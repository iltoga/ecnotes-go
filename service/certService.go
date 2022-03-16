package service

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/lib/cryptoUtil"
	"github.com/iltoga/ecnotes-go/model"
)

type CertService interface {
	CountCerts() (int, error)
	LoadCerts(pwd string) error
	SaveCerts(pwd string) error
	GetCert(name string) (*model.EncKey, error)
	AddCert(cert model.EncKey) error
	RemoveCert(name string) error
}

type CertServiceImpl struct {
	Keys         map[string]model.EncKey
	KeysMutex    *sync.Mutex
	Loaded       bool
	KeysFilePath string
}

// NewCertService creates new CertService
func NewCertService(keysFilePath string) *CertServiceImpl {
	return &CertServiceImpl{
		Keys:         make(map[string]model.EncKey),
		KeysMutex:    &sync.Mutex{},
		Loaded:       false,
		KeysFilePath: keysFilePath,
	}
}

// CountCerts returns number of certs in certificate store
func (cs *CertServiceImpl) CountCerts() (int, error) {
	if cs.KeysFilePath == "" {
		return 0, errors.New("cannot load certs file because file path variable is empty")
	}
	keysFile, err := os.OpenFile(cs.KeysFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer keysFile.Close()
	r := io.Reader(keysFile)
	j := json.NewDecoder(r)
	keys := make([]model.EncKey, 0)
	err = j.Decode(&keys)
	if err != nil {
		return 0, err
	}

	return len(keys), nil
}

// SaveCerts saves certs to file, encrypts them and writes them to file
func (cs *CertServiceImpl) SaveCerts(pwd string) error {
	if len(cs.Keys) > 0 {
		keysFile, err := os.OpenFile(cs.KeysFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		defer keysFile.Close()
		w := io.Writer(keysFile)
		j := json.NewEncoder(w)

		// save keys to file as array instead of map to preserve order
		keys := keysToArray(cs.Keys)
		// encrypt keys
		for idx, cert := range keys {
			encKey, err := cryptoUtil.EncryptMessage(cert.Key, pwd)
			if err != nil {
				return err
			}
			keys[idx].Key = encKey
		}
		return j.Encode(keys)
	}
	return nil
}

// LoadCerts loads certs from file, decrypts them and adds them to map
func (cs *CertServiceImpl) LoadCerts(pwd string) error {
	if cs.KeysFilePath == "" {
		return errors.New("cannot load certs file because file path variable is empty")
	}
	cs.KeysMutex.Lock()
	defer cs.KeysMutex.Unlock()
	keysFile, err := os.OpenFile(cs.KeysFilePath, os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer keysFile.Close()
	r := io.Reader(keysFile)
	j := json.NewDecoder(r)
	keys := make([]model.EncKey, 0)
	err = j.Decode(&keys)
	if err != nil {
		return err
	}

	// decrypt keys
	for idx, cert := range keys {
		decKey, err := cryptoUtil.DecryptMessage(cert.Key, pwd)
		if err != nil {
			return err
		}
		keys[idx].Key = decKey
	}
	cs.Keys = keysToMap(keys)
	cs.Loaded = true
	return nil
}

// GetCert returns cert by name
func (cs *CertServiceImpl) GetCert(name string) (*model.EncKey, error) {
	cs.KeysMutex.Lock()
	defer cs.KeysMutex.Unlock()
	if cs.Loaded {
		if cert, ok := cs.Keys[name]; ok {
			return &cert, nil
		}
	}
	return nil, errors.New(common.ERR_CERT_NOT_FOUND)
}

// AddCert adds cert to map
func (cs *CertServiceImpl) AddCert(cert model.EncKey) error {
	cs.KeysMutex.Lock()
	defer cs.KeysMutex.Unlock()
	if cs.Loaded {
		if _, ok := cs.Keys[cert.Name]; ok {
			return errors.New("key already exists")
		}
	}
	cs.Keys[cert.Name] = cert
	return nil
}

// RemoveCert removes cert from map
func (cs *CertServiceImpl) RemoveCert(name string) error {
	if !cs.Loaded {
		return errors.New(common.ERR_CERT_NOT_FOUND)
	}
	cs.KeysMutex.Lock()
	defer cs.KeysMutex.Unlock()
	if _, ok := cs.Keys[name]; ok {
		delete(cs.Keys, name)
		return nil
	}
	return errors.New(common.ERR_CERT_NOT_FOUND)
}

// keysToArray converts map to array
func keysToArray(keys map[string]model.EncKey) []model.EncKey {
	var keysArray []model.EncKey
	for _, key := range keys {
		keysArray = append(keysArray, key)
	}
	return keysArray
}

// keysToMap converts array to map
func keysToMap(keysArray []model.EncKey) map[string]model.EncKey {
	res := make(map[string]model.EncKey)
	for _, key := range keysArray {
		res[key.Name] = key
	}
	return res
}
