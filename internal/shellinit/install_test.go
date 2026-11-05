package shellinit

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeFS is an in-memory Filesystem for tests, avoiding any real I/O.
type fakeFS struct {
	files map[string][]byte

	copyCalls   []string
	writeCalls  []string
	renameCalls []string
}

func newFakeFS() *fakeFS {
	return &fakeFS{files: map[string][]byte{}}
}

func (f *fakeFS) ReadFile(path string) ([]byte, error) {
	data, ok := f.files[path]
	if !ok {
		return nil, fs.ErrNotExist
	}
	return data, nil
}

func (f *fakeFS) WriteFile(path string, data []byte, _ fs.FileMode) error {
	f.writeCalls = append(f.writeCalls, path)
	f.files[path] = append([]byte(nil), data...)
	return nil
}

func (f *fakeFS) CopyFile(src, dest string) error {
	f.copyCalls = append(f.copyCalls, src+"->"+dest)
	data, ok := f.files[src]
	if !ok {
		return fs.ErrNotExist
	}
	f.files[dest] = append([]byte(nil), data...)
	return nil
}

func (f *fakeFS) Rename(oldPath, newPath string) error {
	f.renameCalls = append(f.renameCalls, oldPath+"->"+newPath)
	data, ok := f.files[oldPath]
	if !ok {
		return fs.ErrNotExist
	}
	f.files[newPath] = data
	delete(f.files, oldPath)
	return nil
}

type erroringFS struct {
	*fakeFS
	readErr  error
	copyErr  error
	writeErr error
}

func (f *erroringFS) ReadFile(path string) ([]byte, error) {
	if f.readErr != nil {
		return nil, f.readErr
	}
	return f.fakeFS.ReadFile(path)
}

func (f *erroringFS) CopyFile(src, dest string) error {
	if f.copyErr != nil {
		return f.copyErr
	}
	return f.fakeFS.CopyFile(src, dest)
}

func (f *erroringFS) WriteFile(path string, data []byte, perm fs.FileMode) error {
	if f.writeErr != nil {
		return f.writeErr
	}
	return f.fakeFS.WriteFile(path, data, perm)
}

func newInitializer(env Env, fs Filesystem) *Initializer {
	return &Initializer{Env: env, FS: fs}
}

func TestNew_ReturnsInitializerWithSystemDefaults(t *testing.T) {
	init := New()
	require.NotNil(t, init)
	require.IsType(t, OSFS{}, init.FS)
	require.NotNil(t, init.Env.GetEnv)
}

func TestInitializer_DetectShell(t *testing.T) {
	env := Env{GOOS: "linux", GetEnv: envWith(map[string]string{"SHELL": "/bin/zsh"})}
	init := newInitializer(env, newFakeFS())
	require.Equal(t, "zsh", init.DetectShell())
}

func TestInitializer_Resolve_ExplicitShell(t *testing.T) {
	init := newInitializer(Env{}, newFakeFS())
	spec, err := init.Resolve("fish")
	require.NoError(t, err)
	require.Equal(t, "fish", spec.Name)
}

func TestInitializer_Resolve_DetectsWhenEmpty(t *testing.T) {
	env := Env{GOOS: "linux", GetEnv: envWith(map[string]string{"SHELL": "/bin/bash"})}
	init := newInitializer(env, newFakeFS())
	spec, err := init.Resolve("   ")
	require.NoError(t, err)
	require.Equal(t, "bash", spec.Name)
}

func TestInitializer_Resolve_UnknownShell(t *testing.T) {
	init := newInitializer(Env{}, newFakeFS())
	_, err := init.Resolve("tcsh")
	require.Error(t, err)
}

func TestInitializer_Snippet(t *testing.T) {
	init := newInitializer(Env{}, newFakeFS())
	spec, err := Lookup("bash")
	require.NoError(t, err)
	text, err := init.Snippet(spec)
	require.NoError(t, err)
	require.NotEmpty(t, text)
}

func TestInitializer_TargetPath_Override(t *testing.T) {
	init := newInitializer(Env{Home: "/home/alice"}, newFakeFS())
	path, err := init.TargetPath(Spec{Display: "Bash"}, "/custom/path")
	require.NoError(t, err)
	require.Equal(t, "/custom/path", path)
}

