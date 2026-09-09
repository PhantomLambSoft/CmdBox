package runtime

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// --- pointer helpers ---

func boolPtr(b bool) *bool                          { return &b }
func stringPtr(s string) *string                    { return &s }
func intPtr(i int) *int                             { return &i }
func envPtr(m map[string]string) *map[string]string { return &m }

// --- isMultiline ---

func TestIsMultiline(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{"single line", "echo hello", false},
		{"two lines", "echo a\necho b", true},
		{"trailing newline only", "echo hello\n", false},
		{"leading and trailing newlines only", "\n\necho hello\n\n", false},
		{"leading/trailing newlines around multiple lines", "\necho a\necho b\n", true},
		{"empty string", "", false},
		{"only newlines", "\n\n\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMultiline(tt.command); got != tt.want {
				t.Fatalf("isMultiline(%q) = %v, want %v", tt.command, got, tt.want)
			}
		})
	}
}

// --- mergeEnv ---

func TestMergeEnvNoOverridesReturnsBaseEnviron(t *testing.T) {
	t.Setenv("CMDBOX_TEST_MERGE_ENV", "base-value")

	got := mergeEnv(nil)

	found := false
	for _, kv := range got {
		if kv == "CMDBOX_TEST_MERGE_ENV=base-value" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("mergeEnv(nil) did not contain base environment variable CMDBOX_TEST_MERGE_ENV=base-value; got %v", got)
	}
}

func TestMergeEnvOverridesAddNewKey(t *testing.T) {
	got := mergeEnv(map[string]string{"CMDBOX_TEST_NEW_KEY": "new-value"})

	if !containsEnv(got, "CMDBOX_TEST_NEW_KEY", "new-value") {
		t.Fatalf("mergeEnv did not add override key; got %v", got)
	}
}

func TestMergeEnvOverridesReplaceExistingKey(t *testing.T) {
	t.Setenv("CMDBOX_TEST_OVERRIDE_KEY", "original-value")

	got := mergeEnv(map[string]string{"CMDBOX_TEST_OVERRIDE_KEY": "overridden-value"})

	if !containsEnv(got, "CMDBOX_TEST_OVERRIDE_KEY", "overridden-value") {
		t.Fatalf("mergeEnv did not override existing key; got %v", got)
	}
	count := 0
	for _, kv := range got {
		if strings.HasPrefix(kv, "CMDBOX_TEST_OVERRIDE_KEY=") {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("mergeEnv produced %d entries for overridden key, want 1", count)
	}
}

func containsEnv(env []string, key, value string) bool {
	for _, kv := range env {
		if kv == key+"="+value {
			return true
		}
	}
	return false
}

// --- scriptSuffixForShell ---

func TestScriptSuffixForShell(t *testing.T) {
	wantDefault := ".sh"
	if runtime.GOOS == "windows" {
		wantDefault = ".cmd"
	}

	tests := []struct {
		name  string
		shell string
		want  string
	}{
		{"default", "default", wantDefault},
		{"cmd", "cmd", ".cmd"},
		{"cmd.exe", "cmd.exe", ".cmd"},
		{"powershell", "powershell", ".ps1"},
		{"pwsh", "pwsh", ".ps1"},
		{"fish", "fish", ".fish"},
		{"bash", "bash", ".sh"},
		{"zsh", "zsh", ".sh"},
		{"unknown", "someothershell", ".sh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scriptSuffixForShell(tt.shell); got != tt.want {
				t.Fatalf("scriptSuffixForShell(%q) = %q, want %q", tt.shell, got, tt.want)
			}
		})
	}
}

// --- scriptHeaderForShell ---

func TestScriptHeaderForShell(t *testing.T) {
	tests := []struct {
		name  string
		shell string
		want  string
	}{
		{"bash", "bash", "#!/usr/bin/env bash\n"},
		{"zsh", "zsh", "#!/usr/bin/env zsh\n"},
		{"fish", "fish", "#!/usr/bin/env fish\n"},
		{"default", "default", ""},
		{"cmd", "cmd", ""},
		{"powershell", "powershell", ""},
		{"unknown", "someothershell", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := scriptHeaderForShell(tt.shell); got != tt.want {
				t.Fatalf("scriptHeaderForShell(%q) = %q, want %q", tt.shell, got, tt.want)
			}
		})
	}
}

