import os
import sys
import re
import shutil
from dataclasses import dataclass
from typing import Callable
from pathlib import Path
from importlib import resources

import typer
import psutil

from cmdbox.cli.ui.console import ConsoleUI


START_MARK = "# >>> cmdbox shell integration >>>"
END_MARK = "# <<< cmdbox shell integration <<<"


def load_integration_text(filename: str) -> str:
    return (
        resources.files("cmdbox.integrations")
        .joinpath(filename)
        .read_text(encoding="utf-8")
        .rstrip()
        + "\n"
    )


def upsert_marked_block(profile_path: Path, block_text: str) -> None:
    profile_path.parent.mkdir(parents=True, exist_ok=True)
    existing = profile_path.read_text(encoding="utf-8") if profile_path.exists() else ""

    marked = f"{START_MARK}\n{block_text.rstrip()}\n{END_MARK}\n"

    pattern = re.compile(
        re.escape(START_MARK) + r".*?" + re.escape(END_MARK) + r"\n?",
        flags=re.DOTALL,
    )

    if pattern.search(existing):
        new_text = pattern.sub(marked, existing)
    else:
        sep = "\n" if existing and not existing.endswith("\n") else ""
        new_text = existing + sep + marked

    if profile_path.exists():
        backup = profile_path.with_suffix(profile_path.suffix + ".bak")
        shutil.copy2(profile_path, backup)

    profile_path.write_text(new_text, encoding="utf-8")


def default_bashrc() -> Path:
    return Path.home() / ".bashrc"


def default_zshrc() -> Path:
    return Path.home() / ".zshrc"


def default_fish_function() -> Path:
    return Path.home() / ".config" / "fish" / "functions" / "cb.fish"


def default_powershell_profile() -> Path:
    # Approximation that works for typical pwsh installations
    # Users can override with --path if needed
    docs = os.environ.get("USERPROFILE")
    if not docs:
        return (
            Path.home()
            / "Documents"
            / "PowerShell"
            / "Microsoft.PowerShell_profile.ps1"
        )
    return Path(docs) / "Documents" / "PowerShell" / "Microsoft.PowerShell_profile.ps1"


@dataclass(frozen=True)
class ShellSpec:
    name: str
    filename: str
    default_path_fn: callable | None
    install_mode: str  # "profile_block" or "write_file" or "wrapper_hint"


SHELLS: dict[str, ShellSpec] = {
    "bash": ShellSpec("bash", "bash.sh", default_bashrc, "profile_block"),
    "zsh": ShellSpec("zsh", "zsh.sh", default_zshrc, "profile_block"),
    "fish": ShellSpec("fish", "fish.fish", default_fish_function, "write_file"),
    "powershell": ShellSpec(
        "powershell", "powershell.ps1", default_powershell_profile, "profile_block"
    ),
    "pwsh": ShellSpec(
        "pwsh", "powershell.ps1", default_powershell_profile, "profile_block"
    ),
    "cmd": ShellSpec("cmd", "cmd.bat", None, "wrapper_hint"),
}


def run_init_command(
    *,
    shell: str = None,
    install: bool = False,
    path: str = None,
    get_console: Callable[[], ConsoleUI],
):
    """
    Executes the initialization command for shell integration, handling snippet output,
    installation, and providing specific configurations based on the operating system shell.

    This function manages different installation modes (e.g., updating profile files,
    writing integration files, or providing wrapper hints) and facilitates smooth integration
    of the application with the user's shell environment.

    Args:
        shell (str): The shell name provided for integration. Must match one of the expected keys
            in the `SHELLS` dictionary (case-insensitive).
        install (bool): Specifies whether to install the integration snippet to the shell's
            configuration file or output the snippet to stdout. Defaults to `False`.
        path (str, optional): A custom path to the file to which the integration snippet
            will be written when `install` is `True`. If not provided, defaults to the
            path determined by the shell specification.
        get_console (Callable[[], ConsoleUI]): A callable returning an instance of `ConsoleUI`
            to handle console output and user feedback.

    Raises:
        typer.BadParameter: If an invalid shell name is provided that does not exist in the
            `SHELLS` dictionary.
    """
    console = get_console()

    if not shell:
        shell = detect_shell()
    shell_key = shell.lower()
    if shell_key not in SHELLS:
        raise typer.BadParameter(f"Invalid shell: {shell}")

    spec = SHELLS[shell_key]
    snippet = load_integration_text(spec.filename)

    if not install:
        line_sep = "-----------------------------------------------------"
        console.print(line_sep, snippet, line_sep, sep="\n")
        console.warning(
            "TODO: provide link to website page for manual install instructions"
        )
        return

    if spec.install_mode == "profile_block":
        target = path or spec.default_path_fn()
        upsert_marked_block(target, snippet)
        console.success(f"Shell integration installed successfully in {target}.")
        console.print("Restart your shell for changes to take effect.")
        return

    if spec.install_mode == "write_file":
        target = path or spec.default_path_fn()
        if target.exists():
            backup = target.with_suffix(target.suffix + ".bak")
            shutil.copy2(target, backup)
        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(snippet, encoding="utf-8")
        console.success(f"Shell integration installed written to: {target}.")
        console.print("Restart your shell for changes to take effect.")
        return

    if spec.install_mode == "wrapper_hint":
        # TODO: Provide link to documentation website when available
        message = (
            "cmd.exe integration is only supported as a wrapper script, not a profile snippet. "
            'To install cmd.exe integration, create a "cb.cmd" file on your PATH containing the '
            "contents below:"
        )
        console.print(message)
        console.print(snippet, end="")
        return


def run_detect_shell(*, get_console: Callable[[], ConsoleUI]):
    """
    Detects the current shell being used and prints it to the console.

    This function retrieves a `ConsoleUI` object using the provided `get_console`
    callable. It then detects the current shell and sends a message with the name
    of the detected shell to be displayed on the console.

    Args:
        get_console (Callable[[], ConsoleUI]): A callable that, when invoked,
            returns an instance of a `ConsoleUI` object used for displaying the
            detected shell information.
    """
    console = get_console()
    shell = detect_shell()
    console.print(f"Detected shell: {shell}")


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
    parent = psutil.Process(os.getppid()).parent()
    name = (parent.name() or "").lower()

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
