from typing import Sequence

from cmdbox.exceptions import CmdboxError


class FieldSelectionError(CmdboxError):
    pass


class UnknownFieldError(FieldSelectionError):

    def __init__(
        self, unknown: str, allowed: Sequence[str], context: str | None = None
    ):
        self.unknown = unknown
        self.allowed = allowed
        self.context = context
        msg = f'Unknown field "{unknown}".'
        if context:
            msg += f" ({context})"
        msg += f" Allowed fields: {', '.join(allowed)}"
        super().__init__(msg)


class EmptyFieldSelectionError(FieldSelectionError):

    def __init__(self, context: str | None = None):
        msg = "No fields specified"
        if context:
            msg += f" ({context})"
        super().__init__(msg)


class HistoryIndexError(CmdboxError):

    def __init__(self, index: int):
        super().__init__(f"No history entry at index {index}.")


class ImportValidationError(CmdboxError):
    """Raised when an import fails validation before any writes occur"""

    pass


class ImportCycleError(ImportValidationError):
    """Raised when import data contains a circular reference."""

    def __init__(self, cycle: list[str]):
        self.cycle = cycle
        super().__init__(f"Circular reference detected: {' -> '.join(cycle)}")
