package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
)

// --- fakes ---

func cmdKey(alias string, profileName *string) string {
	p := ""
	if profileName != nil {
		p = *profileName
	}
	return p + "|" + alias
}

type fakeCommandService struct {
	commands map[string]*models.Command

	getOrNilErr    error
	getWithTagsErr error
	createErr      error
	updateErr      error
	addTagsErr     error
	removeTagsErr  error

	createCalls     []CreateCommandConfig
	updateCalls     []UpdateCommandConfig
	addTagsCalls    map[string][]string
	removeTagsCalls map[string][]string
}

func newFakeCommandService() *fakeCommandService {
	return &fakeCommandService{
		commands:        map[string]*models.Command{},
		addTagsCalls:    map[string][]string{},
		removeTagsCalls: map[string][]string{},
	}
}

func (f *fakeCommandService) seed(alias string, profileName *string, tags []models.Tag) {
	f.commands[cmdKey(alias, profileName)] = &models.Command{Alias: alias, Tags: tags}
}

func (f *fakeCommandService) GetCommandOrNil(alias string, profileName *string) (*models.Command, error) {
	if f.getOrNilErr != nil {
		return nil, f.getOrNilErr
	}
	return f.commands[cmdKey(alias, profileName)], nil
}

func (f *fakeCommandService) GetCommandWithTags(alias string, profileName *string) (*models.Command, error) {
	if f.getWithTagsErr != nil {
		return nil, f.getWithTagsErr
	}
	cmd, ok := f.commands[cmdKey(alias, profileName)]
	if !ok {
		return nil, fmt.Errorf("fake: command %q not found", alias)
	}
	return cmd, nil
}

func (f *fakeCommandService) CreateCommand(input CreateCommandConfig) (*models.Command, error) {
	f.createCalls = append(f.createCalls, input)
	if f.createErr != nil {
		return nil, f.createErr
	}
	cmd := &models.Command{Alias: input.Alias}
	f.commands[cmdKey(input.Alias, input.ProfileName)] = cmd
	return cmd, nil
}

func (f *fakeCommandService) UpdateCommand(alias string, profileName *string, input UpdateCommandConfig) (*models.Command, error) {
	f.updateCalls = append(f.updateCalls, input)
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return f.commands[cmdKey(alias, profileName)], nil
}

func (f *fakeCommandService) AddTags(alias string, tagNames []string, profileName *string) (repository.TagAttachResult, error) {
	f.addTagsCalls[alias] = append(f.addTagsCalls[alias], tagNames...)
	if f.addTagsErr != nil {
		return repository.TagAttachResult{}, f.addTagsErr
	}
	return repository.TagAttachResult{Added: tagNames}, nil
}

func (f *fakeCommandService) RemoveTags(alias string, tagNames []string, profileName *string) (repository.TagDetachResult, error) {
	f.removeTagsCalls[alias] = append(f.removeTagsCalls[alias], tagNames...)
	if f.removeTagsErr != nil {
		return repository.TagDetachResult{}, f.removeTagsErr
	}
	return repository.TagDetachResult{Removed: tagNames}, nil
}

func (f *fakeCommandService) DeleteCommand(alias string, profileName *string) error {
	panic("not implemented in fake")
}

func (f *fakeCommandService) GetCommand(alias string, profileName *string) (*models.Command, error) {
	panic("not implemented in fake")
}

func (f *fakeCommandService) GetCommandByID(id uint, profileName *string) (*models.Command, error) {
	panic("not implemented in fake")
}

func (f *fakeCommandService) ListCommands(orderBy string, tagNames []string, limit *int, profileName *string) ([]models.Command, error) {
	panic("not implemented in fake")
}

func (f *fakeCommandService) SearchCommands(query string, fields []string, limit *int, profileName *string) ([]models.Command, error) {
	panic("not implemented in fake")
}

func (f *fakeCommandService) MoveCommand(alias string, targetProfileName, profileName *string) (*models.Command, error) {
	panic("not implemented in fake")
}

func (f *fakeCommandService) CopyCommand(alias string, targetProfileName, newAlias, profileName *string) (*models.Command, error) {
	panic("not implemented in fake")
}

