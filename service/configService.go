package service

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/iltoga/ecnotes-go/lib/common"
	toml "github.com/pelletier/go-toml"
)

// ConfigService ....
type ConfigService interface {
	GetResourcePath() string
	GetGlobal(key string) (string, error)
	GetGlobalBytes(key string) ([]byte, error)
	SetGlobal(key string, value string)
	SetGlobalBytes(key string, value []byte)
	GetConfig(key string) (string, error)
	GetConfigBytes(key string) ([]byte, error)
	SetConfig(key string, value string) error
	SetConfigBytes(key string, value []byte) error
	LoadConfig() error
	ParseConfigTree(configTree *toml.Tree)
	SaveConfig() error
}

// ConfigServiceImpl ....
type ConfigServiceImpl struct {
	ResourcePath string
	Config       map[string]string // configuration from config file
	Globals      map[string]string // global variables (loaded in memory only)
	Loaded       bool
	ConfigMux    *sync.RWMutex
	GlobalsMux   *sync.RWMutex
}

// NewConfigService ....
func NewConfigService() (ConfigService, error) {
	srv := &ConfigServiceImpl{
		ResourcePath: common.DEFAULT_RESOURCE_PATH,
		Config:       make(map[string]string),
		Globals:      make(map[string]string),
		Loaded:       false,
		ConfigMux:    &sync.RWMutex{},
		GlobalsMux:   &sync.RWMutex{},
	}
	err := srv.init()
	return srv, err
}

// setDefaultConfig set some default values in the config map if they are not set yet
func (c *ConfigServiceImpl) setDefaultConfig() {
	// set default config for db path
	if _, ok := c.Config[common.CONFIG_KVDB_PATH]; !ok {
		c.Config[common.CONFIG_KVDB_PATH] = filepath.Join(c.ResourcePath, common.DEFAULT_DB_PATH)
	}
	// set default config for log file path and level
	if _, ok := c.Config[common.CONFIG_LOG_FILE_PATH]; !ok {
		c.Config[common.CONFIG_LOG_FILE_PATH] = filepath.Join(c.ResourcePath, common.DEFAULT_LOG_FILE_PATH)
	}
	// set default config for key_file_path
	if _, ok := c.Config[common.CONFIG_KEY_FILE_PATH]; !ok {
		c.Config[common.CONFIG_KEY_FILE_PATH] = filepath.Join(c.ResourcePath, common.DEFAULT_KEY_FILE_PATH)
	}
	if _, ok := c.Config[common.CONFIG_LOG_LEVEL]; !ok {
		c.Config[common.CONFIG_LOG_LEVEL] = common.DEFAULT_LOG_LEVEL
	}
	// STEF delete this
	// // set default config for encryption algorithm
	// if _, ok := c.Config[common.CONFIG_ENCRYPTION_ALGORITHM]; !ok {
	// 	c.Config[common.CONFIG_ENCRYPTION_ALGORITHM] = common.DEFAULT_ENCRYPTION_ALGORITHM
	// }

	// set default config for google credentials file path (defaults to user home directory .config/ecnotes)
	// note: the directory will be automatically created, but you must manually copy the file inside it
	// in order to use google sheets service (see google-sheets-service.go)
	if _, ok := c.Config[common.CONFIG_GOOGLE_PROVIDER_PATH]; !ok {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Could not get user's home directory")
		}
		googleProviderPath := filepath.Join(
			homeDir,
			".config",
			"ecnotes",
			"providers",
			"google",
		)
		c.Config[common.CONFIG_GOOGLE_PROVIDER_PATH] = googleProviderPath
		// create the directory if not exists
		if _, err := os.Stat(googleProviderPath); os.IsNotExist(err) {
			if err := os.MkdirAll(googleProviderPath, 0755); err != nil {
				log.Fatalf("Failed to create directory: %s", filepath.Dir(googleProviderPath))
			}
		}
		// set default credentials file
		c.Config[common.CONFIG_GOOGLE_CREDENTIALS_FILE_PATH] = filepath.Join(
			googleProviderPath,
			"cred_serviceaccount.json",
		)
	}
}

func (c *ConfigServiceImpl) init() error {
	// check if the default resource path exists
	if _, err := os.Stat(c.ResourcePath); os.IsNotExist(err) {
		// if not, try to get the user's home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		// create the default resource path if not exists
		resourcePath := filepath.Join(homeDir, ".config", "ecnotes")
		if err := os.MkdirAll(resourcePath, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %s", resourcePath)
		}
		c.ResourcePath = resourcePath
	}

	// create the default config file if not exists, otherwise update it if needed (if the config file is outdated)
	configFilePath := c.getConfigFilePath()
	_, err := os.Stat(configFilePath)
	if os.IsNotExist(err) {
		// create the default config file
		f, err := os.Create(configFilePath)
		if err != nil {
			return fmt.Errorf("failed to create config file: %s", configFilePath)
		}
		f.Close()
	} else {
		// update the config file in case we updated the default configuration with new values
		if err := c.LoadConfig(); err != nil {
			return fmt.Errorf("failed to load config file: %s", configFilePath)
		}
	}
	// set some default values (if not set yet)
	c.setDefaultConfig()
	if err := c.SaveConfig(); err != nil {
		return fmt.Errorf("failed to save config file: %s", configFilePath)
	}
	return nil
}

