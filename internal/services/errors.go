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
	ErrImportFile          = errors.New("import file error")
	ErrUnsupportedVersion  = errors.New("unsupported import file version")
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

type ImportFileError struct {
	Path string
	Err  error
}

func (e *ImportFileError) Error() string {
	return fmt.Sprintf("could not read import file %q: %s", e.Path, e.Err)
}

func (e *ImportFileError) Unwrap() error {
	return e.Err
}

func (e *ImportFileError) Is(target error) bool {
	return target == ErrImportFile
}
