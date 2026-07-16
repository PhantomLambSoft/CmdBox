import logging
import os
import platform
import subprocess
from typing import Callable

import typer

from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.cli.ui.editor import edit_text_fullscreen, EditCanceled, edit_text_in_editor
from cmdbox.cli.ui.presenters.settings_presenter import render_settings_show
from cmdbox.logging_setup.log_decorators import log_action
from cmdbox.settings.settings_service import SettingsService
from cmdbox.core.paths import APP_DATA_DIR

log = logging.getLogger(__name__)


@log_action(__name__, "run_edit_settings")
def run_edit_settings(
    external: bool = True,
    *,
    get_settings_service: Callable[[], SettingsService],
    get_console: Callable[[], ConsoleUI],
):
    settings_service = get_settings_service()
    console = get_console()

    def editor_fn(initial_text: str) -> str:
        if external:
            return edit_text_in_editor(
                initial_text, suffix=".toml", title_hint="CmdBox Settings"
            )
        else:
            return edit_text_fullscreen(initial_text, title="Edit Settings")

    try:
        settings_service.edit(editor_fn)
        console.success("Settings saved.")
    except EditCanceled:
        console.info("Settings not saved.")
        raise typer.Exit(code=0)
    except Exception as exc:
        log.error("Error saving settings.", exc_info=True)
        console.error(f"Error saving settings: {exc}")
        raise typer.Exit(code=1)


@log_action(__name__, "run_show_settings")
def run_show_settings(
    fields: str | None = None,
    *,
    get_settings_service: Callable[[], SettingsService],
    get_console: Callable[[], ConsoleUI],
):
    parsed_fields = fields.split(",") if fields else None

    settings_service = get_settings_service()
    console = get_console()

    settings = settings_service.get()
    output = render_settings_show(settings, parsed_fields)
    console.print(output)


@log_action(__name__, "run_open_data_dir")
def run_open_data_dir(
    print_only: bool = False, *, get_console: Callable[[], ConsoleUI]
) -> None:
    if print_only:
        console = get_console()
        console.print(APP_DATA_DIR)
        return
    system = platform.system()
    if system == "Windows":
        os.startfile(APP_DATA_DIR)
    elif system == "Darwin":
        subprocess.run(["open", str(APP_DATA_DIR)], check=True)
    else:
        subprocess.run(["xdg-open", str(APP_DATA_DIR)], check=True)
