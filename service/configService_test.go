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

func (s *ConfigServiceTestSuite) TestParseConfigTree(t *assert.TestingT) {
	config, _ := toml.Load(`
crt_file = "./main.crt"
key_file = "./main.key"
`)

	s.mockConfig = config
	s.configService, _ = service.NewConfigService()
	s.configService.ParseConfigTree(config)
	fileCrt, _ := s.configService.GetConfig("crt_file")
	assert.Equal(*t, "./main.crt", fileCrt)
	fileKey, _ := s.configService.GetConfig("key_file")
	assert.Equal(*t, "./main.key", fileKey)
}

func (s *ConfigServiceTestSuite) TestConfigServiceImpl_GetConfig(t *testing.T) {
	type fields struct {
		ResourcePath string
		Config       map[string]string
		Globals      map[string]string
		Loaded       bool
		ConfigMux    *sync.RWMutex
		GlobalsMux   *sync.RWMutex
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
				ConfigMux:  &sync.RWMutex{},
				GlobalsMux: &sync.RWMutex{},
				Loaded:     true,
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
				ConfigMux:  &sync.RWMutex{},
				GlobalsMux: &sync.RWMutex{},
				Loaded:     true,
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
				Loaded:     false,
				ConfigMux:  &sync.RWMutex{},
				GlobalsMux: &sync.RWMutex{},
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

// TestGetConfig tests the GetConfig method of the ConfigServiceImpl type.
func TestGetConfig(t *testing.T) {
	// mock some Config map and Globals map
	configMap := map[string]string{
		"test_key":                "test value",
		"test_encoded_key":        "4C6F72656D20697073756D20646F6C6F722073697420616D65742C20636F6E73656374657475722061646970697363696E6720656C69742E204E756E632070656C6C656E746573717565206D617373612076656E656E61746973206C6563747573206D6F6C6C69732C20696E2074656D706F72206E65717565206672696E67696C6C612E",
		"test_invalid_hex_string": "4C6F72656D20697073756D20646F6C6F722073697420616D65742C20636F6E736563746574757220616469706",
	}
	globalsMap := configMap

	type fields struct {
		ResourcePath string
		Config       map[string]string
		Globals      map[string]string
		Loaded       bool
		ConfigMux    *sync.RWMutex
		GlobalsMux   *sync.RWMutex
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
			name: "GetConfig_clear_text_key",
			fields: fields{
				Config:     configMap,
				Globals:    globalsMap,
				ConfigMux:  &sync.RWMutex{},
				GlobalsMux: &sync.RWMutex{},
				Loaded:     true,
			},
			args: args{
				key: "test_key",
			},
			want:    "test value",
			wantErr: false,
		},
		{
			name: "GetConfig_encoded_key",
			fields: fields{
				Config:     configMap,
				Globals:    globalsMap,
				ConfigMux:  &sync.RWMutex{},
				GlobalsMux: &sync.RWMutex{},
				Loaded:     true,
			},
			args: args{
				key: "test_encoded_key",
			},
			want:    "4C6F72656D20697073756D20646F6C6F722073697420616D65742C20636F6E73656374657475722061646970697363696E6720656C69742E204E756E632070656C6C656E746573717565206D617373612076656E656E61746973206C6563747573206D6F6C6C69732C20696E2074656D706F72206E65717565206672696E67696C6C612E",
			wantErr: false,
		},
		{
			name: "GetConfig_invalid_hex_string",
			fields: fields{
				Config:     configMap,
				Globals:    globalsMap,
				ConfigMux:  &sync.RWMutex{},
				GlobalsMux: &sync.RWMutex{},
				Loaded:     true,
			},
			args: args{
				key: "test_invalid_hex_string",
			},
			want:    "4C6F72656D20697073756D20646F6C6F722073697420616D65742C20636F6E736563746574757220616469706",
			wantErr: false,
		},
		{
			name: "GetConfig_Error_Loading_Config_File",
			fields: fields{
				Loaded:     false,
				ConfigMux:  &sync.RWMutex{},
				GlobalsMux: &sync.RWMutex{},
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

// TestGetGlobal tests the GetGlobal method of the ConfigServiceImpl type.
func TestGetGlobal(t *testing.T) {
	// mock some Config map and Globals map
	configMap := map[string]string{
		"test_key":                "test value",
		"test_encoded_key":        "4C6F72656D20697073756D20646F6C6F722073697420616D65742C20636F6E73656374657475722061646970697363696E6720656C69742E204E756E632070656C6C656E746573717565206D617373612076656E656E61746973206C6563747573206D6F6C6C69732C20696E2074656D706F72206E65717565206672696E67696C6C612E",
		"test_invalid_hex_string": "4C6F72656D20697073756D20646F6C6F722073697420616D65742C20636F6E736563746574757220616469706",
	}
	globalsMap := configMap

	type fields struct {
		ResourcePath string
		Config       map[string]string
		Globals      map[string]string
		Loaded       bool
		ConfigMux    *sync.RWMutex
		GlobalsMux   *sync.RWMutex
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
			name: "GetConfig_clear_text_key",
			fields: fields{
				Config:     configMap,
				Globals:    globalsMap,
				ConfigMux:  &sync.RWMutex{},
				GlobalsMux: &sync.RWMutex{},
				Loaded:     true,
			},
			args: args{
				key: "test_key",
			},
			want:    "test value",
			wantErr: false,
		},
		{
			name: "GetConfig_encoded_key",
			fields: fields{
				Config:     configMap,
				Globals:    globalsMap,
				ConfigMux:  &sync.RWMutex{},
				GlobalsMux: &sync.RWMutex{},
				Loaded:     true,
			},
			args: args{
				key: "test_encoded_key",
			},
			want:    "4C6F72656D20697073756D20646F6C6F722073697420616D65742C20636F6E73656374657475722061646970697363696E6720656C69742E204E756E632070656C6C656E746573717565206D617373612076656E656E61746973206C6563747573206D6F6C6C69732C20696E2074656D706F72206E65717565206672696E67696C6C612E",
			wantErr: false,
		},
		{
			name: "GetConfig_invalid_hex_string",
			fields: fields{
				Config:     configMap,
				Globals:    globalsMap,
				ConfigMux:  &sync.RWMutex{},
				GlobalsMux: &sync.RWMutex{},
				Loaded:     true,
			},
			args: args{
				key: "test_invalid_hex_string",
			},
			want:    "4C6F72656D20697073756D20646F6C6F722073697420616D65742C20636F6E736563746574757220616469706",
			wantErr: false,
		},
		{
			name: "GetConfig_Error_Loading_Config_File",
			fields: fields{
				Loaded:     false,
				ConfigMux:  &sync.RWMutex{},
				GlobalsMux: &sync.RWMutex{},
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
			got, err := c.GetGlobal(tt.args.key)
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
