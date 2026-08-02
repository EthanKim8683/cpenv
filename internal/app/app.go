package app

import (
	"net/http"

	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/state"
)

type App struct {
	Cfg        *config.Config
	StateStore state.Store
	HTTPClient *http.Client
	WorkingDir string
}
