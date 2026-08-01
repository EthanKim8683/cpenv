package config

import (
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Port         string `env:"PORT" envDefault:"8080"`
	EnvsDir      string `env:"ENVS_DIR,notEmpty"`
	StatePath    string `env:"STATE_PATH,notEmpty"`
	TemplatesDir string `env:"TEMPLATES_DIR,notEmpty"`
}

func (c *Config) Validate() error {
	var errs error
	if port, err := strconv.Atoi(c.Port); err != nil || !(1 <= port && port <= 65535) {
		errs = errors.Join(errs, fmt.Errorf("PORT %q must be an integer between 1 and 65535", c.Port))
	}
	if !filepath.IsAbs(c.EnvsDir) {
		errs = errors.Join(errs, fmt.Errorf("ENVS_DIR %q must be an absolute path", c.EnvsDir))
	}
	if !filepath.IsAbs(c.StatePath) {
		errs = errors.Join(errs, fmt.Errorf("STATE_PATH %q must be an absolute path", c.StatePath))
	}
	if !filepath.IsAbs(c.TemplatesDir) {
		errs = errors.Join(errs, fmt.Errorf("TEMPLATES_DIR %q must be an absolute path", c.TemplatesDir))
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
