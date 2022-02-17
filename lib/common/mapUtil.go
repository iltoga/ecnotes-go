package common

// GetMapValOrNil gets the value from the map and returns nil if the key is not found
func GetMapValOrNil(m map[string]interface{}, key string) interface{} {
	if val, ok := m[key]; ok {
		return val
	}
	return nil
}
