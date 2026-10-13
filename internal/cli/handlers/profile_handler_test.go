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

// fakeProfileService is a hand-rolled stub implementing services.ProfileService that records the
// arguments it was called with and returns canned results, so handler tests can assert on
// argument mapping without touching a real database.
type fakeProfileService struct {
	createName string
	createDesc *string
	createProf *models.Profile
	createErr  error

	updateName  string
	updateInput services.UpdateProfileConfig
	updateProf  *models.Profile
	updateErr   error

	deleteCalled bool
	deleteName   string
	deleteForce  bool
	deleteResult repository.ProfileDeleteResult
	deleteErr    error

	getName string
	getProf *models.Profile
	getErr  error

	listOrder string
	listLimit *int
	listProfs []models.Profile
	listErr   error

	searchTerm   string
	searchFields []string
	searchLimit  *int
	searchProfs  []models.Profile
	searchErr    error

	statusResult services.ProfileStatus
	statusErr    error

	switchCalled         bool
	switchCommandCalled  bool
	switchVariableCalled bool
	switchSettingsCalled bool
	switchName           string
	switchState          *models.ProfileState
	switchErr            error
	switchCommandErr     error
	switchVariableErr    error
	switchSettingsErr    error
}

func (f *fakeProfileService) CreateProfile(name string, description *string) (*models.Profile, error) {
	f.createName = name
	f.createDesc = description
	return f.createProf, f.createErr
}

func (f *fakeProfileService) UpdateProfile(name string, input services.UpdateProfileConfig) (*models.Profile, error) {
	f.updateName = name
	f.updateInput = input
	return f.updateProf, f.updateErr
}

func (f *fakeProfileService) DeleteProfile(name string, force bool) (repository.ProfileDeleteResult, error) {
	f.deleteCalled = true
	f.deleteName = name
	f.deleteForce = force
	return f.deleteResult, f.deleteErr
}

func (f *fakeProfileService) GetProfile(name string) (*models.Profile, error) {
	f.getName = name
	return f.getProf, f.getErr
}

func (f *fakeProfileService) ListProfiles(orderBy string, limit *int) ([]models.Profile, error) {
	f.listOrder = orderBy
	f.listLimit = limit
	return f.listProfs, f.listErr
}

func (f *fakeProfileService) SearchProfiles(query string, fields []string, limit *int) ([]models.Profile, error) {
	f.searchTerm = query
	f.searchFields = fields
	f.searchLimit = limit
	return f.searchProfs, f.searchErr
}

func (f *fakeProfileService) GetActiveCommandProfile() (*models.Profile, error) {
	return nil, nil
}

func (f *fakeProfileService) GetActiveVariableProfile() (*models.Profile, error) {
	return nil, nil
}

func (f *fakeProfileService) GetActiveSettingsProfile() (*models.Profile, error) {
	return nil, nil
}

func (f *fakeProfileService) SwitchProfile(name string) (*models.ProfileState, error) {
	f.switchCalled = true
	f.switchName = name
	return f.switchState, f.switchErr
}

func (f *fakeProfileService) SwitchCommandProfile(name string) (*models.ProfileState, error) {
	f.switchCommandCalled = true
	f.switchName = name
	return f.switchState, f.switchCommandErr
}

func (f *fakeProfileService) SwitchVariableProfile(name string) (*models.ProfileState, error) {
	f.switchVariableCalled = true
	f.switchName = name
	return f.switchState, f.switchVariableErr
}

func (f *fakeProfileService) SwitchSettingsProfile(name string) (*models.ProfileState, error) {
	f.switchSettingsCalled = true
	f.switchName = name
	return f.switchState, f.switchSettingsErr
}

func (f *fakeProfileService) GetStatus() (services.ProfileStatus, error) {
	return f.statusResult, f.statusErr
}

