package services

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
)

// --- test harness ---

type exportServiceHarness struct {
	svc         *ExportService
	cmdSvc      *fakeCommandService
	varSvc      *fakeVariableService
	profileRepo *fakeProfileRepo
}

func newExportServiceHarness() *exportServiceHarness {
	cmdSvc := newFakeCommandService()
	varSvc := newFakeVariableService()
	profileRepo := &fakeProfileRepo{
		activeProfile:         &models.Profile{Name: "cmd-active"},
		activeVariableProfile: &models.Profile{Name: "var-active"},
	}

	svc := NewExportService(cmdSvc, varSvc, profileRepo)

	return &exportServiceHarness{
		svc:         svc,
		cmdSvc:      cmdSvc,
		varSvc:      varSvc,
		profileRepo: profileRepo,
	}
}

func seedCommandFull(cs *fakeCommandService, alias string, profileName *string, template string, tags []models.Tag) {
	cs.seed(alias, profileName, tags)
	cs.commands[cmdKey(alias, profileName)].Template = template
}

func seedVariableFull(vs *fakeVariableService, name string, profileName *string, value string, tags []models.Tag) {
	vs.seed(name, profileName, tags)
	vs.variables[cmdKey(name, profileName)].Value = value
}

func readExportDoc(t *testing.T, path string) TransferDocument {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading export file: %v", err)
	}
	var doc TransferDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("unmarshal export file: %v", err)
	}
	return doc
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

// --- resolveOutputPath ---

func TestResolveOutputPathNilUsesCwd(t *testing.T) {
	path, err := resolveOutputPath(nil, "cmds")
	if err != nil {
		t.Fatalf("resolveOutputPath() error = %v", err)
	}
	cwd, _ := os.Getwd()
	wantPrefix := filepath.Join(cwd, "cmdbox-cmds-"+time.Now().Format("2006-01-02"))
	if filepath.Dir(path) != cwd {
		t.Fatalf("resolveOutputPath() = %q, want directory %q", path, cwd)
	}
	if !hasSuffixJSON(path) || filepath.Base(path) != filepath.Base(wantPrefix+".json") {
		t.Fatalf("resolveOutputPath() = %q, want basename %q", path, filepath.Base(wantPrefix+".json"))
	}
}

func hasSuffixJSON(path string) bool {
	return filepath.Ext(path) == ".json"
}

func TestResolveOutputPathExplicitFile(t *testing.T) {
	explicit := filepath.Join(t.TempDir(), "out.json")
	path, err := resolveOutputPath(&explicit, "vars")
	if err != nil {
		t.Fatalf("resolveOutputPath() error = %v", err)
	}
	if path != explicit {
		t.Fatalf("resolveOutputPath() = %q, want %q", path, explicit)
	}
}

func TestResolveOutputPathExplicitDirectory(t *testing.T) {
	dir := t.TempDir()
	path, err := resolveOutputPath(&dir, "all")
	if err != nil {
		t.Fatalf("resolveOutputPath() error = %v", err)
	}
	if filepath.Dir(path) != dir {
		t.Fatalf("resolveOutputPath() = %q, want directory %q", path, dir)
	}
	if filepath.Base(path) != "cmdbox-all-"+time.Now().Format("2006-01-02")+".json" {
		t.Fatalf("resolveOutputPath() basename = %q, want date-stamped filename", filepath.Base(path))
	}
}

// --- buildDocument ---

func TestBuildDocumentSetsFields(t *testing.T) {
	cmds := []TransferCommand{{Alias: "a"}}
	vars := []TransferVariable{{Name: "v"}}

	doc := buildDocument("cmds", cmds, vars, "cmd-profile", "var-profile")

	if doc.Version != "1" {
		t.Errorf("Version = %q, want %q", doc.Version, "1")
	}
	if doc.Type != "cmds" {
		t.Errorf("Type = %q, want %q", doc.Type, "cmds")
	}
	if doc.CommandProfile != "cmd-profile" || doc.VariableProfile != "var-profile" {
		t.Errorf("profiles = %q/%q, want cmd-profile/var-profile", doc.CommandProfile, doc.VariableProfile)
	}
	if len(doc.Commands) != 1 || doc.Commands[0].Alias != "a" {
		t.Errorf("Commands = %v, want [{Alias: a}]", doc.Commands)
	}
	if len(doc.Variables) != 1 || doc.Variables[0].Name != "v" {
		t.Errorf("Variables = %v, want [{Name: v}]", doc.Variables)
	}
	if _, err := time.Parse(time.RFC3339, doc.ExportedAt); err != nil {
		t.Errorf("ExportedAt = %q, not a valid RFC3339 timestamp: %v", doc.ExportedAt, err)
	}
}

