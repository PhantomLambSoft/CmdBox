from typing import Callable

from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.cli.ui.presenters.export_presenter import render_export_result
from cmdbox.logging_setup.log_decorators import log_action
from cmdbox.services.export_service import ExportService


@log_action(__name__, "run_export_cmds")
def run_export_cmds(
    *,
    aliases: list[str] | None,
    tag: str | None,
    flatten: bool,
    output: str | None,
    get_export_service: Callable[[], ExportService],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    export_service = get_export_service()
    result = export_service.export_cmds(
        aliases=aliases, tag=tag, flatten=flatten, output_path=output
    )
    for warning in result.warnings:
        console.warning(warning)
    console.print(render_export_result(result))


@log_action(__name__, "run_export_vars")
def run_export_vars(
    *,
    names: list[str] | None,
    tag: str | None,
    flatten: bool,
    output: str | None,
    get_export_service: Callable[[], ExportService],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    export_service = get_export_service()
    result = export_service.export_vars(
        names=names, tag=tag, flatten=flatten, output_path=output
    )
    for warning in result.warnings:
        console.warning(warning)
    console.print(render_export_result(result))


@log_action(__name__, "run_export_all")
def run_export_all(
    *,
    flatten: bool,
    output: str | None,
    get_export_service: Callable[[], ExportService],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    export_service = get_export_service()
    result = export_service.export_all(flatten=flatten, output_path=output)
    for warning in result.warnings:
        console.warning(warning)
    console.print(render_export_result(result))
