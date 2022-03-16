package certservice

import (
	"encoding/hex"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type certServiceTest struct {
	suite.Suite
}

// SetupTest ....
func (s *certServiceTest) SetupTest() {
	fmt.Println("SetupTest...")
}

// TearDownTest ....
func (s *certServiceTest) TearDownTest() {
	fmt.Println("TearDownTest...")
}

// TestSuite ....
func TestSuite(t *testing.T) {
	suite.Run(t, new(certServiceTest))
}

func (s *certServiceTest) TestCertServiceImpl_SaveAndLoadCerts() {
	type fields struct {
		keys         map[string]model.EncKey
		keysMutex    *sync.Mutex
		loaded       bool
		keysFilePath string
	}
	testKey1, _ := hex.DecodeString(
		"d9dc8e1ae286be3d1600c642c1ee203d183e9905a5edf30f7ffd1a118c16bd64ecb7dc416c2c8bf9d3650f8be7d39794bcad97408d0a98ec01584e69160e919740f5f1deac8f90484494589257f0780a404131176d7510ae97cab4e1d94b7608c13bc5945cd1007decd35a03123beb952d0105f12a6ae49efa0417a9a96c9586f2f64cc83c2f118fb7ff1b9f027f4335494c1b34f9bc5c8362c019c795a9fa7cbfe05a29ec7044781fa1da2fddaa459d1171a10f24bbb85dbd47f68c981b78ae87fd1767b82a3acd67f39a5660e88996349ccb1a503a93d2c6fb8b2071684cc44ee10fa6e30741aba967e93349b6a8dee0b19e7c3c3a2225437ca05eb81b86c470c874475f2f26806811b94330d5461b7a6022a7e3540e55ef1777b6",
	)
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Save_Ok",
			fields: fields{
				loaded:       false,
				keysFilePath: "./testKeys.json",
				keys: map[string]model.EncKey{
					"testKey1": {
						Name: "testKey1",
						Algo: common.ENCRYPTION_ALGORITHM_AES_256_CBC,
						Key:  testKey1,
					},
				},
				keysMutex: &sync.Mutex{},
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			cs := &service.CertServiceImpl{
				Keys:         tt.fields.keys,
				KeysMutex:    tt.fields.keysMutex,
				Loaded:       tt.fields.loaded,
				KeysFilePath: tt.fields.keysFilePath,
			}
			if err := cs.SaveCerts("1234"); (err != nil) != tt.wantErr {
				t.Errorf("CertServiceImpl.SaveCerts() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err := cs.LoadCerts("1234"); err != nil {
				t.Errorf("CertServiceImpl.LoadCerts() error = %v", err)
			} else {
				assert.NotNil(t, cs.Keys["testKey1"])
				fmt.Println(cs.Keys["testKey1"].Name)
				fmt.Println(cs.Keys["testKey1"].Algo)
				assert.Equal(t, "testKey1", cs.Keys["testKey1"].Name, "Test key different from what expected")
			}

			// delete file generated by this test
			if err := os.Remove(cs.KeysFilePath); err != nil {
				t.Error("Cannot clean test environment")
			}
		})
	}
}