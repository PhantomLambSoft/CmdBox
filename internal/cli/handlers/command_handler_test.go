package handlers

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

// fakeCommandService is a hand-rolled stub implementing services.CommandService that records
// the arguments it was called with and returns canned results, so handler tests can assert on
// argument mapping without touching a real database.
type fakeCommandService struct {
	createInput services.CreateCommandConfig
	createCmd   *models.Command
	createErr   error

	updateAlias   string
	updateProfile *string
	updateInput   services.UpdateCommandConfig
	updateCmd     *models.Command
	updateErr     error

	deleteCalled  bool
	deleteAlias   string
	deleteProfile *string
	deleteErr     error

	getAlias   string
	getProfile *string
	getCmd     *models.Command
	getErr     error

	getWithTagsAlias   string
	getWithTagsProfile *string
	getWithTagsCmd     *models.Command
	getWithTagsErr     error

	listOrder   string
	listTags    []string
	listLimit   *int
	listProfile *string
	listCmds    []models.Command
	listErr     error

	searchTerm    string
	searchFields  []string
	searchLimit   *int
	searchProfile *string
	searchCmds    []models.Command
	searchErr     error

	addTagsAlias   string
	addTagsNames   []string
	addTagsProfile *string
	addTagsResult  repository.TagAttachResult
	addTagsErr     error

	removeTagsAlias   string
	removeTagsNames   []string
	removeTagsProfile *string
	removeTagsResult  repository.TagDetachResult
	removeTagsErr     error
}

func (f *fakeCommandService) CreateCommand(input services.CreateCommandConfig) (*models.Command, error) {
	f.createInput = input
	return f.createCmd, f.createErr
}

func (f *fakeCommandService) UpdateCommand(alias string, profileName *string, input services.UpdateCommandConfig) (*models.Command, error) {
	f.updateAlias = alias
	f.updateProfile = profileName
	f.updateInput = input
	return f.updateCmd, f.updateErr
}

func (f *fakeCommandService) DeleteCommand(alias string, profileName *string) error {
	f.deleteCalled = true
	f.deleteAlias = alias
	f.deleteProfile = profileName
	return f.deleteErr
}

func (f *fakeCommandService) AddTags(alias string, tagNames []string, profileName *string) (repository.TagAttachResult, error) {
	f.addTagsAlias = alias
	f.addTagsNames = tagNames
	f.addTagsProfile = profileName
	return f.addTagsResult, f.addTagsErr
}

func (f *fakeCommandService) RemoveTags(alias string, tagNames []string, profileName *string) (repository.TagDetachResult, error) {
	f.removeTagsAlias = alias
	f.removeTagsNames = tagNames
	f.removeTagsProfile = profileName
	return f.removeTagsResult, f.removeTagsErr
}

func (f *fakeCommandService) GetCommand(alias string, profileName *string) (*models.Command, error) {
	f.getAlias = alias
	f.getProfile = profileName
	return f.getCmd, f.getErr
}

func (f *fakeCommandService) GetCommandWithTags(alias string, profileName *string) (*models.Command, error) {
	f.getWithTagsAlias = alias
	f.getWithTagsProfile = profileName
	return f.getWithTagsCmd, f.getWithTagsErr
}

func (f *fakeCommandService) GetCommandOrNil(alias string, profileName *string) (*models.Command, error) {
	return nil, nil
}

func (f *fakeCommandService) GetCommandByID(id uint, profileName *string) (*models.Command, error) {
	return nil, nil
}

func (f *fakeCommandService) ListCommands(orderBy string, tagNames []string, limit *int, profileName *string) ([]models.Command, error) {
	f.listOrder = orderBy
	f.listTags = tagNames
	f.listLimit = limit
	f.listProfile = profileName
	return f.listCmds, f.listErr
}

func (f *fakeCommandService) SearchCommands(query string, fields []string, limit *int, profileName *string) ([]models.Command, error) {
	f.searchTerm = query
	f.searchFields = fields
	f.searchLimit = limit
	f.searchProfile = profileName
	return f.searchCmds, f.searchErr
}

func (f *fakeCommandService) MoveCommand(alias string, targetProfileName, profileName *string) (*models.Command, error) {
	return nil, nil
}

func (f *fakeCommandService) CopyCommand(alias string, targetProfileName, newAlias, profileName *string) (*models.Command, error) {
	return nil, nil
}

