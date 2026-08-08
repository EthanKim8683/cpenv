package cli

type state struct {
	LastTemplatePath string
}

type stateStore struct {
	path string
}

func (ss *stateStore) load() (*state, error) {

}

func (ss *stateStore) save(s *state) error {

}
