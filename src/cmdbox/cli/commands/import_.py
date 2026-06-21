from pathlib import Path
from typing import Annotated

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
) -> None:
    run_import_file(
        path=path,
        overwrite=overwrite,
        preview=preview,
        get_import_service=container.get_import_service,
        get_console=container.get_console,
    )
