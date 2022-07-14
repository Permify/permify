package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	Path        = "config/config.yaml"
	DefaultPath = "default.config.yaml"
)

type (
	// Config -.
	Config struct {
		App      `yaml:"app"`
		HTTP     `yaml:"http"`
		Log      `yaml:"logger"`
		Database `yaml:"database"`
	}

	// App -.
	App struct {
		Name    string `env-required:"true" yaml:"name"`
		Version string `env-required:"true" yaml:"version"`
	}

	// HTTP -.
	HTTP struct {
		Port string `env-required:"true" yaml:"port"`
	}

	// Log -.
	Log struct {
		Level string `env-required:"true" yaml:"log_level"`
	}

	// Database -.
	Database struct {
		Listen struct {
			Connection string `env-required:"true" yaml:"connection"`
			PoolMax    int    `env-required:"true" yaml:"pool_max"`
			URL        string `env-required:"true" yaml:"url"`
		} `env-required:"true" yaml:"listen"`

		Write struct {
			Connection string `env-required:"true" yaml:"connection"`
			PoolMax    int    `env-required:"true" yaml:"pool_max"`
			URL        string `env-required:"true" yaml:"url"`
		} `env-required:"true" yaml:"write"`
	}
)

// NewConfig returns app config.
func NewConfig() (*Config, error) {
	cfg := &Config{}

	err := cleanenv.ReadConfig("./"+Path, cfg)
	if err != nil {
		err = cleanenv.ReadConfig("./"+DefaultPath, cfg)
		if err != nil {
			return nil, fmt.Errorf("config error: %w", err)
		}
	}

	return cfg, nil
}
