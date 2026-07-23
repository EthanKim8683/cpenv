package workspace

import (
	"log/slog"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/afero"
)

type mutex interface {
	TryRLock() (bool, error)
	Unlock() error
}

type archiver struct {
	workspaceFs afero.Fs
	archiveFs   afero.Fs
	mu          mutex
	log         *slog.Logger
}

func (a *archiver) archive(event fsnotify.Event) error {
	locked, err := a.mu.TryRLock()
	if err != nil {
		return nil // gotta figure out what err means here
	}
	if !locked {
		a.log.Info("archive: workspace is being modified by cpenv, not archiving for now")
		return nil
	}
	defer a.mu.Unlock()

	return nil
}
