package service_test

import (
	"sync"
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

func TestConfigServiceImpl_GetConfig(t *testing.T) {
	type fields struct {
		ResourcePath string
		Config       map[string]string
		Globals      map[string]string
		Loaded       bool
		ConfigMux    *sync.Mutex
		GlobalsMux   *sync.Mutex
	}
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "GetConfig_Loaded",
			fields: fields{
				Config: map[string]string{
					"test_key": "test value",
				},
				Loaded: true,
			},
			args: args{
				key: "test_key",
			},
			want:    "test value",
			wantErr: false,
		},
		{
			name: "GetConfig_InvalidKey",
			fields: fields{
				Config: map[string]string{
					"test_key": "test value",
				},
				Loaded: true,
			},
			args: args{
				key: "test_key1",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "GetConfig_Error_Loading_Config_File",
			fields: fields{
				Loaded: false,
			},
			args: args{
				key: "test_key",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &service.ConfigServiceImpl{
				ResourcePath: tt.fields.ResourcePath,
				Config:       tt.fields.Config,
				Globals:      tt.fields.Globals,
				Loaded:       tt.fields.Loaded,
				ConfigMux:    tt.fields.ConfigMux,
				GlobalsMux:   tt.fields.GlobalsMux,
			}
			got, err := c.GetConfig(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConfigServiceImpl.GetConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConfigServiceImpl.GetConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
