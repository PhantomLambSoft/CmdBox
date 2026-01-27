from rich.text import Text

from cmdbox.cli.ui.primitives import (
    banner,
    section,
    tag_block,
    stack,
    spacer,
    code_block,
)
from cmdbox.repositories.results import TagAttachResult, TagDetachResult
from cmdbox.runtime.results import ExecutionResult


def render_execution_result(result: ExecutionResult, *, title: str = "Run"):
    ok = result.exit_code == 0
    status = "success" if ok else "error"

    # Wrap stdout border in error color if status is error
    stdout_border = "ui.border" if ok else "status.error"
    # If ok, wrap stderr border in warning color, otherwise error
    stderr_border = "status.error" if not ok else "status.warning"

    parts = [
        banner(
            title=title if ok else f"{title} failed",
            subtitle=f"Exit code: {result.exit_code}",
            status=status,
        ),
        spacer(1),
        code_block(
            result.command,
            title="Command",
            border_style="ui.border",
            style="run.command",
        ),
    ]

    if result.stdout and result.stdout.strip():
        parts += [
            section(
                "STDOUT",
                Text(result.stdout.rstrip("\n"), style="run.stdout"),
                border_style=stdout_border,
            ),
        ]

    if result.stderr and result.stderr.strip():
        parts += [
            section(
                "STDERR",
                Text(result.stderr.rstrip("\n"), style="run.stderr"),
                border_style=stderr_border,
            ),
        ]

    return stack(*parts)


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
