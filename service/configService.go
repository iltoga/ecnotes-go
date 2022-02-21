package service

import (
	"errors"
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
	SetGlobal(key string, value string)
	GetConfig(key string) (string, error)
	SetConfig(key string, value string) error
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
func NewConfigService() ConfigService {
	srv := &ConfigServiceImpl{
		ResourcePath: common.DEFAULT_RESOURCE_PATH,
		Config:       make(map[string]string),
		Globals:      make(map[string]string),
		Loaded:       false,
		ConfigMux:    &sync.RWMutex{},
		GlobalsMux:   &sync.RWMutex{},
	}
	srv.init()
	return srv
}

// setDefaultConfig set some default values in the config file if not exists
func (c *ConfigServiceImpl) setDefaultConfig() {
	if _, ok := c.Config[common.CONFIG_KVDB_PATH]; !ok {
		c.Config[common.CONFIG_KVDB_PATH] = filepath.Join(c.ResourcePath, common.DEFAULT_DB_PATH)
	}
}

func (c *ConfigServiceImpl) init() {
	// check if the default resource path exists
	if _, err := os.Stat(c.ResourcePath); os.IsNotExist(err) {
		// if not, try to get the user's home directory
		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Could not get user's home directory")
		}
		// create the default resource path if not exists
		resourcePath := filepath.Join(homeDir, ".config", "ecnotes")
		if err := os.MkdirAll(resourcePath, 0755); err != nil {
			log.Fatalf("Failed to create resource path: %s", resourcePath)
		}
		c.ResourcePath = resourcePath
	}

	// create the default config file if not exists
	configFilePath := c.getConfigFilePath()
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		f, err := os.Create(configFilePath)
		if err != nil {
			log.Fatalf("Failed to create config file: %s", configFilePath)
		}
		defer f.Close()
		// set some default values
		c.setDefaultConfig()
		if err := toml.NewEncoder(f).Encode(c.Config); err != nil {
			log.Fatalf("Failed to write default config file: %s", configFilePath)
		}
	}
}

// GetResourcePath  ....
func (c *ConfigServiceImpl) GetResourcePath() string {
	return c.ResourcePath
}

// GetConfig ....
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

// GetGlobal ....
func (c *ConfigServiceImpl) GetGlobal(key string) (string, error) {
	c.GlobalsMux.RLock()
	defer c.GlobalsMux.RUnlock()
	if val, ok := c.Globals[key]; ok {
		return val, nil
	}
	return "", errors.New("key not found")
}

// SetGlobal ....
func (c *ConfigServiceImpl) SetGlobal(key string, value string) {
	c.GlobalsMux.Lock()
	defer c.GlobalsMux.Unlock()
	c.Globals[key] = value
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
