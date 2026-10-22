package ui

import (
	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

func paint(dst, src lipgloss.Style) lipgloss.Style {
	return dst.
		Foreground(src.GetForeground()).
		Background(src.GetBackground()).
		Bold(src.GetBold()).
		Italic(src.GetItalic()).
		Faint(src.GetFaint()).
		Underline(src.GetUnderline())
}

func HuhTheme(p Palette) huh.Theme {
	return huh.ThemeFunc(func(isDark bool) *huh.Styles {
		s := huh.ThemeBase(isDark)

		f := &s.Focused
		f.Base = f.Base.BorderForeground(p.Border)
		f.Title = paint(f.Title, p.Title)
		f.Description = paint(f.Description, p.Muted)
		f.ErrorIndicator = paint(f.ErrorIndicator, p.Error)
		f.ErrorMessage = paint(f.ErrorMessage, p.Error)

		f.TextInput.Prompt = paint(f.TextInput.Prompt, p.Info)
		f.TextInput.Cursor = paint(f.TextInput.Cursor, p.Info)
		f.TextInput.Placeholder = paint(f.TextInput.Placeholder, p.Muted)

		f.SelectSelector = paint(f.SelectSelector, p.Info)
		f.SelectedOption = paint(f.SelectedOption, p.Success)

		s.Blurred = s.Focused
		s.Blurred.Base = s.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
		return s
	})
}
