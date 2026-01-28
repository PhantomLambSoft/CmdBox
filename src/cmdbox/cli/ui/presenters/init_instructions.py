import os
from pathlib import Path


"""
Instructions for installing the snippet into the user's shell configuration file.
"""


bash_path = os.path.join(Path.home(), ".bashrc")
bash = [
    f'Locate your shell configuration file (usually located at: "{bash_path}").',
    "Add the code snippet to the end of the file.",
    "Restart your shell session for the changes to take effect.",
]

zsh_path = os.path.join(Path.home(), ".zshrc")
zsh = [
    f'Locate your shell configuration file (usually located at: "{zsh_path}").',
    "Add the code snippet to the end of the file.",
    "Restart your shell session for the changes to take effect.",
]

fish_path = os.path.join(Path.home(), ".config", "fish", "config.fish")
fish = [
    f'Locate your shell configuration file (usually located at: "{fish_path}").',
    "Add the code snippet to the end of the file.",
    "Restart your shell session for the changes to take effect.",
]

powershell_path = os.path.join(
    Path.home(), "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1"
)
powershell = [
    f'Locate your powershell configuration file (usually located at: "{powershell_path}"). ',
    "If you do not find the correct configuration file here, you can run the command `$PROFILE` from your powershell terminal to output the exact location.",
    "Add the code snippet to the end of the file.",
    "Restart your shell session for the changes to take effect.",
]

pwsh = powershell

cmd = [
    "cmd.exe integration is only supported as a wrapper script, not a profile snippet.",
    'Create a file named "cbe.cmd" (the file can be created anywhere on your system).',
    "Paste the snippet into this file and save it.",
    "Ensure the file is on your PATH.",
    "Restart your shell for changes to take effect.",
]


def get_instructions(shell: str) -> list[str]:
    return globals()[shell.lower()]
