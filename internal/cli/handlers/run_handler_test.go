package handlers

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/prompts"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/resolve"
	"github.com/PhantomLambSoft/CmdBox/internal/runtime"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

// fakeRunService is a hand-rolled stub implementing services.RunService that records the
// arguments it was called with and returns canned results, so handler tests can assert on
// argument mapping without touching a real database, resolver, or subprocess executor.
type fakeRunService struct {
	missingAlias   string
	missingVars    map[string]string
	missingProfile *string
	missingResult  []string
	missingErr     error

	runAlias     string
	runOverrides services.RuntimeOverrides
	runVars      map[string]string
	runProfile   *string
	runResult    runtime.ExecutionResult
	runErr       error

	previewAlias     string
	previewOverrides services.RuntimeOverrides
	previewVars      map[string]string
	previewProfile   *string
	previewResult    resolve.Result
	previewCtx       runtime.RunContext
	previewErr       error
}

func (f *fakeRunService) CollectMissingVars(cmdAlias string, runtimeVars map[string]string, profileName *string) ([]string, error) {
	f.missingAlias = cmdAlias
	f.missingVars = runtimeVars
	f.missingProfile = profileName
	return f.missingResult, f.missingErr
}

func (f *fakeRunService) Run(cmdAlias string, runtimeOverrides services.RuntimeOverrides, runtimeVars map[string]string, profileName *string) (runtime.ExecutionResult, error) {
	f.runAlias = cmdAlias
	f.runOverrides = runtimeOverrides
	f.runVars = runtimeVars
	f.runProfile = profileName
	return f.runResult, f.runErr
}

func (f *fakeRunService) Preview(cmdAlias string, runtimeOverrides services.RuntimeOverrides, runtimeVars map[string]string, profileName *string) (resolve.Result, runtime.RunContext, error) {
	f.previewAlias = cmdAlias
	f.previewOverrides = runtimeOverrides
	f.previewVars = runtimeVars
	f.previewProfile = profileName
	return f.previewResult, f.previewCtx, f.previewErr
}

// freshTestPrompter builds a real *prompts.Prompter bound to whatever os.Stdin is current at
// call time. container.GetPrompter is memoized process-wide via sync.OnceValues, so if tests let
// the real one get built it freezes onto the first test's redirected stdin and every later test's
// withStdin swap is silently ignored. Overriding the var with this (non-memoized) constructor
// keeps each test's prompt reads bound to its own stdin pipe.
func freshTestPrompter() (*prompts.Prompter, error) {
	return prompts.NewPrompter(ui.NewPalette(settings.UIColors{}), false), nil
}

// withFakeRunContainer swaps container.GetRunService, container.GetConsole, and
// container.GetPrompter for the duration of the test, restoring the originals on cleanup.
func withFakeRunContainer(t *testing.T, svc services.RunService, svcErr error) {
	t.Helper()

	origSvc := container.GetRunService
	origConsole := container.GetConsole
	origPrompter := container.GetPrompter
	t.Cleanup(func() {
		container.GetRunService = origSvc
		container.GetConsole = origConsole
		container.GetPrompter = origPrompter
	})

	container.GetRunService = func() (services.RunService, error) {
		return svc, svcErr
	}
	container.GetConsole = func() (*ui.Console, error) {
		return ui.NewConsole(ui.NewPalette(settings.UIColors{}), false, "never", 0, 0, 0), nil
	}
	container.GetPrompter = freshTestPrompter
}

// withFailingRunConsole makes container.GetConsole return an error, to exercise handler error
// paths that occur before any console output happens.
func withFailingRunConsole(t *testing.T, svc services.RunService) {
	t.Helper()
	origSvc := container.GetRunService
	origConsole := container.GetConsole
	origPrompter := container.GetPrompter
	t.Cleanup(func() {
		container.GetRunService = origSvc
		container.GetConsole = origConsole
		container.GetPrompter = origPrompter
	})
	container.GetRunService = func() (services.RunService, error) {
		return svc, nil
	}
	container.GetConsole = func() (*ui.Console, error) {
		return nil, errors.New("console boom")
	}
	container.GetPrompter = freshTestPrompter
}

