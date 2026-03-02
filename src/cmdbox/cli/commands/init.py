import logging
from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.common.errors import make_cli_guard
from cmdbox.cli.handlers.init_handler import run_detect_shell, run_init_command

app = typer.Typer(no_args_is_help=False)

cli_guard = make_cli_guard(container.get_console)

log = logging.getLogger(__name__)


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
    log.debug("init.init called. shell=%s, install=%s", shell, install)
    run_init_command(
        shell=shell, install=install, path=path, get_console=container.get_console
    )


@app.command(name="shell", hidden=True)
@cli_guard
def detect_shell() -> None:
    """Detects and prints the current shell being used."""
    log.debug("init.shell called.")
    run_detect_shell(get_console=container.get_console)
