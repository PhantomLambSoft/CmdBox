from dataclasses import dataclass
from rich.theme import Theme as RichTheme


@dataclass(frozen=True)
class Theme:
    rich: RichTheme

    success: str
    error: str
    warning: str
    info: str
    debug: str
    muted: str

    command_alias: str
    command_template: str
    command_description: str
    command_date_created: str
    command_last_updated: str
    command_used: str
    command_last_used: str

    variable_name: str
    variable_value: str
    variable_date_created: str
    variable_last_updated: str

    tag_name: str
    tag_description: str
    tag_date_created: str
    tag_last_updated: str

    run_preview_command: str
    run_preview_step_kind: str
    run_preview_step_key: str
    run_preview_step_expanded_to: str
