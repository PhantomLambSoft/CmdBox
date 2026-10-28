package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHistoryCmd_Metadata(t *testing.T) {
	assert.Equal(t, "history", historyCmd.Name())
	assert.Equal(t, []string{"hist"}, historyCmd.Aliases)
	assert.Contains(t, rootCmd.Commands(), historyCmd)
}

func TestHistoryCmd_Subcommands_Wired(t *testing.T) {
	want := []*cobra.Command{historyListCmd, historyShowCmd, historyRerunCmd, historyLastCmd, historyClearCmd}
	for _, c := range want {
		assert.Contains(t, historyCmd.Commands(), c, "expected %s to be a child of historyCmd", c.Name())
	}
}

func TestBangBangCmd_IsRootLevelAndHidden(t *testing.T) {
	assert.Equal(t, "!!", bangBangCmd.Name())
	assert.True(t, bangBangCmd.Hidden)
	assert.Contains(t, rootCmd.Commands(), bangBangCmd)
	assert.NotContains(t, historyCmd.Commands(), bangBangCmd)
}

func TestHistoryListCmd_Args(t *testing.T) {
	assert.NoError(t, historyListCmd.ValidateArgs([]string{}))
	assert.Error(t, historyListCmd.ValidateArgs([]string{"extra"}))
}

func TestHistoryShowCmd_Args(t *testing.T) {
	assert.Error(t, historyShowCmd.ValidateArgs([]string{}))
	assert.NoError(t, historyShowCmd.ValidateArgs([]string{"1"}))
	assert.Error(t, historyShowCmd.ValidateArgs([]string{"1", "2"}))
}

func TestHistoryRerunCmd_Args(t *testing.T) {
	assert.Error(t, historyRerunCmd.ValidateArgs([]string{}))
	assert.NoError(t, historyRerunCmd.ValidateArgs([]string{"1"}))
}

func TestHistoryLastCmd_Args(t *testing.T) {
	assert.NoError(t, historyLastCmd.ValidateArgs([]string{}))
	assert.Error(t, historyLastCmd.ValidateArgs([]string{"extra"}))
}

func TestHistoryClearCmd_Args(t *testing.T) {
	assert.NoError(t, historyClearCmd.ValidateArgs([]string{}))
	assert.Error(t, historyClearCmd.ValidateArgs([]string{"extra"}))
}

func TestBangBangCmd_Args(t *testing.T) {
	assert.NoError(t, bangBangCmd.ValidateArgs([]string{}))
	assert.Error(t, bangBangCmd.ValidateArgs([]string{"extra"}))
}

func TestHistoryListCmd_Flags(t *testing.T) {
	f := historyListCmd.Flags()
	assertFlag(t, f, "alias", "", "string", "")
	assertFlag(t, f, "order", "", "string", "")
	assertFlag(t, f, "field", "", "stringSlice", "[]")
	assertFlag(t, f, "limit", "", "int", "0")
	assertFlag(t, f, "page", "", "bool", "false")
	assertFlag(t, f, "no-page", "", "bool", "false")
	assertFlag(t, f, "profile", "p", "string", "")
	assertMutuallyExclusive(t, historyListCmd, "page", "no-page")
}

func TestHistoryShowCmd_Flags(t *testing.T) {
	f := historyShowCmd.Flags()
	assertFlag(t, f, "alias", "", "string", "")
	assertFlag(t, f, "profile", "p", "string", "")
}

func TestHistoryRerunCmd_Flags(t *testing.T) {
	f := historyRerunCmd.Flags()
	assertFlag(t, f, "alias", "", "string", "")
	assertFlag(t, f, "profile", "p", "string", "")
}

func TestHistoryClearCmd_Flags(t *testing.T) {
	f := historyClearCmd.Flags()
	assertFlag(t, f, "yes", "y", "bool", "false")
	assertFlag(t, f, "alias", "", "string", "")
	assertFlag(t, f, "profile", "p", "string", "")
}

func TestRunHistoryList_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyListCmd)
	assert.NoError(t, runHistoryList(cmd, []string{}))
}

func TestRunHistoryList_WithLimitAndAlias(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyListCmd)
	require.NoError(t, cmd.Flags().Set("limit", "10"))
	require.NoError(t, cmd.Flags().Set("alias", "deploy"))
	assert.NoError(t, runHistoryList(cmd, []string{}))
}

func TestRunHistoryList_WithOrderAndField(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyListCmd)
	require.NoError(t, cmd.Flags().Set("order", "ran_at"))
	require.NoError(t, cmd.Flags().Set("field", "alias,ran_at"))
	assert.NoError(t, runHistoryList(cmd, []string{}))
}

