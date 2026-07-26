package config

import "github.com/caarlos0/env/v11"

type Config struct {
	TemplatesPath string `env:"TEMPLATES_PATH"`
	WorkspacePath string `env:"WORKSPACE_PATH"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
