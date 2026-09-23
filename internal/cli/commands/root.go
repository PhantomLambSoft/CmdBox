package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cb",
	Short: "CmdBox - a tool for storing and recalling commands with many helpful quality of life features",
}

func Execute() {
	args := resolveArgs(os.Args[1:])
	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
