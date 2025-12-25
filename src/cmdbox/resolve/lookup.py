from typing import Protocol, Optional

from cmdbox.resolve.types import CommandRecord, VariableRecord


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