package workspace

import (
	"encoding/json"
	"fmt"
	"maps"

	_ "embed"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
	starlarkjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"google.golang.org/protobuf/encoding/protojson"
)

//go:embed lib.star
var libStar string
var libGlobals starlark.StringDict

func mergeStringDicts(dicts ...starlark.StringDict) starlark.StringDict {
	out := make(starlark.StringDict)
	for _, dict := range dicts {
		maps.Copy(out, dict)
	}
	return out
}

func render(scaffoldsFs afero.Fs, scaffoldFile string, problem *problemv1.Problem) (*Entry, error) {
	thread := &starlark.Thread{}

	problemJSON, err := protojson.Marshal(problem)
	if err != nil {
		return nil, err
	}

	problemValue, err := starlark.Call(
		thread,
		starlarkjson.Module.Members["decode"],
		starlark.Tuple{starlark.String(problemJSON)},
		nil,
	)
	if err != nil {
		return nil, err
	}

	scaffold, err := afero.ReadFile(scaffoldsFs, scaffoldFile)
	if err != nil {
		return nil, err
	}

	globals, err := starlark.ExecFileOptions(
		&syntax.FileOptions{
			While:           true,
			TopLevelControl: true,
		},
		thread,
		scaffoldFile,
		scaffold,
		mergeStringDicts(libGlobals, starlark.StringDict{
			"problem": problemValue,
		}),
	)
	if err != nil {
		return nil, err
	}

	workspaceValue, ok := globals["workspace"]
	if !ok {
		return nil, fmt.Errorf("workspace not found")
	}

	entryJSONValue, err := starlark.Call(
		thread,
		starlarkjson.Module.Members["encode"],
		starlark.Tuple{workspaceValue},
		nil,
	)
	if err != nil {
		return nil, err
	}

	entryJSON, ok := starlark.AsString(entryJSONValue)
	if !ok {
		return nil, fmt.Errorf("workspace is not a string")
	}

	var entry Entry
	if err := json.Unmarshal([]byte(entryJSON), &entry); err != nil {
		return nil, err
	}
	return &entry, nil
}

func init() {
	var err error
	libGlobals, err = starlark.ExecFileOptions(
		&syntax.FileOptions{},
		&starlark.Thread{},
		"lib.star",
		libStar,
		nil,
	)
	if err != nil {
		panic(err)
	}
}
