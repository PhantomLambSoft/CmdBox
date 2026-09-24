package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProfileCmd_Metadata(t *testing.T) {
	assert.Equal(t, "profile", profileCmd.Name())
	assert.Contains(t, rootCmd.Commands(), profileCmd)
}

func TestProfileCmd_Subcommands_Wired(t *testing.T) {
	want := []*cobra.Command{
		profileAddCmd, profileGetCmd, profileUpdateCmd, profileListCmd, profileDeleteCmd,
		profileStatusCmd, profileSwitchCmd, profileSwitchCmdCmd, profileSwitchVarCmd, profileSwitchSettingsCmd,
	}
	for _, c := range want {
		assert.Contains(t, profileCmd.Commands(), c, "expected %s to be a child of profileCmd", c.Name())
	}
}

func TestProfileCmd_Args(t *testing.T) {
	assert.NoError(t, profileCmd.ValidateArgs([]string{}))
	assert.NoError(t, profileCmd.ValidateArgs([]string{"myname"}))
	assert.Error(t, profileCmd.ValidateArgs([]string{"myname", "extra"}))
}

// profileCmd must run runProfileBareShorthand so `cb profile <name>` actually switches the
// profile - it was previously defined but never wired up as RunE.
func TestProfileCmd_HasBareShorthandRunE(t *testing.T) {
	require.NotNil(t, profileCmd.RunE)
	assert.Equal(t, functionName(runProfileBareShorthand), functionName(profileCmd.RunE))
}

func TestProfileAddCmd_Args(t *testing.T) {
	assert.Error(t, profileAddCmd.ValidateArgs([]string{}))
	assert.NoError(t, profileAddCmd.ValidateArgs([]string{"name"}))
	assert.NoError(t, profileAddCmd.ValidateArgs([]string{"name", "desc"}))
	assert.Error(t, profileAddCmd.ValidateArgs([]string{"name", "desc", "extra"}))
}

// profileGetCmd's Use is "get <name>" (one argument) and the handler only reads args[0],
// so exactly one argument must be required - it was previously misconfigured as ExactArgs(2).
func TestProfileGetCmd_Args(t *testing.T) {
	assert.Error(t, profileGetCmd.ValidateArgs([]string{}))
	assert.NoError(t, profileGetCmd.ValidateArgs([]string{"name"}))
	assert.Error(t, profileGetCmd.ValidateArgs([]string{"name", "extra"}))
}

func TestProfileUpdateCmd_ArgsAndAliases(t *testing.T) {
	assert.Error(t, profileUpdateCmd.ValidateArgs([]string{}))
	assert.NoError(t, profileUpdateCmd.ValidateArgs([]string{"name"}))
	assert.Equal(t, []string{"edit"}, profileUpdateCmd.Aliases)
}

func TestProfileListCmd_Aliases(t *testing.T) {
	assert.Equal(t, []string{"ls"}, profileListCmd.Aliases)
	assert.NoError(t, profileListCmd.ValidateArgs([]string{}))
}

func TestProfileDeleteCmd_ArgsAndAliases(t *testing.T) {
	assert.ElementsMatch(t, []string{"rm", "del", "remove"}, profileDeleteCmd.Aliases)
	assert.Error(t, profileDeleteCmd.ValidateArgs([]string{}))
	assert.NoError(t, profileDeleteCmd.ValidateArgs([]string{"name"}))
}

func TestProfileStatusCmd_ArgsAndAliases(t *testing.T) {
	assert.Equal(t, []string{"where"}, profileStatusCmd.Aliases)
	assert.NoError(t, profileStatusCmd.ValidateArgs([]string{}))
	assert.Error(t, profileStatusCmd.ValidateArgs([]string{"extra"}))
}

func TestProfileSwitchCmd_ArgsAndAliases(t *testing.T) {
	assert.Equal(t, []string{"set"}, profileSwitchCmd.Aliases)
	assert.Error(t, profileSwitchCmd.ValidateArgs([]string{}))
	assert.NoError(t, profileSwitchCmd.ValidateArgs([]string{"name"}))
}

func TestProfileSwitchCmdCmd_Args(t *testing.T) {
	assert.Error(t, profileSwitchCmdCmd.ValidateArgs([]string{}))
	assert.NoError(t, profileSwitchCmdCmd.ValidateArgs([]string{"name"}))
}

func TestProfileSwitchVarCmd_Args(t *testing.T) {
	assert.Error(t, profileSwitchVarCmd.ValidateArgs([]string{}))
	assert.NoError(t, profileSwitchVarCmd.ValidateArgs([]string{"name"}))
}

func TestProfileSwitchSettingsCmd_Args(t *testing.T) {
	assert.Error(t, profileSwitchSettingsCmd.ValidateArgs([]string{}))
	assert.NoError(t, profileSwitchSettingsCmd.ValidateArgs([]string{"name"}))
}

func TestProfileAddCmd_Flags(t *testing.T) {
	assertFlag(t, profileAddCmd.Flags(), "interactive", "i", "bool", "false")
}