type fakeVariableService struct {
	variables map[string]*models.Variable

	getOrNilErr    error
	getWithTagsErr error
	createErr      error
	updateErr      error
	addTagsErr     error
	removeTagsErr  error

	createCalls     []CreateVariableConfig
	updateCalls     []UpdateVariableConfig
	addTagsCalls    map[string][]string
	removeTagsCalls map[string][]string
}

func newFakeVariableService() *fakeVariableService {
	return &fakeVariableService{
		variables:       map[string]*models.Variable{},
		addTagsCalls:    map[string][]string{},
		removeTagsCalls: map[string][]string{},
	}
}

func (f *fakeVariableService) seed(name string, profileName *string, tags []models.Tag) {
	f.variables[cmdKey(name, profileName)] = &models.Variable{Name: name, Tags: tags}
}

func (f *fakeVariableService) GetVariableOrNil(name string, profileName *string) (*models.Variable, error) {
	if f.getOrNilErr != nil {
		return nil, f.getOrNilErr
	}
	return f.variables[cmdKey(name, profileName)], nil
}

func (f *fakeVariableService) GetVariableWithTags(name string, profileName *string) (*models.Variable, error) {
	if f.getWithTagsErr != nil {
		return nil, f.getWithTagsErr
	}
	v, ok := f.variables[cmdKey(name, profileName)]
	if !ok {
		return nil, fmt.Errorf("fake: variable %q not found", name)
	}
	return v, nil
}

func (f *fakeVariableService) CreateVariable(input CreateVariableConfig) (*models.Variable, error) {
	f.createCalls = append(f.createCalls, input)
	if f.createErr != nil {
		return nil, f.createErr
	}
	v := &models.Variable{Name: input.Name, Value: input.Value}
	f.variables[cmdKey(input.Name, input.ProfileName)] = v
	return v, nil
}

func (f *fakeVariableService) UpdateVariable(name string, profileName *string, input UpdateVariableConfig) (*models.Variable, error) {
	f.updateCalls = append(f.updateCalls, input)
	if f.updateErr != nil {
		return nil, f.updateErr
	}
	return f.variables[cmdKey(name, profileName)], nil
}

func (f *fakeVariableService) AddTags(name string, tagNames []string, profileName *string) (repository.TagAttachResult, error) {
	f.addTagsCalls[name] = append(f.addTagsCalls[name], tagNames...)
	if f.addTagsErr != nil {
		return repository.TagAttachResult{}, f.addTagsErr
	}
	return repository.TagAttachResult{Added: tagNames}, nil
}

func (f *fakeVariableService) RemoveTags(name string, tagNames []string, profileName *string) (repository.TagDetachResult, error) {
	f.removeTagsCalls[name] = append(f.removeTagsCalls[name], tagNames...)
	if f.removeTagsErr != nil {
		return repository.TagDetachResult{}, f.removeTagsErr
	}
	return repository.TagDetachResult{Removed: tagNames}, nil
}

func (f *fakeVariableService) DeleteVariable(name string, profileName *string) error {
	panic("not implemented in fake")
}

func (f *fakeVariableService) GetVariable(name string, profileName *string) (*models.Variable, error) {
	panic("not implemented in fake")
}

func (f *fakeVariableService) GetVariableByID(id uint, profileName *string) (*models.Variable, error) {
	panic("not implemented in fake")
}

func (f *fakeVariableService) ListVariables(orderBy string, tagNames []string, limit *int, profileName *string) ([]models.Variable, error) {
	panic("not implemented in fake")
}

func (f *fakeVariableService) SearchVariables(query string, fields []string, limit *int, profileName *string) ([]models.Variable, error) {
	panic("not implemented in fake")
}

func (f *fakeVariableService) MoveVariable(name string, targetProfileName, profileName *string) (*models.Variable, error) {
	panic("not implemented in fake")
}

func (f *fakeVariableService) CopyVariable(name string, targetProfileName, newName, profileName *string) (*models.Variable, error) {
	panic("not implemented in fake")
}

type fakeTagService struct {
	existing map[string]*models.Tag

	getOrNilErr error
	createErr   error

	createCalls []string
}

