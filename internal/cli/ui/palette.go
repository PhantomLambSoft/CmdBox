package ui

import (
	"image/color"

	"charm.land/lipgloss/v2"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

type Palette struct {
	Title        lipgloss.Style
	Subtitle     lipgloss.Style
	Muted        lipgloss.Style
	PanelTitle   lipgloss.Style
	Border       color.Color
	TableHeader  lipgloss.Style
	Caption      lipgloss.Style
	KVKey        lipgloss.Style
	KVValue      lipgloss.Style
	Success      lipgloss.Style
	Info         lipgloss.Style
	Warning      lipgloss.Style
	Error        lipgloss.Style
	Debug        lipgloss.Style
	Code         lipgloss.Style
	CodeInline   lipgloss.Style
	CodeBlock    lipgloss.Style
	EntityName   lipgloss.Style
	EntityID     lipgloss.Style
	EntityCount  lipgloss.Style
	EntityTime   lipgloss.Style
	TagPill      lipgloss.Style
	TagPillMuted lipgloss.Style
	RunCommand   lipgloss.Style
	RunStdout    lipgloss.Style
	RunStderr    lipgloss.Style
	TraceKind    lipgloss.Style
	TraceKey     lipgloss.Style
	TraceValue   lipgloss.Style
}

func styleOf(c settings.ColorStyle) lipgloss.Style {
	s := lipgloss.NewStyle()
	if c.Fg != "" {
		s = s.Foreground(lipgloss.Color(c.Fg))
	}
	if c.Bg != "" {
		s = s.Background(lipgloss.Color(c.Bg))
	}
	return s.Bold(c.Bold).Italic(c.Italic).Faint(c.Faint).Underline(c.Underline)
}

func colorOf(c settings.ColorStyle) color.Color {
	if c.Fg == "" {
		return lipgloss.NoColor{}
	}
	return lipgloss.Color(c.Fg)
}

func NewPalette(c settings.UIColors) Palette {
	return Palette{
		Title:        styleOf(c.Title),
		Subtitle:     styleOf(c.Subtitle),
		Muted:        styleOf(c.Muted),
		Border:       colorOf(c.Border),
		PanelTitle:   styleOf(c.PanelTitle),
		TableHeader:  styleOf(c.TableHeader),
		Caption:      styleOf(c.Caption),
		KVKey:        styleOf(c.KVKey),
		KVValue:      styleOf(c.KVValue),
		Success:      styleOf(c.Success),
		Info:         styleOf(c.Info),
		Warning:      styleOf(c.Warning),
		Error:        styleOf(c.Error),
		Debug:        styleOf(c.Debug),
		Code:         styleOf(c.Code),
		CodeInline:   styleOf(c.CodeInline),
		CodeBlock:    styleOf(c.CodeBlock),
		EntityName:   styleOf(c.EntityName),
		EntityID:     styleOf(c.EntityID),
		EntityCount:  styleOf(c.EntityCount),
		EntityTime:   styleOf(c.EntityTime),
		TagPill:      styleOf(c.TagPill),
		TagPillMuted: styleOf(c.TagPillMuted),
		RunCommand:   styleOf(c.RunCommand),
		RunStdout:    styleOf(c.RunStdout),
		RunStderr:    styleOf(c.RunStderr),
		TraceKind:    styleOf(c.TraceKind),
		TraceKey:     styleOf(c.TraceKey),
		TraceValue:   styleOf(c.TraceValue),
	}
}
