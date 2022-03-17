package service

import (
	"errors"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
)

// KeyManagementServiceAES interface for key management service implementation (key generation, etc)
type KeyManagementServiceAES struct {
	key     []byte
	keyName string
}

// NewKeyManagementServiceAES  the key management service interface using the AES key generation scheme
func NewKeyManagementServiceAES() KeyManagementService {
	return &KeyManagementServiceAES{}
}

// GenerateKey generate a new key for AES (symmetric) encryption
func (kms *KeyManagementServiceAES) GenerateKey() ([]byte, error) {
	// generate a new key
	keyStr, err := cryptoUtil.SecureRandomStr(common.ENCRYPTION_KEY_LENGTH)
	if err != nil {
		return nil, err
	}
	kms.key = []byte(keyStr)
	kms.keyName = "default"
	return kms.key, nil
}

// GetPublicKey get the public key
// Since AES is symmetric, the public key is the same as the private key
func (kms *KeyManagementServiceAES) GetPublicKey() ([]byte, error) {
	return kms.GetPrivateKey()
}

// GetPrivateKey validates and get back the private key in bytes
func (kms *KeyManagementServiceAES) GetPrivateKey() ([]byte, error) {
	// get the key
	key := kms.key
	if key == nil {
		return nil, errors.New(common.ERR_NO_KEY)
	}
	return key, nil
}

// ImportKey import a key into the key management service
func (kms *KeyManagementServiceAES) ImportKey(key []byte, keyName string) error {
	kms.key = key
	kms.keyName = keyName
	return nil
}

// CryptoServiceAES interface for crypto service implementation (encryption, signing, etc)
type CryptoServiceAES struct {
	keyManagementService KeyManagementService
}

// NewCryptoServiceAES  the crypto service interface using the AES key generation scheme
func NewCryptoServiceAES(keyManagementService KeyManagementService) CryptoService {
	return &CryptoServiceAES{keyManagementService}
}

// Encrypt plaintext using AES encryption
func (cs *CryptoServiceAES) Encrypt(plaintext []byte) ([]byte, error) {
	// get the key
	key, err := cs.keyManagementService.GetPublicKey()
	if err != nil {
		return nil, err
	}
	// TODO: avoid double cast to []byte by refactoring the cryptoUtil.AESEncryptMessage
	ciphertext, err := cryptoUtil.EncryptMessage(plaintext, string(key))
	if err != nil {
		return nil, err
	}
	return []byte(ciphertext), nil
}

// Decrypt ciphertext
func (cs *CryptoServiceAES) Decrypt(ciphertext []byte) ([]byte, error) {
	key, err := cs.keyManagementService.GetPublicKey()
	if err != nil {
		return nil, err
	}
	// decrypt the ciphertext
	// TODO: avoid double cast to []byte by refactoring the cryptoUtil.AESDecryptMessage
	plaintext, err := cryptoUtil.DecryptMessage(ciphertext, string(key))
	if err != nil {
		return nil, err
	}
	return []byte(plaintext), nil
}

// Sign This method always return err because signing is proper of public key cryptography and not of symmetric cryptography
func (cs *CryptoServiceAES) Sign(plaintext []byte) ([]byte, error) {
	return nil, errors.New(common.ERR_SYMMETRIC_KEY_SIGNING_NOT_IMPLEMENTED)
}

// Verify This method always return err because signing is proper of public key cryptography and not of symmetric cryptography
func (cs *CryptoServiceAES) Verify(plaintext, signature []byte) error {
	return errors.New(common.ERR_SYMMETRIC_KEY_SIGNING_NOT_IMPLEMENTED)
}

// GetKeyManager get the key management service
func (cs *CryptoServiceAES) GetKeyManager() KeyManagementService {
	return cs.keyManagementService
}

// GetAlgorithm get the algorithm name
func (cs *CryptoServiceAES) GetAlgorithm() string {
	return common.ENCRYPTION_ALGORITHM_AES_256_CBC
}
