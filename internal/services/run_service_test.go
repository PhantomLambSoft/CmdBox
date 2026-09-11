package services

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
	"github.com/PhantomLambSoft/CmdBox/internal/repository/validate"
	"github.com/PhantomLambSoft/CmdBox/internal/resolve"
	"github.com/PhantomLambSoft/CmdBox/internal/runtime"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

// --- test fixtures ---

// fakeSettingsProvider lets tests control the Settings snapshot RunService reads without touching disk.
type fakeSettingsProvider struct {
	settings settings.Settings
}

func (f *fakeSettingsProvider) Get() settings.Settings {
	return f.settings
}

func defaultTestSettings() settings.Settings {
	return settings.Settings{
		ExecutionSettings: settings.DefaultExecutionSettings(),
		History:           settings.DefaultHistorySettings(),
	}
}

// spyCommandRepo wraps a real CommandRepository and, when recordUseErr is set, forces RecordUse to fail
// without disturbing any other behavior - used to exercise RunService's "log and continue" paths.
type spyCommandRepo struct {
	repository.CommandRepository
	recordUseErr   error
	recordUseCalls []uint
}

func (s *spyCommandRepo) RecordUse(commandID uint) error {
	s.recordUseCalls = append(s.recordUseCalls, commandID)
	if s.recordUseErr != nil {
		return s.recordUseErr
	}
	return s.CommandRepository.RecordUse(commandID)
}

// spyHistoryRepo wraps a real HistoryRepository, forcing Record to fail when recordErr is set.
type spyHistoryRepo struct {
	repository.HistoryRepository
	recordErr error
}

func (s *spyHistoryRepo) Record(entry repository.HistoryEntry) (*models.CommandHistory, error) {
	if s.recordErr != nil {
		return nil, s.recordErr
	}
	return s.HistoryRepository.Record(entry)
}

// spyProfileRepo wraps a real ProfileRepository, forcing GetActiveVariableProfile or RecordUse to fail
// when the corresponding error field is set.
type spyProfileRepo struct {
	repository.ProfileRepository
	activeVariableProfileErr error
	recordUseErr             error
	recordUseCalls           []models.Profile
}

func (s *spyProfileRepo) GetActiveVariableProfile() (*models.Profile, error) {
	if s.activeVariableProfileErr != nil {
		return nil, s.activeVariableProfileErr
	}
	return s.ProfileRepository.GetActiveVariableProfile()
}

func (s *spyProfileRepo) RecordUse(profile models.Profile) error {
	s.recordUseCalls = append(s.recordUseCalls, profile)
	if s.recordUseErr != nil {
		return s.recordUseErr
	}
	return s.ProfileRepository.RecordUse(profile)
}

type runServiceHarness struct {
	svc              *RunService
	cmdRepo          repository.CommandRepository
	varRepo          repository.VariableRepository
	profileRepo      repository.ProfileRepository
	historyRepo      repository.HistoryRepository
	settingsProvider *fakeSettingsProvider
	spyCmd           *spyCommandRepo
	spyHistory       *spyHistoryRepo
	spyProfile       *spyProfileRepo
	db               *gorm.DB
}