// --- tagFilter ---

func TestTagFilterNilTag(t *testing.T) {
	if got := tagFilter(nil); got != nil {
		t.Fatalf("tagFilter(nil) = %v, want nil", got)
	}
}

func TestTagFilterWithTag(t *testing.T) {
	tag := "important"
	got := tagFilter(&tag)
	if len(got) != 1 || got[0] != "important" {
		t.Fatalf("tagFilter(&%q) = %v, want [%q]", tag, got, tag)
	}
}

// --- collectDeepCommands ---

func TestCollectDeepCommandsFollowsReferences(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "top", profile, "run <cmd:dep>", nil)
	seedCommandFull(h.cmdSvc, "dep", profile, "echo dep", nil)

	collected, order, err := collectDeepCommands([]string{"top"}, h.cmdSvc, profile)
	if err != nil {
		t.Fatalf("collectDeepCommands() error = %v", err)
	}
	if len(collected) != 2 {
		t.Fatalf("collected = %v, want 2 entries", collected)
	}
	if len(order) != 2 || order[0] != "top" || order[1] != "dep" {
		t.Fatalf("order = %v, want [top dep]", order)
	}
}

func TestCollectDeepCommandsSkipsMissingReference(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "top", profile, "run <cmd:ghost>", nil)

	collected, order, err := collectDeepCommands([]string{"top"}, h.cmdSvc, profile)
	if err != nil {
		t.Fatalf("collectDeepCommands() error = %v", err)
	}
	if _, ok := collected["ghost"]; ok {
		t.Fatalf("collected contains ghost, want it skipped silently")
	}
	if len(order) != 1 || order[0] != "top" {
		t.Fatalf("order = %v, want [top]", order)
	}
}

func TestCollectDeepCommandsHandlesCycle(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "a", profile, "<cmd:b>", nil)
	seedCommandFull(h.cmdSvc, "b", profile, "<cmd:a>", nil)

	collected, order, err := collectDeepCommands([]string{"a"}, h.cmdSvc, profile)
	if err != nil {
		t.Fatalf("collectDeepCommands() error = %v", err)
	}
	if len(collected) != 2 || len(order) != 2 {
		t.Fatalf("collected/order = %v/%v, want 2 entries each with no infinite loop", collected, order)
	}
}

func TestCollectDeepCommandsPropagatesError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.cmdSvc.getOrNilErr = sentinel

	_, _, err := collectDeepCommands([]string{"a"}, h.cmdSvc, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("collectDeepCommands() error = %v, want wrapping %v", err, sentinel)
	}
}

// --- collectDeepVariables ---

func TestCollectDeepVariablesFromNamesAndCommandReferences(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedVariableFull(h.varSvc, "direct", profile, "1", nil)
	seedVariableFull(h.varSvc, "fromcmd", profile, "2", nil)
	commands := map[string]*models.Command{
		"top": {Alias: "top", Template: "<var:fromcmd>"},
	}

	collected, order, err := collectDeepVariables([]string{"direct"}, commands, h.varSvc, profile)
	if err != nil {
		t.Fatalf("collectDeepVariables() error = %v", err)
	}
	if len(collected) != 2 {
		t.Fatalf("collected = %v, want 2 entries", collected)
	}
	if !contains(order, "direct") || !contains(order, "fromcmd") {
		t.Fatalf("order = %v, want to contain direct and fromcmd", order)
	}
}

