import json
from typing import Sequence, Callable

from cmdbox.cli.ui.primitives import (
    col,
    pluralize,
    table_panel,
    kv_table,
    section,
    tag_block,
)
from cmdbox.models import Command

COMMAND_COLUMNS: dict[str, tuple[str, dict, Callable[[Command], object]]] = {
    "alias": (
        "Alias",
        {"style": "entity.name", "no_wrap": True},
        lambda c: c.alias,
    ),
    "template": (
        "Template",
        {"style": "code.inline", "overflow": "fold"},
        lambda c: c.template,
    ),
    "description": (
        "Description",
        {"overflow": "fold"},
        lambda c: c.description,
    ),
    "cwd": (
        "Working Directory",
        {"overflow": "fold"},
        lambda c: c.cwd,
    ),
    "shell": (
        "Shell",
        {"no_wrap": True},
        lambda c: c.shell,
    ),
    "env": (
        "Environment",
        {"overflow": "fold"},
        lambda c: format_env(c.env),
    ),
    "timeout": (
        "Timeout",
        {"no_wrap": True},
        lambda c: f"{c.timeout}s" if c.timeout is not None else None,
    ),
    "used": (
        "Used",
        {"style": "entity.count", "no_wrap": True, "justify": "right"},
        lambda c: c.used,
    ),
    "last_used": (
        "Last Used",
        {"style": "entity.time", "no_wrap": True},
        lambda c: c.last_used,
    ),
    "date_created": (
        "Created",
        {"style": "entity.time", "no_wrap": True},
        lambda c: c.date_created,
    ),
    "last_updated": (
        "Updated",
        {"style": "entity.time", "no_wrap": True},
        lambda c: c.last_updated,
    ),
    "tags": (
        "Tags",
        {"overflow": "fold"},
        lambda c: tag_block([x.tag.name for x in c.tags]),
    ),
}

DEFAULT_FIELDS = ["alias", "template", "description"]


def render_command_created(command: Command):
    rendered_command = render_command(command)
    return section(
        title=f"Command '{command.alias}' created",
        body=rendered_command,
        border_style="status.success",
    )


def render_command(command: Command):
    # Conditional fields will only be shown if they are not None
    conditional_fields = ["cwd", "shell", "env", "timeout"]
    rows = []
    for key, value in COMMAND_COLUMNS.items():
        header, _, extractor = value
        extracted_value = extractor(command)
        if key in conditional_fields and extracted_value is None:
            continue
        rows.append((header, extracted_value))
    cmd_display = kv_table(rows)
    return cmd_display


def render_command_list(
    commands: Sequence[Command], *, title: str = None, fields: list[str] | None = None
):
    active_fields = fields or DEFAULT_FIELDS
    active_fields = [f for f in active_fields if f in COMMAND_COLUMNS]

    columns = []
    extractors = []

    for field in active_fields:
        header, col_args, extractor = COMMAND_COLUMNS[field]
        columns.append(col(header, **col_args))
        extractors.append(extractor)

    rows = [tuple(extractor(c) for extractor in extractors) for c in commands]

    caption = f"{pluralize(len(commands), 'command')} found"

    return table_panel(
        title=title or "Commands",
        columns=columns,
        rows=rows,
        caption=caption,
    )


def render_command_updated(command: Command):
    rendered_command = render_command(command)
    return section(
        title=f"Command '{command.alias}' updated",
        body=rendered_command,
        border_style="status.success",
    )


def render_command_deleted(command: Command):
    rendered_command = render_command(command)
    return section(
        title=f"Command '{command.alias}' deleted",
        body=rendered_command,
        border_style="status.success",
    )


def format_env(env: str | None) -> str | None:
    """
    Formats the provided environment string by parsing it as JSON and converting the
    key-value pairs into a newline-separated string. If the string is not valid JSON
    or parsing fails, the original string is returned.

    Args:
        env (str | None): A string representing the environment in JSON format, or
            None if no environment is provided.

    Returns:
        str | None: A formatted string with key-value pairs separated by newlines
            if the input was parsed successfully; otherwise, the original string or
            None if the input was None.
    """
    if not env:
        return None
    try:
        parsed = json.loads(env)
        return "\n".join(f"{k}={v}" for k, v in parsed.items())
    except (json.JSONDecodeError, AttributeError):
        return env
