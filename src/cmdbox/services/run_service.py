import json
from typing import Callable

from cmdbox.models import Command
from cmdbox.repositories.command_repository import CommandRepository
from cmdbox.repositories.history_repository import HistoryRepository
from cmdbox.resolve.resolver import Resolver
from cmdbox.resolve.type_defs import ResolveResult
from cmdbox.runtime.executor import Executor, RunContext
from cmdbox.runtime.results import ExecutionResult
from cmdbox.settings.models import Settings


class RunService:
    """
    Provides services for executing and previewing commands.

    This class manages the execution of commands and their previews. It interacts with
    a command repository, a resolver for processing commands, and an executor for
    executing commands. The `run` method is used to execute a command, while the `preview`
    method enables retrieving a resolved result of a command without executing it.

    Attributes:
        repo (CommandRepository): The repository that stores and provides access to
            command definitions based on their alias.
        resolver (Resolver): Encapsulates the logic required to resolve a command
            template into its executable form.
        executor (Executor): Responsible for executing resolved commands and
            returning the results.
    """

    def __init__(
        self,
        repo: CommandRepository,
        resolver: Resolver,
        executor: Executor,
        history_repo: HistoryRepository | None = None,
        get_settings: Callable[[], Settings] | None = None,
    ) -> None:
        self._repo = repo
        self._resolver = resolver
        self._executor = executor
        self._history_repo = history_repo
        self._get_settings = get_settings

    def run(
        self,
        command_alias: str,
        ctx: RunContext | None = None,
        runtime_vars: dict[str, str] | None = None,
    ) -> ExecutionResult:
        """
        Executes a command based on the given alias.

        This method retrieves a command associated with the provided alias, resolves its
        template using a resolver, and then executes the resolved command text. Execution
        context stored on the command (cwd, shell, env, timeout) is merged with any
        runtime-supplied context, with runtime values taking precedence.

        Args:
            command_alias (str): The alias for the command to be executed.
            ctx (RunContext | None): The context for running the command.
            runtime_vars (dict[str, str] | None): Runtime variables to be used during command execution.

        Returns:
            ExecutionResult: The result of executing the command.
        """
        cmd = self._repo.get_by_alias(command_alias)
        resolved_cmd = self._resolver.resolve(cmd.template, runtime_vars=runtime_vars)
        ctx = self.build_context(cmd, ctx)
        result = self._executor.run(resolved_cmd.text, ctx=ctx)
        self._repo.record_use(cmd.id)
        self.record_history(
            alias=command_alias,
            template=cmd.template,
            resolved=resolved_cmd.text,
            runtime_vars=runtime_vars,
            exit_code=result.exit_code,
        )
        return result

    def preview(
        self,
        command_alias: str,
        runtime_vars: dict[str, str] | None = None,
        ctx: RunContext | None = None,
    ) -> tuple[ResolveResult, RunContext | None]:
        """
        Retrieves and resolves a command template based on its alias.

        This method takes a command alias, retrieves the associated command template,
        and resolves it to produce a complete command configuration or data structure.
        This resolved template is then returned as a `ResolveResult`.

        Args:
            command_alias (str): The alias of the command to be resolved.
            runtime_vars (dict[str, str] | None): Runtime variables to be used during command resolution.
            ctx (RunContext | None): The context in which the command is being executed.

        Returns:
            tuple[ResolveResult, RunContext | None]: The resolved result of the command template and the effective context.
        """
        cmd = self._repo.get_by_alias(command_alias)
        resolved = self._resolver.resolve(cmd.template, runtime_vars=runtime_vars)
        effective_ctx = self.build_context(cmd, ctx)
        return resolved, effective_ctx

    def record_history(
        self,
        *,
        alias: str,
        template: str,
        resolved: str,
        runtime_vars: dict[str, str] | None = None,
        exit_code: int | None = None,
    ) -> None:
        """
        Records command execution history if the history feature is enabled and
        properly configured. This function interacts with an underlying
        repository to store details about the executed command including alias,
        template, resolved string, runtime variables, and exit code.

        Args:
            alias (str): The alias or the name of the command executed.
            template (str): The template string representing the command.
            resolved (str): The fully resolved string of the executed command.
            runtime_vars (dict[str, str] | None): Runtime variables used during
                the execution of the command. Defaults to None.
            exit_code (int | None): Exit code that indicates the result of the
                command execution. Defaults to None.
        """
        if not self._history_repo or not self._get_settings:
            return
        settings = self._get_settings()
        if not settings.history.enabled:
            return
        self._history_repo.record(
            alias=alias,
            template=template,
            resolved=resolved,
            variables_used=runtime_vars or None,
            exit_code=exit_code,
            limit=settings.history.limit_per_command,
        )

    def collect_missing_vars(
        self, command_alias: str, runtime_vars: dict[str, str] | None = None
    ) -> list[str]:
        cmd = self._repo.get_by_alias(command_alias)
        missing = self._resolver.collect_missing_vars(
            cmd.template, runtime_vars=runtime_vars
        )
        return missing

    def build_context(
        self, cmd: Command, runtime_ctx: RunContext | None
    ) -> RunContext | None:
        """
        Constructs a runtime context by merging command-level settings with runtime-level
        settings. Resolves priority based on input from `runtime_ctx` when available, or
        falls back to `cmd` for default values. If no significant configuration changes
        are detected, returns `None`.

        Args:
            cmd (Command): The command object containing default execution settings, including
                environment variables, working directory, shell, timeout, and more.
            runtime_ctx (RunContext | None): An optional runtime context object that may override
                specific command-level settings, such as environment variables and capture
                preferences.

        Returns:
            RunContext | None: Returns a `RunContext` object with the resulting configuration
            or `None` if no significant changes are produced by the merge.
        """
        stored_env = json.loads(cmd.env) if cmd.env else {}
        runtime_env = getattr(runtime_ctx, "env", None) or {}
        merged_env = {**stored_env, **runtime_env} or None

        cwd = (
            runtime_ctx.cwd if runtime_ctx and runtime_ctx.cwd is not None else None
        ) or cmd.cwd
        shell = (
            runtime_ctx.shell if runtime_ctx and runtime_ctx.shell is not None else None
        ) or cmd.shell
        timeout = (
            runtime_ctx.timeout
            if (runtime_ctx and runtime_ctx.timeout is not None)
            else cmd.timeout
        )

        capture = runtime_ctx.capture if runtime_ctx else False
        emit = runtime_ctx.emit if runtime_ctx else False
        verbose = runtime_ctx.verbose if runtime_ctx else False

        if not any([cwd, shell, merged_env, timeout, capture, emit, verbose]):
            return None

        return RunContext(
            cwd=cwd,
            shell=shell,
            env=merged_env,
            timeout=timeout,
            capture=capture,
            emit=emit,
            verbose=verbose,
        )
