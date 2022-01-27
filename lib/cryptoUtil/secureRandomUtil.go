package cryptoUtil

import (
	"crypto/rand"
	"encoding/hex"

	"golang.org/x/crypto/sha3"
)

const chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// SecureRandomStr returns a random string of the given length
//	length: the length of the string to return
//	returns: a random string of the given length
//					 error: if an error occurs
//	notes: uses crypto/rand to generate random bytes
func SecureRandomStr(length int) (string, error) {
	bytes := make([]byte, length)

	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}

	return string(bytes), nil
}

// Hash ....
func Hash(s string) string {
	h := sha3.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
