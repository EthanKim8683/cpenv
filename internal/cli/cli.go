package cli

import (
	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/gen/active_problem/v1/active_problemv1connect"
	focusv1connect "github.com/EthanKim8683/cpenv/internal/gen/focus/v1/focusv1connect"
	"github.com/EthanKim8683/cpenv/internal/gen/submissions/v1/submissionsv1connect"
	"github.com/EthanKim8683/cpenv/internal/gen/submit/v1/submitv1connect"
)

type CLI struct {
	Cfg                 *config.Config
	CWD                 string
	ActiveProblemClient active_problemv1connect.ActiveProblemServiceClient
	FocusClient         focusv1connect.FocusServiceClient
	SubmissionsClient   submissionsv1connect.SubmissionsServiceClient
	SubmitClient        submitv1connect.SubmitServiceClient
	Preferences         Preferences
}
