// Package service contains all business-logic for the application.
//
// # Architecture contract
//
// The UI layer (package ui) MUST stay thin:
//   - UI files may only contain widget construction, layout, and event wiring.
//   - Any operation that touches data, crypto, configuration, or the file system
//     belongs in a service or lib/util, NOT in a ui file.
//   - If you find yourself writing hex.DecodeString, pbkdf2, or certService calls
//     inside a fyne widget callback, stop and add/extend a service method instead.
//
// KeyService is the single point of truth for encryption-key lifecycle:
//   - generating keys
//   - loading / saving keys (with or without password)
//   - recovery-payload creation and verification
//   - key rotation / re-encryption of notes
//
// All methods return plain Go errors; the UI layer is responsible for deciding
// how to surface them (notification, dialog, log, etc.).
package service

import (
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/model"
)

// RecoverySetupResult holds everything returned from CreateRecoveryPayload so the
// caller (UI) can persist the fields without knowing internal key-derivation details.
type RecoverySetupResult struct {
	// EncryptedKeyHex is the master key encrypted with the PBKDF2-derived recovery password,
	// hex-encoded and ready to be stored in config as keyName_recovery.
	EncryptedKeyHex string
	// Salt is the random salt used for PBKDF2; must be stored in config as keyName_recovery_salt.
	Salt string
}

// KeyService manages the full lifecycle of encryption keys.
// Implementations must be safe for concurrent use.
type KeyService interface {
	// TryAutoLoad attempts to load and activate the default key with an empty
	// password (passwordless keys). Returns true when the key was loaded
	// successfully – in that case the caller should skip showing any auth dialog.
	TryAutoLoad() (bool, error)

	// LoadKey validates password and activates the named cert in the crypto service.
	// Corresponds to the "Confirm" action in the Decrypt Encryption Key dialog.
	LoadKey(keyName, password string) error

	// GenerateKey creates a new encryption key for the given algorithm, saves it
	// to the cert store with the supplied password (may be empty), and optionally
	// marks it as the default key in configuration.
	// If securityQuestion and securityAnswer are both non-empty a recovery payload
	// is created and persisted atomically with the key generation.
	GenerateKey(keyName, algo, password string, setDefault bool, securityQuestion, securityAnswer string) (model.EncKey, error)

	// CreateRecoveryPayload derives a per-key PBKDF2 key from securityAnswer and
	// encrypts rawKey with it. The caller must persist the returned fields.
	// A brute-force delay is deliberately NOT applied here – it belongs in the UI
	// (because only interactive paths need it, not programmatic ones).
	CreateRecoveryPayload(rawKey []byte, securityAnswer string) (RecoverySetupResult, error)

	// VerifyAndRecoverKey decrypts the stored recovery payload with the supplied
	// answer, saves the cert under newPassword, and activates the key.
	// It applies a 1-second brute-force delay internally to protect against
	// automated UI-level attacks.
	VerifyAndRecoverKey(keyName, answer, newPassword string) error

	// RotateKey re-encrypts all existing notes to use the current active key.
	// Call this after GenerateKey or VerifyAndRecoverKey succeeds.
	RotateKey(notes []model.Note, newCert model.EncKey) error

	// ImportKey decrypts an exported key payload (format "ALGO:HEX"), validates
	// the algorithm, adds it to the cert store, and re-encrypts all notes.
	ImportKey(encodedKey, algo, password string) (model.EncKey, error)

	// ExportKeyForClipboard returns a portable string ("ALGO:ENCRYPTED_HEX") that
	// can be copied to the clipboard and later imported via ImportKey.
	// The raw key is encrypted with password before export.
	ExportKeyForClipboard(password string) (string, error)

	// HasRecovery reports whether a recovery payload exists for the given key name.
	// The UI uses this to decide whether to show the "Forgot Password?" button.
	HasRecovery(keyName string) bool
}

// KeyServiceImpl is the production implementation of KeyService.
type KeyServiceImpl struct {
	certService   CertService
	confService   ConfigService
	cryptoService CryptoServiceFactory
	noteService   NoteService
}

// NewKeyService constructs a ready-to-use KeyService.
func NewKeyService(
	certService CertService,
	confService ConfigService,
	cryptoService CryptoServiceFactory,
	noteService NoteService,
) KeyService {
	return &KeyServiceImpl{
		certService:   certService,
		confService:   confService,
		cryptoService: cryptoService,
		noteService:   noteService,
	}
}

