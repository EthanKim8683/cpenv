package app

// func TestDefaultTemplate(t *testing.T) {
// 	t.Parallel()

// 	t.Run("state template", func(t *testing.T) {
// 		t.Parallel()

// 		statePath := filepath.Join(t.TempDir(), "state.json")
// 		stateTmpl := "state"

// 		stateStore := state.NewFileStore(statePath)
// 		stateStore.Save(&state.State{
// 			Template: stateTmpl,
// 		})

// 		a := &App{
// 			StateStore: stateStore,
// 		}

// 		tmpl, err := a.defaultTemplate()
// 		assert.NoError(t, err)
// 		assert.Equal(t, stateTmpl, tmpl)
// 	})

// 	t.Run("empty state template", func(t *testing.T) {
// 		t.Parallel()

// 		statePath := filepath.Join(t.TempDir(), "state.json")
// 		tmplDir := filepath.Join("testdata", "known-templates")

// 		stateStore := state.NewFileStore(statePath)

// 		a := &App{
// 			Cfg: &config.Config{
// 				TemplatesDir: tmplDir,
// 			},
// 			StateStore: stateStore,
// 		}

// 		tmpl, err := a.defaultTemplate()
// 		assert.NoError(t, err)
// 		assert.True(t, slices.Contains([]string{
// 			filepath.Join(tmplDir, "1.star"),
// 			filepath.Join(tmplDir, "2.star"),
// 		}, tmpl))
// 	})
// }

// func TestResolveTemplate(t *testing.T) {
// 	t.Parallel()

// 	t.Run("absolute path", func(t *testing.T) {
// 		t.Parallel()

// 		absTmpl := filepath.Join(t.TempDir(), "absolute.star")

// 		a := &App{}

// 		tmpl, err := a.resolveTemplate(absTmpl)
// 		assert.NoError(t, err)
// 		assert.Equal(t, absTmpl, tmpl)
// 	})

// 	t.Run("relative path", func(t *testing.T) {
// 		t.Parallel()

// 		tmplDir := filepath.Join(t.TempDir(), "relative")
// 		relTmpl := "relative.star"

// 		a := &App{
// 			Cfg: &config.Config{
// 				TemplatesDir: tmplDir,
// 			},
// 		}

// 		tmpl, err := a.resolveTemplate(relTmpl)
// 		assert.NoError(t, err)
// 		assert.Equal(t, filepath.Join(tmplDir, relTmpl), tmpl)
// 	})

// 	t.Run("empty path", func(t *testing.T) {
// 		t.Parallel()

// 		tmplDir := filepath.Join("testdata", "known-templates")
// 		statePath := filepath.Join(t.TempDir(), "state.json")
// 		stateTmpl := "state"

// 		stateStore := state.NewFileStore(statePath)
// 		stateStore.Save(&state.State{
// 			Template: stateTmpl,
// 		})

// 		a := &App{
// 			Cfg: &config.Config{
// 				TemplatesDir: tmplDir,
// 			},
// 			StateStore: stateStore,
// 		}

// 		tmpl, err := a.resolveTemplate("")
// 		assert.NoError(t, err)
// 		assert.Equal(t, stateTmpl, tmpl)
// 	})
// }
