from dataclasses import dataclass


@dataclass(frozen=True)
class UIStyle:
    # Core
    title: str = "bold"
    subtitle: str = "dim"
    muted: str = "dim"
    border: str = "dim"
    panel_title: str = "bold"

    # Tables
    table_header: str = "bold"
    caption: str = "dim"

    # Key/value panels
    kv_key: str = "dim"
    kv_value: str = ""

    # Status
    success: str = "green"
    info: str = "cyan"
    warning: str = "yellow"
    error: str = "red"
    debug: str = "dim blue"

    # Code
    code: str = "cyan"
    code_inline: str = "cyan"
    code_block: str = "dim cyan"

    # Entity accents
    entity_name: str = "bold"
    entity_id: str = "magenta"
    entity_count: str = "bold"
    entity_time: str = "dim"

    # Tags
    tag_pill: str = "bold white on dark_green"
    tag_pill_muted: str = "white on grey23"

    # Execution and previews
    run_command: str = "cyan"
    run_stdout: str = ""
    run_stderr: str = "bold red"
    trace_kind: str = "dim"
    trace_key: str = "magenta"
    trace_value: str = "cyan"


@dataclass(frozen=True)
class DefaultFields:
    command_output = ["alias", "template", "description"]
    command_search = ["alias", "template", "description"]
    variable_output = ["name", "value"]
    variable_search = ["name", "value"]
    tag_output = ["name", "description"]
    tag_search = ["name", "description"]


@dataclass(frozen=True)
class UiSettings:
    use_color: bool = True
    colors: UIStyle = UIStyle()


@dataclass(frozen=True)
class ExecutionSettings:
    default_shell: str = "auto"  # auto | bash | zsh | pwsh | cmd
    capture_output: bool = False


@dataclass(frozen=True)
class Settings:
    ui: UiSettings = UiSettings()
    execution_settings: ExecutionSettings = ExecutionSettings()
    default_fields: DefaultFields = DefaultFields()
