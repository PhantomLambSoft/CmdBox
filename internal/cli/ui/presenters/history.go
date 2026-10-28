package presenters

import (
	"fmt"
	"strconv"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
)

type HistoryEntry struct {
	Index int
	models.CommandHistory
}

var HistoryColumns = []Column[HistoryEntry]{
	{Key: "index", Header: "#", Render: func(h HistoryEntry, p ui.Palette) string {
		if h.Index <= 0 {
			return p.Muted.Render("-")
		}
		return p.Muted.Render(strconv.Itoa(h.Index))
	}},
	{Key: "id", Header: "ID", Render: func(h HistoryEntry, p ui.Palette) string {
		return p.EntityID.Render(h.ID)
	}},
	{Key: "alias", Header: "Alias", Render: func(h HistoryEntry, p ui.Palette) string {
		return p.EntityName.Render(h.Alias)
	}},
	{Key: "template", Header: "Template", Render: func(h HistoryEntry, p ui.Palette) string {
		return p.CodeInline.Render(h.Template)
	}},
	{Key: "resolved", Header: "Resolved", Render: func(h HistoryEntry, p ui.Palette) string {
		return p.CodeBlock.Render(h.Resolved)
	}},
	{Key: "variables_used", Header: "Variables Used", Render: func(h HistoryEntry, p ui.Palette) string {
		return p.CodeInline.Render(deref(h.VariablesUsed))
	}},
	{Key: "exit_code", Header: "Exit", Render: func(h HistoryEntry, p ui.Palette) string {
		if h.ExitCode == nil {
			return p.Muted.Render("-")
		}
		if *h.ExitCode == 0 {
			return p.Success.Render("0")
		}
		return p.Error.Render(strconv.Itoa(*h.ExitCode))
	}},
	{Key: "ran_at", Header: "Ran At", Render: func(h HistoryEntry, p ui.Palette) string {
		return p.Muted.Render(h.RanAt.Format("2006-01-02 15:04:05"))
	}},
}

var DefaultHistoryFields = []string{"index", "id", "alias", "ran_at", "exit_code"}

func convertEntry(entry models.CommandHistory, index int) HistoryEntry {
	return HistoryEntry{Index: index + 1, CommandHistory: entry}
}

func RenderHistoryRecord(hist models.CommandHistory, palette ui.Palette) string {
	entry := convertEntry(hist, -1) // set to less than 0 so it renders as blank
	return renderDetail(entry, HistoryColumns, palette)
}

func RenderHistoryList(history []models.CommandHistory, title string, fields []string, palette ui.Palette) string {
	var cols []Column[HistoryEntry]
	if len(fields) == 0 {
		cols = selectColumns(HistoryColumns, DefaultHistoryFields)
	} else {
		cols = selectColumns(HistoryColumns, fields)
	}

	entries := make([]HistoryEntry, len(history))
	for i, e := range history {
		entries[i] = convertEntry(e, i)
	}

	headers, rows := RenderTable(entries, cols, palette)
	caption := ui.Pluralize(len(rows), "entries found", "entries found")
	return ui.TablePanel(title, headers, rows, caption, palette)
}

func RenderHistoryCleared(count int, alias *string, palette ui.Palette) string {
	var content string
	if alias != nil {
		content = fmt.Sprintf("%d history entries for alias %s cleared", count, *alias)
	} else {
		content = fmt.Sprintf("%d history entries cleared", count)
	}
	return ui.SectionSuccess("", content, palette)
}
