package handlers

import (
	"fmt"
	"strings"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/presenters"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

type AddCmdArgs struct {
	Alias          string
	Template       string
	HasTemplate    bool
	Description    string
	HasDescription bool
	Interactive    bool
	Cwd            string
	Shell          string
	Env            []string // raw "key=value,key=value" entries, to be parsed in handler
	Timeout        int
	HasTimeout     bool
	Profile        string
}

type UpdateCmdArgs struct {
	CurrentAlias   string
	NewAlias       string
	Template       string
	HasTemplate    bool
	Description    string
	HasDescription bool
	Cwd            string
	HasCwd         bool
	ClearCwd       bool
	Shell          string
	HasShell       bool
	ClearShell     bool
	Env            []string
	HasEnv         bool
	ClearEnv       bool
	Timeout        int
	HasTimeout     bool
	ClearTimeout   bool
	SetPairs       string
	EditMode       bool
	EditFields     string
	Profile        string
}

type ListCmdArgs struct {
	ListParams
	Tags []string
}

type SearchCmdArgs struct {
	SearchParams
}

type DeleteCmdArgs struct {
	Alias   string
	Profile string
}

type TagCmdArgs struct {
	TagParams
	Alias string
}

func parseEnvPairs(pairs []string) (map[string]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}
	out := make(map[string]string, len(pairs))
	for _, p := range pairs {
		k, v, ok := strings.Cut(p, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --env value %q, expected KEY=value", p)
		}
		out[strings.TrimSpace(k)] = v
	}
	return out, nil
}

func RunAddCmd(args AddCmdArgs) error {
	commandService, err := container.GetCommandService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	env, err := parseEnvPairs(args.Env)
	if err != nil {
		return err
	}

	input := services.CreateCommandConfig{
		Alias:       args.Alias,
		Template:    args.Template,
		ProfileName: stringOrNil(args.Profile),
		Cwd:         stringOrNil(args.Cwd),
		Shell:       stringOrNil(args.Shell),
		Env:         env,
	}
	if args.HasDescription {
		input.Description = &args.Description
	}
	if args.HasTimeout {
		input.Timeout = &args.Timeout
	}
	cmd, err := commandService.CreateCommand(input)
	if err != nil {
		return fmt.Errorf("creating command: %w", err)
	}
	console.Print(presenters.RenderCommandCreated(*cmd, console.Palette()))
	return nil
}

func RunGetCmd(name, profile string) error {
	commandService, err := container.GetCommandService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	cmd, err := commandService.GetCommandWithTags(name, stringOrNil(profile))
	if err != nil {
		return fmt.Errorf("getting command: %w", err)
	}
	console.Print(presenters.RenderCommand(*cmd, console.Palette()))
	return nil
}

func RunUpdateCmd(args UpdateCmdArgs) error {
	commandService, err := container.GetCommandService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	input := services.UpdateCommandConfig{
		ClearCwd:     args.ClearCwd,
		ClearShell:   args.ClearShell,
		ClearEnv:     args.ClearEnv,
		ClearTimeout: args.ClearTimeout,
	}
	if args.NewAlias != "" {
		input.NewAlias = &args.NewAlias
	}
	if args.HasTemplate {
		input.Template = &args.Template
	}
	if args.HasDescription {
		input.Description = &args.Description
	}
	if args.HasCwd {
		input.Cwd = &args.Cwd
	}
	if args.HasShell {
		input.Shell = &args.Shell
	}
	if args.HasTimeout {
		input.Timeout = &args.Timeout
	}
	if args.HasEnv {
		env, err := parseEnvPairs(args.Env)
		if err != nil {
			return err
		}
		input.Env = env
	}

	cmd, err := commandService.UpdateCommand(args.CurrentAlias, stringOrNil(args.Profile), input)
	if err != nil {
		return fmt.Errorf("updating command: %w", err)
	}
	console.Print(presenters.RenderCommandUpdated(*cmd, console.Palette()))
	return nil
}

func RunListCmd(args ListCmdArgs) error {
	commandService, err := container.GetCommandService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}
	var limit *int
	if args.HasLimit {
		limit = &args.Limit
	}

	cmds, err := commandService.ListCommands(args.Order, args.Tags, limit, stringOrNil(args.Profile))
	if err != nil {
		return fmt.Errorf("listing commands: %w", err)
	}

	rendered := presenters.RenderCommandList(cmds, "Commands", args.Fields, console.Palette())
	rowCount := len(cmds)
	return console.PrintPaged(rendered, &rowCount, args.Page)
}

func RunSearchCmd(args SearchCmdArgs) error {
	commandService, err := container.GetCommandService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}
	var limit *int
	if args.HasLimit {
		limit = &args.Limit
	}

	cmds, err := commandService.SearchCommands(args.Term, args.SearchIn, limit, stringOrNil(args.Profile))
	if err != nil {
		return fmt.Errorf("searching commands: %w", err)
	}

	rendered := presenters.RenderCommandList(cmds, "Search Results", args.Fields, console.Palette())
	rowCount := len(cmds)
	return console.PrintPaged(rendered, &rowCount, args.Page)
}

func RunDeleteCmd(args DeleteCmdArgs) error {
	commandService, err := container.GetCommandService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	profile := stringOrNil(args.Profile)
	cmd, err := commandService.GetCommand(args.Alias, profile)
	if err != nil {
		return fmt.Errorf("getting command: %w", err)
	}
	if err := commandService.DeleteCommand(args.Alias, profile); err != nil {
		return fmt.Errorf("deleting command: %w", err)
	}

	console.Print(presenters.RenderCommandDeleted(*cmd, console.Palette()))
	return nil
}

func RunTagCmd(args TagCmdArgs) error {
	commandService, err := container.GetCommandService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	result, err := commandService.AddTags(args.Alias, args.TagNames, stringOrNil(args.Profile))
	if err != nil {
		return fmt.Errorf("adding tags: %w", err)
	}
	if len(result.Added) > 0 {
		console.Success(fmt.Sprintf("Added tags to '%s': %s", args.Alias, strings.Join(result.Added, ", ")))
	}
	if len(result.Existing) > 0 {
		console.Muted(fmt.Sprintf("Already tagged with: %s", strings.Join(result.Existing, ", ")))
	}
	return nil
}

func RunUntagCmd(args TagCmdArgs) error {
	commandService, err := container.GetCommandService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	result, err := commandService.RemoveTags(args.Alias, args.TagNames, stringOrNil(args.Profile))
	if err != nil {
		return fmt.Errorf("removing tags: %w", err)
	}
	if len(result.Removed) > 0 {
		console.Success(fmt.Sprintf("Removed tags from '%s': %s", args.Alias, strings.Join(result.Removed, ", ")))
	}
	if len(result.NotAttached) > 0 {
		console.Muted(fmt.Sprintf("Not attached: %s", strings.Join(result.NotAttached, ", ")))
	}
	return nil
}
