package logging

import (
	"log/slog"
	"testing"

	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

func newTestSettings(consoleLevel, fileLevel string, fileEnabled bool) *settings.Settings {
	s := settings.DefaultSettings()
	s.Logging.ConsoleLevel = consoleLevel
	s.Logging.File.Enabled = fileEnabled
	s.Logging.File.Level = fileLevel
	return &s
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  slog.Level
	}{
		{"debug upper", "DEBUG", slog.LevelDebug},
		{"debug lower", "debug", slog.LevelDebug},
		{"info", "INFO", slog.LevelInfo},
		{"warn", "WARN", slog.LevelWarn},
		{"warning alias not recognized by slog", "WARNING", slog.LevelInfo},
		{"error", "ERROR", slog.LevelError},
		{"critical upper", "CRITICAL", LevelCritical},
		{"critical lower", "critical", LevelCritical},
		{"critical mixed case", "Critical", LevelCritical},
		{"unknown defaults to info", "NOTALEVEL", slog.LevelInfo},
		{"empty defaults to info", "", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLevel(tt.input)
			if got != tt.want {
				t.Fatalf("parseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetConsoleLevel(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		debug   bool
		setting string
		want    slog.Level
	}{
		{"debug flag overrides everything", false, true, "ERROR", slog.LevelDebug},
		{"debug flag overrides verbose", true, true, "ERROR", slog.LevelDebug},
		{"verbose flag maps to info", true, false, "ERROR", slog.LevelInfo},
		{"neither flag falls back to settings", false, false, "ERROR", slog.LevelError},
		{"neither flag with critical setting", false, false, "CRITICAL", LevelCritical},
		{"neither flag with unrecognized setting defaults to info", false, false, "BOGUS", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSettings(tt.setting, "INFO", false)
			got := GetConsoleLevel(s, tt.verbose, tt.debug)
			if got != tt.want {
				t.Fatalf("GetConsoleLevel(verbose=%v, debug=%v, setting=%q) = %v, want %v",
					tt.verbose, tt.debug, tt.setting, got, tt.want)
			}
		})
	}
}

func TestGetFileLevel(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
		debug   bool
		setting string
		want    slog.Level
	}{
		{"debug flag overrides everything", false, true, "ERROR", slog.LevelDebug},
		{"debug flag overrides verbose", true, true, "ERROR", slog.LevelDebug},
		{"verbose flag maps to info", true, false, "ERROR", slog.LevelInfo},
		{"neither flag falls back to settings", false, false, "ERROR", slog.LevelError},
		{"neither flag with critical setting", false, false, "CRITICAL", LevelCritical},
		{"neither flag with unrecognized setting defaults to info", false, false, "BOGUS", slog.LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSettings("INFO", tt.setting, false)
			got := GetFileLevel(s, tt.verbose, tt.debug)
			if got != tt.want {
				t.Fatalf("GetFileLevel(verbose=%v, debug=%v, setting=%q) = %v, want %v",
					tt.verbose, tt.debug, tt.setting, got, tt.want)
			}
		})
	}
}

func TestGetFileEnabled(t *testing.T) {
	tests := []struct {
		name           string
		settingEnabled bool
		override       *bool
		want           bool
	}{
		{"nil override falls back to settings true", true, nil, true},
		{"nil override falls back to settings false", false, nil, false},
		{"override true wins over settings false", false, boolPtr(true), true},
		{"override false wins over settings true", true, boolPtr(false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newTestSettings("INFO", "INFO", tt.settingEnabled)
			got := GetFileEnabled(s, tt.override)
			if got != tt.want {
				t.Fatalf("GetFileEnabled(settingEnabled=%v, override=%v) = %v, want %v",
					tt.settingEnabled, tt.override, got, tt.want)
			}
		})
	}
}

func boolPtr(b bool) *bool {
	return &b
}
