package logging

import (
	"log/slog"
	"strings"

	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

const LevelCritical = slog.Level(12)

type LogConfig struct {
	ConsoleLevel slog.Level
	FileEnabled  bool
	FileLevel    slog.Level
	FilePath     string
	MaxBytes     int64
	Backups      int
}

func BuildLogConfig(settings *settings.Settings, verbose, debug bool, fileLogs *bool) LogConfig {
	return LogConfig{}
}

func GetConsoleLevel(settings *settings.Settings, verbose, debug bool) slog.Level {
	if debug {
		return slog.LevelDebug
	}
	if verbose {
		return slog.LevelInfo
	}
	return parseLevel(settings.Logging.ConsoleLevel)
}

func GetFileEnabled(settings *settings.Settings, fileLogs *bool) bool {
	if fileLogs != nil {
		return *fileLogs
	}
	return settings.Logging.File.Enabled
}

func GetFileLevel(settings *settings.Settings, verbose, debug bool) slog.Level {
	if debug {
		return slog.LevelDebug
	}
	if verbose {
		return slog.LevelInfo
	}
	return parseLevel(settings.Logging.File.Level)
}

// parseLevel converts a logging level string to a slog.Level. Defaults to LevelInfo for unrecognized input.
func parseLevel(levelStr string) slog.Level {
	if strings.EqualFold(levelStr, "CRITICAL") {
		return LevelCritical
	}
	var lvl slog.Level
	if err := lvl.UnmarshalText([]byte(levelStr)); err != nil {
		return slog.LevelInfo
	}
	return lvl
}
