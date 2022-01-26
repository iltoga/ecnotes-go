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

type ConfigService interface {
	GetConfig(key string) (string, error)
	SetConfig(key string, value string) error
	LoadConfig() error
	ParseConfigTree(configTree *toml.Tree)
	SaveConfig() error
}

// ConfigServiceImpl ....
type ConfigServiceImpl struct {
	config map[string]string
	loaded bool
	mutex  *sync.Mutex
}

// NewConfigService ....
func NewConfigService() ConfigService {
	return &ConfigServiceImpl{
		config: make(map[string]string),
		loaded: false,
		mutex:  &sync.Mutex{},
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
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.config[key] = value
	return nil
}

// ParseConfigTree ....
func (c *ConfigServiceImpl) ParseConfigTree(configTree *toml.Tree) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
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
