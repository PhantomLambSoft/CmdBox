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
