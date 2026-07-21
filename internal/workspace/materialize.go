package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/afero"
)

func materialize(ctx context.Context, fs afero.Fs, entry *entry) error {
	// entry is either a root directory (see encode.star) or a subdirectory.
	var errs error
	for name, entry := range entry.Dir.Entries {
		// Exactly one of File or Dir is set (see encode.star).
		if entry.File != nil {
			err := afero.WriteFile(fs, name, []byte(entry.File.Content), 0644)
			if err != nil {
				errs = errors.Join(errs, fmt.Errorf("materialize %q: write file: %w", name, err))
			}
		} else {
			if err := materialize(ctx, afero.NewBasePathFs(fs, name), entry); err != nil {
				errs = errors.Join(errs, fmt.Errorf("materialize %q: %w", name, err))
			}
		}
	}
	return errs
}
