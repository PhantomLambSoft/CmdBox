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