// setupRunServiceTest builds a RunService backed by a real in-memory sqlite DB (matching the pattern used
// throughout this package's other service tests) plus a real resolve.Resolver and runtime.Executor, since
// both are concrete types on RunService rather than interfaces and cannot be swapped for mocks. strict
// controls the resolver's strict mode (affects how unresolved command references are treated).
func setupRunServiceTest(t *testing.T, strict bool, withHistory bool) *runServiceHarness {
	t.Helper()

	dsn := fmt.Sprintf("file:run-service-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			t.Fatalf("close sql db: %v", closeErr)
		}
	})

	if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	if err := db.AutoMigrate(
		&models.Profile{},
		&models.ProfileState{},
		&models.Command{},
		&models.Variable{},
		&models.Tag{},
		&models.CommandTag{},
		&models.CommandHistory{},
	); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}

	defaultDescription := "Automatically created default profile."
	defaultProfile := models.Profile{Name: repository.DefaultProfileName, Description: &defaultDescription}
	if err := db.Create(&defaultProfile).Error; err != nil {
		t.Fatalf("seed default profile: %v", err)
	}

	state := models.ProfileState{
		ActiveCommandProfileID:  defaultProfile.ID,
		ActiveVariableProfileID: defaultProfile.ID,
		ActiveSettingsProfileID: defaultProfile.ID,
	}
	if err := db.Create(&state).Error; err != nil {
		t.Fatalf("seed profile state: %v", err)
	}

	profileRepo := repository.NewProfileRepository(db, validate.NewProfileValidator(nil))
	cmdRepo := repository.NewCommandRepository(db, profileRepo, validate.NewCommandValidator(nil))
	varRepo := repository.NewVariableRepository(db, profileRepo, validate.NewVariableValidator(nil))

	spyCmd := &spyCommandRepo{CommandRepository: cmdRepo}
	spyProfile := &spyProfileRepo{ProfileRepository: profileRepo}

	var historyRepo repository.HistoryRepository
	var spyHistory *spyHistoryRepo
	if withHistory {
		realHistoryRepo := repository.NewHistoryRepository(db, profileRepo)
		spyHistory = &spyHistoryRepo{HistoryRepository: realHistoryRepo}
		historyRepo = spyHistory
	}

	lookup := resolve.NewLookup(spyCmd, varRepo)
	resolver := resolve.NewResolver(lookup, strict, 10)
	executor := *runtime.NewExecutor()

	settingsProvider := &fakeSettingsProvider{settings: defaultTestSettings()}

	svc := NewRunService(RunServiceConfig{
		CmdRepo:          spyCmd,
		Resolver:         resolver,
		Executor:         executor,
		ProfileRepo:      spyProfile,
		HistoryRepo:      historyRepo,
		SettingsProvider: settingsProvider,
	})

	return &runServiceHarness{
		svc:              svc,
		cmdRepo:          spyCmd,
		varRepo:          varRepo,
		profileRepo:      spyProfile,
		historyRepo:      historyRepo,
		settingsProvider: settingsProvider,
		spyCmd:           spyCmd,
		spyHistory:       spyHistory,
		spyProfile:       spyProfile,
		db:               db,
	}
}

func mustCreateRunServiceCommand(t *testing.T, cmdRepo repository.CommandRepository, cfg repository.CommandCreateConfig) *models.Command {
	t.Helper()
	cmd, err := cmdRepo.Create(cfg)
	if err != nil {
		t.Fatalf("create command %+v: %v", cfg, err)
	}
	return cmd
}

func mustCreateRunServiceVariable(t *testing.T, varRepo repository.VariableRepository, cfg repository.VariableCreateConfig) *models.Variable {
	t.Helper()
	v, err := varRepo.Create(cfg)
	if err != nil {
		t.Fatalf("create variable %+v: %v", cfg, err)
	}
	return v
}

func mustCreateRunServiceProfile(t *testing.T, profileRepo repository.ProfileRepository, name string) *models.Profile {
	t.Helper()
	p, err := profileRepo.Create(name, nil)
	if err != nil {
		t.Fatalf("create profile %q: %v", name, err)
	}
	return p
}

func intPtr(i int) *int    { return &i }
func boolPtr(b bool) *bool { return &b }

// --- resolveProfile ---

