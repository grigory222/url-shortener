package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Storage     string `yaml:"storage"`
	Addr        string `yaml:"addr"`
	DatabaseURL string `yaml:"database_url"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file %q: %w", path, err)
	}

	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	return &cfg, nil
}

func (c *Config) validate() error {
	switch c.Storage {
	case "memory":
	case "postgres":
		if c.DatabaseURL == "" {
			return fmt.Errorf("database_url is required when storage is \"postgres\"")
		}
	case "":
		return fmt.Errorf("storage must be set (memory | postgres)")
	default:
		return fmt.Errorf("unknown storage %q: choose memory or postgres", c.Storage)
	}

	if c.Addr == "" {
		c.Addr = ":8080"
	}

	return nil
}
