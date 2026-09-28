package presenters

import "github.com/PhantomLambSoft/CmdBox/internal/cli/ui"

type Column[T any] struct {
	Key      string
	Header   string
	Render   func(T, ui.Palette) string // Returns final rendered text
	Optional bool
}

func selectColumns[T any](all []Column[T], fields []string) []Column[T] {
	if len(fields) == 0 {
		return all
	}
	byKey := make(map[string]Column[T], len(all))
	for _, c := range all {
		byKey[c.Key] = c
	}
	out := make([]Column[T], 0, len(fields))
	for _, f := range fields {
		if c, ok := byKey[f]; ok {
			out = append(out, c)
		}
	}
	return out
}

func RenderList[T any](title string, items []T, columns []Column[T], palette ui.Palette) string {
	headers := make([]string, len(columns))
	for i, c := range columns {
		headers[i] = c.Header
	}
	rows := make([][]string, len(items))
	for i, item := range items {
		row := make([]string, len(columns))
		for j, c := range columns {
			row[j] = c.Render(item, palette)
		}
		rows[i] = row
	}
	caption := ui.Pluralize(len(items), "item", "")
	return ui.TablePanel(title, headers, rows, caption, palette)
}

func renderDetail[T any](item T, columns []Column[T], palette ui.Palette) string {
	var rows [][2]string
	for _, c := range columns {
		v := c.Render(item, palette)
		if c.Optional && v == "" {
			continue
		}
		rows = append(rows, [2]string{c.Header, v})
	}
	return ui.KVTable(rows, palette)
}
