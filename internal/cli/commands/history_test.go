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
	assertFlag(t, f, "limit", "", "int", "0")
	assertFlag(t, f, "alias", "", "string", "")
}

func TestHistoryClearCmd_Flags(t *testing.T) {
	f := historyClearCmd.Flags()
	assertFlag(t, f, "yes", "y", "bool", "false")
	assertFlag(t, f, "alias", "", "string", "")
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

func TestRunHistoryShow(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyShowCmd)
	assert.NoError(t, runHistoryShow(cmd, []string{"1"}))
}

func TestRunHistoryRerun(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyRerunCmd)
	assert.NoError(t, runHistoryRerun(cmd, []string{"1"}))
}

func TestRunHistoryLast(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyLastCmd)
	assert.NoError(t, runHistoryLast(cmd, []string{}))
}

func TestRunHistoryClear_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyClearCmd)
	assert.NoError(t, runHistoryClear(cmd, []string{}))
}

func TestRunHistoryClear_WithYesAndAlias(t *testing.T) {
	cmd := freshCmdWithFlagsOf(historyClearCmd)
	require.NoError(t, cmd.Flags().Set("yes", "true"))
	require.NoError(t, cmd.Flags().Set("alias", "deploy"))
	assert.NoError(t, runHistoryClear(cmd, []string{}))
}

// bangBangCmd and historyLastCmd are documented as equivalent shortcuts, so they must be
// wired to the same handler function.
func TestBangBangCmd_SharesHandlerWithHistoryLast(t *testing.T) {
	assert.Equal(t, functionName(historyLastCmd.RunE), functionName(bangBangCmd.RunE))
}
