package handlers

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/prompts"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/runtime"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

// fakeHistoryService is a hand-rolled stub implementing services.HistoryService that records the
// arguments it was called with and returns canned results, so handler tests can assert on
// argument mapping without touching a real database.
type fakeHistoryService struct {
	recentAlias   *string
	recentLimit   *int
	recentProfile *string
	recentResult  []models.CommandHistory
	recentErr     error

	refRef     string
	refAlias   *string
	refProfile *string
	refResult  *models.CommandHistory
	refErr     error

	varsEntry  *models.CommandHistory
	varsResult map[string]string
	varsErr    error

	deleteRef     string
	deleteAlias   *string
	deleteProfile *string
	deleteErr     error

	clearCalled  bool
	clearAlias   *string
	clearProfile *string
	clearCount   int
	clearErr     error
}

func (f *fakeHistoryService) GetRecent(alias *string, limit *int, profileName *string) ([]models.CommandHistory, error) {
	f.recentAlias = alias
	f.recentLimit = limit
	f.recentProfile = profileName
	return f.recentResult, f.recentErr
}

func (f *fakeHistoryService) GetByRef(ref string, alias *string, profileName *string) (*models.CommandHistory, error) {
	f.refRef = ref
	f.refAlias = alias
	f.refProfile = profileName
	return f.refResult, f.refErr
}

func (f *fakeHistoryService) GetVariables(entry *models.CommandHistory) (map[string]string, error) {
	f.varsEntry = entry
	return f.varsResult, f.varsErr
}

func (f *fakeHistoryService) DeleteByRef(ref string, alias *string, profileName *string) error {
	f.deleteRef = ref
	f.deleteAlias = alias
	f.deleteProfile = profileName
	return f.deleteErr
}

func (f *fakeHistoryService) Clear(alias *string, profileName *string) (int, error) {
	f.clearCalled = true
	f.clearAlias = alias
	f.clearProfile = profileName
	return f.clearCount, f.clearErr
}

// withFakeHistoryContainer swaps container.GetHistoryService and container.GetConsole for the
// duration of the test, restoring the originals on cleanup.
func withFakeHistoryContainer(t *testing.T, svc services.HistoryService, svcErr error) {
	t.Helper()

	origSvc := container.GetHistoryService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetHistoryService = origSvc
		container.GetConsole = origConsole
	})

	container.GetHistoryService = func() (services.HistoryService, error) {
		return svc, svcErr
	}
	container.GetConsole = func() (*ui.Console, error) {
		return ui.NewConsole(ui.NewPalette(settings.UIColors{}), false, "never", 0, 0, 0), nil
	}
}

// withFailingHistoryConsole makes container.GetConsole return an error, to exercise handler error
// paths that occur before any console output happens.
func withFailingHistoryConsole(t *testing.T, svc services.HistoryService) {
	t.Helper()
	origSvc := container.GetHistoryService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetHistoryService = origSvc
		container.GetConsole = origConsole
	})
	container.GetHistoryService = func() (services.HistoryService, error) {
		return svc, nil
	}
	container.GetConsole = func() (*ui.Console, error) {
		return nil, errors.New("console boom")
	}
}

// withFakeHistoryRerunContainer swaps container.GetHistoryService, container.GetRunService, and
// container.GetConsole for the duration of the test, restoring the originals on cleanup. Used for
// RunRerunHistory / RunLastHistory, which need both services.
func withFakeHistoryRerunContainer(t *testing.T, histSvc services.HistoryService, histErr error, runSvc services.RunService, runErr error) {
	t.Helper()

	origHist := container.GetHistoryService
	origRun := container.GetRunService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetHistoryService = origHist
		container.GetRunService = origRun
		container.GetConsole = origConsole
	})

	container.GetHistoryService = func() (services.HistoryService, error) {
		return histSvc, histErr
	}
	container.GetRunService = func() (services.RunService, error) {
		return runSvc, runErr
	}
	container.GetConsole = func() (*ui.Console, error) {
		return ui.NewConsole(ui.NewPalette(settings.UIColors{}), false, "never", 0, 0, 0), nil
	}
}

