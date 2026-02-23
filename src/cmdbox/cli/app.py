from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.ui.presenters.app_presenter import render_version
from cmdbox.version import __version__
from cmdbox.database import ensure_schema, get_db
from .commands.command_crud import app as command_crud_app
from .commands.command_run import app as command_run_app
from .commands.variable_crud import app as variable_crud_app
from .commands.tag_crud import app as tag_crud_app
from .commands.init import app as init_app
from .commands.settings import app as settings_app


app = typer.Typer(
    name="cb",
    help="CmdBox is a CLI tool for storing and recalling commands with many helpful quality of life features.",
    no_args_is_help=True,
)

app.add_typer(command_crud_app, name="cmd", help="CRUD operations for commands.")
app.add_typer(variable_crud_app, name="var", help="CRUD operations for variables.")
app.add_typer(tag_crud_app, name="tag", help="CRUD operations for tags.")
app.add_typer(command_run_app)
app.add_typer(init_app)
app.add_typer(settings_app, name="settings", help="Manage CmdBox settings.")


def is_test_callback(value: bool):
    if value:
        get_db(testing=True)
        console = container.get_console()
        console.info("Testing mode is active, database is in memory.")


def version_callback(value: bool):
    if value:
        console = container.get_console()
        rendered_version = render_version(__version__)
        console.print(rendered_version)
        raise typer.Exit()


@app.callback()
def common(
    test: Annotated[
        bool,
        typer.Option(
            "--test",
            "-t",
            callback=is_test_callback,
            is_flag=True,
            help="Enables testing mode.  Database will be created in memory and will not affect the "
            "applications persistent database.",
        ),
    ] = None,
    version: Annotated[
        bool,
        typer.Option(
            "--version",
            "-V",
            callback=version_callback,
            is_eager=True,
            help="Print the app version and exit.",
        ),
    ] = None,
) -> None:
    ensure_schema()


def main() -> None:
    app()


if __name__ == "__main__":
    main()
