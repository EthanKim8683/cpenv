package main

import (
	"fmt"
	"os"

	"github.com/EthanKim8683/cpenv/cmd/cpenv/command"
)

func main() {
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "cpenv: %v\n", err)
		os.Exit(1)
	}
}
