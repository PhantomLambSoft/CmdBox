//go:build !linux && !windows

package shellinit

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLookupProcess_CurrentProcess(t *testing.T) {
	ppid, name, err := lookupProcess(os.Getpid())
	require.NoError(t, err)
	require.Equal(t, os.Getppid(), ppid)
	require.NotEmpty(t, name)
}

func TestLookupProcess_UnknownPidReturnsError(t *testing.T) {
	_, _, err := lookupProcess(1 << 30)
	require.Error(t, err)
}
