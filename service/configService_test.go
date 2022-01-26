package service_test

import (
	"testing"

	"github.com/iltoga/ecnotes-go/service"
	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
)

type suite struct {
	configService service.ConfigService
	mockConfig    *toml.Tree
}

type ConfigServiceTestSuite struct {
	suite
}

func (s *ConfigServiceTestSuite) TestParseConfigTree(t *testing.T) {
	config, _ := toml.Load(`
[certs]
crt_file = "./main.crt"
key_file = "./main.key"
`)

	s.mockConfig = config
	s.configService = service.NewConfigService()
	s.configService.ParseConfigTree(config)
	fileCrt, _ := s.configService.GetConfig("crt_file")
	assert.Equal(t, "./main.crt", fileCrt)
	fileKey, _ := s.configService.GetConfig("key_file")
	assert.Equal(t, "./main.key", fileKey)
}
