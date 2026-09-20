package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:                "run <alias> [--flag value ...]",
	Short:              "Run a saved command by its alias",
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE:               runRun,
}

var previewCmd = &cobra.Command{
	Use:                "preview <alias> [--flag value ...]",
	Short:              "Show a fully resolved command without running it",
	Args:               cobra.MinimumNArgs(1),
	DisableFlagParsing: true,
	RunE:               runPreview,
}

func init() {
	rootCmd.AddCommand(runCmd, previewCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	if wantsHelp(args) {
		return cmd.Help()
	}
	alias, ctx, profile, previewFlag, runtimeVars, err := parseInvocation(args)
	if err != nil {
		return err
	}
	if previewFlag {
		return handlers.RunPreviewCommand(alias, runtimeVars, ctx, profile)
	}
	return handlers.RunRunCommand(alias, runtimeVars, ctx, profile)
}

func runPreview(cmd *cobra.Command, args []string) error {
	if wantsHelp(args) {
		return cmd.Help()
	}
	alias, ctx, profile, _, runtimeVars, err := parseInvocation(args)
	if err != nil {
		return err
	}
	return handlers.RunPreviewCommand(alias, runtimeVars, ctx, profile)
}

func wantsHelp(args []string) bool {
	for _, a := range args {
		if a == "--help" || a == "-h" {
			return true
		}
	}
	return false
}

func parseInvocation(args []string) (
	alias string, ctx handlers.RawRunContext, profile string, previewFlag bool, runtimeVars map[string]string, err error,
) {
	runtimeVars = map[string]string{}
	var positional []string

	i := 0
	for i < len(args) {
		tok := args[i]
		switch tok {
		case "--preview", "-p":
			previewFlag = true
			i++
		case "--cwd", "-d":
			ctx.Cwd, i, err = takeValue(args, i)
			ctx.HasCwd = true
		case "--env":
			var v string
			v, i, err = takeValue(args, i)
			ctx.Env = append(ctx.Env, v)
		case "--capture", "-c":
			ctx.Capture = true
			i++
		case "--shell", "-s":
			ctx.Shell, i, err = takeValue(args, i)
			ctx.HasShell = true
		case "--timeout", "-t":
			var v string
			v, i, err = takeValue(args, i)
			if err == nil {
				ctx.Timeout, err = strconv.Atoi(v)
			}
		case "--emit", "-e":
			ctx.Emit = true
			i++
		case "--verbose":
			v := true
			ctx.Verbose = &v
			i++
		case "--no-verbose":
			v := false
			ctx.Verbose = &v
			i++
		case "--profile":
			profile, i, err = takeValue(args, i)
		default:
			if strings.HasPrefix(tok, "--") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				runtimeVars[strings.TrimPrefix(tok, "--")] = args[i+1]
				i += 2
			} else if strings.HasPrefix(tok, "--") {
				i++ // ignore flag with no value
			} else {
				positional = append(positional, tok)
				i++
			}
		}
		if err != nil {
			return "", ctx, "", false, nil, err
		}
	}
	if len(positional) == 0 {
		return "", ctx, "", false, nil, fmt.Errorf("missing required argument: alias")
	}
	return positional[0], ctx, profile, previewFlag, runtimeVars, nil
}

func takeValue(args []string, i int) (string, int, error) {
	if i+1 >= len(args) {
		return "", i, fmt.Errorf("%s requires a value", args[i])
	}
	return args[i+1], i + 2, nil
}
