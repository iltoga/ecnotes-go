package cryptoUtil

import "hash/fnv"

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
