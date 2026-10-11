package handlers

import (
	"errors"
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

// fakeVariableService is a hand-rolled stub implementing services.VariableService that records
// the arguments it was called with and returns canned results, so handler tests can assert on
// argument mapping without touching a real database.
type fakeVariableService struct {
	createInput services.CreateVariableConfig
	createVar   *models.Variable
	createErr   error

	updateName    string
	updateProfile *string
	updateInput   services.UpdateVariableConfig
	updateVar     *models.Variable
	updateErr     error

	deleteCalled  bool
	deleteName    string
	deleteProfile *string
	deleteErr     error

	getName    string
	getProfile *string
	getVar     *models.Variable
	getErr     error

	getWithTagsName    string
	getWithTagsProfile *string
	getWithTagsVar     *models.Variable
	getWithTagsErr     error

	listOrder   string
	listTags    []string
	listLimit   *int
	listProfile *string
	listVars    []models.Variable
	listErr     error

	searchTerm    string
	searchFields  []string
	searchLimit   *int
	searchProfile *string
	searchVars    []models.Variable
	searchErr     error

	addTagsName    string
	addTagsNames   []string
	addTagsProfile *string
	addTagsResult  repository.TagAttachResult
	addTagsErr     error

	removeTagsName    string
	removeTagsNames   []string
	removeTagsProfile *string
	removeTagsResult  repository.TagDetachResult
	removeTagsErr     error
}

func (f *fakeVariableService) CreateVariable(input services.CreateVariableConfig) (*models.Variable, error) {
	f.createInput = input
	return f.createVar, f.createErr
}

func (f *fakeVariableService) UpdateVariable(name string, profileName *string, input services.UpdateVariableConfig) (*models.Variable, error) {
	f.updateName = name
	f.updateProfile = profileName
	f.updateInput = input
	return f.updateVar, f.updateErr
}

func (f *fakeVariableService) DeleteVariable(name string, profileName *string) error {
	f.deleteCalled = true
	f.deleteName = name
	f.deleteProfile = profileName
	return f.deleteErr
}

func (f *fakeVariableService) AddTags(name string, tagNames []string, profileName *string) (repository.TagAttachResult, error) {
	f.addTagsName = name
	f.addTagsNames = tagNames
	f.addTagsProfile = profileName
	return f.addTagsResult, f.addTagsErr
}

func (f *fakeVariableService) RemoveTags(name string, tagNames []string, profileName *string) (repository.TagDetachResult, error) {
	f.removeTagsName = name
	f.removeTagsNames = tagNames
	f.removeTagsProfile = profileName
	return f.removeTagsResult, f.removeTagsErr
}

func (f *fakeVariableService) GetVariable(name string, profileName *string) (*models.Variable, error) {
	f.getName = name
	f.getProfile = profileName
	return f.getVar, f.getErr
}

func (f *fakeVariableService) GetVariableWithTags(name string, profileName *string) (*models.Variable, error) {
	f.getWithTagsName = name
	f.getWithTagsProfile = profileName
	return f.getWithTagsVar, f.getWithTagsErr
}

func (f *fakeVariableService) GetVariableOrNil(name string, profileName *string) (*models.Variable, error) {
	return nil, nil
}

func (f *fakeVariableService) GetVariableByID(id uint, profileName *string) (*models.Variable, error) {
	return nil, nil
}

func (f *fakeVariableService) ListVariables(orderBy string, tagNames []string, limit *int, profileName *string) ([]models.Variable, error) {
	f.listOrder = orderBy
	f.listTags = tagNames
	f.listLimit = limit
	f.listProfile = profileName
	return f.listVars, f.listErr
}

func (f *fakeVariableService) SearchVariables(query string, fields []string, limit *int, profileName *string) ([]models.Variable, error) {
	f.searchTerm = query
	f.searchFields = fields
	f.searchLimit = limit
	f.searchProfile = profileName
	return f.searchVars, f.searchErr
}

func (f *fakeVariableService) MoveVariable(name string, targetProfileName, profileName *string) (*models.Variable, error) {
	return nil, nil
}

func (f *fakeVariableService) CopyVariable(name string, targetProfileName, newName, profileName *string) (*models.Variable, error) {
	return nil, nil
}

// withFakeVariableContainer swaps container.GetVariableService and container.GetConsole for the
// duration of the test, restoring the originals on cleanup. The console is built lazily (at call
// time) so that tests which redirect os.Stdout via captureStdout after this setup runs still have
// their output captured.
func withFakeVariableContainer(t *testing.T, svc services.VariableService, svcErr error) {
	t.Helper()

	origSvc := container.GetVariableService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetVariableService = origSvc
		container.GetConsole = origConsole
	})

	container.GetVariableService = func() (services.VariableService, error) {
		return svc, svcErr
	}
	container.GetConsole = func() (*ui.Console, error) {
		return ui.NewConsole(ui.NewPalette(settings.UIColors{}), false, "never", 0, 0, 0), nil
	}
}

