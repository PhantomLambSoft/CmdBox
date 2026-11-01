package commands

import (
	"errors"
	"testing"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
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

// varGetCmd previously had no --profile flag registered at all, even though runVarGet reads
// it via getProfile(cmd) - the flag silently had no effect. Guards against that regression.
func TestVarGetCmd_Flags(t *testing.T) {
	assertFlag(t, varGetCmd.Flags(), "profile", "p", "string", "")
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

func TestRunVarAdd_BuildsArgsFromPositionalsAndFlags(t *testing.T) {
	var got handlers.AddVarArgs
	stubHandlerCall(t, &callRunAddVar, func(args handlers.AddVarArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(varAddCmd)
	require.NoError(t, cmd.Flags().Set("interactive", "true"))
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	assert.NoError(t, runVarAdd(cmd, []string{"name", "value"}))

	assert.Equal(t, handlers.AddVarArgs{
		Name: "name", Value: "value", HasValue: true, Interactive: true, Profile: "work",
	}, got)
}

func TestRunVarAdd_NoArgsLeavesNameAndValueUnset(t *testing.T) {
	var got handlers.AddVarArgs
	stubHandlerCall(t, &callRunAddVar, func(args handlers.AddVarArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(varAddCmd)
	assert.NoError(t, runVarAdd(cmd, []string{}))
	assert.Equal(t, "", got.Name)
	assert.False(t, got.HasValue)
}

func TestRunVarAdd_PropagatesHandlerError(t *testing.T) {
	stubHandlerCall(t, &callRunAddVar, func(args handlers.AddVarArgs) error {
		return errors.New("boom")
	})

	cmd := freshCmdWithFlagsOf(varAddCmd)
	assert.EqualError(t, runVarAdd(cmd, []string{}), "boom")
}

func TestRunVarGet_ForwardsNameAndProfile(t *testing.T) {
	var gotName, gotProfile string
	stubHandlerCall(t, &callRunGetVar, func(name, profile string) error {
		gotName, gotProfile = name, profile
		return nil
	})

	cmd := freshCmdWithFlagsOf(varGetCmd)
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	assert.NoError(t, runVarGet(cmd, []string{"name"}))
	assert.Equal(t, "name", gotName)
	assert.Equal(t, "work", gotProfile)
}

func TestRunVarUpdate_HappyPath(t *testing.T) {
	var got handlers.UpdateVarArgs
	stubHandlerCall(t, &callRunUpdateVar, func(args handlers.UpdateVarArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(varUpdateCmd)
	require.NoError(t, cmd.Flags().Set("value", "newvalue"))
	assert.NoError(t, runVarUpdate(cmd, []string{"name"}))

	assert.Equal(t, "name", got.CurrentName)
	assert.Equal(t, "newvalue", got.NewValue)
	assert.True(t, got.HasValue)
}

func TestRunVarList_HappyPath(t *testing.T) {
	var got handlers.ListVarArgs
	stubHandlerCall(t, &callRunListVar, func(args handlers.ListVarArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(varListCmd)
	require.NoError(t, cmd.Flags().Set("tag", "a,b"))
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	assert.NoError(t, runVarList(cmd, []string{}))

	assert.Equal(t, []string{"a", "b"}, got.Tags)
	assert.Equal(t, "work", got.Profile)
}

func TestRunVarList_PageAndNoPageConflict(t *testing.T) {
	stubHandlerCall(t, &callRunListVar, func(args handlers.ListVarArgs) error {
		t.Fatal("handler should not be called when --page and --no-page conflict")
		return nil
	})

	cmd := freshCmdWithFlagsOf(varListCmd)
	require.NoError(t, cmd.Flags().Set("page", "true"))
	require.NoError(t, cmd.Flags().Set("no-page", "true"))
	assert.Error(t, runVarList(cmd, []string{}))
}

func TestRunVarSearch_HappyPath(t *testing.T) {
	var got handlers.SearchParams
	stubHandlerCall(t, &callRunSearchVar, func(args handlers.SearchParams) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(varSearchCmd)
	require.NoError(t, cmd.Flags().Set("in", "name,value"))
	assert.NoError(t, runVarSearch(cmd, []string{"term"}))

	assert.Equal(t, "term", got.Term)
	assert.Equal(t, []string{"name", "value"}, got.SearchIn)
}

func TestRunVarDelete_ForwardsNameAndProfile(t *testing.T) {
	var got handlers.DeleteVarArgs
	stubHandlerCall(t, &callRunDeleteVar, func(args handlers.DeleteVarArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(varDeleteCmd)
	require.NoError(t, cmd.Flags().Set("profile", "work"))
	assert.NoError(t, runVarDelete(cmd, []string{"name"}))
	assert.Equal(t, handlers.DeleteVarArgs{Name: "name", Profile: "work"}, got)
}

func TestRunVarAttachTags_UsesPositionalTagNames(t *testing.T) {
	var got handlers.TagVarArgs
	stubHandlerCall(t, &callRunTagVar, func(args handlers.TagVarArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(varAttachTagCmd)
	assert.NoError(t, runVarAttachTags(cmd, []string{"name", "tag1", "tag2"}))
	assert.Equal(t, "name", got.Name)
	assert.Equal(t, []string{"tag1", "tag2"}, got.TagNames)
}

func TestRunVarDetachTags_UsesPositionalTagNames(t *testing.T) {
	var got handlers.TagVarArgs
	stubHandlerCall(t, &callRunUntagVar, func(args handlers.TagVarArgs) error {
		got = args
		return nil
	})

	cmd := freshCmdWithFlagsOf(varDetachTagCmd)
	assert.NoError(t, runVarDetachTags(cmd, []string{"name", "tag1"}))
	assert.Equal(t, "name", got.Name)
	assert.Equal(t, []string{"tag1"}, got.TagNames)
}

// runVarAttachTags/runVarDetachTags must call distinct handlers (RunTagVar vs RunUntagVar) -
// this is the pattern the (previously buggy) cmd equivalents were fixed to match.
func TestVarAttachAndDetachTag_UseDistinctHandlerFunctions(t *testing.T) {
	assert.NotEqual(t,
		functionName(varAttachTagCmd.RunE),
		functionName(varDetachTagCmd.RunE),
	)
}