func newFakeTagService() *fakeTagService {
	return &fakeTagService{existing: map[string]*models.Tag{}}
}

func (f *fakeTagService) GetTagOrNil(name string) (*models.Tag, error) {
	if f.getOrNilErr != nil {
		return nil, f.getOrNilErr
	}
	return f.existing[name], nil
}

func (f *fakeTagService) CreateTag(name string, description *string) (*models.Tag, error) {
	f.createCalls = append(f.createCalls, name)
	if f.createErr != nil {
		return nil, f.createErr
	}
	tag := &models.Tag{Name: name}
	f.existing[name] = tag
	return tag, nil
}

func (f *fakeTagService) UpdateTag(name string, config *UpdateTagConfig) (*models.Tag, error) {
	panic("not implemented in fake")
}

func (f *fakeTagService) DeleteTag(name string) error {
	panic("not implemented in fake")
}

func (f *fakeTagService) GetTag(name string) (*models.Tag, error) {
	panic("not implemented in fake")
}

func (f *fakeTagService) GetTagById(id uint) (*models.Tag, error) {
	panic("not implemented in fake")
}

func (f *fakeTagService) ListTags(orderBy string, limit *int) ([]models.Tag, error) {
	panic("not implemented in fake")
}

func (f *fakeTagService) Search(query string, fields []string, limit *int) ([]models.Tag, error) {
	panic("not implemented in fake")
}

type fakeProfileRepo struct {
	repository.ProfileRepository
	activeProfile *models.Profile
	activeErr     error
}

func (f *fakeProfileRepo) GetActiveCommandProfile() (*models.Profile, error) {
	if f.activeErr != nil {
		return nil, f.activeErr
	}
	return f.activeProfile, nil
}

// --- test harness ---

type importServiceHarness struct {
	svc         *ImportService
	cmdSvc      *fakeCommandService
	varSvc      *fakeVariableService
	tagSvc      *fakeTagService
	profileRepo *fakeProfileRepo
}

func newImportServiceHarness() *importServiceHarness {
	cmdSvc := newFakeCommandService()
	varSvc := newFakeVariableService()
	tagSvc := newFakeTagService()
	profileRepo := &fakeProfileRepo{activeProfile: &models.Profile{Name: repository.DefaultProfileName}}

	svc := NewImportService(cmdSvc, varSvc, tagSvc, profileRepo)

	return &importServiceHarness{
		svc:         svc,
		cmdSvc:      cmdSvc,
		varSvc:      varSvc,
		tagSvc:      tagSvc,
		profileRepo: profileRepo,
	}
}

func writeImportFile(t *testing.T, doc ImportDocument) string {
	t.Helper()
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal import doc: %v", err)
	}
	path := filepath.Join(t.TempDir(), "import.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write import file: %v", err)
	}
	return path
}

func minimalDoc() ImportDocument {
	return ImportDocument{Version: "1", Type: "cmdbox-export"}
}

func sortedCopy(s []string) []string {
	out := make([]string, len(s))
	copy(out, s)
	sort.Strings(out)
	return out
}

// --- buildDependencyGraph / cycle detection ---

func TestBuildDependencyGraph(t *testing.T) {
	doc := ImportDocument{
		Commands: []ImportCommand{
			{Alias: "build", Template: "go build <var:target> && <cmd:test>"},
		},
		Variables: []ImportVariable{
			{Name: "target", Value: "./..."},
		},
	}

	deps := buildDependencyGraph(doc)

	cmdDeps := deps["command:build"]
	if len(cmdDeps) != 2 || sortedCopy(cmdDeps)[0] != "command:test" || sortedCopy(cmdDeps)[1] != "variable:target" {
		t.Fatalf("deps[command:build] = %v, want [command:test variable:target]", cmdDeps)
	}

	varDeps, ok := deps["variable:target"]
	if !ok || len(varDeps) != 0 {
		t.Fatalf("deps[variable:target] = %v, want empty slice present", varDeps)
	}
}

