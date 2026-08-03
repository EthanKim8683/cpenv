package app

import (
	"net/http"

	"github.com/EthanKim8683/cpenv/internal/config"
)

type App struct {
	Cfg        *config.Config
	HTTPClient *http.Client
	WorkingDir string
}
