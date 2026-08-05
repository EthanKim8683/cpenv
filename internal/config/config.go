package config

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port string `env:"CPENV_PORT" envDefault:"8683"`
}

func (c *Config) Validate() error {
	var errs error
	if port, err := strconv.Atoi(c.Port); err != nil || !(1 <= port && port <= 65535) {
		errs = errors.Join(errs, fmt.Errorf("CPENV_PORT %q must be an integer between 1 and 65535", c.Port))
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