func TestBuildDependencyGraphVariableReferences(t *testing.T) {
	doc := ImportDocument{
		Variables: []ImportVariable{
			{Name: "full", Value: "<var:base>/<cmd:setup>"},
		},
	}

	deps := buildDependencyGraph(doc)

	varDeps := sortedCopy(deps["variable:full"])
	if len(varDeps) != 2 || varDeps[0] != "command:setup" || varDeps[1] != "variable:base" {
		t.Fatalf("deps[variable:full] = %v, want [command:setup variable:base]", varDeps)
	}
}

func TestValidateNoCyclesDiamondIsNotACycle(t *testing.T) {
	// a depends on b and c, both of which depend on d - d gets visited twice but is never on the
	// stack the second time, exercising findCycle's "already visited, not a cycle" branch.
	deps := map[string][]string{
		"command:a": {"command:b", "command:c"},
		"command:b": {"command:d"},
		"command:c": {"command:d"},
		"command:d": {},
	}
	if err := validateNoCycles(deps); err != nil {
		t.Fatalf("validateNoCycles() error = %v, want nil", err)
	}
}

func TestValidateNoCyclesNoCycle(t *testing.T) {
	deps := map[string][]string{
		"command:a": {"command:b"},
		"command:b": {},
	}
	if err := validateNoCycles(deps); err != nil {
		t.Fatalf("validateNoCycles() error = %v, want nil", err)
	}
}

func TestValidateNoCyclesDirectCycle(t *testing.T) {
	deps := map[string][]string{
		"command:a": {"command:b"},
		"command:b": {"command:a"},
	}
	if err := validateNoCycles(deps); !errors.Is(err, ErrImportCycle) {
		t.Fatalf("validateNoCycles() error = %v, want ErrImportCycle", err)
	}
}

func TestValidateNoCyclesSelfCycle(t *testing.T) {
	deps := map[string][]string{
		"command:a": {"command:a"},
	}
	if err := validateNoCycles(deps); !errors.Is(err, ErrImportCycle) {
		t.Fatalf("validateNoCycles() error = %v, want ErrImportCycle", err)
	}
}

func TestValidateNoCyclesDependencyOnUnknownNodeIsFine(t *testing.T) {
	deps := map[string][]string{
		"command:a": {"variable:missing"},
	}
	if err := validateNoCycles(deps); err != nil {
		t.Fatalf("validateNoCycles() error = %v, want nil", err)
	}
}

// --- parseImportFile ---

func TestParseImportFileMissingFile(t *testing.T) {
	_, err := parseImportFile(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err == nil {
		t.Fatal("parseImportFile() error = nil, want error")
	}
	var fileErr *ImportFileError
	if !errors.As(err, &fileErr) {
		t.Fatalf("parseImportFile() error = %v, want *ImportFileError", err)
	}
	if !errors.Is(err, ErrImportFile) {
		t.Fatalf("parseImportFile() error does not match ErrImportFile sentinel")
	}
}

func TestParseImportFileInvalidJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bad.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	_, err := parseImportFile(path)
	var fileErr *ImportFileError
	if !errors.As(err, &fileErr) {
		t.Fatalf("parseImportFile() error = %v, want *ImportFileError", err)
	}
}

func TestParseImportFileUnsupportedVersion(t *testing.T) {
	path := writeImportFile(t, ImportDocument{Version: "999"})
	_, err := parseImportFile(path)
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("parseImportFile() error = %v, want ErrUnsupportedVersion", err)
	}
	if !errors.Is(err, ErrImportFile) {
		t.Fatalf("parseImportFile() error does not match ErrImportFile sentinel")
	}
}

func TestParseImportFileValid(t *testing.T) {
	path := writeImportFile(t, minimalDoc())
	doc, err := parseImportFile(path)
	if err != nil {
		t.Fatalf("parseImportFile() error = %v", err)
	}
	if doc.Version != "1" {
		t.Fatalf("doc.Version = %q, want %q", doc.Version, "1")
	}
}

// --- resolveProfileName ---

func TestResolveProfileNameExplicit(t *testing.T) {
	h := newImportServiceHarness()
	h.profileRepo.activeErr = errors.New("should not be called")

	name, err := h.svc.resolveProfileName(strPtr("work"))
	if err != nil {
		t.Fatalf("resolveProfileName() error = %v", err)
	}
	if name != "work" {
		t.Fatalf("resolveProfileName() = %q, want %q", name, "work")
	}
}

