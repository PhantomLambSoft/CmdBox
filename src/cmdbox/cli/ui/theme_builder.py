from rich.theme import Theme as RichTheme

from cmdbox.settings.models import Settings


def build_theme(settings: Settings) -> RichTheme:
    c = settings.ui.colors

    style_map = {
        "ui.title": c.title,
        "ui.subtitle": c.subtitle,
        "ui.muted": c.muted,
        "ui.dim": c.muted,
        "ui.border": c.border,
        "ui.panel_title": c.panel_title,
        # Tables
        "ui.table_header": c.table_header,
        "ui.caption": c.caption,
        # Key/value panels
        "ui.kv.key": c.kv_key,
        "ui.kv.value": c.kv_value,
        # Status
        "status.success": c.success,
        "status.info": c.info,
        "status.warning": c.warning,
        "status.error": c.error,
        # Code semantics (not a font switch, just a consistent code look)
        "code": c.code,
        "code.inline": c.code_inline,
        "code.block": c.code_block,
        # Common entity accents (optional, keep minimal)
        "entity.name": c.entity_name,
        "entity.id": c.entity_id,
        "entity.count": c.entity_count,
        "entity.time": c.entity_time,
        # Tags
        "tag.pill": c.tag_pill,
        "tag.pill_muted": c.tag_pill_muted,
        # Execution and previews
        "run.command": c.run_command,
        "run.output": c.run_output,
        "run.trace.kind": c.trace_kind,
        "run.trace.key": c.trace_key,
        "run.trace.value": c.trace_value,
    }

    return RichTheme(style_map)
