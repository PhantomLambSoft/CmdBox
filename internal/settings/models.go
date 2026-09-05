package settings

import (
	"strconv"
	"strings"
)

type UIColors struct {
	Title        string `toml:"title"`
	Subtitle     string `toml:"subtitle"`
	Muted        string `toml:"muted"`
	Border       string `toml:"border"`
	PanelTitle   string `toml:"panel_title"`
	TableHeader  string `toml:"table_header"`
	Caption      string `toml:"caption"`
	KVKey        string `toml:"kv_key"`
	KVValue      string `toml:"kv_value"`
	Success      string `toml:"success"`
	Info         string `toml:"info"`
	Warning      string `toml:"warning"`
	Error        string `toml:"error"`
	Debug        string `toml:"debug"`
	Code         string `toml:"code"`
	CodeInline   string `toml:"code_inline"`
	CodeBlock    string `toml:"code_block"`
	EntityName   string `toml:"entity_name"`
	EntityID     string `toml:"entity_id"`
	EntityCount  string `toml:"entity_count"`
	EntityTime   string `toml:"entity_time"`
	TagPill      string `toml:"tag_pill"`
	TagPillMuted string `toml:"tag_pill_muted"`
	RunCommand   string `toml:"run_command"`
	RunStdout    string `toml:"run_stdout"`
	RunStderr    string `toml:"run_stderr"`
	TraceKind    string `toml:"trace_kind"`
	TraceKey     string `toml:"trace_key"`
	TraceValue   string `toml:"trace_value"`
}

func DefaultUIColors() UIColors {
	return UIColors{
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
}

type DefaultFields struct {
	CommandOutput        []string `toml:"command_output"`
	CommandSearch        []string `toml:"command_search"`
	CommandListLimit     int      `toml:"command_list_limit"`
	CommandDefaultOrder  string   `toml:"command_default_order"`
	VariableOutput       []string `toml:"variable_output"`
	VariableSearch       []string `toml:"variable_search"`
	VariableListLimit    int      `toml:"variable_list_limit"`
	VariableDefaultOrder string   `toml:"variable_default_order"`
	TagOutput            []string `toml:"tag_output"`
	TagSearch            []string `toml:"tag_search"`
	TagListLimit         int      `toml:"tag_list_limit"`
	TagDefaultOrder      string   `toml:"tag_default_order"`
	ProfileOutput        []string `toml:"profile_output"`
	ProfileSearch        []string `toml:"profile_search"`
	ProfileListLimit     int      `toml:"profile_list_limit"`
	ProfileDefaultOrder  string   `toml:"profile_default_order"`
}

func DefaultDefaultFields() DefaultFields {
	return DefaultFields{
		CommandOutput:    []string{"alias", "template", "description"},
		CommandSearch:    []string{"alias", "template", "description"},
		CommandListLimit: 25, CommandDefaultOrder: "alias",
		VariableOutput:    []string{"name", "value"},
		VariableSearch:    []string{"name", "value"},
		VariableListLimit: 25, VariableDefaultOrder: "name",
		TagOutput:    []string{"name", "description"},
		TagSearch:    []string{"name", "description"},
		TagListLimit: 25, TagDefaultOrder: "name",
		ProfileOutput:    []string{"name", "description", "date_created"},
		ProfileSearch:    []string{"name", "description"},
		ProfileListLimit: 20, ProfileDefaultOrder: "name",
	}
}

type FieldAliases struct {
	AliasMapping map[string][]string `toml:"alias_mapping"`
}

func DefaultFieldAliases() FieldAliases {
	return FieldAliases{
		AliasMapping: map[string][]string{
			"alias":        {"a", "al"},
			"template":     {"t", "temp"},
			"description":  {"d", "desc"},
			"date_created": {"dc", "created"},
			"last_updated": {"lu", "updated"},
			"used":         {"u"},
			"last_used":    {"luse"},
		},
	}
}

// AliasMap generates a map where keys are lowercase aliases and values are the original field names from AliasMapping.
func (f FieldAliases) AliasMap() map[string]string {
	out := make(map[string]string)
	for fieldName, aliases := range f.AliasMapping {
		for _, a := range aliases {
			out[strings.ToLower(a)] = fieldName
		}
	}
	return out
}

type UISettings struct {
	UseColor      bool     `toml:"use_color"`
	Colors        UIColors `toml:"colors"`
	PagerMode     string   `toml:"pager_mode"`
	PagerMinRows  int      `toml:"pager_min_rows"`
	PagerPageStep int      `toml:"pager_page_step"`
	PagerLineStep int      `toml:"pager_line_step"`
}

func DefaultUISettings() UISettings {
	return UISettings{
		UseColor: true, Colors: DefaultUIColors(),
		PagerMode: "auto", PagerMinRows: 25, PagerPageStep: 15, PagerLineStep: 1,
	}
}

type ExecutionSettings struct {
	DefaultShell   string `toml:"default_shell"`
	CaptureOutput  bool   `toml:"capture_output"`
	DefaultVerbose bool   `toml:"default_verbose"`
}

func DefaultExecutionSettings() ExecutionSettings {
	return ExecutionSettings{DefaultShell: "auto", CaptureOutput: false, DefaultVerbose: false}
}

type ByteSize int64

// UnmarshalTOML implements unstable.Unmarshaler. data is the raw TOML literal for the value (e.g. `2048` or
// `"5mb"`, quotes included), as delivered by a Decoder with EnableUnmarshalerInterface set.
func (b *ByteSize) UnmarshalTOML(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if n := len(raw); n >= 2 && (raw[0] == '"' || raw[0] == '\'') && raw[n-1] == raw[0] {
		if unquoted, err := strconv.Unquote(raw); err == nil {
			raw = unquoted
		} else {
			raw = raw[1 : n-1]
		}
	}

	parsed, err := ParseByteSize(raw)
	if err != nil {
		return err
	}
	*b = ByteSize(parsed)
	return nil
}

func (b ByteSize) MarshalTOML() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(b), 10)), nil
}

