package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PhantomLambSoft/CmdBox/internal/atomicfile"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
	"github.com/PhantomLambSoft/CmdBox/internal/resolve"
)

const exportLimit = 10_000

type ExportResult struct {
	Path               string
	Commands           []string
	Variables          []string
	TransientCommands  []string
	TransientVariables []string
	Warnings           []string
	CommandProfile     string
	VariableProfile    string
}

// resolveOutputPath resolves the output file path based on the provided directory or uses the current working directory.
// If outputPath is nil, it generates a file path in the current directory. If it's a directory, appends the filename.
// Returns the fully resolved output path or an error in case of failure.
func resolveOutputPath(outputPath *string, typeLabel string) (string, error) {
	filename := fmt.Sprintf("cmdbox-%s-%s.json", typeLabel, time.Now().Format("2006-01-02"))

	if outputPath == nil {
		cwd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("getting working directory %w", err)
		}
		return filepath.Join(cwd, filename), nil
	}

	if info, err := os.Stat(*outputPath); err == nil && info.IsDir() {
		return filepath.Join(*outputPath, filename), nil
	}
	return *outputPath, nil
}

func buildDocument(typeLabel string, commands []TransferCommand, variables []TransferVariable, cmdProfile, varProfile string) TransferDocument {
	return TransferDocument{
		Version:         "1",
		Type:            typeLabel,
		ExportedAt:      time.Now().UTC().Format(time.RFC3339),
		CommandProfile:  cmdProfile,
		VariableProfile: varProfile,
		Commands:        commands,
		Variables:       variables,
	}
}

// tagFilter returns a slice containing the dereferenced tag if it is not nil, otherwise returns nil.
func tagFilter(tag *string) []string {
	if tag == nil {
		return nil
	}
	return []string{*tag}
}

// collectDeepCommands recursively collects commands and their references starting from the provided aliases.
// It returns a map of commands keyed by alias, the order they were collected, and any error encountered.
func collectDeepCommands(aliases []string, cmdService CommandService, profileName *string) (map[string]*models.Command, []string, error) {
	collected := make(map[string]*models.Command)
	var order []string
	stack := append([]string(nil), aliases...)

	for len(stack) > 0 {
		alias := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, ok := collected[alias]; ok {
			continue
		}
		cmd, err := cmdService.GetCommandOrNil(alias, profileName)
		if err != nil {
			return nil, nil, fmt.Errorf("getting command %q %w", alias, err)
		}
		if cmd == nil {
			continue // reference to a command that doesn't exist, skip silently
		}
		collected[alias] = cmd
		order = append(order, alias)

		for _, ref := range resolve.ExtractReferences(cmd.Template) {
			if ref.Kind == resolve.RefKindCommand {
				if _, ok := collected[ref.Key]; !ok {
					stack = append(stack, ref.Key)
				}
			}
		}
	}

	return collected, order, nil
}

// collectDeepVariables collects all variables and their references needed for execution, resolving them recursively.
// It returns a map of variables, their order of resolution, and any encountered error.
func collectDeepVariables(names []string, commands map[string]*models.Command, varService VariableService, profileName *string) (map[string]*models.Variable, []string, error) {
	collected := make(map[string]*models.Variable)
	var order []string
	stack := append([]string(nil), names...)

	for _, cmd := range commands {
		for _, ref := range resolve.ExtractReferences(cmd.Template) {
			if ref.Kind == resolve.RefKindVariable {
				if _, ok := collected[ref.Key]; !ok {
					stack = append(stack, ref.Key)
				}
			}
		}
	}

	for len(stack) > 0 {
		name := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, ok := collected[name]; ok {
			continue
		}
		v, err := varService.GetVariableOrNil(name, profileName)
		if err != nil {
			return nil, nil, fmt.Errorf("getting variable %q %w", name, err)
		}
		if v == nil {
			continue
		}
		collected[name] = v
		order = append(order, name)

		for _, ref := range resolve.ExtractReferences(v.Value) {
			if ref.Kind == resolve.RefKindVariable {
				if _, ok := collected[ref.Key]; !ok {
					stack = append(stack, ref.Key)
				}
			}
		}
	}

	return collected, order, nil
}

