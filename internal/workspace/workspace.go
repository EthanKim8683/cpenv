package workspace

type Workspace struct {
	Path string
}

func NewWorkspace(path string) *Workspace {
	return &Workspace{Path: path}
}

func (w *Workspace) Scaffold() error {
	return nil
}
