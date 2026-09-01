package cli

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveFirst(t *testing.T) {
	t.Parallel()

	t.Run("coverage", func(t *testing.T) {
		t.Parallel()

		res, err := resolveFirst("", []resolver[int]{
			func(name string) (int, error) {
				return 0, skipResolver
			},
			func(name string) (int, error) {
				return 0, errors.New("error")
			},
			func(name string) (int, error) {
				return 1, nil
			},
			func(name string) (int, error) {
				return 2, nil
			},
		})
		require.NoError(t, err)
		assert.Equal(t, 1, res)
	})

	t.Run("joins errors", func(t *testing.T) {
		t.Parallel()

		res, err := resolveFirst("", []resolver[int]{
			func(name string) (int, error) {
				return 0, errors.New("error 1")
			},
			func(name string) (int, error) {
				return 0, skipResolver
			},
			func(name string) (int, error) {
				return 0, errors.New("error 2")
			},
		})
		assert.ErrorContains(t, err, "error 1")
		assert.ErrorContains(t, err, "error 2")
		assert.Zero(t, res)
	})
}

func TestResolveTemplateAbsPath(t *testing.T) {
	t.Parallel()

	name, err := filepath.Abs(filepath.Join("testdata", "resolve-template", "abs.star"))
	require.NoError(t, err)

	tmpl, err := resolveTemplateAbsPath()(name)
	require.NoError(t, err)
	assert.Equal(t, name, tmpl.path)
}

func TestResolveTemplateRelPath(t *testing.T) {
	t.Parallel()

	name := "rel.star"
	dir, err := filepath.Abs(filepath.Join("testdata", "resolve-template", "dir"))
	require.NoError(t, err)

	tmpl, err := resolveTemplateRelPath(dir)(name)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, name), tmpl.path)
}

type fakePreferences struct {
	defaultTemplate string
}

func (p *fakePreferences) DefaultTemplate() (string, error) {
	return p.defaultTemplate, nil
}

func (p *fakePreferences) SetDefaultTemplate(path string) error {
	return nil
}

func TestResolveTemplateDefault(t *testing.T) {
	t.Parallel()

	path, err := filepath.Abs(filepath.Join("testdata", "resolve-template", "default.star"))
	require.NoError(t, err)

	p := &fakePreferences{defaultTemplate: path}

	tmpl, err := resolveTemplateDefault(p)("")
	require.NoError(t, err)
	assert.Equal(t, path, tmpl.path)
}

func TestResolveTemplateGlob(t *testing.T) {
	t.Parallel()

	dir, err := filepath.Abs(filepath.Join("testdata", "resolve-template", "dir"))
	require.NoError(t, err)

	tmpl, err := resolveTemplateGlob(dir)("")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "rel.star"), tmpl.path)
}

func TestResolveSolutionAbsPath(t *testing.T) {
	t.Parallel()

	name, err := filepath.Abs(filepath.Join("testdata", "resolve-solution", "abs.cpp"))
	require.NoError(t, err)

	sol, err := resolveSolutionAbsPath()(name)
	require.NoError(t, err)
	assert.Equal(t, name, sol.path)
}

func TestResolveSolutionRelPath(t *testing.T) {
	t.Parallel()

	name := "rel.cpp"
	dir, err := filepath.Abs(filepath.Join("testdata", "resolve-solution", "dir"))
	require.NoError(t, err)

	sol, err := resolveSolutionRelPath(dir)(name)
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, name), sol.path)
}

func TestResolveSolutionGlob(t *testing.T) {
	t.Parallel()

	dir, err := filepath.Abs(filepath.Join("testdata", "resolve-solution", "dir"))
	require.NoError(t, err)

	sol, err := resolveSolutionGlob(dir)("")
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(dir, "sol.cpp"), sol.path)
}
