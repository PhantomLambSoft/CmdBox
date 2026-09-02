package services

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmptyFieldSelection = errors.New("no fields specified")
	ErrUnknownHistoryIndex = errors.New("no history at this index")
	ErrNoTargetProfile     = errors.New("no target profile specified")
	ErrImportCycle         = errors.New("import cycle detected")
	ErrImportFile          = errors.New("import file not found")
)

type UnknownFieldError struct {
	Unknown string
	Allowed []string
	Context string
}

func (e *UnknownFieldError) Error() string {
	msg := fmt.Sprintf("unknown field %q", e.Unknown)
	if e.Context != "" {
		msg += fmt.Sprintf(" (%s)", e.Context)
	}
	msg += fmt.Sprintf(", allowed fields: %s", strings.Join(e.Allowed, ", "))
	return msg
}
