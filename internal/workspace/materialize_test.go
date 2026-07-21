package workspace

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaterialize(t *testing.T) {
	tests := map[string]struct {
		dir       *Dir
		wantFiles map[string]string
		wantErr   error
	}{
		"file": {
			dir: &Dir{
				Entries: map[string]*Entry{
					"foo": {
						File: &File{
							Content: "foo",
						},
					},
				},
			},
			wantFiles: map[string]string{
				"foo": "foo",
			},
			wantErr: nil,
		},
		"dir": {
			dir: &Dir{
				Entries: map[string]*Entry{
					"foo": {
						Dir: &Dir{
							Entries: map[string]*Entry{
								"bar": {
									File: &File{
										Content: "bar",
									},
								},
							},
						},
					},
				},
			},
			wantFiles: map[string]string{
				"foo/bar": "bar",
			},
			wantErr: nil,
		},
		"both file and dir": {
			dir: &Dir{
				Entries: map[string]*Entry{
					"foo": {
						File: &File{
							Content: "foo",
						},
						Dir: &Dir{
							Entries: map[string]*Entry{},
						},
					},
				},
			},
			wantErr: errors.New("invalid entry"),
		},
		"neither file nor dir": {
			dir: &Dir{
				Entries: map[string]*Entry{
					"foo": {},
				},
			},
			wantErr: errors.New("invalid entry"),
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			gotFs := afero.NewMemMapFs()
			err := materialize(context.Background(), gotFs, test.dir)
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
