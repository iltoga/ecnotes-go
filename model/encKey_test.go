package model

import (
	"encoding/json"
	"testing"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestByteStringJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := ByteString([]byte("ecnotes"))
	data, err := json.Marshal(original)
	require.NoError(t, err)
	assert.Equal(t, `"65636e6f746573"`, string(data))

	var decoded ByteString
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original, decoded)
}

func TestEncKeyJSONRoundTrip(t *testing.T) {
	t.Parallel()

	original := EncKey{
		Name: "alpha",
		Algo: common.ENCRYPTION_ALGORITHM_AES_256_CBC,
		Key:  ByteString([]byte("alpha-key")),
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded EncKey
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, original, decoded)
}
