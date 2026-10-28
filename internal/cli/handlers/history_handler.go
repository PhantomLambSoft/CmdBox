package handlers

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/presenters"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

type ListHistoryArgs struct {
	Alias    string
	HasAlias bool
	ListParams
}

type ClearHistoryArgs struct {
	Yes        bool
	Alias      string
	HasAlias   bool
	Profile    string
	HasProfile bool
}

func RunListHistory(args ListHistoryArgs) error {
	historyService, err := container.GetHistoryService()
	if err != nil {
		return fmt.Errorf("getting history service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}
	var alias *string
	if args.HasAlias {
		alias = &args.Alias
	}
	var limit *int
	if args.HasLimit {
		limit = &args.Limit
	}

	history, err := historyService.GetRecent(alias, limit, stringOrNil(args.Profile))
	if err != nil {
		return fmt.Errorf("getting recent history: %w", err)
	}

	rendered := presenters.RenderHistoryList(history, "History", args.Fields, console.Palette())
	rowCount := len(history)
	return console.PrintPaged(rendered, &rowCount, args.Page)
}

func RunShowHistory(indexOrID string, alias, profileName *string) error {
	historyService, err := container.GetHistoryService()
	if err != nil {
		return fmt.Errorf("getting history service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	entry, err := historyService.GetByRef(indexOrID, alias, profileName)
	if err != nil {
		return fmt.Errorf("getting history entry: %w", err)
	}

	renderedEntry := presenters.RenderHistoryRecord(*entry, console.Palette())
	console.Print(renderedEntry)
	return nil
}

func RunRerunHistory(indexOrID string, alias, profileName *string) error {
	historyService, err := container.GetHistoryService()
	if err != nil {
		return fmt.Errorf("getting history service: %w", err)
	}
	runService, err := container.GetRunService()
	if err != nil {
		return fmt.Errorf("getting run service: %w", err)
	}

	entry, err := historyService.GetByRef(indexOrID, alias, profileName)
	if err != nil {
		return fmt.Errorf("getting history entry: %w", err)
	}
	variables, err := historyService.GetVariables(entry)
	if err != nil {
		return fmt.Errorf("getting history entry variables: %w", err)
	}

	if _, err = runService.Run(entry.Alias, services.RuntimeOverrides{}, variables, profileName); err != nil {
		return fmt.Errorf("running history entry: %w", err)
	}
	// TODO: should this result be output like the original command run would have been?
	return nil
}

func RunLastHistory() error {
	return RunRerunHistory("1", nil, nil)
}

func RunClearHistory(args ClearHistoryArgs) error {
	if !args.Yes {
		prompter, err := container.GetPrompter()
		if err != nil {
			return fmt.Errorf("getting prompter: %w", err)
		}
		confirmed, err := prompter.PromptConfirmation("Are you sure you want to clear the history?")
		if err != nil {
			return err
		}
		if !confirmed {
			return nil
		}
	}

	historyService, err := container.GetHistoryService()
	if err != nil {
		return fmt.Errorf("getting history service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	var alias *string
	if args.HasAlias {
		alias = &args.Alias
	}
	var profile *string
	if args.HasProfile {
		profile = &args.Profile
	}

	count, err := historyService.Clear(alias, profile)
	if err != nil {
		return fmt.Errorf("clearing history: %w", err)
	}

	console.Print(presenters.RenderHistoryCleared(count, alias, console.Palette()))
	return nil
}
