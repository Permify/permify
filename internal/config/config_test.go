package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig_FileNotFound(t *testing.T) {
	// Test the NewConfig function
	cfg, err := NewConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check if default values are applied correctly
	assert.Equal(t, "3476", cfg.Server.HTTP.Port)
	assert.Equal(t, "3478", cfg.Server.GRPC.Port)
	assert.Equal(t, "info", cfg.Log.Level)
}

func TestNewConfig(t *testing.T) {
	configContent := []byte(`
server:
  http:
    enabled: true
    port: "8080"
  grpc:
    port: "9090"
logger:
  level: "debug"
`)

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "new-config-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir) // Clean up after the test

	// Create a temporary config file
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(tmpFile, configContent, 0o666)
	assert.NoError(t, err)

	// Set the config path in viper to the temporary directory
	viper.AddConfigPath(tmpDir)

	// Test the NewConfig function
	cfg, err := NewConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check if values from the sample config file are loaded correctly
	assert.Equal(t, "8080", cfg.Server.HTTP.Port)
	assert.Equal(t, "9090", cfg.Server.GRPC.Port)
	assert.Equal(t, "debug", cfg.Log.Level)

	// Check if default values are applied correctly
	assert.False(t, cfg.Server.HTTP.TLSConfig.Enabled)
	assert.False(t, cfg.Server.GRPC.TLSConfig.Enabled)
	assert.False(t, cfg.Profiler.Enabled)
	assert.False(t, cfg.Tracer.Enabled)
	assert.Equal(t, "otlp", cfg.Meter.Exporter)
}

func TestNewConfig_InvalidConfig(t *testing.T) {
	configContent := []byte(`
invalid_config_content
`)

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "invalid-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) // Clean up after the test

	// Create a temporary config file
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(tmpFile, configContent, 0o666)
	require.NoError(t, err)

	// Set the config path in viper to the temporary directory
	viper.AddConfigPath(tmpDir)

	// Test the NewConfig function
	cfg, err := NewConfig()
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestNewConfig_PartialConfig(t *testing.T) {
	configContent := []byte(`
server:
  http:
    port: "8081"
tracer:
  enabled: true
  exporter: "jaeger"
  endpoint: "http://localhost:14268/api/traces"
database:
  engine: "postgres"
  uri: "postgres://user:password@localhost/dbname"
`)

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "partial-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) // Clean up after the test

	// Create a temporary config file
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(tmpFile, configContent, 0o666)
	require.NoError(t, err)

	// Set the config path in viper to the temporary directory
	viper.AddConfigPath(tmpDir)

	// Test the NewConfig function
	cfg, err := NewConfig()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// Check if values from the sample config file are loaded correctly
	assert.Equal(t, "8081", cfg.Server.HTTP.Port)
	assert.True(t, cfg.Tracer.Enabled)
	assert.Equal(t, "jaeger", cfg.Tracer.Exporter)
	assert.Equal(t, "http://localhost:14268/api/traces", cfg.Tracer.Endpoint)
	assert.Equal(t, "postgres", cfg.Database.Engine)
	assert.Equal(t, "postgres://user:password@localhost/dbname", cfg.Database.URI)

	// Check if default values are applied correctly
	assert.Equal(t, "3478", cfg.Server.GRPC.Port)
	assert.False(t, cfg.Server.HTTP.TLSConfig.Enabled)
	assert.False(t, cfg.Server.GRPC.TLSConfig.Enabled)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.Equal(t, "otlp", cfg.Meter.Exporter)
	assert.Equal(t, "telemetry.permify.co", cfg.Meter.Endpoint)
}

func Test_isYAML(t *testing.T) {
	type args struct {
		file string
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "valid yaml file",
			args: args{
				file: "testdata/valid.yaml",
			},
			wantErr: assert.NoError,
		},
		{
			name: "invalid yaml file",
			args: args{
				file: "testdata/invalid.json",
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, isYAML(tt.args.file), fmt.Sprintf("isYAML(%v)", tt.args.file))
		})
	}
}

func TestNewConfigWithFile(t *testing.T) {
	configContent := []byte(`
server:
  http:
    enabled: true
    port: "8080"
  grpc:
    port: "9090"
logger:
  level: "debug"
`)

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "new-config-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir) // Clean up after the test

	// Create a temporary config file
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(tmpFile, configContent, 0o666)
	assert.NoError(t, err)

	cfg, err := NewConfigWithFile(tmpFile)
	assert.NotNil(t, cfg)
	assert.NoError(t, err)

	assert.Equal(t, true, cfg.Server.Enabled)
	assert.Equal(t, "8080", cfg.Server.HTTP.Port)
	assert.Equal(t, "9090", cfg.GRPC.Port)
	assert.Equal(t, "debug", cfg.Log.Level)
}

func TestNewConfigWithFile_InvalidConfig(t *testing.T) {
	configContent := []byte(`
invalid config
`)

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "new-config-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir) // Clean up after the test

	// Create a temporary config file
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	err = os.WriteFile(tmpFile, configContent, 0o666)
	assert.NoError(t, err)

	cfg, err := NewConfigWithFile(tmpFile)
	assert.Nil(t, cfg)
	assert.Error(t, err)
}
