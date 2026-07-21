package workspace

import (
	"context"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
)

type Workspace struct {
	ArchiveFs   afero.Fs
	ScaffoldFs  afero.Fs
	WorkspaceFs afero.Fs
}

func (w *Workspace) Focus(ctx context.Context, tabId string, problem *problemv1.Problem) error {
}
