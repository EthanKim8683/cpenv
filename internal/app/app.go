package app

import (
	"net/http"

	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/state"
)

type App struct {
	cfg        *config.Config
	stateStore state.Store
	httpClient *http.Client
}

func NewApp(cfg *config.Config, httpClient *http.Client) *App {
	return &App{
		cfg:        cfg,
		stateStore: state.NewFileStore(cfg.StatePath),
		httpClient: httpClient,
	}
}
