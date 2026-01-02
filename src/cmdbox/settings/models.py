from dataclasses import dataclass


@dataclass(frozen=True)
class DefaultColors:
    red: str = "#a54242"
    green: str = "#8c9440"
    orange: str = "#de935f"
    blue: str = "#5f819d"
    purple: str = "#85678f"
    teal: str = "#5e8d87"
    light_grey: str = "#707880"
    dark_grey: str = "#373b41"
    pink: str = "#cc6666"
    yellow_green: str = "#b5bd68"
    yellow: str = "#f0c674"
    light_blue: str = "#81a2be"
    light_purple: str = "#b294bb"
    cyan: str = "#8abeb7"
    white: str = "#c5c8c6"


@dataclass(frozen=True)
class UiColors:
    success: str = "green"
    error: str = "red"
    warning: str = "yellow"
    info: str = "blue"
    debug: str = "magenta"
    muted: str = "bright_black"


@dataclass(frozen=True)
class UiCommandColors:
    alias: str = DefaultColors.red
    template: str = DefaultColors.green
    description: str = DefaultColors.orange
    date_created: str = DefaultColors.blue
    last_updated: str = DefaultColors.light_blue
    used: str = DefaultColors.purple
    last_used: str = DefaultColors.light_purple


@dataclass(frozen=True)
class UiVariableColors:
    name: str = DefaultColors.red
    value: str = DefaultColors.green
    date_created: str = DefaultColors.blue
    last_updated: str = DefaultColors.light_blue


@dataclass(frozen=True)
class UiTagColors:
    name: str = DefaultColors.red
    description: str = DefaultColors.orange
    date_created: str = DefaultColors.blue
    last_updated: str = DefaultColors.light_blue


@dataclass(frozen=True)
class RunPreviewColors:
    command: str = DefaultColors.green
    step_kind: str = DefaultColors.teal
    step_key: str = DefaultColors.pink
    step_expanded_to: str = DefaultColors.purple


@dataclass(frozen=True)
class UiSettings:
    use_color: bool = True
    colors: UiColors = UiColors()
    command_colors: UiCommandColors = UiCommandColors()
    variable_colors: UiVariableColors = UiVariableColors()
    tag_colors: UiTagColors = UiTagColors()
    run_preview_colors: RunPreviewColors = RunPreviewColors()


@dataclass(frozen=True)
class ExecutionSettings:
    default_shell: str = "auto"  # auto | bash | zsh | pwsh | cmd
    capture_output: bool = False


@dataclass(frozen=True)
class Settings:
    ui: UiSettings = UiSettings()
    execution_settings: ExecutionSettings = ExecutionSettings()