// withFailingVariableConsole makes container.GetConsole return an error, to exercise handler
// error paths that occur before any console output happens.
func withFailingVariableConsole(t *testing.T, svc services.VariableService) {
	t.Helper()
	origSvc := container.GetVariableService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetVariableService = origSvc
		container.GetConsole = origConsole
	})
	container.GetVariableService = func() (services.VariableService, error) {
		return svc, nil
	}
	container.GetConsole = func() (*ui.Console, error) {
		return nil, errors.New("console boom")
	}
}

// --- RunAddVar ---

func TestRunAddVar_Success(t *testing.T) {
	fake := &fakeVariableService{createVar: &models.Variable{Name: "host"}}
	withFakeVariableContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunAddVar(AddVarArgs{Name: "host", Value: "example.com", HasValue: true, Profile: "work"})
		require.NoError(t, err)
	})

	assert.Contains(t, out, "host")
	assert.Equal(t, "host", fake.createInput.Name)
	assert.Equal(t, "example.com", fake.createInput.Value)
	require.NotNil(t, fake.createInput.ProfileName)
	assert.Equal(t, "work", *fake.createInput.ProfileName)
}

func TestRunAddVar_NoProfile(t *testing.T) {
	fake := &fakeVariableService{createVar: &models.Variable{Name: "host"}}
	withFakeVariableContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunAddVar(AddVarArgs{Name: "host", Value: "x", HasValue: true}))
	})

	assert.Nil(t, fake.createInput.ProfileName)
}

func TestRunAddVar_ServiceError(t *testing.T) {
	fake := &fakeVariableService{createErr: errors.New("duplicate name")}
	withFakeVariableContainer(t, fake, nil)

	err := RunAddVar(AddVarArgs{Name: "host", Value: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating variable")
	assert.Contains(t, err.Error(), "duplicate name")
}

func TestRunAddVar_VariableServiceUnavailable(t *testing.T) {
	withFakeVariableContainer(t, nil, errors.New("db down"))

	err := RunAddVar(AddVarArgs{Name: "host", Value: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

func TestRunAddVar_ConsoleUnavailable(t *testing.T) {
	fake := &fakeVariableService{createVar: &models.Variable{Name: "host"}}
	withFailingVariableConsole(t, fake)

	err := RunAddVar(AddVarArgs{Name: "host", Value: "x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

// --- RunGetVar ---

func TestRunGetVar_Success(t *testing.T) {
	fake := &fakeVariableService{getWithTagsVar: &models.Variable{Name: "host"}}
	withFakeVariableContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunGetVar("host", "work")
		require.NoError(t, err)
	})

	assert.Equal(t, "host", fake.getWithTagsName)
	require.NotNil(t, fake.getWithTagsProfile)
	assert.Equal(t, "work", *fake.getWithTagsProfile)
	assert.Contains(t, out, "host")
}

func TestRunGetVar_NoProfile(t *testing.T) {
	fake := &fakeVariableService{getWithTagsVar: &models.Variable{Name: "host"}}
	withFakeVariableContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunGetVar("host", ""))
	})

	assert.Nil(t, fake.getWithTagsProfile)
}

func TestRunGetVar_ServiceError(t *testing.T) {
	fake := &fakeVariableService{getWithTagsErr: errors.New("not found")}
	withFakeVariableContainer(t, fake, nil)

	err := RunGetVar("missing", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting variable")
	assert.Contains(t, err.Error(), "not found")
}

// --- RunUpdateVar ---

func TestRunUpdateVar_RenameAndValue(t *testing.T) {
	fake := &fakeVariableService{updateVar: &models.Variable{Name: "newname"}}
	withFakeVariableContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunUpdateVar(UpdateVarArgs{
			CurrentName: "oldname",
			NewName:     "newname",
			NewValue:    "newvalue",
			HasValue:    true,
			Profile:     "work",
		})
		require.NoError(t, err)
	})

	assert.Equal(t, "oldname", fake.updateName)
	require.NotNil(t, fake.updateProfile)
	assert.Equal(t, "work", *fake.updateProfile)

	in := fake.updateInput
	require.NotNil(t, in.NewName)
	assert.Equal(t, "newname", *in.NewName)
	require.NotNil(t, in.Value)
	assert.Equal(t, "newvalue", *in.Value)
	assert.Contains(t, out, "newname")
}

func TestRunUpdateVar_RenameOnly(t *testing.T) {
	fake := &fakeVariableService{updateVar: &models.Variable{Name: "newname"}}
	withFakeVariableContainer(t, fake, nil)

	captureStdout(t, func() {
		err := RunUpdateVar(UpdateVarArgs{CurrentName: "oldname", NewName: "newname"})
		require.NoError(t, err)
	})

	in := fake.updateInput
	require.NotNil(t, in.NewName)
	assert.Equal(t, "newname", *in.NewName)
	assert.Nil(t, in.Value)
}

func TestRunUpdateVar_ValueOnly(t *testing.T) {
	fake := &fakeVariableService{updateVar: &models.Variable{Name: "oldname"}}
	withFakeVariableContainer(t, fake, nil)

	captureStdout(t, func() {
		err := RunUpdateVar(UpdateVarArgs{CurrentName: "oldname", NewValue: "newvalue", HasValue: true})
		require.NoError(t, err)
	})

	in := fake.updateInput
	assert.Nil(t, in.NewName)
	require.NotNil(t, in.Value)
	assert.Equal(t, "newvalue", *in.Value)
}

func TestRunUpdateVar_NoChanges(t *testing.T) {
	fake := &fakeVariableService{updateVar: &models.Variable{Name: "oldname"}}
	withFakeVariableContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunUpdateVar(UpdateVarArgs{CurrentName: "oldname"}))
	})

	assert.Nil(t, fake.updateInput.NewName)
	assert.Nil(t, fake.updateInput.Value)
}

