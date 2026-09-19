package commands

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
	"github.com/spf13/cobra"
)

var cmdCmd = &cobra.Command{
	Use:   "cmd",
	Short: "Manage saved commands",
}

var cmdAddCmd = &cobra.Command{
	Use:   "add [alias] [template]",
	Short: "Add a new command",
	Args:  cobra.RangeArgs(0, 2),
	RunE:  runCmdAdd,
}

var cmdGetCmd = &cobra.Command{
	Use:   "get <alias>",
	Short: "Get a command",
	Args:  cobra.ExactArgs(1),
	RunE:  runCmdGet,
}

var cmdUpdateCmd = &cobra.Command{
	Use:     "update <alias>",
	Aliases: []string{"edit"},
	Short:   "Update an existing command",
	Args:    cobra.ExactArgs(1),
	RunE:    runCmdUpdate,
}

var cmdListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List commands",
	RunE:    runCmdList,
}

var cmdSearchCmd = &cobra.Command{
	Use:     "search <term>",
	Aliases: []string{"find"},
	Short:   "Search for commands",
	Args:    cobra.ExactArgs(1),
	RunE:    runCmdSearch,
}

var cmdDeleteCmd = &cobra.Command{
	Use:     "delete <alias>",
	Aliases: []string{"rm", "del", "remove"},
	Short:   "Delete a command",
	Args:    cobra.ExactArgs(1),
	RunE:    runCmdDelete,
}

var cmdTagCmd = &cobra.Command{
	Use:   "tag <alias> <tag>...",
	Short: "Tag a command with one or more tags",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runCmdAttachTags,
}

var cmdUntagCmd = &cobra.Command{
	Use:   "untag <alias> <tag>...",
	Short: "Untag a command",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runCmdDetachTags,
}

func init() {
	cmdAddCmd.Flags().StringP("description", "d", "", "Description of the command")
	cmdAddCmd.Flags().BoolP("interactive", "i", false, "Prompt for every field")
	cmdAddCmd.Flags().String("cwd", "", "Working directory to store on the command")
	cmdAddCmd.Flags().String("shell", "", "Shell to store on the command")
	cmdAddCmd.Flags().StringArray("env", nil, "Environment variable KEY=VALUE (repeatable)")
	cmdAddCmd.Flags().Int("timeout", 0, "Timeout in seconds")
	cmdAddCmd.Flags().StringP("profile", "p", "", "Profile to add this command to")

	cmdUpdateCmd.Flags().StringP("template", "t", "", "New template")
	cmdUpdateCmd.Flags().StringP("description", "d", "", "New description")
	cmdUpdateCmd.Flags().String("alias", "", "New alias (rename)")
	cmdUpdateCmd.Flags().String("cwd", "", "New working directory")
	cmdUpdateCmd.Flags().Bool("clear-cwd", false, "Clear the stored working directory")
	cmdUpdateCmd.Flags().String("shell", "", "New shell")
	cmdUpdateCmd.Flags().Bool("clear-shell", false, "Clear the stored shell")
	cmdUpdateCmd.Flags().StringArray("env", nil, "New environment variables (replaces stored)")
	cmdUpdateCmd.Flags().Bool("clear-env", false, "Clear stored environment variables")
	cmdUpdateCmd.Flags().Int("timeout", 0, "New timeout")
	cmdUpdateCmd.Flags().Bool("clear-timeout", false, "Clear the stored timeout")
	cmdUpdateCmd.Flags().String("set", "", "Comma-separated key=value pairs")
	cmdUpdateCmd.Flags().BoolP("edit", "e", false, "Interactively edit current fields")
	cmdUpdateCmd.Flags().String("edit-fields", "", "Restrict --edit to these fields")
	cmdUpdateCmd.Flags().StringP("profile", "p", "", "Profile to look in")

	cmdUpdateCmd.MarkFlagsMutuallyExclusive("cwd", "clear-cwd")
	cmdUpdateCmd.MarkFlagsMutuallyExclusive("shell", "clear-shell")
	cmdUpdateCmd.MarkFlagsMutuallyExclusive("env", "clear-env")
	cmdUpdateCmd.MarkFlagsMutuallyExclusive("timeout", "clear-timeout")

	cmdListCmd.Flags().StringSlice("tag", nil, "Filter by tag (repeatable)")
	cmdListCmd.Flags().String("order", "", "Field to order by")
	cmdListCmd.Flags().StringSlice("field", nil, "Fields to display (repeatable)")
	cmdListCmd.Flags().Int("limit", 0, "Max results")
	cmdListCmd.Flags().Bool("page", false, "Force paging on")
	cmdListCmd.Flags().Bool("no-page", false, "Force paging off")
	cmdListCmd.Flags().StringP("profile", "p", "", "Profile to list from")
	cmdListCmd.MarkFlagsMutuallyExclusive("page", "no-page")

	cmdSearchCmd.Flags().StringSlice("in", nil, "Fields to search in (repeatable)")
	cmdSearchCmd.Flags().StringSlice("field", nil, "Fields to display (repeatable)")
	cmdSearchCmd.Flags().Int("limit", 0, "Max results")
	cmdSearchCmd.Flags().Bool("page", false, "Force paging on")
	cmdSearchCmd.Flags().Bool("no-page", false, "Force paging off")
	cmdSearchCmd.Flags().StringP("profile", "p", "", "Profile to search in")
	cmdSearchCmd.MarkFlagsMutuallyExclusive("page", "no-page")

	cmdDeleteCmd.Flags().StringP("profile", "p", "", "Profile to delete from")
	cmdTagCmd.Flags().StringP("profile", "p", "", "Profile to look in")
	cmdUntagCmd.Flags().StringP("profile", "p", "", "Profile to look in")

	cmdCmd.AddCommand(
		cmdAddCmd,
		cmdGetCmd,
		cmdUpdateCmd,
		cmdListCmd,
		cmdSearchCmd,
		cmdDeleteCmd,
		cmdTagCmd,
		cmdUntagCmd,
	)
	rootCmd.AddCommand(cmdCmd)
}

