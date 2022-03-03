package common

// IsSupportedEncryptionAlgorithm returns true if the specified encryption algorithm is supported by the application.
// The supported algorithms are defined as constants
func IsSupportedEncryptionAlgorithm(algorithm string) bool {
	switch algorithm {
	case ENCRYPTION_ALGORITHM_AES_256_CBC:
		return true
	case ENCRYPTION_ALGORITHM_RSA_OAEP:
		return true
	default:
		return false
	}
}
