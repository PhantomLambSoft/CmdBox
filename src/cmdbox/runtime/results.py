from dataclasses import dataclass


@dataclass(frozen=True)
class ExecutionResult:
    """
    Represents the result of a command execution.

    This class is a data container that holds details of a command execution
    outcome, including the executed command, its exit status, and associated
    standard output and error streams. It is designed to provide a structured
    result of running an external command in a consistent and immutable format.

    Attributes:
        command (str): The actual command that was executed. Extrapolated
            from the Command.template.
        exit_code (int): The exit code returned by the command.
        stdout (str): The standard output produced by the command.
        stderr (str): The standard error output produced by the command.
    """

    command: str
    exit_code: int
    stdout: str
    stderr: str
