package cli

import (
	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/gen/focus/v1/focusv1connect"
	"github.com/EthanKim8683/cpenv/internal/gen/status/v1/statusv1connect"
	"github.com/EthanKim8683/cpenv/internal/gen/submit/v1/submitv1connect"
)

type CLI struct {
	Cfg          *config.Config
	CWD          string
	FocusClient  focusv1connect.FocusServiceClient
	StatusClient statusv1connect.StatusServiceClient
	SubmitClient submitv1connect.SubmitServiceClient
	Preferences  Preferences
}
