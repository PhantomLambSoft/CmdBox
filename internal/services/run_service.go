package services

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
	"github.com/PhantomLambSoft/CmdBox/internal/resolve"
	"github.com/PhantomLambSoft/CmdBox/internal/runtime"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

// SettingsProvider defines an interface for allowing SettingsService to be mocked for testing without dealing with
// the I/O functionality that comes with SettingsService.
type SettingsProvider interface {
	Get() settings.Settings
}

type RuntimeOverrides struct {
	Cwd     string
	Env     map[string]string
	Capture *bool
	Shell   *string
	Timeout *int
	Emit    bool
	Verbose *bool
}

type RunService struct {
	cmdRepo          repository.CommandRepository
	resolver         resolve.Resolver
	executor         runtime.Executor
	profileRepo      repository.ProfileRepository
	historyRepo      repository.HistoryRepository
	settingsProvider SettingsProvider
}

type RunServiceConfig struct {
	CmdRepo          repository.CommandRepository
	Resolver         resolve.Resolver
	Executor         runtime.Executor
	ProfileRepo      repository.ProfileRepository
	HistoryRepo      repository.HistoryRepository
	SettingsProvider SettingsProvider
}

func NewRunService(config RunServiceConfig) *RunService {
	return &RunService{
		cmdRepo:          config.CmdRepo,
		resolver:         config.Resolver,
		executor:         config.Executor,
		profileRepo:      config.ProfileRepo,
		historyRepo:      config.HistoryRepo,
		settingsProvider: config.SettingsProvider,
	}
}

func (s *RunService) resolveProfile(profileName *string) (*models.Profile, error) {
	if profileName != nil {
		if *profileName != "" {
			profile, err := s.profileRepo.GetByName(*profileName)
			if err != nil {
				return nil, fmt.Errorf("getting profile: %w", err)
			}
			return profile, nil
		}
	}
	return nil, nil
}

func (s *RunService) Run(
	cmdAlias string, runtimeOverrides RuntimeOverrides, runtimeVars map[string]string, profileName *string,
) (runtime.ExecutionResult, error) {
	profile, err := s.resolveProfile(profileName)
	if err != nil {
		return runtime.ExecutionResult{}, err
	}
	var profileID *uint
	if profile != nil {
		profileID = &profile.ID
	}

	command, err := s.cmdRepo.GetByAlias(cmdAlias, profileID)
	if err != nil {
		return runtime.ExecutionResult{}, fmt.Errorf("getting command: %w", err)
	}

	resolveResult, err := s.resolver.Resolve(command.Template, nil, runtimeVars)
	if err != nil {
		return runtime.ExecutionResult{}, fmt.Errorf("resolving command: %w", err)
	}
	resolvedCommand := resolveResult.Text

	effectiveCtx, err := s.buildContext(*command, runtimeOverrides)
	if err != nil {
		return runtime.ExecutionResult{}, fmt.Errorf("building context: %w", err)
	}

	// TODO: when effectiveCtx.Emit is true, executor.Run calls os.Exit(0) directly instead of returning,
	// which kills the calling process (untestable in-process) and would leave executionResult nil -
	// panicking below on `*executionResult` - if that exit path were ever removed. Consider having the
	// executor return a normal result/sentinel for emit instead of exiting itself.
	executionResult, err := s.executor.Run(resolvedCommand, effectiveCtx)
	if err != nil {
		return runtime.ExecutionResult{}, fmt.Errorf("running command: %w", err)
	}

	if err = s.cmdRepo.RecordUse(command.ID); err != nil {
		slog.Warn("failed to record command use", "alias", cmdAlias, "error", err)
	}

	historyInput := RecordHistoryInput{
		Alias:       cmdAlias,
		Template:    command.Template,
		Resolved:    resolvedCommand,
		ProfileID:   profileID,
		RuntimeVars: runtimeVars,
		ExitCode:    executionResult.ExitCode,
	}

	if _, err = s.recordHistory(historyInput); err != nil {
		slog.Warn("failed to record history", "alias", cmdAlias, "error", err)
	}

	if err = s.recordVariableProfileUse(resolveResult); err != nil {
		slog.Warn("failed to record variable profile use", "alias", cmdAlias, "error", err)
	}

	return *executionResult, nil
}

func (s *RunService) Preview(
	cmdAlias string, runtimeOverrides RuntimeOverrides, runtimeVars map[string]string, profileName *string,
) (resolve.Result, runtime.RunContext, error) {
	profile, err := s.resolveProfile(profileName)
	if err != nil {
		return resolve.Result{}, runtime.RunContext{}, err
	}
	var profileID *uint
	if profile != nil {
		profileID = &profile.ID
	}

	command, err := s.cmdRepo.GetByAlias(cmdAlias, profileID)
	if err != nil {
		return resolve.Result{}, runtime.RunContext{}, fmt.Errorf("getting command: %w", err)
	}

	resolveResult, err := s.resolver.Resolve(command.Template, nil, runtimeVars)
	if err != nil {
		return resolve.Result{}, runtime.RunContext{}, fmt.Errorf("resolving command: %w", err)
	}

	effectiveCtx, err := s.buildContext(*command, runtimeOverrides)
	if err != nil {
		return resolve.Result{}, runtime.RunContext{}, fmt.Errorf("building context: %w", err)
	}

	return resolveResult, effectiveCtx, nil
}

