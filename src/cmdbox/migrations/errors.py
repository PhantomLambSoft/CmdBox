from cmdbox.exceptions import CmdboxError


class MigrationError(CmdboxError):
    """
    Exception raised when an error occurs during database migration.
    """

    def __init__(self, message: str) -> None:
        super().__init__(message)