// withFakeProfileContainer swaps container.GetProfileService and container.GetConsole for the
// duration of the test, restoring the originals on cleanup.
func withFakeProfileContainer(t *testing.T, svc services.ProfileService, svcErr error) {
	t.Helper()

	origSvc := container.GetProfileService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetProfileService = origSvc
		container.GetConsole = origConsole
	})

	container.GetProfileService = func() (services.ProfileService, error) {
		return svc, svcErr
	}
	container.GetConsole = func() (*ui.Console, error) {
		return ui.NewConsole(ui.NewPalette(settings.UIColors{}), false, "never", 0, 0, 0), nil
	}
}

// withFailingProfileConsole makes container.GetConsole return an error, to exercise handler error
// paths that occur before any console output happens.
func withFailingProfileConsole(t *testing.T, svc services.ProfileService) {
	t.Helper()
	origSvc := container.GetProfileService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetProfileService = origSvc
		container.GetConsole = origConsole
	})
	container.GetProfileService = func() (services.ProfileService, error) {
		return svc, nil
	}
	container.GetConsole = func() (*ui.Console, error) {
		return nil, errors.New("console boom")
	}
}

func makeProfileState(name string) *models.ProfileState {
	return &models.ProfileState{
		ActiveCommandProfile:  models.Profile{Name: name},
		ActiveVariableProfile: models.Profile{Name: name},
		ActiveSettingsProfile: models.Profile{Name: name},
	}
}

// --- RunAddProfile ---

func TestRunAddProfile_WithDescription(t *testing.T) {
	fake := &fakeProfileService{createProf: &models.Profile{Name: "work"}}
	withFakeProfileContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunAddProfile(AddProfileArgs{Name: "work", HasDescription: true, Description: "work profile"})
		require.NoError(t, err)
	})

	assert.Contains(t, out, "work")
	assert.Equal(t, "work", fake.createName)
	require.NotNil(t, fake.createDesc)
	assert.Equal(t, "work profile", *fake.createDesc)
}

func TestRunAddProfile_NoDescription(t *testing.T) {
	fake := &fakeProfileService{createProf: &models.Profile{Name: "work"}}
	withFakeProfileContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunAddProfile(AddProfileArgs{Name: "work"}))
	})

	assert.Nil(t, fake.createDesc, "no description should be passed as nil, not an empty-string pointer")
}

func TestRunAddProfile_ServiceError(t *testing.T) {
	fake := &fakeProfileService{createErr: errors.New("duplicate name")}
	withFakeProfileContainer(t, fake, nil)

	err := RunAddProfile(AddProfileArgs{Name: "work"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating profile")
	assert.Contains(t, err.Error(), "duplicate name")
}

func TestRunAddProfile_ProfileServiceUnavailable(t *testing.T) {
	withFakeProfileContainer(t, nil, errors.New("db down"))

	err := RunAddProfile(AddProfileArgs{Name: "work"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting profile service")
	assert.Contains(t, err.Error(), "db down")
}

func TestRunAddProfile_ConsoleUnavailable(t *testing.T) {
	fake := &fakeProfileService{createProf: &models.Profile{Name: "work"}}
	withFailingProfileConsole(t, fake)

	err := RunAddProfile(AddProfileArgs{Name: "work"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

// --- RunGetProfile ---

func TestRunGetProfile_Success(t *testing.T) {
	fake := &fakeProfileService{getProf: &models.Profile{Name: "work"}}
	withFakeProfileContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunGetProfile("work")
		require.NoError(t, err)
	})

	assert.Equal(t, "work", fake.getName)
	assert.Contains(t, out, "work")
}

func TestRunGetProfile_ServiceError(t *testing.T) {
	fake := &fakeProfileService{getErr: errors.New("not found")}
	withFakeProfileContainer(t, fake, nil)

	err := RunGetProfile("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting profile")
	assert.Contains(t, err.Error(), "not found")
}

// --- RunUpdateProfile ---

func TestRunUpdateProfile_RenameAndDescription(t *testing.T) {
	fake := &fakeProfileService{updateProf: &models.Profile{Name: "newname"}}
	withFakeProfileContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunUpdateProfile(UpdateProfileArgs{
			CurrentName:    "oldname",
			NewName:        "newname",
			HasDescription: true,
			Description:    "new desc",
		})
		require.NoError(t, err)
	})

	assert.Equal(t, "oldname", fake.updateName)
	in := fake.updateInput
	require.NotNil(t, in.NewName)
	assert.Equal(t, "newname", *in.NewName)
	require.NotNil(t, in.Description)
	assert.Equal(t, "new desc", *in.Description)
	assert.Contains(t, out, "newname")
}

func TestRunUpdateProfile_RenameOnly(t *testing.T) {
	fake := &fakeProfileService{updateProf: &models.Profile{Name: "newname"}}
	withFakeProfileContainer(t, fake, nil)

	captureStdout(t, func() {
		err := RunUpdateProfile(UpdateProfileArgs{CurrentName: "oldname", NewName: "newname"})
		require.NoError(t, err)
	})

	in := fake.updateInput
	require.NotNil(t, in.NewName)
	assert.Equal(t, "newname", *in.NewName)
	assert.Nil(t, in.Description)
}

func TestRunUpdateProfile_NoChanges(t *testing.T) {
	fake := &fakeProfileService{updateProf: &models.Profile{Name: "oldname"}}
	withFakeProfileContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunUpdateProfile(UpdateProfileArgs{CurrentName: "oldname"}))
	})

	assert.Nil(t, fake.updateInput.NewName)
	assert.Nil(t, fake.updateInput.Description)
}

