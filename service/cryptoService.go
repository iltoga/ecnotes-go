package service

import "github.com/iltoga/ecnotes-go/lib/common"

// CryptoService interface for crypto service implementation (encryption, signing, etc)
type CryptoService interface {
	// Encrypt encrypt plaintext
	Encrypt(plaintext []byte) ([]byte, error)
	// Decrypt decrypt ciphertext
	Decrypt(ciphertext []byte) ([]byte, error)
	// Sign sign plaintext
	Sign(plaintext []byte) ([]byte, error)
	// Verify verify signature
	Verify(plaintext, signature []byte) error
	// GetKeyManager get the key management service
	GetKeyManager() KeyManagementService
}

// KeyManagementService interface for key management service implementation (key generation, etc)
type KeyManagementService interface {
	// GenerateKey generate a new key
	GenerateKey() ([]byte, error)
	// GetPublicKey get the public key
	GetPublicKey() ([]byte, error)
	// GetPrivateKey get the private key
	GetPrivateKey() ([]byte, error)
	// ImportKey import a key into the key management service
	ImportKey(key []byte) error
}

// NewCrytpService create a new crypto service bases on the given key management service and algorithm
func NewCryptoService(algorithm string) CryptoService {
	switch algorithm {
	case common.ENCRYPTION_ALGORITHM_AES_256_CBC:
		kms := NewKeyManagementServiceAES()
		return NewCryptoServiceAES(kms)
	case common.ENCRYPTION_ALGORITHM_RSA_OAEP:
		kms := NewKeyManagementServiceRSA()
		return NewCryptoServiceRSA(kms)
	default:
		return nil
	}
}
