package cli

import (
	"context"

	"github.com/EthanKim8683/cpenv/internal/config"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	statusv1 "github.com/EthanKim8683/cpenv/internal/gen/status/v1"
	bolt "go.etcd.io/bbolt"
)

type FocusLoader interface {
	Load() (*problemv1.Problem, error)
}

type SubmissionsTailer interface {
	Tail(limit int) ([]*statusv1.Submission, error)
	TailProblem(problemID string, limit int) ([]*statusv1.Submission, error)
}

type SubmitRequester interface {
	Request(ctx context.Context, problemID string, fileName string, content []byte) error
}

// TODO: fix naming
type CLI struct {
	Cfg             *config.Config
	CWD             string
	DB              *bolt.DB
	FocusLoader     FocusLoader
	Submissions     SubmissionsTailer
	SubmitRequester SubmitRequester
}
