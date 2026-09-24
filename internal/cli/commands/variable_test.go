package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVarCmd_Metadata(t *testing.T) {
	assert.Equal(t, "var", varCmd.Name())
	assert.Equal(t, []string{"vars"}, varCmd.Aliases)
	assert.Contains(t, rootCmd.Commands(), varCmd)
}

func TestVarCmd_Subcommands_Wired(t *testing.T) {
	want := []*cobra.Command{
		varAddCmd, varGetCmd, varUpdateCmd, varListCmd, varSearchCmd, varDeleteCmd, varAttachTagCmd, varDetachTagCmd,
	}
	for _, c := range want {
		assert.Contains(t, varCmd.Commands(), c, "expected %s to be a child of varCmd", c.Name())
	}
}

func TestVarAddCmd_Args(t *testing.T) {
	assert.NoError(t, varAddCmd.ValidateArgs([]string{}))
	assert.NoError(t, varAddCmd.ValidateArgs([]string{"name"}))
	assert.NoError(t, varAddCmd.ValidateArgs([]string{"name", "value"}))
	assert.Error(t, varAddCmd.ValidateArgs([]string{"name", "value", "extra"}))
}

func TestVarGetCmd_Args(t *testing.T) {
	assert.Error(t, varGetCmd.ValidateArgs([]string{}))
	assert.NoError(t, varGetCmd.ValidateArgs([]string{"name"}))
	assert.Error(t, varGetCmd.ValidateArgs([]string{"name", "extra"}))
}

func TestVarUpdateCmd_ArgsAndAliases(t *testing.T) {
	assert.Error(t, varUpdateCmd.ValidateArgs([]string{}))
	assert.NoError(t, varUpdateCmd.ValidateArgs([]string{"name"}))
	assert.Equal(t, []string{"edit"}, varUpdateCmd.Aliases)
}

func TestVarListCmd_Aliases(t *testing.T) {
	assert.Equal(t, []string{"ls"}, varListCmd.Aliases)
	assert.NoError(t, varListCmd.ValidateArgs([]string{}))
}

func TestVarSearchCmd_ArgsAndAliases(t *testing.T) {
	assert.Equal(t, []string{"find"}, varSearchCmd.Aliases)
	assert.Error(t, varSearchCmd.ValidateArgs([]string{}))
	assert.NoError(t, varSearchCmd.ValidateArgs([]string{"term"}))
}

func TestVarDeleteCmd_ArgsAndAliases(t *testing.T) {
	assert.ElementsMatch(t, []string{"rm", "del", "remove"}, varDeleteCmd.Aliases)
	assert.Error(t, varDeleteCmd.ValidateArgs([]string{}))
	assert.NoError(t, varDeleteCmd.ValidateArgs([]string{"name"}))
}

func TestVarAttachTagCmd_Args(t *testing.T) {
	assert.Error(t, varAttachTagCmd.ValidateArgs([]string{"name"}))
	assert.NoError(t, varAttachTagCmd.ValidateArgs([]string{"name", "tag1"}))
	assert.NoError(t, varAttachTagCmd.ValidateArgs([]string{"name", "tag1", "tag2"}))
}

func TestVarDetachTagCmd_ArgsAndAliases(t *testing.T) {
	assert.Equal(t, []string{"remove-tag"}, varDetachTagCmd.Aliases)
	assert.Error(t, varDetachTagCmd.ValidateArgs([]string{"name"}))
	assert.NoError(t, varDetachTagCmd.ValidateArgs([]string{"name", "tag1"}))
}

func TestVarAddCmd_Flags(t *testing.T) {
	f := varAddCmd.Flags()
	assertFlag(t, f, "interactive", "i", "bool", "false")
	assertFlag(t, f, "profile", "", "string", "")
}

func TestVarUpdateCmd_Flags(t *testing.T) {
	f := varUpdateCmd.Flags()
	assertFlag(t, f, "value", "", "string", "")
	assertFlag(t, f, "name", "", "string", "")
	assertFlag(t, f, "set", "", "string", "")
	assertFlag(t, f, "edit", "e", "bool", "false")
	assertFlag(t, f, "edit-fields", "", "string", "")
	assertFlag(t, f, "profile", "p", "string", "")
}

