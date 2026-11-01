package handlers

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/presenters"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

type RunExportCmdArgs struct {
	Aliases    []string
	Tag        string
	Flatten    bool
	Output     string
	CmdProfile string
	VarProfile string
}

type RunExportVarArgs struct {
	Names      []string
	Tag        string
	Flatten    bool
	Output     string
	VarProfile string
}

type RunExportAllArgs struct {
	Flatten    bool
	Output     string
	CmdProfile string
	VarProfile string
}

func RunExportCmds(args RunExportCmdArgs) error {
	exportService, err := container.GetExportService()
	if err != nil {
		return fmt.Errorf("getting export service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	config := services.ExportCmdConfig{
		Aliases:        args.Aliases,
		Tag:            stringOrNil(args.Tag),
		Flatten:        args.Flatten,
		Output:         stringOrNil(args.Output),
		CmdProfileName: stringOrNil(args.CmdProfile),
		VarProfileName: stringOrNil(args.VarProfile),
	}

	result, err := exportService.ExportCmds(config)

	console.Print(presenters.RenderExportResult(result, console.Palette()))
	return nil
}

func RunExportVars(args RunExportVarArgs) error {
	exportService, err := container.GetExportService()
	if err != nil {
		return fmt.Errorf("getting export service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	config := services.ExportVarConfig{
		Names:          args.Names,
		Tag:            stringOrNil(args.Tag),
		Flatten:        args.Flatten,
		Output:         stringOrNil(args.Output),
		VarProfileName: stringOrNil(args.VarProfile),
	}

	result, err := exportService.ExportVars(config)

	console.Print(presenters.RenderExportResult(result, console.Palette()))
	return nil
}

func RunExportAll(args RunExportAllArgs) error {
	exportService, err := container.GetExportService()
	if err != nil {
		return fmt.Errorf("getting export service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	config := services.ExportConfig{
		Flatten:        args.Flatten,
		Output:         stringOrNil(args.Output),
		CmdProfileName: stringOrNil(args.CmdProfile),
		VarProfileName: stringOrNil(args.VarProfile),
	}

	result, err := exportService.ExportAll(config)

	console.Print(presenters.RenderExportResult(result, console.Palette()))
	return nil
}
