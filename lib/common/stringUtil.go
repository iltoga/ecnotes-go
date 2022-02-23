package common

import (
	"encoding/json"
	"strconv"
	"time"
)

// MarshalJSON marshals a struct into JSON
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalJSON unmarshals a JSON string into a struct
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// StringToInt converts a string to an int
func StringToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// StringToInt64 converts a string to an int64
func StringToInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

// StringToTime converts a string to a time
func StringToTime(s string) time.Time {
	t, _ := time.Parse(DefaultTimeFormat, s)
	return t
}

// StringToBool converts a string to a bool
func StringToBool(s string) bool {
	b, _ := strconv.ParseBool(s)
	return b
}
