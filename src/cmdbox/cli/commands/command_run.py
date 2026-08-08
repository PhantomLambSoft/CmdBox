from typing import Annotated, Optional
import logging

import typer

from cmdbox import container
from cmdbox.cli.common.errors import make_cli_guard
from cmdbox.cli.completions.commands import complete_command_aliases
from cmdbox.cli.completions.profiles import complete_profile_names
from cmdbox.cli.handlers.run_handler import (
    run_preview_command,
    RawRunContext,
    run_run_command,
)

app = typer.Typer(no_args_is_help=True)

cli_guard = make_cli_guard(container.get_console)

log = logging.getLogger(__name__)


@app.command(
    context_settings={"allow_extra_args": True, "ignore_unknown_options": True}
)
@cli_guard
def run(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to run.",
            autocompletion=complete_command_aliases,
        ),
    ],
    *,
    preview_cmd: Annotated[
        bool,
        typer.Option(
            "--preview",
            "-p",
            help="Output the command that will be executed without actually running it.",
        ),
    ] = False,
    cwd: Annotated[
        str | None,
        typer.Option(
            "--cwd", "-d", help="The working directory for the command execution."
        ),
    ] = None,
    env: Annotated[
        list[str] | None,
        typer.Option("--env", help="The environment variables to set for the command."),
    ] = None,
    capture: Annotated[
        bool,
        typer.Option("--capture", "-c", help="Capture the command output."),
    ] = False,
    shell: Annotated[
        str | None,
        typer.Option(
            "--shell", "-s", help="The shell to use for the command execution."
        ),
    ] = None,
    timeout: Annotated[
        int | None,
        typer.Option(
            "--timeout",
            "-t",
            help="The number of seconds before the process is killed.",
        ),
    ] = None,
    emit: Annotated[
        bool,
        typer.Option(
            "--emit",
            "-e",
            is_flag=True,
            help="Print the command output to stdout instead of running it in a separate process. "
            "You must then press enter to run the command in your current terminal window.",
            hidden=True,
        ),
    ] = False,
    verbose: Annotated[
        Optional[bool],
        typer.Option(
            "--verbose/--no-verbose",
            help="Outputs additional information alongside the command output.",
        ),
    ] = None,
    profile: Annotated[
        Optional[str],
        typer.Option(
            "--profile",
            help="The profile to find this command in. Defaults to the currently active command profile.",
            autocompletion=complete_profile_names,
        ),
    ] = None,
    ctx: typer.Context,
) -> None:
    """
    Run stored commands by using their alias.
    """
    log.debug(
        "run.run called. alias=%s, preview=%s, cwd_used=%s, env=%s, capture=%s, shell=%s, timeout=%s, emit=%s, profile=%s, verbose=%s",
        alias,
        preview_cmd,
        cwd,
        env,
        capture,
        shell,
        timeout,
        emit,
        profile,
        verbose,
    )
    if preview_cmd:
        preview(
            alias=alias,
            cwd=cwd,
            env=env,
            capture=capture,
            shell=shell,
            timeout=timeout,
            verbose=verbose,
            profile=profile,
            ctx=ctx,
        )
        return

    extra_args = ctx.args if ctx.args else ctx.meta.get("_extra_args", [])
    runtime_vars = parse_runtime_vars(extra_args)

    ctx = RawRunContext(
        cwd=cwd,
        env=env,
        capture=capture,
        shell=shell,
        timeout=timeout,
        emit=emit,
        verbose=verbose,
    )
    run_run_command(
        alias=alias,
        runtime_vars=runtime_vars,
        run_ctx=ctx,
        profile=profile,
        get_run_service=container.get_run_service,
        get_settings=container.get_settings,
        get_console=container.get_console,
    )


@app.command(
    "preview",
    context_settings={"allow_extra_args": True, "ignore_unknown_options": True},
)
@cli_guard
def preview(
    alias: Annotated[
        str,
        typer.Argument(
            help="The alias of the command to run.",
            autocompletion=complete_command_aliases,
        ),
    ],
    *,
    cwd: Annotated[
        str | None,
        typer.Option(
            "--cwd", "-d", help="The working directory for the command execution."
        ),
    ] = None,
    env: Annotated[
        list[str] | None,
        typer.Option(
            "--env", "-e", help="The environment variables to set for the command."
        ),
    ] = None,
    capture: Annotated[
        bool,
        typer.Option(
            "--capture", "-c", is_flag=True, help="Capture the command output."
        ),
    ] = False,
    shell: Annotated[
        str | None,
        typer.Option(
            "--shell", "-s", help="The shell to use for the command execution."
        ),
    ] = None,
    timeout: Annotated[
        int | None,
        typer.Option(
            "--timeout",
            "-t",
            help="The number of seconds before the process is killed.",
        ),
    ] = None,
    verbose: Annotated[
        Optional[bool],
        typer.Option(
            "--verbose/--no-verbose",
            help="Outputs additional information alongside the command output.",
        ),
    ] = None,
    profile: Annotated[
        Optional[str],
        typer.Option(
            "--profile",
            help="The profile to find this command in. Defaults to the currently active command profile.",
            autocompletion=complete_profile_names,
        ),
    ] = None,
    ctx: typer.Context,
) -> None:
    """
    Output the command that will be executed without actually running it.
    """
    log.debug(
        "run.preview called. alias=%s, cwd=%s, env=%s, capture=%s, shell=%s, timeout=%s, verbose=%s, profile=%s",
        alias,
        cwd,
        env,
        capture,
        shell,
        timeout,
        verbose,
        profile,
    )
    runtime_vars = parse_runtime_vars(ctx.args)
    ctx = RawRunContext(
        cwd=cwd, env=env, capture=capture, shell=shell, timeout=timeout, verbose=verbose
    )
    run_preview_command(
        alias=alias,
        runtime_vars=runtime_vars,
        run_ctx=ctx,
        profile=profile,
        get_run_service=container.get_run_service,
        get_settings=container.get_settings,
        get_console=container.get_console,
    )


def parse_runtime_vars(args: list[str]) -> dict[str, str]:
    """
    Parses a list of runtime arguments into a dictionary of key-value pairs.

    This function processes a list of command-line arguments, extracting argument
    keys prefixed with `--` and their corresponding values. Arguments without
    values will not be included in the resulting dictionary. Unrecognized or
    improperly formatted tokens are ignored.

    Args:
        args (list[str]): A list of strings, where each string represents a
            command-line argument. Keys should be prefixed with `--` and may
            optionally be followed by a value.

    Returns:
        dict[str, str]: A dictionary where keys are argument names (without the
            `--` prefix) and values are their associated arguments. Only valid
            key-value pairs are included.
    """
    result = {}
    x = 0
    while x < len(args):
        token = args[x]
        if token.startswith("--"):
            key = token.lstrip("-")
            if x + 1 < len(args) and not args[x + 1].startswith("--"):
                result[key] = args[x + 1]
                x += 2
            else:
                x += 1
        else:
            x += 1
    return result
