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
from cmdbox.repositories.results import TagAttachResult, TagDetachResult


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
