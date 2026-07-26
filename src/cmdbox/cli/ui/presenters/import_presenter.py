from rich.console import Group, RenderableType
from rich.text import Text

from cmdbox.cli.ui.primitives import (
    divider,
    kv_table,
    pluralize,
    section,
    stack,
)
from cmdbox.services.import_service import ImportResult

ACTION_ICONS: dict[str, tuple[str, str]] = {
    "create": ("+", "status.success"),
    "skip": ("~", "ui.muted"),
    "overwrite": ("↑", "status.info"),
}

ACTION_LABELS: dict[str, str] = {
    "create": "create",
    "skip": "skip (already exists)",
    "overwrite": "overwrite",
}


def _action_line(name: str, action: str) -> Text:
    icon, style = ACTION_ICONS[action]
    label = ACTION_LABELS[action]
    t = Text()
    t.append(f" {icon} ", style=style)
    t.append(f"{name:<30}", style="entity.name")
    t.append(label, style=style)
    return t


def _action_group(
    created: list[str],
    skipped: list[str],
    overwritten: list[str],
) -> list[Text]:
    lines: list[Text] = []
    for name in created:
        lines.append(_action_line(name, "create"))
    for name in skipped:
        lines.append(_action_line(name, "skip"))
    for name in overwritten:
        lines.append(_action_line(name, "overwrite"))
    return lines


def render_import_preview(result: ImportResult, *, source: str) -> RenderableType:
    parts: list[RenderableType] = [kv_table([("Source", source)])]

    cmd_lines = _action_group(
        result.commands_created,
        result.commands_skipped,
        result.commands_overwritten,
    )
    var_lines = _action_group(
        result.variables_created,
        result.variables_skipped,
        result.variables_overwritten,
    )

    if cmd_lines:
        parts.append(divider("Commands"))
        parts.append(Group(*cmd_lines))

    if var_lines:
        parts.append(divider("Variables"))
        parts.append(Group(*var_lines))

    total_cmds = len(cmd_lines)
    total_vars = len(var_lines)
    caption_parts = []
    if total_cmds:
        caption_parts.append(pluralize(total_cmds, "command"))
    if total_vars:
        caption_parts.append(pluralize(total_vars, "variable"))
    caption = (
        (", ".join(caption_parts) + " in file")
        if caption_parts
        else "Nothing to import"
    )

    return section(
        title="Import Preview",
        body=stack(*parts),
        caption=caption,
        border_style="status.info",
    )


def render_import_result(result: ImportResult) -> RenderableType:
    rows: list[tuple[str, str]] = []

    cmd_parts = []
    if result.commands_created:
        cmd_parts.append(f"{len(result.commands_created)} created")
    if result.commands_skipped:
        cmd_parts.append(f"{len(result.commands_skipped)} skipped")
    if result.commands_overwritten:
        cmd_parts.append(f"{len(result.commands_overwritten)} overwritten")
    if cmd_parts:
        rows.append(("Commands", ", ".join(cmd_parts)))

    var_parts = []
    if result.variables_created:
        var_parts.append(f"{len(result.variables_created)} created")
    if result.variables_skipped:
        var_parts.append(f"{len(result.variables_skipped)} skipped")
    if result.variables_overwritten:
        var_parts.append(f"{len(result.variables_overwritten)} overwritten")
    if var_parts:
        rows.append(("Variables", ", ".join(var_parts)))

    if not rows:
        rows.append(("", "Nothing was imported."))

    return section(
        title="Import Complete",
        body=kv_table(rows),
        border_style="status.success",
    )