// flattenTemplate processes a template string by recursively resolving embedded references to commands and variables.
// It supports nested templates and handles circular references via a tracking mechanism to prevent infinite loops.
// Returns the flattened template string or an error if resolution fails.
func flattenTemplate(template string, cmdService CommandService, varService VariableService, cmdProfileName, varProfileName *string, seen map[string]struct{}) (string, error) {
	if seen == nil {
		seen = make(map[string]struct{})
	}

	var out strings.Builder
	runes := []rune(template)
	i := 0
	for i < len(runes) {
		ch := runes[i]

		if ch == '\\' && i+1 < len(runes) {
			out.WriteRune(ch)
			out.WriteRune(runes[i+1])
			i += 2
			continue
		}

		if ch != '<' {
			out.WriteRune(ch)
			i++
			continue
		}

		tokenInner, rawToken, nextI := resolve.ReadAngleToken(runes, i)
		if tokenInner == nil {
			out.WriteRune('<')
			i++
			continue
		}

		kind, key := resolve.ParseKindAndKey(*tokenInner)
		prefix := "var"
		if kind == resolve.RefKindCommand {
			prefix = "cmd"
		}
		label := prefix + ":" + key

		if _, already := seen[label]; already {
			out.WriteString("<" + rawToken + ">")
			i = nextI
			continue
		}

		nested := make(map[string]struct{}, len(seen)+1)
		for k := range seen {
			nested[k] = struct{}{}
		}
		nested[label] = struct{}{}

		var resolved string
		var found bool
		if kind == resolve.RefKindCommand {
			rec, err := cmdService.GetCommandOrNil(key, cmdProfileName)
			if err != nil {
				return "", fmt.Errorf("resolving command reference %q %w", key, err)
			}
			if rec != nil {
				found = true
				resolved, err = flattenTemplate(rec.Template, cmdService, varService, cmdProfileName, varProfileName, nested)
				if err != nil {
					return "", err
				}
			}
		} else {
			rec, err := varService.GetVariableOrNil(key, varProfileName)
			if err != nil {
				return "", fmt.Errorf("resolving variable reference %q %w", key, err)
			}
			if rec != nil {
				found = true
				resolved, err = flattenTemplate(rec.Value, cmdService, varService, cmdProfileName, varProfileName, nested)
				if err != nil {
					return "", err
				}
			}
		}

		if found {
			out.WriteString(resolved)
		} else {
			out.WriteString("<" + rawToken + ">")
		}
		i = nextI
	}

	return out.String(), nil
}

func serializeCommand(cmd *models.Command, tags []string, flatten bool, cmdService CommandService, varService VariableService, cmdProfileName, varProfileName *string) (TransferCommand, error) {
	template := cmd.Template
	if flatten {
		flattened, err := flattenTemplate(cmd.Template, cmdService, varService, cmdProfileName, varProfileName, nil)
		if err != nil {
			return TransferCommand{}, err
		}
		template = flattened
	}

	var env map[string]string
	if cmd.Env != nil {
		if err := json.Unmarshal([]byte(*cmd.Env), &env); err != nil {
			return TransferCommand{}, fmt.Errorf("decoding env for command %q %w", cmd.Alias, err)
		}
	}

	return TransferCommand{
		Alias:       cmd.Alias,
		Template:    template,
		Description: cmd.Description,
		Tags:        tags,
		Cwd:         cmd.Cwd,
		Shell:       cmd.Shell,
		Env:         env,
		Timeout:     cmd.Timeout,
	}, nil
}

func serializeVariable(v *models.Variable, tags []string, flatten bool, cmdService CommandService, varService VariableService, cmdProfileName, varProfileName *string) (TransferVariable, error) {
	value := v.Value
	if flatten {
		flattened, err := flattenTemplate(v.Value, cmdService, varService, cmdProfileName, varProfileName, nil)
		if err != nil {
			return TransferVariable{}, err
		}
		value = flattened
	}

	return TransferVariable{
		Name:  v.Name,
		Value: value,
		Tags:  tags,
	}, nil
}

