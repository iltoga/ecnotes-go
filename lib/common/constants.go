package common

// EncryptionKeyAction enum to describe the encryption key action
type EncryptionKeyAction int64

const (
	EncryptionKey_Generate EncryptionKeyAction = iota
	EncryptionKey_Decrypt
	EncryptionKey_Encrypt
	EncryptionKey_Verify
)