func TestRunUpdateVar_ServiceError(t *testing.T) {
	fake := &fakeVariableService{updateErr: errors.New("name taken")}
	withFakeVariableContainer(t, fake, nil)

	err := RunUpdateVar(UpdateVarArgs{CurrentName: "oldname"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "updating variable")
	assert.Contains(t, err.Error(), "name taken")
}

// --- RunListVar ---

func TestRunListVar_WithLimitAndTags(t *testing.T) {
	fake := &fakeVariableService{listVars: []models.Variable{{Name: "a"}, {Name: "b"}}}
	withFakeVariableContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunListVar(ListVarArgs{
			ListParams: ListParams{Order: "name", HasLimit: true, Limit: 10, Profile: "work"},
			Tags:       []string{"infra", "net"},
		})
		require.NoError(t, err)
	})

	assert.Equal(t, "name", fake.listOrder)
	assert.Equal(t, []string{"infra", "net"}, fake.listTags)
	require.NotNil(t, fake.listLimit)
	assert.Equal(t, 10, *fake.listLimit)
	require.NotNil(t, fake.listProfile)
	assert.Equal(t, "work", *fake.listProfile)
	assert.Contains(t, out, "a")
	assert.Contains(t, out, "b")
}

func TestRunListVar_NoLimit(t *testing.T) {
	fake := &fakeVariableService{}
	withFakeVariableContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunListVar(ListVarArgs{}))
	})

	assert.Nil(t, fake.listLimit)
}

func TestRunListVar_ServiceError(t *testing.T) {
	fake := &fakeVariableService{listErr: errors.New("db error")}
	withFakeVariableContainer(t, fake, nil)

	err := RunListVar(ListVarArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing variables")
}

// --- RunSearchVar ---

func TestRunSearchVar_Success(t *testing.T) {
	fake := &fakeVariableService{searchVars: []models.Variable{{Name: "match"}}}
	withFakeVariableContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunSearchVar(SearchParams{Term: "mat", SearchIn: []string{"name"}, HasLimit: true, Limit: 5, Profile: "work"})
		require.NoError(t, err)
	})

	assert.Equal(t, "mat", fake.searchTerm)
	assert.Equal(t, []string{"name"}, fake.searchFields)
	require.NotNil(t, fake.searchLimit)
	assert.Equal(t, 5, *fake.searchLimit)
	require.NotNil(t, fake.searchProfile)
	assert.Equal(t, "work", *fake.searchProfile)
	assert.Contains(t, out, "match")
}

