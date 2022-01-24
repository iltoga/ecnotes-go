package service

import (
	"os"
	"path/filepath"
	"sync"

	toml "github.com/pelletier/go-toml"
)

const (
	CONFIG_FILE_DEF_PATH = "./resources"
)

type ConfigService interface {
	GetConfig(key string) (string, error)
	SetConfig(key string, value string) error
	LoadConfig() error
	ParseConfigTree(configTree *toml.Tree)
	SaveConfig() error
}

type ConfigServiceImpl struct {
	config map[string]string
	loaded bool
	mutex  *sync.Mutex
}

func NewConfigService() ConfigService {
	return &ConfigServiceImpl{
		config: make(map[string]string),
		loaded: false,
		mutex:  &sync.Mutex{},
	}
}

func (c *ConfigServiceImpl) GetConfig(key string) (string, error) {
	if !c.loaded {
		err := c.LoadConfig()
		if err != nil {
			return "", err
		}
	}
	return c.config[key], nil
}

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
func (c *ConfigServiceImpl) LoadConfig() error {
	var (
		configTree, err = toml.LoadFile(filepath.Join(CONFIG_FILE_DEF_PATH, "config.toml"))
	)
	if err != nil {
		return err
	}

	c.ParseConfigTree(configTree)
	return nil
}

func (c *ConfigServiceImpl) SaveConfig() error {
	if f, err := os.Create("config.toml"); err != nil {
		return err
	} else {
		defer f.Close()
		if err := toml.NewEncoder(f).Encode(c.config); err != nil {
			return err
		}
	}
	return nil
}
