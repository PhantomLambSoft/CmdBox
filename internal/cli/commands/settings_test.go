package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSettingsCmd_Metadata(t *testing.T) {
	assert.Equal(t, "settings", settingsCmd.Name())
	assert.Contains(t, rootCmd.Commands(), settingsCmd)
}

func TestSettingsCmd_Subcommands_Wired(t *testing.T) {
	want := []*cobra.Command{settingsShowCmd, settingsEditCmd, settingsDataDirCmd}
	for _, c := range want {
		assert.Contains(t, settingsCmd.Commands(), c, "expected %s to be a child of settingsCmd", c.Name())
	}
}

func TestSettingsShowCmd_Args(t *testing.T) {
	assert.NoError(t, settingsShowCmd.ValidateArgs([]string{}))
	assert.Error(t, settingsShowCmd.ValidateArgs([]string{"extra"}))
}

func TestSettingsEditCmd_Args(t *testing.T) {
	assert.NoError(t, settingsEditCmd.ValidateArgs([]string{}))
	assert.Error(t, settingsEditCmd.ValidateArgs([]string{"extra"}))
}

func TestSettingsDataDirCmd_Args(t *testing.T) {
	assert.NoError(t, settingsDataDirCmd.ValidateArgs([]string{}))
	assert.Error(t, settingsDataDirCmd.ValidateArgs([]string{"extra"}))
}

func TestSettingsShowCmd_Flags(t *testing.T) {
	assertFlag(t, settingsShowCmd.Flags(), "field", "", "stringSlice", "[]")
}

func TestRunSettingsShow_FieldFilterIsRead(t *testing.T) {
	cmd := freshCmdWithFlagsOf(settingsShowCmd)
	require.NoError(t, cmd.Flags().Set("field", "cwd,shell"))
	assert.Equal(t, []string{"cwd", "shell"}, getFields(cmd))
}

func TestSettingsEditCmd_Flags(t *testing.T) {
	assertFlag(t, settingsEditCmd.Flags(), "external", "e", "bool", "false")
}

func TestSettingsDataDirCmd_Flags(t *testing.T) {
	assertFlag(t, settingsDataDirCmd.Flags(), "print", "p", "bool", "false")
}

func TestRunSettingsShow(t *testing.T) {
	cmd := freshCmdWithFlagsOf(settingsShowCmd)
	assert.NoError(t, runSettingsShow(cmd, []string{}))
}

func TestRunSettingsEdit(t *testing.T) {
	cmd := freshCmdWithFlagsOf(settingsEditCmd)
	assert.NoError(t, runSettingsEdit(cmd, []string{}))

	require.NoError(t, cmd.Flags().Set("external", "true"))
	assert.NoError(t, runSettingsEdit(cmd, []string{}))
}

func TestRunSettingsDataDir(t *testing.T) {
	cmd := freshCmdWithFlagsOf(settingsDataDirCmd)
	assert.NoError(t, runSettingsDataDir(cmd, []string{}))

	require.NoError(t, cmd.Flags().Set("print", "true"))
	assert.NoError(t, runSettingsDataDir(cmd, []string{}))
}
