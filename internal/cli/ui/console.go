package ui

import (
	"fmt"
	"io"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/colorprofile"
)

type Console struct {
	out      io.Writer
	isTTY    bool
	palette  Palette
	useColor bool

	pagerMode     string
	pagerMinRows  int
	pagerPageStep int
	pagerLineStep int
}

func NewConsole(
	palette Palette,
	useColor bool,
	pagerMode string,
	pagerMinRows, PagerPageStep, pagerLineStep int,
) *Console {
	detected := colorprofile.Detect(os.Stdout, os.Environ())

	w := colorprofile.NewWriter(os.Stdout, os.Environ())
	if !useColor {
		w.Profile = colorprofile.Ascii
	}

	return &Console{
		out:           w,
		isTTY:         detected != colorprofile.NoTTY,
		palette:       palette,
		useColor:      useColor,
		pagerMode:     pagerMode,
		pagerMinRows:  pagerMinRows,
		pagerPageStep: PagerPageStep,
		pagerLineStep: pagerLineStep,
	}
}

func (c *Console) Palette() Palette {
	return c.palette
}

func (c *Console) UseColor() bool {
	return c.useColor
}

func (c *Console) Print(s string) {
	fmt.Fprintln(c.out, s)
}

func (c *Console) Success(msg string) {
	c.Print(c.palette.Success.Render(msg))
}

func (c *Console) Warning(msg string) {
	c.Print(c.palette.Warning.Render(msg))
}

func (c *Console) Error(msg string) {
	c.Print(c.palette.Error.Render(msg))
}

func (c *Console) Info(msg string) {
	c.Print(c.palette.Info.Render(msg))
}

func (c *Console) Muted(msg string) {
	c.Print(c.palette.Muted.Render(msg))
}

func (c *Console) Debug(msg string) {
	c.Print(c.palette.Debug.Render(msg))
}

func (c *Console) PrintPaged(rendered string, rowCount *int, force *bool) error {
	if !c.shouldPage(rowCount, force) {
		c.Print(rendered)
		return nil
	}
	return RunPager(rendered, c.palette, c.pagerPageStep, c.pagerLineStep)
}

func (c *Console) shouldPage(rowCount *int, force *bool) bool {
	if force != nil {
		return *force
	}
	if !c.isTTY {
		return false // never page redirected output
	}
	switch c.pagerMode {
	case "never":
		return false
	case "always":
		return true
	}
	if rowCount == nil {
		return true
	}
	return *rowCount > c.pagerMinRows
}

func (c *Console) DumpStyles() {
	label := func(s string) string {
		return lipgloss.NewStyle().Width(16).Render(s)
	}

	entries := []struct {
		name  string
		style lipgloss.Style
	}{
		{"title", c.palette.Title},
		{"subtitle", c.palette.Subtitle},
		{"muted", c.palette.Muted},
		{"panel_title", c.palette.PanelTitle},
		{"table_header", c.palette.TableHeader},
		{"caption", c.palette.Caption},
		{"kv_key", c.palette.KVKey},
		{"kv_value", c.palette.KVValue},
		{"success", c.palette.Success},
		{"info", c.palette.Info},
		{"warning", c.palette.Warning},
		{"error", c.palette.Error},
		{"debug", c.palette.Debug},
		{"code", c.palette.Code},
		{"code_inline", c.palette.CodeInline},
		{"code_block", c.palette.CodeBlock},
		{"entity_name", c.palette.EntityName},
		{"entity_id", c.palette.EntityID},
		{"entity_count", c.palette.EntityCount},
		{"entity_time", c.palette.EntityTime},
		{"tag_pill", c.palette.TagPill},
		{"tag_pill_muted", c.palette.TagPillMuted},
		{"run_command", c.palette.RunCommand},
		{"run_stdout", c.palette.RunStdout},
		{"run_stderr", c.palette.RunStderr},
		{"trace_kind", c.palette.TraceKind},
		{"trace_key", c.palette.TraceKey},
		{"trace_value", c.palette.TraceValue},
	}

	for _, e := range entries {
		c.Print(label(e.name) + " " + e.style.Render("sample text — "+e.name))
	}

	// Border is a bare color.Color, not a Style, so it needs its own line
	// rather than fitting the entries table above.
	borderStyle := lipgloss.NewStyle().Foreground(c.palette.Border)
	c.Print(label("border") + " " + borderStyle.Render("sample text — border"))
}