func TestCollectDeepVariablesSkipsMissing(t *testing.T) {
	h := newExportServiceHarness()
	collected, order, err := collectDeepVariables([]string{"ghost"}, nil, h.varSvc, strPtr("default"))
	if err != nil {
		t.Fatalf("collectDeepVariables() error = %v", err)
	}
	if len(collected) != 0 || len(order) != 0 {
		t.Fatalf("collected/order = %v/%v, want empty", collected, order)
	}
}

func TestCollectDeepVariablesHandlesCycle(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedVariableFull(h.varSvc, "x", profile, "<var:y>", nil)
	seedVariableFull(h.varSvc, "y", profile, "<var:x>", nil)

	collected, order, err := collectDeepVariables([]string{"x"}, nil, h.varSvc, profile)
	if err != nil {
		t.Fatalf("collectDeepVariables() error = %v", err)
	}
	if len(collected) != 2 || len(order) != 2 {
		t.Fatalf("collected/order = %v/%v, want 2 entries each with no infinite loop", collected, order)
	}
}

func TestCollectDeepVariablesPropagatesError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.varSvc.getOrNilErr = sentinel

	_, _, err := collectDeepVariables([]string{"x"}, nil, h.varSvc, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("collectDeepVariables() error = %v, want wrapping %v", err, sentinel)
	}
}

// --- flattenTemplate ---

func TestFlattenTemplateResolvesVariableAndCommandRefs(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "dep", profile, "echo <var:name>", nil)
	seedVariableFull(h.varSvc, "name", profile, "world", nil)

	got, err := flattenTemplate("run: <cmd:dep>", h.cmdSvc, h.varSvc, profile, profile, nil)
	if err != nil {
		t.Fatalf("flattenTemplate() error = %v", err)
	}
	if got != "run: echo world" {
		t.Fatalf("flattenTemplate() = %q, want %q", got, "run: echo world")
	}
}

func TestFlattenTemplateHandlesEscapedBackslash(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedVariableFull(h.varSvc, "x", profile, "resolved", nil)

	template := `pre\<var:x>post`
	got, err := flattenTemplate(template, h.cmdSvc, h.varSvc, profile, profile, nil)
	if err != nil {
		t.Fatalf("flattenTemplate() error = %v", err)
	}
	if got != template {
		t.Fatalf("flattenTemplate() = %q, want unchanged %q", got, template)
	}
}

func TestFlattenTemplateLeavesUnterminatedTokenAsIs(t *testing.T) {
	h := newExportServiceHarness()
	template := "abc<def"
	got, err := flattenTemplate(template, h.cmdSvc, h.varSvc, nil, nil, nil)
	if err != nil {
		t.Fatalf("flattenTemplate() error = %v", err)
	}
	if got != template {
		t.Fatalf("flattenTemplate() = %q, want unchanged %q", got, template)
	}
}

func TestFlattenTemplateLeavesMissingReferenceAsRawToken(t *testing.T) {
	h := newExportServiceHarness()
	got, err := flattenTemplate("<var:missing>", h.cmdSvc, h.varSvc, nil, nil, nil)
	if err != nil {
		t.Fatalf("flattenTemplate() error = %v", err)
	}
	if got != "<var:missing>" {
		t.Fatalf("flattenTemplate() = %q, want %q", got, "<var:missing>")
	}
}

func TestFlattenTemplateBreaksSelfReferentialCycle(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "a", profile, "<cmd:a>", nil)

	got, err := flattenTemplate("<cmd:a>", h.cmdSvc, h.varSvc, profile, profile, nil)
	if err != nil {
		t.Fatalf("flattenTemplate() error = %v", err)
	}
	if got != "<cmd:a>" {
		t.Fatalf("flattenTemplate() = %q, want %q (cycle should stop resolving)", got, "<cmd:a>")
	}
}

