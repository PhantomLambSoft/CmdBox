package commands

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:     "tag",
	Aliases: []string{"tags"},
	Short:   "Manage tags",
}

var tagAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a tag",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runTagAdd,
}

var tagGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a tag",
	Args:  cobra.ExactArgs(1),
	RunE:  runTagGet,
}

var tagUpdateCmd = &cobra.Command{
	Use:     "update <name>",
	Aliases: []string{"edit"},
	Short:   "Update a tag",
	Args:    cobra.ExactArgs(1),
	RunE:    runTagUpdate,
}

var tagListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List tags",
	RunE:    runTagList,
}

var tagSearchCmd = &cobra.Command{
	Use:     "search <term>",
	Aliases: []string{"find"},
	Short:   "Search tags",
	Args:    cobra.ExactArgs(1),
	RunE:    runTagSearch,
}

var tagDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm", "del", "remove"},
	Short:   "Delete a tag",
	Args:    cobra.ExactArgs(1),
	RunE:    runTagDelete,
}

func init() {
	tagAddCmd.Flags().BoolP("interactive", "i", false, "Prompt for every field")

	tagUpdateCmd.Flags().String("description", "", "New description")
	tagUpdateCmd.Flags().String("name", "", "New name (rename)")
	tagUpdateCmd.Flags().String("set", "", "Comma-separated key=value pairs")
	tagUpdateCmd.Flags().BoolP("edit", "e", false, "Interactively edit current fields")
	tagUpdateCmd.Flags().String("edit-fields", "", "Restrict --edit to these fields")

	tagListCmd.Flags().String("order", "", "Field to order by")
	tagListCmd.Flags().StringSlice("field", nil, "Fields to display (repeatable)")
	tagListCmd.Flags().Int("limit", 0, "Max results")
	tagListCmd.Flags().Bool("page", false, "Force paging on")
	tagListCmd.Flags().Bool("no-page", false, "Force paging off")
	tagListCmd.MarkFlagsMutuallyExclusive("page", "no-page")

	tagSearchCmd.Flags().StringSlice("in", nil, "Fields to search in (repeatable)")
	tagSearchCmd.Flags().StringSlice("field", nil, "Fields to display (repeatable)")
	tagSearchCmd.Flags().Int("limit", 0, "Max results")
	tagSearchCmd.Flags().Bool("page", false, "Force paging on")
	tagSearchCmd.Flags().Bool("no-page", false, "Force paging off")
	tagSearchCmd.MarkFlagsMutuallyExclusive("page", "no-page")

	tagCmd.AddCommand(
		tagAddCmd,
		tagGetCmd,
		tagUpdateCmd,
		tagListCmd,
		tagSearchCmd,
		tagDeleteCmd,
	)
	rootCmd.AddCommand(tagCmd)
}

func runTagAdd(cmd *cobra.Command, args []string) error {
	raw := handlers.AddTagArgs{
		Name:        args[0],
		Interactive: getInteractiveFlag(cmd),
	}
	if len(args) > 1 {
		raw.Description = args[1]
		raw.HasDescription = true
	}
	return handlers.RunAddTag(raw)
}

func runTagGet(cmd *cobra.Command, args []string) error {
	return handlers.RunGetTag(args[0])
}

func runTagUpdate(cmd *cobra.Command, args []string) error {
	editMode := getEditMode(cmd)
	editFields := getEditFields(cmd)

	if editFields != "" && !editMode {
		return fmt.Errorf("--edit-fields can only be used with --edit")
	}

	raw := handlers.UpdateTagArgs{
		CurrentName: args[0],
		NewName:     getStringFlag(cmd, "name"),
		SetPairs:    getStringFlag(cmd, "set"),
		EditMode:    editMode,
		EditFields:  editFields,
	}

	if cmd.Flags().Changed("description") {
		raw.Description = getStringFlag(cmd, "description")
		raw.HasDescription = true
	}

	return handlers.RunUpdateTag(raw)
}

func runTagList(cmd *cobra.Command, args []string) error {
	page, err := getPage(cmd)
	if err != nil {
		return err
	}
	raw := handlers.ListTagArgs{
		Order:   getOrder(cmd),
		Fields:  getFields(cmd),
		Page:    page,
		Profile: getProfile(cmd),
	}

	if cmd.Flags().Changed("limit") {
		raw.Limit = getLimit(cmd)
		raw.HasLimit = true
	}

	return handlers.RunListTag(raw)
}

func runTagSearch(cmd *cobra.Command, args []string) error {
	searchIn, _ := cmd.Flags().GetStringSlice("in")
	page, err := getPage(cmd)
	if err != nil {
		return err
	}
	raw := handlers.SearchTagArgs{
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

	return handlers.RunSearchTag(raw)
}

func runTagDelete(cmd *cobra.Command, args []string) error {
	return handlers.RunDeleteTag(handlers.DeleteTagArgs{Name: args[0]})
}
