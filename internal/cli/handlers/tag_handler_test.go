package handlers

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

// fakeTagService is a hand-rolled stub implementing services.TagService that records the
// arguments it was called with and returns canned results, so handler tests can assert on
// argument mapping without touching a real database.
type fakeTagService struct {
	createName string
	createDesc *string
	createTag  *models.Tag
	createErr  error

	updateName  string
	updateInput services.UpdateTagConfig
	updateTag   *models.Tag
	updateErr   error

	deleteCalled bool
	deleteName   string
	deleteErr    error

	getName string
	getTag  *models.Tag
	getErr  error

	listOrder string
	listLimit *int
	listTags  []models.Tag
	listErr   error

	searchTerm   string
	searchFields []string
	searchLimit  *int
	searchTags   []models.Tag
	searchErr    error
}

func (f *fakeTagService) CreateTag(name string, description *string) (*models.Tag, error) {
	f.createName = name
	f.createDesc = description
	return f.createTag, f.createErr
}

func (f *fakeTagService) UpdateTag(name string, config services.UpdateTagConfig) (*models.Tag, error) {
	f.updateName = name
	f.updateInput = config
	return f.updateTag, f.updateErr
}

func (f *fakeTagService) DeleteTag(name string) error {
	f.deleteCalled = true
	f.deleteName = name
	return f.deleteErr
}

func (f *fakeTagService) GetTag(name string) (*models.Tag, error) {
	f.getName = name
	return f.getTag, f.getErr
}

func (f *fakeTagService) GetTagOrNil(name string) (*models.Tag, error) {
	return nil, nil
}

func (f *fakeTagService) GetTagById(id uint) (*models.Tag, error) {
	return nil, nil
}

func (f *fakeTagService) ListTags(orderBy string, limit *int) ([]models.Tag, error) {
	f.listOrder = orderBy
	f.listLimit = limit
	return f.listTags, f.listErr
}

func (f *fakeTagService) Search(query string, fields []string, limit *int) ([]models.Tag, error) {
	f.searchTerm = query
	f.searchFields = fields
	f.searchLimit = limit
	return f.searchTags, f.searchErr
}

// withFakeTagContainer swaps container.GetTagService and container.GetConsole for the duration
// of the test, restoring the originals on cleanup. The console is built lazily (at call time) so
// that tests which redirect os.Stdout via captureStdout after this setup runs still have their
// output captured.
func withFakeTagContainer(t *testing.T, svc services.TagService, svcErr error) {
	t.Helper()

	origSvc := container.GetTagService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetTagService = origSvc
		container.GetConsole = origConsole
	})

	container.GetTagService = func() (services.TagService, error) {
		return svc, svcErr
	}
	container.GetConsole = func() (*ui.Console, error) {
		return ui.NewConsole(ui.NewPalette(settings.UIColors{}), false, "never", 0, 0, 0), nil
	}
}

// withFailingTagConsole makes container.GetConsole return an error, to exercise handler error
// paths that occur before any console output happens.
func withFailingTagConsole(t *testing.T, svc services.TagService) {
	t.Helper()
	origSvc := container.GetTagService
	origConsole := container.GetConsole
	t.Cleanup(func() {
		container.GetTagService = origSvc
		container.GetConsole = origConsole
	})
	container.GetTagService = func() (services.TagService, error) {
		return svc, nil
	}
	container.GetConsole = func() (*ui.Console, error) {
		return nil, errors.New("console boom")
	}
}

// --- RunAddTag ---

func TestRunAddTag_WithDescription(t *testing.T) {
	fake := &fakeTagService{createTag: &models.Tag{Name: "infra"}}
	withFakeTagContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunAddTag(AddTagArgs{Name: "infra", HasDescription: true, Description: "infra tags"})
		require.NoError(t, err)
	})

	assert.Contains(t, out, "infra")
	assert.Equal(t, "infra", fake.createName)
	require.NotNil(t, fake.createDesc)
	assert.Equal(t, "infra tags", *fake.createDesc)
}

func TestRunAddTag_NoDescription(t *testing.T) {
	fake := &fakeTagService{createTag: &models.Tag{Name: "infra"}}
	withFakeTagContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunAddTag(AddTagArgs{Name: "infra"}))
	})

	assert.Nil(t, fake.createDesc, "no description should be passed as nil, not an empty-string pointer")
}

