from cmdbox.exceptions import CmdboxError


class UnknownCommandError(CmdboxError):

    def __init__(self, cmd_id: int) -> None:
        super().__init__(f"Command with ID '{cmd_id}' not found.")


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


class UnknownVariableError(CmdboxError):

    def __init__(self, var_id: int) -> None:
        super().__init__(f"Variable with ID '{var_id}' not found.")


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


class UpdateError(CmdboxError):

    def __init__(self, message: str | None = None) -> None:
        super().__init__(f"Failed to update: {message}")


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


class UnknownHistoryEntryError(CmdboxError):

    def __init__(self, ref: str):
        super().__init__(f"No history entry found matching '{ref}'.")


class AmbiguousHistoryIdEntryError(CmdboxError):

    def __init__(self, prefix: str):
        super().__init__(
            f"ID prefix '{prefix}' matches multiple history entries. Use more ID characters."
        )


class UnknownProfileError(CmdboxError):

    """
    Exception raised when attempting to switch to a profile that cannot be found
    by name in the database.

    Attributes:
        name (str): The name of the profile that could not be found.
    """

    def __init__(self, name: str) -> None:
        super().__init__(f"Profile '{name}' not found.")


class ProfileConflictError(CmdboxError):

    """
    Exception raised when attempting to create a profile with a name that
    already exists.

    Attributes:
        name (str): The name of the profile that caused the conflict.
    """

    def __init__(self, name: str) -> None:
        super().__init__(f"Profile '{name}' already exists.")


class ProfileNotEmptyError(CmdboxError):

    """
    Exception raised when attempting to delete a profile that still contains
    commands or variables without using '--force'.

    Attributes:
        name (str): The name of the profile that is not empty.
        command_count (int): The number of commands in the profile.
        variable_count (int): The number of variables in the profile.
    """

    def __init__(self, name: str, command_count: int, variable_count: int) -> None:
        super().__init__(
            f"Profile '{name}' still has {command_count} command(s) and "
            f"{variable_count} variable(s). Use --force to delete them along "
            f"with the profile."
        )


class ActiveProfileDeleteError(CmdboxError):

    """
    Exception raised when attempting to delete a profile that is currently active for
    commands, variables, or settings.

    Attributes:
        name (str): The name of the profile that is currently active.
    """

    def __init__(self, name: str) -> None:
        super().__init__(
            f"Profile '{name}' is currently active and cannot be deleted. "
            f"Switch to a different profile first."
        )
