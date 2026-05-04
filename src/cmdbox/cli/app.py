import uuid
from typing import Annotated

import typer

from cmdbox import container
from cmdbox.cli.commands.alias_fallback import AliasFallbackGroup
from cmdbox.cli.ui.presenters.app_presenter import render_version
from cmdbox.logging_setup.log_handlers import configure_logging
from cmdbox.logging_setup.log_config import build_log_config, get_logger
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
    cls=AliasFallbackGroup,
    help="CmdBox is a CLI tool for storing and recalling commands with many helpful quality of life features.",
    no_args_is_help=True,
)

app.add_typer(
    command_crud_app,
    name="cmd",
    help="CRUD (Create, Read, Update, Delete) operations for commands.",
)
app.add_typer(variable_crud_app, name="var", help="CRUD operations for variables.")
app.add_typer(tag_crud_app, name="tag", help="CRUD operations for tags.")
app.add_typer(command_run_app)
app.add_typer(init_app)
app.add_typer(settings_app, name="settings", help="Manage CmdBox settings.")


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
            help="Enables testing mode.  Database will be created in memory and will not affect the "
            "applications persistent database.",
        ),
    ] = False,
    verbose: Annotated[
        bool,
        typer.Option(
            "--verbose",
            "-v",
            help="Enable additional diagnostic output in the terminal. Sets console log level to INFO.",
        ),
    ] = False,
    debug: Annotated[
        bool,
        typer.Option(
            "--debug",
            "-d",
            help="Enable full diagnostic output in the terminal. Sets console log level to DEBUG.",
        ),
    ] = False,
    file_logs: Annotated[
        bool | None,
        typer.Option(
            "--file-logs/--no-file-logs",
            help="Enable/disables writing diagnostic logs to a file. Defaults to settings.",
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
    if test:
        get_db(testing=True)

    settings = container.get_settings()

    run_id = uuid.uuid4().hex[:6]
    log_config = build_log_config(
        settings=settings, verbose=verbose, debug=debug, file_logs=file_logs
    )
    configure_logging(log_config, run_id=run_id)

    log = get_logger()
    log.debug(
        f"startup: test={test}, verbose={verbose}, debug={debug}, file_logs={file_logs}"
    )
    log.debug(f"file_logging={log_config.file_enabled} path={log_config.file_path}")

    ensure_schema()

    if test:
        console = container.get_console()
        console.info("Testing mode is active, database is in memory.")


def main() -> None:
    app()


if __name__ == "__main__":
    main()
