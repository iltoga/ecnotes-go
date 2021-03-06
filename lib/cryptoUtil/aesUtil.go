package cryptoUtil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"
)

// EncryptMessage encrypts a message using a password
func EncryptMessage(message []byte, passPhrase string) ([]byte, error) {
	// hash the password with SHA256 (just to make some noise)
	key := Hash(passPhrase)
	return EncryptAES256(key, message)
}

// DecryptMessage decrypts an encrypted message using a password
func DecryptMessage(message []byte, passPhrase string) ([]byte, error) {
	// hash the password with SHA256 (just to make some noise)
	key := Hash(passPhrase)
	return DecryptAES256(key, message)
}

// EncryptAES256 encrypts a message using AES-256-GCM
//  key is the encryption key (must be 32 bytes)
//  message is the string to be encrypted
//  returns the (b64 encoded) encrypted message, or an error if the key is not 32 bytes
func EncryptAES256(key []byte, plaintext []byte) (encmess []byte, err error) {
	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	// Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	// https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	// Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	// Encrypt the data using aesGCM.Seal
	// Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	// returns to base64 encoded string
	encmess = ciphertext
	return
}

// DecryptAES256 ....
func DecryptAES256(key []byte, securemess []byte) (decodedmess []byte, err error) {
	// Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	// Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	// Get the nonce size
	nonceSize := aesGCM.NonceSize()

	// Extract the nonce from the encrypted data
	nonce, ciphertext := securemess[:nonceSize], securemess[nonceSize:]

	// Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return
	}
	decodedmess = plaintext
	return
}