func TestRunUpdateProfile_ServiceError(t *testing.T) {
	fake := &fakeProfileService{updateErr: errors.New("name taken")}
	withFakeProfileContainer(t, fake, nil)

	err := RunUpdateProfile(UpdateProfileArgs{CurrentName: "oldname"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "updating profile")
	assert.Contains(t, err.Error(), "name taken")
}

func TestRunUpdateProfile_ProfileServiceUnavailable(t *testing.T) {
	withFakeProfileContainer(t, nil, errors.New("db down"))

	err := RunUpdateProfile(UpdateProfileArgs{CurrentName: "oldname"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting profile service")
}

func TestRunUpdateProfile_ConsoleUnavailable(t *testing.T) {
	fake := &fakeProfileService{updateProf: &models.Profile{Name: "oldname"}}
	withFailingProfileConsole(t, fake)

	err := RunUpdateProfile(UpdateProfileArgs{CurrentName: "oldname"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

// --- RunListProfile ---

func TestRunListProfile_WithLimit(t *testing.T) {
	fake := &fakeProfileService{listProfs: []models.Profile{{Name: "a"}, {Name: "b"}}}
	withFakeProfileContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunListProfile(ListProfileArgs{ListParams: ListParams{Order: "name", HasLimit: true, Limit: 10}})
		require.NoError(t, err)
	})

	assert.Equal(t, "name", fake.listOrder)
	require.NotNil(t, fake.listLimit)
	assert.Equal(t, 10, *fake.listLimit)
	assert.Contains(t, out, "a")
	assert.Contains(t, out, "b")
}

func TestRunListProfile_NoLimit(t *testing.T) {
	fake := &fakeProfileService{}
	withFakeProfileContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunListProfile(ListProfileArgs{}))
	})

	assert.Nil(t, fake.listLimit)
}

func TestRunListProfile_ServiceError(t *testing.T) {
	fake := &fakeProfileService{listErr: errors.New("db error")}
	withFakeProfileContainer(t, fake, nil)

	err := RunListProfile(ListProfileArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing profiles")
	assert.Contains(t, err.Error(), "db error")
}

func TestRunListProfile_ProfileServiceUnavailable(t *testing.T) {
	withFakeProfileContainer(t, nil, errors.New("db down"))

	err := RunListProfile(ListProfileArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting profile service")
}

// --- RunDeleteProfile ---

func TestRunDeleteProfile_Success(t *testing.T) {
	fake := &fakeProfileService{
		getProf:      &models.Profile{Name: "gone"},
		deleteResult: repository.ProfileDeleteResult{CommandsDeleted: 2, VariablesDeleted: 3},
	}
	withFakeProfileContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunDeleteProfile(DeleteProfileArgs{Name: "gone"})
		require.NoError(t, err)
	})

	assert.Equal(t, "gone", fake.getName)
	assert.True(t, fake.deleteCalled)
	assert.Equal(t, "gone", fake.deleteName)
	assert.False(t, fake.deleteForce)
	assert.Contains(t, out, "gone")
}

func TestRunDeleteProfile_Force(t *testing.T) {
	fake := &fakeProfileService{getProf: &models.Profile{Name: "gone"}}
	withFakeProfileContainer(t, fake, nil)

	captureStdout(t, func() {
		err := RunDeleteProfile(DeleteProfileArgs{Name: "gone", Force: true})
		require.NoError(t, err)
	})

	assert.True(t, fake.deleteForce)
}

func TestRunDeleteProfile_GetFails_DoesNotDelete(t *testing.T) {
	fake := &fakeProfileService{getErr: errors.New("not found")}
	withFakeProfileContainer(t, fake, nil)

	err := RunDeleteProfile(DeleteProfileArgs{Name: "missing"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting profile")
	assert.False(t, fake.deleteCalled, "delete should not be attempted when the lookup fails")
}

func TestRunDeleteProfile_DeleteFails(t *testing.T) {
	fake := &fakeProfileService{getProf: &models.Profile{Name: "gone"}, deleteErr: errors.New("locked")}
	withFakeProfileContainer(t, fake, nil)

	err := RunDeleteProfile(DeleteProfileArgs{Name: "gone"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleting profile")
	assert.Contains(t, err.Error(), "locked")
}

func TestRunDeleteProfile_ProfileServiceUnavailable(t *testing.T) {
	withFakeProfileContainer(t, nil, errors.New("db down"))

	err := RunDeleteProfile(DeleteProfileArgs{Name: "gone"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting profile service")
}

// --- RunProfileStatus ---

func TestRunProfileStatus_Success(t *testing.T) {
	fake := &fakeProfileService{statusResult: services.ProfileStatus{
		CommandProfile:  "work",
		VariableProfile: "work",
		SettingsProfile: "work",
	}}
	withFakeProfileContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunProfileStatus()
		require.NoError(t, err)
	})

	assert.Contains(t, out, "work")
}

func TestRunProfileStatus_ServiceError(t *testing.T) {
	fake := &fakeProfileService{statusErr: errors.New("boom")}
	withFakeProfileContainer(t, fake, nil)

	err := RunProfileStatus()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting profile status")
	assert.Contains(t, err.Error(), "boom")
}

func TestRunProfileStatus_ProfileServiceUnavailable(t *testing.T) {
	withFakeProfileContainer(t, nil, errors.New("db down"))

	err := RunProfileStatus()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting profile service")
}

func TestRunProfileStatus_ConsoleUnavailable(t *testing.T) {
	fake := &fakeProfileService{}
	withFailingProfileConsole(t, fake)

	err := RunProfileStatus()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

// --- RunSwitchProfile ---

func TestRunSwitchProfile_DefaultSwitchesAllThree(t *testing.T) {
	fake := &fakeProfileService{switchState: makeProfileState("work")}
	withFakeProfileContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunSwitchProfile(SwitchProfileArgs{Name: "work"})
		require.NoError(t, err)
	})

	assert.True(t, fake.switchCalled, "default switch should call the combined SwitchProfile")
	assert.False(t, fake.switchCommandCalled)
	assert.False(t, fake.switchVariableCalled)
	assert.False(t, fake.switchSettingsCalled)
	assert.Equal(t, "work", fake.switchName)
	assert.Contains(t, out, "work")
	assert.Contains(t, out, "command")
	assert.Contains(t, out, "variable")
	assert.Contains(t, out, "settings")
}

func TestRunSwitchProfile_CmdOnly(t *testing.T) {
	fake := &fakeProfileService{switchState: makeProfileState("work")}
	withFakeProfileContainer(t, fake, nil)

	captureStdout(t, func() {
		err := RunSwitchProfile(SwitchProfileArgs{Name: "work", Cmd: true})
		require.NoError(t, err)
	})

	assert.True(t, fake.switchCommandCalled)
	assert.False(t, fake.switchCalled)
	assert.False(t, fake.switchVariableCalled)
	assert.False(t, fake.switchSettingsCalled)
}

func TestRunSwitchProfile_VarOnly(t *testing.T) {
	fake := &fakeProfileService{switchState: makeProfileState("work")}
	withFakeProfileContainer(t, fake, nil)

	captureStdout(t, func() {
		err := RunSwitchProfile(SwitchProfileArgs{Name: "work", Var: true})
		require.NoError(t, err)
	})

	assert.True(t, fake.switchVariableCalled)
	assert.False(t, fake.switchCommandCalled)
	assert.False(t, fake.switchSettingsCalled)
}

func TestRunSwitchProfile_SettingsOnly(t *testing.T) {
	fake := &fakeProfileService{switchState: makeProfileState("work")}
	withFakeProfileContainer(t, fake, nil)

	captureStdout(t, func() {
		err := RunSwitchProfile(SwitchProfileArgs{Name: "work", Settings: true})
		require.NoError(t, err)
	})

	assert.True(t, fake.switchSettingsCalled)
	assert.False(t, fake.switchCommandCalled)
	assert.False(t, fake.switchVariableCalled)
}

func TestRunSwitchProfile_CmdAndVar(t *testing.T) {
	fake := &fakeProfileService{switchState: makeProfileState("work")}
	withFakeProfileContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunSwitchProfile(SwitchProfileArgs{Name: "work", Cmd: true, Var: true})
		require.NoError(t, err)
	})

	assert.True(t, fake.switchCommandCalled)
	assert.True(t, fake.switchVariableCalled)
	assert.False(t, fake.switchSettingsCalled)
	assert.Contains(t, out, "command")
	assert.Contains(t, out, "variable")
	assert.NotContains(t, out, "settings profile")
}

func TestRunSwitchProfile_DefaultError(t *testing.T) {
	fake := &fakeProfileService{switchErr: errors.New("no such profile")}
	withFakeProfileContainer(t, fake, nil)

	err := RunSwitchProfile(SwitchProfileArgs{Name: "missing"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "switching command profiles")
	assert.Contains(t, err.Error(), "no such profile")
}

func TestRunSwitchProfile_CmdError(t *testing.T) {
	fake := &fakeProfileService{switchCommandErr: errors.New("no such profile")}
	withFakeProfileContainer(t, fake, nil)

	err := RunSwitchProfile(SwitchProfileArgs{Name: "missing", Cmd: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "switching command profiles")
}

func TestRunSwitchProfile_VarErrorStopsBeforeSettings(t *testing.T) {
	fake := &fakeProfileService{switchVariableErr: errors.New("no such profile")}
	withFakeProfileContainer(t, fake, nil)

	err := RunSwitchProfile(SwitchProfileArgs{Name: "missing", Var: true, Settings: true})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "switching variable profiles")
	assert.False(t, fake.switchSettingsCalled, "settings switch should not run after an earlier switch fails")
}

func TestRunSwitchProfile_ProfileServiceUnavailable(t *testing.T) {
	withFakeProfileContainer(t, nil, errors.New("db down"))

	err := RunSwitchProfile(SwitchProfileArgs{Name: "work"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting profile service")
}

func TestRunSwitchProfile_ConsoleUnavailable(t *testing.T) {
	fake := &fakeProfileService{switchState: makeProfileState("work")}
	withFailingProfileConsole(t, fake)

	err := RunSwitchProfile(SwitchProfileArgs{Name: "work"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}
