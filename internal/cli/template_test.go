package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/EthanKim8683/cpenv/internal/config"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"
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

func removeFiles(t *testing.T, dir string) {
	t.Helper()

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)

	for _, entry := range entries {
		require.NoError(t, os.RemoveAll(filepath.Join(dir, entry.Name())))
	}
}

func TestRenderTemplate(t *testing.T) {
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

func TestCLI_resolveTemplate(t *testing.T) {
	t.Parallel()

	homeDir := filepath.Join(t.TempDir(), "home")
	cwd := filepath.Join(t.TempDir(), "cwd")

	db, err := bolt.Open(filepath.Join(t.TempDir(), "db.db"), 0600, nil)
	require.NoError(t, err)
	cli := &CLI{
		Cfg: &config.Config{HomeDir: homeDir},
		CWD: cwd,
		DB:  db,
	}

	t.Run("void", func(t *testing.T) {
		path, err := cli.resolveTemplate("")
		require.NoError(t, err)
		assert.Equal(t, "", path)
	})

	t.Run("glob", func(t *testing.T) {
		relName := "rel.star"

		require.NoError(t, os.MkdirAll(cli.templatesDir(), 0755))
		require.NoError(t, os.WriteFile(filepath.Join(cli.templatesDir(), relName), nil, 0644))

		path, err := cli.resolveTemplate("")
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(cli.templatesDir(), relName), path)
	})

	t.Run("default template", func(t *testing.T) {
		defaultPath := filepath.Join(t.TempDir(), "default.star")

		require.NoError(t, os.WriteFile(defaultPath, nil, 0644))
		_ = db.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucket(templateBucketKey)
			require.NoError(t, err)
			require.NoError(t, b.Put(defaultTemplateKey, []byte(defaultPath)))
			return nil
		})

		path, err := cli.resolveTemplate("")
		require.NoError(t, err)
		assert.Equal(t, defaultPath, path)
	})

	t.Run("templates dir relative path", func(t *testing.T) {
		relName := "rel.star"

		path, err := cli.resolveTemplate(relName)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(cli.templatesDir(), relName), path)
	})

	t.Run("cwd relative path", func(t *testing.T) {
		relName := "rel.star"

		require.NoError(t, os.MkdirAll(cwd, 0755))
		require.NoError(t, os.WriteFile(filepath.Join(cwd, relName), nil, 0644))

		path, err := cli.resolveTemplate(relName)
		require.NoError(t, err)
		assert.Equal(t, filepath.Join(cwd, relName), path)
	})

	t.Run("absolute path", func(t *testing.T) {
		absName := filepath.Join(t.TempDir(), "abs.star")
		require.NoError(t, os.WriteFile(absName, nil, 0644))

		path, err := cli.resolveTemplate(absName)
		require.NoError(t, err)
		assert.Equal(t, absName, path)
	})
}

func TestCLI_renderTemplate(t *testing.T) {
	t.Parallel()

	t.Run("default template round trip", func(t *testing.T) {
		t.Parallel()

		homeDir := filepath.Join(t.TempDir(), "home")
		cwd := filepath.Join(t.TempDir(), "cwd")
		name, err := filepath.Abs(filepath.Join("testdata", "template", "minimal.star"))
		require.NoError(t, err)
		problem1 := &problemv1.Problem{Id: "id1"}
		problem2 := &problemv1.Problem{Id: "id2"}

		db, err := bolt.Open(filepath.Join(t.TempDir(), "db.db"), 0600, nil)
		require.NoError(t, err)
		cli := &CLI{
			Cfg: &config.Config{HomeDir: homeDir},
			CWD: cwd,
			DB:  db,
		}

		require.NoError(t, cli.renderTemplate(name, cwd, problem1))
		assert.Equal(t, map[string]string{"id": "id1"}, readFiles(t, cwd))

		removeFiles(t, cwd)

		require.NoError(t, cli.renderTemplate("", cwd, problem2))
		assert.Equal(t, map[string]string{"id": "id2"}, readFiles(t, cwd))
	})
}
