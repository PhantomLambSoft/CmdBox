package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagCmd_Metadata(t *testing.T) {
	assert.Equal(t, "tag", tagCmd.Name())
	assert.Equal(t, []string{"tags"}, tagCmd.Aliases)
	assert.Contains(t, rootCmd.Commands(), tagCmd)
}

func TestTagCmd_Subcommands_Wired(t *testing.T) {
	want := []*cobra.Command{tagAddCmd, tagGetCmd, tagUpdateCmd, tagListCmd, tagSearchCmd, tagDeleteCmd}
	for _, c := range want {
		assert.Contains(t, tagCmd.Commands(), c, "expected %s to be a child of tagCmd", c.Name())
	}
}

func TestTagAddCmd_Args(t *testing.T) {
	assert.Error(t, tagAddCmd.ValidateArgs([]string{}))
	assert.NoError(t, tagAddCmd.ValidateArgs([]string{"name"}))
	assert.NoError(t, tagAddCmd.ValidateArgs([]string{"name", "desc"}))
	assert.Error(t, tagAddCmd.ValidateArgs([]string{"name", "desc", "extra"}))
}

func TestTagGetCmd_Args(t *testing.T) {
	assert.Error(t, tagGetCmd.ValidateArgs([]string{}))
	assert.NoError(t, tagGetCmd.ValidateArgs([]string{"name"}))
	assert.Error(t, tagGetCmd.ValidateArgs([]string{"name", "extra"}))
}

func TestTagUpdateCmd_ArgsAndAliases(t *testing.T) {
	assert.Error(t, tagUpdateCmd.ValidateArgs([]string{}))
	assert.NoError(t, tagUpdateCmd.ValidateArgs([]string{"name"}))
	assert.Equal(t, []string{"edit"}, tagUpdateCmd.Aliases)
}

func TestTagListCmd_Aliases(t *testing.T) {
	assert.Equal(t, []string{"ls"}, tagListCmd.Aliases)
	assert.NoError(t, tagListCmd.ValidateArgs([]string{}))
}

func TestTagSearchCmd_ArgsAndAliases(t *testing.T) {
	assert.Equal(t, []string{"find"}, tagSearchCmd.Aliases)
	assert.Error(t, tagSearchCmd.ValidateArgs([]string{}))
	assert.NoError(t, tagSearchCmd.ValidateArgs([]string{"term"}))
}

func TestTagDeleteCmd_ArgsAndAliases(t *testing.T) {
	assert.ElementsMatch(t, []string{"rm", "del", "remove"}, tagDeleteCmd.Aliases)
	assert.Error(t, tagDeleteCmd.ValidateArgs([]string{}))
	assert.NoError(t, tagDeleteCmd.ValidateArgs([]string{"name"}))
}

func TestTagAddCmd_Flags(t *testing.T) {
	assertFlag(t, tagAddCmd.Flags(), "interactive", "i", "bool", "false")
}

func TestTagUpdateCmd_Flags(t *testing.T) {
	f := tagUpdateCmd.Flags()
	assertFlag(t, f, "description", "", "string", "")
	assertFlag(t, f, "name", "", "string", "")
	assertFlag(t, f, "set", "", "string", "")
	assertFlag(t, f, "edit", "e", "bool", "false")
	assertFlag(t, f, "edit-fields", "", "string", "")
}

func TestTagListCmd_Flags(t *testing.T) {
	f := tagListCmd.Flags()
	assertFlag(t, f, "order", "", "string", "")
	assertFlag(t, f, "field", "", "stringSlice", "[]")
	assertFlag(t, f, "limit", "", "int", "0")
	assertFlag(t, f, "page", "", "bool", "false")
	assertFlag(t, f, "no-page", "", "bool", "false")
	assertMutuallyExclusive(t, tagListCmd, "page", "no-page")
}

func TestTagSearchCmd_Flags(t *testing.T) {
	f := tagSearchCmd.Flags()
	assertFlag(t, f, "in", "", "stringSlice", "[]")
	assertFlag(t, f, "field", "", "stringSlice", "[]")
	assertFlag(t, f, "limit", "", "int", "0")
	assertFlag(t, f, "page", "", "bool", "false")
	assertFlag(t, f, "no-page", "", "bool", "false")
	assertMutuallyExclusive(t, tagSearchCmd, "page", "no-page")
}

func TestRunTagAdd_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(tagAddCmd)
	assert.NoError(t, runTagAdd(cmd, []string{"name"}))
	assert.NoError(t, runTagAdd(cmd, []string{"name", "desc"}))
}

func TestRunTagGet(t *testing.T) {
	cmd := freshCmdWithFlagsOf(tagGetCmd)
	assert.NoError(t, runTagGet(cmd, []string{"name"}))
}

func TestRunTagUpdate_EditFieldsRequiresEditFlag(t *testing.T) {
	cmd := freshCmdWithFlagsOf(tagUpdateCmd)
	require.NoError(t, cmd.Flags().Set("edit-fields", "description"))
	err := runTagUpdate(cmd, []string{"name"})
	assert.EqualError(t, err, "--edit-fields can only be used with --edit")
}

func TestRunTagUpdate_EditFieldsWithEditFlagIsAllowed(t *testing.T) {
	cmd := freshCmdWithFlagsOf(tagUpdateCmd)
	require.NoError(t, cmd.Flags().Set("edit", "true"))
	require.NoError(t, cmd.Flags().Set("edit-fields", "description"))
	assert.NoError(t, runTagUpdate(cmd, []string{"name"}))
}

func TestRunTagUpdate_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(tagUpdateCmd)
	require.NoError(t, cmd.Flags().Set("description", "new desc"))
	assert.NoError(t, runTagUpdate(cmd, []string{"name"}))
}

func TestRunTagList_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(tagListCmd)
	assert.NoError(t, runTagList(cmd, []string{}))
}

func TestRunTagList_PageAndNoPageConflict(t *testing.T) {
	cmd := freshCmdWithFlagsOf(tagListCmd)
	require.NoError(t, cmd.Flags().Set("page", "true"))
	require.NoError(t, cmd.Flags().Set("no-page", "true"))
	assert.Error(t, runTagList(cmd, []string{}))
}

func TestRunTagSearch_HappyPath(t *testing.T) {
	cmd := freshCmdWithFlagsOf(tagSearchCmd)
	assert.NoError(t, runTagSearch(cmd, []string{"term"}))
}

func TestRunTagSearch_PageAndNoPageConflict(t *testing.T) {
	cmd := freshCmdWithFlagsOf(tagSearchCmd)
	require.NoError(t, cmd.Flags().Set("page", "true"))
	require.NoError(t, cmd.Flags().Set("no-page", "true"))
	assert.Error(t, runTagSearch(cmd, []string{"term"}))
}

func TestRunTagDelete(t *testing.T) {
	cmd := freshCmdWithFlagsOf(tagDeleteCmd)
	assert.NoError(t, runTagDelete(cmd, []string{"name"}))
}
