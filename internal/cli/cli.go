package cli

import (
	"context"

	"github.com/EthanKim8683/cpenv/internal/config"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	submissionv1 "github.com/EthanKim8683/cpenv/internal/gen/submission/v1"
	submitv1 "github.com/EthanKim8683/cpenv/internal/gen/submit/v1"
	bolt "go.etcd.io/bbolt"
)

type FocusedProblemLoader interface {
	Load() (*problemv1.Problem, error)
}

type SubmissionTailer interface {
	Tail(limit int) ([]*submissionv1.Submission, error)
	TailProblem(problemID string, limit int) ([]*submissionv1.Submission, error)
}

type Submitter interface {
	Submit(ctx context.Context, req *submitv1.SubmitRequest) (*submitv1.SubmitResponse, error)
}

type CLI struct {
	Cfg            *config.Config
	CWD            string
	DB             *bolt.DB
	FocusedProblem FocusedProblemLoader
	Submissions    SubmissionTailer
	Submitter      Submitter
}
