package handlers

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/presenters"
	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/prompts"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

type RawRunContext struct {
	Cwd        string
	HasCwd     bool
	Env        []string
	Capture    bool
	Shell      string
	HasShell   bool
	Timeout    int
	HasTimeout bool
	Emit       bool
	Verbose    *bool // tri-state
}

func RunRunCommand(alias string, runtimeVars map[string]string, ctx RawRunContext, profile *string) error {
	runService, err := container.GetRunService()
	if err != nil {
		return fmt.Errorf("getting run service: %w", err)
	}

	missingVars, err := runService.CollectMissingVars(alias, runtimeVars, profile)
	if err != nil {
		return fmt.Errorf("collecting missing vars: %w", err)
	}

	if len(missingVars) > 0 {
		for _, varName := range missingVars {
			value, err := prompts.PromptForMissingVars(varName)
			if err != nil {
				return fmt.Errorf("prompting for %s: %w", varName, err)
			}
			runtimeVars[varName] = value
		}
	}

	overrides, err := toRuntimeOverrides(ctx)
	if err != nil {
		return fmt.Errorf("building runtime overrides from context: %w", err)
	}

	executionResult, err := runService.Run(alias, overrides, runtimeVars, profile)
	if err != nil {
		return fmt.Errorf("running command: %w", err)
	}

	// Run only has output when verbose is true
	if ctx.Verbose != nil && *ctx.Verbose && !ctx.Emit {
		console, err := container.GetConsole()
		if err != nil {
			return fmt.Errorf("getting console: %w", err)
		}
		console.Print(presenters.RenderExecutionResult(executionResult, "Execution Result", console.Palette()))
	}

	return nil
}

func RunPreviewCommand(alias string, runtimeVars map[string]string, ctx RawRunContext, profile *string) error {
	runService, err := container.GetRunService()
	if err != nil {
		return fmt.Errorf("getting run service: %w", err)
	}

	missingVars, err := runService.CollectMissingVars(alias, runtimeVars, profile)
	if err != nil {
		return fmt.Errorf("collecting missing vars: %w", err)
	}

	if len(missingVars) > 0 {
		for _, varName := range missingVars {
			value, err := prompts.PromptForMissingVars(varName)
			if err != nil {
				return fmt.Errorf("prompting for %s: %w", varName, err)
			}
			runtimeVars[varName] = value
		}
	}

	overrides, err := toRuntimeOverrides(ctx)
	if err != nil {
		return fmt.Errorf("building runtime overrides from context: %w", err)
	}

	result, effectiveCtx, err := runService.Preview(alias, overrides, runtimeVars, profile)
	if err != nil {
		return fmt.Errorf("previewing run: %w", err)
	}

	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	renderedResult := presenters.RenderPreviewResult(result, "Execution Preview", true, effectiveCtx, console.Palette())
	console.Print(renderedResult)

	return nil
}

func toRuntimeOverrides(raw RawRunContext) (services.RuntimeOverrides, error) {
	var overrides services.RuntimeOverrides

	if raw.HasCwd {
		overrides.Cwd = raw.Cwd
	}
	if raw.HasShell {
		overrides.Shell = &raw.Shell
	}
	if raw.HasTimeout {
		overrides.Timeout = &raw.Timeout
	}
	if len(raw.Env) > 0 {
		env, err := parseEnvPairs(raw.Env)
		if err != nil {
			return services.RuntimeOverrides{}, err
		}
		overrides.Env = env
	}

	return overrides, nil
}
