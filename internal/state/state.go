package state

import focusv1 "github.com/EthanKim8683/cpenv/gen/focus/v1"

type State struct {
	Focus                *focusv1.Focus
	LastUsedTemplateName string
}
