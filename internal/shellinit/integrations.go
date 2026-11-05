package shellinit

import (
	"embed"
	"fmt"
	"strings"
)

//go:embed integrations
var integrationFS embed.FS

// LoadIntegration loads an embedded integration file by name, normalizes line endings, and ensures one trailing newline.
func LoadIntegration(name string) (string, error) {
	data, err := integrationFS.ReadFile("integrations/" + name)
	if err != nil {
		return "", fmt.Errorf("loading integration %s: %w", name, err)
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	return strings.TrimRight(text, "\t\n") + "\n", nil
}
