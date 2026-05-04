from rich.text import Text

from cmdbox.cli.ui.primitives import (
    to_text,
    col,
    table_panel,
    pluralize,
    kv_table,
    section,
    status_line,
)
from cmdbox.models import CommandHistory


def _short_id(entry: CommandHistory) -> str:
    return entry.id[:6]


def _format_exit(exit_code: int | None) -> Text:
    if exit_code is None:
        return to_text("—", style="ui.muted")
    if exit_code == 0:
        return to_text("0", style="status.success")
    return to_text(str(exit_code), style="status.error")


HISTORY_LIST_COLUMNS = [
    col("#", style="ui.muted", justify="right", no_wrap=True),
    col("ID", style="entity.id", no_wrap=True),
    col("Alias", style="entity.name", no_wrap=True),
    col("Ran At", style="entity.time", no_wrap=True),
    col("Exit", justify="center", no_wrap=True),
]


def _list_rows(entries: list[CommandHistory]) -> list[tuple]:
    return [
        (
            str(i),
            _short_id(entry),
            entry.alias,
            to_text(entry.ran_at),
            _format_exit(entry.exit_code),
        )
        for i, entry in enumerate(entries, start=1)
    ]


def render_history_list(entries: list[CommandHistory]):
    return table_panel(
        title="History",
        columns=HISTORY_LIST_COLUMNS,
        rows=_list_rows(entries),
        caption=pluralize(len(entries), "entry", "entries"),
    )


def render_history_entry(entry: CommandHistory, variables: dict | None):
    rows = [
        ("ID", to_text(_short_id(entry), style="entity.id")),
        ("Alias", to_text(entry.alias, style="entity.name")),
        ("Ran At", to_text(entry.ran_at, style="entity.time")),
        ("Exit Code", _format_exit(entry.exit_code)),
        ("Template", to_text(entry.template, style="code.inline")),
        ("Resolved", to_text(entry.resolved, style="code.inline")),
    ]

    if variables:
        for k, v in variables.items():
            rows.append((f"  <{k}>", to_text(v, style="code.inline")))

    return section(
        title=f"History Entry - {_short_id(entry)}",
        body=kv_table(rows),
    )


def render_history_cleared(count: int, alias: str | None):
    scope = f" for '{alias}'" if alias else ""
    return status_line(
        f"Cleared {pluralize(count, 'entry', 'entries')}{scope}.",
        status="success",
    )