func TestRunAddTag_ServiceError(t *testing.T) {
	fake := &fakeTagService{createErr: errors.New("duplicate name")}
	withFakeTagContainer(t, fake, nil)

	err := RunAddTag(AddTagArgs{Name: "infra"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "creating tag")
	assert.Contains(t, err.Error(), "duplicate name")
}

func TestRunAddTag_TagServiceUnavailable(t *testing.T) {
	withFakeTagContainer(t, nil, errors.New("db down"))

	err := RunAddTag(AddTagArgs{Name: "infra"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting tag service")
	assert.Contains(t, err.Error(), "db down")
}

func TestRunAddTag_ConsoleUnavailable(t *testing.T) {
	fake := &fakeTagService{createTag: &models.Tag{Name: "infra"}}
	withFailingTagConsole(t, fake)

	err := RunAddTag(AddTagArgs{Name: "infra"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting console")
}

// --- RunGetTag ---

func TestRunGetTag_Success(t *testing.T) {
	fake := &fakeTagService{getTag: &models.Tag{Name: "infra"}}
	withFakeTagContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunGetTag("infra")
		require.NoError(t, err)
	})

	assert.Equal(t, "infra", fake.getName)
	assert.Contains(t, out, "infra")
}

func TestRunGetTag_ServiceError(t *testing.T) {
	fake := &fakeTagService{getErr: errors.New("not found")}
	withFakeTagContainer(t, fake, nil)

	err := RunGetTag("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting tag")
	assert.Contains(t, err.Error(), "not found")
}

// --- RunUpdateTag ---

func TestRunUpdateTag_RenameAndDescription(t *testing.T) {
	fake := &fakeTagService{updateTag: &models.Tag{Name: "newname"}}
	withFakeTagContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunUpdateTag(UpdateTagArgs{
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

func TestRunUpdateTag_RenameOnly(t *testing.T) {
	fake := &fakeTagService{updateTag: &models.Tag{Name: "newname"}}
	withFakeTagContainer(t, fake, nil)

	captureStdout(t, func() {
		err := RunUpdateTag(UpdateTagArgs{CurrentName: "oldname", NewName: "newname"})
		require.NoError(t, err)
	})

	in := fake.updateInput
	require.NotNil(t, in.NewName)
	assert.Equal(t, "newname", *in.NewName)
	assert.Nil(t, in.Description)
}

func TestRunUpdateTag_NoChanges(t *testing.T) {
	fake := &fakeTagService{updateTag: &models.Tag{Name: "oldname"}}
	withFakeTagContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunUpdateTag(UpdateTagArgs{CurrentName: "oldname"}))
	})

	assert.Nil(t, fake.updateInput.NewName)
	assert.Nil(t, fake.updateInput.Description)
}

func TestRunUpdateTag_ServiceError(t *testing.T) {
	fake := &fakeTagService{updateErr: errors.New("name taken")}
	withFakeTagContainer(t, fake, nil)

	err := RunUpdateTag(UpdateTagArgs{CurrentName: "oldname"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "updating tag")
	assert.Contains(t, err.Error(), "name taken")
}

// --- RunListTag ---

func TestRunListTag_WithLimit(t *testing.T) {
	fake := &fakeTagService{listTags: []models.Tag{{Name: "a"}, {Name: "b"}}}
	withFakeTagContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunListTag(ListTagArgs{ListParams: ListParams{Order: "name", HasLimit: true, Limit: 10}})
		require.NoError(t, err)
	})

	assert.Equal(t, "name", fake.listOrder)
	require.NotNil(t, fake.listLimit)
	assert.Equal(t, 10, *fake.listLimit)
	assert.Contains(t, out, "a")
	assert.Contains(t, out, "b")
}

func TestRunListTag_NoLimit(t *testing.T) {
	fake := &fakeTagService{}
	withFakeTagContainer(t, fake, nil)

	captureStdout(t, func() {
		require.NoError(t, RunListTag(ListTagArgs{}))
	})

	assert.Nil(t, fake.listLimit)
}

func TestRunListTag_ServiceError(t *testing.T) {
	fake := &fakeTagService{listErr: errors.New("db error")}
	withFakeTagContainer(t, fake, nil)

	err := RunListTag(ListTagArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "listing tags")
}

// --- RunSearchTag ---

func TestRunSearchTag_Success(t *testing.T) {
	fake := &fakeTagService{searchTags: []models.Tag{{Name: "match"}}}
	withFakeTagContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunSearchTag(SearchTagArgs{SearchParams: SearchParams{Term: "mat", SearchIn: []string{"name"}, HasLimit: true, Limit: 5}})
		require.NoError(t, err)
	})

	assert.Equal(t, "mat", fake.searchTerm)
	assert.Equal(t, []string{"name"}, fake.searchFields)
	require.NotNil(t, fake.searchLimit)
	assert.Equal(t, 5, *fake.searchLimit)
	assert.Contains(t, out, "match")
}

func TestRunSearchTag_ServiceError(t *testing.T) {
	fake := &fakeTagService{searchErr: errors.New("boom")}
	withFakeTagContainer(t, fake, nil)

	err := RunSearchTag(SearchTagArgs{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "searching tags")
}

// --- RunDeleteTag ---

func TestRunDeleteTag_Success(t *testing.T) {
	fake := &fakeTagService{getTag: &models.Tag{Name: "gone"}}
	withFakeTagContainer(t, fake, nil)

	out := captureStdout(t, func() {
		err := RunDeleteTag(DeleteTagArgs{Name: "gone"})
		require.NoError(t, err)
	})

	assert.Equal(t, "gone", fake.getName)
	assert.True(t, fake.deleteCalled)
	assert.Equal(t, "gone", fake.deleteName)
	assert.Contains(t, out, "gone")
}

func TestRunDeleteTag_GetFails_DoesNotDelete(t *testing.T) {
	fake := &fakeTagService{getErr: errors.New("not found")}
	withFakeTagContainer(t, fake, nil)

	err := RunDeleteTag(DeleteTagArgs{Name: "missing"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "getting tag")
	assert.False(t, fake.deleteCalled, "delete should not be attempted when the lookup fails")
}

func TestRunDeleteTag_DeleteFails(t *testing.T) {
	fake := &fakeTagService{getTag: &models.Tag{Name: "gone"}, deleteErr: errors.New("locked")}
	withFakeTagContainer(t, fake, nil)

	err := RunDeleteTag(DeleteTagArgs{Name: "gone"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleting tag")
	assert.Contains(t, err.Error(), "locked")
}
