package handlers

import (
	"errors"
	"strings"
	"testing"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

// fakeSettingsProvider implements SettingsProvider without ever touching a
// real settings file. Edit intentionally does NOT invoke the editFn it is
// given: in production that function launches a real fullscreen TUI editor
// or a real external editor process, and the handler tests only care about
// how RunEditSettings reacts to Edit's return value, not what Edit itself
// does with the callback.
type fakeSettingsProvider struct {
	settings settings.Settings

	editResult settings.Settings
	editErr    error
	editCalls  int
	lastEditFn func(current string) (string, error)
}

func (f *fakeSettingsProvider) Get() settings.Settings { return f.settings }

func (f *fakeSettingsProvider) Edit(editFn func(current string) (string, error)) (settings.Settings, error) {
	f.editCalls++
	f.lastEditFn = editFn
	if f.editErr != nil {
		return settings.Settings{}, f.editErr
	}
	return f.editResult, nil
}

// fakeConsole implements ConsoleWriter by recording calls instead of
// writing to a real terminal.
type fakeConsole struct {
	palette ui.Palette
	printed []string
	errored []string
	muted   []string
	success []string
}

func (c *fakeConsole) Palette() ui.Palette { return c.palette }
func (c *fakeConsole) Print(s string)      { c.printed = append(c.printed, s) }
func (c *fakeConsole) Error(s string)      { c.errored = append(c.errored, s) }
func (c *fakeConsole) Muted(s string)      { c.muted = append(c.muted, s) }
func (c *fakeConsole) Success(s string)    { c.success = append(c.success, s) }

func withFakeSettingsService(t *testing.T, p *fakeSettingsProvider) {
	t.Helper()
	restore := SetSettingsServiceForTest(func() (SettingsProvider, error) { return p, nil })
	t.Cleanup(restore)
}

func withFailingSettingsService(t *testing.T, err error) {
	t.Helper()
	restore := SetSettingsServiceForTest(func() (SettingsProvider, error) { return nil, err })
	t.Cleanup(restore)
}

func withFakeSettingsConsole(t *testing.T, c *fakeConsole) {
	t.Helper()
	restore := SetConsoleForTest(func() (ConsoleWriter, error) { return c, nil })
	t.Cleanup(restore)
}

func withFailingSettingsConsole(t *testing.T, err error) {
	t.Helper()
	restore := SetConsoleForTest(func() (ConsoleWriter, error) { return nil, err })
	t.Cleanup(restore)
}

func withFakeAppDataDir(t *testing.T, dir string, err error) {
	t.Helper()
	restore := SetAppDataDirForTest(func() (string, error) { return dir, err })
	t.Cleanup(restore)
}

func withFakeOpenDataDir(t *testing.T, fn func(dir string) error) {
	t.Helper()
	restore := SetOpenDataDirForTest(fn)
	t.Cleanup(restore)
}

// ---- RunShowSettings ----

func TestRunShowSettings_NoFieldsRendersEverything(t *testing.T) {
	withFakeSettingsService(t, &fakeSettingsProvider{settings: settings.Settings{}})
	console := &fakeConsole{}
	withFakeSettingsConsole(t, console)

	if err := RunShowSettings(nil); err != nil {
		t.Fatalf("RunShowSettings() error = %v", err)
	}
	if len(console.printed) != 1 {
		t.Fatalf("console.Print called %d times, want 1", len(console.printed))
	}
	if console.printed[0] == "" {
		t.Error("expected non-empty rendered output")
	}
	if len(console.errored) != 0 {
		t.Errorf("console.Error should not be called, got %v", console.errored)
	}
}

func TestRunShowSettings_KnownField(t *testing.T) {
	withFakeSettingsService(t, &fakeSettingsProvider{settings: settings.Settings{}})
	console := &fakeConsole{}
	withFakeSettingsConsole(t, console)

	if err := RunShowSettings([]string{"history.enabled"}); err != nil {
		t.Fatalf("RunShowSettings() error = %v", err)
	}
	if len(console.printed) != 1 {
		t.Fatalf("console.Print called %d times, want 1", len(console.printed))
	}
}

func TestRunShowSettings_UnknownFieldPrintsErrorAndReturnsNil(t *testing.T) {
	withFakeSettingsService(t, &fakeSettingsProvider{settings: settings.Settings{}})
	console := &fakeConsole{}
	withFakeSettingsConsole(t, console)

	err := RunShowSettings([]string{"does.not.exist"})
	if err != nil {
		t.Fatalf("RunShowSettings() error = %v, want nil (error should be reported via console)", err)
	}
	if len(console.errored) != 1 {
		t.Fatalf("console.Error called %d times, want 1", len(console.errored))
	}
	if !strings.Contains(console.errored[0], "does.not.exist") {
		t.Errorf("console.Error message = %q, want it to mention the bad field", console.errored[0])
	}
	if len(console.printed) != 0 {
		t.Errorf("console.Print should not be called on an unknown field, got %v", console.printed)
	}
}

func TestRunShowSettings_SettingsServiceErrorIsWrapped(t *testing.T) {
	wantErr := errors.New("boom")
	withFailingSettingsService(t, wantErr)

	err := RunShowSettings(nil)
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("RunShowSettings() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestRunShowSettings_ConsoleErrorIsWrapped(t *testing.T) {
	withFakeSettingsService(t, &fakeSettingsProvider{settings: settings.Settings{}})
	wantErr := errors.New("no console")
	withFailingSettingsConsole(t, wantErr)

	err := RunShowSettings(nil)
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("RunShowSettings() error = %v, want wrapped %v", err, wantErr)
	}
}

// ---- RunEditSettings ----

func TestRunEditSettings_SuccessPrintsSuccessMessage(t *testing.T) {
	provider := &fakeSettingsProvider{editResult: settings.Settings{}}
	withFakeSettingsService(t, provider)
	console := &fakeConsole{}
	withFakeSettingsConsole(t, console)

	if err := RunEditSettings(false); err != nil {
		t.Fatalf("RunEditSettings() error = %v", err)
	}
	if provider.editCalls != 1 {
		t.Fatalf("Edit called %d times, want 1", provider.editCalls)
	}
	if provider.lastEditFn == nil {
		t.Error("expected a non-nil editFn to be passed to Edit")
	}
	if len(console.success) != 1 {
		t.Fatalf("console.Success called %d times, want 1", len(console.success))
	}
	if len(console.muted) != 0 {
		t.Errorf("console.Muted should not be called on success, got %v", console.muted)
	}
}

func TestRunEditSettings_ExternalAlsoCallsEdit(t *testing.T) {
	provider := &fakeSettingsProvider{editResult: settings.Settings{}}
	withFakeSettingsService(t, provider)
	withFakeSettingsConsole(t, &fakeConsole{})

	if err := RunEditSettings(true); err != nil {
		t.Fatalf("RunEditSettings() error = %v", err)
	}
	if provider.editCalls != 1 {
		t.Fatalf("Edit called %d times, want 1", provider.editCalls)
	}
	if provider.lastEditFn == nil {
		t.Error("expected a non-nil editFn to be passed to Edit")
	}
}

func TestRunEditSettings_CanceledIsNotAnError(t *testing.T) {
	provider := &fakeSettingsProvider{editErr: ui.ErrEditCanceled}
	withFakeSettingsService(t, provider)
	console := &fakeConsole{}
	withFakeSettingsConsole(t, console)

	err := RunEditSettings(false)
	if err != nil {
		t.Fatalf("RunEditSettings() error = %v, want nil on cancel", err)
	}
	if len(console.muted) != 1 {
		t.Fatalf("console.Muted called %d times, want 1", len(console.muted))
	}
	if len(console.success) != 0 {
		t.Errorf("console.Success should not be called on cancel, got %v", console.success)
	}
}

func TestRunEditSettings_GenericEditErrorIsWrapped(t *testing.T) {
	wantErr := errors.New("invalid TOML")
	provider := &fakeSettingsProvider{editErr: wantErr}
	withFakeSettingsService(t, provider)
	withFakeSettingsConsole(t, &fakeConsole{})

	err := RunEditSettings(false)
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("RunEditSettings() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestRunEditSettings_SettingsServiceErrorIsWrapped(t *testing.T) {
	wantErr := errors.New("boom")
	withFailingSettingsService(t, wantErr)

	err := RunEditSettings(false)
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("RunEditSettings() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestRunEditSettings_ConsoleErrorIsWrapped(t *testing.T) {
	withFakeSettingsService(t, &fakeSettingsProvider{})
	wantErr := errors.New("no console")
	withFailingSettingsConsole(t, wantErr)

	err := RunEditSettings(false)
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("RunEditSettings() error = %v, want wrapped %v", err, wantErr)
	}
}

// ---- RunSettingsDataDir ----

func TestRunSettingsDataDir_PrintRendersPathWithoutOpening(t *testing.T) {
	withFakeAppDataDir(t, "/fake/data/dir", nil)
	console := &fakeConsole{}
	withFakeSettingsConsole(t, console)
	withFakeOpenDataDir(t, func(dir string) error {
		t.Fatalf("openDataDir should not be called when print=true, got dir %q", dir)
		return nil
	})

	if err := RunSettingsDataDir(true); err != nil {
		t.Fatalf("RunSettingsDataDir() error = %v", err)
	}
	if len(console.printed) != 1 {
		t.Fatalf("console.Print called %d times, want 1", len(console.printed))
	}
	if !strings.Contains(console.printed[0], "/fake/data/dir") {
		t.Errorf("console.Print output = %q, want it to contain the data dir", console.printed[0])
	}
}

func TestRunSettingsDataDir_PrintConsoleErrorIsWrapped(t *testing.T) {
	withFakeAppDataDir(t, "/fake/data/dir", nil)
	wantErr := errors.New("no console")
	withFailingSettingsConsole(t, wantErr)

	err := RunSettingsDataDir(true)
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("RunSettingsDataDir() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestRunSettingsDataDir_OpensWithoutConsole(t *testing.T) {
	withFakeAppDataDir(t, "/fake/data/dir", nil)
	restore := SetConsoleForTest(func() (ConsoleWriter, error) {
		t.Fatal("getConsole should not be called when print=false")
		return nil, nil
	})
	t.Cleanup(restore)

	var openedDir string
	withFakeOpenDataDir(t, func(dir string) error {
		openedDir = dir
		return nil
	})

	if err := RunSettingsDataDir(false); err != nil {
		t.Fatalf("RunSettingsDataDir() error = %v", err)
	}
	if openedDir != "/fake/data/dir" {
		t.Errorf("openDataDir called with %q, want %q", openedDir, "/fake/data/dir")
	}
}

func TestRunSettingsDataDir_OpenErrorIsWrapped(t *testing.T) {
	withFakeAppDataDir(t, "/fake/data/dir", nil)
	wantErr := errors.New("xdg-open not found")
	withFakeOpenDataDir(t, func(dir string) error { return wantErr })

	err := RunSettingsDataDir(false)
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("RunSettingsDataDir() error = %v, want wrapped %v", err, wantErr)
	}
}

func TestRunSettingsDataDir_AppDataDirErrorIsWrapped(t *testing.T) {
	wantErr := errors.New("no home dir")
	withFakeAppDataDir(t, "", wantErr)

	err := RunSettingsDataDir(false)
	if err == nil || !errors.Is(err, wantErr) {
		t.Fatalf("RunSettingsDataDir() error = %v, want wrapped %v", err, wantErr)
	}
}