type LoggingFileSettings struct {
	Enabled  bool     `toml:"enabled"`
	Level    string   `toml:"level"`
	MaxBytes ByteSize `toml:"max_bytes"`
	Backups  int      `toml:"backups"`
}

func DefaultLoggingFileSettings() LoggingFileSettings {
	return LoggingFileSettings{Enabled: false, Level: "INFO", MaxBytes: 1_000_000, Backups: 3}
}

type LoggingSettings struct {
	ConsoleLevel string              `toml:"console_level"`
	File         LoggingFileSettings `toml:"file"`
}

func DefaultLoggingSettings() LoggingSettings {
	return LoggingSettings{ConsoleLevel: "WARNING", File: DefaultLoggingFileSettings()}
}

type HistorySettings struct {
	Enabled         bool `toml:"enabled"`
	LimitPerCommand int  `toml:"limit_per_command"`
}

func DefaultHistorySettings() HistorySettings {
	return HistorySettings{Enabled: true, LimitPerCommand: 100}
}

type Settings struct {
	UI                UISettings        `toml:"ui"`
	ExecutionSettings ExecutionSettings `toml:"execution_settings"`
	DefaultFields     DefaultFields     `toml:"default_fields"`
	FieldAliases      FieldAliases      `toml:"field_aliases"`
	Logging           LoggingSettings   `toml:"logging"`
	History           HistorySettings   `toml:"history"`
}

func DefaultSettings() Settings {
	return Settings{
		UI:                DefaultUISettings(),
		ExecutionSettings: DefaultExecutionSettings(),
		DefaultFields:     DefaultDefaultFields(),
		FieldAliases:      DefaultFieldAliases(),
		Logging:           DefaultLoggingSettings(),
		History:           DefaultHistorySettings(),
	}
}
