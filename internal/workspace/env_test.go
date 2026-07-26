package workspace

import (
	"os"
	"testing"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readFiles(t *testing.T, fs afero.Fs) map[string]string {
	t.Helper()

	files := make(map[string]string)
	require.NoError(t, afero.Walk(fs, ".", func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)

		if info.IsDir() {
			return nil
		}

		content, err := afero.ReadFile(fs, path)
		require.NoError(t, err)
		files[path] = string(content)

		return nil
	}))
	return files
}

func TestEnv(t *testing.T) {
	t.Parallel()

	t.Run("init", func(t *testing.T) {
		t.Parallel()

		templatesFs := afero.NewBasePathFs(afero.NewOsFs(), "./testdata")

		t.Run("happy path", func(t *testing.T) {
			t.Parallel()

			fs := afero.NewMemMapFs()
			problem := &problemv1.Problem{
				Id:   "id",
				Type: problemv1.ProblemType_PROBLEM_TYPE_STDIO_BATCH,
				Samples: []*problemv1.ProblemSample{
					{
						Input:  "input 0",
						Output: "output 0",
					},
					{
						Input:  "input 1",
						Output: "output 1",
					},
				},
			}
			env := NewEnv(fs, problem)

			require.NoError(t, env.Init(templatesFs, "happy-path.star"))

			files := readFiles(t, fs)
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

			env := NewEnv(afero.NewMemMapFs(), nil)

			err := env.Init(templatesFs, "decode-error.star")
			assert.Error(t, err)
			assert.ErrorContains(t, err, "template \"decode-error.star\": ")
			assert.ErrorContains(t, err, "file 0: expected string path, got int")
			assert.ErrorContains(t, err, "file 0: expected string content, got int")
			assert.ErrorContains(t, err, "file 2: expected string path, got int")
			assert.ErrorContains(t, err, "file \"4\": expected string content, got int")
		})
	})
}
