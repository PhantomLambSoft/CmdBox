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
    cmd_profile: Annotated[
        Optional[str],
        typer.Option(
            "--cmd_profile",
            "-p",
            help="The profile to export commands from. Defaults to the currently active profile.",
        ),
    ] = None,
    var_profile: Annotated[
        Optional[str],
        typer.Option(
            "--var_profile",
            "-v",
            help="The profile to export variables from. Defaults to the currently active profile.",
        ),
    ] = None,
) -> None:
    run_export_cmds(
        aliases=aliases,
        tag=tag,
        flatten=flatten,
        output=output,
        cmd_profile=cmd_profile,
        var_profile=var_profile,
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
    var_profile: Annotated[
        Optional[str],
        typer.Option(
            "--var_profile",
            "-v",
            help="The profile to export variables from. Defaults to the currently active profile.",
        ),
    ] = None,
) -> None:
    run_export_vars(
        names=names,
        tag=tag,
        flatten=flatten,
        output=output,
        var_profile=var_profile,
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
    cmd_profile: Annotated[
        Optional[str],
        typer.Option(
            "--cmd_profile",
            "-p",
            help="The profile to export commands from. Defaults to the currently active profile.",
        ),
    ] = None,
    var_profile: Annotated[
        Optional[str],
        typer.Option(
            "--var_profile",
            "-v",
            help="The profile to export variables from. Defaults to the currently active profile.",
        ),
    ] = None,
) -> None:
    run_export_all(
        flatten=flatten,
        output=output,
        cmd_profile=cmd_profile,
        var_profile=var_profile,
        get_export_service=container.get_export_service,
        get_console=container.get_console,
    )
