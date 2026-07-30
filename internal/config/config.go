package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port         string `env:"PORT"`
	EnvsDir      string `env:"ENVS_DIR"`
	StatePath    string `env:"STATE_PATH"`
	TemplatesDir string `env:"TEMPLATES_DIR"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	return &cfg, nil
}