// withStdin temporarily replaces os.Stdin with a reader over the given content, restoring the
// original on cleanup. Used to drive the non-interactive fallback path in
// prompts.PromptForMissingVars (term.IsTerminal is false under `go test`, so it reads a line from
// os.Stdin rather than opening an interactive TUI prompt).
func withStdin(t *testing.T, content string) {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)

	orig := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = orig
	})

	go func() {
		_, _ = io.WriteString(w, content)
		_ = w.Close()
	}()
}

func boolPtr(b bool) *bool { return &b }

// --- RunRunCommand ---

func TestRunRunCommand_NoMissingVars(t *testing.T) {
	fake := &fakeRunService{runResult: runtime.ExecutionResult{ExitCode: 0}}
	withFakeRunContainer(t, fake, nil)

	err := RunRunCommand("build", map[string]string{"foo": "bar"}, RawRunContext{}, nil)
	require.NoError(t, err)

	assert.Equal(t, "build", fake.missingAlias)
	assert.Equal(t, "build", fake.runAlias)
	assert.Equal(t, "bar", fake.runVars["foo"])
}

func TestRunRunCommand_PromptsForMissingVars(t *testing.T) {
	fake := &fakeRunService{missingResult: []string{"name"}, runResult: runtime.ExecutionResult{ExitCode: 0}}
	withFakeRunContainer(t, fake, nil)
	withStdin(t, "alice\n")

	runtimeVars := map[string]string{}
	err := RunRunCommand("greet", runtimeVars, RawRunContext{}, nil)
	require.NoError(t, err)

	assert.Equal(t, "alice", runtimeVars["name"], "the prompted value should be merged into the caller's runtimeVars map")
	assert.Equal(t, "alice", fake.runVars["name"], "the service should receive the filled-in value")
}

