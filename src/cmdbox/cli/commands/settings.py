from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.common.errors import make_cli_guard
from cmdbox.cli.handlers import settings_handler

app = typer.Typer(no_args_is_help=True)

cli_guard = make_cli_guard(container.get_console)


@app.command("edit")
@cli_guard
def edit(
    external: Annotated[
        bool,
        typer.Option(
            "--external",
            "-e",
            help="Edit settings file directly in terminal.",
        ),
    ] = False,
) -> None:
    """
    Edits the settings file. If the `--external` flag is used, the settings file will be
    opened in the default text editor. Otherwise, the settings will be edited interactively
    in the terminal.
    """
    settings_handler.run_edit_settings(
        external,
        get_settings_service=container.get_settings_service,
        get_console=container.get_console,
    )
