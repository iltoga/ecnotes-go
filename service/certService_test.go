package service_test

import (
	"path/filepath"
	"testing"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/model"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCertService_RoundTripAndMutations(t *testing.T) {
	dir := t.TempDir()
	keyFile := filepath.Join(dir, "key_store.json")

	certService := service.NewCertService(keyFile)
	alphaKey := []byte("alpha-key-32-bytes-alpha-key-32-by")
	betaKey := []byte("beta-key-32-bytes-beta-key-32-byt")

	require.NoError(t, certService.AddCert(model.EncKey{
		Name: "alpha",
		Algo: common.ENCRYPTION_ALGORITHM_AES_256_CBC,
		Key:  alphaKey,
	}))
	require.NoError(t, certService.AddCert(model.EncKey{
		Name: "beta",
		Algo: common.ENCRYPTION_ALGORITHM_AES_256_CBC,
		Key:  betaKey,
	}))
	require.NoError(t, certService.SaveCerts("secret"))

	count, err := certService.CountCerts()
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	loaded := service.NewCertService(keyFile)
	require.NoError(t, loaded.LoadCerts("secret"))

	alphaCert, err := loaded.GetCert("alpha")
	require.NoError(t, err)
	assert.Equal(t, alphaKey, []byte(alphaCert.Key))

	betaCert, err := loaded.GetCert("beta")
	require.NoError(t, err)
	assert.Equal(t, betaKey, []byte(betaCert.Key))

	require.NoError(t, loaded.AddCert(model.EncKey{
		Name: "gamma",
		Algo: common.ENCRYPTION_ALGORITHM_AES_256_CBC,
		Key:  []byte("gamma-key-32-bytes-gamma-key-32-"),
	}))

	_, err = loaded.GetCert("gamma")
	require.NoError(t, err)

	require.NoError(t, loaded.RemoveCert("alpha"))
	_, err = loaded.GetCert("alpha")
	assert.Error(t, err)
}

