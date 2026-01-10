import os
import sys
from shutil import which


def build_shell_command(command: str, preferred_shell: str | None = None) -> list[str]:
    """
    Builds a shell command list based on the provided command and preferred shell, with fallback mechanisms for
    platform compatibility and environment-specific configurations.

    This function determines the appropriate shell and constructs a command list for execution. It handles both
    Windows and Unix-like environments, prioritizing the specified preferred shell or defaulting to environment
    variables and standard fallback options. If no suitable shell is found, it raises a RuntimeError.

    Args:
        command (str): The shell command to execute.
        preferred_shell (str | None): The preferred shell to use for executing the command. If None, the function
            attempts to determine an appropriate shell based on the platform, environment, and fallback defaults.

    Returns:
        list[str]: A list containing the shell command and its arguments.

    Raises:
        RuntimeError: If no usable shell is found for the current system.
    """
    candidates: list[tuple[str, list[str]]] = []

    # Windows options
    if sys.platform.startswith("win"):
        if preferred_shell:
            candidates.append(
                (preferred_shell, _windows_shell_args(preferred_shell, command))
            )

        env_shell = os.environ.get("CMDBOX_SHELL")
        if env_shell:
            candidates.append((env_shell, _windows_shell_args(env_shell, command)))

        # Known good fallbacks
        candidates.extend(
            [
                ("pwsh", ["pwsh", "-NoProfile", "-Command", command]),
                ("powershell", ["powershell", "-NoProfile", "-Command", command]),
                ("cmd.exe", ["cmd.exe", "/C", command]),
            ]
        )

        for exe, args in candidates:
            if which(exe):
                return args

        # Last resort option
        return ["cmd.exe", "/C", command]

    # Unix options
    if preferred_shell:
        candidates.append((preferred_shell, [preferred_shell, "-lc", command]))

    env_shell = os.environ.get("SHELL")
    if env_shell:
        candidates.append((env_shell, [env_shell, "-lc", command]))

    candidates.extend(
        [
            ("/bin/bash", ["/bin/bash", "-lc", command]),
            ("/bin/sh", ["/bin/sh", "-lc", command]),
        ]
    )

    for exe, args in candidates:
        if os.path.isabs(exe):
            if os.path.exists(exe):
                return args
        elif which(exe):
            return args

    raise RuntimeError("No usable shell found for this system")


def _windows_shell_args(shell: str, command: str) -> list[str]:
    shell = shell.lower()

    if shell in ("pwsh", "powershell"):
        return [shell, "-NoProfile", "-Command", command]

    if shell in ("cmd", "cmd.exe"):
        return ["cmd.exe", "/C", command]

    # Treat unknown shell as executable
    return [shell, command]
