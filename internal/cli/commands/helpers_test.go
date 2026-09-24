package commands

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func newTestCmd() *cobra.Command {
	return &cobra.Command{Use: "test"}
}

func TestGetStringFlag(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().String("name", "default", "")

	assert.Equal(t, "default", getStringFlag(cmd, "name"))

	require := cmd.Flags().Set("name", "value")
	assert.NoError(t, require)
	assert.Equal(t, "value", getStringFlag(cmd, "name"))
}

func TestGetStringFlag_UnregisteredReturnsZeroValue(t *testing.T) {
	cmd := newTestCmd()
	assert.Equal(t, "", getStringFlag(cmd, "missing"))
}

func TestGetBoolFlag(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().Bool("force", false, "")

	assert.False(t, getBoolFlag(cmd, "force"))

	assert.NoError(t, cmd.Flags().Set("force", "true"))
	assert.True(t, getBoolFlag(cmd, "force"))
}

func TestGetBoolFlag_UnregisteredReturnsZeroValue(t *testing.T) {
	cmd := newTestCmd()
	assert.False(t, getBoolFlag(cmd, "missing"))
}

func TestGetProfile(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().String("profile", "", "")

	assert.Equal(t, "", getProfile(cmd))

	assert.NoError(t, cmd.Flags().Set("profile", "work"))
	assert.Equal(t, "work", getProfile(cmd))
}

func TestGetDescriptionFlag(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().String("description", "", "")

	assert.NoError(t, cmd.Flags().Set("description", "a desc"))
	assert.Equal(t, "a desc", getDescriptionFlag(cmd))
}

func TestGetInteractiveFlag(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().Bool("interactive", false, "")

	assert.False(t, getInteractiveFlag(cmd))

	assert.NoError(t, cmd.Flags().Set("interactive", "true"))
	assert.True(t, getInteractiveFlag(cmd))
}

func TestGetEditMode(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().Bool("edit", false, "")

	assert.False(t, getEditMode(cmd))

	assert.NoError(t, cmd.Flags().Set("edit", "true"))
	assert.True(t, getEditMode(cmd))
}

func TestGetEditFields(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().String("edit-fields", "", "")

	assert.Equal(t, "", getEditFields(cmd))

	assert.NoError(t, cmd.Flags().Set("edit-fields", "name,value"))
	assert.Equal(t, "name,value", getEditFields(cmd))
}

func TestGetTags(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().StringSlice("tag", nil, "")

	assert.Empty(t, getTags(cmd))

	assert.NoError(t, cmd.Flags().Set("tag", "a,b"))
	assert.Equal(t, []string{"a", "b"}, getTags(cmd))
}

// getTags must read the "tag" flag, not "tags" - every real command in this package
// registers the filter flag as singular "tag". This guards against the flag name
// regressing back to the plural mismatch.
func TestGetTags_DoesNotReadPluralFlagName(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().StringSlice("tags", []string{"should-be-ignored"}, "")

	assert.Empty(t, getTags(cmd))
}

func TestGetLimit(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().Int("limit", 0, "")

	assert.Equal(t, 0, getLimit(cmd))

	assert.NoError(t, cmd.Flags().Set("limit", "25"))
	assert.Equal(t, 25, getLimit(cmd))
}

func TestGetFields(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().StringSlice("field", nil, "")

	assert.Empty(t, getFields(cmd))

	assert.NoError(t, cmd.Flags().Set("field", "alias,template"))
	assert.Equal(t, []string{"alias", "template"}, getFields(cmd))
}

func TestGetOrder(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().String("order", "", "")

	assert.Equal(t, "", getOrder(cmd))

	assert.NoError(t, cmd.Flags().Set("order", "alias"))
	assert.Equal(t, "alias", getOrder(cmd))
}

func TestGetPage_NeitherFlagSet(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().Bool("page", false, "")
	cmd.Flags().Bool("no-page", false, "")

	page, err := getPage(cmd)
	assert.NoError(t, err)
	assert.Nil(t, page)
}

func TestGetPage_PageSet(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().Bool("page", false, "")
	cmd.Flags().Bool("no-page", false, "")
	assert.NoError(t, cmd.Flags().Set("page", "true"))

	page, err := getPage(cmd)
	assert.NoError(t, err)
	if assert.NotNil(t, page) {
		assert.True(t, *page)
	}
}

func TestGetPage_NoPageSet(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().Bool("page", false, "")
	cmd.Flags().Bool("no-page", false, "")
	assert.NoError(t, cmd.Flags().Set("no-page", "true"))

	page, err := getPage(cmd)
	assert.NoError(t, err)
	if assert.NotNil(t, page) {
		assert.False(t, *page)
	}
}

func TestGetPage_BothSetIsAnError(t *testing.T) {
	cmd := newTestCmd()
	cmd.Flags().Bool("page", false, "")
	cmd.Flags().Bool("no-page", false, "")
	assert.NoError(t, cmd.Flags().Set("page", "true"))
	assert.NoError(t, cmd.Flags().Set("no-page", "true"))

	page, err := getPage(cmd)
	assert.Error(t, err)
	assert.Nil(t, page)
}

// getPage panics if "page"/"no-page" aren't registered on the command at all - this is a
// deliberate defensive check for a programmer error (a command wired up without the
// flags), not a user-facing error path, so it's documented here rather than treated as
// something to make graceful.
func TestGetPage_PanicsWhenFlagsNotRegistered(t *testing.T) {
	cmd := newTestCmd()
	assert.Panics(t, func() {
		_, _ = getPage(cmd)
	})
}
