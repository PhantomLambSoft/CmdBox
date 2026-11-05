package shellinit

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUnknownShellError_Message(t *testing.T) {
	err := &UnknownShellError{Name: "tcsh", Valid: []string{"bash", "zsh"}}
	require.Equal(t, "invalid shell: tcsh (expected one of :bash, zsh", err.Error())
}

func TestUnknownShellError_EmptyValidList(t *testing.T) {
	err := &UnknownShellError{Name: "tcsh", Valid: nil}
	require.Equal(t, "invalid shell: tcsh (expected one of :", err.Error())
}

func TestNoDefaultPathError_Message(t *testing.T) {
	err := &NoDefaultPathError{Shell: "fish"}
	require.Equal(t, "no default install path available for shell: fish. Use --path", err.Error())
}

func TestErrUnterminatedBlock_Message(t *testing.T) {
	require.Contains(t, ErrUnterminatedBlock.Error(), "start marker")
}