func TestRunRunCommand_PromptError(t *testing.T) {
	fake := &fakeRunService{missingResult: []string{"name"}}
	withFakeRunContainer(t, fake, nil)
	// Close stdin immediately so the plain-prompt reader hits EOF instead of blocking.
	withStdin(t, "")

	err := RunRunCommand("greet", map[string]string{}, RawRunContext{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "prompting for name")
}

func TestRunRunCommand_CollectMissingVarsError(t *testing.T) {
	fake := &fakeRunService{missingErr: errors.New("no such command")}
	withFakeRunContainer(t, fake, nil)

	err := RunRunCommand("missing", map[string]string{}, RawRunContext{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collecting missing vars")
	assert.Contains(t, err.Error(), "no such command")
}

func TestRunRunCommand_InvalidEnvOverride(t *testing.T) {
	fake := &fakeRunService{}
	withFakeRunContainer(t, fake, nil)

	err := RunRunCommand("build", map[string]string{}, RawRunContext{Env: []string{"NOEQUALS"}}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "building runtime overrides from context")
	assert.Contains(t, err.Error(), "NOEQUALS")
}

func TestRunRunCommand_ServiceRunError(t *testing.T) {
	fake := &fakeRunService{runErr: errors.New("exec failed")}
	withFakeRunContainer(t, fake, nil)

	err := RunRunCommand("build", map[string]string{}, RawRunContext{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "running command")
	assert.Contains(t, err.Error(), "exec failed")
}

func TestRunRunCommand_OverridesPassedThrough(t *testing.T) {
	fake := &fakeRunService{}
	withFakeRunContainer(t, fake, nil)

	profile := "work"
	err := RunRunCommand("build", map[string]string{}, RawRunContext{
		Cwd: "/tmp", HasCwd: true,
		Shell: "bash", HasShell: true,
		Timeout: 30, HasTimeout: true,
		Env:     []string{"KEY=value"},
		Emit:    true,
		Verbose: boolPtr(true),
	}, &profile)
	require.NoError(t, err)

	assert.Equal(t, "/tmp", fake.runOverrides.Cwd)
	require.NotNil(t, fake.runOverrides.Shell)
	assert.Equal(t, "bash", *fake.runOverrides.Shell)
	require.NotNil(t, fake.runOverrides.Timeout)
	assert.Equal(t, 30, *fake.runOverrides.Timeout)
	assert.Equal(t, "value", fake.runOverrides.Env["KEY"])
	assert.True(t, fake.runOverrides.Emit)
	require.NotNil(t, fake.runOverrides.Verbose)
	assert.True(t, *fake.runOverrides.Verbose)
	require.NotNil(t, fake.runProfile)
	assert.Equal(t, "work", *fake.runProfile)
}

func TestRunRunCommand_OverridesOmittedWhenNotSet(t *testing.T) {
	fake := &fakeRunService{}
	withFakeRunContainer(t, fake, nil)

	err := RunRunCommand("build", map[string]string{}, RawRunContext{}, nil)
	require.NoError(t, err)

	assert.Equal(t, "", fake.runOverrides.Cwd)
	assert.Nil(t, fake.runOverrides.Shell)
	assert.Nil(t, fake.runOverrides.Timeout)
	assert.Nil(t, fake.runOverrides.Env)
	assert.False(t, fake.runOverrides.Emit)
}

func TestRunRunCommand_VerboseAndNotEmit_PrintsOutput(t *testing.T) {
	fake := &fakeRunService{runResult: runtime.ExecutionResult{Command: "echo hi", ExitCode: 0, Stdout: "hi\n"}}
	withFakeRunContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunRunCommand("build", map[string]string{}, RawRunContext{Verbose: boolPtr(true)}, nil)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "echo hi")
	assert.Contains(t, out, "hi")
}

func TestRunRunCommand_VerboseFalse_NoOutput(t *testing.T) {
	fake := &fakeRunService{runResult: runtime.ExecutionResult{Command: "echo hi", ExitCode: 0, Stdout: "hi\n"}}
	withFakeRunContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunRunCommand("build", map[string]string{}, RawRunContext{Verbose: boolPtr(false)}, nil)
		require.NoError(t, err)
	})

	assert.Empty(t, out)
}

func TestRunRunCommand_VerboseNil_NoOutput(t *testing.T) {
	fake := &fakeRunService{runResult: runtime.ExecutionResult{Command: "echo hi", ExitCode: 0}}
	withFakeRunContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunRunCommand("build", map[string]string{}, RawRunContext{}, nil)
		require.NoError(t, err)
	})

	assert.Empty(t, out)
}

func TestRunRunCommand_VerboseAndEmit_NoOutput(t *testing.T) {
	fake := &fakeRunService{runResult: runtime.ExecutionResult{Command: "echo hi", ExitCode: 0, Stdout: "hi\n"}}
	withFakeRunContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunRunCommand("build", map[string]string{}, RawRunContext{Verbose: boolPtr(true), Emit: true}, nil)
		require.NoError(t, err)
	})

	assert.Empty(t, out, "emit mode should never print the execution result panel")
}

