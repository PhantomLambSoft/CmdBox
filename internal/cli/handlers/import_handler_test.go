package handlers

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

// withFakeImportContainer swaps container.GetImportService and container.GetConsole for the
// duration of the test, restoring the originals on cleanup. The import service it installs is a
// real *services.ImportService wired to the given fakes, so handler tests exercise the actual
// argument translation (RunImportFile -> services.ImportConfig) without touching a real database.
func withFakeImportContainer(t *testing.T, cmdSvc services.CommandService, varSvc services.VariableService, tagSvc services.TagService, profileRepo *fakeExportProfileRepo) {
	t.Helper()

	origSvc := container.GetImportService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetImportService = origSvc
		container.GetConsole = origConsole
	})

	container.GetImportService = func() (*services.ImportService, error) {
		return services.NewImportService(cmdSvc, varSvc, tagSvc, profileRepo), nil
	}
	container.GetConsole = func() (*ui.Console, error) {
		return ui.NewConsole(ui.NewPalette(settings.UIColors{}), false, "never", 0, 0, 0), nil
	}
}

// withFailingImportService makes container.GetImportService return an error, simulating a
// container resolution failure (e.g. the database being unavailable) before any import logic
// runs.
func withFailingImportService(t *testing.T, svcErr error) {
	t.Helper()

	orig := container.GetImportService
	t.Cleanup(func() {
		container.GetImportService = orig
	})
	container.GetImportService = func() (*services.ImportService, error) {
		return nil, svcErr
	}
}

// withFailingImportConsole makes container.GetConsole return an error, to exercise handler error
// paths that occur after the import itself has already succeeded.
func withFailingImportConsole(t *testing.T, cmdSvc services.CommandService, varSvc services.VariableService, tagSvc services.TagService, profileRepo *fakeExportProfileRepo) {
	t.Helper()

	origSvc := container.GetImportService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetImportService = origSvc
		container.GetConsole = origConsole
	})

	container.GetImportService = func() (*services.ImportService, error) {
		return services.NewImportService(cmdSvc, varSvc, tagSvc, profileRepo), nil
	}
	container.GetConsole = func() (*ui.Console, error) {
		return nil, errors.New("console boom")
	}
}

// writeImportFile writes doc as JSON to a fresh file under t.TempDir(), so the import never reads
// anything outside the test's own sandbox.
func writeImportFile(t *testing.T, name string, doc map[string]any) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	data, err := json.Marshal(doc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0o600))
	return path
}

func baseImportDoc() map[string]any {
	return map[string]any{
		"version":          "1",
		"type":             "cmdbox-transfer",
		"command_profile":  "work",
		"variable_profile": "work",
		"commands":         []any{},
		"variables":        []any{},
	}
}

// --- RunImportFile ---

func TestRunImportFile_Success_CreatesCommandsAndVariables(t *testing.T) {
	doc := baseImportDoc()
	doc["commands"] = []any{map[string]any{"alias": "build", "template": "go build ./..."}}
	doc["variables"] = []any{map[string]any{"name": "name", "value": "world"}}
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, &fakeExportProfileRepo{})

	stdout := captureStdout(t, func() {
		err := RunImportFile(RunImportArgs{Path: path, Profile: "work"})
		require.NoError(t, err)
	})

	assert.Equal(t, "build", cmdSvc.createInput.Alias)
	assert.Equal(t, "name", varSvc.createInput.Name)
	assert.Contains(t, stdout, "Import complete")
	assert.Contains(t, stdout, "build")
	assert.Contains(t, stdout, "name")
}

func TestRunImportFile_Preview_DoesNotCreateAnything(t *testing.T) {
	doc := baseImportDoc()
	doc["commands"] = []any{map[string]any{"alias": "build", "template": "go build ./..."}}
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, &fakeExportProfileRepo{})

	stdout := captureStdout(t, func() {
		err := RunImportFile(RunImportArgs{Path: path, Profile: "work", Preview: true})
		require.NoError(t, err)
	})

	assert.Empty(t, cmdSvc.createInput.Alias, "preview should not create the command")
	assert.Contains(t, stdout, "build")
}

func TestRunImportFile_Overwrite_UpdatesExistingCommand(t *testing.T) {
	doc := baseImportDoc()
	doc["commands"] = []any{map[string]any{"alias": "build", "template": "go build ./... -v"}}
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{
		orNilCmd:       &models.Command{Alias: "build"},
		getWithTagsCmd: &models.Command{Alias: "build"},
	}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, &fakeExportProfileRepo{})

	stdout := captureStdout(t, func() {
		err := RunImportFile(RunImportArgs{Path: path, Profile: "work", Overwrite: true})
		require.NoError(t, err)
	})

	assert.Equal(t, "build", cmdSvc.updateAlias)
	assert.Contains(t, stdout, "Commands overwritten")
}

