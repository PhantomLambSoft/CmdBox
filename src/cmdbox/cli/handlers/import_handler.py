from pathlib import Path
from typing import Callable

import typer

from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.cli.ui.presenters.import_presenter import (
    render_import_preview,
    render_import_result,
)
from cmdbox.logging_setup.log_decorators import log_action
from cmdbox.services.errors import ImportCycleError, ImportFileError
from cmdbox.services.import_service import ImportService


@log_action(__name__, "run_import_file")
def run_import_file(
    *,
    path: Path,
    overwrite: bool,
    preview: bool,
    profile: str | None,
    get_import_service: Callable[[], ImportService],
    get_console: Callable[[], ConsoleUI],
) -> None:
    console = get_console()
    try:
        import_service = get_import_service()
        result = import_service.import_file(
            path, preview=preview, overwrite=overwrite, profile=profile
        )
    except ImportFileError as e:
        console.error(str(e))
        raise typer.Exit(code=1)
    except ImportCycleError as e:
        console.error(
            f"Import rejected - circular dependency detected: {' -> '.join(e.cycle)}"
        )
        raise typer.Exit(code=1)

    if preview:
        console.print(render_import_preview(result, source=str(path)))
    else:
        console.print(render_import_result(result))