// withFakeContainer swaps container.GetCommandService and container.GetConsole for the duration
// of the test, restoring the originals on cleanup. Pass a nil svc or console to simulate a
// container resolution failure for that dependency.
func withFakeContainer(t *testing.T, svc services.CommandService, svcErr error) {
	t.Helper()

	origSvc := container.GetCommandService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetCommandService = origSvc
		container.GetConsole = origConsole
	})

	container.GetCommandService = func() (services.CommandService, error) {
		return svc, svcErr
	}

	// Built lazily, at call time rather than here, so that tests which redirect os.Stdout via
	// captureStdout (after this setup runs) still have their output captured: Console wraps
	// whatever os.Stdout happens to be at construction time, not a live reference to it.
	container.GetConsole = func() (*ui.Console, error) {
		return ui.NewConsole(ui.NewPalette(settings.UIColors{}), false, "never", 0, 0, 0), nil
	}
}

// withFailingConsole makes container.GetConsole return an error, to exercise handler error paths
// that occur before any console output happens.
func withFailingConsole(t *testing.T, svc services.CommandService) {
	t.Helper()
	origSvc := container.GetCommandService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetCommandService = origSvc
		container.GetConsole = origConsole
	})
	container.GetCommandService = func() (services.CommandService, error) {
		return svc, nil
	}
	container.GetConsole = func() (*ui.Console, error) {
		return nil, errors.New("console boom")
	}
}

// captureStdout redirects os.Stdout for the duration of fn and returns everything written to it.
// ui.Console writes to a wrapped os.Stdout internally, so this is the only way to observe its
// output from outside the ui package.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	fn()

	require.NoError(t, w.Close())
	os.Stdout = orig
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	return string(out)
}

// --- RunAddCmd ---

func TestRunAddCmd_Success(t *testing.T) {
	fake := &fakeCommandService{createCmd: &models.Command{Alias: "build"}}
	withFakeContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunAddCmd(AddCmdArgs{
			Alias:          "build",
			Template:       "go build ./...",
			HasDescription: true,
			Description:    "builds the project",
			Cwd:            "/tmp",
			Shell:          "bash",
			Env:            []string{"FOO=bar"},
			HasTimeout:     true,
			Timeout:        30,
			Profile:        "work",
		})
		require.NoError(t, err)
	})

	assert.Contains(t, out, "build")
	assert.Equal(t, "build", fake.createInput.Alias)
	assert.Equal(t, "go build ./...", fake.createInput.Template)
	require.NotNil(t, fake.createInput.Description)
	assert.Equal(t, "builds the project", *fake.createInput.Description)
	require.NotNil(t, fake.createInput.Cwd)
	assert.Equal(t, "/tmp", *fake.createInput.Cwd)
	require.NotNil(t, fake.createInput.Shell)
	assert.Equal(t, "bash", *fake.createInput.Shell)
	require.NotNil(t, fake.createInput.Timeout)
	assert.Equal(t, 30, *fake.createInput.Timeout)
	require.NotNil(t, fake.createInput.ProfileName)
	assert.Equal(t, "work", *fake.createInput.ProfileName)
	assert.Equal(t, map[string]string{"FOO": "bar"}, fake.createInput.Env)
}

func TestRunAddCmd_OptionalFieldsOmitted(t *testing.T) {
	fake := &fakeCommandService{createCmd: &models.Command{Alias: "x"}}
	withFakeContainer(t, fake, nil)

	captureStdout(t, func() {
		err := RunAddCmd(AddCmdArgs{Alias: "x", Template: "echo hi"})
		require.NoError(t, err)
	})

	assert.Nil(t, fake.createInput.Description)
	assert.Nil(t, fake.createInput.Timeout)
	assert.Nil(t, fake.createInput.Cwd)
	assert.Nil(t, fake.createInput.Shell)
	assert.Nil(t, fake.createInput.ProfileName)
}

