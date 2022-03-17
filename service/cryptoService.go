package service

import (
	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
)

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
	// GetAlgorithm get the algorithm
	GetAlgorithm() string
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
	ImportKey(key []byte, keyName string) error
	// GetCertificate get the certificate for the given key
	GetCertificate() model.EncKey
}

// NewCrytpServiceFactory create a new crypto service bases on the given key management service and algorithm and inject it into the CryptoServiceImpl
func NewCryptoServiceFactory(algorithm string) CryptoService {
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

// to store the crypto service and allow to switch between crypto services

type CryptoServiceFactory interface {
	GetSrv() CryptoService
	SetSrv(srv CryptoService)
}

type CryptoServiceFactoryImpl struct {
	Srv CryptoService
}

func (c *CryptoServiceFactoryImpl) GetSrv() CryptoService {
	return c.Srv
}

func (c *CryptoServiceFactoryImpl) SetSrv(srv CryptoService) {
	c.Srv = srv
}
