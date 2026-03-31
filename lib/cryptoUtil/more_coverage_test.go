package cryptoUtil_test

import (
	"encoding/hex"
	"testing"

	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashHelpersAndWrappers(t *testing.T) {
	t.Parallel()

	input := "ecnotes"
	fnvHash := cryptoUtil.FNV32a(input)
	assert.Equal(t, fnvHash, cryptoUtil.IndexFromString(input))
	assert.NotZero(t, fnvHash)

	encoded := cryptoUtil.EncodedHash(input)
	decoded, err := cryptoUtil.DecodedHash(encoded)
	require.NoError(t, err)
	assert.Equal(t, cryptoUtil.Hash(input), decoded)

	_, err = cryptoUtil.DecodedHash("not-hex")
	assert.Error(t, err)

	ciphertext, err := cryptoUtil.EncryptMessage([]byte("secret note"), "password")
	require.NoError(t, err)

	plaintext, err := cryptoUtil.DecryptMessage(ciphertext, "password")
	require.NoError(t, err)
	assert.Equal(t, []byte("secret note"), plaintext)

	// sanity-check that the encoded hash is valid hex
	_, err = hex.DecodeString(encoded)
	require.NoError(t, err)
}
