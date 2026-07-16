package cpenv

import (
	"fmt"
	"os"

	"github.com/EthanKim8683/cpenv/cmd/cpenv/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "cpenv: ", err)
		os.Exit(1)
	}
}
