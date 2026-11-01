package handlers

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

// fakeExportProfileRepo is a hand-rolled stub implementing repository.ProfileRepository. Export
// flows only ever call GetActiveCommandProfile/GetActiveVariableProfile (when no explicit profile
// is given), so every other method is an unused stub.
type fakeExportProfileRepo struct {
	activeCmdProfile *models.Profile
	activeCmdErr     error
	activeVarProfile *models.Profile
	activeVarErr     error
}

func (f *fakeExportProfileRepo) WithTx(tx *gorm.DB) repository.ProfileRepository { return f }

func (f *fakeExportProfileRepo) Create(name string, description *string) (*models.Profile, error) {
	return nil, nil
}

func (f *fakeExportProfileRepo) GetByName(name string) (*models.Profile, error) { return nil, nil }

func (f *fakeExportProfileRepo) GetByID(id uint) (*models.Profile, error) { return nil, nil }

func (f *fakeExportProfileRepo) ListAll(orderBy string, limit int) ([]models.Profile, error) {
	return nil, nil
}

func (f *fakeExportProfileRepo) Search(query string, fields []string, limit int) ([]models.Profile, error) {
	return nil, nil
}

func (f *fakeExportProfileRepo) Update(profile *models.Profile, input repository.ProfileUpdateConfig) (*models.Profile, error) {
	return nil, nil
}

func (f *fakeExportProfileRepo) Delete(profile *models.Profile, force bool) (repository.ProfileDeleteResult, error) {
	return repository.ProfileDeleteResult{}, nil
}

func (f *fakeExportProfileRepo) RecordUse(profile models.Profile) error { return nil }

func (f *fakeExportProfileRepo) GetState() (*models.ProfileState, error) { return nil, nil }

func (f *fakeExportProfileRepo) GetStateWithProfiles() (*models.ProfileState, error) {
	return nil, nil
}

func (f *fakeExportProfileRepo) SetActiveProfile(config repository.SetProfileConfig) (*models.ProfileState, error) {
	return nil, nil
}

func (f *fakeExportProfileRepo) GetActiveCommandProfile() (*models.Profile, error) {
	return f.activeCmdProfile, f.activeCmdErr
}

func (f *fakeExportProfileRepo) GetActiveVariableProfile() (*models.Profile, error) {
	return f.activeVarProfile, f.activeVarErr
}

func (f *fakeExportProfileRepo) GetActiveSettingsProfile() (*models.Profile, error) {
	return nil, nil
}

// withFakeExportContainer swaps container.GetExportService and container.GetConsole for the
// duration of the test, restoring the originals on cleanup. The export service it installs is a
// real *services.ExportService wired to the given fakes, so handler tests exercise the actual
// argument translation (RunExportCmds/Vars/All -> services.Export*Config) without touching a real
// database or writing outside of a test's own temp directory.
func withFakeExportContainer(t *testing.T, cmdSvc services.CommandService, varSvc services.VariableService, profileRepo repository.ProfileRepository) {
	t.Helper()

	origSvc := container.GetExportService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetExportService = origSvc
		container.GetConsole = origConsole
	})

	container.GetExportService = func() (*services.ExportService, error) {
		return services.NewExportService(cmdSvc, varSvc, profileRepo), nil
	}
	container.GetConsole = func() (*ui.Console, error) {
		return ui.NewConsole(ui.NewPalette(settings.UIColors{}), false, "never", 0, 0, 0), nil
	}
}

// withFailingExportService makes container.GetExportService return an error, simulating a
// container resolution failure (e.g. the database being unavailable) before any export logic
// runs.
func withFailingExportService(t *testing.T, svcErr error) {
	t.Helper()

	origSvc := container.GetExportService
	t.Cleanup(func() {
		container.GetExportService = origSvc
	})
	container.GetExportService = func() (*services.ExportService, error) {
		return nil, svcErr
	}
}

// withFailingExportConsole makes container.GetConsole return an error, to exercise handler error
// paths that occur before any console output happens.
func withFailingExportConsole(t *testing.T, cmdSvc services.CommandService, varSvc services.VariableService, profileRepo repository.ProfileRepository) {
	t.Helper()

	origSvc := container.GetExportService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetExportService = origSvc
		container.GetConsole = origConsole
	})

	container.GetExportService = func() (*services.ExportService, error) {
		return services.NewExportService(cmdSvc, varSvc, profileRepo), nil
	}
	container.GetConsole = func() (*ui.Console, error) {
		return nil, errors.New("console boom")
	}
}

