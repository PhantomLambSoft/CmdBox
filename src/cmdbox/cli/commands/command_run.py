from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.commands.command_crud import cli_guard
from cmdbox.cli.completions.commands import complete_command_aliases
from cmdbox.cli.handlers.run_handler import (
    run_preview_command,
    RawRunContext,
    run_run_command,
)


app = typer.Typer(no_args_is_help=True)


@app.command()
@cli_guard
def run(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to run.",
            autocompletion=complete_command_aliases,
        ),
    ],
    preview_cmd: Annotated[bool, typer.Option("--preview", "-p", is_flag=True)] = False,
    cwd: Annotated[
        str,
        typer.Option(
            "--cwd", "-d", help="The working directory for the command execution."
        ),
    ] = None,
    env: Annotated[
        list[str],
        typer.Option(
            "--env", "-e", help="The environment variables to set for the command."
        ),
    ] = None,
    capture: Annotated[
        bool,
        typer.Option(
            "--capture", "-c", is_flag=True, help="Capture the command output."
        ),
    ] = False,
    shell: Annotated[
        str,
        typer.Option(
            "--shell", "-s", help="The shell to use for the command execution."
        ),
    ] = None,
):
    if preview_cmd:
        preview(alias=alias, cwd=cwd, env=env, capture=capture, shell=shell)
        return
    ctx = RawRunContext(cwd=cwd, env=env, capture=capture, shell=shell)
    run_run_command(
        alias=alias,
        run_ctx=ctx,
        get_run_service=container.get_run_service,
        get_console=container.get_console,
    )


@app.command("preview")
@cli_guard
def preview(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to run.",
            autocompletion=complete_command_aliases,
        ),
    ],
    cwd: Annotated[
        str,
        typer.Option(
            "--cwd", "-d", help="The working directory for the command execution."
        ),
    ] = None,
    env: Annotated[
        list[str],
        typer.Option(
            "--env", "-e", help="The environment variables to set for the command."
        ),
    ] = None,
    capture: Annotated[
        bool,
        typer.Option(
            "--capture", "-c", is_flag=True, help="Capture the command output."
        ),
    ] = False,
    shell: Annotated[
        str,
        typer.Option(
            "--shell", "-s", help="The shell to use for the command execution."
        ),
    ] = None,
):
    ctx = RawRunContext(cwd=cwd, env=env, capture=capture, shell=shell)
    run_preview_command(
        alias=alias,
        run_ctx=ctx,
        get_run_service=container.get_run_service,
        get_console=container.get_console,
    )
