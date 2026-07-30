from dataclasses import dataclass, field

from cmdbox.common.parsing import parse_byte_size


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
    command_output: list[str] = field(
        default_factory=lambda: ["alias", "template", "description"]
    )
    command_search: list[str] = field(
        default_factory=lambda: ["alias", "template", "description"]
    )
    command_list_limit: int = 25
    command_default_order: str = "alias"
    variable_output: list[str] = field(default_factory=lambda: ["name", "value"])
    variable_search: list[str] = field(default_factory=lambda: ["name", "value"])
    variable_list_limit: int = 25
    variable_default_order: str = "name"
    tag_output: list[str] = field(default_factory=lambda: ["name", "description"])
    tag_search: list[str] = field(default_factory=lambda: ["name", "description"])
    tag_list_limit: int = 25
    tag_default_order: str = "name"
    profile_output: list[str] = field(
        default_factory=lambda: ["name", "description", "date_created"]
    )
    profile_search: list[str] = field(default_factory=lambda: ["name", "description"])
    profile_list_limit: int = 20
    profile_default_order: str = "name"


@dataclass(frozen=True)
class FieldAliases:
    alias_mapping: dict[str, list[str]] = field(
        default_factory=lambda: {
            "alias": ["a", "al"],
            "template": ["t", "temp"],
            "description": ["d", "desc"],
            "date_created": ["dc", "created"],
            "last_updated": ["lu", "updated"],
            "used": ["u"],
            "last_used": ["lu"],
        }
    )

    @property
    def alias_map(self) -> dict[str, str]:
        out: dict[str, str] = {}
        for field_name, aliases in self.alias_mapping.items():
            for alias in aliases:
                out[alias.lower()] = field_name
        return out


@dataclass(frozen=True)
class UiSettings:
    use_color: bool = True
    colors: UIStyle = UIStyle()
    pager_mode: str = "auto"  # auto | always | never
    pager_min_rows: int = 25
    pager_page_step: int = 15
    pager_line_step: int = 1


@dataclass(frozen=True)
class ExecutionSettings:
    default_shell: str = "auto"  # auto | bash | zsh | pwsh | cmd
    capture_output: bool = False
    default_verbose: bool = False


@dataclass(frozen=True)
class LoggingFileSettings:
    enabled: bool = False
    level: str = "INFO"
    max_bytes: int | str = 1_000_000
    backups: int = 3

    def __post_init__(self):
        object.__setattr__(self, "max_bytes", parse_byte_size(self.max_bytes))


@dataclass(frozen=True)
class LoggingSettings:
    console_level: str = "WARNING"
    file: LoggingFileSettings = LoggingFileSettings()


@dataclass(frozen=True)
class HistorySettings:
    enabled: bool = True
    limit_per_command: int = 100  # 0 = unlimited


@dataclass(frozen=True)
class Settings:
    ui: UiSettings = UiSettings()
    execution_settings: ExecutionSettings = ExecutionSettings()
    default_fields: DefaultFields = DefaultFields()
    field_aliases: FieldAliases = FieldAliases()
    logging: LoggingSettings = LoggingSettings()
    history: HistorySettings = HistorySettings()
