package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readFiles(t *testing.T, dir string) map[string]string {
	t.Helper()

	files := make(map[string]string)
	filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		require.NoError(t, err)

		if info.IsDir() {
			return nil
		}

		content, err := os.ReadFile(path)
		require.NoError(t, err)

		path, err = filepath.Rel(dir, path)
		require.NoError(t, err)

		files[path] = string(content)
		return nil
	})
	return files
}

func TestTemplate(t *testing.T) {
	t.Parallel()

	t.Run("coverage", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join("testdata", "template", "coverage.star")
		dir := t.TempDir()
		problem := &problemv1.Problem{
			Id:   "id",
			Type: problemv1.ProblemType_PROBLEM_TYPE_STDIO_BATCH,
			Samples: []*problemv1.ProblemSample{
				{Input: "input 0", Output: "output 0"},
				{Input: "input 1", Output: "output 1"},
			},
		}

		src, err := os.ReadFile(path)
		require.NoError(t, err)

		tl := &template{path: path, src: src}
		require.NoError(t, tl.render(dir, problem))
		files := readFiles(t, dir)
		assert.Equal(t, map[string]string{
			"id":               "id",
			"type":             "PROBLEM_TYPE_STDIO_BATCH",
			"samples/0/input":  "input 0",
			"samples/0/output": "output 0",
			"samples/1/input":  "input 1",
			"samples/1/output": "output 1",
		}, files)
	})

	t.Run("decode error", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join("testdata", "template", "decode-error.star")
		dir := t.TempDir()

		src, err := os.ReadFile(path)
		require.NoError(t, err)

		tl := &template{path: path, src: src}
		err = tl.render(dir, nil)
		require.ErrorContains(t, err, "decode files")
		assert.ErrorContains(t, err, "0: expected string file name, got int")
		assert.ErrorContains(t, err, "0: expected string content, got int")
		assert.ErrorContains(t, err, "2: expected string file name, got int")
		assert.ErrorContains(t, err, "\"4\": expected string content, got int")
	})
}

func TestResolveTemplate(t *testing.T) {
	t.Parallel()

	cwd := filepath.Join(t.TempDir(), "cwd")
	templatesDir := filepath.Join(t.TempDir(), "templates")
	defaultTemplate := filepath.Join(cwd, "default.star")
	relPath := "template.star"
	cwdRelPath := filepath.Join(cwd, "template.star")
	templatesDirRelPath := filepath.Join(templatesDir, "template.star")
	absPath := filepath.Join(t.TempDir(), "template.star")

	require.NoError(t, os.MkdirAll(cwd, 0755))
	require.NoError(t, os.MkdirAll(templatesDir, 0755))
	require.NoError(t, os.WriteFile(defaultTemplate, nil, 0644))
	require.NoError(t, os.WriteFile(cwdRelPath, nil, 0644))
	require.NoError(t, os.WriteFile(templatesDirRelPath, nil, 0644))
	require.NoError(t, os.WriteFile(absPath, nil, 0644))

	t.Run("no templates", func(t *testing.T) {
		t.Parallel()

		tl, err := resolveTemplate("", t.TempDir(), t.TempDir(), "")
		require.NoError(t, err)
		assert.Nil(t, tl)
	})

	t.Run("glob", func(t *testing.T) {
		t.Parallel()

		tl, err := resolveTemplate("", t.TempDir(), templatesDir, "")
		require.NoError(t, err)
		assert.Equal(t, templatesDirRelPath, tl.path)
	})

	t.Run("default template", func(t *testing.T) {
		t.Parallel()

		tl, err := resolveTemplate("", t.TempDir(), t.TempDir(), defaultTemplate)
		require.NoError(t, err)
		assert.Equal(t, defaultTemplate, tl.path)
	})

	t.Run("relative to templates dir", func(t *testing.T) {
		t.Parallel()

		tl, err := resolveTemplate(relPath, t.TempDir(), templatesDir, defaultTemplate)
		require.NoError(t, err)
		assert.Equal(t, templatesDirRelPath, tl.path)
	})

	t.Run("relative to cwd", func(t *testing.T) {
		t.Parallel()

		tl, err := resolveTemplate(relPath, cwd, templatesDir, defaultTemplate)
		require.NoError(t, err)
		assert.Equal(t, cwdRelPath, tl.path)
	})

	t.Run("absolute path", func(t *testing.T) {
		t.Parallel()

		tl, err := resolveTemplate(absPath, cwd, templatesDir, defaultTemplate)
		require.NoError(t, err)
		assert.Equal(t, absPath, tl.path)
	})
}
