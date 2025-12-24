class CmdboxError(Exception):
    """
    Base class for exceptions in the cmdbox module.

    This custom exception class serves as a base for all exceptions raised
    within the cmdbox module. It provides a means to handle module-specific
    errors in a structured manner.
    """

    pass


class UnknownAliasError(CmdboxError):
    """
    Exception raised when an unknown alias is encountered.

    This error is raised when a requested alias is not found within the
    storage system.

    Attributes:
        alias (str): The alias that was not found.
    """

    def __init__(self, alias: str) -> None:
        super().__init__(f"Alias '{alias}' not found.")


class AliasConflictError(CmdboxError):
    """
    Indicates a conflict caused by an already existing alias.

    This exception is raised when attempting to create or use an alias that
    already exists.

    Attributes:
        alias (str): The alias string that caused the conflict.
    """

    def __init__(self, alias: str) -> None:
        super().__init__(f"Alias '{alias}' already exists.")


class UnknownNameError(CmdboxError):
    """
    Exception raised for accessing an unknown variable.

    This exception is raised when a requested variable is not found in
    the storage system.

    Attributes:
        variable (str): The name of the variable that was not found.
    """

    def __init__(self, name: str) -> None:
        super().__init__(f"Variable name '{name}' not found.")


class NameConflictError(CmdboxError):
    """
    Exception caused by an already existing variable.

    This exception is raised when attempting to create or use a variable that
    already exists.

    Attributes:
        name (str): The name of the variable that caused the conflict.
    """

    def __init__(self, name: str) -> None:
        super().__init__(f"Name '{name}' already exists.")


class UnknownTagError(CmdboxError):

    def __init__(self, tag_name: str) -> None:
        super().__init__(f"Tag '{tag_name}' not found.")


class TagAttachError(CmdboxError):

    pass


class TagDetachError(CmdboxError):
    pass


class ValidationError(CmdboxError):
    """
    Represents an error related to invalid data.

    This class is used to handle exceptions that occur due to invalid data.
    Invalid data is data that technically can be stored in the database, but
    is against the use principles allowed by the application.
    ex: Creating a command with an empty alias or template.
    """

    pass


class ResolutionError(CmdboxError):
    """
    Represents an error encountered during resolution processes.

    This exception is raised when an error is encountered during the command
    template resolution process.  For example: a circular reference in which
    command A references command B, which references command A.
    """

    pass


class CommandSyntaxError(ResolutionError):
    """
    Represents an error related to command syntax.

    This class is used to handle exceptions that occur due to syntax errors
    in command parsing or execution.
    """

    pass


class UnknownReference(ResolutionError):
    """
    Represents an error related to references in a string being resolved.

    This class is used to handle exceptions when a reference is stored in a
    string being resolved, but the reference cannot be found in the database.
    """

    def __init__(self, kind: str, key: str):
        super().__init__(f"Unknown {kind}: {key}")
        self.kind = kind
        self.key = key


class MaxDepthExceeded(ResolutionError):
    """
    Exception raised when the maximum depth is exceeded.

    Represents a specific error condition where a resolution process exceeds
    the allowed or configured maximum depth. This exception is typically used
    to prevent excessively deep recursion by enforcing depth limits.
    """

    def __init__(self, max_depth: int):
        super().__init__(f"Maximum resolution depth exceeded: ({max_depth})")
        self.max_depth = max_depth


class CycleDetectionError(ResolutionError):
    """
    An error raised when a circular reference is detected.

    Raised when a circular reference is detected while attempting to resolve
    a string.  Ex: Command A references command B, which references command A.
    """

    def __init__(self, path: list[str]):
        super().__init__(f"Cycle detected: {' -> '.join(path)}")
        self.path = path
