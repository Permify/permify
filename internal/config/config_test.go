package config

import (
	"io/ioutil"
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
	tmpDir, err := ioutil.TempDir("", "new-config-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir) // Clean up after the test

	// Create a temporary config file
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	err = ioutil.WriteFile(tmpFile, configContent, 0o666)
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
	tmpDir, err := ioutil.TempDir("", "invalid-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) // Clean up after the test

	// Create a temporary config file
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	err = ioutil.WriteFile(tmpFile, configContent, 0o666)
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
	tmpDir, err := ioutil.TempDir("", "partial-config-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) // Clean up after the test

	// Create a temporary config file
	tmpFile := filepath.Join(tmpDir, "config.yaml")
	err = ioutil.WriteFile(tmpFile, configContent, 0o666)
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