func TestRunServiceResolveProfile(t *testing.T) {
	t.Run("nil profileName returns nil profile", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		profile, err := h.svc.resolveProfile(nil)
		if err != nil {
			t.Fatalf("resolveProfile() error = %v", err)
		}
		if profile != nil {
			t.Fatalf("resolveProfile() = %+v, want nil", profile)
		}
	})

	t.Run("empty string profileName returns nil profile", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		profile, err := h.svc.resolveProfile(strPtr(""))
		if err != nil {
			t.Fatalf("resolveProfile() error = %v", err)
		}
		if profile != nil {
			t.Fatalf("resolveProfile() = %+v, want nil", profile)
		}
	})

	t.Run("known profileName resolves the profile", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		mustCreateRunServiceProfile(t, h.profileRepo, "work")

		profile, err := h.svc.resolveProfile(strPtr("work"))
		if err != nil {
			t.Fatalf("resolveProfile() error = %v", err)
		}
		if profile == nil || profile.Name != "work" {
			t.Fatalf("resolveProfile() = %+v, want profile named work", profile)
		}
	})

	t.Run("unknown profileName returns wrapped ErrProfileNotFound", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		_, err := h.svc.resolveProfile(strPtr("missing"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("resolveProfile() error = %v, want ErrProfileNotFound", err)
		}
	})
}

// --- buildContext ---

func TestRunServiceBuildContext(t *testing.T) {
	t.Run("uses settings defaults when nothing else is set", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		h.settingsProvider.settings.ExecutionSettings = settings.ExecutionSettings{
			DefaultShell: "cmd.exe", CaptureOutput: true, DefaultVerbose: true,
		}

		ctx, err := h.svc.buildContext(models.Command{}, RuntimeOverrides{})
		if err != nil {
			t.Fatalf("buildContext() error = %v", err)
		}
		if ctx.Shell != "cmd.exe" || !ctx.Capture || !ctx.Verbose {
			t.Fatalf("ctx = %+v, want defaults from settings applied", ctx)
		}
		if ctx.Cwd != "" || ctx.Timeout != 0 || ctx.Emit {
			t.Fatalf("ctx = %+v, want zero cwd/timeout/emit", ctx)
		}
	})

	t.Run("command fields are used when no runtime override given", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		cwd := "C:/work"
		shell := "pwsh"
		timeout := 30
		cmd := models.Command{Cwd: &cwd, Shell: &shell, Timeout: &timeout}

		ctx, err := h.svc.buildContext(cmd, RuntimeOverrides{})
		if err != nil {
			t.Fatalf("buildContext() error = %v", err)
		}
		if ctx.Cwd != cwd || ctx.Shell != shell || ctx.Timeout != timeout {
			t.Fatalf("ctx = %+v, want command's cwd/shell/timeout", ctx)
		}
	})

	t.Run("runtime overrides win over command fields", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		cmdCwd, cmdShell, cmdTimeout := "C:/cmd", "cmd.exe", 10
		cmd := models.Command{Cwd: &cmdCwd, Shell: &cmdShell, Timeout: &cmdTimeout}

		overrides := RuntimeOverrides{
			Cwd:     "C:/override",
			Shell:   strPtr("pwsh"),
			Timeout: intPtr(99),
			Capture: boolPtr(true),
			Verbose: boolPtr(true),
			Emit:    true,
		}

		ctx, err := h.svc.buildContext(cmd, overrides)
		if err != nil {
			t.Fatalf("buildContext() error = %v", err)
		}
		if ctx.Cwd != "C:/override" || ctx.Shell != "pwsh" || ctx.Timeout != 99 {
			t.Fatalf("ctx = %+v, want overrides to win", ctx)
		}
		if !ctx.Capture || !ctx.Verbose || !ctx.Emit {
			t.Fatalf("ctx = %+v, want capture/verbose/emit true", ctx)
		}
	})

	t.Run("runtime env is merged over stored env", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		storedEnv := `{"A":"stored-a","B":"stored-b"}`
		cmd := models.Command{Env: &storedEnv}

		ctx, err := h.svc.buildContext(cmd, RuntimeOverrides{Env: map[string]string{"B": "runtime-b", "C": "runtime-c"}})
		if err != nil {
			t.Fatalf("buildContext() error = %v", err)
		}
		want := map[string]string{"A": "stored-a", "B": "runtime-b", "C": "runtime-c"}
		if len(ctx.Env) != len(want) {
			t.Fatalf("ctx.Env = %v, want %v", ctx.Env, want)
		}
		for k, v := range want {
			if ctx.Env[k] != v {
				t.Fatalf("ctx.Env[%q] = %q, want %q", k, ctx.Env[k], v)
			}
		}
	})

	t.Run("invalid stored env JSON returns error", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		badEnv := "not-json"
		cmd := models.Command{Env: &badEnv}

		_, err := h.svc.buildContext(cmd, RuntimeOverrides{})
		if err == nil {
			t.Fatal("buildContext() error = nil, want error")
		}
	})
}

