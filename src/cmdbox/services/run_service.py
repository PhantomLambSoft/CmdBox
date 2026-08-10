import json
from typing import Callable

from cmdbox.models import Command, Profile
from cmdbox.repositories.command_repository import CommandRepository
from cmdbox.repositories.history_repository import HistoryRepository
from cmdbox.repositories.profile_repository import ProfileRepository
from cmdbox.resolve.resolver import Resolver
from cmdbox.resolve.type_defs import ResolveResult, RefKind
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
        profile_repo: ProfileRepository,
        history_repo: HistoryRepository | None = None,
        get_settings: Callable[[], Settings] | None = None,
    ) -> None:
        self._repo = repo
        self._resolver = resolver
        self._executor = executor
        self._profile_repo = profile_repo
        self._history_repo = history_repo
        self._get_settings = get_settings

    def _resolve_profile(self, profile: str | None) -> Profile | None:
        return self._profile_repo.get_by_name(profile) if profile else None

    def run(
        self,
        command_alias: str,
        ctx: RunContext | None = None,
        runtime_vars: dict[str, str] | None = None,
        profile: str | None = None,
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
            profile (str | None): The name of the profile to use for command execution.

        Returns:
            ExecutionResult: The result of executing the command.
        """
        resolved_profile = self._resolve_profile(profile)
        cmd = self._repo.get_by_alias(command_alias, profile=resolved_profile)
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
            profile=resolved_profile,
        )
        self.record_variable_profile_use(resolved_cmd)
        return result

    def preview(
        self,
        command_alias: str,
        runtime_vars: dict[str, str] | None = None,
        ctx: RunContext | None = None,
        profile: str | None = None,
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
            profile (str | None): The name of the profile to be used for command resolution. Defaults to None.

        Returns:
            tuple[ResolveResult, RunContext | None]: The resolved result of the command template and the
            effective context.
        """
        resolved_profile = self._resolve_profile(profile)
        cmd = self._repo.get_by_alias(command_alias, profile=resolved_profile)
        resolved = self._resolver.resolve(cmd.template, runtime_vars=runtime_vars)
        effective_ctx = self.build_context(cmd, ctx)
        return resolved, effective_ctx

    def record_history(
        self,
        *,
        alias: str,
        template: str,
        resolved: str,
        profile: Profile | None,
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
            profile (Profile): The profile used for command execution.
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
            profile=profile,
        )

    def record_variable_profile_use(self, resolved_cmd: ResolveResult) -> None:
        """
        Records the usage of a variable profile if a stored variable is detected in
        the resolution trace of the provided command.

        This method inspects the trace of the provided `resolved_cmd` to determine
        if a variable with a reference kind of `VARIABLE` and a source of `stored`
        was used. If such a variable is found, the method retrieves the active
        variable profile ID from the profile repository and records its usage.

        Args:
            resolved_cmd: The resolution result containing the trace of steps
                executed and their metadata.

        """
        used_stored_variable = any(
            step.kind == RefKind.VARIABLE and step.source == "stored"
            for step in resolved_cmd.trace
        )
        if used_stored_variable:
            active_variable_profile_id = self._profile_repo.get_state().active_variable_profile_id
            self._profile_repo.record_use(active_variable_profile_id)

    def collect_missing_vars(
        self,
        command_alias: str,
        runtime_vars: dict[str, str] | None = None,
        profile: str | None = None,
    ) -> list[str]:
        resolved_profile = self._resolve_profile(profile)
        cmd = self._repo.get_by_alias(command_alias, profile=resolved_profile)
        missing = self._resolver.collect_missing_vars(
            cmd.template, runtime_vars=runtime_vars
        )
        return missing

    def build_context(self, cmd: Command, runtime_ctx: RunContext | None = None) -> RunContext | None:
        """
        Builds and returns a runtime execution context by consolidating information from the command and
        the current runtime context.

        The method integrates the environment variables, working directory, shell, timeout, output
        capture settings, verbosity, and other runtime parameters to form a comprehensive context. If
        certain properties are not provided in the runtime context, it falls back to values specified
        in the command or default application settings.

        Args:
            cmd (Command): The command object containing configuration and parameters for execution.
            runtime_ctx (RunContext | None): The optional runtime context that can override specific
                properties of the command settings.

        Returns:
            RunContext | None: A consolidated execution context containing the merged settings to be
                used during command execution.
        """
        stored_env = json.loads(cmd.env) if cmd.env else {}
        runtime_env = getattr(runtime_ctx, "env", None) or {}
        merged_env = {**stored_env, **runtime_env} or None

        cwd = (
          runtime_ctx.cwd if runtime_ctx and runtime_ctx.cwd is not None else None
        ) or cmd.cwd

        settings = self._get_settings()

        if runtime_ctx and runtime_ctx.shell is not None:
            shell = runtime_ctx.shell
        elif cmd.shell is not None:
            shell = cmd.shell
        else:
            shell = settings.execution_settings.default_shell

        if runtime_ctx and runtime_ctx.timeout is not None:
            timeout = runtime_ctx.timeout
        else:
            timeout = cmd.timeout

        if runtime_ctx and runtime_ctx.capture is not None:
            capture = runtime_ctx.capture
        else:
            capture = settings.execution_settings.capture_output

        if runtime_ctx and runtime_ctx.verbose is not None:
            verbose = runtime_ctx.verbose
        else:
            verbose = settings.execution_settings.default_verbose

        emit = runtime_ctx.emit if runtime_ctx else False

        return RunContext(
            cwd=cwd,
            shell=shell,
            env=merged_env,
            timeout=timeout,
            capture=capture,
            emit=emit,
            verbose=verbose,
        )
