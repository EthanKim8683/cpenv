package workspace

import (
	"testing"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dirEntry(entries map[string]*entry) *entry {
	return &entry{Dir: &dir{Entries: entries}}
}

func fileEntry(content string) *entry {
	return &entry{File: &file{Content: content}}
}

func TestRender(t *testing.T) {
	scaffoldsFs := afero.NewBasePathFs(afero.NewOsFs(), "./testdata")

	t.Run("happy path", func(t *testing.T) {
		problem := &problemv1.Problem{
			Id:   "123",
			Type: problemv1.ProblemType_PROBLEM_TYPE_STDIO_BATCH,
			Samples: []*problemv1.ProblemSample{
				{
					Input:  "1 2",
					Output: "3",
				},
				{
					Input:  "4 5",
					Output: "6",
				},
			},
		}

		wantEntry := dirEntry(map[string]*entry{
			"id":   fileEntry("123"),
			"type": fileEntry("PROBLEM_TYPE_STDIO_BATCH"),
			"samples": dirEntry(map[string]*entry{
				"0": dirEntry(map[string]*entry{
					"input":  fileEntry("1 2"),
					"output": fileEntry("3"),
				}),
				"1": dirEntry(map[string]*entry{
					"input":  fileEntry("4 5"),
					"output": fileEntry("6"),
				}),
			}),
		})

		entry, err := render(scaffoldsFs, "happy-path.star", problem)
		require.NoError(t, err)
		assert.Equal(t, wantEntry, entry)
	})

	t.Run("encode error", func(t *testing.T) {
		_, err := render(scaffoldsFs, "encode-error.star", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "encode: want dict, got string")
	})

	t.Run("expand error", func(t *testing.T) {
		_, err := render(scaffoldsFs, "expand-error.star", nil)
		require.Error(t, err)
		assert.ErrorContains(t, err, "expand: want string or dict, got int")
	})
}