// --- recordHistory ---

func TestRunServiceRecordHistory(t *testing.T) {
	t.Run("disabled history returns nil without touching the repo", func(t *testing.T) {
		h := setupRunServiceTest(t, false, true)
		h.settingsProvider.settings.History.Enabled = false

		got, err := h.svc.recordHistory(RecordHistoryInput{Alias: "a", Template: "t", Resolved: "r"})
		if err != nil {
			t.Fatalf("recordHistory() error = %v", err)
		}
		if got != nil {
			t.Fatalf("recordHistory() = %+v, want nil", got)
		}
	})

	t.Run("nil history repo returns nil even when enabled", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		h.settingsProvider.settings.History.Enabled = true

		got, err := h.svc.recordHistory(RecordHistoryInput{Alias: "a", Template: "t", Resolved: "r"})
		if err != nil {
			t.Fatalf("recordHistory() error = %v", err)
		}
		if got != nil {
			t.Fatalf("recordHistory() = %+v, want nil", got)
		}
	})

	t.Run("records an entry with the expected fields", func(t *testing.T) {
		h := setupRunServiceTest(t, false, true)
		h.settingsProvider.settings.History.Enabled = true
		h.settingsProvider.settings.History.LimitPerCommand = 42

		entry, err := h.svc.recordHistory(RecordHistoryInput{
			Alias:       "build",
			Template:    "go build <pkg>",
			Resolved:    "go build ./...",
			RuntimeVars: map[string]string{"pkg": "./..."},
			ExitCode:    3,
		})
		if err != nil {
			t.Fatalf("recordHistory() error = %v", err)
		}
		if entry == nil {
			t.Fatal("recordHistory() = nil, want an entry")
		}
		if entry.Alias != "build" || entry.Template != "go build <pkg>" || entry.Resolved != "go build ./..." {
			t.Fatalf("entry = %+v, want alias/template/resolved to match input", entry)
		}
		if entry.ExitCode == nil || *entry.ExitCode != 3 {
			t.Fatalf("entry.ExitCode = %v, want 3", entry.ExitCode)
		}
		if entry.VariablesUsed == nil || !strings.Contains(*entry.VariablesUsed, `"pkg":"./..."`) {
			t.Fatalf("entry.VariablesUsed = %v, want marshaled runtime vars", entry.VariablesUsed)
		}
	})

	t.Run("empty runtime vars leaves VariablesUsed nil", func(t *testing.T) {
		h := setupRunServiceTest(t, false, true)
		h.settingsProvider.settings.History.Enabled = true

		entry, err := h.svc.recordHistory(RecordHistoryInput{Alias: "a", Template: "t", Resolved: "r"})
		if err != nil {
			t.Fatalf("recordHistory() error = %v", err)
		}
		if entry.VariablesUsed != nil {
			t.Fatalf("entry.VariablesUsed = %v, want nil", *entry.VariablesUsed)
		}
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		h := setupRunServiceTest(t, false, true)
		h.settingsProvider.settings.History.Enabled = true
		h.spyHistory.recordErr = errors.New("boom")

		_, err := h.svc.recordHistory(RecordHistoryInput{Alias: "a", Template: "t", Resolved: "r"})
		if err == nil {
			t.Fatal("recordHistory() error = nil, want error")
		}
	})
}

