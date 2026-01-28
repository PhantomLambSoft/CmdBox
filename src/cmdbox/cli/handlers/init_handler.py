import shutil
from pathlib import Path
from typing import Callable

import typer

from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.cli.ui.presenters.init_presenter import (
    render_install_instructions,
    render_install_success,
    render_shell_output,
)
from cmdbox.init.detect import detect_shell
from cmdbox.init.io import load_integration_text, upsert_marked_block
from cmdbox.init.specs import SHELLS


def run_init_command(
    *,
    shell: str = None,
    install: bool = False,
    path: str = None,
    get_console: Callable[[], ConsoleUI],
) -> None:
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
        title = f"Install instructions for {spec.name}"
        console.print(
            render_install_instructions(
                snippet, shell=shell, title=title, include_help_text=shell != "cmd"
            )
        )
        return

    target = (
        Path(path)
        if path
        else (spec.default_path_fn() if spec.default_path_fn else None)
    )

    if spec.install_mode == "profile_block":
        if target is None:
            raise typer.BadParameter(
                f"No default install path available for shell: {spec.name}."
            )
        upsert_marked_block(target, snippet)
        console.print(render_install_success())
        return

    if spec.install_mode == "write_file":
        if target is None:
            raise typer.BadParameter(
                f"No default install path available for shell: {spec.name}."
            )

        if target.exists():
            backup = target.with_suffix(target.suffix + ".bak")
            shutil.copy2(target, backup)

        target.parent.mkdir(parents=True, exist_ok=True)
        target.write_text(snippet, encoding="utf-8")
        console.print(render_install_success())
        return

    if spec.install_mode == "wrapper_hint":
        title = "Unable to install snippet for cmd shell"
        console.print(
            render_install_instructions(
                snippet, shell="cmd", title=title, include_help_text=False
            )
        )
        return


def run_detect_shell(*, get_console: Callable[[], ConsoleUI]) -> None:
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
    console.print(render_shell_output(shell))
