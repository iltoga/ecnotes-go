package common

import "encoding/json"

// MarshalJSON marshals a struct into JSON
func MarshalJSON(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// UnmarshalJSON unmarshals a JSON string into a struct
func UnmarshalJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
