from cmdbox.repositories.command_repository import CommandRepository
from cmdbox.resolve.resolver import Resolver
from cmdbox.resolve.types import ResolveResult
from cmdbox.runtime.executor import Executor
from cmdbox.runtime.results import ExecutionResult


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
    ) -> None:
        self._repo = repo
        self._resolver = resolver
        self._executor = executor

    def run(self, command_alias: str) -> ExecutionResult:
        """
        Executes a command based on the given alias.

        This method retrieves a command associated with the provided alias, resolves its
        template using a resolver, and then executes the resolved command text.

        Args:
            command_alias (str): The alias for the command to be executed.

        Returns:
            ExecutionResult: The result of executing the command.
        """
        cmd = self._repo.get_by_alias(command_alias)
        resolved_cmd = self._resolver.resolve(cmd.template)
        return self._executor.run(resolved_cmd.text)

    def preview(self, command_alias: str) -> ResolveResult:
        """
        Retrieves and resolves a command template based on its alias.

        This method takes a command alias, retrieves the associated command template,
        and resolves it to produce a complete command configuration or data structure.
        This resolved template is then returned as a `ResolveResult`.

        Args:
            command_alias (str): The alias of the command to be resolved.

        Returns:
            ResolveResult: The resolved result of the command template.
        """
        cmd = self._repo.get_by_alias(command_alias)
        return self._resolver.resolve(cmd.template)
