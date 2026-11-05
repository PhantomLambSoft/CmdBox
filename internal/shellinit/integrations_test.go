package shellinit

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadIntegration_KnownFiles(t *testing.T) {
	for _, name := range []string{"bash.sh", "zsh.sh", "fish.fish", "powershell.ps1", "cmd.bat"} {
		t.Run(name, func(t *testing.T) {
			text, err := LoadIntegration(name)
			require.NoError(t, err)
			require.NotEmpty(t, text)
			require.False(t, strings.Contains(text, "\r\n"), "should not contain CRLF after normalization")
			require.True(t, strings.HasSuffix(text, "\n"), "should end with exactly one newline")
			require.False(t, strings.HasSuffix(text, "\n\n"), "should not have trailing blank lines")
		})
	}
}

func TestLoadIntegration_MatchesSpecsTable(t *testing.T) {
	for _, name := range ShellNames() {
		spec, err := Lookup(name)
		require.NoError(t, err)
		if spec.Integration == "" {
			continue
		}
		_, err = LoadIntegration(spec.Integration)
		require.NoError(t, err, "integration %s referenced by spec %s should load", spec.Integration, name)
	}
}

func TestLoadIntegration_UnknownFile(t *testing.T) {
	_, err := LoadIntegration("does-not-exist.sh")
	require.Error(t, err)
	require.Contains(t, err.Error(), "does-not-exist.sh")
}
