package shellinit

import (
	"errors"
	"fmt"
	"strings"
)

type UnknownShellError struct {
	Name  string
	Valid []string
}

func (e *UnknownShellError) Error() string {
	return fmt.Sprintf("invalid shell: %s (expected one of :%s", e.Name, strings.Join(e.Valid, ", "))
}

type NoDefaultPathError struct {
	Shell string
}

func (e *NoDefaultPathError) Error() string {
	return fmt.Sprintf("no default install path available for shell: %s. Use --path", e.Shell)
}

var ErrUnterminatedBlock = errors.New(
	"found the CmdBox start marker without its end marker; remove the partial block by hand and retry",
)
