package config

import (
	"errors"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port    int    `env:"CPENV_PORT" envDefault:"8683"`
	HomeDir string `env:"CPENV_HOME,notEmpty"`
}

func (c *Config) Validate() error {
	var errs error
	if !(1 <= c.Port && c.Port <= 65535) {
		errs = errors.Join(errs, fmt.Errorf("CPENV_PORT %d must be between 1 and 65535", c.Port))
	}
	return errs
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}
	return &cfg, nil
}