func runCmdAdd(cmd *cobra.Command, args []string) error {
	profile := getProfile(cmd)
	interactive := getInteractiveFlag(cmd)
	description := getDescriptionFlag(cmd)
	cwd := getStringFlag(cmd, "cwd")
	shell := getStringFlag(cmd, "shell")
	env, _ := cmd.Flags().GetStringArray("env")

	raw := handlers.AddCmdArgs{
		Description: description,
		Interactive: interactive,
		Cwd:         cwd,
		Shell:       shell,
		Env:         env,
		Profile:     profile,
	}
	if len(args) > 0 {
		raw.Alias = args[0]
	}
	if len(args) > 1 {
		raw.Template = args[1]
		raw.HasTemplate = true
	}
	if cmd.Flags().Changed("timeout") {
		raw.Timeout, _ = cmd.Flags().GetInt("timeout")
	}
	return handlers.RunAddCmd(raw)
}

func runCmdGet(cmd *cobra.Command, args []string) error {
	profile := getProfile(cmd)
	return handlers.RunGetCmd(args[0], profile)
}

func runCmdUpdate(cmd *cobra.Command, args []string) error {
	editMode := getEditMode(cmd)
	editFields := getEditFields(cmd)

	if editFields != "" && !editMode {
		return fmt.Errorf("--edit-fields can only be used with --edit")
	}

	raw := handlers.UpdateCmdArgs{
		CurrentAlias: args[0],
		EditMode:     editMode,
		EditFields:   editFields,
		SetPairs:     getStringFlag(cmd, "set"),
		Profile:      getProfile(cmd),
	}

	if cmd.Flags().Changed("template") {
		raw.Template = getStringFlag(cmd, "template")
		raw.HasTemplate = true
	}
	if cmd.Flags().Changed("description") {
		raw.Description = getStringFlag(cmd, "description")
		raw.HasDescription = true
	}
	if cmd.Flags().Changed("alias") {
		raw.NewAlias = getStringFlag(cmd, "alias")
	}
	if cmd.Flags().Changed("cwd") {
		raw.Cwd = getStringFlag(cmd, "cwd")
		raw.HasCwd = true
	}
	raw.ClearCwd = getBoolFlag(cmd, "clear-cwd")
	if cmd.Flags().Changed("shell") {
		raw.Shell = getStringFlag(cmd, "shell")
		raw.HasShell = true
	}
	raw.ClearShell = getBoolFlag(cmd, "clear-shell")
	if cmd.Flags().Changed("env") {
		raw.Env, _ = cmd.Flags().GetStringArray("env")
		raw.HasEnv = true
	}
	raw.ClearEnv = getBoolFlag(cmd, "clear-env")
	if cmd.Flags().Changed("timeout") {
		raw.Timeout, _ = cmd.Flags().GetInt("timeout")
		raw.HasTimeout = true
	}
	raw.ClearTimeout = getBoolFlag(cmd, "clear-timeout")

	return handlers.RunUpdateCmd(raw)
}

func runCmdList(cmd *cobra.Command, args []string) error {
	page, err := getPage(cmd)
	if err != nil {
		return err
	}
	raw := handlers.ListCmdArgs{
		Order:   getOrder(cmd),
		Fields:  getFields(cmd),
		Page:    page,
		Profile: getProfile(cmd),
		Tags:    getTags(cmd),
	}
	if cmd.Flags().Changed("limit") {
		raw.Limit = getLimit(cmd)
		raw.HasLimit = true
	}
	return handlers.RunListCmd(raw)
}

func runCmdSearch(cmd *cobra.Command, args []string) error {
	page, err := getPage(cmd)
	if err != nil {
		return err
	}
	searchIn, _ := cmd.Flags().GetStringSlice("in")
	raw := handlers.SearchCmdArgs{
		Term:     args[0],
		SearchIn: searchIn,
		Fields:   getFields(cmd),
		Page:     page,
		Profile:  getProfile(cmd),
	}
	if cmd.Flags().Changed("limit") {
		raw.Limit = getLimit(cmd)
		raw.HasLimit = true
	}
	return handlers.RunSearchCmd(raw)
}

func runCmdDelete(cmd *cobra.Command, args []string) error {
	return handlers.RunDeleteCmd(handlers.DeleteCmdArgs{
		Alias:   args[0],
		Profile: getProfile(cmd),
	})
}

func runCmdAttachTags(cmd *cobra.Command, args []string) error {
	return handlers.RunTagCmd(handlers.TagCmdArgs{
		Alias:    args[0],
		TagNames: getTags(cmd),
		Profile:  getProfile(cmd),
	})
}

func runCmdDetachTags(cmd *cobra.Command, args []string) error {
	return handlers.RunTagCmd(handlers.TagCmdArgs{
		Alias:    args[0],
		TagNames: getTags(cmd),
		Profile:  getProfile(cmd),
	})
}
