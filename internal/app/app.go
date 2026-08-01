package app

import (
	"net/http"

	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/state"
)

type App struct {
	cfg        *config.Config
	store      state.Store
	httpClient *http.Client
}