func TestRunAddCmd_InvalidEnv(t *testing.T) {
	fake := &fakeCommandService{}
	withFakeContainer(t, fake, nil)

	err := RunAddCmd(AddCmdArgs{Alias: "x", Template: "echo hi", Env: []string{"NOEQUALS"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --env value")
	assert.Zero(t, fake.createInput.Alias, "service should not be called when env parsing fails")
}

func TestRunAddCmd_ServiceError(t *testing.T) {
	fake := &fakeCommandService{createErr: errors.New("duplicate alias")}
	withFakeContainer(t, fake, nil)

	err := RunAddCmd(AddCmdArgs{Alias: "x", Template: "echo hi"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating command")
	assert.Contains(t, err.Error(), "duplicate alias")
}

func TestRunAddCmd_CommandServiceUnavailable(t *testing.T) {
	withFakeContainer(t, nil, errors.New("db down"))

	err := RunAddCmd(AddCmdArgs{Alias: "x", Template: "echo hi"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting command service")
}

func TestRunAddCmd_ConsoleUnavailable(t *testing.T) {
	fake := &fakeCommandService{createCmd: &models.Command{Alias: "x"}}
	withFailingConsole(t, fake)

	err := RunAddCmd(AddCmdArgs{Alias: "x", Template: "echo hi"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

// --- RunGetCmd ---

func TestRunGetCmd_Success(t *testing.T) {
	fake := &fakeCommandService{getWithTagsCmd: &models.Command{Alias: "build"}}
	withFakeContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunGetCmd("build", "work")
		require.NoError(t, err)
	})

	assert.Equal(t, "build", fake.getWithTagsAlias)
	require.NotNil(t, fake.getWithTagsProfile)
	assert.Equal(t, "work", *fake.getWithTagsProfile)
	assert.Contains(t, out, "build")
}

func TestRunGetCmd_NoProfile(t *testing.T) {
	fake := &fakeCommandService{getWithTagsCmd: &models.Command{Alias: "build"}}
	withFakeContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunGetCmd("build", ""))
	})

	assert.Nil(t, fake.getWithTagsProfile)
}

func TestRunGetCmd_ServiceError(t *testing.T) {
	fake := &fakeCommandService{getWithTagsErr: errors.New("not found")}
	withFakeContainer(t, fake, nil)

	err := RunGetCmd("missing", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting command")
	assert.Contains(t, err.Error(), "not found")
}

// --- RunUpdateCmd ---

func TestRunUpdateCmd_SetFields(t *testing.T) {
	fake := &fakeCommandService{updateCmd: &models.Command{Alias: "new"}}
	withFakeContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunUpdateCmd(UpdateCmdArgs{
			CurrentAlias:   "old",
			NewAlias:       "new",
			HasTemplate:    true,
			Template:       "echo updated",
			HasDescription: true,
			Description:    "new desc",
			HasCwd:         true,
			Cwd:            "/new",
			HasShell:       true,
			Shell:          "zsh",
			HasEnv:         true,
			Env:            []string{"A=1"},
			HasTimeout:     true,
			Timeout:        5,
			Profile:        "work",
		})
		require.NoError(t, err)
	})

	assert.Equal(t, "old", fake.updateAlias)
	require.NotNil(t, fake.updateProfile)
	assert.Equal(t, "work", *fake.updateProfile)

	in := fake.updateInput
	require.NotNil(t, in.NewAlias)
	assert.Equal(t, "new", *in.NewAlias)
	require.NotNil(t, in.Template)
	assert.Equal(t, "echo updated", *in.Template)
	require.NotNil(t, in.Description)
	assert.Equal(t, "new desc", *in.Description)
	require.NotNil(t, in.Cwd)
	assert.Equal(t, "/new", *in.Cwd)
	require.NotNil(t, in.Shell)
	assert.Equal(t, "zsh", *in.Shell)
	require.NotNil(t, in.Timeout)
	assert.Equal(t, 5, *in.Timeout)
	assert.Equal(t, map[string]string{"A": "1"}, in.Env)
	assert.Contains(t, out, "new")
}

func TestRunUpdateCmd_ClearFlags(t *testing.T) {
	fake := &fakeCommandService{updateCmd: &models.Command{Alias: "old"}}
	withFakeContainer(t, fake, nil)

	captureStdout(t, func() {
		err := RunUpdateCmd(UpdateCmdArgs{
			CurrentAlias: "old",
			ClearCwd:     true,
			ClearShell:   true,
			ClearEnv:     true,
			ClearTimeout: true,
		})
		require.NoError(t, err)
	})

	in := fake.updateInput
	assert.True(t, in.ClearCwd)
	assert.True(t, in.ClearShell)
	assert.True(t, in.ClearEnv)
	assert.True(t, in.ClearTimeout)
	assert.Nil(t, in.NewAlias)
	assert.Nil(t, in.Template)
	assert.Nil(t, in.Description)
	assert.Nil(t, in.Cwd)
	assert.Nil(t, in.Shell)
	assert.Nil(t, in.Timeout)
}

func TestRunUpdateCmd_InvalidEnv(t *testing.T) {
	fake := &fakeCommandService{}
	withFakeContainer(t, fake, nil)

	err := RunUpdateCmd(UpdateCmdArgs{CurrentAlias: "old", HasEnv: true, Env: []string{"bad"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --env value")
	assert.Empty(t, fake.updateAlias, "service should not be called when env parsing fails")
}

func TestRunUpdateCmd_ServiceError(t *testing.T) {
	fake := &fakeCommandService{updateErr: errors.New("alias taken")}
	withFakeContainer(t, fake, nil)

	err := RunUpdateCmd(UpdateCmdArgs{CurrentAlias: "old"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "updating command")
	assert.Contains(t, err.Error(), "alias taken")
}

// --- RunListCmd ---

func TestRunListCmd_WithLimitAndTags(t *testing.T) {
	fake := &fakeCommandService{listCmds: []models.Command{{Alias: "a"}, {Alias: "b"}}}
	withFakeContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunListCmd(ListCmdArgs{
			ListParams: ListParams{Order: "alias", HasLimit: true, Limit: 10, Profile: "work"},
			Tags:       []string{"db", "net"},
		})
		require.NoError(t, err)
	})

	assert.Equal(t, "alias", fake.listOrder)
	assert.Equal(t, []string{"db", "net"}, fake.listTags)
	require.NotNil(t, fake.listLimit)
	assert.Equal(t, 10, *fake.listLimit)
	require.NotNil(t, fake.listProfile)
	assert.Equal(t, "work", *fake.listProfile)
	assert.Contains(t, out, "a")
	assert.Contains(t, out, "b")
}

func TestRunListCmd_NoLimit(t *testing.T) {
	fake := &fakeCommandService{}
	withFakeContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunListCmd(ListCmdArgs{}))
	})

	assert.Nil(t, fake.listLimit)
}

func TestRunListCmd_ServiceError(t *testing.T) {
	fake := &fakeCommandService{listErr: errors.New("db error")}
	withFakeContainer(t, fake, nil)

	err := RunListCmd(ListCmdArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing commands")
}

// --- RunSearchCmd ---

func TestRunSearchCmd_Success(t *testing.T) {
	fake := &fakeCommandService{searchCmds: []models.Command{{Alias: "match"}}}
	withFakeContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunSearchCmd(SearchCmdArgs{
			SearchParams: SearchParams{Term: "bui", SearchIn: []string{"alias"}, HasLimit: true, Limit: 5, Profile: "work"},
		})
		require.NoError(t, err)
	})

	assert.Equal(t, "bui", fake.searchTerm)
	assert.Equal(t, []string{"alias"}, fake.searchFields)
	require.NotNil(t, fake.searchLimit)
	assert.Equal(t, 5, *fake.searchLimit)
	require.NotNil(t, fake.searchProfile)
	assert.Equal(t, "work", *fake.searchProfile)
	assert.Contains(t, out, "match")
}

func TestRunSearchCmd_ServiceError(t *testing.T) {
	fake := &fakeCommandService{searchErr: errors.New("boom")}
	withFakeContainer(t, fake, nil)

	err := RunSearchCmd(SearchCmdArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "searching commands")
}

// --- RunDeleteCmd ---

func TestRunDeleteCmd_Success(t *testing.T) {
	fake := &fakeCommandService{getCmd: &models.Command{Alias: "gone"}}
	withFakeContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunDeleteCmd(DeleteCmdArgs{Alias: "gone", Profile: "work"})
		require.NoError(t, err)
	})

	assert.Equal(t, "gone", fake.getAlias)
	require.NotNil(t, fake.getProfile)
	assert.Equal(t, "work", *fake.getProfile)
	assert.True(t, fake.deleteCalled)
	assert.Equal(t, "gone", fake.deleteAlias)
	assert.Contains(t, out, "gone")
}