// --- recordVariableProfileUse ---

func TestRunServiceRecordVariableProfileUse(t *testing.T) {
	t.Run("no trace steps is a no-op", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		if err := h.svc.recordVariableProfileUse(resolve.Result{}); err != nil {
			t.Fatalf("recordVariableProfileUse() error = %v", err)
		}
		if len(h.spyProfile.recordUseCalls) != 0 {
			t.Fatalf("RecordUse called %d times, want 0", len(h.spyProfile.recordUseCalls))
		}
	})

	t.Run("runtime-sourced variable is not counted as stored use", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		trace := []resolve.TraceStep{{Kind: resolve.RefKindVariable, Key: "name", Source: "runtime"}}

		if err := h.svc.recordVariableProfileUse(resolve.Result{Trace: trace}); err != nil {
			t.Fatalf("recordVariableProfileUse() error = %v", err)
		}
		if len(h.spyProfile.recordUseCalls) != 0 {
			t.Fatalf("RecordUse called %d times, want 0", len(h.spyProfile.recordUseCalls))
		}
	})

	t.Run("command trace steps are not counted as stored variable use", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		trace := []resolve.TraceStep{{Kind: resolve.RefKindCommand, Key: "build", Source: ""}}

		if err := h.svc.recordVariableProfileUse(resolve.Result{Trace: trace}); err != nil {
			t.Fatalf("recordVariableProfileUse() error = %v", err)
		}
		if len(h.spyProfile.recordUseCalls) != 0 {
			t.Fatalf("RecordUse called %d times, want 0", len(h.spyProfile.recordUseCalls))
		}
	})

	t.Run("stored variable use records against the active variable profile", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		trace := []resolve.TraceStep{{Kind: resolve.RefKindVariable, Key: "name", Source: "stored"}}

		if err := h.svc.recordVariableProfileUse(resolve.Result{Trace: trace}); err != nil {
			t.Fatalf("recordVariableProfileUse() error = %v", err)
		}
		if len(h.spyProfile.recordUseCalls) != 1 {
			t.Fatalf("RecordUse called %d times, want 1", len(h.spyProfile.recordUseCalls))
		}
		if h.spyProfile.recordUseCalls[0].Name != repository.DefaultProfileName {
			t.Fatalf("RecordUse called with %+v, want default profile", h.spyProfile.recordUseCalls[0])
		}
	})

	t.Run("propagates GetActiveVariableProfile errors", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		h.spyProfile.activeVariableProfileErr = errors.New("boom")
		trace := []resolve.TraceStep{{Kind: resolve.RefKindVariable, Key: "name", Source: "stored"}}

		err := h.svc.recordVariableProfileUse(resolve.Result{Trace: trace})
		if err == nil {
			t.Fatal("recordVariableProfileUse() error = nil, want error")
		}
	})

	t.Run("propagates RecordUse errors", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		h.spyProfile.recordUseErr = errors.New("boom")
		trace := []resolve.TraceStep{{Kind: resolve.RefKindVariable, Key: "name", Source: "stored"}}

		err := h.svc.recordVariableProfileUse(resolve.Result{Trace: trace})
		if err == nil {
			t.Fatal("recordVariableProfileUse() error = nil, want error")
		}
	})
}

// --- CollectMissingVars ---

