package extension

import "fmt"

type ExtensionError struct {
	Msg string
}

func (e *ExtensionError) Error() string {
	return fmt.Sprintf("extension error: %s", e.Msg)
}
