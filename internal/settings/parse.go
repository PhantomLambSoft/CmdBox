package settings

import (
	"fmt"
	"strconv"
	"strings"
)

type byteUnit struct {
	suffix     string
	multiplier int64
}

// byteUnits defines a list of byte units with their suffixes and corresponding multipliers for size conversion.
// Order matters here, so this is a slice of byteUnits instead of a map.
var byteUnits = []byteUnit{
	{"kb", 1024},
	{"mb", 1024 * 1024},
	{"gb", 1024 * 1024 * 1024},
	{"b", 1},
}

// ParseByteSize parses a human-readable byte size string (e.g., "1mb", "500kb") and returns its value in bytes as int64.
func ParseByteSize(value string) (int64, error) {
	cleaned := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(value), " ", ""))

	for _, u := range byteUnits {
		if strings.HasSuffix(cleaned, u.suffix) {
			numberStr := cleaned[:len(cleaned)-len(u.suffix)]
			number, err := strconv.ParseFloat(numberStr, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid byte size: %s", value)
			}
			return int64(number * float64(u.multiplier)), nil
		}
	}

	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid byte size: %s", value)
	}
	return n, nil
}
