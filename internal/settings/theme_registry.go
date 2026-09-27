package settings

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

//go:embed themes/*.toml
var embeddedThemes embed.FS

type Theme struct {
	Name   string   `toml:"name"`
	Colors UIColors `toml:"colors"`
}

func LoadTheme(name, userThemeDir string) (Theme, error) {
	if userThemeDir != "" {
		if data, err := os.ReadFile(filepath.Join(userThemeDir, name+".toml")); err == nil {
			return decodeTheme(data)
		}
	}
	data, err := embeddedThemes.ReadFile("themes/" + name + ".toml")
	if err != nil {
		return Theme{}, fmt.Errorf("theme %q not found", name)
	}
	return decodeTheme(data)
}

func decodeTheme(data []byte) (Theme, error) {
	var t Theme
	err := toml.NewDecoder(strings.NewReader(string(data))).EnableUnmarshalerInterface().Decode(&t)
	return t, err
}

func ListThemes(userThemeDir string) []string {
	seen := map[string]bool{}
	var names []string

	entries, _ := fs.Glob(embeddedThemes, "themes/*.toml")
	for _, e := range entries {
		n := strings.TrimSuffix(filepath.Base(e), ".toml")
		seen[n] = true
		names = append(names, n)
	}
	if userThemeDir != "" {
		_ = filepath.WalkDir(userThemeDir, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || filepath.Ext(p) != ".toml" {
				return nil
			}
			if n := strings.TrimSuffix(filepath.Base(p), ".toml"); !seen[n] {
				names = append(names, n)
			}
			return nil
		})
	}
	return names
}

// MergeColorOverrides updates a base UIColors structure with overrides and returns the updated result and unknown keys.
func MergeColorOverrides(base UIColors, overrides map[string]ColorStyle) (UIColors, []string) {
	out := base
	var unknown []string
	for key, val := range overrides {
		switch key {
		case "title":
			out.Title = val
		case "subtitle":
			out.Subtitle = val
		case "muted":
			out.Muted = val
		case "border":
			out.Border = val
		case "panel_title":
			out.PanelTitle = val
		case "table_header":
			out.TableHeader = val
		case "caption":
			out.Caption = val
		case "kv_key":
			out.KVKey = val
		case "kv_value":
			out.KVValue = val
		case "success":
			out.Success = val
		case "info":
			out.Info = val
		case "warning":
			out.Warning = val
		case "error":
			out.Error = val
		case "debug":
			out.Debug = val
		case "code":
			out.Code = val
		case "code_inline":
			out.CodeInline = val
		case "code_block":
			out.CodeBlock = val
		case "entity_name":
			out.EntityName = val
		case "entity_id":
			out.EntityID = val
		case "entity_count":
			out.EntityCount = val
		case "entity_time":
			out.EntityTime = val
		case "tag_pill":
			out.TagPill = val
		case "tag_pill_muted":
			out.TagPillMuted = val
		case "run_command":
			out.RunCommand = val
		case "run_stdout":
			out.RunStdout = val
		case "run_stderr":
			out.RunStderr = val
		case "trace_kind":
			out.TraceKind = val
		case "trace_key":
			out.TraceKey = val
		case "trace_value":
			out.TraceValue = val
		default:
			unknown = append(unknown, key)
		}
	}
	return out, unknown
}
