from cmdbox.cli.ui.theme import Theme
from rich.theme import Theme as RichTheme


def build_theme(settings) -> Theme:
    colors = settings.ui.colors
    cmd = settings.ui.command_colors
    var = settings.ui.variable_colors
    tag = settings.ui.tag_colors

    style_map = {
        "success": colors.success,
        "error": colors.error,
        "warning": colors.warning,
        "info": colors.info,
        "debug": colors.debug,
        "muted": colors.muted,
        "command.alias": cmd.alias,
        "command.template": cmd.template,
        "command.description": cmd.description,
        "command.date_created": cmd.date_created,
        "command.last_updated": cmd.last_updated,
        "command.used": cmd.used,
        "command.last_used": cmd.last_used,
        "variable.name": var.name,
        "variable.value": var.value,
        "variable.date_created": var.date_created,
        "variable.last_updated": var.last_updated,
        "tag.name": tag.name,
        "tag.description": tag.description,
        "tag.date_created": tag.date_created,
        "tag.last_updated": tag.last_updated,
    }

    return Theme(
        rich=RichTheme(style_map),
        success="success",
        error="error",
        warning="warning",
        info="info",
        debug="debug",
        muted="muted",
        command_alias="command.alias",
        command_template="command.template",
        command_description="command.description",
        command_date_created="command.date_created",
        command_last_updated="command.last_updated",
        command_used="command.used",
        command_last_used="command.last_used",
        variable_name="variable.name",
        variable_value="variable.value",
        variable_date_created="variable.date_created",
        variable_last_updated="variable.last_updated",
        tag_name="tag.name",
        tag_description="tag.description",
        tag_date_created="tag.date_created",
        tag_last_updated="tag.last_updated",
    )