func TestFlattenTemplatePropagatesCommandError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.cmdSvc.getOrNilErr = sentinel

	_, err := flattenTemplate("<cmd:a>", h.cmdSvc, h.varSvc, nil, nil, nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("flattenTemplate() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestFlattenTemplatePropagatesVariableError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.varSvc.getOrNilErr = sentinel

	_, err := flattenTemplate("<var:x>", h.cmdSvc, h.varSvc, nil, nil, nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("flattenTemplate() error = %v, want wrapping %v", err, sentinel)
	}
}

// --- serializeCommand / serializeVariable ---

func TestSerializeCommandNoFlatten(t *testing.T) {
	h := newExportServiceHarness()
	desc := "a description"
	cmd := &models.Command{Alias: "a", Template: "<cmd:dep>", Description: &desc}

	got, err := serializeCommand(cmd, []string{"t1"}, false, h.cmdSvc, h.varSvc, nil, nil)
	if err != nil {
		t.Fatalf("serializeCommand() error = %v", err)
	}
	if got.Template != "<cmd:dep>" {
		t.Fatalf("Template = %q, want unresolved %q", got.Template, "<cmd:dep>")
	}
	if got.Description == nil || *got.Description != desc {
		t.Fatalf("Description = %v, want %q", got.Description, desc)
	}
	if len(got.Tags) != 1 || got.Tags[0] != "t1" {
		t.Fatalf("Tags = %v, want [t1]", got.Tags)
	}
}

func TestSerializeCommandFlattenResolvesNestedRefs(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "dep", profile, "echo dep", nil)
	cmd := &models.Command{Alias: "top", Template: "<cmd:dep>"}

	got, err := serializeCommand(cmd, nil, true, h.cmdSvc, h.varSvc, profile, profile)
	if err != nil {
		t.Fatalf("serializeCommand() error = %v", err)
	}
	if got.Template != "echo dep" {
		t.Fatalf("Template = %q, want %q", got.Template, "echo dep")
	}
}

func TestSerializeCommandInvalidEnvJSONErrors(t *testing.T) {
	h := newExportServiceHarness()
	badEnv := "{not json"
	cmd := &models.Command{Alias: "a", Template: "echo hi", Env: &badEnv}

	_, err := serializeCommand(cmd, nil, false, h.cmdSvc, h.varSvc, nil, nil)
	if err == nil {
		t.Fatal("serializeCommand() error = nil, want error decoding env")
	}
}

func TestSerializeVariableNoFlatten(t *testing.T) {
	h := newExportServiceHarness()
	v := &models.Variable{Name: "x", Value: "<var:y>"}

	got, err := serializeVariable(v, []string{"t1"}, false, h.cmdSvc, h.varSvc, nil, nil)
	if err != nil {
		t.Fatalf("serializeVariable() error = %v", err)
	}
	if got.Value != "<var:y>" {
		t.Fatalf("Value = %q, want unresolved %q", got.Value, "<var:y>")
	}
	if len(got.Tags) != 1 || got.Tags[0] != "t1" {
		t.Fatalf("Tags = %v, want [t1]", got.Tags)
	}
}

func TestSerializeVariableFlatten(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedVariableFull(h.varSvc, "y", profile, "resolved", nil)
	v := &models.Variable{Name: "x", Value: "<var:y>"}

	got, err := serializeVariable(v, nil, true, h.cmdSvc, h.varSvc, profile, profile)
	if err != nil {
		t.Fatalf("serializeVariable() error = %v", err)
	}
	if got.Value != "resolved" {
		t.Fatalf("Value = %q, want %q", got.Value, "resolved")
	}
}

// --- resolveCommandProfileName / resolveVariableProfileName ---

func TestResolveCommandProfileNameExplicit(t *testing.T) {
	h := newExportServiceHarness()
	h.profileRepo.activeErr = errors.New("should not be called")

	name, err := h.svc.resolveCommandProfileName(strPtr("work"))
	if err != nil {
		t.Fatalf("resolveCommandProfileName() error = %v", err)
	}
	if name != "work" {
		t.Fatalf("resolveCommandProfileName() = %q, want %q", name, "work")
	}
}

func TestResolveCommandProfileNameFallsBackToActive(t *testing.T) {
	h := newExportServiceHarness()
	name, err := h.svc.resolveCommandProfileName(nil)
	if err != nil {
		t.Fatalf("resolveCommandProfileName() error = %v", err)
	}
	if name != "cmd-active" {
		t.Fatalf("resolveCommandProfileName() = %q, want %q", name, "cmd-active")
	}
}

