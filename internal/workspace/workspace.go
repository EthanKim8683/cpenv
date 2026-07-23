package workspace

import (
	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/spf13/afero"
)

type Workspace struct {
	ArchiveFs   afero.Fs
	ScaffoldsFs afero.Fs
	WorkspaceFs afero.Fs
}

func (w *Workspace) Focus(problem *problemv1.Problem) error {
	return nil
}

func (w *Workspace) Rescaffold(scaffoldFile string) error {
	return nil
}
