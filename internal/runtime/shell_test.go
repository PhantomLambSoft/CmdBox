package runtime

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// mkExe creates an empty file inside dir with the given name so it can be
// discovered by which()/exec.LookPath without touching anything outside dir.
func mkExe(t *testing.T, dir, name string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("echo"), 0o755); err != nil {
		t.Fatalf("write fake exe %s: %v", path, err)
	}
}

// unsetEnv unsets key for the duration of the test and restores its previous
// value (or absence) afterward.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	old, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv(key, old)
		} else {
			_ = os.Unsetenv(key)
		}
	})
}

// isolatePath points PATH at an empty (or fake-exe-populated) temp directory
// so which() only ever sees what the test explicitly places there.
func isolatePath(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("PATH", dir)
	return dir
}

// --- which ---

func TestWhichFindsExecutableOnPath(t *testing.T) {
	dir := isolatePath(t)
	mkExe(t, dir, "myfake.exe")

	if !which("myfake") {
		t.Fatal("which(\"myfake\") = false, want true (should resolve via PATH)")
	}
}

func TestWhichFindsAbsolutePathThatExists(t *testing.T) {
	dir := t.TempDir()
	mkExe(t, dir, "myfake.exe")
	path := filepath.Join(dir, "myfake.exe")

	if !which(path) {
		t.Fatalf("which(%q) = false, want true", path)
	}
}

func TestWhichReturnsFalseWhenNotFound(t *testing.T) {
	isolatePath(t)

	if which("definitely-not-a-real-cmdbox-test-executable") {
		t.Fatal("which() = true, want false for nonexistent executable")
	}
}

func TestWhichReturnsFalseForNonexistentAbsolutePath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.exe")

	if which(path) {
		t.Fatalf("which(%q) = true, want false", path)
	}
}

// --- windowsShellArgs ---

func TestWindowsShellArgs(t *testing.T) {
	tests := []struct {
		name    string
		shell   string
		command string
		want    []string
	}{
		{"pwsh lowercase", "pwsh", "echo hi", []string{"pwsh", "-NoProfile", "-Command", "echo hi"}},
		{"pwsh mixed case", "PWsh", "echo hi", []string{"pwsh", "-NoProfile", "-Command", "echo hi"}},
		{"powershell lowercase", "powershell", "echo hi", []string{"powershell", "-NoProfile", "-Command", "echo hi"}},
		{"powershell mixed case", "PowerShell", "echo hi", []string{"powershell", "-NoProfile", "-Command", "echo hi"}},
		{"cmd short form", "cmd", "dir", []string{"cmd.exe", "/C", "dir"}},
		{"cmd.exe", "cmd.exe", "dir", []string{"cmd.exe", "/C", "dir"}},
		{"CMD.EXE uppercase", "CMD.EXE", "dir", []string{"cmd.exe", "/C", "dir"}},
		{"unknown shell treated as literal executable", "wsl", "ls", []string{"wsl", "ls"}},
		{"unknown shell with absolute path", `C:\Git\bin\bash.exe`, "ls", []string{`c:\git\bin\bash.exe`, "ls"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := windowsShellArgs(tt.shell, tt.command)
			if !equalSlices(got, tt.want) {
				t.Fatalf("windowsShellArgs(%q, %q) = %v, want %v", tt.shell, tt.command, got, tt.want)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- BuildShellCommand (Windows branch) ---

func skipIfNotWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("windows-specific behavior; skipping on non-windows host")
	}
}

func TestBuildShellCommandWindowsNoShellsAvailableFallsBackToCmd(t *testing.T) {
	skipIfNotWindows(t)
	isolatePath(t)
	unsetEnv(t, "CMDBOX_SHELL")

	got, err := BuildShellCommand("dir", "")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}
	want := []string{"cmd.exe", "/C", "dir"}
	if !equalSlices(got, want) {
		t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
	}
}

func TestBuildShellCommandWindowsPreferredShellTakesPriority(t *testing.T) {
	skipIfNotWindows(t)
	dir := isolatePath(t)
	unsetEnv(t, "CMDBOX_SHELL")
	mkExe(t, dir, "pwsh.exe")
	mkExe(t, dir, "cmd.exe")

	got, err := BuildShellCommand("echo hi", "pwsh")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}
	want := []string{"pwsh", "-NoProfile", "-Command", "echo hi"}
	if !equalSlices(got, want) {
		t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
	}
}

func TestBuildShellCommandWindowsPreferredShellUnavailableFallsThrough(t *testing.T) {
	skipIfNotWindows(t)
	dir := isolatePath(t)
	unsetEnv(t, "CMDBOX_SHELL")
	mkExe(t, dir, "powershell.exe")

	got, err := BuildShellCommand("echo hi", "pwsh")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}
	want := []string{"powershell", "-NoProfile", "-Command", "echo hi"}
	if !equalSlices(got, want) {
		t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
	}
}

func TestBuildShellCommandWindowsEnvShellUsedWhenNoPreferred(t *testing.T) {
	skipIfNotWindows(t)
	dir := isolatePath(t)
	mkExe(t, dir, "powershell.exe")
	t.Setenv("CMDBOX_SHELL", "powershell")

	got, err := BuildShellCommand("echo hi", "")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}
	want := []string{"powershell", "-NoProfile", "-Command", "echo hi"}
	if !equalSlices(got, want) {
		t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
	}
}

