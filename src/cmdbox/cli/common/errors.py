from functools import wraps
from typing import Callable

import typer

from cmdbox.cli.ui.console import ConsoleUI


def make_cli_guard(
    get_console: Callable[[], ConsoleUI],
) -> Callable[[Callable], Callable]:
    """
    Creates a decorator to handle exceptions in CLI commands and provide uniform error
    handling and messaging via a ConsoleUI instance.

    The decorator wraps any given function, intercepting exceptions raised during its execution.
    If an unexpected exception occurs, it is logged using the ConsoleUI's error handler,
    and the program exits with a specific error code.

    Args:
        get_console: Callable that returns a `ConsoleUI` instance. Used to handle error
            messages and output them in a consistent manner during CLI execution.

    Returns:
        Callable: A decorator function that processes the given function, wrapping it with
        exception handling and error reporting logic.
    """

    def cli_guard(fn):
        @wraps(fn)
        def wrapper(*args, **kwargs):
            try:
                return fn(*args, **kwargs)
            except typer.BadParameter:
                raise
            except typer.Exit:
                raise
            except Exception as exc:
                console = get_console()
                console.error(f"{exc}")
                raise typer.Exit(code=1)

        return wrapper

    return cli_guard