func TestResolveCommandProfileNameActiveError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.profileRepo.activeErr = sentinel

	_, err := h.svc.resolveCommandProfileName(nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("resolveCommandProfileName() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestResolveVariableProfileNameExplicit(t *testing.T) {
	h := newExportServiceHarness()
	h.profileRepo.activeVariableErr = errors.New("should not be called")

	name, err := h.svc.resolveVariableProfileName(strPtr("work"))
	if err != nil {
		t.Fatalf("resolveVariableProfileName() error = %v", err)
	}
	if name != "work" {
		t.Fatalf("resolveVariableProfileName() = %q, want %q", name, "work")
	}
}

func TestResolveVariableProfileNameFallsBackToActive(t *testing.T) {
	h := newExportServiceHarness()
	name, err := h.svc.resolveVariableProfileName(nil)
	if err != nil {
		t.Fatalf("resolveVariableProfileName() error = %v", err)
	}
	if name != "var-active" {
		t.Fatalf("resolveVariableProfileName() = %q, want %q", name, "var-active")
	}
}

func TestResolveVariableProfileNameActiveError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.profileRepo.activeVariableErr = sentinel

	_, err := h.svc.resolveVariableProfileName(nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("resolveVariableProfileName() error = %v, want wrapping %v", err, sentinel)
	}
}

// --- commandTagNames / variableTagNames ---

func TestCommandTagNamesMapsNames(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	h.cmdSvc.seed("a", profile, []models.Tag{{Name: "t1"}, {Name: "t2"}})

	got, err := h.svc.commandTagNames("a", profile)
	if err != nil {
		t.Fatalf("commandTagNames() error = %v", err)
	}
	sort.Strings(got)
	if len(got) != 2 || got[0] != "t1" || got[1] != "t2" {
		t.Fatalf("commandTagNames() = %v, want [t1 t2]", got)
	}
}

func TestCommandTagNamesPropagatesError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.cmdSvc.getWithTagsErr = sentinel

	_, err := h.svc.commandTagNames("a", strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("commandTagNames() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestVariableTagNamesMapsNames(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	h.varSvc.seed("v", profile, []models.Tag{{Name: "t1"}})

	got, err := h.svc.variableTagNames("v", profile)
	if err != nil {
		t.Fatalf("variableTagNames() error = %v", err)
	}
	if len(got) != 1 || got[0] != "t1" {
		t.Fatalf("variableTagNames() = %v, want [t1]", got)
	}
}

func TestVariableTagNamesPropagatesError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.varSvc.getWithTagsErr = sentinel

	_, err := h.svc.variableTagNames("v", strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("variableTagNames() error = %v, want wrapping %v", err, sentinel)
	}
}

// --- writeExportDoc ---

func TestWriteExportDocWritesIndentedJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out.json")
	doc := TransferDocument{Version: "1", Type: "cmds"}

	if err := writeExportDoc(path, doc); err != nil {
		t.Fatalf("writeExportDoc() error = %v", err)
	}
	got := readExportDoc(t, path)
	if got.Version != "1" || got.Type != "cmds" {
		t.Fatalf("readback doc = %+v, want version=1 type=cmds", got)
	}
}

func TestWriteExportDocPropagatesAtomicFileError(t *testing.T) {
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("write blocker file: %v", err)
	}
	path := filepath.Join(blocker, "out.json")

	err := writeExportDoc(path, TransferDocument{})
	if err == nil {
		t.Fatal("writeExportDoc() error = nil, want error")
	}
}

// --- ExportCmds ---

func TestExportCmdsExplicitAliasesFlattenWritesFileAndResult(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "dep", profile, "echo dep", nil)
	seedCommandFull(h.cmdSvc, "top", profile, "<cmd:dep>", []models.Tag{{Name: "t1"}})
	out := filepath.Join(t.TempDir(), "out.json")

	result, err := h.svc.ExportCmds([]string{"top"}, nil, true, &out, profile, profile)
	if err != nil {
		t.Fatalf("ExportCmds() error = %v", err)
	}
	if len(result.Commands) != 1 || result.Commands[0] != "top" {
		t.Fatalf("Commands = %v, want [top]", result.Commands)
	}
	if len(result.TransientCommands) != 0 {
		t.Fatalf("TransientCommands = %v, want none when flattening", result.TransientCommands)
	}

	doc := readExportDoc(t, out)
	if len(doc.Commands) != 1 || doc.Commands[0].Template != "echo dep" {
		t.Fatalf("exported command = %+v, want flattened template %q", doc.Commands, "echo dep")
	}
}

