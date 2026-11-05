package shellinit

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnv_env_NilGetEnvReturnsEmpty(t *testing.T) {
	e := Env{}
	require.Equal(t, "", e.env("PATH"))
}

func TestEnv_env_DelegatesToGetEnv(t *testing.T) {
	e := Env{GetEnv: envWith(map[string]string{"FOO": "bar"})}
	require.Equal(t, "bar", e.env("FOO"))
}

func TestSystemEnv_PopulatesFields(t *testing.T) {
	env := SystemEnv()
	require.Equal(t, runtime.GOOS, env.GOOS)
	require.NotNil(t, env.GetEnv)
	require.NotNil(t, env.Output)
	require.NotNil(t, env.Parent)
	require.Greater(t, env.Ppid, 0)
}

func TestRunOutput_Success(t *testing.T) {
	// "go" is guaranteed to be on PATH in this test environment; this is a
	// read-only invocation with no side effects on the system.
	out, err := runOutput("go", "version")
	require.NoError(t, err)
	require.Contains(t, out, "go version")
}

func TestRunOutput_UnknownCommandReturnsError(t *testing.T) {
	_, err := runOutput("cmdbox-definitely-not-a-real-binary-xyz")
	require.Error(t, err)
}
