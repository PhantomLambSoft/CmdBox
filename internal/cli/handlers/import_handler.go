package handlers

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/presenters"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

type RunImportArgs struct {
	Path      string
	Overwrite bool
	Preview   bool
	Profile   string
}

func RunImportFile(args RunImportArgs) error {
	importService, err := container.GetImportService()
	if err != nil {
		return fmt.Errorf("failed to get import service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("failed to get console: %w", err)
	}

	config := services.ImportConfig{
		ImportPath:  args.Path,
		Overwrite:   args.Overwrite,
		Preview:     args.Preview,
		ProfileName: stringOrNil(args.Profile),
	}

	result, err := importService.ImportFile(config)
	if err != nil {
		return fmt.Errorf("failed to import file: %w", err)
	}

	console.Print(presenters.RenderImportResult(result, console.Palette()))
	return nil
}
