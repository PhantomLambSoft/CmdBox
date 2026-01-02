from dataclasses import dataclass


@dataclass(frozen=True)
class DefaultColors:
    name: str = "#a54242"
    template: str = "#8c9440"
    description: str = "#de935f"
    date_created: str = "#5f819d"
    used: str = "#85678f"
    teal: str = "#5e8d87"
    light_grey: str = "#707880"
    dark_grey: str = "#373b41"
    pink: str = "#cc6666"
    yellow_green: str = "#b5bd68"
    yellow: str = "#f0c674"
    updated: str = "#81a2be"
    last_used: str = "#b294bb"
    light_blue: str = "#8abeb7"
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
    alias: str = DefaultColors.name
    template: str = DefaultColors.template
    description: str = DefaultColors.description
    date_created: str = DefaultColors.date_created
    last_updated: str = DefaultColors.updated
    used: str = DefaultColors.used
    last_used: str = DefaultColors.last_used


@dataclass(frozen=True)
class UiVariableColors:
    name: str = DefaultColors.name
    value: str = DefaultColors.template
    date_created: str = DefaultColors.date_created
    last_updated: str = DefaultColors.updated


@dataclass(frozen=True)
class UiTagColors:
    name: str = DefaultColors.name
    description: str = DefaultColors.description
    date_created: str = DefaultColors.date_created
    last_updated: str = DefaultColors.updated


@dataclass(frozen=True)
class UiSettings:
    use_color: bool = True
    colors: UiColors = UiColors()
    command_colors: UiCommandColors = UiCommandColors()
    variable_colors: UiVariableColors = UiVariableColors()
    tag_colors: UiTagColors = UiTagColors()


@dataclass(frozen=True)
class ExecutionSettings:
    default_shell: str = "auto"  # auto | bash | zsh | pwsh | cmd
    capture_output: bool = False


@dataclass(frozen=True)
class Settings:
    ui: UiSettings = UiSettings()
    execution_settings: ExecutionSettings = ExecutionSettings()
