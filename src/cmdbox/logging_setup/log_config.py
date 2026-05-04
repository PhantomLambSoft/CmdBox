import logging
from dataclasses import dataclass
from pathlib import Path

from cmdbox.core.paths import get_log_file_path

LOGGER_NAME = "cmdbox"


@dataclass(frozen=True)
class LogConfig:
    console_level: int
    file_enabled: bool
    file_level: int
    file_path: Path
    max_bytes: int
    backups: int


def _level(level_str: str) -> int:
    s = (level_str or "").upper().strip()
    return getattr(logging, s, logging.INFO)


def build_log_config(
    settings, *, verbose: bool, debug: bool, file_logs: bool | None
) -> LogConfig:
    """
    Builds and returns a LogConfig instance based on the provided settings and flags.

    This function determines the appropriate logging configuration for the application
    by evaluating the verbosity, debug settings, and file-based logging preferences. It
    retrieves logging levels and file configurations from the provided settings.

    Args:
        settings: Application-specific configuration object containing logging
            settings such as file size limits and backup count.
        verbose: Flag indicating whether verbose logging is enabled.
        debug: Flag indicating whether debug-level logging is enabled.
        file_logs: Flag indicating whether file-based logging is enabled. If None,
            it defaults to the settings specified in the application configuration.

    Returns:
        LogConfig: A fully constructed LogConfig instance containing the logging
        configuration details, such as console logging level, file logging level,
        file path, maximum file size, and backup count.
    """
    console_level = get_console_level(settings, verbose=verbose, debug=debug)

    file_enabled = get_file_enabled(settings, file_logs=file_logs)
    file_level = get_file_level(settings, verbose=verbose, debug=debug)

    file_path = get_log_file_path()

    return LogConfig(
        console_level=console_level,
        file_enabled=file_enabled,
        file_level=file_level,
        file_path=file_path,
        max_bytes=settings.logging.file.max_bytes,
        backups=settings.logging.file.backups,
    )


def get_console_level(settings, *, verbose: bool, debug: bool) -> int:
    """
    Determines and returns the appropriate logging level for console output.

    Args:
        settings: Application settings object containing logging configuration.
        verbose: Flag indicating whether verbose mode is enabled.
        debug: Flag indicating whether debug mode is enabled.

    Returns:
        int: Numerical value representing the logging level.
    """
    if debug:
        return logging.DEBUG
    elif verbose:
        return logging.INFO
    else:
        return _level(settings.logging.console_level)


def get_file_enabled(settings, *, file_logs: bool | None) -> bool:
    """
    Returns whether file logging is enabled based on the provided settings and optional override.

    Args:
        settings: Application configuration containing logging settings.
        file_logs (bool | None): Optional override for enabling or disabling file logging.

    Returns:
        bool: True if file logging is enabled; otherwise, False.
    """
    if file_logs is not None:
        return file_logs
    return bool(settings.logging.file.enabled)


def get_file_level(settings, *, verbose: bool, debug: bool) -> int:
    """
    Determines the appropriate logging level for file-based logging based on the provided settings
    and debug/verbose flags.

    Args:
        settings: Configuration settings that include logging level information.
        verbose: Enables verbose logging if set to True.
        debug: Enables debug-level logging if set to True.

    Returns:
        int: The resolved logging level for file-based logging.
    """
    if debug:
        return logging.DEBUG
    elif verbose:
        return logging.INFO
    else:
        return _level(settings.logging.file.level)


def get_logger() -> logging.Logger:
    return logging.getLogger(LOGGER_NAME)
