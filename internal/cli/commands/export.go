package commands

import (
	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
	"github.com/spf13/cobra"
)

func runExportCmds(cmd *cobra.Command, args []string) error {
	aliases, _ := cmd.Flags().GetStringArray("aliases")
	tag := getStringFlag(cmd, "tag")
	flatten := getBoolFlag(cmd, "flatten")
	output := getStringFlag(cmd, "output")
	cmdProfile := getStringFlag(cmd, "cmd-profile")
	varProfile := getStringFlag(cmd, "var-profile")

	raw := handlers.RunExportCmdArgs{
		Aliases:    aliases,
		Tag:        tag,
		Flatten:    flatten,
		Output:     output,
		CmdProfile: cmdProfile,
		VarProfile: varProfile,
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
