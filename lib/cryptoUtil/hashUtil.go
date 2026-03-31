package cryptoUtil

import (
	"crypto/sha256"
	"encoding/hex"
	"hash/fnv"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

// FNV32a hashes using fnv32a algorithm
func FNV32a(text string) uint32 {
	algorithm := fnv.New32a()
	algorithm.Write([]byte(text))
	return algorithm.Sum32()
}

// IndexFromString returns the unique (hash) index of an arbitrary string
func IndexFromString(text string) uint32 {
	return FNV32a(text)
}

// GenerateRecoveryPassword derives a strong key from a security answer using PBKDF2-HMAC-SHA256.
// The salt is unique per key (randomly generated at setup time and persisted in config),
// preventing cross-key rainbow table attacks.  600 000 iterations matches the OWASP 2023
// recommendation for PBKDF2-HMAC-SHA256, making offline brute-force expensive.
func GenerateRecoveryPassword(answers []string, salt []byte) string {
	combined := ""
	for _, ans := range answers {
		combined += strings.TrimSpace(strings.ToLower(ans))
	}
	// Use 600,000 iterations as recommended by OWASP for PBKDF2-HMAC-SHA256
	keyBytes := pbkdf2.Key([]byte(combined), salt, 600000, 32, sha256.New)
	return hex.EncodeToString(keyBytes)
}
