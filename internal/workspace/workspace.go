package workspace

import (
	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
)

type Workspace struct {
}

func (w *Workspace) Focus(problem *problemv1.Problem) error {
	return nil
}

func (w *Workspace) Scaffold(scaffoldFile string) error {
	return nil
}
