import logging
from typing import Annotated, Optional

import typer

from cmdbox import container
from cmdbox.cli.handlers.history_handlers import (
    run_history_list,
    run_history_show,
    run_history_rerun,
    run_history_clear,
    run_rerun_last,
)
from cmdbox.cli.common.errors import make_cli_guard

app = typer.Typer(help="View and re-run command execution history.")

cli_guard = make_cli_guard(container.get_console)

log = logging.getLogger(__name__)


@app.command("list")
@cli_guard
def history_list(
    alias: Annotated[
        Optional[str],
        typer.Option("--alias", "-a", help="Filter history to a specific alias."),
    ] = None,
    limit: Annotated[
        int,
        typer.Option("--limit", "-l", help="The number of history entries to show."),
    ] = 25,
) -> None:
    """List recent command executions."""
    run_history_list(
        alias=alias,
        limit=limit,
        get_history_service=container.get_history_service,
        get_console=container.get_console,
    )


@app.command("show")
@cli_guard
def history_show(
    ref: Annotated[
        str,
        typer.Argument(help="Entry index (from list) or ID prefix."),
    ],
) -> None:
    """Show full details for a single history entry."""
    run_history_show(
        ref=ref,
        get_history_service=container.get_history_service,
        get_console=container.get_console,
    )


@app.command("rerun")
@cli_guard
def history_rerun(
    ref: Annotated[
        str,
        typer.Argument(help="Entry index (from list) or ID prefix."),
    ],
) -> None:
    """Re-execute a past command invocation using the same variables."""
    run_history_rerun(
        ref=ref,
        get_history_service=container.get_history_service,
        get_run_service=container.get_run_service,
    )


@app.command("clear")
@cli_guard
def history_clear(
    alias: Annotated[
        Optional[str],
        typer.Option("--alias", "-a", help="Clear history only for a specific alias."),
    ] = None,
    yes: Annotated[
        bool,
        typer.Option(
            "--yes",
            "-y",
            help="Skip confirmation prompt.",
        ),
    ] = False,
) -> None:
    """Clear command history."""
    run_history_clear(
        alias=alias,
        yes=yes,
        get_history_service=container.get_history_service,
        get_console=container.get_console,
    )


@app.command("last")
@cli_guard
def rerun_last() -> None:
    """Re-executes the last command run."""
    run_rerun_last(
        get_run_service=container.get_run_service,
        get_history_service=container.get_history_service,
        get_console=container.get_console,
    )
