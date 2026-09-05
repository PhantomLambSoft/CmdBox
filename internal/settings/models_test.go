package settings

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/pelletier/go-toml/v2"
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
	want := LoggingFileSettings{Enabled: false, Level: "INFO", MaxBytes: 1_000_000, Backups: 3}
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

// --- ByteSize ---

type byteSizeHolder struct {
	MaxBytes ByteSize `toml:"max_bytes"`
}

// unmarshalByteSizeHolder decodes into h the way Repository.Load and Service.Edit do: with
// EnableUnmarshalerInterface so ByteSize's UnmarshalTOML hook is actually invoked.
func unmarshalByteSizeHolder(t *testing.T, data string, h *byteSizeHolder) error {
	t.Helper()
	return toml.NewDecoder(strings.NewReader(data)).EnableUnmarshalerInterface().Decode(h)
}

// marshalByteSizeHolder encodes h the way Repository.Save and Service.Edit do: with
// EnableMarshalerInterface so ByteSize's MarshalTOML hook is actually invoked.
func marshalByteSizeHolder(t *testing.T, h byteSizeHolder) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).EnableMarshalerInterface().Encode(h); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	return buf.Bytes()
}

func TestByteSizeUnmarshalTOML(t *testing.T) {
	t.Run("unmarshals a plain integer", func(t *testing.T) {
		var h byteSizeHolder
		if err := unmarshalByteSizeHolder(t, "max_bytes = 2048", &h); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if h.MaxBytes != 2048 {
			t.Fatalf("MaxBytes = %d, want 2048", h.MaxBytes)
		}
	})

	t.Run("unmarshals a human-readable string via ParseByteSize", func(t *testing.T) {
		var h byteSizeHolder
		if err := unmarshalByteSizeHolder(t, `max_bytes = "5mb"`, &h); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if h.MaxBytes != ByteSize(5*1024*1024) {
			t.Fatalf("MaxBytes = %d, want %d", h.MaxBytes, 5*1024*1024)
		}
	})

	t.Run("unmarshals a single-quoted literal string", func(t *testing.T) {
		var h byteSizeHolder
		if err := unmarshalByteSizeHolder(t, "max_bytes = '2kb'", &h); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if h.MaxBytes != ByteSize(2*1024) {
			t.Fatalf("MaxBytes = %d, want %d", h.MaxBytes, 2*1024)
		}
	})

	t.Run("invalid string returns an error", func(t *testing.T) {
		var h byteSizeHolder
		if err := unmarshalByteSizeHolder(t, `max_bytes = "not-a-size"`, &h); err == nil {
			t.Fatalf("Decode() error = nil, want error")
		}
	})

	t.Run("UnmarshalTOML called directly parses a human-readable string", func(t *testing.T) {
		var b ByteSize
		if err := b.UnmarshalTOML([]byte(`"5mb"`)); err != nil {
			t.Fatalf("UnmarshalTOML() error = %v", err)
		}
		if b != ByteSize(5*1024*1024) {
			t.Fatalf("UnmarshalTOML() = %d, want %d", b, 5*1024*1024)
		}
	})

	t.Run("UnmarshalTOML called directly propagates invalid string errors", func(t *testing.T) {
		var b ByteSize
		if err := b.UnmarshalTOML([]byte(`"not-a-size"`)); err == nil {
			t.Fatalf("UnmarshalTOML() error = nil, want error")
		}
	})
}

func TestByteSizeMarshalTOML(t *testing.T) {
	t.Run("round-trips through the encoder and decoder", func(t *testing.T) {
		h := byteSizeHolder{MaxBytes: 500}
		data := marshalByteSizeHolder(t, h)

		var roundTripped byteSizeHolder
		if err := unmarshalByteSizeHolder(t, string(data), &roundTripped); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if roundTripped.MaxBytes != 500 {
			t.Fatalf("round-tripped MaxBytes = %d, want 500", roundTripped.MaxBytes)
		}
	})

	t.Run("MarshalTOML called directly returns the plain integer", func(t *testing.T) {
		b := ByteSize(500)
		data, err := b.MarshalTOML()
		if err != nil {
			t.Fatalf("MarshalTOML() error = %v", err)
		}
		if string(data) != "500" {
			t.Fatalf("MarshalTOML() = %q, want %q", data, "500")
		}
	})
}
