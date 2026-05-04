from typing import Callable

from cmdbox.cli.prompts.prompts import prompt_for_confirm
from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.cli.ui.presenters.history_presenter import (
    render_history_list,
    render_history_entry,
    render_history_cleared,
)
from cmdbox.logging_setup.log_decorators import log_action
from cmdbox.runtime.executor import RunContext
from cmdbox.services.history_service import HistoryService
from cmdbox.services.run_service import RunService


@log_action(__name__, "run_history_list")
def run_history_list(
    *,
    alias: str | None,
    limit: int,
    get_history_service: Callable[[], HistoryService],
    get_console: Callable[[], ConsoleUI],
) -> None:
    service = get_history_service()
    entries = service.get_recent(alias=alias, limit=limit)
    console = get_console()
    if not entries:
        console.info("No history found")
        return
    console.print(render_history_list(entries))


@log_action(__name__, "run_history_show")
def run_history_show(
    *,
    ref: str,
    get_history_service: Callable[[], HistoryService],
    get_console: Callable[[], ConsoleUI],
) -> None:
    service = get_history_service()
    entry = service.get_by_ref(ref)
    variables = service.get_variables(entry)
    console = get_console()
    console.print(render_history_entry(entry, variables))


@log_action(__name__, "run_history_rerun")
def run_history_rerun(
    *,
    ref: str,
    get_history_service: Callable[[], HistoryService],
    get_run_service: Callable[[], RunService],
) -> None:
    history_service = get_history_service()
    entry = history_service.get_by_ref(ref)
    variables = history_service.get_variables(entry)
    run_service = get_run_service()
    run_service.run(entry.alias, ctx=RunContext(), runtime_vars=variables)


@log_action(__name__, "run_history_clear")
def run_history_clear(
    *,
    alias: str | None,
    yes: bool,
    get_history_service: Callable[[], HistoryService],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    if not yes:
        scope = f" for '{alias}'" if alias else ""
        if not prompt_for_confirm(f"Clear all history{scope}?"):
            console.info("Aborted")
            return
    service = get_history_service()
    count = service.clear(alias=alias)
    console.print(render_history_cleared(count, alias))
