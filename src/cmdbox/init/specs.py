import os
from dataclasses import dataclass
from pathlib import Path
from typing import Callable


@dataclass(frozen=True)
class ShellSpec:
    name: str
    filename: str
    default_path_fn: Callable[[], Path] | None
    install_mode: str  # "profile_block" or "write_file" or "wrapper_hint"


def default_bashrc() -> Path:
    return Path.home() / ".bashrc"


def default_zshrc() -> Path:
    return Path.home() / ".zshrc"


def default_fish_function() -> Path:
    return (
        Path.home() / ".config" / "fish" / "functions" / "cb.fish"
    )  # TODO: Does this need to be cbe.fish?


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
