package model

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type EncKey struct {
	Name string     `json:"name"`
	Algo string     `json:"algo"`
	Key  ByteString `json:"key"`
}

type ByteString []byte

// MarshalJSON serializes ByteArray to hex
func (s ByteString) MarshalJSON() ([]byte, error) {
	bytes, err := json.Marshal(fmt.Sprintf("%x", string(s)))
	return bytes, err
}

// UnmarshalJSON deserializes ByteArray to hex
func (s *ByteString) UnmarshalJSON(data []byte) error {
	var x string
	err := json.Unmarshal(data, &x)
	if err == nil {
		str, e := hex.DecodeString(x)
		*s = ByteString([]byte(str))
		err = e
	}
	return err
}