func TestInitializer_TargetPath_OverrideExpandsHome(t *testing.T) {
	init := newInitializer(Env{Home: "/home/alice"}, newFakeFS())

	path, err := init.TargetPath(Spec{}, "~")
	require.NoError(t, err)
	require.Equal(t, "/home/alice", path)

	path, err = init.TargetPath(Spec{}, "~/bashrc")
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/home/alice", "bashrc"), path)
}

func TestInitializer_TargetPath_OverrideNoHomeLeavesTildeAsIs(t *testing.T) {
	init := newInitializer(Env{Home: ""}, newFakeFS())
	path, err := init.TargetPath(Spec{}, "~/bashrc")
	require.NoError(t, err)
	require.Equal(t, "~/bashrc", path)
}

func TestInitializer_TargetPath_UsesDefaultPathFunc(t *testing.T) {
	init := newInitializer(Env{Home: "/home/alice"}, newFakeFS())
	spec, err := Lookup("bash")
	require.NoError(t, err)
	// override GOOS via env stored on the initializer
	init.Env.GOOS = "linux"
	path, err := init.TargetPath(spec, "")
	require.NoError(t, err)
	require.Equal(t, filepath.Join("/home/alice", ".bashrc"), path)
}

func TestInitializer_TargetPath_NoDefaultPath(t *testing.T) {
	init := newInitializer(Env{}, newFakeFS())
	_, err := init.TargetPath(Spec{Display: "CMD"}, "")
	var noPath *NoDefaultPathError
	require.ErrorAs(t, err, &noPath)
	require.Equal(t, "CMD", noPath.Shell)
}

func TestInstall_NewFileIsCreated(t *testing.T) {
	fsys := newFakeFS()
	init := newInitializer(Env{}, fsys)
	spec := Spec{Display: "Bash", Mode: ModeProfileBlock, DefaultPath: func(Env) (string, error) { return "/home/alice/.bashrc", nil }}

	res, err := init.Install(spec, "export FOO=bar", "")
	require.NoError(t, err)
	require.True(t, res.Changed)
	require.Empty(t, res.BackupPath)
	require.Equal(t, "/home/alice/.bashrc", res.Path)
	require.Empty(t, fsys.copyCalls, "no backup should be made for a new file")

	got := string(fsys.files["/home/alice/.bashrc"])
	require.Contains(t, got, "export FOO=bar")
	require.Contains(t, got, StartMark)
}

func TestInstall_ExistingFileGetsBackedUpAndUpdated(t *testing.T) {
	fsys := newFakeFS()
	fsys.files["/home/alice/.bashrc"] = []byte("alias ll='ls -la'\n")
	init := newInitializer(Env{}, fsys)
	spec := Spec{Display: "Bash", Mode: ModeProfileBlock, DefaultPath: func(Env) (string, error) { return "/home/alice/.bashrc", nil }}

	res, err := init.Install(spec, "export FOO=bar", "")
	require.NoError(t, err)
	require.True(t, res.Changed)
	require.Equal(t, "/home/alice/.bashrc.bak", res.BackupPath)
	require.Equal(t, []string{"/home/alice/.bashrc->/home/alice/.bashrc.bak"}, fsys.copyCalls)

	backup := string(fsys.files["/home/alice/.bashrc.bak"])
	require.Equal(t, "alias ll='ls -la'\n", backup)

	updated := string(fsys.files["/home/alice/.bashrc"])
	require.Contains(t, updated, "alias ll='ls -la'")
	require.Contains(t, updated, "export FOO=bar")
}

func TestInstall_SecondInstallIsNoOp(t *testing.T) {
	fsys := newFakeFS()
	init := newInitializer(Env{}, fsys)
	spec := Spec{Display: "Bash", Mode: ModeProfileBlock, DefaultPath: func(Env) (string, error) { return "/home/alice/.bashrc", nil }}

	_, err := init.Install(spec, "export FOO=bar", "")
	require.NoError(t, err)

	fsys.copyCalls = nil
	fsys.writeCalls = nil

	res, err := init.Install(spec, "export FOO=bar", "")
	require.NoError(t, err)
	require.False(t, res.Changed)
	require.Empty(t, res.BackupPath)
	require.Empty(t, fsys.copyCalls)
	require.Empty(t, fsys.writeCalls)
}