func TestRunSearchVar_ServiceError(t *testing.T) {
	fake := &fakeVariableService{searchErr: errors.New("boom")}
	withFakeVariableContainer(t, fake, nil)

	err := RunSearchVar(SearchParams{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "searching variables")
}

// --- RunDeleteVar ---

func TestRunDeleteVar_Success(t *testing.T) {
	fake := &fakeVariableService{getVar: &models.Variable{Name: "gone"}}
	withFakeVariableContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunDeleteVar(DeleteVarArgs{Name: "gone", Profile: "work"})
		require.NoError(t, err)
	})

	assert.Equal(t, "gone", fake.getName)
	require.NotNil(t, fake.getProfile)
	assert.Equal(t, "work", *fake.getProfile)
	assert.True(t, fake.deleteCalled)
	assert.Equal(t, "gone", fake.deleteName)
	assert.Contains(t, out, "gone")
}

func TestRunDeleteVar_GetFails_DoesNotDelete(t *testing.T) {
	fake := &fakeVariableService{getErr: errors.New("not found")}
	withFakeVariableContainer(t, fake, nil)

	err := RunDeleteVar(DeleteVarArgs{Name: "missing"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting variable")
	assert.False(t, fake.deleteCalled, "delete should not be attempted when the lookup fails")
}

func TestRunDeleteVar_DeleteFails(t *testing.T) {
	fake := &fakeVariableService{getVar: &models.Variable{Name: "gone"}, deleteErr: errors.New("locked")}
	withFakeVariableContainer(t, fake, nil)

	err := RunDeleteVar(DeleteVarArgs{Name: "gone"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleting variable")
	assert.Contains(t, err.Error(), "locked")
}

// --- RunTagVar ---

func TestRunTagVar_AddedAndExisting(t *testing.T) {
	fake := &fakeVariableService{addTagsResult: repository.TagAttachResult{
		Added:    []string{"new1"},
		Existing: []string{"old1"},
	}}
	withFakeVariableContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunTagVar(TagVarArgs{TagParams: TagParams{TagNames: []string{"new1", "old1"}, Profile: "work"}, Name: "host"})
		require.NoError(t, err)
	})

	assert.Equal(t, "host", fake.addTagsName)
	assert.Equal(t, []string{"new1", "old1"}, fake.addTagsNames)
	require.NotNil(t, fake.addTagsProfile)
	assert.Equal(t, "work", *fake.addTagsProfile)
	assert.Contains(t, out, "new1")
	assert.Contains(t, out, "old1")
	assert.Contains(t, out, "Already tagged with: old1")
}

func TestRunTagVar_NoOutputWhenResultEmpty(t *testing.T) {
	fake := &fakeVariableService{}
	withFakeVariableContainer(t, fake, nil)

	out := captureStdout(t, func() {
		require.NoError(t, RunTagVar(TagVarArgs{Name: "host"}))
	})

	assert.Empty(t, out)
}

func TestRunTagVar_ServiceError(t *testing.T) {
	fake := &fakeVariableService{addTagsErr: errors.New("no such variable")}
	withFakeVariableContainer(t, fake, nil)

	err := RunTagVar(TagVarArgs{Name: "host"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "adding tags")
}

// --- RunUntagVar ---

func TestRunUntagVar_RemovedAndNotAttached(t *testing.T) {
	fake := &fakeVariableService{removeTagsResult: repository.TagDetachResult{
		Removed:     []string{"gone1"},
		NotAttached: []string{"missing1"},
	}}
	withFakeVariableContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunUntagVar(TagVarArgs{TagParams: TagParams{TagNames: []string{"gone1", "missing1"}, Profile: "work"}, Name: "host"})
		require.NoError(t, err)
	})

	assert.Equal(t, "host", fake.removeTagsName)
	assert.Equal(t, []string{"gone1", "missing1"}, fake.removeTagsNames)
	require.NotNil(t, fake.removeTagsProfile)
	assert.Equal(t, "work", *fake.removeTagsProfile)
	assert.Contains(t, out, "gone1")
	assert.Contains(t, out, "missing1")
}

func TestRunUntagVar_ServiceError(t *testing.T) {
	fake := &fakeVariableService{removeTagsErr: errors.New("no such variable")}
	withFakeVariableContainer(t, fake, nil)

	err := RunUntagVar(TagVarArgs{Name: "host"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "removing tags")
}
