from rich.theme import Theme


CMDBOX_THEME = Theme(
    {
        # Core text
        "ui.title": "bold",
        "ui.subtitle": "dim",
        "ui.muted": "dim",
        "ui.dim": "dim",
        "ui.border": "dim",
        "ui.panel_title": "bold",
        # Tables
        "ui.table_header": "bold",
        "ui.caption": "dim",
        # Key/value panels
        "ui.kv.key": "dim",
        "ui.kv.value": "",
        # Status
        "status.success": "bold green",
        "status.info": "bold cyan",
        "status.warning": "bold yellow",
        "status.error": "bold red",
        # Code semantics (not a font switch, just a consistent code look)
        "code": "cyan",
        "code.inline": "cyan",
        "code.block": "cyan",
        # Common entity accents (optional, keep minimal)
        "entity.name": "bold",
        "entity.id": "magenta",
        "entity.count": "bold",
        "entity.time": "dim",
        # Tags
        "tag.pill": "bold white on dark_green",
        "tag.pill_muted": "white on grey23",
        # Execution and previews
        "run.command": "cyan",
        "run.output": "",
        "run.trace.kind": "dim",
        "run.trace.key": "magenta",
        "run.trace.value": "cyan",
    }
)