// --- buildScriptExecArgs ---

func TestBuildScriptExecArgs(t *testing.T) {
	wantDefault := []string{"sh", "/tmp/script.sh"}
	if runtime.GOOS == "windows" {
		wantDefault = []string{"cmd.exe", "/d", "/s", "/c", "/tmp/script.sh"}
	}

	tests := []struct {
		name  string
		shell string
		want  []string
	}{
		{"default", "default", wantDefault},
		{"cmd", "cmd", []string{"cmd.exe", "/d", "/s", "/c", "/tmp/script.sh"}},
		{"cmd.exe", "cmd.exe", []string{"cmd.exe", "/d", "/s", "/c", "/tmp/script.sh"}},
		{"powershell", "powershell", []string{"powershell", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", "/tmp/script.sh"}},
		{"pwsh", "pwsh", []string{"pwsh", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", "/tmp/script.sh"}},
		{"fish", "fish", []string{"fish", "/tmp/script.sh"}},
		{"zsh", "zsh", []string{"zsh", "/tmp/script.sh"}},
		{"bash", "bash", []string{"bash", "/tmp/script.sh"}},
		{"unknown", "someothershell", []string{"sh", "/tmp/script.sh"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildScriptExecArgs("/tmp/script.sh", tt.shell)
			if !equalSlices(got, tt.want) {
				t.Fatalf("buildScriptExecArgs(%q, %q) = %v, want %v", "/tmp/script.sh", tt.shell, got, tt.want)
			}
		})
	}
}

// --- exitCodeFrom ---

func TestExitCodeFromNilErrorUsesProcessState(t *testing.T) {
	cmd := exitCodeCmd(0)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run exit-0 command: %v", err)
	}

	code, ok := exitCodeFrom(cmd, nil)
	if !ok {
		t.Fatal("exitCodeFrom() ok = false, want true")
	}
	if code != 0 {
		t.Fatalf("exitCodeFrom() code = %d, want 0", code)
	}
}

func TestExitCodeFromExitError(t *testing.T) {
	cmd := exitCodeCmd(1)
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected exit-1 command to return an error")
	}

	code, ok := exitCodeFrom(cmd, err)
	if !ok {
		t.Fatal("exitCodeFrom() ok = false, want true")
	}
	if code == 0 {
		t.Fatalf("exitCodeFrom() code = %d, want nonzero", code)
	}
}

// exitCodeCmd returns a *exec.Cmd that, when run, exits with the given code.
func exitCodeCmd(code int) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.Command("cmd.exe", "/C", "exit", strconv.Itoa(code))
	}
	return exec.Command("sh", "-c", "exit "+strconv.Itoa(code))
}

func TestExitCodeFromNonExitError(t *testing.T) {
	cmd := exec.Command("cmdbox-definitely-does-not-exist-binary")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected error for nonexistent binary")
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		t.Fatal("expected a non-ExitError (e.g. exec: not found)")
	}

	_, ok := exitCodeFrom(cmd, err)
	if ok {
		t.Fatal("exitCodeFrom() ok = true, want false for non-ExitError")
	}
}

// --- Executor.Run integration tests ---

func TestExecutorRunSimpleCommandCapturesStdout(t *testing.T) {
	e := NewExecutor()
	result, err := e.Run("echo hello", RunContext{Capture: boolPtr(true)})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result == nil {
		t.Fatal("Run() result = nil, want non-nil")
	}
	if result.ExitCode != 0 {
		t.Fatalf("Run() ExitCode = %d, want 0 (stderr=%q)", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "hello") {
		t.Fatalf("Run() Stdout = %q, want to contain %q", result.Stdout, "hello")
	}
}

func TestExecutorRunNonZeroExitCode(t *testing.T) {
	e := NewExecutor()
	command := "exit 3"
	result, err := e.Run(command, RunContext{Capture: boolPtr(true)})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.ExitCode != 3 {
		t.Fatalf("Run() ExitCode = %d, want 3", result.ExitCode)
	}
}

func TestExecutorRunMultilineScript(t *testing.T) {
	e := NewExecutor()
	command := "echo line1\necho line2"
	result, err := e.Run(command, RunContext{Capture: boolPtr(true)})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Run() ExitCode = %d, want 0 (stderr=%q)", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "line1") || !strings.Contains(result.Stdout, "line2") {
		t.Fatalf("Run() Stdout = %q, want to contain both line1 and line2", result.Stdout)
	}
}

