package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/afero"
)

func materialize(ctx context.Context, fs afero.Fs, dir *Dir) error {
	var errs error
	for name, entry := range dir.Entries {
		switch {
		case entry.File != nil && entry.Dir == nil:
			errs = errors.Join(errs, afero.WriteFile(fs, name, []byte(entry.File.Content), 0644))
		case entry.Dir != nil && entry.File == nil:
			errs = errors.Join(errs, materialize(ctx, afero.NewBasePathFs(fs, name), entry.Dir))
		default:
			errs = errors.Join(errs, fmt.Errorf("invalid entry: %v", entry))
		}
	}
	return errs
}