func TestRunServiceCollectMissingVars(t *testing.T) {
	t.Run("reports unresolved variables", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		mustCreateRunServiceVariable(t, h.varRepo, repository.VariableCreateConfig{Name: "known", Value: "v"})
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "build", Template: "<known> <missing>"})

		missing, err := h.svc.CollectMissingVars("build", nil, nil)
		if err != nil {
			t.Fatalf("CollectMissingVars() error = %v", err)
		}
		if len(missing) != 1 || missing[0] != "missing" {
			t.Fatalf("missing = %v, want [missing]", missing)
		}
	})

	t.Run("runtime vars satisfy otherwise-missing variables", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "build", Template: "<name>"})

		missing, err := h.svc.CollectMissingVars("build", map[string]string{"name": "value"}, nil)
		if err != nil {
			t.Fatalf("CollectMissingVars() error = %v", err)
		}
		if len(missing) != 0 {
			t.Fatalf("missing = %v, want none", missing)
		}
	})

	t.Run("unknown alias returns error", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		_, err := h.svc.CollectMissingVars("does-not-exist", nil, nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("CollectMissingVars() error = %v, want ErrUnknownAlias", err)
		}
	})

	t.Run("unknown profile name returns error", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		_, err := h.svc.CollectMissingVars("build", nil, strPtr("missing-profile"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("CollectMissingVars() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("scopes command lookup to the named profile", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		other := mustCreateRunServiceProfile(t, h.profileRepo, "other")
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "build", Template: "<x>", ProfileID: &other.ID})

		_, err := h.svc.CollectMissingVars("build", nil, nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("CollectMissingVars() under default profile error = %v, want ErrUnknownAlias", err)
		}

		missing, err := h.svc.CollectMissingVars("build", nil, strPtr("other"))
		if err != nil {
			t.Fatalf("CollectMissingVars() under explicit profile error = %v", err)
		}
		if len(missing) != 1 || missing[0] != "x" {
			t.Fatalf("missing = %v, want [x]", missing)
		}
	})
}

// --- Preview ---

func TestRunServicePreview(t *testing.T) {
	t.Run("resolves the template and builds context without side effects", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		mustCreateRunServiceVariable(t, h.varRepo, repository.VariableCreateConfig{Name: "name", Value: "world"})
		cmd := mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "greet", Template: "echo hello <name>"})

		result, ctx, err := h.svc.Preview("greet", RuntimeOverrides{}, nil, nil)
		if err != nil {
			t.Fatalf("Preview() error = %v", err)
		}
		if result.Text != "echo hello world" {
			t.Fatalf("result.Text = %q, want %q", result.Text, "echo hello world")
		}
		if ctx.Shell == "" {
			t.Fatalf("ctx.Shell = %q, want a resolved shell", ctx.Shell)
		}
		if cmd.Used != 0 {
			t.Fatalf("cmd.Used = %d, want 0 (Preview must not record use)", cmd.Used)
		}
		if len(h.spyProfile.recordUseCalls) != 0 {
			t.Fatalf("ProfileRepository.RecordUse called %d times, want 0 (Preview must not record use)", len(h.spyProfile.recordUseCalls))
		}
	})

	t.Run("propagates profile resolution errors", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		_, _, err := h.svc.Preview("greet", RuntimeOverrides{}, nil, strPtr("missing"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("Preview() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("unknown alias returns error", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		_, _, err := h.svc.Preview("does-not-exist", RuntimeOverrides{}, nil, nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("Preview() error = %v, want ErrUnknownAlias", err)
		}
	})

	t.Run("strict resolver returns error for unresolved command reference", func(t *testing.T) {
		h := setupRunServiceTest(t, true, false)
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "build", Template: "<cmd:missing>"})

		_, _, err := h.svc.Preview("build", RuntimeOverrides{}, nil, nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("Preview() error = %v, want wrapped ErrUnknownAlias", err)
		}
	})

	t.Run("invalid stored env returns error", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "build", Template: "echo hi", Env: map[string]string{}})
		// Corrupt the stored env directly so buildContext's json.Unmarshal fails.
		if err := h.db.Model(&models.Command{}).Where("alias = ?", "build").Update("env", "not-json").Error; err != nil {
			t.Fatalf("corrupt env: %v", err)
		}

		_, _, err := h.svc.Preview("build", RuntimeOverrides{}, nil, nil)
		if err == nil {
			t.Fatal("Preview() error = nil, want error")
		}
	})
}

// --- Run ---

