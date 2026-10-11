package presenters

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
)

var VariableColumns = []Column[models.Variable]{
	{Key: "name", Header: "Name", Render: func(v models.Variable, p ui.Palette) string {
		return p.EntityName.Render(v.Name)
	}},
	{Key: "value", Header: "Value", Render: func(v models.Variable, p ui.Palette) string {
		return p.CodeInline.Render(v.Value)
	}},
	{Key: "date_created", Header: "Created", Render: func(c models.Variable, p ui.Palette) string {
		return p.EntityTime.Render(formatTime(&c.CreatedAt))
	}},
	{Key: "last_updated", Header: "Updated", Render: func(c models.Variable, p ui.Palette) string {
		return p.EntityTime.Render(formatTime(&c.UpdatedAt))
	}},
	{Key: "tags", Header: "Tags", Render: func(c models.Variable, p ui.Palette) string {
		return ui.TagBlock(tagNames(c.Tags), p)
	}},
}

var DefaultVariableFields = []string{"name", "value"}

func RenderVariable(variable models.Variable, palette ui.Palette) string {
	return renderDetail(variable, VariableColumns, palette)
}

func RenderVariableCreated(variable models.Variable, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Variable '%s' created", variable.Name), RenderVariable(variable, palette), palette)
}

func RenderVariableList(vars []models.Variable, title string, fields []string, palette ui.Palette) string {
	var cols []Column[models.Variable]
	if len(fields) == 0 {
		cols = selectColumns(VariableColumns, DefaultVariableFields)
	} else {
		cols = selectColumns(VariableColumns, fields)
	}
	headers, rows := RenderTable(vars, cols, palette)
	caption := ui.Pluralize(len(rows), "variable found", "variables found")
	return ui.TablePanel(title, headers, rows, caption, palette)
}

func RenderVariableUpdated(variable models.Variable, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Variable '%s' updated", variable.Name), RenderVariable(variable, palette), palette)
}

func RenderVariableDeleted(variable models.Variable, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Variable '%s' deleted", variable.Name), RenderVariable(variable, palette), palette)
}

func RenderVariableMoved(variable models.Variable, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Variable '%s' moved", variable.Name), RenderVariable(variable, palette), palette)
}

func RenderVariableCopied(variable models.Variable, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Variable '%s' copied", variable.Name), RenderVariable(variable, palette), palette)
}
