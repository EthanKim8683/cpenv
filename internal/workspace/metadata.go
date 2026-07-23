package workspace

import (
	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
)

const metadataFile = ".metadata.json"

type metadata struct {
	ScaffoldFile string
	ScaffoldHash string
	Problem      problemv1.Problem
}
