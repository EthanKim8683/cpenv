package workspace

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/afero"
)

func materialize(ctx context.Context, fs afero.Fs, entry *Entry) error {
	file := entry.File
	dir := entry.Dir
	switch {
	case file != nil && dir == nil:
		return afero.WriteFile(fs, file.Name, []byte(file.Content), 0644)
	case file == nil && dir != nil:
		if err := fs.MkdirAll(dir.Name, 0755); err != nil {
			return err
		}

		var errs error
		subFs := afero.NewBasePathFs(fs, dir.Name)
		for _, subEntry := range dir.Entries {
			errors.Join(errs, materialize(ctx, subFs, subEntry))
		}
		return nil
	default:
		return fmt.Errorf("invalid entry: %v", entry)
	}
}
