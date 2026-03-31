package service_test

import (
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"testing"

	toml "github.com/pelletier/go-toml"
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ──────────────────────────────────────────────────────────────────────────────
// Minimal fakes (only what KeyService actually calls)
// ──────────────────────────────────────────────────────────────────────────────

type fakeCertService struct {
	mu    sync.Mutex
	certs map[string]model.EncKey
	count int // CountCerts return value
	loadErr error
}

func newFakeCertService() *fakeCertService {
	return &fakeCertService{certs: make(map[string]model.EncKey)}
}
func (f *fakeCertService) CountCerts() (int, error)        { return f.count, nil }
func (f *fakeCertService) LoadCerts(pwd string) error      { return f.loadErr }
func (f *fakeCertService) SaveCerts(pwd string) error      { return nil }
func (f *fakeCertService) GetCert(name string) (*model.EncKey, error) {
	f.mu.Lock(); defer f.mu.Unlock()
	if c, ok := f.certs[name]; ok {
		return &c, nil
	}
	return nil, errors.New(common.ERR_CERT_NOT_FOUND)
}
func (f *fakeCertService) AddCert(cert model.EncKey) error {
	f.mu.Lock(); defer f.mu.Unlock()
	f.certs[cert.Name] = cert
	return nil
}
func (f *fakeCertService) RemoveCert(name string) error {
	f.mu.Lock(); defer f.mu.Unlock()
	delete(f.certs, name)
	return nil
}

// fakeConfService stores key→value pairs in memory.
type fakeConfService struct {
	mu   sync.RWMutex
	data map[string]string
}

func newFakeConfService() *fakeConfService { return &fakeConfService{data: make(map[string]string)} }
func (f *fakeConfService) GetResourcePath() string { return "" }
func (f *fakeConfService) GetGlobal(key string) (string, error) {
	f.mu.RLock(); defer f.mu.RUnlock()
	if v, ok := f.data[key]; ok { return v, nil }
	return "", errors.New("not found")
}
func (f *fakeConfService) GetGlobalBytes(key string) ([]byte, error) { return nil, nil }
func (f *fakeConfService) SetGlobal(key, value string) {
	f.mu.Lock(); defer f.mu.Unlock(); f.data[key] = value
}
func (f *fakeConfService) SetGlobalBytes(key string, value []byte) {}
func (f *fakeConfService) GetConfig(key string) (string, error) {
	f.mu.RLock(); defer f.mu.RUnlock()
	if v, ok := f.data[key]; ok { return v, nil }
	return "", errors.New("not found")
}
func (f *fakeConfService) GetConfigBytes(key string) ([]byte, error) { return nil, nil }
func (f *fakeConfService) SetConfig(key, value string) error {
	f.mu.Lock(); defer f.mu.Unlock(); f.data[key] = value; return nil
}
func (f *fakeConfService) SetConfigBytes(key string, value []byte) error { return nil }
func (f *fakeConfService) LoadConfig() error                             { return nil }
func (f *fakeConfService) ParseConfigTree(t *toml.Tree)                 {}
func (f *fakeConfService) SaveConfig() error                             { return nil }

// fakeNoteService — only ReEncryptNotes and GetNotes are called by KeyService.
type fakeNoteService struct{ reEncCalled bool }

func (f *fakeNoteService) ReEncryptNotes(notes []model.Note, cert model.EncKey) error {
	f.reEncCalled = true; return nil
}
func (f *fakeNoteService) SaveEncryptedNotes(notes []model.Note) error             { return nil }
func (f *fakeNoteService) GetNotes() ([]model.Note, error)                         { return nil, nil }
func (f *fakeNoteService) GetNote(id int) (*model.Note, error)                     { return nil, nil }
func (f *fakeNoteService) GetNoteWithContent(id int) (*model.Note, error)          { return nil, nil }
func (f *fakeNoteService) GetNoteIDFromTitle(title string) int                     { return 0 }
func (f *fakeNoteService) GetTitles() []string                                     { return nil }
func (f *fakeNoteService) SearchNotes(q string, fuzzy bool) ([]string, error)      { return nil, nil }
func (f *fakeNoteService) CreateNote(n *model.Note) error                          { return nil }
func (f *fakeNoteService) UpdateNote(n *model.Note) error                          { return nil }
func (f *fakeNoteService) UpdateNoteTitle(old, new string) (int, error)            { return 0, nil }
func (f *fakeNoteService) UpdateNoteContent(n *model.Note) error                   { return nil }
func (f *fakeNoteService) DeleteNote(id int) error                                 { return nil }
func (f *fakeNoteService) EncryptNote(n *model.Note) error                         { return nil }
func (f *fakeNoteService) DecryptNote(n *model.Note) error                         { return nil }

// ──────────────────────────────────────────────────────────────────────────────
// Tests
// ──────────────────────────────────────────────────────────────────────────────

func newTestKeyService() (service.KeyService, *fakeCertService, *fakeConfService, *fakeNoteService) {
	cert := newFakeCertService()
	conf := newFakeConfService()
	note := &fakeNoteService{}
	crypto := &service.CryptoServiceFactoryImpl{}
	ks := service.NewKeyService(cert, conf, crypto, note)
	return ks, cert, conf, note
}

func TestKeyService_GenerateKey_StoresAndActivates(t *testing.T) {
	ks, certSvc, confSvc, _ := newTestKeyService()

	cert, err := ks.GenerateKey("myKey", common.ENCRYPTION_ALGORITHM_AES_256_CBC, "secret", true, "", "")
	require.NoError(t, err)
	assert.Equal(t, "myKey", cert.Name)
	assert.Equal(t, common.ENCRYPTION_ALGORITHM_AES_256_CBC, cert.Algo)
	assert.NotEmpty(t, cert.Key)

	// Must be stored in cert store
	stored, err := certSvc.GetCert("myKey")
	require.NoError(t, err)
	assert.Equal(t, cert.Name, stored.Name)

	// Default key name must be persisted
	defName, err := confSvc.GetConfig(common.CONFIG_CUR_ENCRYPTION_KEY_NAME)
	require.NoError(t, err)
	assert.Equal(t, "myKey", defName)
}

func TestKeyService_GenerateKey_WithRecovery_PersistsMetadata(t *testing.T) {
	ks, _, confSvc, _ := newTestKeyService()

	_, err := ks.GenerateKey(
		"myKey", common.ENCRYPTION_ALGORITHM_AES_256_CBC, "",
		true, "What is your pet's name?", "fluffy",
	)
	require.NoError(t, err)

	question, err := confSvc.GetConfig("myKey_recovery_question")
	require.NoError(t, err)
	assert.Equal(t, "What is your pet's name?", question)

	salt, err := confSvc.GetConfig("myKey_recovery_salt")
	require.NoError(t, err)
	assert.NotEmpty(t, salt, "random salt must be stored")

	encHex, err := confSvc.GetConfig("myKey_recovery")
	require.NoError(t, err)
	assert.NotEmpty(t, encHex, "recovery payload must be stored")
}

func TestKeyService_GenerateKey_RejectsUnsupportedAlgorithm(t *testing.T) {
	ks, _, _, _ := newTestKeyService()
	_, err := ks.GenerateKey("myKey", "bogus", "", false, "", "")
	assert.Error(t, err)
}

func TestKeyService_VerifyAndRecoverKey_CorrectAnswer(t *testing.T) {
	ks, _, confSvc, _ := newTestKeyService()

	_, err := ks.GenerateKey(
		"myKey", common.ENCRYPTION_ALGORITHM_AES_256_CBC, "oldPwd",
		true, "Pet name?", "fluffy",
	)
	require.NoError(t, err)

	// Simulate loading the cert store so GetCert works after recovery
	confSvc.SetConfig(common.CONFIG_CUR_ENCRYPTION_KEY_NAME, "myKey")

	err = ks.VerifyAndRecoverKey("myKey", "fluffy", "newPwd")
	require.NoError(t, err)
}

func TestKeyService_VerifyAndRecoverKey_WrongAnswer(t *testing.T) {
	ks, _, _, _ := newTestKeyService()

	_, err := ks.GenerateKey(
		"myKey", common.ENCRYPTION_ALGORITHM_AES_256_CBC, "oldPwd",
		true, "Pet name?", "fluffy",
	)
	require.NoError(t, err)

	err = ks.VerifyAndRecoverKey("myKey", "wrongAnswer", "newPwd")
	assert.Error(t, err, "wrong answer must return an error")
}

func TestKeyService_TryAutoLoad(t *testing.T) {
	ks, certSvc, confSvc, _ := newTestKeyService()

	require.NoError(t, certSvc.AddCert(model.EncKey{
		Name: "auto",
		Algo: common.ENCRYPTION_ALGORITHM_AES_256_CBC,
		Key:  []byte("auto-key-32-bytes-auto-key-32-by"),
	}))
	require.NoError(t, confSvc.SetConfig(common.CONFIG_CUR_ENCRYPTION_KEY_NAME, "auto"))

	ok, err := ks.TryAutoLoad()
	require.NoError(t, err)
	assert.True(t, ok)

	certSvc.loadErr = errors.New("cipher: message authentication failed")
	ok, err = ks.TryAutoLoad()
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestKeyService_LoadKey(t *testing.T) {
	ks, certSvc, confSvc, _ := newTestKeyService()

	require.NoError(t, certSvc.AddCert(model.EncKey{
		Name: "manual",
		Algo: common.ENCRYPTION_ALGORITHM_AES_256_CBC,
		Key:  []byte("manual-key-32-bytes-manual-key-"),
	}))
	require.NoError(t, confSvc.SetConfig(common.CONFIG_CUR_ENCRYPTION_KEY_NAME, "manual"))

	require.NoError(t, ks.LoadKey("manual", "secret"))

	certSvc.loadErr = errors.New("cipher: message authentication failed")
	err := ks.LoadKey("manual", "wrong")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid password")
}

func TestKeyService_HasRecovery(t *testing.T) {
	ks, _, confSvc, _ := newTestKeyService()

	assert.False(t, ks.HasRecovery("missing"))
	require.NoError(t, confSvc.SetConfig("with_recovery_recovery_question", "What is your pet's name?"))
	assert.True(t, ks.HasRecovery("with_recovery"))
}

func TestKeyService_CreateRecoveryPayload_UniquePerCall(t *testing.T) {
	ks, _, _, _ := newTestKeyService()
	rawKey := []byte("test-raw-key-32-bytes-long-xxxxxx")

	r1, err := ks.CreateRecoveryPayload(rawKey, "fluffy")
	require.NoError(t, err)
	r2, err := ks.CreateRecoveryPayload(rawKey, "fluffy")
	require.NoError(t, err)

	// Different salts → different ciphertext even for same answer
	assert.NotEqual(t, r1.Salt, r2.Salt)
	assert.NotEqual(t, r1.EncryptedKeyHex, r2.EncryptedKeyHex)
}

func TestKeyService_ImportKey_InvalidAlgo(t *testing.T) {
	ks, _, _, _ := newTestKeyService()
	_, err := ks.ImportKey("somehex", "unsupported-algo", "pwd")
	assert.Error(t, err)
}

func TestKeyService_ExportKeyForClipboard_NoKeyConfigured(t *testing.T) {
	ks, _, _, _ := newTestKeyService()
	_, err := ks.ExportKeyForClipboard("pwd")
	assert.Error(t, err)
}

func TestKeyService_ExportImportClipboard_RoundTrip(t *testing.T) {
	ks, certSvc, confSvc, noteSvc := newTestKeyService()

	originalKey := []byte("0123456789abcdef0123456789abcdef")
	require.NoError(t, certSvc.AddCert(model.EncKey{
		Name: "original",
		Algo: common.ENCRYPTION_ALGORITHM_AES_256_CBC,
		Key:  originalKey,
	}))
	require.NoError(t, confSvc.SetConfig(common.CONFIG_CUR_ENCRYPTION_KEY_NAME, "original"))

	exported, err := ks.ExportKeyForClipboard("transport-password")
	require.NoError(t, err)

	prefix, payload, ok := strings.Cut(exported, ":")
	require.True(t, ok)
	assert.Equal(t, common.ENCRYPTION_ALGORITHM_AES_256_CBC, prefix)
	_, err = hex.DecodeString(payload)
	require.NoError(t, err)

	imported, err := ks.ImportKey(exported, "", "transport-password")
	require.NoError(t, err)
	assert.Equal(t, "Imported key", imported.Name)
	assert.Equal(t, common.ENCRYPTION_ALGORITHM_AES_256_CBC, imported.Algo)
	assert.Equal(t, originalKey, []byte(imported.Key))

	stored, err := certSvc.GetCert("Imported key")
	require.NoError(t, err)
	assert.Equal(t, originalKey, []byte(stored.Key))

	defaultName, err := confSvc.GetConfig(common.CONFIG_CUR_ENCRYPTION_KEY_NAME)
	require.NoError(t, err)
	assert.Equal(t, "Imported key", defaultName)
	assert.True(t, noteSvc.reEncCalled)
}

func TestKeyService_RotateKey_DelegatesToNoteService(t *testing.T) {
	ks, _, _, noteSvc := newTestKeyService()
	cert := model.EncKey{Name: "k", Algo: common.ENCRYPTION_ALGORITHM_AES_256_CBC}
	err := ks.RotateKey([]model.Note{}, cert)
	require.NoError(t, err)
	assert.True(t, noteSvc.reEncCalled)
}
