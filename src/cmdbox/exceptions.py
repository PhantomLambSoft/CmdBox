class CmdboxError(Exception):
    """
    Base class for exceptions in the cmdbox module.

    This custom exception class serves as a base for all exceptions raised
    within the cmdbox module. It provides a means to handle module-specific
    errors in a structured manner.
    """

    pass


class UnknownAlias(CmdboxError):
    """
    Exception raised when an unknown alias is encountered.

    This error is raised when a requested alias is not found within the
    storage system.

    Attributes:
        alias (str): The alias that was not found.
    """

    def __init__(self, alias: str) -> None:
        super().__init__(f"Alias {alias} not found.")


class AliasConflict(CmdboxError):
    """
    Indicates a conflict caused by an already existing alias.

    This exception is raised when attempting to create or use an alias that
    already exists.

    Attributes:
        alias (str): The alias string that caused the conflict.
    """

    def __init__(self, alias: str) -> None:
        super().__init__(f"Alias {alias} already exists.")


class UnknownVariable(CmdboxError):
    """
    Exception raised for accessing an unknown variable.

    This exception is raised when a requested variable is not found in
    the storage system.

    Attributes:
        variable (str): The name of the variable that was not found.
    """

    def __init__(self, variable: str) -> None:
        super().__init__(f"Variable {variable} not found.")


class VariableConflict(CmdboxError):
    """
    Exception caused by an already existing variable.

    This exception is raised when attempting to create or use a variable that
    already exists.

    Attributes:
        variable (str): The name of the variable that caused the conflict.
    """

    def __init__(self, variable: str) -> None:
        super().__init__(f"Variable {variable} already exists.")


class ResolutionError(CmdboxError):
    """
    Represents an error encountered during resolution processes.

    This exception is raised when an error is encountered during the command
    template resolution process.  For example: a circular reference in which
    command A references command B, which references command A.
    """

    pass


class CommandSyntaxError(CmdboxError):
    """
    Represents an error related to command syntax.

    This class is used to handle exceptions that occur due to syntax errors
    in command parsing or execution.
    """

    pass
