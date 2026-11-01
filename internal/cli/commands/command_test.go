package commands

import (
	"errors"
	"testing"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
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

// cmdGetCmd previously had no --profile flag registered at all, even though runCmdGet reads
// it via getProfile(cmd) - the flag silently had no effect. Guards against that regression.
func TestCmdGetCmd_Flags(t *testing.T) {
	assertFlag(t, cmdGetCmd.Flags(), "profile", "p", "string", "")
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
	var got handlers.AddCmdArgs
	stubHandlerCall(t, &callRunAddCmd, func(args handlers.AddCmdArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdAddCmd)
	require.NoError(t, cmd.Flags().Set("timeout", "30"))
	require.NoError(t, cmd.Flags().Set("description", "desc"))
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	err := runCmdAdd(cmd, []string{"myalias", "echo hi"})
	assert.NoError(t, err)

	assert.Equal(t, "myalias", got.Alias)
	assert.Equal(t, "echo hi", got.Template)
	assert.True(t, got.HasTemplate)
	assert.Equal(t, "desc", got.Description)
	assert.Equal(t, 30, got.Timeout)
	assert.Equal(t, "work", got.Profile)
}

func TestRunCmdAdd_NoArgsLeavesAliasAndTemplateUnset(t *testing.T) {
	var got handlers.AddCmdArgs
	stubHandlerCall(t, &callRunAddCmd, func(args handlers.AddCmdArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdAddCmd)
	err := runCmdAdd(cmd, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "", got.Alias)
	assert.False(t, got.HasTemplate)
}

func TestRunCmdAdd_PropagatesHandlerError(t *testing.T) {
	stubHandlerCall(t, &callRunAddCmd, func(args handlers.AddCmdArgs) error {
		return errors.New("boom")
	})

	cmd := freshCmdWithFlagsOf(cmdAddCmd)
	assert.EqualError(t, runCmdAdd(cmd, []string{}), "boom")
}

func TestRunCmdGet_ForwardsAliasAndProfile(t *testing.T) {
	var gotAlias, gotProfile string
	stubHandlerCall(t, &callRunGetCmd, func(name, profile string) error {
		gotAlias, gotProfile = name, profile
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdGetCmd)
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	assert.NoError(t, runCmdGet(cmd, []string{"myalias"}))
	assert.Equal(t, "myalias", gotAlias)
	assert.Equal(t, "work", gotProfile)
}

func TestRunCmdUpdate_EditFieldsRequiresEditFlag(t *testing.T) {
	stubHandlerCall(t, &callRunUpdateCmd, func(args handlers.UpdateCmdArgs) error {
		t.Fatal("handler should not be called when validation fails")
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdUpdateCmd)
	require.NoError(t, cmd.Flags().Set("edit-fields", "template"))
	err := runCmdUpdate(cmd, []string{"myalias"})
	assert.EqualError(t, err, "--edit-fields can only be used with --edit")
}

func TestRunCmdUpdate_EditFieldsWithEditFlagIsAllowed(t *testing.T) {
	var got handlers.UpdateCmdArgs
	stubHandlerCall(t, &callRunUpdateCmd, func(args handlers.UpdateCmdArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdUpdateCmd)
	require.NoError(t, cmd.Flags().Set("edit", "true"))
	require.NoError(t, cmd.Flags().Set("edit-fields", "template"))
	err := runCmdUpdate(cmd, []string{"myalias"})
	assert.NoError(t, err)
	assert.True(t, got.EditMode)
	assert.Equal(t, "template", got.EditFields)
}

func TestRunCmdUpdate_HappyPath(t *testing.T) {
	var got handlers.UpdateCmdArgs
	stubHandlerCall(t, &callRunUpdateCmd, func(args handlers.UpdateCmdArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdUpdateCmd)
	require.NoError(t, cmd.Flags().Set("template", "echo hi"))
	require.NoError(t, cmd.Flags().Set("description", "desc"))
	require.NoError(t, cmd.Flags().Set("cwd", "/tmp"))
	require.NoError(t, cmd.Flags().Set("timeout", "10"))
	err := runCmdUpdate(cmd, []string{"myalias"})
	assert.NoError(t, err)

	assert.Equal(t, "myalias", got.CurrentAlias)
	assert.Equal(t, "echo hi", got.Template)
	assert.True(t, got.HasTemplate)
	assert.Equal(t, "desc", got.Description)
	assert.True(t, got.HasDescription)
	assert.Equal(t, "/tmp", got.Cwd)
	assert.True(t, got.HasCwd)
	assert.Equal(t, 10, got.Timeout)
	assert.True(t, got.HasTimeout)
}

func TestRunCmdUpdate_PropagatesHandlerError(t *testing.T) {
	stubHandlerCall(t, &callRunUpdateCmd, func(args handlers.UpdateCmdArgs) error {
		return errors.New("boom")
	})

	cmd := freshCmdWithFlagsOf(cmdUpdateCmd)
	assert.EqualError(t, runCmdUpdate(cmd, []string{"myalias"}), "boom")
}

func TestRunCmdList_HappyPath(t *testing.T) {
	var got handlers.ListCmdArgs
	stubHandlerCall(t, &callRunListCmd, func(args handlers.ListCmdArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdListCmd)
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	err := runCmdList(cmd, []string{})
	assert.NoError(t, err)
	assert.Equal(t, "work", got.Profile)
}

func TestRunCmdList_PageAndNoPageConflict(t *testing.T) {
	stubHandlerCall(t, &callRunListCmd, func(args handlers.ListCmdArgs) error {
		t.Fatal("handler should not be called when --page and --no-page conflict")
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdListCmd)
	require.NoError(t, cmd.Flags().Set("page", "true"))
	require.NoError(t, cmd.Flags().Set("no-page", "true"))
	err := runCmdList(cmd, []string{})
	assert.Error(t, err)
}

func TestRunCmdSearch_HappyPath(t *testing.T) {
	var got handlers.SearchCmdArgs
	stubHandlerCall(t, &callRunSearchCmd, func(args handlers.SearchCmdArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdSearchCmd)
	err := runCmdSearch(cmd, []string{"term"})
	assert.NoError(t, err)
	assert.Equal(t, "term", got.Term)
}

func TestRunCmdDelete_ForwardsAliasAndProfile(t *testing.T) {
	var got handlers.DeleteCmdArgs
	stubHandlerCall(t, &callRunDeleteCmd, func(args handlers.DeleteCmdArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdDeleteCmd)
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	assert.NoError(t, runCmdDelete(cmd, []string{"myalias"}))
	assert.Equal(t, handlers.DeleteCmdArgs{Alias: "myalias", Profile: "work"}, got)
}

func TestRunCmdAttachTags_UsesPositionalTagNames(t *testing.T) {
	var got handlers.TagCmdArgs
	stubHandlerCall(t, &callRunTagCmd, func(args handlers.TagCmdArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdTagCmd)
	assert.NoError(t, runCmdAttachTags(cmd, []string{"myalias", "tag1", "tag2"}))
	assert.Equal(t, "myalias", got.Alias)
	assert.Equal(t, []string{"tag1", "tag2"}, got.TagNames)
}

func TestRunCmdDetachTags_UsesPositionalTagNames(t *testing.T) {
	var got handlers.TagCmdArgs
	stubHandlerCall(t, &callRunUntagCmd, func(args handlers.TagCmdArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(cmdUntagCmd)
	assert.NoError(t, runCmdDetachTags(cmd, []string{"myalias", "tag1"}))
	assert.Equal(t, "myalias", got.Alias)
	assert.Equal(t, []string{"tag1"}, got.TagNames)
}

// runCmdDetachTags must call the untag handler, not the attach handler - this was a
// copy/paste bug. This pins the wiring itself: the two subcommands must not share the same
// RunE function.
func TestCmdTagAndUntag_UseDistinctHandlerFunctions(t *testing.T) {
	assert.NotEqual(t,
		functionName(cmdTagCmd.RunE),
		functionName(cmdUntagCmd.RunE),
		"cmdTagCmd and cmdUntagCmd must not be wired to the same RunE function",
	)
}
