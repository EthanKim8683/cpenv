package workspace

import problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"

const stateFile = ".cpenv-state.json"

type state struct {
	ScaffoldFile string
	Problem      problemv1.Problem
}
