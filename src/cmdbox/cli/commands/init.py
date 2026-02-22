from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.commands.command_crud import cli_guard
from cmdbox.cli.handlers.init_handler import run_detect_shell, run_init_command

app = typer.Typer(no_args_is_help=False)


@app.command()
@cli_guard
def init(
    shell: Annotated[
        str,
        typer.Argument(
            help="The shell to initialize. bash, zsh, fish, powershell, or cmd"
        ),
    ] = None,
    install: Annotated[
        bool,
        typer.Option(
            "--install",
            "-i",
            is_flag=True,
            help="Install into the shell profile where applicable. If False, the profile modification will be"
            "printed for you to make the profile changes manually.  This may not work on all profiles.",
        ),
    ] = False,
    path: Annotated[
        str, typer.Option("--path", "-p", help="The path to the shell profile.")
    ] = None,
) -> None:
    """
    Initialize CmdBox for a shell environment.
    """
    run_init_command(
        shell=shell, install=install, path=path, get_console=container.get_console
    )


@app.command(name="shell", hidden=True)
@cli_guard
def detect_shell() -> None:
    """Detects and prints the current shell being used."""
    run_detect_shell(get_console=container.get_console)
