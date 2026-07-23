package workspace

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

func copyFs(srcFs afero.Fs, dstFs afero.Fs) error {
	if err := afero.Walk(srcFs, ".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if err := dstFs.MkdirAll(path, info.Mode()); err != nil {
				return fmt.Errorf("mkdir dst %q: %w", path, err)
			}
			return nil
		}

		in, err := srcFs.Open(path)
		if err != nil {
			return fmt.Errorf("open src %q: %w", path, err)
		}
		defer in.Close()

		perm := info.Mode().Perm()
		out, err := dstFs.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
		if err != nil {
			return fmt.Errorf("create dst %q: %w", path, err)
		}

		if _, err := io.Copy(out, in); err != nil {
			_ = out.Close()
			return fmt.Errorf("copy %q: %w", path, err)
		}

		if err := out.Close(); err != nil {
			return fmt.Errorf("close dst %q: %w", path, err)
		}

		if err := dstFs.Chmod(path, perm); err != nil {
			return fmt.Errorf("chmod dst %q: %w", path, err)
		}

		return nil
	}); err != nil {
		return fmt.Errorf("copy fs: walk: %w", err)
	}
	return nil
}

func clearFs(fs afero.Fs) error {
	entries, err := afero.ReadDir(fs, ".")
	if err != nil {
		return fmt.Errorf("clear fs: read dir: %w", err)
	}

	for _, entry := range entries {
		if err := fs.RemoveAll(entry.Name()); err != nil {
			return fmt.Errorf("clear fs: remove %q: %w", entry.Name(), err)
		}
	}
	return nil
}

func replaceFs(srcFs afero.Fs, dstFs afero.Fs) error {
	if err := clearFs(dstFs); err != nil {
		return fmt.Errorf("replace fs: clear dst fs: %w", err)
	}

	if err := copyFs(srcFs, dstFs); err != nil {
		return fmt.Errorf("replace fs: copy fs: %w", err)
	}

	return nil
}

func replaceDir(srcFs afero.Fs, dstFs afero.Fs, dir string) error {
	dir = filepath.Clean(dir)
	if !filepath.IsLocal(dir) || dir == "." {
		return errors.New("replace dir: dir must be local and not root")
	}

	tmpDir := dir + ".tmp"
	swpDir := dir + ".swp"

	if err := dstFs.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("replace dir: remove leftover tmp dir: %w", err)
	}

	if err := dstFs.RemoveAll(swpDir); err != nil {
		return fmt.Errorf("replace dir: remove leftover swp dir: %w", err)
	}

	if err := dstFs.MkdirAll(tmpDir, 0755); err != nil {
		return fmt.Errorf("replace dir: create tmp dir: %w", err)
	}

	if err := copyFs(srcFs, afero.NewBasePathFs(dstFs, tmpDir)); err != nil {
		_ = dstFs.RemoveAll(tmpDir)
		return fmt.Errorf("replace dir: copy fs: %w", err)
	}

	if err := dstFs.Rename(dir, swpDir); err != nil && !errors.Is(err, fs.ErrNotExist) {
		_ = dstFs.RemoveAll(tmpDir)
		return fmt.Errorf("replace dir: rename dir: %w", err)
	}

	if err := dstFs.Rename(tmpDir, dir); err != nil {
		_ = dstFs.Rename(swpDir, dir)
		_ = dstFs.RemoveAll(tmpDir)
		return fmt.Errorf("replace dir: rename tmp dir: %w", err)
	}

	if err := dstFs.RemoveAll(swpDir); err != nil {
		return fmt.Errorf("replace dir: remove swp dir: %w", err)
	}

	return nil
}