type ExportService struct {
	CmdService  CommandService
	VarService  VariableService
	ProfileRepo repository.ProfileRepository
}

func NewExportService(cmdService CommandService, varService VariableService, profileRepo repository.ProfileRepository) *ExportService {
	return &ExportService{CmdService: cmdService, VarService: varService, ProfileRepo: profileRepo}
}

func (s *ExportService) resolveCommandProfileName(profileName *string) (string, error) {
	if profileName != nil {
		return *profileName, nil
	}
	profile, err := s.ProfileRepo.GetActiveCommandProfile()
	if err != nil {
		return "", fmt.Errorf("getting active command profile %w", err)
	}
	return profile.Name, nil
}

func (s *ExportService) resolveVariableProfileName(profileName *string) (string, error) {
	if profileName != nil {
		return *profileName, nil
	}
	profile, err := s.ProfileRepo.GetActiveVariableProfile()
	if err != nil {
		return "", fmt.Errorf("getting active variable profile %w", err)
	}
	return profile.Name, nil
}

func (s *ExportService) ExportCmds(aliases []string, tag *string, flatten bool, outputPath *string, cmdProfileName, varProfileName *string) (ExportResult, error) {
	resolvedCmdProfile, err := s.resolveCommandProfileName(cmdProfileName)
	if err != nil {
		return ExportResult{}, fmt.Errorf("resolving command profile %w", err)
	}
	resolvedVarProfile, err := s.resolveVariableProfileName(varProfileName)
	if err != nil {
		return ExportResult{}, fmt.Errorf("resolving variable profile %w", err)
	}

	path, err := resolveOutputPath(outputPath, "cmds")
	if err != nil {
		return ExportResult{}, fmt.Errorf("resolving output path %w", err)
	}

	result := ExportResult{
		Path:            path,
		CommandProfile:  resolvedCmdProfile,
		VariableProfile: resolvedVarProfile,
	}

	targetAliases := aliases
	if len(targetAliases) == 0 {
		limit := exportLimit
		cmds, err := s.CmdService.ListCommands("", tagFilter(tag), &limit, &resolvedCmdProfile)
		if err != nil {
			return ExportResult{}, fmt.Errorf("listing commands %w", err)
		}
		targetAliases = make([]string, len(cmds))
		for i, c := range cmds {
			targetAliases[i] = c.Alias
		}
	}

	var serializedCmds []TransferCommand
	var serializedVars []TransferVariable

	if flatten {
		for _, alias := range targetAliases {
			cmd, err := s.CmdService.GetCommandOrNil(alias, &resolvedCmdProfile)
			if err != nil {
				return ExportResult{}, fmt.Errorf("getting command %q %w", alias, err)
			}
			if cmd == nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Command %s not found", alias))
				continue
			}
			tags, err := s.commandTagNames(alias, &resolvedCmdProfile)
			if err != nil {
				return ExportResult{}, err
			}
			serialized, err := serializeCommand(cmd, tags, true, s.CmdService, s.VarService, &resolvedCmdProfile, &resolvedVarProfile)
			if err != nil {
				return ExportResult{}, err
			}
			serializedCmds = append(serializedCmds, serialized)
			result.Commands = append(result.Commands, alias)
		}
	} else {
		collectedCmds, cmdOrder, err := collectDeepCommands(targetAliases, s.CmdService, &resolvedCmdProfile)
		if err != nil {
			return ExportResult{}, err
		}
		collectedVars, varOrder, err := collectDeepVariables(nil, collectedCmds, s.VarService, &resolvedVarProfile)
		if err != nil {
			return ExportResult{}, err
		}

		targetSet := make(map[string]struct{}, len(targetAliases))
		for _, a := range targetAliases {
			targetSet[a] = struct{}{}
		}
		for _, alias := range targetAliases {
			if _, ok := collectedCmds[alias]; !ok {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Command %s not found", alias))
			}
		}

		for _, alias := range cmdOrder {
			tags, err := s.commandTagNames(alias, &resolvedCmdProfile)
			if err != nil {
				return ExportResult{}, err
			}
			serialized, err := serializeCommand(collectedCmds[alias], tags, false, s.CmdService, s.VarService, &resolvedCmdProfile, &resolvedVarProfile)
			if err != nil {
				return ExportResult{}, err
			}
			serializedCmds = append(serializedCmds, serialized)

			if _, ok := targetSet[alias]; ok {
				result.Commands = append(result.Commands, alias)
			} else {
				result.TransientCommands = append(result.TransientCommands, alias)
			}
		}

		for _, name := range varOrder {
			tags, err := s.variableTagNames(name, &resolvedVarProfile)
			if err != nil {
				return ExportResult{}, err
			}
			serialized, err := serializeVariable(collectedVars[name], tags, false, s.CmdService, s.VarService, &resolvedCmdProfile, &resolvedVarProfile)
			if err != nil {
				return ExportResult{}, err
			}
			serializedVars = append(serializedVars, serialized)
			result.TransientVariables = append(result.TransientVariables, name)
		}
	}

	doc := buildDocument("cmds", serializedCmds, serializedVars, resolvedCmdProfile, resolvedVarProfile)
	if err := writeExportDoc(result.Path, doc); err != nil {
		return ExportResult{}, err
	}
	return result, nil
}

