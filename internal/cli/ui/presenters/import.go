package presenters

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

func RenderImportResult(result services.ImportResult, palette ui.Palette) string {
	header := ui.KVTable([][2]string{{"Imported from: ", result.Path}, {"Imported into profile: ", result.Profile}}, palette)
	parts := []string{header}

	if len(result.CommandsCreated) > 0 {
		list := ui.BulletList(result.CommandsCreated, palette)
		caption := fmt.Sprintf("Commands created - %d", len(result.CommandsCreated))
		section := ui.Banner(caption, list, palette.Success, palette)
		parts = append(parts, section)
	}

	if len(result.CommandsSkipped) > 0 {
		list := ui.BulletList(result.CommandsSkipped, palette)
		caption := fmt.Sprintf("Commands skipped - %d", len(result.CommandsSkipped))
		section := ui.Banner(caption, list, palette.Muted, palette)
		parts = append(parts, section)
	}

	if len(result.CommandsOverwritten) > 0 {
		list := ui.BulletList(result.CommandsOverwritten, palette)
		caption := fmt.Sprintf("Commands overwritten - %d", len(result.CommandsOverwritten))
		section := ui.Banner(caption, list, palette.Info, palette)
		parts = append(parts, section)
	}

	if len(result.VariablesCreated) > 0 {
		list := ui.BulletList(result.VariablesCreated, palette)
		caption := fmt.Sprintf("Variables created - %d", len(result.VariablesCreated))
		section := ui.Banner(caption, list, palette.Success, palette)
		parts = append(parts, section)
	}

	if len(result.VariablesSkipped) > 0 {
		list := ui.BulletList(result.VariablesSkipped, palette)
		caption := fmt.Sprintf("Variables skipped - %d", len(result.VariablesSkipped))
		section := ui.Banner(caption, list, palette.Muted, palette)
		parts = append(parts, section)
	}

	if len(result.VariablesOverwritten) > 0 {
		list := ui.BulletList(result.VariablesOverwritten, palette)
		caption := fmt.Sprintf("Variables overwritten - %d", len(result.VariablesOverwritten))
		section := ui.Banner(caption, list, palette.Info, palette)
		parts = append(parts, section)
	}

	joinedParts := ui.Stack(parts...)
	return ui.SectionSuccess("Import complete", joinedParts, palette)
}