// TryAutoLoad tries to load the default key with an empty password.
// Returns (true, nil) on success, (false, nil) when no passwordless key exists,
// or (false, err) when an unexpected error occurs.
func (ks *KeyServiceImpl) TryAutoLoad() (bool, error) {
	keyName, err := ks.confService.GetConfig(common.CONFIG_CUR_ENCRYPTION_KEY_NAME)
	if err != nil || keyName == "" {
		return false, nil
	}
	if err := ks.certService.LoadCerts(""); err != nil {
		if strings.Contains(err.Error(), "message authentication failed") {
			return false, nil // not an error – just a password-protected key
		}
		return false, fmt.Errorf("error loading cert store: %w", err)
	}
	cert, err := ks.certService.GetCert(keyName)
	if err != nil {
		return false, fmt.Errorf("configured key %q not found: %w", keyName, err)
	}
	if !common.IsSupportedEncryptionAlgorithm(cert.Algo) {
		return false, fmt.Errorf("unsupported encryption algorithm for key %q: %s", keyName, cert.Algo)
	}
	ks.cryptoService.SetSrv(NewCryptoServiceFactory(cert.Algo))
	if err = ks.cryptoService.GetSrv().GetKeyManager().ImportKey(cert.Key, cert.Name); err != nil {
		return false, err
	}
	return true, nil
}

// LoadKey validates the password and activates the named cert.
func (ks *KeyServiceImpl) LoadKey(keyName, password string) error {
	if err := ks.certService.LoadCerts(password); err != nil {
		if strings.Contains(err.Error(), "message authentication failed") {
			return fmt.Errorf("invalid password: %w", err)
		}
		return fmt.Errorf("error loading cert store: %w", err)
	}
	cert, err := ks.certService.GetCert(keyName)
	if err != nil {
		return fmt.Errorf("key %q not found: %w", keyName, err)
	}
	if !common.IsSupportedEncryptionAlgorithm(cert.Algo) {
		return fmt.Errorf("unsupported encryption algorithm for key %q: %s", keyName, cert.Algo)
	}
	ks.cryptoService.SetSrv(NewCryptoServiceFactory(cert.Algo))
	if err = ks.cryptoService.GetSrv().GetKeyManager().ImportKey(cert.Key, cert.Name); err != nil {
		return fmt.Errorf("error importing key: %w", err)
	}
	return nil
}

// GenerateKey generates a new encryption key, stores it, creates an optional recovery
// payload, and optionally marks it as the application default.
func (ks *KeyServiceImpl) GenerateKey(keyName, algo, password string, setDefault bool, securityQuestion, securityAnswer string) (model.EncKey, error) {
	if !common.IsSupportedEncryptionAlgorithm(algo) {
		return model.EncKey{}, fmt.Errorf("unsupported encryption algorithm: %q", algo)
	}
	ks.cryptoService.SetSrv(NewCryptoServiceFactory(algo))
	rawKey, err := ks.cryptoService.GetSrv().GetKeyManager().GenerateKey()
	if err != nil {
		return model.EncKey{}, fmt.Errorf("error generating key: %w", err)
	}

	var recoveryPayload *RecoverySetupResult
	if securityQuestion != "" && securityAnswer != "" {
		payload, err := ks.CreateRecoveryPayload(rawKey, securityAnswer)
		if err != nil {
			return model.EncKey{}, err
		}
		recoveryPayload = &payload
	}

	cert := model.EncKey{Name: keyName, Algo: algo, Key: rawKey}
	if err := ks.certService.AddCert(cert); err != nil {
		return model.EncKey{}, fmt.Errorf("error adding key to cert store: %w", err)
	}

	if err := ks.certService.SaveCerts(password); err != nil {
		return model.EncKey{}, fmt.Errorf("error saving cert store: %w", err)
	}
	if recoveryPayload != nil {
		if err := ks.persistRecoveryMetadata(keyName, securityQuestion, recoveryPayload.Salt, recoveryPayload.EncryptedKeyHex, algo); err != nil {
			return model.EncKey{}, fmt.Errorf("error persisting recovery metadata: %w", err)
		}
	}
	if setDefault {
		if err := ks.confService.SetConfig(common.CONFIG_CUR_ENCRYPTION_KEY_NAME, keyName); err != nil {
			return model.EncKey{}, fmt.Errorf("error persisting default key name: %w", err)
		}
		if err := ks.confService.SaveConfig(); err != nil {
			return model.EncKey{}, fmt.Errorf("error saving config: %w", err)
		}
	}
	return cert, nil
}

