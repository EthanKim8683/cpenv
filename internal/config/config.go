package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port          string `env:"CPENV_PORT" envDefault:"8080"`
	StatePath     string `env:"CPENV_STATE_PATH,notEmpty"`
	TemplatesDir  string `env:"CPENV_TEMPLATES_DIR,notEmpty"`
	WorkspacesDir string `env:"CPENV_WORKSPACES_DIR,notEmpty"`
}

func (c *Config) Validate() error {
	var errs error
	if port, err := strconv.Atoi(c.Port); err != nil || !(1 <= port && port <= 65535) {
		errs = errors.Join(errs, fmt.Errorf("CPENV_PORT %q must be an integer between 1 and 65535", c.Port))
	}
	if !filepath.IsAbs(c.StatePath) {
		errs = errors.Join(errs, fmt.Errorf("CPENV_STATE_PATH %q must be an absolute path", c.StatePath))
	}
	if !filepath.IsAbs(c.TemplatesDir) {
		errs = errors.Join(errs, fmt.Errorf("CPENV_TEMPLATES_DIR %q must be an absolute path", c.TemplatesDir))
	}
	if !filepath.IsAbs(c.WorkspacesDir) {
		errs = errors.Join(errs, fmt.Errorf("CPENV_WORKSPACES_DIR %q must be an absolute path", c.WorkspacesDir))
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
