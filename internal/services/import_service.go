package services

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/PhantomLambSoft/CmdBox/internal/repository"
	"github.com/PhantomLambSoft/CmdBox/internal/resolve"
)

var supportedVersions = []string{"1"}

type ImportResult struct {
	CommandsCreated      []string
	CommandsSkipped      []string
	CommandsOverwritten  []string
	VariablesCreated     []string
	VariablesSkipped     []string
	VariablesOverwritten []string
	Profile              string
	Preview              bool
}

// makeLabel combines a reference kind and key into a single string, separated by a colon.
func makeLabel(kind resolve.RefKind, key string) string {
	return string(kind) + ":" + key
}

func buildDependencyGraph(importData TransferDocument) map[string][]string {
	deps := make(map[string][]string)

	for _, command := range importData.Commands {
		label := makeLabel(resolve.RefKindCommand, command.Alias)
		references := resolve.ExtractReferences(command.Template)
		refLabels := make([]string, len(references))
		for i, ref := range references {
			refLabels[i] = makeLabel(ref.Kind, ref.Key)
		}
		deps[label] = refLabels
	}

	for _, variable := range importData.Variables {
		label := makeLabel(resolve.RefKindVariable, variable.Name)
		references := resolve.ExtractReferences(variable.Value)
		refLabels := make([]string, len(references))
		for i, ref := range references {
			refLabels[i] = makeLabel(ref.Kind, ref.Key)
		}
		deps[label] = refLabels
	}

	return deps
}

func findCycle(start string, getDependencies func(string) []string) []string {
	stack := make([]string, 0)
	onStack := make(map[string]struct{})
	visited := make(map[string]struct{})

	var visit func(node string) []string
	visit = func(node string) []string {
		if _, ok := onStack[node]; ok {
			cycleStart := slices.Index(stack, node)
			cycle := make([]string, len(stack[cycleStart:]))
			copy(cycle, stack[cycleStart:])
			return append(cycle, node)
		}
		if _, ok := visited[node]; ok {
			return nil
		}

		visited[node] = struct{}{}
		stack = append(stack, node)
		onStack[node] = struct{}{}

		for _, dep := range getDependencies(node) {
			if result := visit(dep); result != nil {
				return result
			}
		}

		stack = stack[:len(stack)-1]
		delete(onStack, node)
		return nil
	}

	return visit(start)
}

func validateNoCycles(deps map[string][]string) error {
	getDeps := func(label string) []string {
		return deps[label]
	}

	for label := range deps {
		cycle := findCycle(label, getDeps)
		if cycle != nil {
			return ErrImportCycle
		}
	}
	return nil
}

func parseImportFile(path string) (TransferDocument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return TransferDocument{}, &ImportFileError{Path: path, Err: err}
	}

	var doc TransferDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return TransferDocument{}, &ImportFileError{Path: path, Err: err}
	}

	if !slices.Contains(supportedVersions, doc.Version) {
		return TransferDocument{}, &ImportFileError{Path: path, Err: ErrUnsupportedVersion}
	}

	return doc, nil
}

type ImportService struct {
	CmdService  CommandService
	VarService  VariableService
	TagService  TagService
	ProfileRepo repository.ProfileRepository
	Result      ImportResult
}

func NewImportService(
	cmdService CommandService,
	varService VariableService,
	tagService TagService,
	profileRepo repository.ProfileRepository,
) *ImportService {
	return &ImportService{
		CmdService:  cmdService,
		VarService:  varService,
		TagService:  tagService,
		ProfileRepo: profileRepo,
		Result:      ImportResult{},
	}
}

// resolveProfileName determines the effective profile name by using the provided value or fetching the active profile.
func (s *ImportService) resolveProfileName(profileName *string) (string, error) {
	if profileName != nil {
		return *profileName, nil
	}
	profile, err := s.ProfileRepo.GetActiveCommandProfile()
	if err != nil {
		return "", fmt.Errorf("getting active command profile %w", err)
	}
	return profile.Name, nil
}

type actionType string

const (
	actionCreate    actionType = "create"
	actionOverwrite actionType = "overwrite"
	actionSkip      actionType = "skip"
)

type commandAction struct {
	Data   TransferCommand
	Action actionType
}

type variableAction struct {
	Data   TransferVariable
	Action actionType
}

