package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
	"golang.org/x/crypto/sha3"
)

// KeyManagementServiceRSAImpl implementation of the key management service interface
type KeyManagementServiceRSAImpl struct {
	key     []byte
	keyName string
}

// NewKeyManagementServiceRSA  the key management service interface using the RSA OAEP key generation scheme
func NewKeyManagementServiceRSA() KeyManagementService {
	return &KeyManagementServiceRSAImpl{}
}

// GetCertificate get the certificate of the key
func (kms *KeyManagementServiceRSAImpl) GetCertificate() model.EncKey {
	return model.EncKey{
		Key:  kms.key,
		Name: kms.keyName,
		Algo: common.ENCRYPTION_ALGORITHM_RSA_OAEP,
	}
}

// GenerateKey generate a new key
func (kms *KeyManagementServiceRSAImpl) GenerateKey() ([]byte, error) {
	// generate a new key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	// marshal the key
	keyBytes := x509.MarshalPKCS1PrivateKey(key)

	// store the key
	kms.key = keyBytes
	kms.keyName = "default"
	return keyBytes, nil
}

// GetPublicKey get the public key
func (kms *KeyManagementServiceRSAImpl) GetPublicKey() ([]byte, error) {
	// get the key
	key := kms.key
	if key == nil {
		return nil, nil
	}
	// unmarshal the key
	keyBlock, err := x509.ParsePKCS1PrivateKey(key)
	if err != nil {
		return nil, err
	}
	if keyBlock == nil {
		return nil, nil
	}
	// marshal the key
	keyBytes := x509.MarshalPKCS1PublicKey(&keyBlock.PublicKey)
	return keyBytes, nil
}

// GetPrivateKey validates and get back the private key in bytes
func (kms *KeyManagementServiceRSAImpl) GetPrivateKey() ([]byte, error) {
	// get the key
	key := kms.key
	if key == nil {
		return nil, nil
	}
	// unmarshal the key
	keyBlock, _ := x509.ParsePKCS1PrivateKey(key)
	if keyBlock == nil {
		return nil, nil
	}
	// marshal the key
	keyBytes := x509.MarshalPKCS1PrivateKey(keyBlock)
	return keyBytes, nil
}

// ImportKey validate and import the key
func (kms *KeyManagementServiceRSAImpl) ImportKey(key []byte, keyName string) error {
	// validate the key
	keyBlock, err := x509.ParsePKCS1PrivateKey(key)
	if err != nil {
		return err
	}
	// marshal and store the key
	kms.key = x509.MarshalPKCS1PrivateKey(keyBlock)
	kms.keyName = keyName
	return nil
}

// CryptoServiceRSAImpl implementation of the crypto service interface
type CryptoServiceRSAImpl struct {
	keyManagementService KeyManagementService
}

// NewCryptoServiceRSA NewCryptoServiceRSA implement the crypto service interface using the key management service and RSA OAEP encryption scheme
func NewCryptoServiceRSA(keyManagementService KeyManagementService) CryptoService {
	return &CryptoServiceRSAImpl{keyManagementService: keyManagementService}
}

// Encrypt encrypt plaintext using the key management service
func (cs *CryptoServiceRSAImpl) Encrypt(plaintext []byte) ([]byte, error) {
	publicKey, err := cs.keyManagementService.GetPublicKey()
	if err != nil {
		return nil, err
	}
	// parse the rsa public key
	rsaPublicKey, err := x509.ParsePKCS1PublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	// encrypt the plaintext
	ciphertext, err := rsa.EncryptOAEP(sha3.New256(), rand.Reader, rsaPublicKey, plaintext, []byte{})
	if err != nil {
		return nil, err
	}
	return ciphertext, nil
}

// Decrypt decrypt ciphertext using the key management service
func (cs *CryptoServiceRSAImpl) Decrypt(ciphertext []byte) ([]byte, error) {
	// get the private key
	privateKey, err := cs.keyManagementService.GetPrivateKey()
	if err != nil {
		return nil, err
	}
	// parse the rsa private key
	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	// decrypt the ciphertext
	plaintext, err := rsa.DecryptOAEP(sha3.New256(), rand.Reader, rsaPrivateKey, ciphertext, []byte{})
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

// Sign sign plaintext using the key management service
func (cs *CryptoServiceRSAImpl) Sign(plaintext []byte) ([]byte, error) {
	// get the private key
	privateKey, err := cs.keyManagementService.GetPrivateKey()
	if err != nil {
		return nil, err
	}
	// parse the rsa private key
	rsaPrivateKey, err := x509.ParsePKCS1PrivateKey(privateKey)
	if err != nil {
		return nil, err
	}
	// sign the plaintext
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, 0, plaintext)
	if err != nil {
		return nil, err
	}
	return signature, nil
}

// Verify verify signature using the key management service
func (cs *CryptoServiceRSAImpl) Verify(plaintext, signature []byte) error {
	// get the public key
	publicKey, err := cs.keyManagementService.GetPublicKey()
	if err != nil {
		return err
	}
	// parse the rsa public key
	rsaPublicKey, err := x509.ParsePKCS1PublicKey(publicKey)
	if err != nil {
		return err
	}
	// verify the signature
	err = rsa.VerifyPKCS1v15(rsaPublicKey, 0, plaintext, signature)
	if err != nil {
		return err
	}
	return nil
}

// GetKeyManagementService get the key management service
func (cs *CryptoServiceRSAImpl) GetKeyManager() KeyManagementService {
	return cs.keyManagementService
}

// GetAlgorithm get the algorithm
func (cs *CryptoServiceRSAImpl) GetAlgorithm() string {
	return common.ENCRYPTION_ALGORITHM_RSA_OAEP
}
