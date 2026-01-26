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
from cmdbox.repositories.results import TagAttachResult, TagDetachResult

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
        {"style": "entity.time"},
        lambda c: c.date_created,
    ),
    "last_updated": (
        "Updated",
        {"style": "entity.time"},
        lambda c: c.last_updated,
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
    return kv_table(rows)


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


def render_tag_attach_result(result: TagAttachResult):
    added_tag_block = tag_block(tags=result.added)
    existing_tag_block = tag_block(tags=result.existing, style="tag.pill_muted")
    renderable = []
    if len(result.added) > 0:
        added_section = section(
            title=f"{len(result.added)} tags added successfully",
            body=added_tag_block,
            border_style="status.success",
        )
        renderable.append(added_section)
    if len(result.existing) > 0:
        existing_section = section(
            title=f"{len(result.existing)} tags already exist",
            body=existing_tag_block,
            border_style="ui.dim",
        )
        renderable.append(existing_section)
    return stack(*renderable)


def render_tag_detach_result(result: TagDetachResult):
    removed_tag_block = tag_block(tags=result.removed)
    not_attached_tag_block = tag_block(tags=result.not_attached, style="tag.pill_muted")
    renderable = []
    if len(result.removed) > 0:
        removed_section = section(
            title=f"{len(result.removed)} tags removed successfully",
            body=removed_tag_block,
            border_style="status.success",
        )
        renderable.append(removed_section)
    if len(result.not_attached) > 0:
        not_attached_section = section(
            title=f"{len(result.not_attached)} tags not attached",
            body=not_attached_tag_block,
            border_style="ui.dim",
        )
        renderable.append(not_attached_section)
    return stack(*renderable)
