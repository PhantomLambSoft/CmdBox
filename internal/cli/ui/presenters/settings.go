package presenters

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
)

var ErrUnknownSettingsField = errors.New("unknown settings field")

type settingsField struct {
	Section string
	Path    string
	Key     string
	Value   string
}

func field(section, path, value string) settingsField {
	key := path
	if i := strings.LastIndex(path, "."); i >= 0 {
		key = path[i+1:]
	}
	return settingsField{Section: section, Path: path, Key: key, Value: value}
}

func collectAllFields(s settings.Settings) []settingsField {
	var f []settingsField
	f = append(f, collectUIFields(s.UI)...)
	f = append(f, collectExecutionFields(s.ExecutionSettings)...)
	f = append(f, collectDefaultFieldsFields(s.DefaultFields)...)
	f = append(f, collectFieldAliasesFields(s.FieldAliases)...)
	f = append(f, collectLoggingFields(s.Logging)...)
	f = append(f, collectHistoryFields(s.History)...)
	return f
}

func collectUIFields(u settings.UISettings) []settingsField {
	fields := []settingsField{
		field("ui", "ui.use_color", strconv.FormatBool(u.UseColor)),
		field("ui", "ui.theme", u.Theme),
		field("ui", "ui.pager_mode", u.PagerMode),
		field("ui", "ui.pager_min_rows", strconv.Itoa(u.PagerMinRows)),
		field("ui", "ui.pager_page_step", strconv.Itoa(u.PagerPageStep)),
		field("ui", "ui.pager_line_step", strconv.Itoa(u.PagerLineStep)),
	}
	if len(u.ColorOverrides) > 0 {
		fields = append(fields, field("ui", "ui.color_overrides", fmt.Sprintf("%d key(s) overridden", len(u.ColorOverrides))))
	}
	return fields
}

func collectExecutionFields(e settings.ExecutionSettings) []settingsField {
	return []settingsField{
		field("execution_settings", "execution_settings.default_shell", e.DefaultShell),
		field("execution_settings", "execution_settings.capture_output", strconv.FormatBool(e.CaptureOutput)),
		field("execution_settings", "execution_settings.default_verbose", strconv.FormatBool(e.DefaultVerbose)),
	}
}

func collectDefaultFieldsFields(d settings.DefaultFields) []settingsField {
	j := func(ss []string) string { return strings.Join(ss, ", ") }
	return []settingsField{
		field("default_fields", "default_fields.command_output", j(d.CommandOutput)),
		field("default_fields", "default_fields.command_search", j(d.CommandSearch)),
		field("default_fields", "default_fields.command_list_limit", strconv.Itoa(d.CommandListLimit)),
		field("default_fields", "default_fields.command_default_order", d.CommandDefaultOrder),
		field("default_fields", "default_fields.variable_output", j(d.VariableOutput)),
		field("default_fields", "default_fields.variable_search", j(d.VariableSearch)),
		field("default_fields", "default_fields.variable_list_limit", strconv.Itoa(d.VariableListLimit)),
		field("default_fields", "default_fields.variable_default_order", d.VariableDefaultOrder),
		field("default_fields", "default_fields.tag_output", j(d.TagOutput)),
		field("default_fields", "default_fields.tag_search", j(d.TagSearch)),
		field("default_fields", "default_fields.tag_list_limit", strconv.Itoa(d.TagListLimit)),
		field("default_fields", "default_fields.tag_default_order", d.TagDefaultOrder),
		field("default_fields", "default_fields.profile_output", j(d.ProfileOutput)),
		field("default_fields", "default_fields.profile_search", j(d.ProfileSearch)),
		field("default_fields", "default_fields.profile_list_limit", strconv.Itoa(d.ProfileListLimit)),
		field("default_fields", "default_fields.profile_default_order", d.ProfileDefaultOrder),
	}
}