func TestRunDeleteCmd_GetFails_DoesNotDelete(t *testing.T) {
	fake := &fakeCommandService{getErr: errors.New("not found")}
	withFakeContainer(t, fake, nil)

	err := RunDeleteCmd(DeleteCmdArgs{Alias: "missing"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting command")
	assert.False(t, fake.deleteCalled, "delete should not be attempted when the lookup fails")
}

func TestRunDeleteCmd_DeleteFails(t *testing.T) {
	fake := &fakeCommandService{getCmd: &models.Command{Alias: "gone"}, deleteErr: errors.New("locked")}
	withFakeContainer(t, fake, nil)

	err := RunDeleteCmd(DeleteCmdArgs{Alias: "gone"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleting command")
	assert.Contains(t, err.Error(), "locked")
}

// --- RunTagCmd ---

func TestRunTagCmd_AddedAndExisting(t *testing.T) {
	fake := &fakeCommandService{addTagsResult: repository.TagAttachResult{
		Added:    []string{"new1"},
		Existing: []string{"old1"},
	}}
	withFakeContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunTagCmd(TagCmdArgs{TagParams: TagParams{TagNames: []string{"new1", "old1"}, Profile: "work"}, Alias: "build"})
		require.NoError(t, err)
	})

	assert.Equal(t, "build", fake.addTagsAlias)
	assert.Equal(t, []string{"new1", "old1"}, fake.addTagsNames)
	require.NotNil(t, fake.addTagsProfile)
	assert.Equal(t, "work", *fake.addTagsProfile)
	assert.Contains(t, out, "new1")
	assert.Contains(t, out, "old1")
}

func TestRunTagCmd_NoOutputWhenResultEmpty(t *testing.T) {
	fake := &fakeCommandService{}
	withFakeContainer(t, fake, nil)

	out := captureStdout(t, func() {
		require.NoError(t, RunTagCmd(TagCmdArgs{Alias: "build"}))
	})

	assert.Empty(t, out)
}

func TestRunTagCmd_ServiceError(t *testing.T) {
	fake := &fakeCommandService{addTagsErr: errors.New("no such command")}
	withFakeContainer(t, fake, nil)

	err := RunTagCmd(TagCmdArgs{Alias: "build"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "adding tags")
}

// --- RunUntagCmd ---

func TestRunUntagCmd_RemovedAndNotAttached(t *testing.T) {
	fake := &fakeCommandService{removeTagsResult: repository.TagDetachResult{
		Removed:     []string{"gone1"},
		NotAttached: []string{"missing1"},
	}}
	withFakeContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunUntagCmd(TagCmdArgs{TagParams: TagParams{TagNames: []string{"gone1", "missing1"}, Profile: "work"}, Alias: "build"})
		require.NoError(t, err)
	})

	assert.Equal(t, "build", fake.removeTagsAlias)
	assert.Equal(t, []string{"gone1", "missing1"}, fake.removeTagsNames)
	require.NotNil(t, fake.removeTagsProfile)
	assert.Equal(t, "work", *fake.removeTagsProfile)
	assert.Contains(t, out, "gone1")
	assert.Contains(t, out, "missing1")
}

func TestRunUntagCmd_ServiceError(t *testing.T) {
	fake := &fakeCommandService{removeTagsErr: errors.New("no such command")}
	withFakeContainer(t, fake, nil)

	err := RunUntagCmd(TagCmdArgs{Alias: "build"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "removing tags")
}

// --- helpers ---

func TestParseEnvPairs(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		m, err := parseEnvPairs(nil)
		require.NoError(t, err)
		assert.Nil(t, m)
	})

	t.Run("valid", func(t *testing.T) {
		m, err := parseEnvPairs([]string{"A=1", " B = 2"})
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"A": "1", "B": " 2"}, m)
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := parseEnvPairs([]string{"NOEQUALS"})
		require.Error(t, err)
		assert.Contains(t, err.Error(), `"NOEQUALS"`)
	})

	t.Run("value with equals sign", func(t *testing.T) {
		m, err := parseEnvPairs([]string{"A=b=c"})
		require.NoError(t, err)
		assert.Equal(t, map[string]string{"A": "b=c"}, m)
	})
}

func TestStringOrNil(t *testing.T) {
	assert.Nil(t, stringOrNil(""))
	require.NotNil(t, stringOrNil("x"))
	assert.Equal(t, "x", *stringOrNil("x"))
}
