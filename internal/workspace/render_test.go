package workspace

import (
	"testing"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dirFrom(entries map[string]*Entry) *Entry {
	return &Entry{
		Dir: &Dir{
			Entries: entries,
		},
	}
}

func fileFrom(content string) *Entry {
	return &Entry{
		File: &File{
			Content: content,
		},
	}
}

func TestRender(t *testing.T) {
	scaffoldsFs := afero.NewBasePathFs(afero.NewOsFs(), "./testdata")
	scaffoldFile := "scaffold.star"

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

	wantEntry := dirFrom(map[string]*Entry{
		"id":   fileFrom("123"),
		"type": fileFrom("PROBLEM_TYPE_STDIO_BATCH"),
		"samples": dirFrom(map[string]*Entry{
			"0": dirFrom(map[string]*Entry{
				"input":  fileFrom("1 2"),
				"output": fileFrom("3"),
			}),
			"1": dirFrom(map[string]*Entry{
				"input":  fileFrom("4 5"),
				"output": fileFrom("6"),
			}),
		}),
	})

	entry, err := render(scaffoldsFs, scaffoldFile, problem)
	require.NoError(t, err)
	assert.Equal(t, wantEntry, entry)
}
