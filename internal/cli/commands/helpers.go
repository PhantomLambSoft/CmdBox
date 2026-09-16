package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func getStringFlag(cmd *cobra.Command, flagName string) string {
	flag, _ := cmd.Flags().GetString(flagName)
	return flag
}

func getProfile(cmd *cobra.Command) string {
	profile, _ := cmd.Flags().GetString("profile")
	return profile
}

func getDescriptionFlag(cmd *cobra.Command) string {
	description, _ := cmd.Flags().GetString("description")
	return description
}

func getInteractiveFlag(cmd *cobra.Command) bool {
	interactive, _ := cmd.Flags().GetBool("interactive")
	return interactive
}

func getEditMode(cmd *cobra.Command) bool {
	editMode, _ := cmd.Flags().GetBool("edit")
	return editMode
}

func getEditFields(cmd *cobra.Command) string {
	editFields, _ := cmd.Flags().GetString("edit-fields")
	return editFields
}

func getTags(cmd *cobra.Command) []string {
	tags, _ := cmd.Flags().GetStringSlice("tags")
	return tags
}

func getLimit(cmd *cobra.Command) int {
	limit, _ := cmd.Flags().GetInt("limit")
	return limit
}

func getFields(cmd *cobra.Command) []string {
	fields, _ := cmd.Flags().GetStringSlice("field")
	return fields
}

func getOrder(cmd *cobra.Command) string {
	order, _ := cmd.Flags().GetString("order")
	return order
}

func getPage(cmd *cobra.Command) (*bool, error) {
	page, err := cmd.Flags().GetBool("page")
	if err != nil {
		panic(fmt.Sprintf("flag lookup failed: %v", err))
	}
	noPage, err := cmd.Flags().GetBool("no-page")
	if err != nil {
		panic(fmt.Sprintf("flag lookup failed: %v", err))
	}

	if page && noPage {
		return nil, fmt.Errorf("--page and --no-page cannot be used together")
	}
	if page {
		t := true
		return &t, nil
	}
	if noPage {
		f := false
		return &f, nil
	}
	return nil, nil
}
