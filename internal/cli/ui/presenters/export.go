package presenters

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

func RenderExportResult(result services.ExportResult, palette ui.Palette) string {
	outputPath := ui.KVTable([][2]string{{"Saved to:", result.Path}}, palette)
	parts := []string{outputPath}

	if len(result.Commands) > 0 {
		list := ui.BulletList(result.Commands, palette)
		caption := fmt.Sprintf("Commands exported from profile '%s' - %d", result.CommandProfile, len(result.Commands))
		cmdSection := ui.Banner(caption, list, palette.Info, palette)
		parts = append(parts, cmdSection)
	}

	if len(result.TransientCommands) > 0 {
		list := ui.BulletList(result.TransientCommands, palette)
		caption := fmt.Sprintf("Transient commands exported - %d", len(result.TransientCommands))
		cmdSection := ui.Banner(caption, list, palette.Info, palette)
		parts = append(parts, cmdSection)
	}

	if len(result.Variables) > 0 {
		list := ui.BulletList(result.Variables, palette)
		caption := fmt.Sprintf("Variables exported from profile '%s' - %d", result.VariableProfile, len(result.Variables))
		varSection := ui.Banner(caption, list, palette.Info, palette)
		parts = append(parts, varSection)
	}

	if len(result.TransientVariables) > 0 {
		list := ui.BulletList(result.TransientVariables, palette)
		caption := fmt.Sprintf("Transient variables exported - %d", len(result.TransientVariables))
		varSection := ui.Banner(caption, list, palette.Info, palette)
		parts = append(parts, varSection)
	}

	if len(result.Warnings) > 0 {
		list := ui.BulletList(result.Warnings, palette)
		warnSection := ui.Banner("Warnings", list, palette.Warning, palette)
		parts = append(parts, warnSection)
	}

	joinedParts := ui.Stack(parts...)
	return ui.SectionSuccess("Export Complete", joinedParts, palette)
}