func TestProfileUpdateCmd_Flags(t *testing.T) {
	f := profileUpdateCmd.Flags()
	assertFlag(t, f, "description", "", "string", "")
	assertFlag(t, f, "name", "", "string", "")
	assertFlag(t, f, "set", "", "string", "")
	assertFlag(t, f, "edit", "e", "bool", "false")
	assertFlag(t, f, "edit-fields", "", "string", "")
}

func TestProfileListCmd_Flags(t *testing.T) {
	f := profileListCmd.Flags()
	assertFlag(t, f, "order", "", "string", "")
	assertFlag(t, f, "field", "", "stringSlice", "[]")
	assertFlag(t, f, "limit", "", "int", "0")
	assertFlag(t, f, "page", "", "bool", "false")
	assertFlag(t, f, "no-page", "", "bool", "false")
	assertMutuallyExclusive(t, profileListCmd, "page", "no-page")
}

func TestProfileDeleteCmd_Flags(t *testing.T) {
	assertFlag(t, profileDeleteCmd.Flags(), "force", "", "bool", "false")
}

func TestProfileCmd_ScopeFlags(t *testing.T) {
	f := profileCmd.Flags()
	assertFlag(t, f, "cmd", "", "bool", "false")
	assertFlag(t, f, "var", "", "bool", "false")
	assertFlag(t, f, "settings", "", "bool", "false")
}

func TestProfileSwitchCmd_ScopeFlags(t *testing.T) {
	f := profileSwitchCmd.Flags()
	assertFlag(t, f, "cmd", "", "bool", "false")
	assertFlag(t, f, "var", "", "bool", "false")
	assertFlag(t, f, "settings", "", "bool", "false")
}

func TestRunProfileBareShorthand_NoArgsShowsHelp(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileCmd)
	assert.NoError(t, runProfileBareShorthand(cmd, []string{}))
}

func TestRunProfileBareShorthand_WithNameSwitches(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileCmd)
	assert.NoError(t, runProfileBareShorthand(cmd, []string{"myprofile"}))
}

func TestRunProfileAdd_NameIsSetFromFirstArg(t *testing.T) {
	// AddProfileArgs.Name previously wasn't populated at all; this exercises the code path
	// end-to-end (the stub handler can't be inspected, but this at least proves args[0] is
	// read without panicking and the call succeeds).
	cmd := freshCmdWithFlagsOf(profileAddCmd)
	assert.NoError(t, runProfileAdd(cmd, []string{"myprofile"}))
	assert.NoError(t, runProfileAdd(cmd, []string{"myprofile", "a description"}))
}

func TestRunProfileGet(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileGetCmd)
	assert.NoError(t, runProfileGet(cmd, []string{"myprofile"}))
}

func TestRunProfileUpdate_EditFieldsRequiresEditFlag(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileUpdateCmd)
	require.NoError(t, cmd.Flags().Set("edit-fields", "description"))
	err := runProfileUpdate(cmd, []string{"myprofile"})
	assert.EqualError(t, err, "--edit-fields can only be used with --edit")
}

func TestRunProfileUpdate_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileUpdateCmd)
	require.NoError(t, cmd.Flags().Set("description", "new desc"))
	assert.NoError(t, runProfileUpdate(cmd, []string{"myprofile"}))
}

func TestRunProfileList_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileListCmd)
	assert.NoError(t, runProfileList(cmd, []string{}))
}

func TestRunProfileList_PageAndNoPageConflict(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileListCmd)
	require.NoError(t, cmd.Flags().Set("page", "true"))
	require.NoError(t, cmd.Flags().Set("no-page", "true"))
	assert.Error(t, runProfileList(cmd, []string{}))
}

func TestRunProfileDelete(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileDeleteCmd)
	assert.NoError(t, runProfileDelete(cmd, []string{"myprofile"}))
}

func TestRunProfileStatus(t *testing.T) {
	assert.NoError(t, runProfileStatus(profileStatusCmd, []string{}))
}

func TestRunProfileSwitch_ReadsScopeFlags(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileSwitchCmd)
	require.NoError(t, cmd.Flags().Set("cmd", "true"))
	assert.NoError(t, runProfileSwitch(cmd, []string{"myprofile"}))
}

func TestRunProfileSwitchCmd(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileSwitchCmdCmd)
	assert.NoError(t, runProfileSwitchCmd(cmd, []string{"myprofile"}))
}

func TestRunProfileSwitchVar(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileSwitchVarCmd)
	assert.NoError(t, runProfileSwitchVar(cmd, []string{"myprofile"}))
}

func TestRunProfileSwitchSettings(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileSwitchSettingsCmd)
	assert.NoError(t, runProfileSwitchSettings(cmd, []string{"myprofile"}))
}

func TestSwitchArgsFromFlags(t *testing.T) {
	cmd := freshCmdWithFlagsOf(profileSwitchCmd)
	require.NoError(t, cmd.Flags().Set("var", "true"))

	got := switchArgsFromFlags(cmd, "myprofile")
	assert.Equal(t, "myprofile", got.Name)
	assert.False(t, got.Cmd)
	assert.True(t, got.Var)
	assert.False(t, got.Settings)
}
