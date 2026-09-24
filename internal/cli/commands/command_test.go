package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdCmd_Metadata(t *testing.T) {
	assert.Equal(t, "cmd", cmdCmd.Name())
	assert.Equal(t, []string{"cmds"}, cmdCmd.Aliases)
	assert.Contains(t, rootCmd.Commands(), cmdCmd)
}

func TestCmdCmd_Subcommands_Wired(t *testing.T) {
	want := []*cobra.Command{
		cmdAddCmd, cmdGetCmd, cmdUpdateCmd, cmdListCmd, cmdSearchCmd, cmdDeleteCmd, cmdTagCmd, cmdUntagCmd,
	}
	for _, c := range want {
		assert.Contains(t, cmdCmd.Commands(), c, "expected %s to be a child of cmdCmd", c.Name())
	}
}

func TestCmdAddCmd_Args(t *testing.T) {
	assert.NoError(t, cmdAddCmd.ValidateArgs([]string{}))
	assert.NoError(t, cmdAddCmd.ValidateArgs([]string{"alias"}))
	assert.NoError(t, cmdAddCmd.ValidateArgs([]string{"alias", "template"}))
	assert.Error(t, cmdAddCmd.ValidateArgs([]string{"alias", "template", "extra"}))
}

func TestCmdGetCmd_Args(t *testing.T) {
	assert.Error(t, cmdGetCmd.ValidateArgs([]string{}))
	assert.NoError(t, cmdGetCmd.ValidateArgs([]string{"alias"}))
	assert.Error(t, cmdGetCmd.ValidateArgs([]string{"alias", "extra"}))
}

func TestCmdUpdateCmd_ArgsAndAliases(t *testing.T) {
	assert.Error(t, cmdUpdateCmd.ValidateArgs([]string{}))
	assert.NoError(t, cmdUpdateCmd.ValidateArgs([]string{"alias"}))
	assert.Equal(t, []string{"edit"}, cmdUpdateCmd.Aliases)
}

func TestCmdListCmd_Aliases(t *testing.T) {
	assert.Equal(t, []string{"ls"}, cmdListCmd.Aliases)
	assert.NoError(t, cmdListCmd.ValidateArgs([]string{}))
}

func TestCmdSearchCmd_ArgsAndAliases(t *testing.T) {
	assert.Equal(t, []string{"find"}, cmdSearchCmd.Aliases)
	assert.Error(t, cmdSearchCmd.ValidateArgs([]string{}))
	assert.NoError(t, cmdSearchCmd.ValidateArgs([]string{"term"}))
}

func TestCmdDeleteCmd_ArgsAndAliases(t *testing.T) {
	assert.ElementsMatch(t, []string{"rm", "del", "remove"}, cmdDeleteCmd.Aliases)
	assert.Error(t, cmdDeleteCmd.ValidateArgs([]string{}))
	assert.NoError(t, cmdDeleteCmd.ValidateArgs([]string{"alias"}))
}

func TestCmdTagCmd_Args(t *testing.T) {
	assert.Error(t, cmdTagCmd.ValidateArgs([]string{"alias"}))
	assert.NoError(t, cmdTagCmd.ValidateArgs([]string{"alias", "tag1"}))
	assert.NoError(t, cmdTagCmd.ValidateArgs([]string{"alias", "tag1", "tag2"}))
}

func TestCmdUntagCmd_Args(t *testing.T) {
	assert.Error(t, cmdUntagCmd.ValidateArgs([]string{"alias"}))
	assert.NoError(t, cmdUntagCmd.ValidateArgs([]string{"alias", "tag1"}))
}

func TestCmdAddCmd_Flags(t *testing.T) {
	f := cmdAddCmd.Flags()
	assertFlag(t, f, "description", "d", "string", "")
	assertFlag(t, f, "interactive", "i", "bool", "false")
	assertFlag(t, f, "cwd", "", "string", "")
	assertFlag(t, f, "shell", "", "string", "")
	assertFlag(t, f, "env", "", "stringArray", "[]")
	assertFlag(t, f, "timeout", "", "int", "0")
	assertFlag(t, f, "profile", "p", "string", "")
}

func TestCmdUpdateCmd_Flags(t *testing.T) {
	f := cmdUpdateCmd.Flags()
	assertFlag(t, f, "template", "t", "string", "")
	assertFlag(t, f, "description", "d", "string", "")
	assertFlag(t, f, "alias", "", "string", "")
	assertFlag(t, f, "cwd", "", "string", "")
	assertFlag(t, f, "clear-cwd", "", "bool", "false")
	assertFlag(t, f, "shell", "", "string", "")
	assertFlag(t, f, "clear-shell", "", "bool", "false")
	assertFlag(t, f, "env", "", "stringArray", "[]")
	assertFlag(t, f, "clear-env", "", "bool", "false")
	assertFlag(t, f, "timeout", "", "int", "0")
	assertFlag(t, f, "clear-timeout", "", "bool", "false")
	assertFlag(t, f, "set", "", "string", "")
	assertFlag(t, f, "edit", "e", "bool", "false")
	assertFlag(t, f, "edit-fields", "", "string", "")
	assertFlag(t, f, "profile", "p", "string", "")
}