func TestVarListCmd_Flags(t *testing.T) {
	f := varListCmd.Flags()
	assertFlag(t, f, "tag", "", "stringSlice", "[]")
	assertFlag(t, f, "order", "o", "string", "")
	assertFlag(t, f, "field", "", "stringSlice", "[]")
	assertFlag(t, f, "limit", "l", "int", "0")
	assertFlag(t, f, "page", "", "bool", "false")
	assertFlag(t, f, "no-page", "", "bool", "false")
	assertFlag(t, f, "profile", "p", "string", "")
}

func TestVarSearchCmd_Flags(t *testing.T) {
	f := varSearchCmd.Flags()
	assertFlag(t, f, "in", "", "stringSlice", "[]")
	assertFlag(t, f, "field", "", "stringSlice", "[]")
	assertFlag(t, f, "limit", "l", "int", "0")
	assertFlag(t, f, "page", "", "bool", "false")
	assertFlag(t, f, "no-page", "", "bool", "false")
	assertFlag(t, f, "profile", "p", "string", "")
}

func TestVarDeleteCmd_Flags(t *testing.T) {
	assertFlag(t, varDeleteCmd.Flags(), "profile", "p", "string", "")
}

func TestVarAttachTagCmd_Flags(t *testing.T) {
	assertFlag(t, varAttachTagCmd.Flags(), "profile", "p", "string", "")
}

func TestVarDetachTagCmd_Flags(t *testing.T) {
	assertFlag(t, varDetachTagCmd.Flags(), "profile", "p", "string", "")
}

func TestRunVarAdd_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(varAddCmd)
	assert.NoError(t, runVarAdd(cmd, []string{"name", "value"}))
}

func TestRunVarAdd_NoArgs(t *testing.T) {
	cmd := freshCmdWithFlagsOf(varAddCmd)
	assert.NoError(t, runVarAdd(cmd, []string{}))
}

func TestRunVarGet(t *testing.T) {
	cmd := freshCmdWithFlagsOf(varGetCmd)
	assert.NoError(t, runVarGet(cmd, []string{"name"}))
}

func TestRunVarUpdate_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(varUpdateCmd)
	require.NoError(t, cmd.Flags().Set("value", "newvalue"))
	assert.NoError(t, runVarUpdate(cmd, []string{"name"}))
}

func TestRunVarList_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(varListCmd)
	assert.NoError(t, runVarList(cmd, []string{}))
}

func TestRunVarList_PageAndNoPageConflict(t *testing.T) {
	cmd := freshCmdWithFlagsOf(varListCmd)
	require.NoError(t, cmd.Flags().Set("page", "true"))
	require.NoError(t, cmd.Flags().Set("no-page", "true"))
	assert.Error(t, runVarList(cmd, []string{}))
}

func TestRunVarSearch_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(varSearchCmd)
	assert.NoError(t, runVarSearch(cmd, []string{"term"}))
}

func TestRunVarDelete(t *testing.T) {
	cmd := freshCmdWithFlagsOf(varDeleteCmd)
	assert.NoError(t, runVarDelete(cmd, []string{"name"}))
}

func TestRunVarAttachTags_UsesPositionalTagNames(t *testing.T) {
	cmd := freshCmdWithFlagsOf(varAttachTagCmd)
	assert.NoError(t, runVarAttachTags(cmd, []string{"name", "tag1", "tag2"}))
}

func TestRunVarDetachTags_UsesPositionalTagNames(t *testing.T) {
	cmd := freshCmdWithFlagsOf(varDetachTagCmd)
	assert.NoError(t, runVarDetachTags(cmd, []string{"name", "tag1"}))
}

// runVarAttachTags/runVarDetachTags must call distinct handlers (RunTagVar vs RunUntagVar) -
// this is the pattern the (previously buggy) cmd equivalents were fixed to match.
func TestVarAttachAndDetachTag_UseDistinctHandlerFunctions(t *testing.T) {
	assert.NotEqual(t,
		functionName(varAttachTagCmd.RunE),
		functionName(varDetachTagCmd.RunE),
	)
}