// A nonexistent profile name must surface a "profile not found" error. Before --profile was
// wired into handlers.ListHistoryArgs, this flag was silently ignored and the active profile's
// history was listed instead, so this also guards against that regression.
func TestRunHistoryList_WithUnknownProfile_PropagatesError(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyListCmd)
	require.NoError(t, cmd.Flags().Set("profile", "cmdbox-test-nonexistent-profile"))
	err := runHistoryList(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "profile")
}

func TestRunHistoryList_PageAndNoPageConflict(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyListCmd)
	require.NoError(t, cmd.Flags().Set("page", "true"))
	require.NoError(t, cmd.Flags().Set("no-page", "true"))
	assert.Error(t, runHistoryList(cmd, []string{}))
}

// seedHistoryEntry adds a command under the given (test-unique) alias and runs it once, producing
// exactly one new, real command_histories row - the most recent one (index "1") - without relying
// on any other test having left data behind. Each test that needs an existing history entry seeds
// its own alias so tests stay independent of execution order.
func seedHistoryEntry(t *testing.T, alias string) {
	t.Helper()
	addCmd := freshCmdWithFlagsOf(cmdAddCmd)
	require.NoError(t, runCmdAdd(addCmd, []string{alias, "echo " + alias}))
	require.NoError(t, runRun(runCmd, []string{alias}))
}

func TestRunHistoryShow(t *testing.T) {
	seedHistoryEntry(t, "history-test-show")
	cmd := freshCmdWithFlagsOf(historyShowCmd)
	assert.NoError(t, runHistoryShow(cmd, []string{"1"}))
}

// A nonexistent profile name must surface a "profile not found" error, proving --profile reaches
// the handler rather than being silently dropped.
func TestRunHistoryShow_WithUnknownProfile_PropagatesError(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyShowCmd)
	require.NoError(t, cmd.Flags().Set("alias", "deploy"))
	require.NoError(t, cmd.Flags().Set("profile", "cmdbox-test-nonexistent-profile"))
	err := runHistoryShow(cmd, []string{"1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "profile")
}

func TestRunHistoryRerun(t *testing.T) {
	seedHistoryEntry(t, "history-test-rerun")
	cmd := freshCmdWithFlagsOf(historyRerunCmd)
	assert.NoError(t, runHistoryRerun(cmd, []string{"1"}))
}

// A nonexistent profile name must surface a "profile not found" error, proving --profile reaches
// the handler rather than being silently dropped.
func TestRunHistoryRerun_WithUnknownProfile_PropagatesError(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyRerunCmd)
	require.NoError(t, cmd.Flags().Set("profile", "cmdbox-test-nonexistent-profile"))
	err := runHistoryRerun(cmd, []string{"1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "profile")
}

func TestRunHistoryLast(t *testing.T) {
	seedHistoryEntry(t, "history-test-last")
	cmd := freshCmdWithFlagsOf(historyLastCmd)
	assert.NoError(t, runHistoryLast(cmd, []string{}))
}

// TestRunHistoryClear_HappyPath (without --yes) is intentionally not covered here: that path
// drives a real interactive confirmation prompt (huh.NewConfirm().Run()), which has no mockable
// seam in this test binary and hangs indefinitely outside a real terminal.

// An unscoped TestRunHistoryClear_WithYes (--yes, no --alias) is deliberately not included here:
// this test package has no database isolation and runs against the real local CmdBox history, so
// an unfiltered clear would delete the developer's actual command history. Clear is only
// exercised scoped to an alias below.

func TestRunHistoryClear_WithYesAndAlias(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyClearCmd)
	require.NoError(t, cmd.Flags().Set("yes", "true"))
	require.NoError(t, cmd.Flags().Set("alias", "deploy"))
	assert.NoError(t, runHistoryClear(cmd, []string{}))
}

// A nonexistent profile name must surface a "profile not found" error, proving --profile reaches
// the handler rather than being silently dropped.
func TestRunHistoryClear_WithUnknownProfile_PropagatesError(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyClearCmd)
	require.NoError(t, cmd.Flags().Set("yes", "true"))
	require.NoError(t, cmd.Flags().Set("profile", "cmdbox-test-nonexistent-profile"))
	err := runHistoryClear(cmd, []string{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "profile")
}

// bangBangCmd and historyLastCmd are documented as equivalent shortcuts, so they must be
// wired to the same handler function.
func TestBangBangCmd_SharesHandlerWithHistoryLast(t *testing.T) {
	assert.Equal(t, functionName(historyLastCmd.RunE), functionName(bangBangCmd.RunE))
}
