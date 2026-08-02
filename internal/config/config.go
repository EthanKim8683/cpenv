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
	Home          string `env:"CPENV_HOME"`
	StatePath     string `env:"CPENV_STATE_PATH"`
	TemplatesDir  string `env:"CPENV_TEMPLATES_DIR"`
	WorkspacesDir string `env:"CPENV_WORKSPACES_DIR"`
}

func (c *Config) Validate() error {
	var errs error

	if port, err := strconv.Atoi(c.Port); err != nil || !(1 <= port && port <= 65535) {
		errs = errors.Join(errs, fmt.Errorf("CPENV_PORT %q must be an integer between 1 and 65535", c.Port))
	}

	if c.Home != "" {
		if !filepath.IsAbs(c.Home) {
			errs = errors.Join(errs, fmt.Errorf("CPENV_HOME %q must be an absolute path", c.Home))
		} else {
			if c.StatePath == "" {
				c.StatePath = filepath.Join(c.Home, "state.json")
			}
			if c.TemplatesDir == "" {
				c.TemplatesDir = filepath.Join(c.Home, "templates")
			}
			if c.WorkspacesDir == "" {
				c.WorkspacesDir = filepath.Join(c.Home, "workspaces")
			}
		}
	}

	if c.StatePath != "" && !filepath.IsAbs(c.StatePath) {
		errs = errors.Join(errs, fmt.Errorf("CPENV_STATE_PATH %q must be an absolute path", c.StatePath))
	}
	if c.TemplatesDir != "" && !filepath.IsAbs(c.TemplatesDir) {
		errs = errors.Join(errs, fmt.Errorf("CPENV_TEMPLATES_DIR %q must be an absolute path", c.TemplatesDir))
	}
	if c.WorkspacesDir != "" && !filepath.IsAbs(c.WorkspacesDir) {
		errs = errors.Join(errs, fmt.Errorf("CPENV_WORKSPACES_DIR %q must be an absolute path", c.WorkspacesDir))
	}

	if c.StatePath == "" || c.TemplatesDir == "" || c.WorkspacesDir == "" {
		errs = errors.Join(errs, fmt.Errorf("CPENV_HOME or CPENV_STATE_PATH, CPENV_TEMPLATES_DIR, and CPENV_WORKSPACES_DIR must be set"))
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
