import logging
from typing import Callable
from dataclasses import dataclass

import typer

from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.cli.ui.presenters.result_presenter import (
    render_execution_result,
    render_preview_result,
)
from cmdbox.runtime.executor import RunContext
from cmdbox.services.run_service import RunService
from cmdbox.settings.models import Settings
from cmdbox.logging_setup.log_decorators import log_action

log = logging.getLogger(__name__)


@dataclass(frozen=False)
class RawRunContext:
    """
    This dataclass is effectively the same as the actual RunContext
    except that it takes a different 'pre-parsed' version of the
    env argument.  This class takes the argument in a format that
    can be supplied by the user, which must then be parsed into a
    format that can be used by the RunContext.
    """

    cwd: str | None = None
    env: list[str] | str | None = None
    capture: bool = False
    shell: str | None = None
    emit: bool = False
    verbose: bool | None = False


@log_action(__name__, "run_run_command")
def run_run_command(
    *,
    alias: str,
    run_ctx: RawRunContext | None = None,
    get_run_service: Callable[[], RunService],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
):
    if run_ctx and run_ctx.verbose is None:
        settings = get_settings()
        run_ctx.verbose = settings.execution_settings.default_verbose
    run_service = get_run_service()
    run_ctx = get_run_ctx(run_ctx) if run_ctx else RunContext()
    ex_result = run_service.run(alias, ctx=run_ctx)
    if run_ctx.verbose and not run_ctx.emit:
        console = get_console()
        console.print(render_execution_result(ex_result))


@log_action(__name__, "run_preview_command")
def run_preview_command(
    *,
    alias: str,
    run_ctx: RawRunContext | None = None,
    get_run_service: Callable[[], RunService],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
):
    if run_ctx and run_ctx.verbose is None:
        settings = get_settings()
        run_ctx.verbose = settings.execution_settings.default_verbose
    run_service = get_run_service()
    run_ctx = get_run_ctx(run_ctx) if run_ctx else RunContext()
    prev_result = run_service.preview(alias)
    rendered_result = render_preview_result(prev_result, ctx=run_ctx)
    console = get_console()
    console.print(rendered_result)


def get_run_ctx(raw_run_ctx: RawRunContext | None) -> RunContext:
    """
    Creates and returns a `RunContext` object based on the provided `raw_run_ctx`.

    If the `raw_run_ctx` is not provided (i.e., is None), a default `RunContext` is
    created and returned. If `raw_run_ctx` is provided, its properties are used to
    initialize the `RunContext`, and its environment is parsed before being passed
    to the new context.

    Args:
        raw_run_ctx: An instance of `RawRunContext` or None. If provided, it must
            contain context information such as current working directory, shell
            configuration, and environment variables.

    Returns:
        RunContext: A `RunContext` object constructed using the provided
        `raw_run_ctx` or with default values if no `raw_run_ctx` is provided.
    """
    if raw_run_ctx is None:
        return RunContext()

    env = parse_env(raw_run_ctx.env)
    return RunContext(
        cwd=raw_run_ctx.cwd,
        env=env,
        capture=raw_run_ctx.capture,
        shell=raw_run_ctx.shell,
        emit=raw_run_ctx.emit,
        verbose=raw_run_ctx.verbose,
    )


def parse_env(env: list[str] | str | None) -> dict[str, str] | None:
    """
    Parses a list, string, or None value into a dictionary where each key-value pair
    represents an environment variable, with the key and value extracted from the
    input format. If the input is None, returns None.

    Args:
        env (list[str] | str | None): A list of strings, a comma-separated string,
            or None, where each string is in the format "key=value".

    Returns:
        dict[str, str] | None: A dictionary containing the parsed environment
            variables as key-value pairs, or None if the input is None.

    Raises:
        typer.BadParameter: If an entry in the input does not follow the
            "key=value" format.
    """
    if env is None:
        return None

    ret = {}

    if isinstance(env, str):
        env = env.split(",")

    def split_pair(pair: str) -> tuple[str, str]:
        if "=" not in pair:
            log.error(f"Invalid environment variable format: {pair}")
            raise typer.BadParameter(
                "Invalid environment variable format. Each env must be in the format of key=value."
            )
        k, v = pair.split("=", maxsplit=1)
        log.debug(f"Parsed env: {k}={v}")
        return k, v

    for pair in env:
        if "," in pair:
            for k, v in map(split_pair, pair.split(",")):
                ret[k] = v
        else:
            k, v = split_pair(pair)
            ret[k] = v

    log.debug(f"Parsed env: {ret}")
    return ret