// CreateRecoveryPayload derives a PBKDF2 key from the security answer (using a
// fresh random salt) and encrypts rawKey with it.
func (ks *KeyServiceImpl) CreateRecoveryPayload(rawKey []byte, securityAnswer string) (RecoverySetupResult, error) {
	salt, err := cryptoUtil.SecureRandomStr(32)
	if err != nil {
		return RecoverySetupResult{}, fmt.Errorf("error generating recovery salt: %w", err)
	}
	recoveryPwd := cryptoUtil.GenerateRecoveryPassword([]string{securityAnswer}, []byte(salt))
	encRawKey, err := cryptoUtil.EncryptMessage(rawKey, recoveryPwd)
	if err != nil {
		return RecoverySetupResult{}, fmt.Errorf("error encrypting recovery payload: %w", err)
	}
	return RecoverySetupResult{
		EncryptedKeyHex: hex.EncodeToString(encRawKey),
		Salt:            salt,
	}, nil
}

// PersistRecoveryMetadata stores everything related to a recovery payload in config.
// Separate from CreateRecoveryPayload so the UI controls when config is written.
func (ks *KeyServiceImpl) persistRecoveryMetadata(keyName, question, saltStr, encHex, algo string) error {
	if err := ks.confService.SetConfig(keyName+"_recovery", encHex); err != nil {
		return err
	}
	if err := ks.confService.SetConfig(keyName+"_recovery_question", question); err != nil {
		return err
	}
	if err := ks.confService.SetConfig(keyName+"_recovery_salt", saltStr); err != nil {
		return err
	}
	if err := ks.confService.SetConfig(keyName+"_algo", algo); err != nil {
		return err
	}
	return ks.confService.SaveConfig()
}

// VerifyAndRecoverKey decrypts the stored recovery payload, saves a fresh cert
// store with newPassword, and activates the key in the crypto service.
// A 1-second delay is applied to rate-limit interactive brute-force attempts.
func (ks *KeyServiceImpl) VerifyAndRecoverKey(keyName, answer, newPassword string) error {
	// Deliberate 1-second delay – protects only the interactive UI path.
	time.Sleep(1 * time.Second)

	encRawKeyStr, err := ks.confService.GetConfig(keyName + "_recovery")
	if err != nil || encRawKeyStr == "" {
		return fmt.Errorf("no recovery data found for key %q", keyName)
	}
	encRawKey, err := hex.DecodeString(encRawKeyStr)
	if err != nil {
		return fmt.Errorf("invalid recovery payload encoding: %w", err)
	}

	saltStr, err := ks.confService.GetConfig(keyName + "_recovery_salt")
	if err != nil || saltStr == "" {
		saltStr = common.RecoveryFallbackSalt // backwards-compat with pre-salt keys
	}

	recoveryPwd := cryptoUtil.GenerateRecoveryPassword([]string{answer}, []byte(saltStr))
	decryptedKey, err := cryptoUtil.DecryptMessage(encRawKey, recoveryPwd)
	if err != nil {
		return fmt.Errorf("incorrect answer: %w", err)
	}

	algo, err := ks.confService.GetConfig(keyName + "_algo")
	if err != nil || algo == "" {
		algo = common.ENCRYPTION_ALGORITHM_AES_256_CBC
	}
	if !common.IsSupportedEncryptionAlgorithm(algo) {
		return fmt.Errorf("unsupported encryption algorithm for recovered key %q: %s", keyName, algo)
	}

	cert := model.EncKey{Name: keyName, Algo: algo, Key: decryptedKey}
	_ = ks.certService.RemoveCert(keyName) // remove stale entry (may not exist yet)
	if err := ks.certService.AddCert(cert); err != nil {
		return fmt.Errorf("error adding recovered key: %w", err)
	}
	if err := ks.certService.SaveCerts(newPassword); err != nil {
		return fmt.Errorf("error saving cert store with new password: %w", err)
	}

	ks.cryptoService.SetSrv(NewCryptoServiceFactory(cert.Algo))
	if err = ks.cryptoService.GetSrv().GetKeyManager().ImportKey(cert.Key, cert.Name); err != nil {
		return fmt.Errorf("error activating recovered key: %w", err)
	}
	return nil
}