// GetResourcePath  ....
func (c *ConfigServiceImpl) GetResourcePath() string {
	return c.ResourcePath
}

// GetConfig return the value of the given key from the config map
func (c *ConfigServiceImpl) GetConfig(key string) (string, error) {
	if err := c.checkAndLoad(); err != nil {
		return "", err
	}
	c.ConfigMux.RLock()
	defer c.ConfigMux.RUnlock()
	if val, ok := c.Config[key]; ok {
		return val, nil
	}
	return "", errors.New("key not found")
}

// GetGlobalBytes same as above, for byte arrays
func (c *ConfigServiceImpl) GetConfigBytes(key string) ([]byte, error) {
	if err := c.checkAndLoad(); err != nil {
		return nil, err
	}
	c.ConfigMux.RLock()
	defer c.ConfigMux.RUnlock()
	if val, ok := c.Config[key]; ok {
		if valHex, err := hex.DecodeString(val); err == nil {
			return valHex, nil
		}
		return nil, errors.New(common.ERR_INVALID_HEX_STRING)
	}
	return nil, errors.New("key not found")
}

// SetConfig ....
func (c *ConfigServiceImpl) SetConfig(key string, value string) error {
	if err := c.checkAndLoad(); err != nil {
		return err
	}
	c.ConfigMux.Lock()
	defer c.ConfigMux.Unlock()
	c.Config[key] = value
	return nil
}

// SetConfigBytes same as above, for byte arrays
// note: the byte array is encoded as hexadecimal string
func (c *ConfigServiceImpl) SetConfigBytes(key string, value []byte) error {
	if err := c.checkAndLoad(); err != nil {
		return err
	}
	c.ConfigMux.Lock()
	defer c.ConfigMux.Unlock()
	c.Config[key] = hex.EncodeToString(value)
	return nil
}

// GetGlobal ....
func (c *ConfigServiceImpl) GetGlobal(key string) (string, error) {
	c.GlobalsMux.RLock()
	defer c.GlobalsMux.RUnlock()
	if val, ok := c.Globals[key]; ok {
		return val, nil
	}
	return "", errors.New("key not found")
}

// GetGlobalBytes same as above, for byte arrays
func (c *ConfigServiceImpl) GetGlobalBytes(key string) ([]byte, error) {
	c.GlobalsMux.RLock()
	defer c.GlobalsMux.RUnlock()
	if val, ok := c.Globals[key]; ok {
		if valHex, err := hex.DecodeString(val); err == nil {
			return valHex, nil
		}
		return nil, errors.New(common.ERR_INVALID_HEX_STRING)
	}
	return nil, errors.New("key not found")
}

// SetGlobal ....
func (c *ConfigServiceImpl) SetGlobal(key string, value string) {
	c.GlobalsMux.Lock()
	defer c.GlobalsMux.Unlock()
	c.Globals[key] = value
}

// SetGlobalBytes same as above, for byte arrays
func (c *ConfigServiceImpl) SetGlobalBytes(key string, value []byte) {
	c.GlobalsMux.Lock()
	defer c.GlobalsMux.Unlock()
	c.Globals[key] = hex.EncodeToString(value)
}

// ParseConfigTree ....
func (c *ConfigServiceImpl) ParseConfigTree(configTree *toml.Tree) {
	c.ConfigMux.Lock()
	defer c.ConfigMux.Unlock()
	for key, value := range configTree.ToMap() {
		c.Config[key] = value.(string)
	}
	c.Loaded = true
}

/* LoadConfig loads the config from the config file.
 *
 * This function is thread safe.
 */
// LoadConfig ....
func (c *ConfigServiceImpl) LoadConfig() error {
	configTree, err := toml.LoadFile(c.getConfigFilePath())
	if err != nil {
		return err
	}
	c.ParseConfigTree(configTree)
	c.Loaded = true
	return nil
}

// SaveConfig ....
func (c *ConfigServiceImpl) SaveConfig() error {
	f, err := os.Create(c.getConfigFilePath())
	if err != nil {
		return err
	}
	defer f.Close()
	c.ConfigMux.RLock()
	defer c.ConfigMux.RUnlock()
	return toml.NewEncoder(f).Encode(c.Config)
}

func (c *ConfigServiceImpl) getConfigFilePath() string {
	return filepath.Join(c.ResourcePath, "config.toml")
}

func (c *ConfigServiceImpl) checkAndLoad() error {
	if !c.Loaded {
		err := c.LoadConfig()
		if err != nil {
			return err
		}
	}
	return nil
}
