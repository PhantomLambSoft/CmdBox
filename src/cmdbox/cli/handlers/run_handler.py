import logging
from typing import Callable
from dataclasses import dataclass

import typer

from cmdbox.cli.prompts.prompts import prompt_for_missing_var
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
    capture: bool | None = None
    shell: str | None = None
    timeout: int | None = None
    emit: bool = False
    verbose: bool | None = None


@log_action(__name__, "run_run_command")
def run_run_command(
    *,
    alias: str,
    runtime_vars: dict[str, str] | None = None,
    run_ctx: RawRunContext | None = None,
    get_run_service: Callable[[], RunService],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
):
    settings = get_settings()
    if run_ctx is None:
        run_ctx = RawRunContext()
    apply_settings_defaults(run_ctx, settings)

    run_service = get_run_service()

    missing = run_service.collect_missing_vars(alias, runtime_vars=runtime_vars)
    if missing:
        for var_name in missing:
            value = prompt_for_missing_var(var_name)
            runtime_vars[var_name] = value

    run_ctx = get_run_ctx(run_ctx)
    ex_result = run_service.run(alias, ctx=run_ctx, runtime_vars=runtime_vars)
    if run_ctx.verbose and not run_ctx.emit:
        console = get_console()
        console.print(render_execution_result(ex_result))


@log_action(__name__, "run_preview_command")
def run_preview_command(
    *,
    alias: str,
    runtime_vars: dict[str, str] | None = None,
    run_ctx: RawRunContext | None = None,
    get_run_service: Callable[[], RunService],
    get_settings: Callable[[], Settings],
    get_console: Callable[[], ConsoleUI],
):
    settings = get_settings()
    if run_ctx is None:
        run_ctx = RawRunContext()
    apply_settings_defaults(run_ctx, settings)

    run_service = get_run_service()

    missing = run_service.collect_missing_vars(alias, runtime_vars=runtime_vars)
    if missing:
        for var_name in missing:
            value = prompt_for_missing_var(var_name)
            runtime_vars[var_name] = value

    run_ctx = get_run_ctx(run_ctx)
    prev_result, effective_ctx = run_service.preview(
        alias, runtime_vars=runtime_vars, ctx=run_ctx
    )
    rendered_result = render_preview_result(prev_result, ctx=effective_ctx)
    console = get_console()
    console.print(rendered_result)


def apply_settings_defaults(run_ctx: RawRunContext, settings: Settings) -> None:
    """
    Adjusts the execution context by applying default settings if certain values
    are not explicitly defined. This ensures that the runtime behavior aligns
    with the provided configuration.

    Args:
        run_ctx: An object representing the current runtime context. It contains
            runtime-specific configurations, such as verbosity, shell usage,
            and output capturing.
        settings: A configuration object that holds execution defaults. These
            defaults are applied to the runtime context when corresponding
            settings are unspecified.
    """
    ex = settings.execution_settings
    if run_ctx.verbose is None:
        run_ctx.verbose = ex.default_verbose
    if run_ctx.capture is None:
        run_ctx.capture = ex.capture_output
    if run_ctx.shell is None:
        run_ctx.shell = ex.default_shell


def get_run_ctx(raw_run_ctx: RawRunContext) -> RunContext:
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
        timeout=raw_run_ctx.timeout,
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