// withFakeHistoryClearContainer swaps container.GetHistoryService, container.GetConsole, and
// container.GetPrompter for the duration of the test, restoring the originals on cleanup. Used for
// RunClearHistory.
func withFakeHistoryClearContainer(t *testing.T, svc services.HistoryService, svcErr error) {
	t.Helper()

	origSvc := container.GetHistoryService
	origConsole := container.GetConsole
	origPrompter := container.GetPrompter
	t.Cleanup(func() {
		container.GetHistoryService = origSvc
		container.GetConsole = origConsole
		container.GetPrompter = origPrompter
	})

	container.GetHistoryService = func() (services.HistoryService, error) {
		return svc, svcErr
	}
	container.GetConsole = func() (*ui.Console, error) {
		return ui.NewConsole(ui.NewPalette(settings.UIColors{}), false, "never", 0, 0, 0), nil
	}
	container.GetPrompter = freshTestPrompter
}

func sampleHistoryEntry(alias string) models.CommandHistory {
	return models.CommandHistory{
		ID:        "hist-1",
		Alias:     alias,
		Template:  "echo <name>",
		Resolved:  "echo alice",
		RanAt:     time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		ProfileID: 1,
	}
}

// --- RunListHistory ---

func TestRunListHistory_NoFilters(t *testing.T) {
	fake := &fakeHistoryService{recentResult: []models.CommandHistory{sampleHistoryEntry("build")}}
	withFakeHistoryContainer(t, fake, nil)

	out := captureStdout(t, func() {
		require.NoError(t, RunListHistory(ListHistoryArgs{}))
	})

	assert.Nil(t, fake.recentAlias, "no --alias flag should pass nil, not an empty-string pointer")
	assert.Nil(t, fake.recentLimit, "no --limit flag should pass nil")
	assert.Nil(t, fake.recentProfile, "no --profile flag should pass nil")
	assert.Contains(t, out, "build")
}

func TestRunListHistory_WithAlias(t *testing.T) {
	fake := &fakeHistoryService{recentResult: []models.CommandHistory{sampleHistoryEntry("build")}}
	withFakeHistoryContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunListHistory(ListHistoryArgs{Alias: "build", HasAlias: true}))
	})

	require.NotNil(t, fake.recentAlias)
	assert.Equal(t, "build", *fake.recentAlias)
}

func TestRunListHistory_WithLimit(t *testing.T) {
	fake := &fakeHistoryService{}
	withFakeHistoryContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunListHistory(ListHistoryArgs{ListParams: ListParams{Limit: 5, HasLimit: true}}))
	})

	require.NotNil(t, fake.recentLimit)
	assert.Equal(t, 5, *fake.recentLimit)
}

func TestRunListHistory_WithProfile(t *testing.T) {
	fake := &fakeHistoryService{}
	withFakeHistoryContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunListHistory(ListHistoryArgs{ListParams: ListParams{Profile: "work"}}))
	})

	require.NotNil(t, fake.recentProfile)
	assert.Equal(t, "work", *fake.recentProfile)
}

