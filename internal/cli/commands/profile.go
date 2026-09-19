package commands

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/handlers"
	"github.com/spf13/cobra"
)

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage profiles",
	Args:  cobra.MaximumNArgs(1),
}

var profileAddCmd = &cobra.Command{
	Use:   "add <name> [description]",
	Short: "Add a new profile",
	Args:  cobra.RangeArgs(1, 2),
	RunE:  runProfileAdd,
}

var profileGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a profile",
	Args:  cobra.ExactArgs(2),
	RunE:  runProfileGet,
}

var profileUpdateCmd = &cobra.Command{
	Use:     "update <name>",
	Aliases: []string{"edit"},
	Short:   "Update a profile",
	Args:    cobra.ExactArgs(1),
	RunE:    runProfileUpdate,
}

var profileListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List profiles",
	RunE:    runProfileList,
}

var profileDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm", "del", "remove"},
	Short:   "Delete a profile",
	Args:    cobra.ExactArgs(1),
	RunE:    runProfileDelete,
}

var profileStatusCmd = &cobra.Command{
	Use:     "status",
	Aliases: []string{"where"},
	Short:   "Show the currently active profile for command, variable, and settings",
	Args:    cobra.NoArgs,
	RunE:    runProfileStatus,
}

var profileSwitchCmd = &cobra.Command{
	Use:     "switch <name>",
	Aliases: []string{"set"},
	Short:   "Switch the active profile",
	Args:    cobra.ExactArgs(1),
	RunE:    runProfileSwitch,
}

var profileSwitchCmdCmd = &cobra.Command{
	Use:   "cmd <name>",
	Short: "Switch the active command profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileSwitchCmd,
}

var profileSwitchVarCmd = &cobra.Command{
	Use:   "var <name>",
	Short: "Switch the active variable profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileSwitchVar,
}

var profileSwitchSettingsCmd = &cobra.Command{
	Use:   "settings <name>",
	Short: "Switch the active settings profile",
	Args:  cobra.ExactArgs(1),
	RunE:  runProfileSwitchSettings,
}

func init() {
	profileAddCmd.Flags().BoolP("interactive", "i", false, "Prompt for every field")

	profileUpdateCmd.Flags().String("description", "", "New description")
	profileUpdateCmd.Flags().String("name", "", "New name (rename)")
	profileUpdateCmd.Flags().String("set", "", "Comma-separated key=value pairs")
	profileUpdateCmd.Flags().BoolP("edit", "e", false, "Interactively edit current fields")
	profileUpdateCmd.Flags().String("edit-fields", "", "Restrict --edit to these fields")

	profileListCmd.Flags().String("order", "", "Field to order by")
	profileListCmd.Flags().StringSlice("field", nil, "Fields to display (repeatable)")
	profileListCmd.Flags().Int("limit", 0, "Max results")
	profileListCmd.Flags().Bool("page", false, "Force paging on")
	profileListCmd.Flags().Bool("no-page", false, "Force paging off")
	profileListCmd.MarkFlagsMutuallyExclusive("page", "no-page")

	profileDeleteCmd.Flags().Bool("force", false, "Delete the profile even if it still has commands or variables assigned")

	profileCmd.Flags().Bool("cmd", false, "Switch only the command profile")
	profileCmd.Flags().Bool("var", false, "Switch only the variable profile")
	profileCmd.Flags().Bool("settings", false, "Switch only the settings profile")

	profileSwitchCmd.Flags().Bool("cmd", false, "Switch only the command profile")
	profileSwitchCmd.Flags().Bool("var", false, "Switch only the variable profile")
	profileSwitchCmd.Flags().Bool("settings", false, "Switch only the settings profile")

	profileCmd.AddCommand(
		profileAddCmd,
		profileGetCmd,
		profileUpdateCmd,
		profileListCmd,
		profileDeleteCmd,
		profileStatusCmd,
		profileSwitchCmd,
		profileSwitchCmdCmd,
		profileSwitchVarCmd,
		profileSwitchSettingsCmd,
	)
	rootCmd.AddCommand(profileCmd)
}

func runProfileBareShorthand(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	return handlers.RunSwitchProfile(switchArgsFromFlags(cmd, args[0]))
}

func runProfileAdd(cmd *cobra.Command, args []string) error {
	raw := handlers.AddProfileArgs{
		Interactive: getInteractiveFlag(cmd),
	}
	if len(args) > 1 {
		raw.Description = args[1]
		raw.HasDescription = true
	}
	return handlers.RunAddProfile(raw)
}

func runProfileGet(cmd *cobra.Command, args []string) error {
	return handlers.RunGetProfile(args[0])
}

func runProfileUpdate(cmd *cobra.Command, args []string) error {
	editMode := getEditMode(cmd)
	editFields := getEditFields(cmd)

	if editFields != "" && !editMode {
		return fmt.Errorf("--edit-fields can only be used with --edit")
	}

	raw := handlers.UpdateProfileArgs{
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
	return handlers.RunUpdateProfile(raw)
}

func runProfileList(cmd *cobra.Command, args []string) error {
	page, err := getPage(cmd)
	if err != nil {
		return err
	}
	raw := handlers.ListProfileArgs{
		Order:  getOrder(cmd),
		Fields: getFields(cmd),
		Page:   page,
	}
	if cmd.Flags().Changed("limit") {
		raw.Limit = getLimit(cmd)
		raw.HasLimit = true
	}
	return handlers.RunListProfile(raw)
}

func runProfileDelete(cmd *cobra.Command, args []string) error {
	return handlers.RunDeleteProfile(handlers.DeleteProfileArgs{
		Name:  args[0],
		Force: getBoolFlag(cmd, "force"),
	})
}

func runProfileStatus(cmd *cobra.Command, args []string) error {
	return handlers.RunProfileStatus()
}

func runProfileSwitch(cmd *cobra.Command, args []string) error {
	return handlers.RunSwitchProfile(switchArgsFromFlags(cmd, args[0]))
}

func runProfileSwitchCmd(cmd *cobra.Command, args []string) error {
	return handlers.RunSwitchProfile(handlers.SwitchProfileArgs{
		Name: args[0],
		Cmd:  true,
	})
}

func runProfileSwitchVar(cmd *cobra.Command, args []string) error {
	return handlers.RunSwitchProfile(handlers.SwitchProfileArgs{
		Name: args[0],
		Var:  true,
	})
}

func runProfileSwitchSettings(cmd *cobra.Command, args []string) error {
	return handlers.RunSwitchProfile(handlers.SwitchProfileArgs{
		Name:     args[0],
		Settings: true,
	})
}

func switchArgsFromFlags(cmd *cobra.Command, name string) handlers.SwitchProfileArgs {
	return handlers.SwitchProfileArgs{
		Name:     name,
		Cmd:      getBoolFlag(cmd, "cmd"),
		Var:      getBoolFlag(cmd, "var"),
		Settings: getBoolFlag(cmd, "settings"),
	}
}