func TestInstall_WriteFileModeIgnoresExistingContent(t *testing.T) {
	fsys := newFakeFS()
	fsys.files["/home/alice/.config/fish/functions/cbe.fish"] = []byte("old content")
	init := newInitializer(Env{}, fsys)
	spec := Spec{
		Display:     "Fish",
		Mode:        ModeWriteFile,
		DefaultPath: func(Env) (string, error) { return "/home/alice/.config/fish/functions/cbe.fish", nil },
	}

	res, err := init.Install(spec, "function cbe\nend\n", "")
	require.NoError(t, err)
	require.True(t, res.Changed)
	require.Equal(t, "function cbe\nend\n", string(fsys.files["/home/alice/.config/fish/functions/cbe.fish"]))
}

func TestInstall_WrapperHintModeCannotInstall(t *testing.T) {
	fsys := newFakeFS()
	init := newInitializer(Env{}, fsys)
	spec := Spec{Display: "CMD", Mode: ModeWrapperHint, DefaultPath: func(Env) (string, error) { return "/some/path", nil }}

	_, err := init.Install(spec, "irrelevant", "")
	require.Error(t, err)
	require.Empty(t, fsys.writeCalls)
}

func TestInstall_TargetPathError(t *testing.T) {
	init := newInitializer(Env{}, newFakeFS())
	spec := Spec{Display: "Bash", Mode: ModeProfileBlock}

	_, err := init.Install(spec, "irrelevant", "")
	var noPath *NoDefaultPathError
	require.ErrorAs(t, err, &noPath)
}

func TestInstall_UnterminatedBlockError(t *testing.T) {
	fsys := newFakeFS()
	fsys.files["/home/alice/.bashrc"] = []byte("before\n" + StartMark + "\nno end marker\n")
	init := newInitializer(Env{}, fsys)
	spec := Spec{Display: "Bash", Mode: ModeProfileBlock, DefaultPath: func(Env) (string, error) { return "/home/alice/.bashrc", nil }}

	_, err := init.Install(spec, "export FOO=bar", "")
	require.ErrorIs(t, err, ErrUnterminatedBlock)
	require.Empty(t, fsys.writeCalls)
}

func TestInstall_ReadFileErrorPropagates(t *testing.T) {
	fsys := &erroringFS{fakeFS: newFakeFS(), readErr: os.ErrPermission}
	init := newInitializer(Env{}, fsys)
	spec := Spec{Display: "Bash", Mode: ModeProfileBlock, DefaultPath: func(Env) (string, error) { return "/home/alice/.bashrc", nil }}

	_, err := init.Install(spec, "export FOO=bar", "")
	require.ErrorIs(t, err, os.ErrPermission)
}

func TestInstall_CopyFileErrorPropagatesBeforeWrite(t *testing.T) {
	fsys := &erroringFS{fakeFS: newFakeFS(), copyErr: os.ErrPermission}
	fsys.files["/home/alice/.bashrc"] = []byte("existing\n")
	init := newInitializer(Env{}, fsys)
	spec := Spec{Display: "Bash", Mode: ModeProfileBlock, DefaultPath: func(Env) (string, error) { return "/home/alice/.bashrc", nil }}

	_, err := init.Install(spec, "export FOO=bar", "")
	require.ErrorIs(t, err, os.ErrPermission)
	require.Empty(t, fsys.writeCalls)
}

func TestInstall_WriteFileErrorPropagates(t *testing.T) {
	fsys := &erroringFS{fakeFS: newFakeFS(), writeErr: os.ErrPermission}
	init := newInitializer(Env{}, fsys)
	spec := Spec{Display: "Bash", Mode: ModeProfileBlock, DefaultPath: func(Env) (string, error) { return "/home/alice/.bashrc", nil }}

	res, err := init.Install(spec, "export FOO=bar", "")
	require.ErrorIs(t, err, os.ErrPermission)
	require.False(t, res.Changed)
}
