package runtime

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Executor struct{}

func NewExecutor() *Executor {
	return &Executor{}
}

func (e *Executor) Run(command string, ctx RunContext) (*ExecutionResult, error) {
	if ctx.Emit != nil && *ctx.Emit {
		emitCommand(command)
		return nil, nil
	}

	var env []string
	if ctx.Env != nil {
		env = mergeEnv(*ctx.Env)
	}

	if isMultiline(command) {
		return e.runMultilineAsScript(command, ctx, env)
	}

	var shell string
	if ctx.Shell != nil {
		shell = *ctx.Shell
	}
	args, err := BuildShellCommand(command, shell)
	if err != nil {
		return nil, fmt.Errorf("building shell command: %w", err)
	}

	var cwd = ""
	var capture = false
	var timeout = 0
	if ctx.Cwd != nil {
		cwd = *ctx.Cwd
	}
	if ctx.Capture != nil {
		capture = *ctx.Capture
	}
	if ctx.Timeout != nil {
		timeout = *ctx.Timeout
	}

	return e.executeCommand(command, args, cwd, env, capture, timeout)
}

func (e *Executor) runMultilineAsScript(command string, ctx RunContext, env []string) (*ExecutionResult, error) {
	var shell = ""
	if ctx.Shell != nil {
		shell = strings.ToLower(*ctx.Shell)
	}
	if shell == "" {
		shell = "default"
	}
	suffix := scriptSuffixForShell(shell)
	scriptBody := strings.TrimRight(strings.ReplaceAll(command, "\r\n", "\n"), "\n") + "\n"

	f, err := os.CreateTemp("", "cmdbox-*"+suffix)
	if err != nil {
		return nil, fmt.Errorf("creating script file: %w", err)
	}

	scriptPath := f.Name()
	defer func() {
		_ = os.Remove(scriptPath)
	}()

	header := scriptHeaderForShell(shell)
	if _, err := f.WriteString(header + scriptBody); err != nil {
		f.Close()
		return nil, fmt.Errorf("writing script file: %w", err)
	}
	if err := f.Close(); err != nil {
		return nil, fmt.Errorf("closing script file: %w", err)
	}

	args := buildScriptExecArgs(scriptPath, shell)

	var cwd = ""
	var capture = false
	var timeout = 0
	if ctx.Cwd != nil {
		cwd = *ctx.Cwd
	}
	if ctx.Capture != nil {
		capture = *ctx.Capture
	}
	if ctx.Timeout != nil {
		timeout = *ctx.Timeout
	}

	return e.executeCommand(command, args, cwd, env, capture, timeout)
}

func (e *Executor) executeCommand(
	command string,
	args []string,
	cwd string,
	env []string,
	captureOutput bool,
	timeoutSeconds int,
) (*ExecutionResult, error) {
	ctx := context.Background()
	if timeoutSeconds > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeoutSeconds)*time.Second)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = cwd
	cmd.Env = env
	configureProcessGroup(cmd) // Platform specific
	cmd.Cancel = func() error {
		killProcessTree(cmd)
		return nil
	}
	cmd.WaitDelay = 5 * time.Second // Give the tree 5 seconds to die before wait gives up

	var stdout, stderr bytes.Buffer
	if captureOutput {
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	err := cmd.Run()

	if ctx.Err() == context.DeadlineExceeded {
		// TODO: log that command timed out
	}

	exitCode, ok := exitCodeFrom(cmd, err)
	if !ok {
		return nil, fmt.Errorf("running command %w", err)
	}
	// TODO: log command completed
	return &ExecutionResult{
		Command:  command,
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}, nil
}

// exitCodeFrom extracts the exit code from an exec.Cmd or an error, returning an indicator of success or failure.
// This distinguishes between "ran and exited non-zero" from "the process never produced an exit code at all."
func exitCodeFrom(cmd *exec.Cmd, err error) (code int, ok bool) {
	if err == nil {
		return cmd.ProcessState.ExitCode(), true
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode(), true
	}
	return -1, false
}

func mergeEnv(overrides map[string]string) []string {
	base := os.Environ()
	if len(overrides) == 0 {
		return base
	}
	merged := make(map[string]string, len(base)+len(overrides))
	for _, kv := range base {
		if k, v, ok := strings.Cut(kv, "="); ok {
			merged[k] = v
		}
	}
	for k, v := range overrides {
		merged[k] = v
	}
	result := make([]string, 0, len(merged))
	for k, v := range merged {
		result = append(result, k+"="+v)
	}
	return result
}

func isMultiline(command string) bool {
	return strings.Contains(strings.Trim(command, "\n"), "\n")
}

func modelLabel(emit bool) string {
	if emit {
		return "emit"
	}
	return "subprocess"
}

func scriptSuffixForShell(shell string) string {
	switch {
	case shell == "default":
		if runtime.GOOS == "windows" {
			return ".cmd"
		}
		return ".sh"
	case strings.Contains(shell, "cmd") || shell == "cmd.exe":
		return ".cmd"
	case strings.Contains(shell, "powershell") || shell == "pwsh":
		return ".ps1"
	case strings.Contains(shell, "fish"):
		return ".fish"
	default:
		return ".sh"
	}
}

func scriptHeaderForShell(shell string) string {
	switch {
	case strings.Contains(shell, "bash"):
		return "#!/usr/bin/env bash\n"
	case strings.Contains(shell, "zsh"):
		return "#!/usr/bin/env zsh\n"
	case strings.Contains(shell, "fish"):
		return "#!/usr/bin/env fish\n"
	default:
		return ""
	}
}

// buildScriptExecArgs constructs and returns the appropriate command-line arguments to execute a script based on the given shell.
func buildScriptExecArgs(scriptPath, shell string) []string {
	switch {
	case shell == "default":
		if runtime.GOOS == "windows" {
			return []string{"cmd.exe", "/d", "/s", "/c", scriptPath}
		}
		return []string{"sh", scriptPath}
	case strings.Contains(shell, "cmd") || shell == "cmd.exe":
		return []string{"cmd.exe", "/d", "/s", "/c", scriptPath}
	case strings.Contains(shell, "powershell") || shell == "pwsh":
		exe := "powershell"
		if strings.Contains(shell, "pwsh") {
			exe = "pwsh"
		}
		return []string{exe, "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Bypass", "-File", scriptPath}
	case strings.Contains(shell, "fish"):
		return []string{"fish", scriptPath}
	case strings.Contains(shell, "zsh"):
		return []string{"zsh", scriptPath}
	case strings.Contains(shell, "bash"):
		return []string{"bash", scriptPath}
	default:
		return []string{"sh", scriptPath}
	}
}

func emitCommand(command string) {
	fmt.Fprintln(os.Stdout, strings.Trim(command, "\n"))
	os.Exit(0)
}
