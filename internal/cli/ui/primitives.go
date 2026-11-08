package ui

import (
	"fmt"
	"image/color"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
	"github.com/charmbracelet/x/term"
)

func GetTerminalWidth() int {
	w, _, err := term.GetSize(os.Stdout.Fd())
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

func GetElementWidth(element string) int {
	return lipgloss.Width(element)
}

func TablePanel(title string, headers []string, rows [][]string, caption string, palette Palette) string {
	tWidth := GetTerminalWidth()
	borderStyle := lipgloss.NewStyle().Foreground(palette.Border)
	t := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.RoundedBorder()).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(false).
		BorderHeader(true).
		BorderColumn(true).
		BorderStyle(borderStyle)
	body := t.Render()

	tableWidth := GetElementWidth(body)
	if tableWidth > tWidth {
		body = t.Width(tWidth - 4).Render()
	}

	capStyle := palette.Caption.PaddingTop(1).Align(lipgloss.Center)
	renderedCaption := capStyle.Render(caption)

	panel := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(palette.Border).
		Padding(0, 1, 0, 1)

	titleStyle := palette.PanelTitle.PaddingBottom(1).Align(lipgloss.Center)

	return panel.Render(lipgloss.JoinVertical(lipgloss.Center, titleStyle.Render(title), body, renderedCaption))
}

func KVTable(rows [][2]string, palette Palette) string {
	data := make([][]string, len(rows))
	for i, r := range rows {
		data[i] = []string{palette.KVKey.Render(r[0]), r[1]}
	}

	kvTable := table.New().
		Rows(data...).
		Border(lipgloss.HiddenBorder()).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 0 {
				return lipgloss.NewStyle().Align(lipgloss.Right)
			}
			return lipgloss.NewStyle().Align(lipgloss.Left)
		})

	renderedTable := kvTable.Render()

	terminalWidth := GetTerminalWidth()
	tableWidth := GetElementWidth(renderedTable)

	if tableWidth < terminalWidth {
		return renderedTable
	}
	kvTable = kvTable.Width(terminalWidth - 4)
	return kvTable.Render()
}

func KVPanel(title string, rows [][2]string, palette Palette) string {
	return Section(title, KVTable(rows, palette), "", palette.Border, palette.PanelTitle)
}

func TagPill(name string, style lipgloss.Style) string {
	return style.Padding(0, 1).Render(name)
}

func TagBlock(tags []string, palette Palette) string {
	out := ""
	for i, t := range tags {
		if i > 0 {
			out += " "
		}
		out += TagPill(t, palette.TagPill)
	}
	return out
}

func Banner(title, subtitle string, statusStyle lipgloss.Style, palette Palette) string {
	content := statusStyle.Render(title)
	if subtitle != "" {
		content += "\n" + palette.Subtitle.Render(subtitle)
	}
	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(statusStyle.GetForeground()).
		Width(GetTerminalWidth()-2).
		Padding(0, 1).
		Render(content)
}

func BulletList(items []string, palette Palette) string {
	if len(items) == 0 {
		return palette.Muted.Render("(none)")
	}
	lines := make([]string, len(items))
	for i, item := range items {
		lines[i] = palette.Muted.Render("• ") + palette.KVValue.Render(item)
	}
	return strings.Join(lines, "\n")
}

func CodeInline(text string, palette Palette) string {
	return palette.CodeInline.Render(text)
}

func CodeBlock(text, title string, palette Palette) string {
	body := lipgloss.NewStyle().Padding(1, 1).Render(palette.CodeBlock.Render(text))
	if title == "" {
		return lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(palette.Border).
			Width(GetTerminalWidth() - 2).
			Render(body)
	}
	return Section(title, body, "", palette.Border, palette.PanelTitle)
}

func Section(title, body, caption string, borderColor color.Color, titleStyle lipgloss.Style) string {
	panel := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Align(lipgloss.Center)

	content := titleStyle.Render(title) + "\n" + body
	if caption != "" {
		content += "\n" + caption
	}

	renderedPanel := panel.Render(content)
	panelWidth := GetElementWidth(renderedPanel)
	terminalWidth := GetTerminalWidth()
	if panelWidth < terminalWidth {
		return renderedPanel
	}
	return panel.Width(terminalWidth - 4).Render(content)
}

func SectionSuccess(title, body string, palette Palette) string {
	return Section(title, body, "", palette.Success.GetForeground(), palette.Success)
}

func SectionError(title, body string, palette Palette) string {
	return Section(title, body, "", palette.Error.GetForeground(), palette.Error)
}

func Divider(title string, width int, palette Palette) string {
	if title == "" {
		return palette.Muted.Render(strings.Repeat("─", width))
	}
	label := " " + title + " "
	side := (width - len(label)) / 2
	if side < 0 {
		side = 0
	}
	line := strings.Repeat("─", side) + label + strings.Repeat("─", width-side-len(label))
	return palette.Muted.Render(line)
}

func Spacer(n int) string {
	if n < 1 {
		return ""
	}
	return strings.Repeat("\n", n)
}

func Centered(s string, width int) string {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center).Render(s)
}

func Stack(parts ...string) string {
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func StackHorz(parts ...string) string {
	return lipgloss.JoinHorizontal(lipgloss.Top, parts...)
}

func SummaryCounts(labels []string, counts []int, palette Palette) string {
	parts := make([]string, len(labels))
	for i, label := range labels {
		parts[i] = palette.Muted.Render(label+":") + " " + palette.Title.Render(fmt.Sprintf("%d", counts[i]))
	}
	return strings.Join(parts, palette.Muted.Render("  "))
}

func Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return "1 " + singular
	}
	if plural == "" {
		plural = singular + "s"
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// JoinWithAnd renders a comma-separated list with "and" between the last two items
func JoinWithAnd(items []string) string {
	return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
}