// tempExportFile returns a path under a fresh t.TempDir() for use as an --output argument, so
// writing the exported document never touches anything outside the test's own sandbox.
func tempExportFile(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(t.TempDir(), name)
}

// --- RunExportCmds ---

func TestRunExportCmds_Success_ExplicitAlias(t *testing.T) {
	cmdSvc := &fakeCommandService{
		orNilCmd:       &models.Command{Alias: "a", Template: "echo hi"},
		getWithTagsCmd: &models.Command{Alias: "a"},
	}
	varSvc := &fakeVariableService{}
	withFakeExportContainer(t, cmdSvc, varSvc, &fakeExportProfileRepo{})
	out := tempExportFile(t, "cmds.json")

	stdout := captureStdout(t, func() {
		err := RunExportCmds(RunExportCmdArgs{
			Aliases:    []string{"a"},
			Output:     out,
			CmdProfile: "work",
			VarProfile: "work",
		})
		require.NoError(t, err)
	})

	assert.Contains(t, stdout, "Export Complete")
	assert.Contains(t, stdout, "a")
	assert.Contains(t, stdout, "work")
	_, statErr := os.Stat(out)
	assert.NoError(t, statErr, "export file should have been written")
}

func TestRunExportCmds_TagFilterAndFlattenPassedThrough(t *testing.T) {
	cmdSvc := &fakeCommandService{
		listCmds:       []models.Command{{Alias: "a"}},
		orNilCmd:       &models.Command{Alias: "a", Template: "echo hi"},
		getWithTagsCmd: &models.Command{Alias: "a"},
	}
	varSvc := &fakeVariableService{}
	withFakeExportContainer(t, cmdSvc, varSvc, &fakeExportProfileRepo{})
	out := tempExportFile(t, "cmds.json")

	captureStdout(t, func() {
		err := RunExportCmds(RunExportCmdArgs{
			Tag:        "keep",
			Flatten:    true,
			Output:     out,
			CmdProfile: "work",
			VarProfile: "work",
		})
		require.NoError(t, err)
	})

	assert.Equal(t, []string{"keep"}, cmdSvc.listTags)
	require.NotNil(t, cmdSvc.listProfile)
	assert.Equal(t, "work", *cmdSvc.listProfile)
}

func TestRunExportCmds_NoProfile_UsesActiveProfilesFromRepo(t *testing.T) {
	cmdSvc := &fakeCommandService{
		orNilCmd:       &models.Command{Alias: "a", Template: "echo hi"},
		getWithTagsCmd: &models.Command{Alias: "a"},
	}
	varSvc := &fakeVariableService{}
	profileRepo := &fakeExportProfileRepo{
		activeCmdProfile: &models.Profile{Name: "cmd-active"},
		activeVarProfile: &models.Profile{Name: "var-active"},
	}
	withFakeExportContainer(t, cmdSvc, varSvc, profileRepo)
	out := tempExportFile(t, "cmds.json")

	stdout := captureStdout(t, func() {
		err := RunExportCmds(RunExportCmdArgs{Aliases: []string{"a"}, Output: out})
		require.NoError(t, err)
	})

	assert.Contains(t, stdout, "cmd-active")
}

func TestRunExportCmds_MissingAlias_WarnsWithoutError(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	withFakeExportContainer(t, cmdSvc, varSvc, &fakeExportProfileRepo{})
	out := tempExportFile(t, "cmds.json")

	stdout := captureStdout(t, func() {
		err := RunExportCmds(RunExportCmdArgs{
			Aliases:    []string{"ghost"},
			Output:     out,
			CmdProfile: "work",
			VarProfile: "work",
		})
		require.NoError(t, err)
	})

	assert.Contains(t, stdout, "Command ghost not found")
}

