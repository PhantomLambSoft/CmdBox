package commands

import (
	"errors"
	"testing"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
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

func TestRunTagAdd_BuildsArgsFromPositionalsAndFlags(t *testing.T) {
	var got handlers.AddTagArgs
	stubHandlerCall(t, &callRunAddTag, func(args handlers.AddTagArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(tagAddCmd)
	assert.NoError(t, runTagAdd(cmd, []string{"name"}))
	assert.Equal(t, "name", got.Name)
	assert.False(t, got.HasDescription)

	assert.NoError(t, runTagAdd(cmd, []string{"name", "desc"}))
	assert.Equal(t, "desc", got.Description)
	assert.True(t, got.HasDescription)
}

func TestRunTagAdd_PropagatesHandlerError(t *testing.T) {
	stubHandlerCall(t, &callRunAddTag, func(args handlers.AddTagArgs) error {
		return errors.New("boom")
	})

	cmd := freshCmdWithFlagsOf(tagAddCmd)
	assert.EqualError(t, runTagAdd(cmd, []string{"name"}), "boom")
}

func TestRunTagGet_ForwardsName(t *testing.T) {
	var gotName string
	stubHandlerCall(t, &callRunGetTag, func(name string) error {
		gotName = name
		return nil
	})

	cmd := freshCmdWithFlagsOf(tagGetCmd)
	assert.NoError(t, runTagGet(cmd, []string{"name"}))
	assert.Equal(t, "name", gotName)
}

func TestRunTagUpdate_EditFieldsRequiresEditFlag(t *testing.T) {
	stubHandlerCall(t, &callRunUpdateTag, func(args handlers.UpdateTagArgs) error {
		t.Fatal("handler should not be called when validation fails")
		return nil
	})

	cmd := freshCmdWithFlagsOf(tagUpdateCmd)
	require.NoError(t, cmd.Flags().Set("edit-fields", "description"))
	err := runTagUpdate(cmd, []string{"name"})
	assert.EqualError(t, err, "--edit-fields can only be used with --edit")
}

func TestRunTagUpdate_EditFieldsWithEditFlagIsAllowed(t *testing.T) {
	var got handlers.UpdateTagArgs
	stubHandlerCall(t, &callRunUpdateTag, func(args handlers.UpdateTagArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(tagUpdateCmd)
	require.NoError(t, cmd.Flags().Set("edit", "true"))
	require.NoError(t, cmd.Flags().Set("edit-fields", "description"))
	assert.NoError(t, runTagUpdate(cmd, []string{"name"}))
	assert.True(t, got.EditMode)
	assert.Equal(t, "description", got.EditFields)
}

func TestRunTagUpdate_HappyPath(t *testing.T) {
	var got handlers.UpdateTagArgs
	stubHandlerCall(t, &callRunUpdateTag, func(args handlers.UpdateTagArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(tagUpdateCmd)
	require.NoError(t, cmd.Flags().Set("description", "new desc"))
	assert.NoError(t, runTagUpdate(cmd, []string{"name"}))
	assert.Equal(t, "name", got.CurrentName)
	assert.Equal(t, "new desc", got.Description)
	assert.True(t, got.HasDescription)
}

func TestRunTagList_HappyPath(t *testing.T) {
	called := false
	stubHandlerCall(t, &callRunListTag, func(args handlers.ListTagArgs) error {
		called = true
		return nil
	})

	cmd := freshCmdWithFlagsOf(tagListCmd)
	assert.NoError(t, runTagList(cmd, []string{}))
	assert.True(t, called)
}

func TestRunTagList_PageAndNoPageConflict(t *testing.T) {
	stubHandlerCall(t, &callRunListTag, func(args handlers.ListTagArgs) error {
		t.Fatal("handler should not be called when --page and --no-page conflict")
		return nil
	})

	cmd := freshCmdWithFlagsOf(tagListCmd)
	require.NoError(t, cmd.Flags().Set("page", "true"))
	require.NoError(t, cmd.Flags().Set("no-page", "true"))
	assert.Error(t, runTagList(cmd, []string{}))
}

func TestRunTagSearch_HappyPath(t *testing.T) {
	var got handlers.SearchTagArgs
	stubHandlerCall(t, &callRunSearchTag, func(args handlers.SearchTagArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(tagSearchCmd)
	assert.NoError(t, runTagSearch(cmd, []string{"term"}))
	assert.Equal(t, "term", got.Term)
}

func TestRunTagSearch_PageAndNoPageConflict(t *testing.T) {
	stubHandlerCall(t, &callRunSearchTag, func(args handlers.SearchTagArgs) error {
		t.Fatal("handler should not be called when --page and --no-page conflict")
		return nil
	})

	cmd := freshCmdWithFlagsOf(tagSearchCmd)
	require.NoError(t, cmd.Flags().Set("page", "true"))
	require.NoError(t, cmd.Flags().Set("no-page", "true"))
	assert.Error(t, runTagSearch(cmd, []string{"term"}))
}

func TestRunTagDelete_ForwardsName(t *testing.T) {
	var got handlers.DeleteTagArgs
	stubHandlerCall(t, &callRunDeleteTag, func(args handlers.DeleteTagArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(tagDeleteCmd)
	assert.NoError(t, runTagDelete(cmd, []string{"name"}))
	assert.Equal(t, handlers.DeleteTagArgs{Name: "name"}, got)
}
