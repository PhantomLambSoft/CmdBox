from rich.console import RenderableType

from cmdbox.cli.ui.primitives import (
    bullet_list,
    divider,
    kv_table,
    pluralize,
    section,
    stack,
)
from cmdbox.services.export_service import ExportResult


def _labeled(name: str, is_transient: bool) -> str:
    return f"{name} (dependency)" if is_transient else name


def render_export_result(result: ExportResult) -> RenderableType:
    parts: list[RenderableType] = [kv_table([("Saved to", str(result.path))])]

    all_cmds = [(a, False) for a in result.commands] + [
        (a, True) for a in result.transient_commands
    ]
    all_vars = [(n, False) for n in result.variables] + [
        (n, True) for n in result.transient_variables
    ]

    if all_cmds:
        parts.append(divider(f"Commands from profile '{result.command_profile}'"))
        parts.append(bullet_list([_labeled(a, dep) for a, dep in all_cmds]))

    if all_vars:
        parts.append(divider(f"Variables from profile '{result.variable_profile}'"))
        parts.append(bullet_list([_labeled(n, dep) for n, dep in all_vars]))

    total_cmds = len(all_cmds)
    total_vars = len(all_vars)
    caption_parts = []
    if total_cmds:
        caption_parts.append(pluralize(total_cmds, "command"))
    if total_vars:
        caption_parts.append(pluralize(total_vars, "variable"))
    caption = (
        (", ".join(caption_parts) + " exported")
        if caption_parts
        else "Nothing exported"
    )

    return section(
        title="Export Complete",
        body=stack(*parts),
        caption=caption,
        border_style="status.success",
    )
