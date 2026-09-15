package commands

import (
	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
	"github.com/spf13/cobra"
)

var varCmd = &cobra.Command{
	Use:   "var",
	Short: "Manage saved variables",
}

var varAddCmd = &cobra.Command{
	Use:   "add [name] [value]",
	Short: "Add a new variable",
	Args:  cobra.RangeArgs(0, 2),
	RunE:  runVarAdd,
}

var varGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a variable",
	Args:  cobra.ExactArgs(1),
	RunE:  runVarGet,
}

var varUpdateCmd = &cobra.Command{
	Use:     "update <name>",
	Aliases: []string{"edit"},
	Short:   "Update a variable",
	Args:    cobra.ExactArgs(1),
	RunE:    runVarUpdate,
}

var varListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List variables",
	RunE:    runVarList,
}

var varSearchCmd = &cobra.Command{
	Use:     "search <term>",
	Aliases: []string{"find"},
	Short:   "Search variables",
	Args:    cobra.ExactArgs(1),
	RunE:    runVarSearch,
}

var varDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm", "del", "remove"},
	Short:   "Delete a variable",
	Args:    cobra.ExactArgs(1),
	RunE:    runVarDelete,
}

var varAttachTagCmd = &cobra.Command{
	Use:   "tag <name> <tag>...",
	Short: "Tag a variable with one or more tags",
	Args:  cobra.MinimumNArgs(2),
	RunE:  runVarAttachTags,
}

var varDetachTagCmd = &cobra.Command{
	Use:     "detach-tag <name> <tag>",
	Aliases: []string{"remove-tag"},
	Short:   "Remove one or more tags from a variable",
	Args:    cobra.MinimumNArgs(2),
	RunE:    runVarDetachTags,
}

func init() {
	varAddCmd.Flags().BoolP("interactive", "i", false, "Prompt for every field")
	varAddCmd.Flags().String("profile", "", "The profile to add this variable to")

	varUpdateCmd.Flags().String("value", "", "New value")
	varUpdateCmd.Flags().String("name", "", "New name (rename)")
	varUpdateCmd.Flags().String("set", "", "Comma-separated key=value pairs")
	varUpdateCmd.Flags().BoolP("edit", "e", false, "Interactively edit current fields")
	varUpdateCmd.Flags().StringP("edit-fields", "ef", "", "Restrict --edit to these fields")
	varUpdateCmd.Flags().String("profile", "", "Profile to look in")

	varListCmd.Flags().StringSlice("tag", nil, "Filter by tag (repeatable)")
	varListCmd.Flags().String("order", "", "Field to order by")
	varListCmd.Flags().StringSlice("field", nil, "Fields to display (repeatable)")
	varListCmd.Flags().Int("limit", 0, "Max results")
	varListCmd.Flags().Bool("page", false, "Force paging on")
	varListCmd.Flags().Bool("no-page", false, "Force paging off")
	varListCmd.Flags().String("profile", "", "Profile to list from")

	varSearchCmd.Flags().StringSlice("in", nil, "Fields to search in (repeatable)")
	varSearchCmd.Flags().StringSlice("field", nil, "Fields to display (repeatable)")
	varSearchCmd.Flags().Int("limit", 0, "Max results")
	varSearchCmd.Flags().Bool("page", false, "Force paging on")
	varSearchCmd.Flags().Bool("no-page", false, "Force paging off")
	varSearchCmd.Flags().String("profile", "", "Profile to search in")

	varDeleteCmd.Flags().String("profile", "", "Profile to delete from")

	varAttachTagCmd.Flags().String("profile", "", "Profile to look in")
	varDetachTagCmd.Flags().String("profile", "", "Profile to look in")

	varCmd.AddCommand(
		varAddCmd,
		varGetCmd,
		varUpdateCmd,
		varListCmd,
		varSearchCmd,
		varDeleteCmd,
		varAttachTagCmd,
		varDetachTagCmd,
	)
	rootCmd.AddCommand(varCmd)
}

func runVarAdd(cmd *cobra.Command, args []string) error {
	interactive, _ := cmd.Flags().GetBool("interactive")
	profile := getProfile(cmd)

	raw := handlers.AddVarArgs{Interactive: interactive, Profile: profile}
	if len(args) > 0 {
		raw.Name = args[0]
	}
	if len(args) > 1 {
		raw.Value = args[1]
		raw.HasValue = true
	}
	return handlers.RunAddVar(raw)
}

func runVarGet(cmd *cobra.Command, args []string) error {
	profile := getProfile(cmd)
	return handlers.RunGetVar(args[0], profile)
}

func runVarUpdate(cmd *cobra.Command, args []string) error {
	editMode, _ := cmd.Flags().GetBool("edit")
	setParis := getStringFlag(cmd, "set")
	editFields := getStringFlag(cmd, "edit-fields")
	newName := getStringFlag(cmd, "name")
	profile := getProfile(cmd)

	raw := handlers.UpdateVarArgs{
		CurrentName: args[0],
		SetPairs:    setParis,
		EditMode:    editMode,
		EditFields:  editFields,
		NewName:     newName,
		Profile:     profile,
	}
	if cmd.Flags().Changed("value") {
		raw.NewValue, _ = cmd.Flags().GetString("value")
		raw.HasValue = true
	}
	return handlers.RunUpdateVar(raw)
}

func runVarList(cmd *cobra.Command, args []string) error {
	tags := getTags(cmd)
	order := getOrder(cmd)
	fields := getFields(cmd)
	profile := getProfile(cmd)

	raw := handlers.ListVarArgs{Tags: tags, Order: order, Fields: fields, Profile: profile}
	if cmd.Flags().Changed("limit") {
		raw.Limit = getLimit(cmd)
		raw.HasLimit = true
	}

	page, err := getPage(cmd)
	if err != nil {
		return err
	}
	raw.Page = page

	return handlers.RunListVar(raw)
}

func runVarSearch(cmd *cobra.Command, args []string) error {
	profile := getProfile(cmd)
	fields := getFields(cmd)
	searchIn, _ := cmd.Flags().GetStringSlice("in")

	raw := handlers.SearchVarArgs{Term: args[0], Fields: fields, SearchIn: searchIn, Profile: profile}
	if cmd.Flags().Changed("limit") {
		raw.Limit = getLimit(cmd)
		raw.HasLimit = true
	}

	page, err := getPage(cmd)
	if err != nil {
		return err
	}
	raw.Page = page

	return handlers.RunSearchVar(raw)
}

func runVarDelete(cmd *cobra.Command, args []string) error {
	profile := getProfile(cmd)
	return handlers.RunDeleteVar(handlers.DeleteVarArgs{Name: args[0], Profile: profile})
}

func runVarAttachTags(cmd *cobra.Command, args []string) error {
	profile := getProfile(cmd)
	return handlers.RunTagVar(handlers.TagVarArgs{
		Name:     args[0],
		TagNames: args[1:],
		Profile:  profile,
	})
}

func runVarDetachTags(cmd *cobra.Command, args []string) error {
	profile := getProfile(cmd)
	return handlers.RunTagVar(handlers.TagVarArgs{
		Name:     args[0],
		TagNames: args[1:],
		Profile:  profile,
	})
}
