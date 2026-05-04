from typing import Protocol, Optional

from .type_defs import CommandRecord, VariableRecord
from ..repositories.command_repository import CommandRepository
from ..repositories.variable_repository import VariableRepository


class ResolverLookup(Protocol):
    """
    Protocol for resolving commands and variables.

    This protocol defines the interface for looking up command and variable
    definitions based on their respective identifiers. It serves as a contract
    for implementing classes to provide resolution mechanisms for commands and
    variables. Implementers of this protocol must define the behavior for
    retrieving both command and variable records, ensuring consistency and
    reliability across different resolutions.

    Methods:
        get_command(alias: str) -> Optional[CommandRecord]:
            Retrieves the CommandRecord associated with the given alias, if it
            exists.

        get_variable(name: str) -> Optional[VariableRecord]:
            Retrieves the VariableRecord associated with the given name, if it
            exists.
    """

    def get_command(self, alias: str) -> Optional[CommandRecord]:
        pass

    def get_variable(self, name: str) -> Optional[VariableRecord]:
        pass


class RepoLookup(ResolverLookup):
    """
    Provides lookup functionality for commands and variables stored in repositories.

    This class serves as an adapter for resolving commands and variables from their
    respective repositories. Command and variable records can be retrieved based on
    aliases or names respectively. It is intended to abstract the underlying repository
    interaction by providing an easy interface for lookups.

    Attributes:
        cmd_repo (CommandRepository): Repository for storing and retrieving command records.
        var_repo (VariableRepository): Repository for managing and accessing variable records.
    """

    def __init__(self, cmd_repo: CommandRepository, var_repo: VariableRepository):
        self._cmd_repo = cmd_repo
        self._var_repo = var_repo

    def get_command(self, alias: str) -> Optional[CommandRecord]:
        cmd = self._cmd_repo.get_by_alias(alias)
        if cmd is None:
            return None
        return CommandRecord(alias=cmd.alias, template=cmd.template)

    def get_variable(self, name: str) -> Optional[VariableRecord]:
        var = self._var_repo.get_by_name(name)
        if var is None:
            return None
        return VariableRecord(name=var.name, value=var.value)


class MemoizedLookup(ResolverLookup):
    """
    Caches and retrieves command and variable lookups for improved performance.

    The MemoizedLookup class acts as a wrapper for a ResolverLookup instance,
    adding caching functionality to minimize redundant lookups. Commands and
    variables are cached after their first retrieval, improving performance
    for subsequent requests. Use this class when frequent lookups are
    expected and caching them can significantly enhance speed.

    Attributes:
        inner (ResolverLookup): The wrapped ResolverLookup instance used for
            retrieving commands and variables.
        cmd_cache (dict): A dictionary used to cache CommandRecord results
            for quick retrieval based on their aliases.
        var_cache (dict): A dictionary used to cache VariableRecord results
            for quick retrieval based on their names.
    """

    def __init__(self, inner: ResolverLookup):
        self._inner = inner
        self._cmd_cache: dict[str, Optional[CommandRecord]] = {}
        self._var_cache: dict[str, Optional[VariableRecord]] = {}

    def get_command(self, alias: str) -> Optional[CommandRecord]:
        """
        Retrieves a command by its alias, leveraging the cache for faster access.

        This method checks if the command associated with the given alias exists
        in the cache. If found, it retrieves the command from the cache. If not,
        it fetches the command from an internal source, stores it in the cache,
        and then returns it.

        Args:
            alias (str): The alias of the command to retrieve.

        Returns:
            Optional[CommandRecord]: The command associated with the alias if
                                     it exists, otherwise None.
        """
        if alias in self._cmd_cache:
            return self._cmd_cache[alias]
        cmd = self._inner.get_command(alias)
        self._cmd_cache[alias] = cmd
        return cmd

    def get_variable(self, name: str) -> Optional[VariableRecord]:
        """
        Retrieves a variable by its alias, leveraging the cache for faster access.

        This method checks if the variable associated with the given name exists
        in the cache. If found, it retrieves the variable from the cache. If not,
        it fetches the variable from an internal source, stores it in the cache,
        and then returns it.

        Args:
            name (str): The name of the variable to retrieve.

        Returns:
            Optional[VariableRecord]: The variable associated with the alias if
                                     it exists, otherwise None.
        """
        if name in self._var_cache:
            return self._var_cache[name]
        var = self._inner.get_variable(name)
        self._var_cache[name] = var
        return var

    def clear(self) -> None:
        self._cmd_cache.clear()
        self._var_cache.clear()
