import os
import sys


def build_shell_command(command: str) -> list[str]:
    """
    Constructs a shell command suitable for execution on the host platform.

    This function builds a list of command-line arguments that represent a shell
    command ready to be executed via a subprocess. It accounts for platform-specific
    differences, such as the use of PowerShell or cmd.exe on Windows and standard
    shells (e.g., bash, zsh) on Unix-like systems. Additionally, it respects
    user-defined shell preferences specified in environment variables.

    Args:
        command (str): The command to be executed by the shell.

    Returns:
        list[str]: A list of strings representing the shell command and its arguments.
    """
    if sys.platform.startswith("win"):
        # Prefer PowerShell if available, otherwise cmd.exe
        comspec = os.environ.get("COMSPEC", "cmd.exe")

        # If the user has set CMDBOX_SHELL=powershell, honor it
        shell = os.environ.get("CMDBOX_SHELL", "").lower()
        if shell in ("pwsh", "powershell", "power_shell", "power-shell"):
            return ["pwsh", "-NoProfile", "-Command", command]

        return [comspec, "/c", command]

    # Linux/macOS
    sh = os.environ.get("SHELL", "/bin/sh")
    return [sh, "-lc", command]
