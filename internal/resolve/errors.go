package resolve

import (
	"strings"
)

type CycleDetectionError struct {
	path []string
}

func (e *CycleDetectionError) Error() string {
	return "cycle detected: " + strings.Join(e.path, " -> ")
}
