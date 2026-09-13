package settings

import (
	"reflect"
	"testing"
)

func TestDefaultUIColors(t *testing.T) {
	got := DefaultUIColors()
	want := UIColors{
		Title:        "bold",
		Subtitle:     "dim",
		Muted:        "dim",
		Border:       "dim",
		PanelTitle:   "bold",
		TableHeader:  "bold",
		Caption:      "dim",
		KVKey:        "dim",
		KVValue:      "",
		Success:      "green",
		Info:         "cyan",
		Warning:      "yellow",
		Error:        "red",
		Debug:        "dim blue",
		Code:         "cyan",
		CodeInline:   "cyan",
		CodeBlock:    "dim cyan",
		EntityName:   "bold",
		EntityID:     "magenta",
		EntityCount:  "bold",
		EntityTime:   "dim",
		TagPill:      "bold white on dark_green",
		TagPillMuted: "white on grey23",
		RunCommand:   "cyan",
		RunStdout:    "",
		RunStderr:    "bold red",
		TraceKind:    "dim",
		TraceKey:     "magenta",
		TraceValue:   "purple3",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DefaultUIColors() = %+v, want %+v", got, want)
	}
}

func TestDefaultDefaultFields(t *testing.T) {
	got := DefaultDefaultFields()
	want := DefaultFields{
		CommandOutput:        []string{"alias", "template", "description"},
		CommandSearch:        []string{"alias", "template", "description"},
		CommandListLimit:     25,
		CommandDefaultOrder:  "alias",
		VariableOutput:       []string{"name", "value"},
		VariableSearch:       []string{"name", "value"},
		VariableListLimit:    25,
		VariableDefaultOrder: "name",
		TagOutput:            []string{"name", "description"},
		TagSearch:            []string{"name", "description"},
		TagListLimit:         25,
		TagDefaultOrder:      "name",
		ProfileOutput:        []string{"name", "description", "date_created"},
		ProfileSearch:        []string{"name", "description"},
		ProfileListLimit:     20,
		ProfileDefaultOrder:  "name",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DefaultDefaultFields() = %+v, want %+v", got, want)
	}
}

func TestFieldAliasesAliasMap(t *testing.T) {
	t.Run("maps default aliases to lowercase field names", func(t *testing.T) {
		aliases := DefaultFieldAliases()
		m := aliases.AliasMap()

		cases := map[string]string{
			"a":       "alias",
			"al":      "template", // sanity check below overrides this
			"t":       "template",
			"temp":    "template",
			"d":       "description",
			"desc":    "description",
			"dc":      "date_created",
			"created": "date_created",
			"lu":      "last_updated",
			"updated": "last_updated",
			"u":       "used",
			"luse":    "last_used",
		}
		// "al" is actually an alias of "alias", not "template" - fix expectation.
		cases["al"] = "alias"

		for alias, want := range cases {
			if got := m[alias]; got != want {
				t.Fatalf("AliasMap()[%q] = %q, want %q", alias, got, want)
			}
		}
	})

	t.Run("lowercases mixed-case aliases", func(t *testing.T) {
		f := FieldAliases{AliasMapping: map[string][]string{
			"alias": {"A", "Al"},
		}}
		m := f.AliasMap()
		if m["a"] != "alias" || m["al"] != "alias" {
			t.Fatalf("AliasMap() = %v, want lowercase keys mapped to alias", m)
		}
		if _, ok := m["A"]; ok {
			t.Fatalf("AliasMap() should not retain original casing as a key")
		}
	})

	t.Run("empty mapping produces empty map", func(t *testing.T) {
		f := FieldAliases{}
		m := f.AliasMap()
		if len(m) != 0 {
			t.Fatalf("AliasMap() = %v, want empty map", m)
		}
	})

	t.Run("unknown alias is absent", func(t *testing.T) {
		m := DefaultFieldAliases().AliasMap()
		if _, ok := m["nonexistent"]; ok {
			t.Fatalf("AliasMap() contains unexpected key %q", "nonexistent")
		}
	})
}

func TestDefaultUISettings(t *testing.T) {
	got := DefaultUISettings()
	if !got.UseColor {
		t.Fatalf("DefaultUISettings().UseColor = false, want true")
	}
	if !reflect.DeepEqual(got.Colors, DefaultUIColors()) {
		t.Fatalf("DefaultUISettings().Colors = %+v, want DefaultUIColors()", got.Colors)
	}
	if got.PagerMode != "auto" || got.PagerMinRows != 25 || got.PagerPageStep != 15 || got.PagerLineStep != 1 {
		t.Fatalf("DefaultUISettings() = %+v, unexpected pager defaults", got)
	}
}

func TestDefaultExecutionSettings(t *testing.T) {
	got := DefaultExecutionSettings()
	want := ExecutionSettings{DefaultShell: "auto", CaptureOutput: false, DefaultVerbose: false}
	if got != want {
		t.Fatalf("DefaultExecutionSettings() = %+v, want %+v", got, want)
	}
}

func TestDefaultLoggingFileSettings(t *testing.T) {
	got := DefaultLoggingFileSettings()
	want := LoggingFileSettings{Enabled: false, Level: "INFO", MaxSizeMB: 10, Backups: 3}
	if got != want {
		t.Fatalf("DefaultLoggingFileSettings() = %+v, want %+v", got, want)
	}
}

func TestDefaultLoggingSettings(t *testing.T) {
	got := DefaultLoggingSettings()
	if got.ConsoleLevel != "WARNING" {
		t.Fatalf("DefaultLoggingSettings().ConsoleLevel = %q, want WARNING", got.ConsoleLevel)
	}
	if got.File != DefaultLoggingFileSettings() {
		t.Fatalf("DefaultLoggingSettings().File = %+v, want DefaultLoggingFileSettings()", got.File)
	}
}

func TestDefaultHistorySettings(t *testing.T) {
	got := DefaultHistorySettings()
	want := HistorySettings{Enabled: true, LimitPerCommand: 100}
	if got != want {
		t.Fatalf("DefaultHistorySettings() = %+v, want %+v", got, want)
	}
}

func TestDefaultSettings(t *testing.T) {
	got := DefaultSettings()
	want := Settings{
		UI:                DefaultUISettings(),
		ExecutionSettings: DefaultExecutionSettings(),
		DefaultFields:     DefaultDefaultFields(),
		FieldAliases:      DefaultFieldAliases(),
		Logging:           DefaultLoggingSettings(),
		History:           DefaultHistorySettings(),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("DefaultSettings() = %+v, want %+v", got, want)
	}
}
