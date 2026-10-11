package presenters

import (
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/PhantomLambSoft/CmdBox/internal/models"
)

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func formatTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func formatEnv(env *string) string {
	if env == nil || *env == "" {
		return ""
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(*env), &m); err != nil {
		return *env
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]string, len(keys))
	for i, k := range keys {
		lines[i] = k + "=" + m[k]
	}
	return strings.Join(lines, "\n")
}

func tagNames(tags []models.Tag) []string {
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	sort.Strings(names)
	return names
}
