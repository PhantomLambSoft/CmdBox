import os
import sys
from pathlib import Path

import psutil


def detect_shell() -> str:
    """
    Detect the shell being used in the current operating environment.

    The function identifies the shell by examining the parent process of the current
    script and its name. On Unix systems, it also checks the `SHELL` environment
    variable as a fallback. For Windows, it supports shell detection for cmd,
    PowerShell, and Git Bash.

    Returns:
        str: The name of the detected shell. If the shell cannot be determined using
        process information, it falls back to environment-based detection.
    """
    parent = psutil.Process(os.getpid()).parent()
    name = ((parent.name() if parent else "") or "").lower()

    # Windows
    if name in {"pwsh.exe", "powershell.exe"}:
        return "powershell"
    if name == "cmd.exe":
        return "cmd"
    if name in {"bash.exe", "git-bash.exe", "msys2.exe"}:
        return "bash"

    # Unix
    if name in {"bash", "zsh", "fish"}:
        return name

    # Fallback
    p = os.environ.get("SHELL")
    if p:
        sh = Path(p).name.lower()
        if sh in {"bash", "zsh", "fish"}:
            return sh

    # Default to env detection if everything else fails
    return detect_shell_env()


def detect_shell_env() -> str:
    """
    Detects the shell being used on the operating system.

    This function determines the active shell based on environment variables,
    platform information, and commonly used shell identifiers. It identifies
    the shell for both Windows and Unix-based systems.

    Returns:
        str: A string representing the detected shell. Possible values include
        'bash', 'powershell', 'cmd', 'zsh', and 'fish'. Defaults to 'bash'
        if unable to determine the specific shell.
    """
    if sys.platform.startswith("win"):
        # Git bash / MSYS
        if (
            os.environ.get("MSYSTEM")
            or os.environ.get("MINGW_PREFIX")
            or os.environ.get("SHELL")
        ):
            return "bash"
        # Powershell
        if os.environ.get("PSModulePath") or os.environ.get(
            "POWERSHELL_DISTRIBUTION_CHANNEL"
        ):
            return "powershell"
        # Default to cmd
        return "cmd"

    sh = os.environ.get("SHELL")
    if sh:
        base = Path(sh).name.lower()
        if base in {"bash", "zsh", "fish"}:
            return base

    return "bash"
