from pathlib import Path
from typing import Annotated, Optional

import typer

from cmdbox import container
from cmdbox.cli.handlers.import_handler import run_import_file

app = typer.Typer(invoke_without_command=True, no_args_is_help=True)


@app.callback()
def import_file(
    path: Annotated[
        Path,
        typer.Argument(help="Path to the JSON export file to import."),
    ],
    overwrite: Annotated[
        bool,
        typer.Option(
            "--overwrite", help="Replace existing items on conflict. Defaults to skip."
        ),
    ] = False,
    preview: Annotated[
        bool,
        typer.Option(
            "--preview", help="Show what would be imported without writing anything."
        ),
    ] = False,
    profile: Annotated[
        Optional[str],
        typer.Option(
            "--profile",
            "-p",
            help="The profile to import into. Defaults to the currently active command profile.",
        ),
    ] = None,
) -> None:
    run_import_file(
        path=path,
        overwrite=overwrite,
        preview=preview,
        profile=profile,
        get_import_service=container.get_import_service,
        get_console=container.get_console,
    )
