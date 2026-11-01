package commands

import (
	"errors"
	"testing"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
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
	var got handlers.ListHistoryArgs
	stubHandlerCall(t, &callRunListHistory, func(args handlers.ListHistoryArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(historyListCmd)
	assert.NoError(t, runHistoryList(cmd, []string{}))
	assert.Equal(t, handlers.ListHistoryArgs{ListParams: handlers.ListParams{Fields: []string{}}}, got)
}

func TestRunHistoryList_WithLimitAndAlias(t *testing.T) {
	var got handlers.ListHistoryArgs
	stubHandlerCall(t, &callRunListHistory, func(args handlers.ListHistoryArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(historyListCmd)
	require.NoError(t, cmd.Flags().Set("limit", "10"))
	require.NoError(t, cmd.Flags().Set("alias", "deploy"))
	assert.NoError(t, runHistoryList(cmd, []string{}))

	assert.Equal(t, 10, got.Limit)
	assert.True(t, got.HasLimit)
	assert.Equal(t, "deploy", got.Alias)
	assert.True(t, got.HasAlias)
}

func TestRunHistoryList_WithOrderAndField(t *testing.T) {
	var got handlers.ListHistoryArgs
	stubHandlerCall(t, &callRunListHistory, func(args handlers.ListHistoryArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(historyListCmd)
	require.NoError(t, cmd.Flags().Set("order", "ran_at"))
	require.NoError(t, cmd.Flags().Set("field", "alias,ran_at"))
	assert.NoError(t, runHistoryList(cmd, []string{}))

	assert.Equal(t, "ran_at", got.Order)
	assert.Equal(t, []string{"alias", "ran_at"}, got.Fields)
}

// --profile must reach handlers.ListHistoryArgs.Profile - before that wiring existed, this
// flag was silently ignored and the active profile's history was listed instead.
func TestRunHistoryList_ProfileIsForwarded(t *testing.T) {
	var got handlers.ListHistoryArgs
	stubHandlerCall(t, &callRunListHistory, func(args handlers.ListHistoryArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(historyListCmd)
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	assert.NoError(t, runHistoryList(cmd, []string{}))
	assert.Equal(t, "work", got.Profile)
}

func TestRunHistoryList_PropagatesHandlerError(t *testing.T) {
	stubHandlerCall(t, &callRunListHistory, func(args handlers.ListHistoryArgs) error {
		return errors.New("boom")
	})

	cmd := freshCmdWithFlagsOf(historyListCmd)
	err := runHistoryList(cmd, []string{})
	assert.EqualError(t, err, "boom")
}

func TestRunHistoryList_PageAndNoPageConflict(t *testing.T) {
	stubHandlerCall(t, &callRunListHistory, func(args handlers.ListHistoryArgs) error {
		t.Fatal("handler should not be called when --page and --no-page conflict")
		return nil
	})

	cmd := freshCmdWithFlagsOf(historyListCmd)
	require.NoError(t, cmd.Flags().Set("page", "true"))
	require.NoError(t, cmd.Flags().Set("no-page", "true"))
	assert.Error(t, runHistoryList(cmd, []string{}))
}

func TestRunHistoryShow_ForwardsIndexAliasAndProfile(t *testing.T) {
	var gotIndex string
	var gotAlias, gotProfile *string
	stubHandlerCall(t, &callRunShowHistory, func(indexOrID string, alias, profileName *string) error {
		gotIndex, gotAlias, gotProfile = indexOrID, alias, profileName
		return nil
	})

	cmd := freshCmdWithFlagsOf(historyShowCmd)
	require.NoError(t, cmd.Flags().Set("alias", "deploy"))
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	assert.NoError(t, runHistoryShow(cmd, []string{"3"}))

	assert.Equal(t, "3", gotIndex)
	require.NotNil(t, gotAlias)
	assert.Equal(t, "deploy", *gotAlias)
	require.NotNil(t, gotProfile)
	assert.Equal(t, "work", *gotProfile)
}

func TestRunHistoryShow_AliasAndProfileNilWhenUnset(t *testing.T) {
	var gotAlias, gotProfile *string
	stubHandlerCall(t, &callRunShowHistory, func(indexOrID string, alias, profileName *string) error {
		gotAlias, gotProfile = alias, profileName
		return nil
	})

	cmd := freshCmdWithFlagsOf(historyShowCmd)
	assert.NoError(t, runHistoryShow(cmd, []string{"1"}))
	assert.Nil(t, gotAlias)
	assert.Nil(t, gotProfile)
}

func TestRunHistoryShow_PropagatesHandlerError(t *testing.T) {
	stubHandlerCall(t, &callRunShowHistory, func(indexOrID string, alias, profileName *string) error {
		return errors.New("profile not found")
	})

	cmd := freshCmdWithFlagsOf(historyShowCmd)
	err := runHistoryShow(cmd, []string{"1"})
	assert.EqualError(t, err, "profile not found")
}

func TestRunHistoryRerun_ForwardsIndexAliasAndProfile(t *testing.T) {
	var gotIndex string
	var gotAlias, gotProfile *string
	stubHandlerCall(t, &callRunRerunHistory, func(indexOrID string, alias, profileName *string) error {
		gotIndex, gotAlias, gotProfile = indexOrID, alias, profileName
		return nil
	})

	cmd := freshCmdWithFlagsOf(historyRerunCmd)
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	assert.NoError(t, runHistoryRerun(cmd, []string{"2"}))

	assert.Equal(t, "2", gotIndex)
	assert.Nil(t, gotAlias)
	require.NotNil(t, gotProfile)
	assert.Equal(t, "work", *gotProfile)
}

func TestRunHistoryRerun_PropagatesHandlerError(t *testing.T) {
	stubHandlerCall(t, &callRunRerunHistory, func(indexOrID string, alias, profileName *string) error {
		return errors.New("profile not found")
	})

	cmd := freshCmdWithFlagsOf(historyRerunCmd)
	err := runHistoryRerun(cmd, []string{"1"})
	assert.EqualError(t, err, "profile not found")
}

func TestRunHistoryLast_DelegatesToHandler(t *testing.T) {
	called := false
	stubHandlerCall(t, &callRunLastHistory, func() error {
		called = true
		return nil
	})

	cmd := freshCmdWithFlagsOf(historyLastCmd)
	assert.NoError(t, runHistoryLast(cmd, []string{}))
	assert.True(t, called)
}

func TestRunHistoryLast_PropagatesHandlerError(t *testing.T) {
	stubHandlerCall(t, &callRunLastHistory, func() error {
		return errors.New("no history")
	})

	assert.EqualError(t, runHistoryLast(historyLastCmd, []string{}), "no history")
}

// TestRunHistoryClear_HappyPath (without --yes) is intentionally not covered here: that path
// drives a real interactive confirmation prompt (huh.NewConfirm().Run()), which has no mockable
// seam in this test binary and hangs indefinitely outside a real terminal.

func TestRunHistoryClear_WithYesAndAlias(t *testing.T) {
	var got handlers.ClearHistoryArgs
	stubHandlerCall(t, &callRunClearHistory, func(args handlers.ClearHistoryArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(historyClearCmd)
	require.NoError(t, cmd.Flags().Set("yes", "true"))
	require.NoError(t, cmd.Flags().Set("alias", "deploy"))
	assert.NoError(t, runHistoryClear(cmd, []string{}))

	assert.True(t, got.Yes)
	assert.Equal(t, "deploy", got.Alias)
	assert.True(t, got.HasAlias)
}

func TestRunHistoryClear_ProfileIsForwarded(t *testing.T) {
	var got handlers.ClearHistoryArgs
	stubHandlerCall(t, &callRunClearHistory, func(args handlers.ClearHistoryArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(historyClearCmd)
	require.NoError(t, cmd.Flags().Set("yes", "true"))
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	assert.NoError(t, runHistoryClear(cmd, []string{}))

	assert.Equal(t, "work", got.Profile)
	assert.True(t, got.HasProfile)
}

func TestRunHistoryClear_PropagatesHandlerError(t *testing.T) {
	stubHandlerCall(t, &callRunClearHistory, func(args handlers.ClearHistoryArgs) error {
		return errors.New("profile not found")
	})

	cmd := freshCmdWithFlagsOf(historyClearCmd)
	require.NoError(t, cmd.Flags().Set("yes", "true"))
	err := runHistoryClear(cmd, []string{})
	assert.EqualError(t, err, "profile not found")
}

// bangBangCmd and historyLastCmd are documented as equivalent shortcuts, so they must be
// wired to the same handler function.
func TestBangBangCmd_SharesHandlerWithHistoryLast(t *testing.T) {
	assert.Equal(t, functionName(historyLastCmd.RunE), functionName(bangBangCmd.RunE))
}
