package app

func (a *App) Focus() error {
	problem, err := a.focusStore.Problem()
	if err != nil {
		return err
	}

	return nil
}
