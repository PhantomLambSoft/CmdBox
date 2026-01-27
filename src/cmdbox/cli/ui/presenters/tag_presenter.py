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
from cmdbox.models import Tag
from cmdbox.repositories.results import TagAttachResult, TagDetachResult


TAG_COLUMNS: dict[str, tuple[str, dict, Callable[[Tag], object]]] = {
    "name": (
        "Name",
        {"style": "entity.name", "no_wrap": True},
        lambda t: t.name,
    ),
    "description": (
        "Description",
        {"overflow": "fold"},
        lambda t: t.description,
    ),
    "date_created": (
        "Created",
        {"style": "entity.time", "no_wrap": True},
        lambda t: t.date_created,
    ),
    "last_updated": (
        "Updated",
        {"style": "entity.time", "no_wrap": True},
        lambda t: t.last_updated,
    ),
}


DEFAULT_FIELDS = ["name", "description"]


def render_tag_created(tag: Tag):
    rendered_tag = render_tag(tag)
    return section(
        title=f"Tag '{tag.name}' created",
        body=rendered_tag,
        border_style="status.success",
    )


def render_tag(tag: Tag):
    rows = []
    for value in TAG_COLUMNS.values():
        header, _, extractor = value
        rows.append((header, extractor(tag)))
    return kv_table(rows)


def render_tag_list(
    tags: Sequence[Tag], *, title: str = None, fields: list[str] | None = None
):
    active_fields = fields or DEFAULT_FIELDS
    active_fields = [f for f in active_fields if f in TAG_COLUMNS]

    columns = []
    extractors = []

    for field in active_fields:
        header, col_args, extractor = TAG_COLUMNS[field]
        columns.append(col(header, **col_args))
        extractors.append(extractor)

    rows = [tuple(extractor(t) for extractor in extractors) for t in tags]

    caption = f"{pluralize(len(tags), 'tag')} found"

    return table_panel(
        title=title or "Tags",
        columns=columns,
        rows=rows,
        caption=caption,
    )


def render_tag_updated(tag: Tag):
    rendered_tag = render_tag(tag)
    return section(
        title=f"Tag '{tag.name}' updated",
        body=rendered_tag,
        border_style="status.success",
    )


def render_tag_deleted(tag: Tag):
    rendered_tag = render_tag(tag)
    return section(
        title=f"Tag '{tag.name}' deleted",
        body=rendered_tag,
        border_style="status.success",
    )
