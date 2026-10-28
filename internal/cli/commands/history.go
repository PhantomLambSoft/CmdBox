package commands

import (
	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
	"github.com/spf13/cobra"
)

var historyCmd = &cobra.Command{
	Use:     "history",
	Aliases: []string{"hist"},
	Short:   "View and re-run past command executions",
}

var historyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List recent command executions",
	Args:  cobra.NoArgs,
	RunE:  runHistoryList,
}

var historyShowCmd = &cobra.Command{
	Use:   "show <index>",
	Short: "Show details of a history entry",
	Args:  cobra.ExactArgs(1),
	RunE:  runHistoryShow,
}

var historyRerunCmd = &cobra.Command{
	Use:   "rerun <index>",
	Short: "Re-execute a past command",
	Args:  cobra.ExactArgs(1),
	RunE:  runHistoryRerun,
}

var historyLastCmd = &cobra.Command{
	Use:   "last",
	Short: "Re-execute the most recent command",
	Args:  cobra.NoArgs,
	RunE:  runHistoryLast,
}

var historyClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear history entries",
	Args:  cobra.NoArgs,
	RunE:  runHistoryClear,
}

// bangBangCmd implements `cb !!`, the documented shortcut for history last.
var bangBangCmd = &cobra.Command{
	Use:    "!!",
	Short:  "Re-execute the most recent command (shortcut for 'history last')",
	Args:   cobra.NoArgs,
	Hidden: true,
	RunE:   runHistoryLast,
}

func init() {
	historyListCmd.Flags().String("alias", "", "Filter to a specific command alias")
	historyListCmd.Flags().String("order", "", "Field to order by")
	historyListCmd.Flags().StringSlice("field", nil, "Fields to display (repeatable)")
	historyListCmd.Flags().Int("limit", 0, "Number of entries to show")
	historyListCmd.Flags().Bool("page", false, "Force paging on")
	historyListCmd.Flags().Bool("no-page", false, "Force paging off")
	historyListCmd.Flags().StringP("profile", "p", "", "Profile to list from")
	historyListCmd.MarkFlagsMutuallyExclusive("page", "no-page")

	historyShowCmd.Flags().StringP("alias", "", "", "Filter to a specific command alias")
	historyShowCmd.Flags().StringP("profile", "p", "", "Profile to list from")

	historyRerunCmd.Flags().StringP("alias", "", "", "Filter to a specific command alias")
	historyRerunCmd.Flags().StringP("profile", "p", "", "Profile to list from")

	historyClearCmd.Flags().BoolP("yes", "y", false, "Skip the confirmation prompt")
	historyClearCmd.Flags().String("alias", "", "Clear history for a specific alias only")
	historyClearCmd.Flags().StringP("profile", "p", "", "Profile to list from")

	historyCmd.AddCommand(historyListCmd, historyShowCmd, historyRerunCmd, historyLastCmd, historyClearCmd)

	rootCmd.AddCommand(historyCmd, bangBangCmd) // bangBangCmd is a root level command, not a history subcommand
}

func runHistoryList(cmd *cobra.Command, args []string) error {
	page, err := getPage(cmd)
	if err != nil {
		return err
	}
	raw := handlers.ListHistoryArgs{
		Order:   getOrder(cmd),
		Fields:  getFields(cmd),
		Page:    page,
		Profile: getProfile(cmd),
	}
	if cmd.Flags().Changed("limit") {
		raw.Limit = getLimit(cmd)
		raw.HasLimit = true
	}
	if cmd.Flags().Changed("alias") {
		raw.Alias = getStringFlag(cmd, "alias")
		raw.HasAlias = true
	}
	return handlers.RunListHistory(raw)
}

func runHistoryShow(cmd *cobra.Command, args []string) error {
	var alias *string
	var profileName *string
	if cmd.Flags().Changed("alias") {
		rawAlias := getStringFlag(cmd, "alias")
		alias = &rawAlias
	}
	if cmd.Flags().Changed("profile") {
		rawProfile := getStringFlag(cmd, "profile")
		profileName = &rawProfile
	}
	return handlers.RunShowHistory(args[0], alias, profileName)
}

func runHistoryRerun(cmd *cobra.Command, args []string) error {
	var alias *string
	var profileName *string
	if cmd.Flags().Changed("alias") {
		rawAlias := getStringFlag(cmd, "alias")
		alias = &rawAlias
	}
	if cmd.Flags().Changed("profile") {
		rawProfile := getStringFlag(cmd, "profile")
		profileName = &rawProfile
	}
	return handlers.RunRerunHistory(args[0], alias, profileName)
}

func runHistoryLast(cmd *cobra.Command, args []string) error {
	return handlers.RunLastHistory()
}

func runHistoryClear(cmd *cobra.Command, args []string) error {
	raw := handlers.ClearHistoryArgs{Yes: getBoolFlag(cmd, "yes")}
	if cmd.Flags().Changed("alias") {
		raw.Alias = getStringFlag(cmd, "alias")
		raw.HasAlias = true
	}
	if cmd.Flags().Changed("profile") {
		raw.Profile = getStringFlag(cmd, "profile")
		raw.HasProfile = true
	}
	return handlers.RunClearHistory(raw)
}