func TestRunRunCommand_PrompterUnavailable(t *testing.T) {
	fake := &fakeRunService{missingResult: []string{"name"}}
	withFakeRunContainer(t, fake, nil)
	container.GetPrompter = func() (*prompts.Prompter, error) {
		return nil, errors.New("terminal unavailable")
	}

	err := RunRunCommand("greet", map[string]string{}, RawRunContext{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting prompter")
	assert.Contains(t, err.Error(), "terminal unavailable")
}

func TestRunRunCommand_RunServiceUnavailable(t *testing.T) {
	withFakeRunContainer(t, nil, errors.New("db down"))

	err := RunRunCommand("build", map[string]string{}, RawRunContext{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting run service")
	assert.Contains(t, err.Error(), "db down")
}

func TestRunRunCommand_ConsoleUnavailable(t *testing.T) {
	fake := &fakeRunService{runResult: runtime.ExecutionResult{ExitCode: 0}}
	withFailingRunConsole(t, fake)

	err := RunRunCommand("build", map[string]string{}, RawRunContext{Verbose: boolPtr(true)}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

func TestRunRunCommand_ConsoleUnavailable_NotNeededWhenNotVerbose(t *testing.T) {
	fake := &fakeRunService{runResult: runtime.ExecutionResult{ExitCode: 0}}
	withFailingRunConsole(t, fake)

	err := RunRunCommand("build", map[string]string{}, RawRunContext{}, nil)
	require.NoError(t, err, "console is only needed for verbose output, so a failing console should not matter here")
}

// --- RunPreviewCommand ---

func TestRunPreviewCommand_NoMissingVars(t *testing.T) {
	fake := &fakeRunService{previewResult: resolve.Result{Text: "echo hi"}}
	withFakeRunContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunPreviewCommand("build", map[string]string{"foo": "bar"}, RawRunContext{}, nil)
		require.NoError(t, err)
	})

	assert.Equal(t, "build", fake.missingAlias)
	assert.Equal(t, "build", fake.previewAlias)
	assert.Equal(t, "bar", fake.previewVars["foo"])
	assert.Contains(t, out, "echo hi")
}

func TestRunPreviewCommand_PromptsForMissingVars(t *testing.T) {
	fake := &fakeRunService{missingResult: []string{"name"}, previewResult: resolve.Result{Text: "echo alice"}}
	withFakeRunContainer(t, fake, nil)
	withStdin(t, "alice\n")

	runtimeVars := map[string]string{}
	captureStdout(t, func() {
		err := RunPreviewCommand("greet", runtimeVars, RawRunContext{}, nil)
		require.NoError(t, err)
	})

	assert.Equal(t, "alice", runtimeVars["name"])
	assert.Equal(t, "alice", fake.previewVars["name"])
}

func TestRunPreviewCommand_PromptError(t *testing.T) {
	fake := &fakeRunService{missingResult: []string{"name"}}
	withFakeRunContainer(t, fake, nil)
	withStdin(t, "")

	err := RunPreviewCommand("greet", map[string]string{}, RawRunContext{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "prompting for name")
}

func TestRunPreviewCommand_CollectMissingVarsError(t *testing.T) {
	fake := &fakeRunService{missingErr: errors.New("no such command")}
	withFakeRunContainer(t, fake, nil)

	err := RunPreviewCommand("missing", map[string]string{}, RawRunContext{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "collecting missing vars")
}

func TestRunPreviewCommand_InvalidEnvOverride(t *testing.T) {
	fake := &fakeRunService{}
	withFakeRunContainer(t, fake, nil)

	err := RunPreviewCommand("build", map[string]string{}, RawRunContext{Env: []string{"NOEQUALS"}}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "building runtime overrides from context")
}

func TestRunPreviewCommand_ServicePreviewError(t *testing.T) {
	fake := &fakeRunService{previewErr: errors.New("resolve failed")}
	withFakeRunContainer(t, fake, nil)

	err := RunPreviewCommand("build", map[string]string{}, RawRunContext{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "previewing run")
	assert.Contains(t, err.Error(), "resolve failed")
}

func TestRunPreviewCommand_OverridesPassedThrough(t *testing.T) {
	fake := &fakeRunService{}
	withFakeRunContainer(t, fake, nil)

	profile := "work"
	captureStdout(t, func() {
		err := RunPreviewCommand("build", map[string]string{}, RawRunContext{
			Cwd: "/tmp", HasCwd: true,
			Shell: "bash", HasShell: true,
			Timeout: 30, HasTimeout: true,
			Env: []string{"KEY=value"},
		}, &profile)
		require.NoError(t, err)
	})

	assert.Equal(t, "/tmp", fake.previewOverrides.Cwd)
	require.NotNil(t, fake.previewOverrides.Shell)
	assert.Equal(t, "bash", *fake.previewOverrides.Shell)
	require.NotNil(t, fake.previewOverrides.Timeout)
	assert.Equal(t, 30, *fake.previewOverrides.Timeout)
	assert.Equal(t, "value", fake.previewOverrides.Env["KEY"])
	require.NotNil(t, fake.previewProfile)
	assert.Equal(t, "work", *fake.previewProfile)
}

func TestRunPreviewCommand_AlwaysPrintsResult(t *testing.T) {
	fake := &fakeRunService{previewResult: resolve.Result{Text: "echo hi"}}
	withFakeRunContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunPreviewCommand("build", map[string]string{}, RawRunContext{Verbose: boolPtr(false)}, nil)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "echo hi", "unlike RunRunCommand, preview output does not depend on the Verbose flag")
}

func TestRunPreviewCommand_PrompterUnavailable(t *testing.T) {
	fake := &fakeRunService{missingResult: []string{"name"}}
	withFakeRunContainer(t, fake, nil)
	container.GetPrompter = func() (*prompts.Prompter, error) {
		return nil, errors.New("terminal unavailable")
	}

	err := RunPreviewCommand("greet", map[string]string{}, RawRunContext{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting prompter")
	assert.Contains(t, err.Error(), "terminal unavailable")
}

func TestRunPreviewCommand_RunServiceUnavailable(t *testing.T) {
	withFakeRunContainer(t, nil, errors.New("db down"))

	err := RunPreviewCommand("build", map[string]string{}, RawRunContext{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting run service")
}

func TestRunPreviewCommand_ConsoleUnavailable(t *testing.T) {
	fake := &fakeRunService{previewResult: resolve.Result{Text: "echo hi"}}
	withFailingRunConsole(t, fake)

	err := RunPreviewCommand("build", map[string]string{}, RawRunContext{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

// --- toRuntimeOverrides ---

func TestToRuntimeOverrides_Defaults(t *testing.T) {
	overrides, err := toRuntimeOverrides(RawRunContext{})
	require.NoError(t, err)

	assert.Equal(t, "", overrides.Cwd)
	assert.Nil(t, overrides.Shell)
	assert.Nil(t, overrides.Timeout)
	assert.Nil(t, overrides.Env)
	assert.False(t, overrides.Emit)
}

func TestToRuntimeOverrides_AllSet(t *testing.T) {
	overrides, err := toRuntimeOverrides(RawRunContext{
		Cwd: "/tmp", HasCwd: true,
		Shell: "zsh", HasShell: true,
		Timeout: 10, HasTimeout: true,
		Env:     []string{"A=1", "B=2"},
		Emit:    true,
		Verbose: boolPtr(true),
	})
	require.NoError(t, err)

	assert.Equal(t, "/tmp", overrides.Cwd)
	require.NotNil(t, overrides.Shell)
	assert.Equal(t, "zsh", *overrides.Shell)
	require.NotNil(t, overrides.Timeout)
	assert.Equal(t, 10, *overrides.Timeout)
	assert.Equal(t, map[string]string{"A": "1", "B": "2"}, overrides.Env)
	assert.True(t, overrides.Emit)
	require.NotNil(t, overrides.Verbose)
	assert.True(t, *overrides.Verbose)
}

func TestToRuntimeOverrides_VerboseFalseIsForwarded(t *testing.T) {
	overrides, err := toRuntimeOverrides(RawRunContext{Verbose: boolPtr(false)})
	require.NoError(t, err)

	require.NotNil(t, overrides.Verbose, "an explicit false should be forwarded, not dropped like a nil")
	assert.False(t, *overrides.Verbose)
}

func TestToRuntimeOverrides_InvalidEnv(t *testing.T) {
	_, err := toRuntimeOverrides(RawRunContext{Env: []string{"bad-pair"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad-pair")
}
