from pathlib import Path
import os
import sys


VENDOR = "SomeGuySoftware"
APP_NAME = "CmdBox"


def get_app_data_dir() -> Path:
    """
    Gets the application data directory for the current platform.

    This function determines the appropriate location where application data
    should be stored for the current operating system. The directory varies
    based on the operating system, ensuring compatibility and proper organization
    of application data.

    Returns:
        Path: A `Path` object representing the full path to the application
        data directory for the current platform.
    """
    if sys.platform == "win32":
        base = Path(os.environ["APPDATA"])
    elif sys.platform == "darwin":
        base = Path.home() / "Library" / "Application Support"
    else:
        base = Path.home() / ".local" / "share"
    return base / VENDOR / APP_NAME


APP_DATA_DIR = get_app_data_dir()
APP_DATA_DIR.mkdir(parents=True, exist_ok=True)


def get_log_dir() -> Path:
    """
    Retrieves the directory path designated for storing log files.

    This function ensures the existence of the 'logs' directory within the application
    data directory. If the directory does not exist, it gets created with necessary
    parent directories.

    Returns:
        Path: A `Path` object representing the log directory.
    """
    log_dir = APP_DATA_DIR / "logs"
    log_dir.mkdir(parents=True, exist_ok=True)
    return log_dir


def get_log_file_path() -> Path:
    return get_log_dir() / "cmdbox.log"
