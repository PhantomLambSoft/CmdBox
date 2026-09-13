package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

const (
	vendor  = "PhantomLamb"
	appName = "CmdBox"
)

func GetAppDataDir() (string, error) {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "Library", "Application Support")
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}

	dir := filepath.Join(base, vendor, appName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func GetLogDir() (string, error) {
	appDataDir, err := GetAppDataDir()
	if err != nil {
		return "", err
	}
	logDir := filepath.Join(appDataDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return "", err
	}
	return logDir, nil
}

func GetLogFilePath() (string, error) {
	logDir, err := GetLogDir()
	if err != nil {
		return "", err
	}
	logFilePath := filepath.Join(logDir, "cmdbox.log")
	return logFilePath, nil
}
