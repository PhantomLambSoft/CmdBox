package presenters

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

var ProfileColumns = []Column[models.Profile]{
	{Key: "name", Header: "Name", Render: func(profile models.Profile, p ui.Palette) string {
		return p.EntityName.Render(profile.Name)
	}},
	{Key: "description", Header: "Description", Render: func(c models.Profile, p ui.Palette) string {
		return deref(c.Description)
	}},
	{Key: "last_used", Header: "Last Used", Render: func(c models.Profile, p ui.Palette) string {
		return p.EntityTime.Render(formatTime(c.LastUsed))
	}},
	{Key: "date_created", Header: "Created", Render: func(c models.Profile, p ui.Palette) string {
		return p.EntityTime.Render(formatTime(&c.CreatedAt))
	}},
	{Key: "last_updated", Header: "Updated", Render: func(c models.Profile, p ui.Palette) string {
		return p.EntityTime.Render(formatTime(&c.UpdatedAt))
	}},
}

var DefaultProfileFields = []string{"name", "description", "last_used"}

func RenderProfile(profile models.Profile, palette ui.Palette) string {
	return renderDetail(profile, ProfileColumns, palette)
}

func RenderProfileCreated(profile models.Profile, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Profile %s created", profile.Name), RenderProfile(profile, palette), palette)
}

func RenderProfileList(profiles []models.Profile, title string, fields []string, palette ui.Palette) string {
	var cols []Column[models.Profile]
	if len(fields) == 0 {
		cols = selectColumns(ProfileColumns, DefaultProfileFields)
	} else {
		cols = selectColumns(ProfileColumns, fields)
	}
	headers, rows := RenderTable(profiles, cols, palette)
	caption := ui.Pluralize(len(rows), "profile found", "profiles found")
	return ui.TablePanel(title, headers, rows, caption, palette)
}

func RenderProfileUpdated(profile models.Profile, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Profile '%s' updated", profile.Name), RenderProfile(profile, palette), palette)
}

func RenderProfileDeleted(profile models.Profile, deleteResult repository.ProfileDeleteResult, palette ui.Palette) string {
	var labels = []string{"Deleted commands", "Deleted variables"}
	var counts = []int{int(deleteResult.CommandsDeleted), int(deleteResult.VariablesDeleted)}
	body := ui.Stack(RenderProfile(profile, palette), ui.SummaryCounts(labels, counts, palette))
	return ui.SectionSuccess(
		fmt.Sprintf("Profile '%s' deleted", profile.Name),
		body,
		palette,
	)
}

func RenderProfileStatus(status services.ProfileStatus, palette ui.Palette) string {
	linked := "No"
	if status.Linked() {
		linked = "Yes"
	}
	var values = [][2]string{
		{"Command profile:", status.CommandProfile},
		{"Variable profile:", status.VariableProfile},
		{"Settings profile:", status.SettingsProfile},
		{"Linked:", linked},
	}
	return ui.KVPanel("Profile Status", values, palette)
}

func RenderProfileSwitched(currentState *models.ProfileState, switchedKinds []string, targetName string, palette ui.Palette) string {
	title := fmt.Sprintf("Switched %s profile to '%s'", ui.JoinWithAnd(switchedKinds), targetName)

	linked := "No"
	if currentState.Linked() {
		linked = "Yes"
	}
	var values = [][2]string{
		{"Command profile:", currentState.ActiveCommandProfile.Name},
		{"Variable profile:", currentState.ActiveVariableProfile.Name},
		{"Settings profile:", currentState.ActiveSettingsProfile.Name},
		{"Linked:", linked},
	}
	return ui.SectionSuccess(title, ui.KVTable(values, palette), palette)
}
