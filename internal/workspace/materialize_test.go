package workspace

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaterialize(t *testing.T) {
	tests := map[string]struct {
		entry     *entry
		wantFiles map[string]string
		wantErr   error
	}{
		"file": {
			entry: dirEntry(map[string]*entry{
				"foo": fileEntry("foo"),
			}),
			wantFiles: map[string]string{
				"foo": "foo",
			},
			wantErr: nil,
		},
		"dir": {
			entry: dirEntry(map[string]*entry{
				"foo": dirEntry(map[string]*entry{
					"bar": fileEntry("bar"),
				}),
			}),
			wantFiles: map[string]string{
				"foo/bar": "bar",
			},
			wantErr: nil,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotFs := afero.NewMemMapFs()
			err := materialize(context.Background(), gotFs, test.entry)
			if test.wantErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, test.wantErr.Error())
				assert.Empty(t, gotFs)
			} else {
				assert.NoError(t, err)

				files := make(map[string]string)
				_ = afero.Walk(gotFs, ".", func(path string, info os.FileInfo, err error) error {
					require.NoError(t, err)
					if info.IsDir() {
						return nil
					}
					content, err := afero.ReadFile(gotFs, path)
					require.NoError(t, err)
					files[filepath.ToSlash(path)] = string(content)
					return nil
				})
				assert.Equal(t, test.wantFiles, files)
			}
		})
	}
}