func TestRunImportFile_SkipsExistingCommandWithoutOverwrite(t *testing.T) {
	doc := baseImportDoc()
	doc["commands"] = []any{map[string]any{"alias": "build", "template": "go build ./..."}}
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{orNilCmd: &models.Command{Alias: "build"}}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, &fakeExportProfileRepo{})

	stdout := captureStdout(t, func() {
		err := RunImportFile(RunImportArgs{Path: path, Profile: "work"})
		require.NoError(t, err)
	})

	assert.Empty(t, cmdSvc.createInput.Alias)
	assert.Empty(t, cmdSvc.updateAlias)
	assert.Contains(t, stdout, "Commands skipped")
	assert.Contains(t, stdout, "build")
}

func TestRunImportFile_NoProfile_UsesActiveProfileFromRepo(t *testing.T) {
	doc := baseImportDoc()
	doc["commands"] = []any{map[string]any{"alias": "build", "template": "go build ./..."}}
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	profileRepo := &fakeExportProfileRepo{activeCmdProfile: &models.Profile{Name: "cmd-active"}}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, profileRepo)

	stdout := captureStdout(t, func() {
		err := RunImportFile(RunImportArgs{Path: path})
		require.NoError(t, err)
	})

	require.NotNil(t, cmdSvc.createInput.ProfileName)
	assert.Equal(t, "cmd-active", *cmdSvc.createInput.ProfileName)
	assert.Contains(t, stdout, "cmd-active")
}

func TestRunImportFile_CreatesMissingTags(t *testing.T) {
	doc := baseImportDoc()
	doc["commands"] = []any{map[string]any{"alias": "build", "template": "go build ./...", "tags": []any{"ci"}}}
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, &fakeExportProfileRepo{})

	captureStdout(t, func() {
		err := RunImportFile(RunImportArgs{Path: path, Profile: "work"})
		require.NoError(t, err)
	})

	assert.Equal(t, []string{"ci"}, tagSvc.createCalls)
}

func TestRunImportFile_DoesNotRecreateExistingTag(t *testing.T) {
	doc := baseImportDoc()
	doc["commands"] = []any{map[string]any{"alias": "build", "template": "go build ./...", "tags": []any{"ci"}}}
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{orNilTag: &models.Tag{Name: "ci"}}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, &fakeExportProfileRepo{})

	captureStdout(t, func() {
		err := RunImportFile(RunImportArgs{Path: path, Profile: "work"})
		require.NoError(t, err)
	})

	assert.Empty(t, tagSvc.createCalls, "tag already exists, should not be recreated")
}

func TestRunImportFile_FileNotFound_Wrapped(t *testing.T) {
	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, &fakeExportProfileRepo{})

	missing := filepath.Join(t.TempDir(), "does-not-exist.json")
	err := RunImportFile(RunImportArgs{Path: missing, Profile: "work"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to import file")
}

func TestRunImportFile_UnsupportedVersion_Wrapped(t *testing.T) {
	doc := baseImportDoc()
	doc["version"] = "999"
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, &fakeExportProfileRepo{})

	err := RunImportFile(RunImportArgs{Path: path, Profile: "work"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to import file")
	assert.ErrorIs(t, err, services.ErrUnsupportedVersion)
}

func TestRunImportFile_ImportCycle_Wrapped(t *testing.T) {
	doc := baseImportDoc()
	// Build a genuine cycle: command "a" references variable "v", variable "v" references command "a".
	doc["commands"] = []any{map[string]any{"alias": "a", "template": "<var:v>"}}
	doc["variables"] = []any{map[string]any{"name": "v", "value": "<cmd:a>"}}
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, &fakeExportProfileRepo{})

	err := RunImportFile(RunImportArgs{Path: path, Profile: "work"})
	require.Error(t, err)
	assert.ErrorIs(t, err, services.ErrImportCycle)
}

func TestRunImportFile_ProfileResolutionError_Wrapped(t *testing.T) {
	doc := baseImportDoc()
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	profileRepo := &fakeExportProfileRepo{activeCmdErr: errors.New("no active profile")}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, profileRepo)

	err := RunImportFile(RunImportArgs{Path: path})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to import file")
	assert.Contains(t, err.Error(), "no active profile")
}

func TestRunImportFile_ServiceError_Wrapped(t *testing.T) {
	doc := baseImportDoc()
	doc["commands"] = []any{map[string]any{"alias": "build", "template": "go build ./..."}}
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{createErr: errors.New("db is locked")}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	withFakeImportContainer(t, cmdSvc, varSvc, tagSvc, &fakeExportProfileRepo{})

	err := RunImportFile(RunImportArgs{Path: path, Profile: "work"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to import file")
	assert.Contains(t, err.Error(), "db is locked")
}

func TestRunImportFile_ImportServiceUnavailable(t *testing.T) {
	withFailingImportService(t, errors.New("db down"))

	err := RunImportFile(RunImportArgs{Path: "unused"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get import service")
	assert.Contains(t, err.Error(), "db down")
}

func TestRunImportFile_ConsoleUnavailable(t *testing.T) {
	doc := baseImportDoc()
	path := writeImportFile(t, "import.json", doc)

	cmdSvc := &fakeCommandService{}
	varSvc := &fakeVariableService{}
	tagSvc := &fakeTagService{}
	withFailingImportConsole(t, cmdSvc, varSvc, tagSvc, &fakeExportProfileRepo{})

	err := RunImportFile(RunImportArgs{Path: path, Profile: "work"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get console")
}
