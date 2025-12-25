class CmdboxError(Exception):
    """
    Base class for exceptions in the cmdbox module.

    This custom exception class serves as a base for all exceptions raised
    within the cmdbox module. It provides a means to handle module-specific
    errors in a structured manner.
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
