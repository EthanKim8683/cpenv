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

		require.NoError(t, renderTemplate(path, dir, problem))

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

		err := renderTemplate(path, dir, nil)
		require.ErrorContains(t, err, "decode files")
		assert.ErrorContains(t, err, "0: expected string file name, got int")
		assert.ErrorContains(t, err, "0: expected string content, got int")
		assert.ErrorContains(t, err, "2: expected string file name, got int")
		assert.ErrorContains(t, err, "\"4\": expected string content, got int")
	})
}