func (s *ExportService) ExportVars(names []string, tag *string, flatten bool, outputPath *string, varProfileName *string) (ExportResult, error) {
	resolvedVarProfile, err := s.resolveVariableProfileName(varProfileName)
	if err != nil {
		return ExportResult{}, fmt.Errorf("resolving variable profile %w", err)
	}

	path, err := resolveOutputPath(outputPath, "vars")
	if err != nil {
		return ExportResult{}, fmt.Errorf("resolving output path %w", err)
	}

	result := ExportResult{Path: path, VariableProfile: resolvedVarProfile}

	targetNames := names
	if len(targetNames) == 0 {
		limit := exportLimit
		vars, err := s.VarService.ListVariables("", tagFilter(tag), &limit, &resolvedVarProfile)
		if err != nil {
			return ExportResult{}, fmt.Errorf("listing variables %w", err)
		}
		targetNames = make([]string, len(vars))
		for i, v := range vars {
			targetNames[i] = v.Name
		}
	}

	var serializedVars []TransferVariable

	if flatten {
		for _, name := range targetNames {
			v, err := s.VarService.GetVariableOrNil(name, &resolvedVarProfile)
			if err != nil {
				return ExportResult{}, fmt.Errorf("getting variable %q %w", name, err)
			}
			if v == nil {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Variable %s not found", name))
				continue
			}
			tags, err := s.variableTagNames(name, &resolvedVarProfile)
			if err != nil {
				return ExportResult{}, err
			}
			serialized, err := serializeVariable(v, tags, true, s.CmdService, s.VarService, nil, &resolvedVarProfile)
			if err != nil {
				return ExportResult{}, err
			}
			serializedVars = append(serializedVars, serialized)
			result.Variables = append(result.Variables, name)
		}
	} else {
		collectedVars, varOrder, err := collectDeepVariables(targetNames, nil, s.VarService, &resolvedVarProfile)
		if err != nil {
			return ExportResult{}, err
		}

		targetSet := make(map[string]struct{}, len(targetNames))
		for _, n := range targetNames {
			targetSet[n] = struct{}{}
		}
		for _, name := range targetNames {
			if _, ok := collectedVars[name]; !ok {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Variable %s not found", name))
			}
		}

		for _, name := range varOrder {
			tags, err := s.variableTagNames(name, &resolvedVarProfile)
			if err != nil {
				return ExportResult{}, err
			}
			serialized, err := serializeVariable(collectedVars[name], tags, false, s.CmdService, s.VarService, nil, &resolvedVarProfile)
			if err != nil {
				return ExportResult{}, err
			}
			serializedVars = append(serializedVars, serialized)

			if _, ok := targetSet[name]; ok {
				result.Variables = append(result.Variables, name)
			} else {
				result.TransientVariables = append(result.TransientVariables, name)
			}
		}
	}

	doc := buildDocument("vars", nil, serializedVars, "", resolvedVarProfile)
	if err := writeExportDoc(result.Path, doc); err != nil {
		return ExportResult{}, err
	}
	return result, nil
}

