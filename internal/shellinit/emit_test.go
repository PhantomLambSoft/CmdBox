package shellinit

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteEmit_WritesToStdoutWhenNoEnv(t *testing.T) {
	var buf bytes.Buffer
	err := WriteEmit("echo hi", &buf, func(string) string { return "" }, func(string, []byte, fs.FileMode) error {
		t.Fatal("writeFile should not be called")
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, "echo hi\n", buf.String())
}

func TestWriteEmit_WritesToFileWhenEnvSet(t *testing.T) {
	var buf bytes.Buffer
	var gotPath string
	var gotData []byte
	var gotPerm fs.FileMode

	err := WriteEmit("echo hi", &buf, func(key string) string {
		require.Equal(t, EmitFileEnv, key)
		return "/tmp/emit-file"
	}, func(path string, data []byte, perm fs.FileMode) error {
		gotPath, gotData, gotPerm = path, data, perm
		return nil
	})

	require.NoError(t, err)
	require.Empty(t, buf.String(), "stdout should not be used when emit file is set")
	require.Equal(t, "/tmp/emit-file", gotPath)
	require.Equal(t, "echo hi\n", string(gotData))
	require.Equal(t, fs.FileMode(0o600), gotPerm)
}

func TestWriteEmit_TrimsSurroundingNewlines(t *testing.T) {
	var buf bytes.Buffer
	err := WriteEmit("\n\necho hi\n\n", &buf, func(string) string { return "" }, nil)
	require.NoError(t, err)
	require.Equal(t, "echo hi\n", buf.String())
}

func TestWriteEmit_PropagatesWriteFileError(t *testing.T) {
	wantErr := errors.New("disk full")
	err := WriteEmit("echo hi", &bytes.Buffer{}, func(string) string { return "/tmp/x" }, func(string, []byte, fs.FileMode) error {
		return wantErr
	})
	require.ErrorIs(t, err, wantErr)
}

func TestWriteEmit_PropagatesStdoutWriteError(t *testing.T) {
	err := WriteEmit("echo hi", failingWriter{}, func(string) string { return "" }, nil)
	require.Error(t, err)
}

func TestEmit_WritesToRealStdoutWhenEnvUnset(t *testing.T) {
	t.Setenv(EmitFileEnv, "")
	require.NoError(t, os.Unsetenv(EmitFileEnv))

	r, w, err := os.Pipe()
	require.NoError(t, err)
	origStdout := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	err = Emit("echo hi")
	require.NoError(t, w.Close())
	os.Stdout = origStdout

	out, readErr := io.ReadAll(r)
	require.NoError(t, readErr)
	require.NoError(t, err)
	require.Equal(t, "echo hi\n", string(out))
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errors.New("broken pipe")
}
