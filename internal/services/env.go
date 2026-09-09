package services

import (
	"encoding/json"
	"fmt"
)

// parseEnv parses a JSON-encoded string into a map of environment variables and returns it or an error if parsing fails.
func parseEnv(source string) (map[string]string, error) {
	if source == "" {
		return nil, nil
	}
	var env map[string]string
	if err := json.Unmarshal([]byte(source), &env); err != nil {
		return nil, fmt.Errorf("parsing env: %w", err)
	}
	return env, nil
}