func TestCmdUpdateCmd_MutuallyExclusiveGroups(t *testing.T) {
	cases := []struct{ a, b string }{
		{"cwd", "clear-cwd"},
		{"shell", "clear-shell"},
		{"env", "clear-env"},
		{"timeout", "clear-timeout"},
	}
	for _, c := range cases {
		t.Run(c.a+"_"+c.b, func(t *testing.T) {
			assertMutuallyExclusive(t, cmdUpdateCmd, c.a, c.b)
		})
	}
}

func TestCmdListCmd_Flags(t *testing.T) {
	f := cmdListCmd.Flags()
	assertFlag(t, f, "tag", "", "stringSlice", "[]")
	assertFlag(t, f, "order", "", "string", "")
	assertFlag(t, f, "field", "", "stringSlice", "[]")
	assertFlag(t, f, "limit", "", "int", "0")
	assertFlag(t, f, "page", "", "bool", "false")
	assertFlag(t, f, "no-page", "", "bool", "false")
	assertFlag(t, f, "profile", "p", "string", "")
	assertMutuallyExclusive(t, cmdListCmd, "page", "no-page")
}

func TestCmdSearchCmd_Flags(t *testing.T) {
	f := cmdSearchCmd.Flags()
	assertFlag(t, f, "in", "", "stringSlice", "[]")
	assertFlag(t, f, "field", "", "stringSlice", "[]")
	assertFlag(t, f, "limit", "", "int", "0")
	assertFlag(t, f, "page", "", "bool", "false")
	assertFlag(t, f, "no-page", "", "bool", "false")
	assertFlag(t, f, "profile", "p", "string", "")
	assertMutuallyExclusive(t, cmdSearchCmd, "page", "no-page")
}

func TestCmdDeleteCmd_Flags(t *testing.T) {
	assertFlag(t, cmdDeleteCmd.Flags(), "profile", "p", "string", "")
}

func TestCmdTagCmd_Flags(t *testing.T) {
	assertFlag(t, cmdTagCmd.Flags(), "profile", "p", "string", "")
}

func TestCmdUntagCmd_Flags(t *testing.T) {
	assertFlag(t, cmdUntagCmd.Flags(), "profile", "p", "string", "")
}

func TestRunCmdAdd_BuildsArgsFromPositionalsAndFlags(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdAddCmd)
	require.NoError(t, cmd.Flags().Set("timeout", "30"))
	err := runCmdAdd(cmd, []string{"myalias", "echo hi"})
	assert.NoError(t, err)
}

func TestRunCmdAdd_NoArgs(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdAddCmd)
	err := runCmdAdd(cmd, []string{})
	assert.NoError(t, err)
}

func TestRunCmdGet(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdGetCmd)
	assert.NoError(t, runCmdGet(cmd, []string{"myalias"}))
}

func TestRunCmdUpdate_EditFieldsRequiresEditFlag(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdUpdateCmd)
	require.NoError(t, cmd.Flags().Set("edit-fields", "template"))
	err := runCmdUpdate(cmd, []string{"myalias"})
	assert.EqualError(t, err, "--edit-fields can only be used with --edit")
}

func TestRunCmdUpdate_EditFieldsWithEditFlagIsAllowed(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdUpdateCmd)
	require.NoError(t, cmd.Flags().Set("edit", "true"))
	require.NoError(t, cmd.Flags().Set("edit-fields", "template"))
	err := runCmdUpdate(cmd, []string{"myalias"})
	assert.NoError(t, err)
}

func TestRunCmdUpdate_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdUpdateCmd)
	require.NoError(t, cmd.Flags().Set("template", "echo hi"))
	require.NoError(t, cmd.Flags().Set("description", "desc"))
	require.NoError(t, cmd.Flags().Set("cwd", "/tmp"))
	require.NoError(t, cmd.Flags().Set("timeout", "10"))
	err := runCmdUpdate(cmd, []string{"myalias"})
	assert.NoError(t, err)
}

func TestRunCmdList_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdListCmd)
	err := runCmdList(cmd, []string{})
	assert.NoError(t, err)
}

func TestRunCmdList_PageAndNoPageConflict(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdListCmd)
	require.NoError(t, cmd.Flags().Set("page", "true"))
	require.NoError(t, cmd.Flags().Set("no-page", "true"))
	err := runCmdList(cmd, []string{})
	assert.Error(t, err)
}

func TestRunCmdSearch_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdSearchCmd)
	err := runCmdSearch(cmd, []string{"term"})
	assert.NoError(t, err)
}

func TestRunCmdDelete(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdDeleteCmd)
	assert.NoError(t, runCmdDelete(cmd, []string{"myalias"}))
}

func TestRunCmdAttachTags(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdTagCmd)
	assert.NoError(t, runCmdAttachTags(cmd, []string{"myalias", "tag1", "tag2"}))
}

func TestRunCmdDetachTags(t *testing.T) {
	cmd := freshCmdWithFlagsOf(cmdUntagCmd)
	assert.NoError(t, runCmdDetachTags(cmd, []string{"myalias", "tag1"}))
}

// runCmdDetachTags must call the untag handler, not the attach handler - this was a
// copy/paste bug. Since the handlers are currently no-op stubs we can't observe which one
// ran through behavior, so this pins the wiring itself: the two subcommands must not share
// the same RunE function.
func TestCmdTagAndUntag_UseDistinctHandlerFunctions(t *testing.T) {
	assert.NotEqual(t,
		functionName(cmdTagCmd.RunE),
		functionName(cmdUntagCmd.RunE),
		"cmdTagCmd and cmdUntagCmd must not be wired to the same RunE function",
	)
}
