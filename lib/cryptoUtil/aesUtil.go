package cryptoUtil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// EncryptWithPassword encrypts a message using a password
// The message is encrypted using AES-256-CFB
func EncryptWithPassword(message string, password string) (encmess string, err error) {
	// hash the password with SHA256 (just to make some noise)
	key := []byte(Hash(password))
	return EncryptAES256(key, message)
}

// DecryptWithPassword decrypts an encrypted message using a password
func DecryptWithPassword(message string, password string) (encmess string, err error) {
	// hash the password with SHA256 (just to make some noise)
	key := []byte(Hash(password))
	return DecryptAES256(key, message)
}

// EncryptAES256 encrypts a message using AES-256-CFB
//  key is the encryption key (must be 32 bytes)
//  message is the string to be encrypted
//  returns the (b64 encoded) encrypted message, or an error if the key is not 32 bytes
func EncryptAES256(key []byte, message string) (encmess string, err error) {
	plainText := []byte(message)

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	// IV needs to be unique, but doesn't have to be secure.
	// It's common to put it at the beginning of the ciphertext.
	cipherText := make([]byte, aes.BlockSize+len(plainText))
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(cipherText[aes.BlockSize:], plainText)

	// returns to base64 encoded string
	encmess = base64.URLEncoding.EncodeToString(cipherText)
	return
}

// DecryptAES256 ....
func DecryptAES256(key []byte, securemess string) (decodedmess string, err error) {
	cipherText, err := base64.URLEncoding.DecodeString(securemess)
	if err != nil {
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return
	}

	if len(cipherText) < aes.BlockSize {
		err = errors.New("ciphertext block size is too short")
		return
	}

	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(cipherText, cipherText)

	decodedmess = string(cipherText)
	return
}
