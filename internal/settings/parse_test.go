package settings

import "testing"

func TestParseByteSize(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{name: "plain bytes with suffix", input: "100b", want: 100},
		{name: "plain number without suffix", input: "100", want: 100},
		{name: "kilobytes", input: "1kb", want: 1024},
		{name: "megabytes", input: "1mb", want: 1024 * 1024},
		{name: "gigabytes", input: "1gb", want: 1024 * 1024 * 1024},
		{name: "uppercase suffix", input: "1MB", want: 1024 * 1024},
		{name: "mixed case suffix", input: "1Mb", want: 1024 * 1024},
		{name: "fractional value", input: "1.5mb", want: int64(1.5 * 1024 * 1024)},
		{name: "internal spaces trimmed", input: "1 mb", want: 1024 * 1024},
		{name: "surrounding whitespace trimmed", input: "  1mb  ", want: 1024 * 1024},
		{name: "zero value", input: "0", want: 0},
		{name: "zero with suffix", input: "0kb", want: 0},
		{name: "negative plain number", input: "-5", want: -5},
		{name: "large kilobyte value", input: "500kb", want: 500 * 1024},
		{name: "empty string is invalid", input: "", wantErr: true},
		{name: "non-numeric string is invalid", input: "abc", wantErr: true},
		{name: "unknown suffix is invalid", input: "5xb", wantErr: true},
		{name: "malformed number before suffix", input: "1.5.5mb", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseByteSize(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseByteSize(%q) error = nil, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseByteSize(%q) error = %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseByteSize(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}

	t.Run("kb suffix takes precedence over b suffix", func(t *testing.T) {
		got, err := ParseByteSize("2kb")
		if err != nil {
			t.Fatalf("ParseByteSize() error = %v", err)
		}
		if got != 2048 {
			t.Fatalf("ParseByteSize(\"2kb\") = %d, want 2048 (kb match, not b match)", got)
		}
	})
}