func collectFieldAliasesFields(fa settings.FieldAliases) []settingsField {
	names := make([]string, 0, len(fa.AliasMapping))
	for name := range fa.AliasMapping {
		names = append(names, name)
	}
	sort.Strings(names)

	fields := make([]settingsField, len(names))
	for i, name := range names {
		fields[i] = field("field_aliases", "field_aliases.alias_mapping."+name, strings.Join(fa.AliasMapping[name], ", "))
	}
	return fields
}

func collectLoggingFields(l settings.LoggingSettings) []settingsField {
	return []settingsField{
		field("logging", "logging.console_level", l.ConsoleLevel),
		field("logging", "logging.file.enabled", strconv.FormatBool(l.File.Enabled)),
		field("logging", "logging.file.level", l.File.Level),
		field("logging", "logging.file.max_size_mb", strconv.Itoa(l.File.MaxSizeMB)),
		field("logging", "logging.file.backups", strconv.Itoa(l.File.Backups)),
	}
}

func collectHistoryFields(h settings.HistorySettings) []settingsField {
	return []settingsField{
		field("history", "history.enabled", strconv.FormatBool(h.Enabled)),
		field("history", "history.limit_per_command", strconv.Itoa(h.LimitPerCommand)),
	}
}

func RenderSettingsShow(s settings.Settings, requested []string, palette ui.Palette) (string, error) {
	all := collectAllFields(s)
	if len(requested) == 0 {
		return renderSectionGroups(groupBySection(all), palette), nil
	}

	var matched []settingsField
	for _, name := range requested {
		if strings.Contains(name, ".") {
			if f, ok := findByPath(all, name); ok {
				matched = append(matched, f)
				continue
			}
			group := findByPrefix(all, name)
			if len(group) == 0 {
				return "", fmt.Errorf("%w: %q", ErrUnknownSettingsField, name)
			}
			matched = append(matched, group...)
			continue
		}

		candidates := findByKey(all, name)
		switch len(candidates) {
		case 0:
			return "", fmt.Errorf("%w: %q", ErrUnknownSettingsField, name)
		case 1:
			matched = append(matched, candidates[0])
		default:
			for _, c := range candidates {
				c.Key = c.Path
				matched = append(matched, c)
			}
		}
	}

	return renderSectionGroups(groupBySection(matched), palette), nil
}

func findByPath(fields []settingsField, path string) (settingsField, bool) {
	for _, f := range fields {
		if f.Path == path {
			return f, true
		}
	}
	return settingsField{}, false
}

func findByKey(fields []settingsField, key string) []settingsField {
	var out []settingsField
	for _, f := range fields {
		if f.Key == key {
			out = append(out, f)
		}
	}
	return out
}

func findByPrefix(fields []settingsField, prefix string) []settingsField {
	var out []settingsField
	p := prefix + "."
	for _, f := range fields {
		if strings.HasPrefix(f.Path, p) {
			out = append(out, f)
		}
	}
	return out
}

func groupBySection(fields []settingsField) map[string][]settingsField {
	out := make(map[string][]settingsField)
	for _, f := range fields {
		out[f.Section] = append(out[f.Section], f)
	}
	return out
}

func renderSectionGroups(bySection map[string][]settingsField, palette ui.Palette) string {
	var sectionBodies []string
	for _, section := range orderedSectionNames() {
		fields := bySection[section]
		if len(fields) == 0 {
			continue
		}
		rows := make([][2]string, len(fields))
		for i, f := range fields {
			rows[i] = [2]string{f.Key, f.Value}
		}
		sectionBodies = append(sectionBodies,
			ui.Section(section, ui.KVTable(rows, palette), "", palette.Info.GetForeground(), palette.PanelTitle),
		)
	}
	return ui.Section("Settings", ui.Stack(sectionBodies...), "", palette.Info.GetForeground(), palette.PanelTitle)
}

func orderedSectionNames() []string {
	return []string{"ui", "execution_settings", "default_fields", "field_aliases", "logging", "history"}
}

func RenderDataDir(dataDir string, palette ui.Palette) string {
	title := palette.Title.Render("Data Directory")
	body := palette.CodeInline.Render(dataDir)
	return ui.Section(title, body, "", palette.Border, palette.PanelTitle)
}