// RotateKey re-encrypts all supplied notes to use newCert.
func (ks *KeyServiceImpl) RotateKey(notes []model.Note, newCert model.EncKey) error {
	if err := ks.noteService.ReEncryptNotes(notes, newCert); err != nil {
		return fmt.Errorf("error re-encrypting notes: %w", err)
	}
	return nil
}

// ImportKey decrypts an exported key payload, adds it to the cert store, and
// re-encrypts all notes.
func (ks *KeyServiceImpl) ImportKey(encodedKey, algo, password string) (model.EncKey, error) {
	if encodedKey == "" {
		return model.EncKey{}, fmt.Errorf("encrypted key is empty")
	}
	if prefix, payload, ok := strings.Cut(encodedKey, ":"); ok && common.IsSupportedEncryptionAlgorithm(prefix) {
		if algo != "" && algo != prefix {
			return model.EncKey{}, fmt.Errorf("algorithm mismatch between payload %q and selected %q", prefix, algo)
		}
		algo = prefix
		encodedKey = payload
	}
	if !common.IsSupportedEncryptionAlgorithm(algo) {
		return model.EncKey{}, fmt.Errorf("unsupported encryption algorithm: %q", algo)
	}
	encKey, err := hex.DecodeString(encodedKey)
	if err != nil {
		return model.EncKey{}, fmt.Errorf("invalid encrypted key encoding: %w", err)
	}
	rawKey, err := cryptoUtil.DecryptMessage(encKey, password)
	if err != nil {
		return model.EncKey{}, fmt.Errorf("error decrypting imported key: %w", err)
	}
	cert := model.EncKey{Name: "Imported key", Algo: algo, Key: rawKey}
	if err := ks.certService.AddCert(cert); err != nil {
		return model.EncKey{}, fmt.Errorf("error adding key to cert store: %w", err)
	}
	if err := ks.certService.SaveCerts(password); err != nil {
		return model.EncKey{}, fmt.Errorf("error saving cert store: %w", err)
	}
	if err := ks.confService.SetConfig(common.CONFIG_CUR_ENCRYPTION_KEY_NAME, cert.Name); err != nil {
		return model.EncKey{}, fmt.Errorf("error persisting key name: %w", err)
	}
	if err := ks.confService.SaveConfig(); err != nil {
		return model.EncKey{}, fmt.Errorf("error saving config: %w", err)
	}
	notes, err := ks.noteService.GetNotes()
	if err != nil {
		return model.EncKey{}, fmt.Errorf("error loading notes for re-encryption: %w", err)
	}
	if err := ks.noteService.ReEncryptNotes(notes, cert); err != nil {
		return model.EncKey{}, fmt.Errorf("error re-encrypting notes: %w", err)
	}
	return cert, nil
}

// ExportKeyForClipboard encrypts the current default raw key with password and
// returns a portable "ALGO:ENCRYPTED_HEX" string suitable for clipboard copy.
// An empty password uses an empty-string AES key (same as passwordless keys).
func (ks *KeyServiceImpl) ExportKeyForClipboard(password string) (string, error) {
	keyName, err := ks.confService.GetConfig(common.CONFIG_CUR_ENCRYPTION_KEY_NAME)
	if err != nil || keyName == "" {
		return "", fmt.Errorf("no default encryption key configured")
	}
	cert, err := ks.certService.GetCert(keyName)
	if err != nil {
		return "", fmt.Errorf("could not load encryption key: %w", err)
	}
	encKey, err := cryptoUtil.EncryptMessage(cert.Key, password)
	if err != nil {
		return "", fmt.Errorf("error encrypting key for export: %w", err)
	}
	return fmt.Sprintf("%s:%s", cert.Algo, hex.EncodeToString(encKey)), nil
}

// HasRecovery reports whether a recovery payload exists for the given key name.
func (ks *KeyServiceImpl) HasRecovery(keyName string) bool {
	q, err := ks.confService.GetConfig(keyName + "_recovery_question")
	return err == nil && q != ""
}