func TestBuildShellCommandWindowsEnvShellTakesPriorityOverFallbacks(t *testing.T) {
	skipIfNotWindows(t)
	dir := isolatePath(t)
	mkExe(t, dir, "mytool.exe")
	mkExe(t, dir, "pwsh.exe")
	t.Setenv("CMDBOX_SHELL", "mytool")

	got, err := BuildShellCommand("echo hi", "")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}
	want := []string{"mytool", "echo hi"}
	if !equalSlices(got, want) {
		t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
	}
}

func TestBuildShellCommandWindowsFallbackOrderPwshBeforePowershellBeforeCmd(t *testing.T) {
	skipIfNotWindows(t)
	dir := isolatePath(t)
	unsetEnv(t, "CMDBOX_SHELL")
	mkExe(t, dir, "powershell.exe")
	mkExe(t, dir, "cmd.exe")

	got, err := BuildShellCommand("echo hi", "")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}
	want := []string{"powershell", "-NoProfile", "-Command", "echo hi"}
	if !equalSlices(got, want) {
		t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
	}
}

func TestBuildShellCommandWindowsFallsBackToCmdWhenOnlyCmdAvailable(t *testing.T) {
	skipIfNotWindows(t)
	dir := isolatePath(t)
	unsetEnv(t, "CMDBOX_SHELL")
	mkExe(t, dir, "cmd.exe")

	got, err := BuildShellCommand("dir", "")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}
	want := []string{"cmd.exe", "/C", "dir"}
	if !equalSlices(got, want) {
		t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
	}
}

func TestBuildShellCommandWindowsCustomPreferredShellNotInKnownList(t *testing.T) {
	skipIfNotWindows(t)
	dir := isolatePath(t)
	unsetEnv(t, "CMDBOX_SHELL")
	mkExe(t, dir, "wsl.exe")

	got, err := BuildShellCommand("ls", "wsl")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}
	want := []string{"wsl", "ls"}
	if !equalSlices(got, want) {
		t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
	}
}

func TestBuildShellCommandWindowsNeverErrors(t *testing.T) {
	skipIfNotWindows(t)
	isolatePath(t)
	unsetEnv(t, "CMDBOX_SHELL")

	// Even with nothing on PATH, the windows branch always has a last-resort
	// return and should never surface ErrNoUsableShell.
	if _, err := BuildShellCommand("dir", ""); err != nil {
		t.Fatalf("BuildShellCommand() error = %v, want nil", err)
	}
}

// --- BuildShellCommand (Unix branch) ---

func skipIfWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("unix-specific behavior; skipping on windows host")
	}
}

func TestBuildShellCommandUnixPreferredShellTakesPriority(t *testing.T) {
	skipIfWindows(t)
	dir := isolatePath(t)
	unsetEnv(t, "SHELL")
	mkExe(t, dir, "zsh")
	mkExe(t, dir, "bash")

	got, err := BuildShellCommand("echo hi", "zsh")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}
	want := []string{"zsh", "-lc", "echo hi"}
	if !equalSlices(got, want) {
		t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
	}
}

func TestBuildShellCommandUnixPreferredShellUnavailableFallsThrough(t *testing.T) {
	skipIfWindows(t)
	dir := isolatePath(t)
	unsetEnv(t, "SHELL")
	mkExe(t, dir, "bash")

	got, err := BuildShellCommand("echo hi", "zsh")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}
	want := []string{"bash", "-lc", "echo hi"}
	if !equalSlices(got, want) {
		t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
	}
}

func TestBuildShellCommandUnixEnvShellUsedWhenNoPreferred(t *testing.T) {
	skipIfWindows(t)
	dir := isolatePath(t)
	envShellPath := filepath.Join(dir, "myshell")
	mkExe(t, dir, "myshell")
	t.Setenv("SHELL", envShellPath)

	got, err := BuildShellCommand("echo hi", "")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}
	want := []string{envShellPath, "-lc", "echo hi"}
	if !equalSlices(got, want) {
		t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
	}
}

func TestBuildShellCommandUnixFallsBackToBinBashWhenAbsolute(t *testing.T) {
	skipIfWindows(t)
	isolatePath(t)
	unsetEnv(t, "SHELL")

	got, err := BuildShellCommand("echo hi", "")
	if err != nil {
		t.Fatalf("BuildShellCommand() error = %v", err)
	}

	if _, statErr := os.Stat("/bin/bash"); statErr == nil {
		want := []string{"/bin/bash", "-lc", "echo hi"}
		if !equalSlices(got, want) {
			t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
		}
	} else if _, statErr := os.Stat("/bin/sh"); statErr == nil {
		want := []string{"/bin/sh", "-c", "echo hi"}
		if !equalSlices(got, want) {
			t.Fatalf("BuildShellCommand() = %v, want %v", got, want)
		}
	}
}

func TestBuildShellCommandUnixReturnsErrorWhenNoShellAvailable(t *testing.T) {
	skipIfWindows(t)
	if _, statErr := os.Stat("/bin/bash"); statErr == nil {
		t.Skip("cannot simulate missing shells: /bin/bash exists on this host and BuildShellCommand checks it directly")
	}
	if _, statErr := os.Stat("/bin/sh"); statErr == nil {
		t.Skip("cannot simulate missing shells: /bin/sh exists on this host and BuildShellCommand checks it directly")
	}
	isolatePath(t)
	unsetEnv(t, "SHELL")

	_, err := BuildShellCommand("echo hi", "")
	if !errors.Is(err, ErrNoUsableShell) {
		t.Fatalf("BuildShellCommand() error = %v, want ErrNoUsableShell", err)
	}
}
