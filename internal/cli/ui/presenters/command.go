package presenters

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
)

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func formatTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

func formatEnv(env *string) string {
	if env == nil || *env == "" {
		return ""
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(*env), &m); err != nil {
		return *env
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	lines := make([]string, len(keys))
	for i, k := range keys {
		lines[i] = k + "=" + m[k]
	}
	return strings.Join(lines, "\n")
}

func tagNames(tags []models.Tag) []string {
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	sort.Strings(names)
	return names
}

var CommandColumns = []Column[models.Command]{
	{Key: "alias", Header: "Alias", Render: func(c models.Command, p ui.Palette) string {
		return p.EntityName.Render(c.Alias)
	}},
	{Key: "template", Header: "Template", Render: func(c models.Command, p ui.Palette) string {
		return p.CodeInline.Render(c.Template)
	}},
	{Key: "description", Header: "Description", Render: func(c models.Command, p ui.Palette) string {
		return deref(c.Description)
	}},
	{Key: "cwd", Header: "Working Directory", Optional: true, Render: func(c models.Command, p ui.Palette) string {
		return deref(c.Cwd)
	}},
	{Key: "shell", Header: "Shell", Optional: true, Render: func(c models.Command, p ui.Palette) string {
		return deref(c.Shell)
	}},
	{Key: "env", Header: "Environment", Optional: true, Render: func(c models.Command, p ui.Palette) string {
		return formatEnv(c.Env)
	}},
	{Key: "timeout", Header: "Timeout", Optional: true, Render: func(c models.Command, p ui.Palette) string {
		if c.Timeout == nil {
			return ""
		}
		return strconv.Itoa(*c.Timeout) + "s"
	}},
	{Key: "used", Header: "Used", Render: func(c models.Command, p ui.Palette) string {
		return p.EntityCount.Render(strconv.Itoa(c.Used))
	}},
	{Key: "last_used", Header: "Last Used", Render: func(c models.Command, p ui.Palette) string {
		return p.EntityTime.Render(formatTime(c.LastUsed))
	}},
	{Key: "date_created", Header: "Created", Render: func(c models.Command, p ui.Palette) string {
		return p.EntityTime.Render(formatTime(&c.CreatedAt))
	}},
	{Key: "last_updated", Header: "Updated", Render: func(c models.Command, p ui.Palette) string {
		return p.EntityTime.Render(formatTime(&c.UpdatedAt))
	}},
	{Key: "tags", Header: "Tags", Render: func(c models.Command, p ui.Palette) string {
		return ui.TagBlock(tagNames(c.Tags), p)
	}},
}

var DefaultCommandFields = []string{"alias", "template", "description"}

func RenderCommand(cmd models.Command, palette ui.Palette) string {
	return renderDetail(cmd, CommandColumns, palette)
}

func RenderCommandCreated(cmd models.Command, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Command '%s' created", cmd.Alias), RenderCommand(cmd, palette), palette)
}

func RenderCommandList(cmds []models.Command, title string, fields []string, palette ui.Palette) string {
	var cols []Column[models.Command]
	if len(fields) == 0 {
		cols = selectColumns(CommandColumns, DefaultCommandFields)
	} else {
		cols = selectColumns(CommandColumns, fields)
	}
	return RenderList(title, cmds, cols, palette)
}

func RenderCommandUpdated(cmd models.Command, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Command '%s' updated", cmd.Alias), RenderCommand(cmd, palette), palette)
}

func RenderCommandDeleted(cmd models.Command, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Command '%s' deleted", cmd.Alias), RenderCommand(cmd, palette), palette)
}

func RenderCommandMoved(cmd models.Command, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Command '%s' moved", cmd.Alias), RenderCommand(cmd, palette), palette)
}

func RenderCommandCopied(cmd models.Command, palette ui.Palette) string {
	return ui.SectionSuccess(fmt.Sprintf("Command '%s' copied", cmd.Alias), RenderCommand(cmd, palette), palette)
}
