package commands

import (
	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
	"github.com/spf13/cobra"
)

var cmdImport = &cobra.Command{
	Use:   "import",
	Short: "Import commands and variables from a file",
	Args:  cobra.ExactArgs(1),
}

func runImport(cmd *cobra.Command, args []string) error {
	overwrite := getBoolFlag(cmd, "overwrite")
	preview := getBoolFlag(cmd, "preview")
	profile := getProfile(cmd)

	raw := handlers.RunImportArgs{
		Path:      args[0],
		Overwrite: overwrite,
		Preview:   preview,
		Profile:   profile,
	}

	return handlers.RunImportFile(raw)
}