func TestResolveProfileNameFallsBackToActive(t *testing.T) {
	h := newImportServiceHarness()
	h.profileRepo.activeProfile = &models.Profile{Name: "active-one"}

	name, err := h.svc.resolveProfileName(nil)
	if err != nil {
		t.Fatalf("resolveProfileName() error = %v", err)
	}
	if name != "active-one" {
		t.Fatalf("resolveProfileName() = %q, want %q", name, "active-one")
	}
}

func TestResolveProfileNameActiveProfileError(t *testing.T) {
	h := newImportServiceHarness()
	sentinel := errors.New("boom")
	h.profileRepo.activeErr = sentinel

	_, err := h.svc.resolveProfileName(nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("resolveProfileName() error = %v, want wrapping %v", err, sentinel)
	}
}

// --- ImportFile: top level plumbing ---

func TestImportFileProfileResolutionError(t *testing.T) {
	h := newImportServiceHarness()
	sentinel := errors.New("boom")
	h.profileRepo.activeErr = sentinel
	path := writeImportFile(t, minimalDoc())

	_, err := h.svc.ImportFile(path, false, false, nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileParseError(t *testing.T) {
	h := newImportServiceHarness()

	_, err := h.svc.ImportFile(filepath.Join(t.TempDir(), "missing.json"), false, false, strPtr("default"))
	var fileErr *ImportFileError
	if !errors.As(err, &fileErr) {
		t.Fatalf("ImportFile() error = %v, want wrapping *ImportFileError", err)
	}
}

func TestImportFileCycleDetected(t *testing.T) {
	h := newImportServiceHarness()
	doc := ImportDocument{
		Version: "1",
		Commands: []ImportCommand{
			{Alias: "a", Template: "<cmd:b>"},
			{Alias: "b", Template: "<cmd:a>"},
		},
	}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, false, false, strPtr("default"))
	if !errors.Is(err, ErrImportCycle) {
		t.Fatalf("ImportFile() error = %v, want ErrImportCycle", err)
	}
}

func TestImportFileGetCommandOrNilError(t *testing.T) {
	h := newImportServiceHarness()
	sentinel := errors.New("boom")
	h.cmdSvc.getOrNilErr = sentinel
	doc := ImportDocument{Version: "1", Commands: []ImportCommand{{Alias: "a", Template: "echo hi"}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, false, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileGetVariableOrNilError(t *testing.T) {
	h := newImportServiceHarness()
	sentinel := errors.New("boom")
	h.varSvc.getOrNilErr = sentinel
	doc := ImportDocument{Version: "1", Variables: []ImportVariable{{Name: "v", Value: "1"}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, false, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

// --- ImportFile: classification ---

func TestImportFileClassifiesNewCommandsAndVariablesAsCreated(t *testing.T) {
	h := newImportServiceHarness()
	doc := ImportDocument{
		Version:   "1",
		Commands:  []ImportCommand{{Alias: "a", Template: "echo hi"}},
		Variables: []ImportVariable{{Name: "v", Value: "1"}},
	}
	path := writeImportFile(t, doc)

	result, err := h.svc.ImportFile(path, false, false, strPtr("default"))
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if len(result.CommandsCreated) != 1 || result.CommandsCreated[0] != "a" {
		t.Fatalf("CommandsCreated = %v, want [a]", result.CommandsCreated)
	}
	if len(result.VariablesCreated) != 1 || result.VariablesCreated[0] != "v" {
		t.Fatalf("VariablesCreated = %v, want [v]", result.VariablesCreated)
	}
	if len(h.cmdSvc.createCalls) != 1 {
		t.Fatalf("CreateCommand called %d times, want 1", len(h.cmdSvc.createCalls))
	}
	if len(h.varSvc.createCalls) != 1 {
		t.Fatalf("CreateVariable called %d times, want 1", len(h.varSvc.createCalls))
	}
}

func TestImportFileClassifiesExistingWithoutOverwriteAsSkipped(t *testing.T) {
	h := newImportServiceHarness()
	h.cmdSvc.seed("a", strPtr("default"), nil)
	h.varSvc.seed("v", strPtr("default"), nil)
	doc := ImportDocument{
		Version:   "1",
		Commands:  []ImportCommand{{Alias: "a", Template: "echo hi"}},
		Variables: []ImportVariable{{Name: "v", Value: "1"}},
	}
	path := writeImportFile(t, doc)

	result, err := h.svc.ImportFile(path, false, false, strPtr("default"))
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if len(result.CommandsSkipped) != 1 || result.CommandsSkipped[0] != "a" {
		t.Fatalf("CommandsSkipped = %v, want [a]", result.CommandsSkipped)
	}
	if len(result.VariablesSkipped) != 1 || result.VariablesSkipped[0] != "v" {
		t.Fatalf("VariablesSkipped = %v, want [v]", result.VariablesSkipped)
	}
	if len(h.cmdSvc.createCalls) != 0 || len(h.cmdSvc.updateCalls) != 0 {
		t.Fatalf("skipped command should not be created or updated: create=%d update=%d", len(h.cmdSvc.createCalls), len(h.cmdSvc.updateCalls))
	}
	if len(h.varSvc.createCalls) != 0 || len(h.varSvc.updateCalls) != 0 {
		t.Fatalf("skipped variable should not be created or updated: create=%d update=%d", len(h.varSvc.createCalls), len(h.varSvc.updateCalls))
	}
}

func TestImportFileClassifiesExistingWithOverwriteAsOverwritten(t *testing.T) {
	h := newImportServiceHarness()
	h.cmdSvc.seed("a", strPtr("default"), nil)
	h.varSvc.seed("v", strPtr("default"), nil)
	doc := ImportDocument{
		Version:   "1",
		Commands:  []ImportCommand{{Alias: "a", Template: "echo hi"}},
		Variables: []ImportVariable{{Name: "v", Value: "1"}},
	}
	path := writeImportFile(t, doc)

	result, err := h.svc.ImportFile(path, true, false, strPtr("default"))
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if len(result.CommandsOverwritten) != 1 || result.CommandsOverwritten[0] != "a" {
		t.Fatalf("CommandsOverwritten = %v, want [a]", result.CommandsOverwritten)
	}
	if len(result.VariablesOverwritten) != 1 || result.VariablesOverwritten[0] != "v" {
		t.Fatalf("VariablesOverwritten = %v, want [v]", result.VariablesOverwritten)
	}
	if len(h.cmdSvc.updateCalls) != 1 {
		t.Fatalf("UpdateCommand called %d times, want 1", len(h.cmdSvc.updateCalls))
	}
	if len(h.varSvc.updateCalls) != 1 {
		t.Fatalf("UpdateVariable called %d times, want 1", len(h.varSvc.updateCalls))
	}
}

// --- ImportFile: preview mode ---

func TestImportFilePreviewDoesNotMutate(t *testing.T) {
	h := newImportServiceHarness()
	doc := ImportDocument{
		Version:  "1",
		Commands: []ImportCommand{{Alias: "a", Template: "echo hi", Tags: []string{"t1"}}},
	}
	path := writeImportFile(t, doc)

	result, err := h.svc.ImportFile(path, false, true, strPtr("default"))
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if !result.Preview {
		t.Fatal("result.Preview = false, want true")
	}
	if len(result.CommandsCreated) != 1 {
		t.Fatalf("CommandsCreated = %v, want classification to still happen in preview", result.CommandsCreated)
	}
	if len(h.cmdSvc.createCalls) != 0 {
		t.Fatalf("CreateCommand called %d times in preview, want 0", len(h.cmdSvc.createCalls))
	}
	if len(h.tagSvc.createCalls) != 0 {
		t.Fatalf("CreateTag called %d times in preview, want 0", len(h.tagSvc.createCalls))
	}
}

func TestImportFileSetsResultProfileAndPreview(t *testing.T) {
	h := newImportServiceHarness()
	path := writeImportFile(t, minimalDoc())

	result, err := h.svc.ImportFile(path, false, false, strPtr("custom-profile"))
	if err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if result.Profile != "custom-profile" {
		t.Fatalf("result.Profile = %q, want %q", result.Profile, "custom-profile")
	}
	if result.Preview {
		t.Fatal("result.Preview = true, want false")
	}
}

// --- ImportFile: tag handling ---

func TestImportFileCreatesMissingTagsButNotExistingOnes(t *testing.T) {
	h := newImportServiceHarness()
	h.tagSvc.existing["already-there"] = &models.Tag{Name: "already-there"}
	doc := ImportDocument{
		Version:  "1",
		Commands: []ImportCommand{{Alias: "a", Template: "echo hi", Tags: []string{"already-there", "new-tag"}}},
	}
	path := writeImportFile(t, doc)

	if _, err := h.svc.ImportFile(path, false, false, strPtr("default")); err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if len(h.tagSvc.createCalls) != 1 || h.tagSvc.createCalls[0] != "new-tag" {
		t.Fatalf("CreateTag calls = %v, want [new-tag]", h.tagSvc.createCalls)
	}
}

func TestImportFileSkipsTagsFromSkippedItems(t *testing.T) {
	h := newImportServiceHarness()
	h.cmdSvc.seed("a", strPtr("default"), nil)
	doc := ImportDocument{
		Version:  "1",
		Commands: []ImportCommand{{Alias: "a", Template: "echo hi", Tags: []string{"only-on-skipped"}}},
	}
	path := writeImportFile(t, doc)

	if _, err := h.svc.ImportFile(path, false, false, strPtr("default")); err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}
	if len(h.tagSvc.createCalls) != 0 {
		t.Fatalf("CreateTag calls = %v, want none for a skipped command's tags", h.tagSvc.createCalls)
	}
}

func TestImportFileGetTagError(t *testing.T) {
	h := newImportServiceHarness()
	sentinel := errors.New("boom")
	h.tagSvc.getOrNilErr = sentinel
	doc := ImportDocument{Version: "1", Commands: []ImportCommand{{Alias: "a", Template: "echo hi", Tags: []string{"t"}}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, false, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileCreateTagError(t *testing.T) {
	h := newImportServiceHarness()
	sentinel := errors.New("boom")
	h.tagSvc.createErr = sentinel
	doc := ImportDocument{Version: "1", Commands: []ImportCommand{{Alias: "a", Template: "echo hi", Tags: []string{"t"}}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, false, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

// --- ImportFile: command create/overwrite error propagation ---

func TestImportFileCreateCommandError(t *testing.T) {
	h := newImportServiceHarness()
	sentinel := errors.New("boom")
	h.cmdSvc.createErr = sentinel
	doc := ImportDocument{Version: "1", Commands: []ImportCommand{{Alias: "a", Template: "echo hi"}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, false, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileCreateVariableError(t *testing.T) {
	h := newImportServiceHarness()
	sentinel := errors.New("boom")
	h.varSvc.createErr = sentinel
	doc := ImportDocument{Version: "1", Variables: []ImportVariable{{Name: "v", Value: "1"}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, false, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileOverwriteCommandGetWithTagsError(t *testing.T) {
	h := newImportServiceHarness()
	h.cmdSvc.seed("a", strPtr("default"), nil)
	sentinel := errors.New("boom")
	h.cmdSvc.getWithTagsErr = sentinel
	doc := ImportDocument{Version: "1", Commands: []ImportCommand{{Alias: "a", Template: "echo hi"}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, true, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileOverwriteCommandUpdateError(t *testing.T) {
	h := newImportServiceHarness()
	h.cmdSvc.seed("a", strPtr("default"), nil)
	sentinel := errors.New("boom")
	h.cmdSvc.updateErr = sentinel
	doc := ImportDocument{Version: "1", Commands: []ImportCommand{{Alias: "a", Template: "echo hi"}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, true, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileOverwriteCommandRemoveTagsError(t *testing.T) {
	h := newImportServiceHarness()
	h.cmdSvc.seed("a", strPtr("default"), []models.Tag{{Name: "old"}})
	sentinel := errors.New("boom")
	h.cmdSvc.removeTagsErr = sentinel
	doc := ImportDocument{Version: "1", Commands: []ImportCommand{{Alias: "a", Template: "echo hi"}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, true, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileOverwriteCommandAddTagsError(t *testing.T) {
	h := newImportServiceHarness()
	h.cmdSvc.seed("a", strPtr("default"), nil)
	sentinel := errors.New("boom")
	h.cmdSvc.addTagsErr = sentinel
	doc := ImportDocument{Version: "1", Commands: []ImportCommand{{Alias: "a", Template: "echo hi", Tags: []string{"new"}}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, true, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileOverwriteCommandTagDiff(t *testing.T) {
	h := newImportServiceHarness()
	h.cmdSvc.seed("a", strPtr("default"), []models.Tag{{Name: "keep"}, {Name: "drop"}})
	doc := ImportDocument{
		Version:  "1",
		Commands: []ImportCommand{{Alias: "a", Template: "echo hi", Tags: []string{"keep", "add"}}},
	}
	path := writeImportFile(t, doc)

	if _, err := h.svc.ImportFile(path, true, false, strPtr("default")); err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}

	removed := sortedCopy(h.cmdSvc.removeTagsCalls["a"])
	if len(removed) != 1 || removed[0] != "drop" {
		t.Fatalf("RemoveTags(a) = %v, want [drop]", removed)
	}
	added := sortedCopy(h.cmdSvc.addTagsCalls["a"])
	if len(added) != 1 || added[0] != "add" {
		t.Fatalf("AddTags(a) = %v, want [add]", added)
	}
}

// --- ImportFile: variable overwrite error propagation ---

func TestImportFileOverwriteVariableGetWithTagsError(t *testing.T) {
	h := newImportServiceHarness()
	h.varSvc.seed("v", strPtr("default"), nil)
	sentinel := errors.New("boom")
	h.varSvc.getWithTagsErr = sentinel
	doc := ImportDocument{Version: "1", Variables: []ImportVariable{{Name: "v", Value: "1"}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, true, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileOverwriteVariableUpdateError(t *testing.T) {
	h := newImportServiceHarness()
	h.varSvc.seed("v", strPtr("default"), nil)
	sentinel := errors.New("boom")
	h.varSvc.updateErr = sentinel
	doc := ImportDocument{Version: "1", Variables: []ImportVariable{{Name: "v", Value: "1"}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, true, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileOverwriteVariableRemoveTagsError(t *testing.T) {
	h := newImportServiceHarness()
	h.varSvc.seed("v", strPtr("default"), []models.Tag{{Name: "old"}})
	sentinel := errors.New("boom")
	h.varSvc.removeTagsErr = sentinel
	doc := ImportDocument{Version: "1", Variables: []ImportVariable{{Name: "v", Value: "1"}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, true, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileOverwriteVariableAddTagsError(t *testing.T) {
	h := newImportServiceHarness()
	h.varSvc.seed("v", strPtr("default"), nil)
	sentinel := errors.New("boom")
	h.varSvc.addTagsErr = sentinel
	doc := ImportDocument{Version: "1", Variables: []ImportVariable{{Name: "v", Value: "1", Tags: []string{"new"}}}}
	path := writeImportFile(t, doc)

	_, err := h.svc.ImportFile(path, true, false, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ImportFile() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestImportFileOverwriteVariableTagDiff(t *testing.T) {
	h := newImportServiceHarness()
	h.varSvc.seed("v", strPtr("default"), []models.Tag{{Name: "keep"}, {Name: "drop"}})
	doc := ImportDocument{
		Version:   "1",
		Variables: []ImportVariable{{Name: "v", Value: "1", Tags: []string{"keep", "add"}}},
	}
	path := writeImportFile(t, doc)

	if _, err := h.svc.ImportFile(path, true, false, strPtr("default")); err != nil {
		t.Fatalf("ImportFile() error = %v", err)
	}

	removed := sortedCopy(h.varSvc.removeTagsCalls["v"])
	if len(removed) != 1 || removed[0] != "drop" {
		t.Fatalf("RemoveTags(v) = %v, want [drop]", removed)
	}
	added := sortedCopy(h.varSvc.addTagsCalls["v"])
	if len(added) != 1 || added[0] != "add" {
		t.Fatalf("AddTags(v) = %v, want [add]", added)
	}
}