func (s *ImportService) ImportFile(path string, overwrite bool, preview bool, profileName *string) (ImportResult, error) {
	resolvedProfileName, err := s.resolveProfileName(profileName)
	if err != nil {
		return ImportResult{}, fmt.Errorf("resolving profile %w", err)
	}
	s.Result.Preview = preview
	s.Result.Profile = resolvedProfileName

	doc, err := parseImportFile(path)
	if err != nil {
		return ImportResult{}, fmt.Errorf("parsing import file %w", err)
	}

	deps := buildDependencyGraph(doc)
	if err = validateNoCycles(deps); err != nil {
		return ImportResult{}, err // Raw correct error, return directly
	}

	var cmdActions []commandAction
	for _, cmdData := range doc.Commands {
		existing, err := s.CmdService.GetCommandOrNil(cmdData.Alias, &resolvedProfileName)
		if err != nil {
			return ImportResult{}, fmt.Errorf("getting command %w", err)
		}

		switch {
		case existing == nil:
			cmdActions = append(cmdActions, commandAction{Data: cmdData, Action: actionCreate})
		case overwrite:
			cmdActions = append(cmdActions, commandAction{Data: cmdData, Action: actionOverwrite})
		default:
			cmdActions = append(cmdActions, commandAction{Data: cmdData, Action: actionSkip})
		}
	}

	var varActions []variableAction
	for _, varData := range doc.Variables {
		existing, err := s.VarService.GetVariableOrNil(varData.Name, &resolvedProfileName)
		if err != nil {
			return ImportResult{}, fmt.Errorf("getting variable %w", err)
		}

		switch {
		case existing == nil:
			varActions = append(varActions, variableAction{Data: varData, Action: actionCreate})
		case overwrite:
			varActions = append(varActions, variableAction{Data: varData, Action: actionOverwrite})
		default:
			varActions = append(varActions, variableAction{Data: varData, Action: actionSkip})
		}
	}

	s.classifyCommands(cmdActions)
	s.classifyVariables(varActions)

	if preview {
		return s.Result, nil
	}

	tagNames := s.collectTagNames(cmdActions, varActions)
	if err = s.ensureTagsExist(tagNames); err != nil {
		return ImportResult{}, err
	}

	if err = s.handleCommands(cmdActions, resolvedProfileName); err != nil {
		return ImportResult{}, err
	}
	if err = s.handleVariables(varActions, resolvedProfileName); err != nil {
		return ImportResult{}, err
	}

	return s.Result, nil
}

func (s *ImportService) classifyCommands(cmdActions []commandAction) {
	for _, action := range cmdActions {
		alias := action.Data.Alias
		if action.Action == actionCreate {
			s.Result.CommandsCreated = append(s.Result.CommandsCreated, alias)
		} else if action.Action == actionOverwrite {
			s.Result.CommandsOverwritten = append(s.Result.CommandsOverwritten, alias)
		} else {
			s.Result.CommandsSkipped = append(s.Result.CommandsSkipped, alias)
		}
	}
}

func (s *ImportService) handleCommands(cmdActions []commandAction, profileName string) error {
	for _, action := range cmdActions {
		if action.Action == actionSkip {
			continue
		}
		alias := action.Data.Alias
		tags := action.Data.Tags

		if action.Action == actionCreate {
			input := CreateCommandConfig{
				Alias:       alias,
				Template:    action.Data.Template,
				Description: action.Data.Description,
				Tags:        tags,
				Cwd:         action.Data.Cwd,
				Shell:       action.Data.Shell,
				Env:         action.Data.Env,
				Timeout:     action.Data.Timeout,
				ProfileName: &profileName,
			}
			if _, err := s.CmdService.CreateCommand(input); err != nil {
				return fmt.Errorf("creating command %w", err)
			}
		} else {
			if err := s.overwriteExistingCommand(action, profileName); err != nil {
				return fmt.Errorf("overwriting existing command %w", err)
			}
		}
	}
	return nil
}

func (s *ImportService) overwriteExistingCommand(action commandAction, profileName string) error {
	existing, err := s.CmdService.GetCommandWithTags(action.Data.Alias, &profileName)
	if err != nil {
		return fmt.Errorf("getting existing command %w", err)
	}
	currentTags := make(map[string]struct{}, len(existing.Tags))
	for _, tag := range existing.Tags {
		currentTags[tag.Name] = struct{}{}
	}
	newTags := make(map[string]struct{}, len(action.Data.Tags))
	for _, tag := range action.Data.Tags {
		newTags[tag] = struct{}{}
	}

	updateConfig := UpdateCommandConfig{
		Template:     &action.Data.Template,
		Description:  action.Data.Description,
		Cwd:          action.Data.Cwd,
		ClearCwd:     false,
		Shell:        action.Data.Shell,
		ClearShell:   false,
		Env:          action.Data.Env,
		ClearEnv:     false,
		Timeout:      action.Data.Timeout,
		ClearTimeout: false,
	}
	if _, err = s.CmdService.UpdateCommand(existing.Alias, &profileName, updateConfig); err != nil {
		return fmt.Errorf("updating command %w", err)
	}

	var toRemove []string
	for t := range currentTags {
		if _, found := newTags[t]; !found {
			toRemove = append(toRemove, t)
		}
	}

	var toAdd []string
	for t, _ := range newTags {
		if _, found := currentTags[t]; !found {
			toAdd = append(toAdd, t)
		}
	}

	if _, err := s.CmdService.RemoveTags(action.Data.Alias, toRemove, &profileName); err != nil {
		return fmt.Errorf("removing tags %w", err)
	}

	if _, err := s.CmdService.AddTags(action.Data.Alias, toAdd, &profileName); err != nil {
		return fmt.Errorf("adding tags %w", err)
	}

	return nil
}