func TestRunListHistory_ServiceError(t *testing.T) {
	fake := &fakeHistoryService{recentErr: errors.New("db error")}
	withFakeHistoryContainer(t, fake, nil)

	err := RunListHistory(ListHistoryArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting recent history")
	assert.Contains(t, err.Error(), "db error")
}

func TestRunListHistory_HistoryServiceUnavailable(t *testing.T) {
	withFakeHistoryContainer(t, nil, errors.New("db down"))

	err := RunListHistory(ListHistoryArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting history service")
	assert.Contains(t, err.Error(), "db down")
}

func TestRunListHistory_ConsoleUnavailable(t *testing.T) {
	fake := &fakeHistoryService{}
	withFailingHistoryConsole(t, fake)

	err := RunListHistory(ListHistoryArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

// --- RunShowHistory ---

func TestRunShowHistory_Success(t *testing.T) {
	entry := sampleHistoryEntry("build")
	fake := &fakeHistoryService{refResult: &entry}
	withFakeHistoryContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunShowHistory("3", nil, nil)
		require.NoError(t, err)
	})

	assert.Equal(t, "3", fake.refRef)
	assert.Nil(t, fake.refAlias)
	assert.Nil(t, fake.refProfile)
	assert.Contains(t, out, "build")
}

func TestRunShowHistory_WithAliasAndProfile(t *testing.T) {
	entry := sampleHistoryEntry("build")
	fake := &fakeHistoryService{refResult: &entry}
	withFakeHistoryContainer(t, fake, nil)

	alias := "build"
	profile := "work"
	captureStdout(t, func() {
		require.NoError(t, RunShowHistory("hist-1", &alias, &profile))
	})

	require.NotNil(t, fake.refAlias)
	assert.Equal(t, "build", *fake.refAlias)
	require.NotNil(t, fake.refProfile)
	assert.Equal(t, "work", *fake.refProfile)
}

func TestRunShowHistory_ServiceError(t *testing.T) {
	fake := &fakeHistoryService{refErr: errors.New("not found")}
	withFakeHistoryContainer(t, fake, nil)

	err := RunShowHistory("99", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting history entry")
	assert.Contains(t, err.Error(), "not found")
}

func TestRunShowHistory_HistoryServiceUnavailable(t *testing.T) {
	withFakeHistoryContainer(t, nil, errors.New("db down"))

	err := RunShowHistory("1", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting history service")
}

func TestRunShowHistory_ConsoleUnavailable(t *testing.T) {
	entry := sampleHistoryEntry("build")
	fake := &fakeHistoryService{refResult: &entry}
	withFailingHistoryConsole(t, fake)

	err := RunShowHistory("1", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

// --- RunRerunHistory ---

func TestRunRerunHistory_Success(t *testing.T) {
	entry := sampleHistoryEntry("build")
	histFake := &fakeHistoryService{refResult: &entry, varsResult: map[string]string{"name": "alice"}}
	runFake := &fakeRunService{runResult: runtime.ExecutionResult{Command: "echo alice", ExitCode: 0}}
	withFakeHistoryRerunContainer(t, histFake, nil, runFake, nil)

	err := RunRerunHistory("3", nil, nil)
	require.NoError(t, err)

	assert.Equal(t, "3", histFake.refRef)
	assert.Equal(t, &entry, histFake.varsEntry)
	assert.Equal(t, "build", runFake.runAlias)
	assert.Equal(t, "alice", runFake.runVars["name"])
}

func TestRunRerunHistory_PassesAliasAndProfile(t *testing.T) {
	entry := sampleHistoryEntry("build")
	histFake := &fakeHistoryService{refResult: &entry}
	runFake := &fakeRunService{}
	withFakeHistoryRerunContainer(t, histFake, nil, runFake, nil)

	alias := "build"
	profile := "work"
	require.NoError(t, RunRerunHistory("hist-1", &alias, &profile))

	require.NotNil(t, histFake.refAlias)
	assert.Equal(t, "build", *histFake.refAlias)
	require.NotNil(t, histFake.refProfile)
	assert.Equal(t, "work", *histFake.refProfile)
	require.NotNil(t, runFake.runProfile)
	assert.Equal(t, "work", *runFake.runProfile)
}

func TestRunRerunHistory_GetByRefError(t *testing.T) {
	histFake := &fakeHistoryService{refErr: errors.New("not found")}
	runFake := &fakeRunService{}
	withFakeHistoryRerunContainer(t, histFake, nil, runFake, nil)

	err := RunRerunHistory("99", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting history entry")
	assert.Contains(t, err.Error(), "not found")
}

func TestRunRerunHistory_GetVariablesError(t *testing.T) {
	entry := sampleHistoryEntry("build")
	histFake := &fakeHistoryService{refResult: &entry, varsErr: errors.New("bad json")}
	runFake := &fakeRunService{}
	withFakeHistoryRerunContainer(t, histFake, nil, runFake, nil)

	err := RunRerunHistory("1", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting history entry variables")
	assert.Contains(t, err.Error(), "bad json")
}

func TestRunRerunHistory_RunError(t *testing.T) {
	entry := sampleHistoryEntry("build")
	histFake := &fakeHistoryService{refResult: &entry}
	runFake := &fakeRunService{runErr: errors.New("command failed")}
	withFakeHistoryRerunContainer(t, histFake, nil, runFake, nil)

	err := RunRerunHistory("1", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "running history entry")
	assert.Contains(t, err.Error(), "command failed")
}

func TestRunRerunHistory_HistoryServiceUnavailable(t *testing.T) {
	runFake := &fakeRunService{}
	withFakeHistoryRerunContainer(t, nil, errors.New("db down"), runFake, nil)

	err := RunRerunHistory("1", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting history service")
	assert.Contains(t, err.Error(), "db down")
}

func TestRunRerunHistory_RunServiceUnavailable(t *testing.T) {
	histFake := &fakeHistoryService{}
	withFakeHistoryRerunContainer(t, histFake, nil, nil, errors.New("exec unavailable"))

	err := RunRerunHistory("1", nil, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting run service")
	assert.Contains(t, err.Error(), "exec unavailable")
}

// --- RunLastHistory ---

func TestRunLastHistory_DelegatesToRerunWithIndexOne(t *testing.T) {
	entry := sampleHistoryEntry("build")
	histFake := &fakeHistoryService{refResult: &entry}
	runFake := &fakeRunService{}
	withFakeHistoryRerunContainer(t, histFake, nil, runFake, nil)

	require.NoError(t, RunLastHistory())

	assert.Equal(t, "1", histFake.refRef)
	assert.Nil(t, histFake.refAlias)
	assert.Nil(t, histFake.refProfile)
}

func TestRunLastHistory_Error(t *testing.T) {
	histFake := &fakeHistoryService{refErr: errors.New("no history")}
	runFake := &fakeRunService{}
	withFakeHistoryRerunContainer(t, histFake, nil, runFake, nil)

	err := RunLastHistory()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no history")
}

// --- RunClearHistory ---

func TestRunClearHistory_YesSkipsPrompt(t *testing.T) {
	fake := &fakeHistoryService{clearCount: 4}
	withFakeHistoryClearContainer(t, fake, nil)
	// Force the prompter to fail; it must never be consulted when Yes is set.
	container.GetPrompter = func() (*prompts.Prompter, error) {
		return nil, errors.New("should not be called")
	}

	out := captureStdout(t, func() {
		err := RunClearHistory(ClearHistoryArgs{Yes: true})
		require.NoError(t, err)
	})

	assert.Nil(t, fake.clearAlias)
	assert.Nil(t, fake.clearProfile)
	assert.Contains(t, out, "4")
}

func TestRunClearHistory_WithAliasAndProfile(t *testing.T) {
	fake := &fakeHistoryService{clearCount: 1}
	withFakeHistoryClearContainer(t, fake, nil)

	captureStdout(t, func() {
		err := RunClearHistory(ClearHistoryArgs{
			Yes: true, Alias: "build", HasAlias: true, Profile: "work", HasProfile: true,
		})
		require.NoError(t, err)
	})

	require.NotNil(t, fake.clearAlias)
	assert.Equal(t, "build", *fake.clearAlias)
	require.NotNil(t, fake.clearProfile)
	assert.Equal(t, "work", *fake.clearProfile)
}

func TestRunClearHistory_NoAliasOrProfilePassesNil(t *testing.T) {
	fake := &fakeHistoryService{}
	withFakeHistoryClearContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunClearHistory(ClearHistoryArgs{Yes: true}))
	})

	assert.Nil(t, fake.clearAlias)
	assert.Nil(t, fake.clearProfile)
}

func TestRunClearHistory_ServiceError(t *testing.T) {
	fake := &fakeHistoryService{clearErr: errors.New("db locked")}
	withFakeHistoryClearContainer(t, fake, nil)

	err := RunClearHistory(ClearHistoryArgs{Yes: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "clearing history")
	assert.Contains(t, err.Error(), "db locked")
}

func TestRunClearHistory_HistoryServiceUnavailable(t *testing.T) {
	withFakeHistoryClearContainer(t, nil, errors.New("db down"))

	err := RunClearHistory(ClearHistoryArgs{Yes: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting history service")
	assert.Contains(t, err.Error(), "db down")
}

func TestRunClearHistory_ConsoleUnavailable(t *testing.T) {
	fake := &fakeHistoryService{}
	origSvc := container.GetHistoryService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetHistoryService = origSvc
		container.GetConsole = origConsole
	})
	container.GetHistoryService = func() (services.HistoryService, error) {
		return fake, nil
	}
	container.GetConsole = func() (*ui.Console, error) {
		return nil, errors.New("console boom")
	}

	err := RunClearHistory(ClearHistoryArgs{Yes: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

func TestRunClearHistory_PrompterUnavailableWhenNotYes(t *testing.T) {
	fake := &fakeHistoryService{}
	withFakeHistoryClearContainer(t, fake, nil)
	container.GetPrompter = func() (*prompts.Prompter, error) {
		return nil, errors.New("terminal unavailable")
	}

	err := RunClearHistory(ClearHistoryArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting prompter")
	assert.Contains(t, err.Error(), "terminal unavailable")
	assert.False(t, fake.clearCalled, "history should not be cleared before confirmation is obtained")
}
