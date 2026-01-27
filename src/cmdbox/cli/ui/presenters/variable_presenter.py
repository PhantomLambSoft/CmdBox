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
from cmdbox.models import Variable


VARIABLE_COLUMNS: dict[str, tuple[str, dict, Callable[[Variable], object]]] = {
    "name": (
        "Name",
        {"style": "entity.name", "no_wrap": True},
        lambda v: v.name,
    ),
    "value": (
        "Value",
        {"style": "code.inline", "overflow": "fold"},
        lambda v: v.value,
    ),
    "date_created": (
        "Created",
        {"style": "entity.time", "no_wrap": True},
        lambda v: v.date_created,
    ),
    "last_updated": (
        "Updated",
        {"style": "entity.time", "no_wrap": True},
        lambda v: v.last_updated,
    ),
    "tags": (
        "Tags",
        {"overflow": "fold"},
        lambda v: tag_block([x.tag.name for x in v.tags]),
    ),
}

DEFAULT_FIELDS = ["name", "value"]


def render_variable_created(variable: Variable):
    pass


def render_variable(variable: Variable):
    rows = []
    for value in VARIABLE_COLUMNS.values():
        header, _, extractor = value
        rows.append((header, extractor(variable)))
    var_display = kv_table(rows)

    tags = [link.tag.name for link in variable.tags]
    if len(tags) > 0:
        tag_display = tag_block(tags=tags)
        return stack(var_display, tag_display)

    return var_display


def render_variable_list(
    variables: Sequence[Variable], *, title: str = None, fields: list[str] | None = None
):
    active_fields = fields or DEFAULT_FIELDS
    active_fields = [f for f in active_fields if f in VARIABLE_COLUMNS]

    columns = []
    extractors = []

    for field in active_fields:
        header, col_args, extractor = VARIABLE_COLUMNS[field]
        columns.append(col(header, **col_args))
        extractors.append(extractor)

    rows = [tuple(extractor(v) for extractor in extractors) for v in variables]

    caption = f"{pluralize(len(variables), 'variable')} found"

    return table_panel(
        title=title or "Variables",
        caption=caption,
        columns=columns,
        rows=rows,
    )


def render_variable_updated(variable: Variable):
    rendered_variable = render_variable(variable)
    return section(
        title=f"Variable '{variable.name}' updated",
        body=rendered_variable,
        border_style="status.success",
    )


def render_variable_deleted(variable: Variable):
    rendered_variable = render_variable(variable)
    return section(
        title=f"Variable '{variable.name}' deleted",
        body=rendered_variable,
        border_style="status.success",
    )