func TestExecutorRunUsesConfiguredCwd(t *testing.T) {
	dir := t.TempDir()
	e := NewExecutor()

	command := "echo test > out.txt"
	result, err := e.Run(command, RunContext{Cwd: stringPtr(dir), Capture: boolPtr(true)})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Run() ExitCode = %d, want 0 (stderr=%q)", result.ExitCode, result.Stderr)
	}

	outPath := filepath.Join(dir, "out.txt")
	if _, err := os.Stat(outPath); err != nil {
		t.Fatalf("expected file %s to be created in Cwd: %v", outPath, err)
	}
}

func TestExecutorRunAppliesEnvOverrides(t *testing.T) {
	e := NewExecutor()

	var command string
	var shell *string
	if runtime.GOOS == "windows" {
		command = "echo %CMDBOX_RUN_TEST_VAR%"
		shell = stringPtr("cmd")
	} else {
		command = "echo $CMDBOX_RUN_TEST_VAR"
	}

	result, err := e.Run(command, RunContext{
		Env:     envPtr(map[string]string{"CMDBOX_RUN_TEST_VAR": "injected-value"}),
		Capture: boolPtr(true),
		Shell:   shell,
	})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Run() ExitCode = %d, want 0 (stderr=%q)", result.ExitCode, result.Stderr)
	}
	if !strings.Contains(result.Stdout, "injected-value") {
		t.Fatalf("Run() Stdout = %q, want to contain %q", result.Stdout, "injected-value")
	}
}

func TestExecutorRunTimeoutKillsProcess(t *testing.T) {
	e := NewExecutor()

	var command string
	if runtime.GOOS == "windows" {
		command = "ping -n 10 127.0.0.1 >nul"
	} else {
		command = "sleep 10"
	}

	start := time.Now()
	result, err := e.Run(command, RunContext{Capture: boolPtr(true), Timeout: intPtr(1)})
	elapsed := time.Since(start)

	if elapsed >= 8*time.Second {
		t.Fatalf("Run() took %v, want it to be killed well before the 10s sleep completed", elapsed)
	}
	if err != nil {
		t.Fatalf("Run() error = %v, want nil (timeout should surface as a nonzero exit code)", err)
	}
	if result == nil {
		t.Fatal("Run() result = nil, want non-nil")
	}
	if result.ExitCode == 0 {
		t.Fatalf("Run() ExitCode = 0, want nonzero for a killed process")
	}
}

func TestExecutorRunZeroValueContextUsesDefaults(t *testing.T) {
	e := NewExecutor()

	// A completely zero-value RunContext (all pointer fields nil) must not panic
	// and should behave as: no capture, no cwd override, no env override, no
	// timeout, default shell, not emit-only.
	result, err := e.Run("exit 0", RunContext{})
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result == nil {
		t.Fatal("Run() result = nil, want non-nil")
	}
	if result.ExitCode != 0 {
		t.Fatalf("Run() ExitCode = %d, want 0", result.ExitCode)
	}
	if result.Stdout != "" || result.Stderr != "" {
		t.Fatalf("Run() with nil Capture should not capture output; got Stdout=%q Stderr=%q", result.Stdout, result.Stderr)
	}
}

func TestExecutorRunUncapturedWritesToOsStdout(t *testing.T) {
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w

	e := NewExecutor()
	result, runErr := e.Run("echo uncaptured-output", RunContext{Capture: boolPtr(false)})

	_ = w.Close()
	os.Stdout = origStdout

	if runErr != nil {
		t.Fatalf("Run() error = %v", runErr)
	}
	if result.Stdout != "" {
		t.Fatalf("Run() with Capture=false Stdout = %q, want empty (output should go to os.Stdout)", result.Stdout)
	}

	buf := make([]byte, 4096)
	n, _ := r.Read(buf)
	_ = r.Close()

	if !strings.Contains(string(buf[:n]), "uncaptured-output") {
		t.Fatalf("os.Stdout pipe = %q, want to contain %q", string(buf[:n]), "uncaptured-output")
	}
}