func (s *ImportService) classifyVariables(varActions []variableAction) {
	for _, action := range varActions {
		name := action.Data.Name
		if action.Action == actionCreate {
			s.Result.VariablesCreated = append(s.Result.VariablesCreated, name)
		} else if action.Action == actionOverwrite {
			s.Result.VariablesOverwritten = append(s.Result.VariablesOverwritten, name)
		} else {
			s.Result.VariablesSkipped = append(s.Result.VariablesSkipped, name)
		}
	}
}

func (s *ImportService) handleVariables(varActions []variableAction, profileName string) error {
	for _, action := range varActions {
		if action.Action == actionSkip {
			continue
		}
		name := action.Data.Name
		tags := action.Data.Tags

		if action.Action == actionCreate {
			input := CreateVariableConfig{
				Name:        name,
				Value:       action.Data.Value,
				Tags:        tags,
				ProfileName: &profileName,
			}
			if _, err := s.VarService.CreateVariable(input); err != nil {
				return fmt.Errorf("creating variable %w", err)
			}
		} else {
			if err := s.overwriteExistingVariable(action, profileName); err != nil {
				return fmt.Errorf("overwriting variable %w", err)
			}
		}
	}
	return nil
}

func (s *ImportService) overwriteExistingVariable(action variableAction, profileName string) error {
	existing, err := s.VarService.GetVariableWithTags(action.Data.Name, &profileName)
	if err != nil {
		return fmt.Errorf("getting existing variable %w", err)
	}
	currentTags := make(map[string]struct{}, len(existing.Tags))
	for _, tag := range existing.Tags {
		currentTags[tag.Name] = struct{}{}
	}
	newTags := make(map[string]struct{})
	for _, tag := range action.Data.Tags {
		newTags[tag] = struct{}{}
	}

	updateConfig := UpdateVariableConfig{
		Value: &action.Data.Value,
	}
	if _, err := s.VarService.UpdateVariable(existing.Name, &profileName, updateConfig); err != nil {
		return fmt.Errorf("updating variable %w", err)
	}

	var toRemove []string
	for t, _ := range currentTags {
		if _, found := newTags[t]; !found {
			toRemove = append(toRemove, t)
		}
	}

	var toAdd []string
	for t := range newTags {
		if _, found := currentTags[t]; !found {
			toAdd = append(toAdd, t)
		}
	}

	if _, err := s.VarService.RemoveTags(existing.Name, toRemove, &profileName); err != nil {
		return fmt.Errorf("removing tags from variable %w", err)
	}

	if _, err := s.VarService.AddTags(existing.Name, toAdd, &profileName); err != nil {
		return fmt.Errorf("adding tags to variable %w", err)
	}

	return nil
}

func (s *ImportService) collectTagNames(cmdActions []commandAction, varActions []variableAction) []string {
	seen := make(map[string]struct{})
	for _, action := range cmdActions {
		if action.Action != actionSkip {
			for _, t := range action.Data.Tags {
				seen[t] = struct{}{}
			}
		}
	}
	for _, action := range varActions {
		if action.Action != actionSkip {
			for _, t := range action.Data.Tags {
				seen[t] = struct{}{}
			}
		}
	}
	names := make([]string, 0, len(seen))
	for t := range seen {
		names = append(names, t)
	}
	return names
}

func (s *ImportService) ensureTagsExist(tagNames []string) error {
	for _, tagName := range tagNames {
		tag, err := s.TagService.GetTagOrNil(tagName)
		if err != nil {
			return fmt.Errorf("getting tag %w", err)
		}
		if tag == nil {
			if _, err = s.TagService.CreateTag(tagName, nil); err != nil {
				return fmt.Errorf("creating tag %w", err)
			}
		}
	}
	return nil
}
