from typing import Sequence, Callable

from cmdbox.cli.ui.primitives import (
    col,
    pluralize,
    table_panel,
    kv_table,
    section,
    tag_block,
    stack,
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
    # TODO: add fields to this
    rows = []
    for value in COMMAND_COLUMNS.values():
        header, _, extractor = value
        rows.append((header, extractor(command)))
    cmd_display = kv_table(rows)

    tags = [link.tag.name for link in command.tags]
    if len(tags) > 0:
        tag_display = tag_block(tags=tags)
        return stack(cmd_display, tag_display)

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
