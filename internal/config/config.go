package config

import "github.com/caarlos0/env/v11"

type Config struct {
	ArchiveDir   string `env:"ARCHIVE_DIR"`
	ScaffoldsDir string `env:"SCAFFOLDS_DIR"`
	WorkspaceDir string `env:"WORKSPACE_DIR"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
