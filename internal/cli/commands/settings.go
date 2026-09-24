package commands

import (
	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
	"github.com/spf13/cobra"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "View and edit your CmdBox configuration",
}

var settingsShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current settings",
	Args:  cobra.NoArgs,
	RunE:  runSettingsShow,
}

var settingsEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit settings",
	Args:  cobra.NoArgs,
	RunE:  runSettingsEdit,
}

var settingsDataDirCmd = &cobra.Command{
	Use:   "data-dir",
	Short: "Open the CmdBox data directory",
	Args:  cobra.NoArgs,
	RunE:  runSettingsDataDir,
}

func init() {
	settingsShowCmd.Flags().StringSlice("field", nil, "Comma-separated list of setting fields to display")

	settingsEditCmd.Flags().BoolP("external", "e", false, "Open the settings file in your default text editor")

	settingsDataDirCmd.Flags().BoolP("print", "p", false, "Print the data directory path instead of opening it")

	settingsCmd.AddCommand(settingsShowCmd, settingsEditCmd, settingsDataDirCmd)
	rootCmd.AddCommand(settingsCmd)
}

func runSettingsShow(cmd *cobra.Command, args []string) error {
	return handlers.RunShowSettings(getFields(cmd))
}

func runSettingsEdit(cmd *cobra.Command, args []string) error {
	return handlers.RunEditSettings(getBoolFlag(cmd, "external"))
}

func runSettingsDataDir(cmd *cobra.Command, args []string) error {
	return handlers.RunSettingsDataDir(getBoolFlag(cmd, "print"))
}