func TestExportCmdsExplicitAliasesFlattenWarnsOnMissing(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	out := filepath.Join(t.TempDir(), "out.json")

	result, err := h.svc.ExportCmds([]string{"ghost"}, nil, true, &out, profile, profile)
	if err != nil {
		t.Fatalf("ExportCmds() error = %v", err)
	}
	if len(result.Warnings) != 1 || result.Warnings[0] != "Command ghost not found" {
		t.Fatalf("Warnings = %v, want [Command ghost not found]", result.Warnings)
	}
	if len(result.Commands) != 0 {
		t.Fatalf("Commands = %v, want none", result.Commands)
	}
}

func TestExportCmdsNoAliasesListsFromRepositoryWithTagFilter(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("work")
	seedCommandFull(h.cmdSvc, "keepme", profile, "echo keep", []models.Tag{{Name: "keep"}})
	seedCommandFull(h.cmdSvc, "skipme", profile, "echo skip", []models.Tag{{Name: "skip"}})
	out := filepath.Join(t.TempDir(), "out.json")
	tag := "keep"

	result, err := h.svc.ExportCmds(nil, &tag, false, &out, profile, profile)
	if err != nil {
		t.Fatalf("ExportCmds() error = %v", err)
	}
	if len(result.Commands) != 1 || result.Commands[0] != "keepme" {
		t.Fatalf("Commands = %v, want [keepme]", result.Commands)
	}
}

func TestExportCmdsDeepCollectSeparatesDirectAndTransient(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "top", profile, "run <cmd:dep> <var:v1>", nil)
	seedCommandFull(h.cmdSvc, "dep", profile, "echo dep", nil)
	seedVariableFull(h.varSvc, "v1", profile, "1", nil)
	out := filepath.Join(t.TempDir(), "out.json")

	result, err := h.svc.ExportCmds([]string{"top"}, nil, false, &out, profile, profile)
	if err != nil {
		t.Fatalf("ExportCmds() error = %v", err)
	}
	if len(result.Commands) != 1 || result.Commands[0] != "top" {
		t.Fatalf("Commands = %v, want [top]", result.Commands)
	}
	if len(result.TransientCommands) != 1 || result.TransientCommands[0] != "dep" {
		t.Fatalf("TransientCommands = %v, want [dep]", result.TransientCommands)
	}
	if len(result.TransientVariables) != 1 || result.TransientVariables[0] != "v1" {
		t.Fatalf("TransientVariables = %v, want [v1]", result.TransientVariables)
	}
}

func TestExportCmdsDeepCollectWarnsOnMissingDirectAlias(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	out := filepath.Join(t.TempDir(), "out.json")

	result, err := h.svc.ExportCmds([]string{"ghost"}, nil, false, &out, profile, profile)
	if err != nil {
		t.Fatalf("ExportCmds() error = %v", err)
	}
	if len(result.Warnings) != 1 || result.Warnings[0] != "Command ghost not found" {
		t.Fatalf("Warnings = %v, want [Command ghost not found]", result.Warnings)
	}
}

