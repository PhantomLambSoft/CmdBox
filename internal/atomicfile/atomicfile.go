package atomicfile

import (
	"fmt"
	"os"
	"path/filepath"
)

// WriteFile writes the given content to the specified destination file.
// It ensures atomic writes by using a temporary file during the write process.
// The function creates all necessary parent directories for the destination file.
func WriteFile(dest, content string) (err error) {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating directory %q: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, filepath.Base(dest)+".*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmp.Name()

	defer func() {
		if err != nil {
			os.Remove(tmpPath)
		}
	}()

	if _, err = tmp.WriteString(content); err != nil {
		tmp.Close()
		return fmt.Errorf("writing to temp file: %w", err)
	}

	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing temp file: %w", err)
	}

	if err = tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err = os.Rename(tmpPath, dest); err != nil {
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}
