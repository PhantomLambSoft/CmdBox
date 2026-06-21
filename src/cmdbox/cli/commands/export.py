from typing import Annotated, Optional

import typer

from cmdbox import container
from cmdbox.cli.handlers.export_handler import (
    run_export_all,
    run_export_cmds,
    run_export_vars,
)

app = typer.Typer(no_args_is_help=True)


@app.command("cmds", help="Export commands and their dependencies to a JSON file.")
def export_cmds(
    aliases: Annotated[
        Optional[list[str]],
        typer.Argument(help="Aliases to export. Exports all commands if omitted."),
    ] = None,
    tag: Annotated[
        Optional[str],
        typer.Option(
            "--tag", "-t", help="Filter by tag. Ignored if aliases are given."
        ),
    ] = None,
    flatten: Annotated[
        bool,
        typer.Option(
            "--flatten", "-f", is_flag=True, help="Inline all references recursively."
        ),
    ] = False,
    output: Annotated[
        Optional[str],
        typer.Option(
            "--output",
            "-o",
            help="Path to the output JSON file. Defaults to current directory.",
        ),
    ] = None,
) -> None:
    run_export_cmds(
        aliases=aliases,
        tag=tag,
        flatten=flatten,
        output=output,
        get_export_service=container.get_export_service,
        get_console=container.get_console,
    )


@app.command("vars", help="Export variables and their dependencies to a JSON file.")
def export_vars(
    names: Annotated[
        Optional[list[str]],
        typer.Argument(
            help="Names of variables to export. Exports all variables if omitted."
        ),
    ] = None,
    tag: Annotated[
        Optional[str],
        typer.Option("--tag", "-t", help="Filter by tag. Ignored if names are given."),
    ] = None,
    flatten: Annotated[
        bool,
        typer.Option(
            "--flatten", "-f", is_flag=True, help="Inline all references recursively."
        ),
    ] = False,
    output: Annotated[
        Optional[str],
        typer.Option(
            "--output",
            "-o",
            help="Path to the output JSON file. Defaults to current directory.",
        ),
    ] = None,
) -> None:
    run_export_vars(
        names=names,
        tag=tag,
        flatten=flatten,
        output=output,
        get_export_service=container.get_export_service,
        get_console=container.get_console,
    )


@app.command(
    "all", help="Export all commands, variables, and their dependencies to a JSON file."
)
def export_all(
    flatten: Annotated[
        bool,
        typer.Option(
            "--flatten", "-f", is_flag=True, help="Inline all references recursively."
        ),
    ] = False,
    output: Annotated[
        Optional[str],
        typer.Option(
            "--output",
            "-o",
            help="Path to the output JSON file. Defaults to current directory.",
        ),
    ] = None,
) -> None:
    run_export_all(
        flatten=flatten,
        output=output,
        get_export_service=container.get_export_service,
        get_console=container.get_console,
    )
