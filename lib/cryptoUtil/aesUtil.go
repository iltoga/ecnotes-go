package cryptoUtil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
)

// EncryptWithPassword encrypts a message using a password
func EncryptWithPassword(message string, password string) (encmess string, err error) {
	// hash the password with SHA256 (just to make some noise)
	key := Hash(password)
	return EncryptAES256(key, message)
}

// DecryptWithPassword decrypts an encrypted message using a password
func DecryptWithPassword(message string, password string) (encmess string, err error) {
	// hash the password with SHA256 (just to make some noise)
	key := Hash(password)
	return DecryptAES256(key, message)
}

// EncryptAES256 encrypts a message using AES-256-GCM
//  key is the encryption key (must be 32 bytes)
//  message is the string to be encrypted
//  returns the (b64 encoded) encrypted message, or an error if the key is not 32 bytes
func EncryptAES256(key []byte, message string) (encmess string, err error) {
	plaintext := []byte(message)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	//Create a new GCM - https://en.wikipedia.org/wiki/Galois/Counter_Mode
	//https://golang.org/pkg/crypto/cipher/#NewGCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	// returns to base64 encoded string
	encmess = fmt.Sprintf("%x", ciphertext)
	return
}

// DecryptAES256 ....
func DecryptAES256(key []byte, securemess string) (decodedmess string, err error) {
	enc, err := hex.DecodeString(securemess)
	if err != nil {
		return "", err
	}

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return
	}
	decodedmess = fmt.Sprintf("%s", plaintext)
	return
}