func (s *ExportService) ExportAll(flatten bool, outputPath *string, cmdProfileName, varProfileName *string) (ExportResult, error) {
	resolvedCmdProfile, err := s.resolveCommandProfileName(cmdProfileName)
	if err != nil {
		return ExportResult{}, fmt.Errorf("resolving command profile %w", err)
	}
	resolvedVarProfile, err := s.resolveVariableProfileName(varProfileName)
	if err != nil {
		return ExportResult{}, fmt.Errorf("resolving variable profile %w", err)
	}

	path, err := resolveOutputPath(outputPath, "all")
	if err != nil {
		return ExportResult{}, fmt.Errorf("resolving output path %w", err)
	}

	result := ExportResult{
		Path:            path,
		CommandProfile:  resolvedCmdProfile,
		VariableProfile: resolvedVarProfile,
	}

	limit := exportLimit
	allCmds, err := s.CmdService.ListCommands("", nil, &limit, &resolvedCmdProfile)
	if err != nil {
		return ExportResult{}, fmt.Errorf("listing commands %w", err)
	}
	allVars, err := s.VarService.ListVariables("", nil, &limit, &resolvedVarProfile)
	if err != nil {
		return ExportResult{}, fmt.Errorf("listing variables %w", err)
	}

	serializedCmds := make([]TransferCommand, 0, len(allCmds))
	for _, cmd := range allCmds {
		tags, err := s.commandTagNames(cmd.Alias, &resolvedCmdProfile)
		if err != nil {
			return ExportResult{}, err
		}
		serialized, err := serializeCommand(&cmd, tags, flatten, s.CmdService, s.VarService, &resolvedCmdProfile, &resolvedVarProfile)
		if err != nil {
			return ExportResult{}, err
		}
		serializedCmds = append(serializedCmds, serialized)
		result.Commands = append(result.Commands, cmd.Alias)
	}

	serializedVars := make([]TransferVariable, 0, len(allVars))
	for _, v := range allVars {
		tags, err := s.variableTagNames(v.Name, &resolvedVarProfile)
		if err != nil {
			return ExportResult{}, err
		}

		serialized, err := serializeVariable(&v, tags, flatten, s.CmdService, s.VarService, &resolvedCmdProfile, &resolvedVarProfile)
		if err != nil {
			return ExportResult{}, err
		}
		serializedVars = append(serializedVars, serialized)
		result.Variables = append(result.Variables, v.Name)
	}

	doc := buildDocument("all", serializedCmds, serializedVars, resolvedCmdProfile, resolvedVarProfile)
	if err := writeExportDoc(result.Path, doc); err != nil {
		return ExportResult{}, err
	}
	return result, nil
}

func (s *ExportService) commandTagNames(alias string, profileName *string) ([]string, error) {
	withTags, err := s.CmdService.GetCommandWithTags(alias, profileName)
	if err != nil {
		return nil, fmt.Errorf("getting tags for command %q %w", alias, err)
	}
	names := make([]string, len(withTags.Tags))
	for i, t := range withTags.Tags {
		names[i] = t.Name
	}
	return names, nil
}

func (s *ExportService) variableTagNames(name string, profileName *string) ([]string, error) {
	withTags, err := s.VarService.GetVariableWithTags(name, profileName)
	if err != nil {
		return nil, fmt.Errorf("getting tags for variable %q %w", name, err)
	}
	names := make([]string, len(withTags.Tags))
	for i, t := range withTags.Tags {
		names[i] = t.Name
	}
	return names, nil
}

func writeExportDoc(path string, doc TransferDocument) error {
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding export document %w", err)
	}
	if err := atomicfile.WriteFile(path, string(data)); err != nil {
		return fmt.Errorf("writing export file %w", err)
	}
	return nil
}
