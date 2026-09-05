package settings

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRepositoryLoad(t *testing.T) {
	t.Run("missing file leaves the target unchanged", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		repo := NewSettingsRepository(path)

		into := DefaultSettings()
		into.History.LimitPerCommand = 42

		if err := repo.Load(&into); err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if into.History.LimitPerCommand != 42 {
			t.Fatalf("Load() modified target when file was missing: LimitPerCommand = %d, want 42", into.History.LimitPerCommand)
		}
	})

	t.Run("existing file overlays values onto the target", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		if err := writeFileForTest(t, path, "[history]\nenabled = false\nlimit_per_command = 7\n"); err != nil {
			t.Fatalf("writing fixture: %v", err)
		}
		repo := NewSettingsRepository(path)

		into := DefaultSettings()
		if err := repo.Load(&into); err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if into.History.Enabled != false || into.History.LimitPerCommand != 7 {
			t.Fatalf("Load() History = %+v, want {Enabled:false LimitPerCommand:7}", into.History)
		}
		// Fields not present in the file should retain the target's existing values.
		if into.ExecutionSettings.DefaultShell != "auto" {
			t.Fatalf("Load() ExecutionSettings.DefaultShell = %q, want unchanged default %q", into.ExecutionSettings.DefaultShell, "auto")
		}
	})

	t.Run("malformed TOML returns an error", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		if err := writeFileForTest(t, path, "not valid toml [[["); err != nil {
			t.Fatalf("writing fixture: %v", err)
		}
		repo := NewSettingsRepository(path)

		into := DefaultSettings()
		if err := repo.Load(&into); err == nil {
			t.Fatalf("Load() error = nil, want error for malformed TOML")
		}
	})
}

func TestRepositorySave(t *testing.T) {
	t.Run("writes settings that can be read back", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nested", "config.toml")
		repo := NewSettingsRepository(path)

		s := DefaultSettings()
		s.History.LimitPerCommand = 55
		s.UI.UseColor = false

		if err := repo.Save(s); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		var loaded Settings
		if err := repo.Load(&loaded); err != nil {
			t.Fatalf("Load() after Save() error = %v", err)
		}
		if loaded.History.LimitPerCommand != 55 {
			t.Fatalf("loaded.History.LimitPerCommand = %d, want 55", loaded.History.LimitPerCommand)
		}
		if loaded.UI.UseColor != false {
			t.Fatalf("loaded.UI.UseColor = %v, want false", loaded.UI.UseColor)
		}
	})

	t.Run("overwrites a previously saved file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "config.toml")
		repo := NewSettingsRepository(path)

		first := DefaultSettings()
		first.History.LimitPerCommand = 1
		if err := repo.Save(first); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		second := DefaultSettings()
		second.History.LimitPerCommand = 2
		if err := repo.Save(second); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		var loaded Settings
		if err := repo.Load(&loaded); err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if loaded.History.LimitPerCommand != 2 {
			t.Fatalf("loaded.History.LimitPerCommand = %d, want 2 (second save should win)", loaded.History.LimitPerCommand)
		}
	})
}

func writeFileForTest(t *testing.T, path, content string) error {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
