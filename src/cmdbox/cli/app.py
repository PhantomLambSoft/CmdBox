from typing import Annotated

import typer

from cmdbox.cli.ui import console
from cmdbox.version import __version__
from cmdbox.database import ensure_schema, get_db
from .commands.command_crud import app as command_crud_app
from .commands.command_run import app as command_run_app
from .commands.variable_crud import app as variable_crud_app
from .commands.tag_crud import app as tag_crud_app


app = typer.Typer(
    name="cb",
    help="CmdBox is a CLI tool for storing and recalling commands with many helpful quality of life features.",
    no_args_is_help=True,
)

app.add_typer(command_crud_app, name="cmd")
app.add_typer(variable_crud_app, name="var")
app.add_typer(tag_crud_app, name="tag")
app.add_typer(command_run_app)


def is_test_callback(value: bool):
    if value:
        get_db(testing=True)
        console.print_success("Testing mode is active, database is in memory.")


def version_callback(value: bool):
    if value:
        typer.echo(f"Version: {__version__}")
        raise typer.Exit()


@app.callback()
def common(
    test: Annotated[
        bool,
        typer.Option(
            "--test",
            callback=is_test_callback,
            is_flag=True,
            is_eager=True,
            help="Enables testing mode.  Database will be created in memory and will not affect the "
            "applications persistent database.",
        ),
    ] = None,
    version: Annotated[
        bool,
        typer.Option(
            "--version",
            "-v",
            callback=version_callback,
            is_eager=True,
            help="Print the app version and exit.",
        ),
    ] = None,
) -> None:
    pass


def main() -> None:
    ensure_schema()
    app()


if __name__ == "__main__":
    main()
