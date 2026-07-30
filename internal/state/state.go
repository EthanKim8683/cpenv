package state

import (
	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
)

type State struct {
	FocusedProblem       *problemv1.Problem
	LastUsedTemplateName string
}