func TestRunExportCmds_ServiceError_Wrapped(t *testing.T) {
	cmdSvc := &fakeCommandService{listErr: errors.New("db is locked")}
	varSvc := &fakeVariableService{}
	withFakeExportContainer(t, cmdSvc, varSvc, &fakeExportProfileRepo{})
	out := tempExportFile(t, "cmds.json")

	stdout := captureStdout(t, func() {
		err := RunExportCmds(RunExportCmdArgs{Output: out, CmdProfile: "work", VarProfile: "work"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exporting commands")
		assert.Contains(t, err.Error(), "listing commands")
		assert.Contains(t, err.Error(), "db is locked")
	})

	assert.Empty(t, stdout, "nothing should be printed when the export fails")
	_, statErr := os.Stat(out)
	assert.True(t, os.IsNotExist(statErr), "no file should be written when the export fails")
}

func TestRunExportCmds_ExportServiceUnavailable(t *testing.T) {
	withFailingExportService(t, errors.New("db down"))

	err := RunExportCmds(RunExportCmdArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting export service")
	assert.Contains(t, err.Error(), "db down")
}

func TestRunExportCmds_ConsoleUnavailable(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	withFailingExportConsole(t, cmdSvc, varSvc, &fakeExportProfileRepo{})

	err := RunExportCmds(RunExportCmdArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

// --- RunExportVars ---

func TestRunExportVars_Success_ExplicitName(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{
		orNilVar:       &models.Variable{Name: "v", Value: "1"},
		getWithTagsVar: &models.Variable{Name: "v"},
	}
	withFakeExportContainer(t, cmdSvc, varSvc, &fakeExportProfileRepo{})
	out := tempExportFile(t, "vars.json")

	stdout := captureStdout(t, func() {
		err := RunExportVars(RunExportVarArgs{
			Names:      []string{"v"},
			Output:     out,
			VarProfile: "work",
		})
		require.NoError(t, err)
	})

	assert.Contains(t, stdout, "Export Complete")
	assert.Contains(t, stdout, "v")
	assert.Contains(t, stdout, "work")
	_, statErr := os.Stat(out)
	assert.NoError(t, statErr, "export file should have been written")
}

func TestRunExportVars_TagFilterAndFlattenPassedThrough(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{
		listVars:       []models.Variable{{Name: "v"}},
		orNilVar:       &models.Variable{Name: "v", Value: "1"},
		getWithTagsVar: &models.Variable{Name: "v"},
	}
	withFakeExportContainer(t, cmdSvc, varSvc, &fakeExportProfileRepo{})
	out := tempExportFile(t, "vars.json")

	captureStdout(t, func() {
		err := RunExportVars(RunExportVarArgs{
			Tag:        "keep",
			Flatten:    true,
			Output:     out,
			VarProfile: "work",
		})
		require.NoError(t, err)
	})

	assert.Equal(t, []string{"keep"}, varSvc.listTags)
	require.NotNil(t, varSvc.listProfile)
	assert.Equal(t, "work", *varSvc.listProfile)
}

func TestRunExportVars_NoProfile_UsesActiveProfileFromRepo(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{
		orNilVar:       &models.Variable{Name: "v", Value: "1"},
		getWithTagsVar: &models.Variable{Name: "v"},
	}
	profileRepo := &fakeExportProfileRepo{activeVarProfile: &models.Profile{Name: "var-active"}}
	withFakeExportContainer(t, cmdSvc, varSvc, profileRepo)
	out := tempExportFile(t, "vars.json")

	stdout := captureStdout(t, func() {
		err := RunExportVars(RunExportVarArgs{Names: []string{"v"}, Output: out})
		require.NoError(t, err)
	})

	assert.Contains(t, stdout, "var-active")
}

func TestRunExportVars_MissingName_WarnsWithoutError(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	withFakeExportContainer(t, cmdSvc, varSvc, &fakeExportProfileRepo{})
	out := tempExportFile(t, "vars.json")

	stdout := captureStdout(t, func() {
		err := RunExportVars(RunExportVarArgs{Names: []string{"ghost"}, Output: out, VarProfile: "work"})
		require.NoError(t, err)
	})

	assert.Contains(t, stdout, "Variable ghost not found")
}

func TestRunExportVars_ServiceError_Wrapped(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{listErr: errors.New("db is locked")}
	withFakeExportContainer(t, cmdSvc, varSvc, &fakeExportProfileRepo{})
	out := tempExportFile(t, "vars.json")

	stdout := captureStdout(t, func() {
		err := RunExportVars(RunExportVarArgs{Output: out, VarProfile: "work"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exporting variables")
		assert.Contains(t, err.Error(), "listing variables")
		assert.Contains(t, err.Error(), "db is locked")
	})

	assert.Empty(t, stdout, "nothing should be printed when the export fails")
	_, statErr := os.Stat(out)
	assert.True(t, os.IsNotExist(statErr), "no file should be written when the export fails")
}

func TestRunExportVars_ExportServiceUnavailable(t *testing.T) {
	withFailingExportService(t, errors.New("db down"))

	err := RunExportVars(RunExportVarArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting export service")
	assert.Contains(t, err.Error(), "db down")
}

func TestRunExportVars_ConsoleUnavailable(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	withFailingExportConsole(t, cmdSvc, varSvc, &fakeExportProfileRepo{})

	err := RunExportVars(RunExportVarArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

// --- RunExportAll ---

func TestRunExportAll_Success_CommandsAndVariablesWithTags(t *testing.T) {
	cmdSvc := &fakeCommandService{
		listCmds:       []models.Command{{Alias: "a"}},
		getWithTagsCmd: &models.Command{Alias: "a", Tags: []models.Tag{{Name: "t1"}}},
	}
	varSvc := &fakeVariableService{
		listVars:       []models.Variable{{Name: "v"}},
		getWithTagsVar: &models.Variable{Name: "v"},
	}
	withFakeExportContainer(t, cmdSvc, varSvc, &fakeExportProfileRepo{})
	out := tempExportFile(t, "all.json")

	stdout := captureStdout(t, func() {
		err := RunExportAll(RunExportAllArgs{Output: out, CmdProfile: "work", VarProfile: "work"})
		require.NoError(t, err)
	})

	assert.Contains(t, stdout, "Export Complete")
	assert.Contains(t, stdout, "a")
	assert.Contains(t, stdout, "v")

	data, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Contains(t, string(data), "t1", "tags should be recorded in the exported document")
}

func TestRunExportAll_FlattenPassedThrough(t *testing.T) {
	cmdSvc := &fakeCommandService{
		listCmds:       []models.Command{{Alias: "a", Template: "echo <var:name>"}},
		getWithTagsCmd: &models.Command{Alias: "a"},
		orNilCmd:       &models.Command{Alias: "a", Template: "echo <var:name>"},
	}
	varSvc := &fakeVariableService{
		listVars:       []models.Variable{{Name: "name", Value: "world"}},
		getWithTagsVar: &models.Variable{Name: "name"},
		orNilVar:       &models.Variable{Name: "name", Value: "world"},
	}
	withFakeExportContainer(t, cmdSvc, varSvc, &fakeExportProfileRepo{})
	out := tempExportFile(t, "all.json")

	captureStdout(t, func() {
		err := RunExportAll(RunExportAllArgs{Flatten: true, Output: out, CmdProfile: "work", VarProfile: "work"})
		require.NoError(t, err)
	})

	data, err := os.ReadFile(out)
	require.NoError(t, err)
	assert.Contains(t, string(data), "echo world")
}

func TestRunExportAll_NoProfile_UsesActiveProfilesFromRepo(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	profileRepo := &fakeExportProfileRepo{
		activeCmdProfile: &models.Profile{Name: "cmd-active"},
		activeVarProfile: &models.Profile{Name: "var-active"},
	}
	withFakeExportContainer(t, cmdSvc, varSvc, profileRepo)
	out := tempExportFile(t, "all.json")

	stdout := captureStdout(t, func() {
		err := RunExportAll(RunExportAllArgs{Output: out})
		require.NoError(t, err)
	})

	assert.Contains(t, stdout, "Saved to")
	_, statErr := os.Stat(out)
	assert.NoError(t, statErr)
}

func TestRunExportAll_CommandProfileResolutionError_Wrapped(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	profileRepo := &fakeExportProfileRepo{activeCmdErr: errors.New("no active profile")}
	withFakeExportContainer(t, cmdSvc, varSvc, profileRepo)
	out := tempExportFile(t, "all.json")

	err := RunExportAll(RunExportAllArgs{Output: out})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "exporting all")
	assert.Contains(t, err.Error(), "no active profile")
}

func TestRunExportAll_ServiceError_Wrapped(t *testing.T) {
	cmdSvc := &fakeCommandService{listErr: errors.New("db is locked")}
	varSvc := &fakeVariableService{}
	withFakeExportContainer(t, cmdSvc, varSvc, &fakeExportProfileRepo{})
	out := tempExportFile(t, "all.json")

	stdout := captureStdout(t, func() {
		err := RunExportAll(RunExportAllArgs{Output: out, CmdProfile: "work", VarProfile: "work"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "exporting all")
		assert.Contains(t, err.Error(), "listing commands")
		assert.Contains(t, err.Error(), "db is locked")
	})

	assert.Empty(t, stdout, "nothing should be printed when the export fails")
	_, statErr := os.Stat(out)
	assert.True(t, os.IsNotExist(statErr), "no file should be written when the export fails")
}

func TestRunExportAll_ExportServiceUnavailable(t *testing.T) {
	withFailingExportService(t, errors.New("db down"))

	err := RunExportAll(RunExportAllArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting export service")
	assert.Contains(t, err.Error(), "db down")
}

func TestRunExportAll_ConsoleUnavailable(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	withFailingExportConsole(t, cmdSvc, varSvc, &fakeExportProfileRepo{})

	err := RunExportAll(RunExportAllArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}
