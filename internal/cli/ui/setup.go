package ui

import (
	"path/filepath"

	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

func Setup(cfg settings.UISettings, dataDir string) (*Console, []string, error) {
	theme, err := settings.LoadTheme(cfg.Theme, userThemeDir(dataDir))
	if err != nil {
		return nil, nil, err
	}
	colors, unknown := settings.MergeColorOverrides(theme.Colors, cfg.ColorOverrides)
	palette := NewPalette(colors)
	console := NewConsole(palette, cfg.UseColor, cfg.PagerMode, cfg.PagerMinRows, cfg.PagerPageStep, cfg.PagerLineStep)
	return console, unknown, nil
}

func userThemeDir(dataDir string) string {
	return filepath.Join(dataDir, "themes")
}
