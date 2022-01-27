package service

import (
	"os"
	"path/filepath"
	"sync"

	toml "github.com/pelletier/go-toml"
)

const (
	configFileDefPath = "./resources"
)

// ConfigService ....
type ConfigService interface {
	GetGlobal(key string) (string, error)
	SetGlobal(key string, value string) error
	GetConfig(key string) (string, error)
	SetConfig(key string, value string) error
	LoadConfig() error
	ParseConfigTree(configTree *toml.Tree)
	SaveConfig() error
}

// ConfigServiceImpl ....
type ConfigServiceImpl struct {
	config     map[string]string // configuration from config file
	globals    map[string]string // global variables (loaded in memory only)
	loaded     bool
	configMux  *sync.Mutex
	globalsMux *sync.Mutex
}

// NewConfigService ....
func NewConfigService() ConfigService {
	return &ConfigServiceImpl{
		config:     make(map[string]string),
		globals:    make(map[string]string),
		loaded:     false,
		configMux:  &sync.Mutex{},
		globalsMux: &sync.Mutex{},
	}
}

// GetConfig ....
func (c *ConfigServiceImpl) GetConfig(key string) (string, error) {
	if !c.loaded {
		err := c.LoadConfig()
		if err != nil {
			return "", err
		}
	}
	return c.config[key], nil
}

// SetConfig ....
func (c *ConfigServiceImpl) SetConfig(key string, value string) error {
	if !c.loaded {
		err := c.LoadConfig()
		if err != nil {
			return err
		}
	}
	c.configMux.Lock()
	defer c.configMux.Unlock()
	c.config[key] = value
	return nil
}

// GetGlobal ....
func (c *ConfigServiceImpl) GetGlobal(key string) (string, error) {
	err := c.LoadConfig()
	if err != nil {
		return "", err
	}
	return c.globals[key], nil
}

// SetGlobal ....
func (c *ConfigServiceImpl) SetGlobal(key string, value string) error {
	c.globalsMux.Lock()
	defer c.globalsMux.Unlock()
	c.globals[key] = value
	return nil
}

// ParseConfigTree ....
func (c *ConfigServiceImpl) ParseConfigTree(configTree *toml.Tree) {
	c.configMux.Lock()
	defer c.configMux.Unlock()
	for key, value := range configTree.ToMap() {
		c.config[key] = value.(string)
	}
	c.loaded = true
}

/* LoadConfig loads the config from the config file.
 *
 * This function is thread safe.
 */
// LoadConfig ....
func (c *ConfigServiceImpl) LoadConfig() error {
	configTree, err := toml.LoadFile(filepath.Join(configFileDefPath, "config.toml"))
	if err != nil {
		return err
	}

	c.ParseConfigTree(configTree)
	return nil
}

// SaveConfig ....
func (c *ConfigServiceImpl) SaveConfig() error {
	f, err := os.Create("config.toml")
	if err != nil {
		return err
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(c.config); err != nil {
		return err
	}
	return nil
}
