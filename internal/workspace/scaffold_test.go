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
	require.NoError(t, afero.Walk(fs, ".", func(
		path string,
		info os.FileInfo,
		err error,
	) error {
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

func TestRender(t *testing.T) {
	t.Parallel()

	scaffoldsFs := afero.NewBasePathFs(afero.NewOsFs(), "./testdata")

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

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()

		workspaceFs := afero.NewMemMapFs()
		require.NoError(t, scaffold(
			workspaceFs,
			scaffoldsFs,
			"happy-path.star",
			problem,
		))

		files := readFiles(t, workspaceFs)
		assert.Equal(t, map[string]string{
			"id":               "id",
			"type":             "PROBLEM_TYPE_STDIO_BATCH",
			"samples/0/input":  "input 0",
			"samples/0/output": "output 0",
			"samples/1/input":  "input 1",
			"samples/1/output": "output 1",
		}, files)
	})

	t.Run("render error", func(t *testing.T) {
		t.Parallel()
		err := scaffold(afero.NewMemMapFs(), scaffoldsFs, "render-error.star", nil)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "scaffold \"render-error.star\": render files: ")
	})

	t.Run("decode error", func(t *testing.T) {
		t.Parallel()
		err := scaffold(afero.NewMemMapFs(), scaffoldsFs, "decode-error.star", nil)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "scaffold \"decode-error.star\": decode files: ")
		assert.ErrorContains(t, err, "decode file 0: expected string path, got int")
		assert.ErrorContains(t, err, "decode file 0: expected string content, got int")
		assert.ErrorContains(t, err, "decode file 2: expected string path, got int")
		assert.ErrorContains(t, err, "decode file \"4\": expected string content, got int")
	})
}
