package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/iltoga/ecnotes-go/lib/common"
	"github.com/iltoga/ecnotes-go/service"
	"github.com/iltoga/ecnotes-go/service/observer"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loadedConfig(values map[string]string) service.ConfigService {
	return &service.ConfigServiceImpl{
		Config:     values,
		Globals:    map[string]string{},
		Loaded:     true,
		ConfigMux:  &sync.RWMutex{},
		GlobalsMux: &sync.RWMutex{},
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("TEST_HELPER_PROCESS") != "1" {
		return
	}

	switch os.Getenv("TEST_HELPER_MODE") {
	case "setupCloseHandler":
		logPath := os.Getenv("TEST_LOG_PATH")
		logFile, err := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		logger := logrus.New()
		logger.SetOutput(logFile)
		_, cancel := context.WithCancel(context.Background())
		setupCloseHandler(cancel, logger, logFile)

		proc, err := os.FindProcess(os.Getpid())
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(3)
		}
		if err := proc.Signal(os.Interrupt); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(4)
		}
		time.Sleep(2 * time.Second)
		os.Exit(5)
	default:
		os.Exit(0)
	}
}

func TestSetupCryptoService(t *testing.T) {
	cryptoFactory, err := setupCryptoService()
	require.NoError(t, err)
	require.NotNil(t, cryptoFactory)
}

func TestSetupLogger(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "logs", "ecnotes.log")
	cfg := loadedConfig(map[string]string{
		common.CONFIG_LOG_LEVEL:     common.LOG_LEVEL_INFO,
		common.CONFIG_LOG_FILE_PATH: logPath,
	})

	logger, logFile, err := setupLogger(cfg)
	require.NoError(t, err)
	require.NotNil(t, logger)
	require.NotNil(t, logFile)
	assert.Equal(t, logrus.InfoLevel, logger.Level)

	require.NoError(t, logFile.Close())
	_, err = os.Stat(logPath)
	require.NoError(t, err)
}

func TestSetupCerts(t *testing.T) {
	cfg := loadedConfig(map[string]string{
		common.CONFIG_KEY_FILE_PATH: filepath.Join(t.TempDir(), "key_store.json"),
	})

	certService, err := setupCerts(cfg)
	require.NoError(t, err)
	require.NotNil(t, certService)
}

func TestSetupDb(t *testing.T) {
	dir := t.TempDir()
	cfg := loadedConfig(map[string]string{
		common.CONFIG_KVDB_PATH: dir,
	})
	cryptoFactory, err := setupCryptoService()
	require.NoError(t, err)

	noteService, err := setupDb(cfg, cryptoFactory, &observer.ObserverImpl{})
	require.NoError(t, err)
	require.NotNil(t, noteService)
}

func TestSetupProviders_NoGoogleSheetID(t *testing.T) {
	cfg := loadedConfig(map[string]string{})
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	err := setupProviders(context.Background(), cfg, nil, &observer.ObserverImpl{}, logger)
	require.NoError(t, err)
}

func TestSetupProviders_InvalidGoogleProviderConfig(t *testing.T) {
	cfg := loadedConfig(map[string]string{
		common.CONFIG_GOOGLE_SHEET_ID:              "sheet-id",
		common.CONFIG_GOOGLE_CREDENTIALS_FILE_PATH: filepath.Join(t.TempDir(), "missing.json"),
	})
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	err := setupProviders(context.Background(), cfg, nil, &observer.ObserverImpl{}, logger)
	require.Error(t, err)
}

func TestSetupConfigService(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(wd))
	})

	projectRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(projectRoot, "resources"), 0o755))
	require.NoError(t, os.Chdir(projectRoot))

	cfg, err := setupConfigService()
	require.NoError(t, err)
	require.NotNil(t, cfg)
}

func TestSetupCloseHandlerAndCleanup(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "cleanup.log")
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", "setupCloseHandler")
	cmd.Env = append(os.Environ(),
		"TEST_HELPER_PROCESS=1",
		"TEST_HELPER_MODE=setupCloseHandler",
		"TEST_LOG_PATH="+logPath,
	)

	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	content, err := os.ReadFile(logPath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Cleanup...")
}
