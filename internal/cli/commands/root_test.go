package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Execute() itself isn't covered here: on error it calls os.Exit(1), which would kill the
// test binary. Its two components - resolveArgs (alias_fallback_test.go) and rootCmd.Execute
// (cobra, and each subcommand's own tests) - are covered separately.
func TestRootCmd_Metadata(t *testing.T) {
	assert.Equal(t, "cb", rootCmd.Name())
	assert.NotEmpty(t, rootCmd.Short)
}

func TestRootCmd_TopLevelCommandsRegistered(t *testing.T) {
	names := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"cmd", "run", "preview", "var", "tag", "history", "settings", "profile", "!!"} {
		assert.True(t, names[want], "expected rootCmd to have subcommand %q", want)
	}
}
