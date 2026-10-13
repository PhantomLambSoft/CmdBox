package handlers

import (
	"fmt"
	"strings"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/presenters"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

type AddVarArgs struct {
	Name        string
	Value       string
	HasValue    bool
	Interactive bool
	Profile     string
}

type UpdateVarArgs struct {
	CurrentName string
	NewName     string
	NewValue    string
	HasValue    bool
	SetPairs    string // raw "key=value,key=value" - parsed in handler function
	EditMode    bool
	EditFields  string
	Profile     string
}

type ListVarArgs struct {
	ListParams
	Tags []string
}

type DeleteVarArgs struct {
	Name    string
	Profile string
}

type TagVarArgs struct {
	TagParams
	Name string
}

func RunAddVar(args AddVarArgs) error {
	variableService, err := container.GetVariableService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	input := services.CreateVariableConfig{
		Name:        args.Name,
		Value:       args.Value,
		ProfileName: stringOrNil(args.Profile),
	}

	variable, err := variableService.CreateVariable(input)
	if err != nil {
		return fmt.Errorf("creating variable: %w", err)
	}
	console.Print(presenters.RenderVariableCreated(*variable, console.Palette()))
	return nil
}

func RunGetVar(name, profile string) error {
	variableService, err := container.GetVariableService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}
	variable, err := variableService.GetVariableWithTags(name, stringOrNil(profile))
	if err != nil {
		return fmt.Errorf("getting variable: %w", err)
	}
	console.Print(presenters.RenderVariable(*variable, console.Palette()))
	return nil
}

func RunUpdateVar(args UpdateVarArgs) error {
	variableService, err := container.GetVariableService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	input := services.UpdateVariableConfig{}
	if args.NewName != "" {
		input.NewName = &args.NewName
	}
	if args.NewValue != "" {
		input.Value = &args.NewValue
	}

	variable, err := variableService.UpdateVariable(args.CurrentName, stringOrNil(args.Profile), input)
	if err != nil {
		return fmt.Errorf("updating variable: %w", err)
	}
	console.Print(presenters.RenderVariableUpdated(*variable, console.Palette()))
	return nil
}

func RunListVar(args ListVarArgs) error {
	variableService, err := container.GetVariableService()
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

	variables, err := variableService.ListVariables(args.Order, args.Tags, limit, stringOrNil(args.Profile))
	if err != nil {
		return fmt.Errorf("listing variables: %w", err)
	}

	rendered := presenters.RenderVariableList(variables, "Variables", args.Fields, console.Palette())
	rowCount := len(variables)
	return console.PrintPaged(rendered, &rowCount, args.Page)
}

func RunSearchVar(args SearchParams) error {
	variableService, err := container.GetVariableService()
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

	variables, err := variableService.SearchVariables(args.Term, args.SearchIn, limit, stringOrNil(args.Profile))
	if err != nil {
		return fmt.Errorf("searching variables: %w", err)
	}

	rendered := presenters.RenderVariableList(variables, "Variable search Results", args.Fields, console.Palette())
	rowCount := len(variables)
	return console.PrintPaged(rendered, &rowCount, args.Page)
}

func RunDeleteVar(args DeleteVarArgs) error {
	variableService, err := container.GetVariableService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	profile := stringOrNil(args.Profile)
	variable, err := variableService.GetVariable(args.Name, profile)
	if err != nil {
		return fmt.Errorf("getting variable: %w", err)
	}
	if err := variableService.DeleteVariable(args.Name, profile); err != nil {
		return fmt.Errorf("deleting variable: %w", err)
	}
	console.Print(presenters.RenderVariableDeleted(*variable, console.Palette()))
	return nil
}

func RunTagVar(args TagVarArgs) error {
	variableService, err := container.GetVariableService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	result, err := variableService.AddTags(args.Name, args.TagNames, stringOrNil(args.Profile))
	if err != nil {
		return fmt.Errorf("adding tags: %w", err)
	}
	if len(result.Added) > 0 {
		console.Success(fmt.Sprintf("Added tags to '%s': %s", args.Name, strings.Join(result.Added, ", ")))
	}
	if (len(result.Existing)) > 0 {
		console.Warning(fmt.Sprintf("Already tagged with: %s", strings.Join(result.Existing, ", ")))
	}
	return nil
}

func RunUntagVar(args TagVarArgs) error {
	variableService, err := container.GetVariableService()
	if err != nil {
		return fmt.Errorf("getting command service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}
	result, err := variableService.RemoveTags(args.Name, args.TagNames, stringOrNil(args.Profile))
	if err != nil {
		return fmt.Errorf("removing tags: %w", err)
	}
	if len(result.Removed) > 0 {
		console.Success(fmt.Sprintf("Removed tags from '%s': %s", args.Name, strings.Join(result.Removed, ", ")))
	}
	if len(result.NotAttached) > 0 {
		console.Muted(fmt.Sprintf("Not attached: %s", strings.Join(result.NotAttached, ", ")))
	}
	return nil
}
