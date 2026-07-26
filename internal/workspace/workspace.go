package workspace

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type Workspace struct {
	log     *slog.Logger
	path    string
	store   *store
	mu      sync.Mutex
	watcher *fsnotify.Watcher
}

func (w *Workspace) Open(name string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.watcher != nil {
		w.watcher.Close()
		w.watcher = nil
	}

	if err := w.store.save(); err != nil {
		return fmt.Errorf("open: save: %w", err)
	}

	if err := w.store.load(name); err != nil {
		return fmt.Errorf("open: load: %w", err)
	}

	return nil
}

func (w *Workspace) Watch(ctx context.Context) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("watch: new watcher: %w", err)
	}
	defer watcher.Close()

	if err := watcher.Add(w.path); err != nil {
		return fmt.Errorf("watch: add workspace path: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return errors.New("watch: watcher closed unexpectedly")
			}

			relPath, err := filepath.Rel(w.path, event.Name)
			if err != nil {
				w.log.Debug("error handling event", "path", event.Name, "error", err)
				continue
			}

			relPath = filepath.ToSlash(relPath)
			if slices.Contains(strings.Split(relPath, "/"), ".git") {
				continue
			}

			if event.Op == fsnotify.Chmod {
				continue
			}

			save()
		case err, ok := <-watcher.Errors:
			if !ok {
				return errors.New("watch: watcher closed unexpectedly")
			}

			if errors.Is(err, fsnotify.ErrEventOverflow) {
				w.log.Warn("event overflow; saving")
				save()
				continue
			}

			w.log.Error("error watching workspace", "error", err)
		}
	}
}

func NewWorkspace(log *slog.Logger, path, baseBranch string) (*Workspace, error) {
	log = log.With("component", "workspace", "path", path)

	store, err := newStore(path, baseBranch)
	if err != nil {
		return nil, fmt.Errorf("new workspace: new store: %w", err)
	}

	return &Workspace{
		log:   log,
		path:  path,
		store: store,
	}, nil
}
