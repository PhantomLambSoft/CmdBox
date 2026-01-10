from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.completions.commands import complete_command_aliases

app = typer.Typer(no_args_is_help=True)


@app.command()
def run(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to run.",
            autocompletion=complete_command_aliases,
        ),
    ],
    preview: Annotated[
        bool,
        typer.Option(
            "--preview",
            "-p",
            is_flag=True,
            help="Preview the command execution, but do not execute.",
        ),
    ] = False,
):
    console = container.get_console()
    run_service = container.get_run_service()
    if preview:
        result = run_service.preview(alias)
        console.print_run_preview(result)
    else:
        ex_result = run_service.run(alias)
        if ex_result.stderr:
            console.error(ex_result.stderr)
