package service_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/stretchr/testify/assert"
)

var (
	priKeyHex = "308204a50201000282010100b2e8436f3ba46d9a75d39d5b130030b6c542c491378dc6082eeaf31c29962a286d871b6dd56ffd80c9cbce63e6b242cbb23f9541c1c97dcde8852221e50db3073b8da2a4ea2610fc0b1eadee23e84160f7aed850b9b2ae7116d798a2112a88c4ded9acd0c32f00baf42655910f81e83c628f41ada17bb096acf26d71e5edf646bfb8b62148a40e7ba7c94be66858468be3cfdaee0dff8327c4844a4e5d2d707b9a54defab23c058ca3482426971ae9e86d675bd9059cb0783e4577cff6f593ca4f7e862ee94a6bf74c420bfbd7132f45aac5a198f91951aa0c26928931643b8c48c4eaf7058046780283a431a30c6159761391a98563f22b233f2344854338fb02030100010282010039c712c8207dd8bbb263b604cc9d1a1e5c945481056ceed083be72e6dc73578818df3237855f9681fa29acacccbb33212f9ea3284a5a351bc3850361e8e444b60840948f27e34546f09c66d56a993e4bff9162e0a72812780945755099b49fd8dc9375e131b7c3479d43a80ca1f2753ad325aab3555c69ca2f6e57741a2a80870c3d710e87ea15cc86a4e31b9bb4006af704ba78128f88944b9bc0828f1d609a1c02e909b8c29c6a788db087d0975c34c370e4e9a8dcec0eda35f66842c7d3b332ce8865b5846a9fe7fc78b03eda127267eec81e2c275b95fa2c75676aa6cbbaa67b4556f27188ffb48a52c1868c6d0cb6004488f0108a102bc67425d345f34102818100e28bec42242c190b566086ebec215cec4d417b9f22cc019addb66e6a1f083b85b9887cb2cfd843ce8590b323470b4fc81f131122468363d503d37fc37ce7da6dda9102798ed61dcb34742d7b9891841977d5688c741298a6a3883be6a2866407b415616140ff58e3d5c4e2c15c8a2d45c647fea4d264a4e49ae7e57ee3a8341b02818100ca2ac6f1e4075a8981e82f91d0453b6cdddccb11a7ad577c09cbb14373546b36f7e6826c3ee43e9a56902dbb644b646bc2f840dc890ca2eac81bf67f7110cd76a63b681a8b79c2395dcbc1a1ef7ec94beecd6f8d30098b49dafcdb77f827a1acfa1e77861e641f9a46c630d3cb03c64f229afafcffcf44f1fcc6aa5971ce9ca1028181008c2b30d9d791a5493b7f6bdb6af5558e2b5ab9c7437b5ffed6f13a2dd4d77e24861fe9afa523d50861e19ec4d3ff2eb4ce6d38abb15f3814a35267f9a73db90b4131798b8691fa4b314034a80544fbabda562362cbaa79e298ca00edf95f176320cc1dbd53bee9dbc5f714a9b8bd11b7db2fce61627fbcfa68d1d45007419a4302818100a006a45bac88359e4afa234d6472a8cb500309aafbf33620b5104b4c7cea01c40d0ea5865172122bd1016771c1bdfbcb6115692228499c5c03f23e783a63767fc8ad95860d895fb8510a8c474670319ead74682c762dd7d7aa4424e51dc52130eefb56d90f0d6a0690a728d73d07cbddb022c531a6bbc67356075ba8597196810281810080a353e311aa5857ccb323c7ee885572fcb7de03c51357c1131eac68abedd2c404e961d0a10d3968838b4108d3bac141f383b7413fe4a08d98253a39344e598576f122d5b57e2db9823b2a3b81b5f932de718c0eed228b1f3e8f2e1e3c82793dc265f16173c4d88da6347c3c79aef43c547a4c012cca5555718ffc3d178e2718"
	pubKeyHex = "3082010a0282010100b2e8436f3ba46d9a75d39d5b130030b6c542c491378dc6082eeaf31c29962a286d871b6dd56ffd80c9cbce63e6b242cbb23f9541c1c97dcde8852221e50db3073b8da2a4ea2610fc0b1eadee23e84160f7aed850b9b2ae7116d798a2112a88c4ded9acd0c32f00baf42655910f81e83c628f41ada17bb096acf26d71e5edf646bfb8b62148a40e7ba7c94be66858468be3cfdaee0dff8327c4844a4e5d2d707b9a54defab23c058ca3482426971ae9e86d675bd9059cb0783e4577cff6f593ca4f7e862ee94a6bf74c420bfbd7132f45aac5a198f91951aa0c26928931643b8c48c4eaf7058046780283a431a30c6159761391a98563f22b233f2344854338fb0203010001"
)

