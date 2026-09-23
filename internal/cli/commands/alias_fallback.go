package commands

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// resolveArgs processes command-line arguments, prepends context command if necessary, and ensures help/completion commands exist.
func resolveArgs(args []string) []string {
	if len(args) == 0 || strings.HasPrefix(args[0], "-") {
		return args
	}

	// "help" and "completion" are registered lazily in Execute(). Force them to exist now so "cb help" is not
	// treated as a stored command.
	rootCmd.InitDefaultHelpCmd()
	rootCmd.InitDefaultCompletionCmd()

	first := args[0]

	if isTopLevelCommand(first) {
		return args
	}
	if isCmdSubcommand(first) {
		return append([]string{"cmd"}, args...)
	}
	return append([]string{"run"}, args...)
}

func isTopLevelCommand(name string) bool {
	return isInCommands(name, rootCmd.Commands())
}

func isCmdSubcommand(name string) bool {
	return isInCommands(name, cmdCmd.Commands())
}

func isInCommands(name string, commands []*cobra.Command) bool {
	for _, c := range commands {
		if name == c.Name() {
			return true
		}
		if slices.Contains(c.Aliases, name) {
			return true
		}
	}
	return false
}
