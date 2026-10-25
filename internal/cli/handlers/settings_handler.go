package handlers

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/presenters"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

// SettingsProvider and ConsoleWriter are provided here for testing purposes
type SettingsProvider interface {
	Get() settings.Settings
	Edit(editFn func(current string) (string, error)) (settings.Settings, error)
}

type ConsoleWriter interface {
	Palette() ui.Palette
	Print(string)
	Error(string)
	Muted(string)
	Success(string)
}

// Indirections over container access and the "open the data directory"
// side effect, so tests can override them via the SetXForTest functions.
// Production code should not reassign these outside of this var block.
var (
	getSettingsService = func() (SettingsProvider, error) { return container.GetSettingsService() }
	getConsole         = func() (ConsoleWriter, error) { return container.GetConsole() }
	getAppDataDir      = container.GetAppDataDir
	openDataDir        = defaultOpenDataDir
)

func SetSettingsServiceForTest(fn func() (SettingsProvider, error)) func() {
	prev := getSettingsService
	getSettingsService = fn
	return func() { getSettingsService = prev }
}

func SetConsoleForTest(fn func() (ConsoleWriter, error)) func() {
	prev := getConsole
	getConsole = fn
	return func() { getConsole = prev }
}

func SetAppDataDirForTest(fn func() (string, error)) func() {
	prev := getAppDataDir
	getAppDataDir = fn
	return func() { getAppDataDir = prev }
}

func SetOpenDataDirForTest(fn func(dir string) error) func() {
	prev := openDataDir
	openDataDir = fn
	return func() { openDataDir = prev }
}

func RunShowSettings(fields []string) error {
	settingsService, err := getSettingsService()
	if err != nil {
		return fmt.Errorf("getting settings service: %w", err)
	}
	settingsValue := settingsService.Get()
	console, err := getConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	renderedOutput, err := presenters.RenderSettingsShow(settingsValue, fields, console.Palette())
	if err != nil {
		if errors.Is(err, presenters.ErrUnknownSettingsField) {
			console.Error(err.Error())
			return nil
		}
		return fmt.Errorf("rendering settings show: %w", err)
	}
	console.Print(renderedOutput)

	return nil
}

func RunEditSettings(external bool) error {
	settingsService, err := getSettingsService()
	if err != nil {
		return fmt.Errorf("getting settings service: %w", err)
	}
	console, err := getConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	editFn := func(current string) (string, error) {
		return ui.EditTextFullscreen(current, "Edit settings", console.Palette())
	}
	if external {
		editFn = func(current string) (string, error) {
			return ui.EditTextExternal(current, ".toml")
		}
	}

	_, err = settingsService.Edit(editFn)
	if err != nil {
		if errors.Is(err, ui.ErrEditCanceled) {
			console.Muted("Edit canceled, no changes made.")
			return nil
		}
		return fmt.Errorf("editing settings: %w", err)
	}
	console.Success("Settings updated.")
	return nil
}

func RunSettingsDataDir(print bool) error {
	dataDir, err := getAppDataDir()
	if err != nil {
		return fmt.Errorf("getting app data dir: %w", err)
	}
	if print {
		console, err := getConsole()
		if err != nil {
			return fmt.Errorf("getting console: %w", err)
		}
		console.Print(presenters.RenderDataDir(dataDir, console.Palette()))
		return nil
	}

	if err := openDataDir(dataDir); err != nil {
		return fmt.Errorf("opening data directory: %w", err)
	}
	return nil
}

func defaultOpenDataDir(dataDir string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", dataDir)
	case "darwin":
		cmd = exec.Command("open", dataDir)
	default:
		cmd = exec.Command("xdg-open", dataDir)
	}

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		// Windows always exits with error code 1, even when it opens the folder without error. This is a known issue.
		// We check for the exit code and platform here, and if it is code 1 on Windows, the error is ignored.
		windowsFalsePositive := runtime.GOOS == "windows" && errors.As(err, &exitErr) && exitErr.ExitCode() == 1
		if !windowsFalsePositive {
			return err
		}
	}
	return nil
}
