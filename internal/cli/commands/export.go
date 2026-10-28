package commands

import (
	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export commands and/or variables",
}

var exportCommandsCmd = &cobra.Command{
	Use:   "cmds",
	Short: "Export commands",
	RunE:  runExportCmds,
}

var exportVariablesCmd = &cobra.Command{
	Use:   "vars",
	Short: "Export variables",
	RunE:  runExportVars,
}

var exportAllCmd = &cobra.Command{
	Use:   "all",
	Short: "Export commands and variables",
	RunE:  runExportAll,
}

func init() {
	exportCommandsCmd.Flags().StringArrayP("aliases", "a", nil, "Export commands with aliases")
	exportCommandsCmd.Flags().StringP("tag", "t", "", "Export commands with tag")
	exportCommandsCmd.Flags().BoolP("flatten", "f", false, "Export commands in flatten mode")
	exportCommandsCmd.Flags().StringP("output", "o", "", "Export commands to file")
	exportCommandsCmd.Flags().StringP("cmd-profile", "c", "", "Export commands to profile")

	exportVariablesCmd.Flags().StringArrayP("names", "n", nil, "Export variables with names")
	exportVariablesCmd.Flags().StringP("tag", "t", "", "Export variables with tag")
	exportVariablesCmd.Flags().BoolP("flatten", "f", false, "Export variables in flatten mode")
	exportVariablesCmd.Flags().StringP("output", "o", "", "Export variables to file")
	exportVariablesCmd.Flags().StringP("var-profile", "v", "", "Export variables to profile")

	exportAllCmd.Flags().StringP("tag", "t", "", "Export commands and variables with tag")
	exportAllCmd.Flags().BoolP("flatten", "f", false, "Export commands and variables in flatten mode")
	exportAllCmd.Flags().StringP("output", "o", "", "Export commands and variables to file")
	exportAllCmd.Flags().StringP("cmd-profile", "c", "", "Export commands and variables to profile")
	exportAllCmd.Flags().StringP("var-profile", "v", "", "Export commands and variables to profile")

	exportCmd.AddCommand(exportCommandsCmd, exportVariablesCmd, exportAllCmd)
	rootCmd.AddCommand(exportCmd)
}

func runExportCmds(cmd *cobra.Command, args []string) error {
	aliases, _ := cmd.Flags().GetStringArray("aliases")
	tag := getStringFlag(cmd, "tag")
	flatten := getBoolFlag(cmd, "flatten")
	output := getStringFlag(cmd, "output")
	cmdProfile := getStringFlag(cmd, "cmd-profile")

	raw := handlers.RunExportCmdArgs{
		Aliases:    aliases,
		Tag:        tag,
		Flatten:    flatten,
		Output:     output,
		CmdProfile: cmdProfile,
	}

	return handlers.RunExportCmds(raw)
}

func runExportVars(cmd *cobra.Command, args []string) error {
	names, _ := cmd.Flags().GetStringArray("names")
	tag := getStringFlag(cmd, "tag")
	flatten := getBoolFlag(cmd, "flatten")
	output := getStringFlag(cmd, "output")
	varProfile := getStringFlag(cmd, "var-profile")

	raw := handlers.RunExportVarArgs{
		Names:      names,
		Tag:        tag,
		Flatten:    flatten,
		Output:     output,
		VarProfile: varProfile,
	}

	return handlers.RunExportVars(raw)
}

func runExportAll(cmd *cobra.Command, args []string) error {
	flatten := getBoolFlag(cmd, "flatten")
	output := getStringFlag(cmd, "output")
	cmdProfile := getStringFlag(cmd, "cmd-profile")
	varProfile := getStringFlag(cmd, "var-profile")

	raw := handlers.RunExportAllArgs{
		Flatten:    flatten,
		Output:     output,
		CmdProfile: cmdProfile,
		VarProfile: varProfile,
	}

	return handlers.RunExportAll(raw)
}
