package service_test

import (
	"path/filepath"
	"sync"
	"testing"

	"github.com/iltoga/ecnotes-go/service"
	toml "github.com/pelletier/go-toml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type suite struct {
	configService service.ConfigService
	mockConfig    *toml.Tree
}

type ConfigServiceTestSuite struct {
	suite
}

func TestParseConfigTree(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	config, _ := toml.Load(`
crt_file = "./main.crt"
key_file = "./main.key"
`)

	configService, _ := service.NewConfigService()
	configService.ParseConfigTree(config)
	fileCrt, _ := configService.GetConfig("crt_file")
	assert.Equal(t, "./main.crt", fileCrt)
	fileKey, _ := configService.GetConfig("key_file")
	assert.Equal(t, "./main.key", fileKey)
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

func TestConfigService_NewConfigService_RoundTrip(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	configService, err := service.NewConfigService()
	require.NoError(t, err)

	expectedResourcePath := filepath.Join(homeDir, ".config", "ecnotes")
	assert.Equal(t, expectedResourcePath, configService.GetResourcePath())

	require.NoError(t, configService.SetConfig("custom_key", "custom_value"))
	require.NoError(t, configService.SetConfigBytes("custom_bytes", []byte{0x01, 0x02}))
	configService.SetGlobal("global_key", "global_value")
	configService.SetGlobalBytes("global_bytes", []byte{0x03, 0x04})
	require.NoError(t, configService.SaveConfig())

	gotConfig, err := configService.GetConfig("custom_key")
	require.NoError(t, err)
	assert.Equal(t, "custom_value", gotConfig)

	gotConfigBytes, err := configService.GetConfigBytes("custom_bytes")
	require.NoError(t, err)
	assert.Equal(t, []byte{0x01, 0x02}, gotConfigBytes)

	gotGlobal, err := configService.GetGlobal("global_key")
	require.NoError(t, err)
	assert.Equal(t, "global_value", gotGlobal)

	gotGlobalBytes, err := configService.GetGlobalBytes("global_bytes")
	require.NoError(t, err)
	assert.Equal(t, []byte{0x03, 0x04}, gotGlobalBytes)

	parsed, err := toml.Load(`parsed_key = "parsed_value"`)
	require.NoError(t, err)
	configService.ParseConfigTree(parsed)

	gotParsed, err := configService.GetConfig("parsed_key")
	require.NoError(t, err)
	assert.Equal(t, "parsed_value", gotParsed)

	require.NoError(t, configService.LoadConfig())
}
