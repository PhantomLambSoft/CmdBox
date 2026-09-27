package settings

import (
	"strings"
)

type UIColors struct {
	Title        ColorStyle `toml:"title"`
	Subtitle     ColorStyle `toml:"subtitle"`
	Muted        ColorStyle `toml:"muted"`
	Border       ColorStyle `toml:"border"`
	PanelTitle   ColorStyle `toml:"panel_title"`
	TableHeader  ColorStyle `toml:"table_header"`
	Caption      ColorStyle `toml:"caption"`
	KVKey        ColorStyle `toml:"kv_key"`
	KVValue      ColorStyle `toml:"kv_value"`
	Success      ColorStyle `toml:"success"`
	Info         ColorStyle `toml:"info"`
	Warning      ColorStyle `toml:"warning"`
	Error        ColorStyle `toml:"error"`
	Debug        ColorStyle `toml:"debug"`
	Code         ColorStyle `toml:"code"`
	CodeInline   ColorStyle `toml:"code_inline"`
	CodeBlock    ColorStyle `toml:"code_block"`
	EntityName   ColorStyle `toml:"entity_name"`
	EntityID     ColorStyle `toml:"entity_id"`
	EntityCount  ColorStyle `toml:"entity_count"`
	EntityTime   ColorStyle `toml:"entity_time"`
	TagPill      ColorStyle `toml:"tag_pill"`
	TagPillMuted ColorStyle `toml:"tag_pill_muted"`
	RunCommand   ColorStyle `toml:"run_command"`
	RunStdout    ColorStyle `toml:"run_stdout"`
	RunStderr    ColorStyle `toml:"run_stderr"`
	TraceKind    ColorStyle `toml:"trace_kind"`
	TraceKey     ColorStyle `toml:"trace_key"`
	TraceValue   ColorStyle `toml:"trace_value"`
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
	UseColor       bool                  `toml:"use_color"`
	Theme          string                `toml:"theme"`
	ColorOverrides map[string]ColorStyle `toml:"color_overrides,omitempty"`
	PagerMode      string                `toml:"pager_mode"`
	PagerMinRows   int                   `toml:"pager_min_rows"`
	PagerPageStep  int                   `toml:"pager_page_step"`
	PagerLineStep  int                   `toml:"pager_line_step"`
}

func DefaultUISettings() UISettings {
	return UISettings{
		UseColor:      true,
		Theme:         "default",
		PagerMode:     "auto",
		PagerMinRows:  25,
		PagerPageStep: 15,
		PagerLineStep: 1,
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

type LoggingFileSettings struct {
	Enabled   bool   `toml:"enabled"`
	Level     string `toml:"level"`
	MaxSizeMB int    `toml:"max_size_mb"`
	Backups   int    `toml:"backups"`
}

func DefaultLoggingFileSettings() LoggingFileSettings {
	return LoggingFileSettings{Enabled: false, Level: "INFO", MaxSizeMB: 10, Backups: 3}
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
