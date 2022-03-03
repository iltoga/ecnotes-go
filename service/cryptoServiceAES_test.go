package service_test

import (
	"bytes"
	"testing"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/stretchr/testify/assert"
)

var (
	aesKey = "shhhhhhhhhhhhhItsaSecret"
)

// mock service.KeyManagementService implementation
type mockAESKeyManagementService struct {
	aesKey []byte
}

// NewMockAESKeyManagementService returns a new mock service.KeyManagementService implementation
func NewMockAESKeyManagementService(priKey []byte) service.KeyManagementService {
	return &mockAESKeyManagementService{priKey}
}

// GenerateKey generates a new mocked key
func (m *mockAESKeyManagementService) GenerateKey() ([]byte, error) {
	return m.aesKey, nil
}

// GetPublicKey returns the mocked public key
func (m *mockAESKeyManagementService) GetPublicKey() ([]byte, error) {
	return m.aesKey, nil
}

// GetPrivateKey returns the mocked private key
func (m *mockAESKeyManagementService) GetPrivateKey() ([]byte, error) {
	return m.aesKey, nil
}

// ImportKey
func (m *mockAESKeyManagementService) ImportKey(key []byte) error {
	m.aesKey = key
	return nil
}

// TestEncrypt tests the encryption of a string (using t *testing.T)
func TestEncryptAES(t *testing.T) {
	// create a new mock service.KeyManagementService implementation
	mockAESKeyManagementService := NewMockAESKeyManagementService([]byte(aesKey))

	// create a new service.EncryptionService implementation
	encryptionService := service.NewCryptoServiceAES(mockAESKeyManagementService)

	// encrypt a string
	testStringB := []byte("test string")
	encrypted, err := encryptionService.Encrypt(testStringB)
	if err != nil {
		t.Errorf("Error while encrypting string: %s", err)
	}

	// decrypt the string
	decrypted, err := encryptionService.Decrypt(encrypted)
	if err != nil {
		t.Errorf("Error while decrypting string: %s", err)
	}

	// check if the decrypted string is the same as the original
	if !bytes.Equal(testStringB, decrypted) {
		t.Errorf("Decrypted string is not the same as the original: %s", decrypted)
	}
}

// TestSignVerify tests the signing and verification of a string (using t *testing.T)
func TestSignAES(t *testing.T) {
	// create a new mock service.KeyManagementService implementation
	mockAESKeyManagementService := NewMockAESKeyManagementService([]byte(aesKey))

	// create a new service.EncryptionService implementation
	signatureService := service.NewCryptoServiceAES(mockAESKeyManagementService)

	// sign a string
	testStringB := []byte("test string")
	signature, err := signatureService.Sign(testStringB)
	assert.Equal(t, common.ERR_SYMMETRIC_KEY_SIGNING_NOT_IMPLEMENTED, err.Error())

	// verify the signature
	err = signatureService.Verify(testStringB, signature)
	assert.Equal(t, common.ERR_SYMMETRIC_KEY_SIGNING_NOT_IMPLEMENTED, err.Error())
}
