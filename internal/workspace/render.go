package workspace

import (
	"encoding/json"
	"fmt"

	_ "embed"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
	starlarkjson "go.starlark.net/lib/json"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
	"google.golang.org/protobuf/encoding/protojson"
)

//go:embed encode.star
var encodeStar string
var encode starlark.Value

func encodeProblem(
	thread *starlark.Thread,
	problem *problemv1.Problem,
) (starlark.Value, error) {
	data, err := protojson.Marshal(problem)
	if err != nil {
		return nil, fmt.Errorf("encode problem: marshal: %w", err)
	}

	value, err := starlark.Call(
		thread,
		starlarkjson.Module.Members["decode"],
		starlark.Tuple{starlark.String(data)},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("encode problem: starlark decode: %w", err)
	}

	return value, nil
}

func renderWorkspace(
	thread *starlark.Thread,
	scaffoldsFs afero.Fs,
	scaffoldFile string,
	problemValue starlark.Value,
) (starlark.Value, error) {
	scaffold, err := afero.ReadFile(scaffoldsFs, scaffoldFile)
	if err != nil {
		return nil, fmt.Errorf("read scaffold %q: %w", scaffoldFile, err)
	}

	globals, err := starlark.ExecFileOptions(
		&syntax.FileOptions{
			While:           true,
			TopLevelControl: true,
		},
		thread,
		scaffoldFile,
		scaffold,
		starlark.StringDict{
			"problem": problemValue,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("exec scaffold %q: %w", scaffoldFile, err)
	}

	value, ok := globals["workspace"]
	if !ok {
		return nil, fmt.Errorf("scaffold %q: missing global workspace", scaffoldFile)
	}
	return value, nil
}

func decodeWorkspace(
	thread *starlark.Thread,
	value starlark.Value,
) (*entry, error) {
	value, err := starlark.Call(
		thread,
		encode,
		starlark.Tuple{value},
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("decode workspace: starlark encode: %w", err)
	}

	data, ok := starlark.AsString(value)
	if !ok {
		return nil, fmt.Errorf("decode workspace: expected string, got %T", value)
	}

	var entry entry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, fmt.Errorf("decode workspace: unmarshal: %w", err)
	}
	return &entry, nil
}

func render(
	scaffoldsFs afero.Fs,
	scaffoldFile string,
	problem *problemv1.Problem,
) (*entry, error) {
	thread := &starlark.Thread{}

	problemValue, err := encodeProblem(thread, problem)
	if err != nil {
		return nil, fmt.Errorf("render %q: %w", scaffoldFile, err)
	}

	workspaceValue, err := renderWorkspace(
		thread,
		scaffoldsFs,
		scaffoldFile,
		problemValue,
	)
	if err != nil {
		return nil, fmt.Errorf("render %q: %w", scaffoldFile, err)
	}

	entry, err := decodeWorkspace(thread, workspaceValue)
	if err != nil {
		return nil, fmt.Errorf("render %q: %w", scaffoldFile, err)
	}

	return entry, nil
}

func init() {
	globals, err := starlark.ExecFileOptions(
		&syntax.FileOptions{
			Recursion: true,
		},
		&starlark.Thread{},
		"encode.star",
		encodeStar,
		starlark.StringDict{
			"json": starlarkjson.Module,
		},
	)
	if err != nil {
		panic(fmt.Errorf("init render: exec encode.star: %w", err))
	}

	var ok bool
	encode, ok = globals["encode"]
	if !ok {
		panic("init render: missing global encode")
	}
}
