package presenters

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
)

var TagColumns = []Column[models.Tag]{
	{Key: "name", Header: "Name", Render: func(t models.Tag, p ui.Palette) string {
		return p.EntityName.Render(t.Name)
	}},
	{Key: "description", Header: "Description", Render: func(t models.Tag, p ui.Palette) string {
		return deref(t.Description)
	}},
	{Key: "date_created", Header: "Created", Render: func(c models.Tag, p ui.Palette) string {
		return p.EntityTime.Render(formatTime(&c.CreatedAt))
	}},
	{Key: "last_updated", Header: "Updated", Render: func(c models.Tag, p ui.Palette) string {
		return p.EntityTime.Render(formatTime(&c.UpdatedAt))
	}},
}

var DefaultTagFields = []string{"name", "description"}

func RenderTag(tag models.Tag, palette ui.Palette) string {
	return renderDetail(tag, TagColumns, palette)
}

func RenderTagCreated(tag models.Tag, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Tag '%s' created", tag.Name), RenderTag(tag, palette), palette)
}

func RenderTagList(tags []models.Tag, title string, fields []string, palette ui.Palette) string {
	var cols []Column[models.Tag]
	if len(fields) == 0 {
		cols = selectColumns(TagColumns, DefaultTagFields)
	} else {
		cols = selectColumns(TagColumns, fields)
	}
	headers, rows := RenderTable(tags, cols, palette)
	caption := ui.Pluralize(len(rows), "tag found", "tags found")
	return ui.TablePanel(title, headers, rows, caption, palette)
}

func RenderTagUpdated(tag models.Tag, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Tag '%s' updated", tag.Name), RenderTag(tag, palette), palette)
}

func RenderTagDeleted(tag models.Tag, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Tag '%s' deleted", tag.Name), RenderTag(tag, palette), palette)
}