type RecordHistoryInput struct {
	Alias       string
	Template    string
	Resolved    string
	ProfileID   *uint
	RuntimeVars map[string]string
	ExitCode    int
}

func (s *RunService) recordHistory(input RecordHistoryInput) (*models.CommandHistory, error) {
	currentSettings := s.settingsProvider.Get()
	if !currentSettings.History.Enabled {
		return nil, nil
	}
	if s.historyRepo == nil {
		return nil, nil
	}

	var variablesUsed *string
	if len(input.RuntimeVars) > 0 {
		marshalled, err := json.Marshal(input.RuntimeVars)
		if err != nil {
			return nil, fmt.Errorf("marshaling runtime vars: %w", err)
		}
		s := string(marshalled)
		variablesUsed = &s
	}

	historyEntry := repository.HistoryEntry{
		Alias:          input.Alias,
		Template:       input.Template,
		Resolved:       input.Resolved,
		VariablesUsed:  variablesUsed,
		ExitCode:       &input.ExitCode,
		RetentionLimit: &currentSettings.History.LimitPerCommand,
		ProfileID:      input.ProfileID,
	}

	history, err := s.historyRepo.Record(historyEntry)
	if err != nil {
		return nil, fmt.Errorf("recording history: %w", err)
	}

	return history, nil
}

func (s *RunService) recordVariableProfileUse(resolvedCommand resolve.Result) error {
	usedStoredVar := false
	for _, step := range resolvedCommand.Trace {
		if step.Kind == resolve.RefKindVariable && step.Source == "stored" {
			usedStoredVar = true
			break
		}
	}
	if !usedStoredVar {
		return nil
	}

	profile, err := s.profileRepo.GetActiveVariableProfile()
	if err != nil {
		return fmt.Errorf("getting active variable profile: %w", err)
	}
	return s.profileRepo.RecordUse(*profile)
}

// CollectMissingVars identifies variables inside a command template that are not provided in the supplied runtimeVars map.
func (s *RunService) CollectMissingVars(cmdAlias string, runtimeVars map[string]string, profileName *string) ([]string, error) {
	profile, err := s.resolveProfile(profileName)
	if err != nil {
		return []string{}, fmt.Errorf("resolving profile: %w", err)
	}
	var profileID *uint
	if profile != nil {
		profileID = &profile.ID
	}
	cmd, err := s.cmdRepo.GetByAlias(cmdAlias, profileID)
	if err != nil {
		return []string{}, fmt.Errorf("getting command: %w", err)
	}
	missing, err := s.resolver.CollectMissingVars(cmd.Template, runtimeVars)
	if err != nil {
		return nil, fmt.Errorf("collecting missing vars: %w", err)
	}
	return missing, nil
}

// BuildContext constructs a runtime.RunContext by combining runtime override configurations with stored command
// configurations and default values taken from settings.
func (s *RunService) buildContext(cmd models.Command, runtimeOverrides RuntimeOverrides) (runtime.RunContext, error) {
	var err error
	envs := make(map[string]string)
	if cmd.Env != nil {
		envs, err = parseEnv(*cmd.Env)
		if err != nil {
			return runtime.RunContext{}, fmt.Errorf("parsing stored env: %w", err)
		}
	}
	if runtimeOverrides.Env != nil {
		maps.Copy(envs, runtimeOverrides.Env) // merge runtime envs into stored envs under envs variable
	}

	currentSettings := s.settingsProvider.Get()

	var cwd = ""
	if runtimeOverrides.Cwd != "" {
		cwd = runtimeOverrides.Cwd
	} else if cmd.Cwd != nil {
		cwd = *cmd.Cwd
	}

	var shell *string
	if runtimeOverrides.Shell != nil {
		shell = runtimeOverrides.Shell
	} else if cmd.Shell != nil {
		shell = cmd.Shell
	} else {
		shell = &currentSettings.ExecutionSettings.DefaultShell
	}

	var timeout = 0
	if runtimeOverrides.Timeout != nil {
		timeout = *runtimeOverrides.Timeout
	} else if cmd.Timeout != nil {
		timeout = *cmd.Timeout
	}

	var capture *bool
	if runtimeOverrides.Capture != nil {
		capture = runtimeOverrides.Capture
	} else {
		capture = &currentSettings.ExecutionSettings.CaptureOutput
	}

	var verbose *bool
	if runtimeOverrides.Verbose != nil {
		verbose = runtimeOverrides.Verbose
	} else {
		verbose = &currentSettings.ExecutionSettings.DefaultVerbose
	}

	emit := runtimeOverrides.Emit

	return runtime.RunContext{
		Cwd:     cwd,
		Env:     envs,
		Shell:   *shell,
		Timeout: timeout,
		Capture: *capture,
		Verbose: *verbose,
		Emit:    emit,
	}, nil
}