// mock service.KeyManagementService implementation
type mockKeyManagementService struct {
	priKey []byte
	pubKey []byte
}

// NewMockKeyManagementService returns a new mock service.KeyManagementService implementation
func NewMockKeyManagementService(priKey, pubKey []byte) service.KeyManagementService {
	return &mockKeyManagementService{priKey, pubKey}
}

func (m *mockKeyManagementService) GetCertificate() model.EncKey {
	priKey, _ := hex.DecodeString(priKeyHex)
	return model.EncKey{
		Key:  priKey,
		Name: "testKey1",
		Algo: common.ENCRYPTION_ALGORITHM_RSA_OAEP,
	}
}

// GenerateKey generates a new mocked key
func (m *mockKeyManagementService) GenerateKey() ([]byte, error) {
	return m.priKey, nil
}

// GetPublicKey returns the mocked public key
func (m *mockKeyManagementService) GetPublicKey() ([]byte, error) {
	return m.pubKey, nil
}

// GetPrivateKey returns the mocked private key
func (m *mockKeyManagementService) GetPrivateKey() ([]byte, error) {
	return m.priKey, nil
}

// ImportKey ....
func (m *mockKeyManagementService) ImportKey(key []byte, keyName string) error {
	m.priKey = key
	return nil
}

// TestEncrypt tests the encryption of a string (using t *testing.T)
func TestEncrypt(t *testing.T) {
	// create a new mock service.KeyManagementService implementation
	priKeyB, _ := hex.DecodeString(priKeyHex)
	pubKeyB, _ := hex.DecodeString(pubKeyHex)
	mockKeyManagementService := NewMockKeyManagementService(priKeyB, pubKeyB)

	// create a new service.EncryptionService implementation
	encryptionService := service.NewCryptoServiceRSA(mockKeyManagementService)

	// encrypt a string
	testStringB := []byte("test string")
	encrypted, err := encryptionService.Encrypt(testStringB)
	if err != nil {
		t.Errorf("Error while encrypting string: %s", err)
	}

	// decrypt the string
	decrypted, err := encryptionService.Decrypt(encrypted)
	if err != nil {
		t.Errorf("Error while decrypting string: %s", err)
	}

	// check if the decrypted string is the same as the original
	if !bytes.Equal(testStringB, decrypted) {
		t.Errorf("Decrypted string is not the same as the original: %s", decrypted)
	}
}

// TestSignVerify tests the signing and verification of a string (using t *testing.T)
func TestSign(t *testing.T) {
	// create a new mock service.KeyManagementService implementation
	priKeyB, _ := hex.DecodeString(priKeyHex)
	pubKeyB, _ := hex.DecodeString(pubKeyHex)
	mockKeyManagementService := NewMockKeyManagementService(priKeyB, pubKeyB)

	// create a new service.SignatureService implementation
	signatureService := service.NewCryptoServiceRSA(mockKeyManagementService)

	// sign a string
	testStringB := []byte("test string")
	signature, err := signatureService.Sign(testStringB)
	if err != nil {
		t.Errorf("Error while signing string: %s", err)
	}

	// verify the signature
	err = signatureService.Verify(testStringB, signature)
	if err != nil {
		t.Errorf("Error while verifying signature: %s", err)
	}
	assert.Nil(t, err, "Wront signature")
}
