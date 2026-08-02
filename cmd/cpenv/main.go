package main

import (
	"os"

	"github.com/EthanKim8683/cpenv/cmd/cpenv/command"
)

func main() {
	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
