package common

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSupportedEncryptionAlgorithm(t *testing.T) {
	t.Parallel()

	assert.True(t, IsSupportedEncryptionAlgorithm(ENCRYPTION_ALGORITHM_AES_256_CBC))
	assert.True(t, IsSupportedEncryptionAlgorithm(ENCRYPTION_ALGORITHM_RSA_OAEP))
	assert.False(t, IsSupportedEncryptionAlgorithm(""))
	assert.False(t, IsSupportedEncryptionAlgorithm("unsupported"))
}

func TestGetMapValOrNil(t *testing.T) {
	t.Parallel()

	payload := map[string]interface{}{
		"str":   "value",
		"count": 42,
	}

	assert.Equal(t, "value", GetMapValOrNil(payload, "str"))
	assert.Equal(t, 42, GetMapValOrNil(payload, "count"))
	assert.Nil(t, GetMapValOrNil(payload, "missing"))
}

func TestMarshalAndUnmarshalJSON(t *testing.T) {
	t.Parallel()

	type sample struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	encoded, err := MarshalJSON(sample{Name: "Ada", Age: 36})
	require.NoError(t, err)

	var decoded sample
	require.NoError(t, UnmarshalJSON(encoded, &decoded))
	assert.Equal(t, sample{Name: "Ada", Age: 36}, decoded)
}

func TestStringHelpers(t *testing.T) {
	t.Parallel()

	assert.Equal(t, 123, StringToInt("123"))
	assert.Equal(t, 0, StringToInt("bad-int"))
	assert.Equal(t, int64(456), StringToInt64("456"))
	assert.Equal(t, int64(0), StringToInt64("bad-int64"))
	assert.True(t, StringToBool("true"))
	assert.False(t, StringToBool("not-a-bool"))

	now := time.Date(2026, 3, 31, 12, 34, 56, 789000000, time.UTC)
	formatted := FormatTime(now)
	assert.True(t, now.Equal(StringToTime(formatted)))
	assert.True(t, now.Equal(TimestampToTime(now.UnixMilli())))
}

func TestGetCurrentHelpers(t *testing.T) {
	t.Parallel()

	currentTime := GetCurrentTime()
	parsed, err := time.Parse(DefaultTimeFormat, currentTime)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), parsed, time.Second)

	timestamp := GetCurrentTimestamp()
	assert.InDelta(t, time.Now().UnixMilli(), timestamp, 2500)
}

func TestGetUserHomeDir(t *testing.T) {
	t.Parallel()

	home := GetUserHomeDir()
	require.NotEmpty(t, home)
	_, err := os.Stat(home)
	require.NoError(t, err)
}

func TestStringHelpersJSONCompatibility(t *testing.T) {
	t.Parallel()

	type payload struct {
		UpdatedAt string `json:"updated_at"`
	}

	raw, err := json.Marshal(payload{UpdatedAt: GetCurrentTime()})
	require.NoError(t, err)

	var decoded payload
	require.NoError(t, json.Unmarshal(raw, &decoded))
	assert.NotEmpty(t, decoded.UpdatedAt)
}