func TestExportCmdsCommandProfileResolutionError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.profileRepo.activeErr = sentinel

	_, err := h.svc.ExportCmds(nil, nil, false, nil, nil, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportCmds() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportCmdsVariableProfileResolutionError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.profileRepo.activeVariableErr = sentinel

	_, err := h.svc.ExportCmds(nil, nil, false, nil, strPtr("default"), nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportCmds() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportCmdsListCommandsError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.cmdSvc.listErr = sentinel

	_, err := h.svc.ExportCmds(nil, nil, false, nil, strPtr("default"), strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportCmds() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportCmdsFlattenGetCommandOrNilError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.cmdSvc.getOrNilErr = sentinel

	_, err := h.svc.ExportCmds([]string{"a"}, nil, true, nil, strPtr("default"), strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportCmds() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportCmdsCollectDeepCommandsError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.cmdSvc.getOrNilErr = sentinel

	_, err := h.svc.ExportCmds([]string{"a"}, nil, false, nil, strPtr("default"), strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportCmds() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportCmdsCommandTagNamesError(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "a", profile, "echo hi", nil)
	sentinel := errors.New("boom")
	h.cmdSvc.getWithTagsErr = sentinel

	_, err := h.svc.ExportCmds([]string{"a"}, nil, true, nil, profile, profile)
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportCmds() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportCmdsWriteError(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "a", profile, "echo hi", nil)
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("write blocker file: %v", err)
	}
	out := filepath.Join(blocker, "out.json")

	_, err := h.svc.ExportCmds([]string{"a"}, nil, true, &out, profile, profile)
	if err == nil {
		t.Fatal("ExportCmds() error = nil, want error")
	}
}

// --- ExportVars ---

func TestExportVarsExplicitNamesFlattenWritesFile(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedVariableFull(h.varSvc, "dep", profile, "resolved", nil)
	seedVariableFull(h.varSvc, "top", profile, "<var:dep>", []models.Tag{{Name: "t1"}})
	out := filepath.Join(t.TempDir(), "out.json")

	result, err := h.svc.ExportVars([]string{"top"}, nil, true, &out, profile)
	if err != nil {
		t.Fatalf("ExportVars() error = %v", err)
	}
	if len(result.Variables) != 1 || result.Variables[0] != "top" {
		t.Fatalf("Variables = %v, want [top]", result.Variables)
	}

	doc := readExportDoc(t, out)
	if len(doc.Variables) != 1 || doc.Variables[0].Value != "resolved" {
		t.Fatalf("exported variable = %+v, want flattened value %q", doc.Variables, "resolved")
	}
}

func TestExportVarsWarnsOnMissingExplicitName(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	out := filepath.Join(t.TempDir(), "out.json")

	result, err := h.svc.ExportVars([]string{"ghost"}, nil, true, &out, profile)
	if err != nil {
		t.Fatalf("ExportVars() error = %v", err)
	}
	if len(result.Warnings) != 1 || result.Warnings[0] != "Variable ghost not found" {
		t.Fatalf("Warnings = %v, want [Variable ghost not found]", result.Warnings)
	}
}

func TestExportVarsNoNamesListsWithTagFilter(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("work")
	seedVariableFull(h.varSvc, "keepme", profile, "1", []models.Tag{{Name: "keep"}})
	seedVariableFull(h.varSvc, "skipme", profile, "2", []models.Tag{{Name: "skip"}})
	out := filepath.Join(t.TempDir(), "out.json")
	tag := "keep"

	result, err := h.svc.ExportVars(nil, &tag, false, &out, profile)
	if err != nil {
		t.Fatalf("ExportVars() error = %v", err)
	}
	if len(result.Variables) != 1 || result.Variables[0] != "keepme" {
		t.Fatalf("Variables = %v, want [keepme]", result.Variables)
	}
}

func TestExportVarsDeepCollectSeparatesDirectAndTransient(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedVariableFull(h.varSvc, "top", profile, "<var:dep>", nil)
	seedVariableFull(h.varSvc, "dep", profile, "1", nil)
	out := filepath.Join(t.TempDir(), "out.json")

	result, err := h.svc.ExportVars([]string{"top"}, nil, false, &out, profile)
	if err != nil {
		t.Fatalf("ExportVars() error = %v", err)
	}
	if len(result.Variables) != 1 || result.Variables[0] != "top" {
		t.Fatalf("Variables = %v, want [top]", result.Variables)
	}
	if len(result.TransientVariables) != 1 || result.TransientVariables[0] != "dep" {
		t.Fatalf("TransientVariables = %v, want [dep]", result.TransientVariables)
	}
}

func TestExportVarsProfileResolutionError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.profileRepo.activeVariableErr = sentinel

	_, err := h.svc.ExportVars(nil, nil, false, nil, nil)
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportVars() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportVarsListVariablesError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.varSvc.listErr = sentinel

	_, err := h.svc.ExportVars(nil, nil, false, nil, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportVars() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportVarsFlattenGetVariableOrNilError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.varSvc.getOrNilErr = sentinel

	_, err := h.svc.ExportVars([]string{"v"}, nil, true, nil, strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportVars() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportVarsVariableTagNamesError(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedVariableFull(h.varSvc, "v", profile, "1", nil)
	sentinel := errors.New("boom")
	h.varSvc.getWithTagsErr = sentinel

	_, err := h.svc.ExportVars([]string{"v"}, nil, true, nil, profile)
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportVars() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportVarsWriteError(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedVariableFull(h.varSvc, "v", profile, "1", nil)
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("write blocker file: %v", err)
	}
	out := filepath.Join(blocker, "out.json")

	_, err := h.svc.ExportVars([]string{"v"}, nil, true, &out, profile)
	if err == nil {
		t.Fatal("ExportVars() error = nil, want error")
	}
}

// --- ExportAll ---

func TestExportAllWritesAllCommandsAndVariables(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "a", profile, "echo a", []models.Tag{{Name: "t1"}})
	seedVariableFull(h.varSvc, "v", profile, "1", nil)
	out := filepath.Join(t.TempDir(), "out.json")

	result, err := h.svc.ExportAll(false, &out, profile, profile)
	if err != nil {
		t.Fatalf("ExportAll() error = %v", err)
	}
	if len(result.Commands) != 1 || result.Commands[0] != "a" {
		t.Fatalf("Commands = %v, want [a]", result.Commands)
	}
	if len(result.Variables) != 1 || result.Variables[0] != "v" {
		t.Fatalf("Variables = %v, want [v]", result.Variables)
	}

	doc := readExportDoc(t, out)
	if doc.Type != "all" {
		t.Fatalf("doc.Type = %q, want %q", doc.Type, "all")
	}
	if len(doc.Commands) != 1 || len(doc.Commands[0].Tags) != 1 || doc.Commands[0].Tags[0] != "t1" {
		t.Fatalf("exported command = %+v, want tag t1", doc.Commands)
	}
}

func TestExportAllFlattenTrue(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedVariableFull(h.varSvc, "name", profile, "world", nil)
	seedCommandFull(h.cmdSvc, "a", profile, "echo <var:name>", nil)
	out := filepath.Join(t.TempDir(), "out.json")

	_, err := h.svc.ExportAll(true, &out, profile, profile)
	if err != nil {
		t.Fatalf("ExportAll() error = %v", err)
	}

	doc := readExportDoc(t, out)
	if len(doc.Commands) != 1 || doc.Commands[0].Template != "echo world" {
		t.Fatalf("exported command = %+v, want flattened template %q", doc.Commands, "echo world")
	}
}

func TestExportAllListCommandsError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.cmdSvc.listErr = sentinel

	_, err := h.svc.ExportAll(false, nil, strPtr("default"), strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportAll() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportAllListVariablesError(t *testing.T) {
	h := newExportServiceHarness()
	sentinel := errors.New("boom")
	h.varSvc.listErr = sentinel

	_, err := h.svc.ExportAll(false, nil, strPtr("default"), strPtr("default"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportAll() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportAllCommandTagError(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedCommandFull(h.cmdSvc, "a", profile, "echo hi", nil)
	sentinel := errors.New("boom")
	h.cmdSvc.getWithTagsErr = sentinel

	_, err := h.svc.ExportAll(false, nil, profile, profile)
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportAll() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportAllVariableTagError(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	seedVariableFull(h.varSvc, "v", profile, "1", nil)
	sentinel := errors.New("boom")
	h.varSvc.getWithTagsErr = sentinel

	_, err := h.svc.ExportAll(false, nil, profile, profile)
	if !errors.Is(err, sentinel) {
		t.Fatalf("ExportAll() error = %v, want wrapping %v", err, sentinel)
	}
}

func TestExportAllWriteError(t *testing.T) {
	h := newExportServiceHarness()
	profile := strPtr("default")
	tmp := t.TempDir()
	blocker := filepath.Join(tmp, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o600); err != nil {
		t.Fatalf("write blocker file: %v", err)
	}
	out := filepath.Join(blocker, "out.json")

	_, err := h.svc.ExportAll(false, &out, profile, profile)
	if err == nil {
		t.Fatal("ExportAll() error = nil, want error")
	}
}
