package commands

import (
	"testing"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeHandlerSettingsProvider and fakeHandlerConsole implement
// handlers.SettingsProvider / handlers.ConsoleWriter without touching the
// real app data directory or settings file, and without invoking the
// editFn they're given (which, in production, launches a real fullscreen
// TUI editor or a real external editor process).
type fakeHandlerSettingsProvider struct {
	value   settings.Settings
	editErr error
}

func (f *fakeHandlerSettingsProvider) Get() settings.Settings { return f.value }

func (f *fakeHandlerSettingsProvider) Edit(func(current string) (string, error)) (settings.Settings, error) {
	if f.editErr != nil {
		return settings.Settings{}, f.editErr
	}
	return f.value, nil
}

type fakeHandlerConsole struct{}

func (fakeHandlerConsole) Palette() ui.Palette { return ui.Palette{} }
func (fakeHandlerConsole) Print(string)        {}
func (fakeHandlerConsole) Error(string)        {}
func (fakeHandlerConsole) Muted(string)        {}
func (fakeHandlerConsole) Success(string)      {}

// withFakeSettingsHandlerDeps substitutes the handlers package's settings
// service, console, app-data-dir, and "open the data directory" side
// effect for the duration of the test, restoring them on cleanup. Without
// this, RunShowSettings/RunEditSettings/RunSettingsDataDir would read the
// real settings file, resolve the real app data directory, and (for
// RunSettingsDataDir with print=false) open a real OS file explorer
// window.
func withFakeSettingsHandlerDeps(t *testing.T) {
	t.Helper()
	restoreSvc := handlers.SetSettingsServiceForTest(func() (handlers.SettingsProvider, error) {
		return &fakeHandlerSettingsProvider{}, nil
	})
	restoreConsole := handlers.SetConsoleForTest(func() (handlers.ConsoleWriter, error) {
		return fakeHandlerConsole{}, nil
	})
	restoreDataDir := handlers.SetAppDataDirForTest(func() (string, error) {
		return "/fake/data/dir", nil
	})
	restoreOpen := handlers.SetOpenDataDirForTest(func(dir string) error { return nil })
	t.Cleanup(func() {
		restoreSvc()
		restoreConsole()
		restoreDataDir()
		restoreOpen()
	})
}

func TestSettingsCmd_Metadata(t *testing.T) {
	assert.Equal(t, "settings", settingsCmd.Name())
	assert.Contains(t, rootCmd.Commands(), settingsCmd)
}

func TestSettingsCmd_Subcommands_Wired(t *testing.T) {
	want := []*cobra.Command{settingsShowCmd, settingsEditCmd, settingsDataDirCmd}
	for _, c := range want {
		assert.Contains(t, settingsCmd.Commands(), c, "expected %s to be a child of settingsCmd", c.Name())
	}
}

func TestSettingsShowCmd_Args(t *testing.T) {
	assert.NoError(t, settingsShowCmd.ValidateArgs([]string{}))
	assert.Error(t, settingsShowCmd.ValidateArgs([]string{"extra"}))
}

func TestSettingsEditCmd_Args(t *testing.T) {
	assert.NoError(t, settingsEditCmd.ValidateArgs([]string{}))
	assert.Error(t, settingsEditCmd.ValidateArgs([]string{"extra"}))
}

func TestSettingsDataDirCmd_Args(t *testing.T) {
	assert.NoError(t, settingsDataDirCmd.ValidateArgs([]string{}))
	assert.Error(t, settingsDataDirCmd.ValidateArgs([]string{"extra"}))
}

func TestSettingsShowCmd_Flags(t *testing.T) {
	assertFlag(t, settingsShowCmd.Flags(), "field", "", "stringSlice", "[]")
}

func TestRunSettingsShow_FieldFilterIsRead(t *testing.T) {
	cmd := freshCmdWithFlagsOf(settingsShowCmd)
	require.NoError(t, cmd.Flags().Set("field", "cwd,shell"))
	assert.Equal(t, []string{"cwd", "shell"}, getFields(cmd))
}

func TestSettingsEditCmd_Flags(t *testing.T) {
	assertFlag(t, settingsEditCmd.Flags(), "external", "e", "bool", "false")
}

func TestSettingsDataDirCmd_Flags(t *testing.T) {
	assertFlag(t, settingsDataDirCmd.Flags(), "print", "p", "bool", "false")
}

func TestRunSettingsShow(t *testing.T) {
	withFakeSettingsHandlerDeps(t)
	cmd := freshCmdWithFlagsOf(settingsShowCmd)
	assert.NoError(t, runSettingsShow(cmd, []string{}))
}

func TestRunSettingsEdit(t *testing.T) {
	withFakeSettingsHandlerDeps(t)
	cmd := freshCmdWithFlagsOf(settingsEditCmd)
	assert.NoError(t, runSettingsEdit(cmd, []string{}))

	require.NoError(t, cmd.Flags().Set("external", "true"))
	assert.NoError(t, runSettingsEdit(cmd, []string{}))
}

func TestRunSettingsDataDir(t *testing.T) {
	withFakeSettingsHandlerDeps(t)
	cmd := freshCmdWithFlagsOf(settingsDataDirCmd)
	assert.NoError(t, runSettingsDataDir(cmd, []string{}))

	require.NoError(t, cmd.Flags().Set("print", "true"))
	assert.NoError(t, runSettingsDataDir(cmd, []string{}))
}