func TestRunServiceRun(t *testing.T) {
	t.Run("happy path executes, records history, and bumps command use", func(t *testing.T) {
		h := setupRunServiceTest(t, false, true)
		h.settingsProvider.settings.History.Enabled = true
		mustCreateRunServiceVariable(t, h.varRepo, repository.VariableCreateConfig{Name: "name", Value: "world"})
		cmd := mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "greet", Template: "echo hello <name>"})

		result, err := h.svc.Run("greet", RuntimeOverrides{Capture: boolPtr(true)}, nil, nil)
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if result.ExitCode != 0 {
			t.Fatalf("result.ExitCode = %d, want 0 (stderr=%q)", result.ExitCode, result.Stderr)
		}
		if !strings.Contains(result.Stdout, "hello") || !strings.Contains(result.Stdout, "world") {
			t.Fatalf("result.Stdout = %q, want to contain %q and %q", result.Stdout, "hello", "world")
		}

		updated, err := h.cmdRepo.GetByID(cmd.ID, nil)
		if err != nil {
			t.Fatalf("GetByID() error = %v", err)
		}
		if updated.Used != 1 {
			t.Fatalf("command Used = %d, want 1", updated.Used)
		}

		recent, err := h.historyRepo.GetRecent(nil, 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(recent) != 1 || recent[0].Alias != "greet" {
			t.Fatalf("history = %+v, want a single greet entry", recent)
		}

		// The template used a stored variable, so use should also be recorded against the active
		// variable profile.
		if len(h.spyProfile.recordUseCalls) != 1 {
			t.Fatalf("ProfileRepository.RecordUse called %d times, want 1", len(h.spyProfile.recordUseCalls))
		}
	})

	t.Run("runtime-supplied variable does not record profile use", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "greet", Template: "echo hello <name>"})

		_, err := h.svc.Run("greet", RuntimeOverrides{Capture: boolPtr(true)}, map[string]string{"name": "runtime-world"}, nil)
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if len(h.spyProfile.recordUseCalls) != 0 {
			t.Fatalf("ProfileRepository.RecordUse called %d times, want 0 (variable came from runtimeVars)", len(h.spyProfile.recordUseCalls))
		}
	})

	t.Run("disabled history does not create an entry", func(t *testing.T) {
		h := setupRunServiceTest(t, false, true)
		h.settingsProvider.settings.History.Enabled = false
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "greet", Template: "echo hi"})

		if _, err := h.svc.Run("greet", RuntimeOverrides{Capture: boolPtr(true)}, nil, nil); err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		recent, err := h.historyRepo.GetRecent(nil, 0, nil)
		if err != nil {
			t.Fatalf("GetRecent() error = %v", err)
		}
		if len(recent) != 0 {
			t.Fatalf("history = %+v, want none while disabled", recent)
		}
	})

	t.Run("nonzero exit code is captured without returning an error", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "fail", Template: "exit 3"})

		result, err := h.svc.Run("fail", RuntimeOverrides{Capture: boolPtr(true)}, nil, nil)
		if err != nil {
			t.Fatalf("Run() error = %v", err)
		}
		if result.ExitCode != 3 {
			t.Fatalf("result.ExitCode = %d, want 3", result.ExitCode)
		}
	})

	t.Run("profile resolution error is returned", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		_, err := h.svc.Run("greet", RuntimeOverrides{}, nil, strPtr("missing"))
		if !errors.Is(err, repository.ErrProfileNotFound) {
			t.Fatalf("Run() error = %v, want ErrProfileNotFound", err)
		}
	})

	t.Run("unknown alias returns error", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		_, err := h.svc.Run("does-not-exist", RuntimeOverrides{}, nil, nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("Run() error = %v, want ErrUnknownAlias", err)
		}
	})

	t.Run("strict resolver error is returned and nothing executes", func(t *testing.T) {
		h := setupRunServiceTest(t, true, true)
		h.settingsProvider.settings.History.Enabled = true
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "build", Template: "<cmd:missing>"})

		_, err := h.svc.Run("build", RuntimeOverrides{}, nil, nil)
		if !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("Run() error = %v, want wrapped ErrUnknownAlias", err)
		}
		recent, gErr := h.historyRepo.GetRecent(nil, 0, nil)
		if gErr != nil {
			t.Fatalf("GetRecent() error = %v", gErr)
		}
		if len(recent) != 0 {
			t.Fatalf("history = %+v, want none when resolution fails before execution", recent)
		}
	})

	t.Run("invalid stored env returns error before executing", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "build", Template: "echo hi", Env: map[string]string{}})
		if err := h.db.Model(&models.Command{}).Where("alias = ?", "build").Update("env", "not-json").Error; err != nil {
			t.Fatalf("corrupt env: %v", err)
		}

		_, err := h.svc.Run("build", RuntimeOverrides{}, nil, nil)
		if err == nil {
			t.Fatal("Run() error = nil, want error")
		}
	})

	t.Run("executor failure is returned", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "build", Template: "echo hi"})

		// A nonexistent working directory makes the underlying process fail to start at all
		// (rather than merely exiting nonzero), which is the error path executor.Run can return.
		_, err := h.svc.Run("build", RuntimeOverrides{Cwd: "Z:/definitely/does/not/exist/cmdbox-test"}, nil, nil)
		if err == nil {
			t.Fatal("Run() error = nil, want error")
		}
		if !strings.Contains(err.Error(), "running command") {
			t.Fatalf("err = %v, want it to wrap the executor failure", err)
		}
	})

	t.Run("command use failure is logged but does not fail Run", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "greet", Template: "echo hi"})
		h.spyCmd.recordUseErr = errors.New("boom")

		if _, err := h.svc.Run("greet", RuntimeOverrides{Capture: boolPtr(true)}, nil, nil); err != nil {
			t.Fatalf("Run() error = %v, want nil (RecordUse failures must only be logged)", err)
		}
	})

	t.Run("history recording failure is logged but does not fail Run", func(t *testing.T) {
		h := setupRunServiceTest(t, false, true)
		h.settingsProvider.settings.History.Enabled = true
		h.spyHistory.recordErr = errors.New("boom")
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "greet", Template: "echo hi"})

		if _, err := h.svc.Run("greet", RuntimeOverrides{Capture: boolPtr(true)}, nil, nil); err != nil {
			t.Fatalf("Run() error = %v, want nil (history failures must only be logged)", err)
		}
	})

	t.Run("variable profile use failure is logged but does not fail Run", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		mustCreateRunServiceVariable(t, h.varRepo, repository.VariableCreateConfig{Name: "name", Value: "world"})
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "greet", Template: "echo hello <name>"})
		h.spyProfile.activeVariableProfileErr = errors.New("boom")

		if _, err := h.svc.Run("greet", RuntimeOverrides{Capture: boolPtr(true)}, nil, nil); err != nil {
			t.Fatalf("Run() error = %v, want nil (variable profile use failures must only be logged)", err)
		}
	})

	t.Run("scopes command lookup to the named profile", func(t *testing.T) {
		h := setupRunServiceTest(t, false, false)
		other := mustCreateRunServiceProfile(t, h.profileRepo, "other")
		mustCreateRunServiceCommand(t, h.cmdRepo, repository.CommandCreateConfig{Alias: "greet", Template: "echo hi", ProfileID: &other.ID})

		if _, err := h.svc.Run("greet", RuntimeOverrides{}, nil, nil); !errors.Is(err, repository.ErrUnknownAlias) {
			t.Fatalf("Run() under default profile error = %v, want ErrUnknownAlias", err)
		}

		if _, err := h.svc.Run("greet", RuntimeOverrides{Capture: boolPtr(true)}, nil, strPtr("other")); err != nil {
			t.Fatalf("Run() under explicit profile error = %v", err)
		}
	})
}
