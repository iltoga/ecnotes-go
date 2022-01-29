package service

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	toml "github.com/pelletier/go-toml"
)

const (
	defaultResourcePath = "./resources"
)

// ConfigService ....
type ConfigService interface {
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
	return &ConfigServiceImpl{
		ResourcePath: defaultResourcePath,
		Config:       make(map[string]string),
		Globals:      make(map[string]string),
		Loaded:       false,
		ConfigMux:    &sync.RWMutex{},
		GlobalsMux:   &sync.RWMutex{},
	}
}

// GetConfig ....
func (c *ConfigServiceImpl) GetConfig(key string) (string, error) {
	if !c.Loaded {
		err := c.LoadConfig()
		if err != nil {
			return "", err
		}
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
	if !c.Loaded {
		err := c.LoadConfig()
		if err != nil {
			return err
		}
	}
	c.ConfigMux.Lock()
	defer c.ConfigMux.Unlock()
	c.Config[key] = value
	return nil
}

// GetGlobal ....
func (c *ConfigServiceImpl) GetGlobal(key string) (string, error) {
	err := c.LoadConfig()
	if err != nil {
		return "", err
	}
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
	if err := toml.NewEncoder(f).Encode(c.Config); err != nil {
		return err
	}
	return nil
}

func (c *ConfigServiceImpl) getConfigFilePath() string {
	return filepath.Join(c.ResourcePath, "config.toml")
}
