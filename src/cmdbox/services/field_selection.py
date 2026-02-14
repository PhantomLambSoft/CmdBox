from typing import Mapping, Sequence
from dataclasses import dataclass

from cmdbox.services.errors import EmptyFieldSelectionError, UnknownFieldError


@dataclass(frozen=True)
class FieldSelectionResolver:
    """
    Resolves and validates field selections based on allowed fields, aliases, and additional rules.

    This class is designed to handle validation and resolution of a set of field selections provided
    by the user. It allows for enforcement of constraints like allowed fields, support for field aliases,
    removal of duplicates, and handling special tokens like "all". You can use it to process raw field
    selections into a sanitized and verified list of valid field names.

    Attributes:
        allowed_fields (list[str]): A list of fields that are explicitly allowed. Any field not
            present in this list will result in a validation error.
        aliases (Mapping[str, str] | None): A mapping of alias names to their corresponding target
            field names. Aliases allow users to refer to fields using alternate identifiers.
        allow_duplicates (bool): A flag indicating whether duplicate fields are permitted in the
            final resolved list. Defaults to False.
        all_token (str): A special token used to indicate that all allowed fields should be selected.
            Defaults to "all".
    """

    allowed_fields: list[str]
    aliases: Mapping[str, str] | None = None
    allow_duplicates: bool = False
    all_token: str = "all"

    def resolve(
        self,
        raw: Sequence[str] | None,
        *,
        default_fields: Sequence[str] | None = None,
        context: str | None = None,
    ) -> list[str]:
        """
        Resolves and validates a list of field names based on the provided raw input, default
        fields, and context. This method processes the field input, checks for special tokens,
        and ensures the fields comply with the allowed fields defined in the current instance.

        Args:
            raw (Sequence[str] | None): A sequence of raw field names provided for resolution.
                If None, the `default_fields` argument or allowed fields will be used as a
                fallback.
            default_fields (Sequence[str] | None, optional): A sequence of default field names
                to use if `raw` is None. If omitted or None, all allowed fields are used.
            context (str | None, optional): An optional string providing additional contextual
                information for validation or error handling.

        Returns:
            list[str]: A list of validated and resolved field names, taking into account allowed
            fields, special tokens, and fallback mechanisms.

        Raises:
            EmptyFieldSelectionError: If `raw` contains no valid field tokens or is equivalent
            to an empty selection within the specified context.
        """
        if not self.allowed_fields:
            return []

        if raw is None:
            fields = (
                list(default_fields)
                if default_fields is not None
                else self.allowed_fields
            )
            return self.validate(fields, context)

        tokens = [x.strip() for x in raw if x.strip()]

        if not tokens:
            raise EmptyFieldSelectionError(context=context)

        if any(x.lower() == self.all_token.lower() for x in tokens):
            return self.allowed_fields

        return self.validate(tokens, context)

    def validate(self, tokens: Sequence[str], context: str | None) -> list[str]:
        """
        Validates a list of tokens against allowed fields and applies transformations.

        This method processes a list of tokens by applying transformations, validating
        their presence in the allowed fields, and handling duplicate entries based on
        the configuration settings. If a token is not allowed, an exception is raised.

        Args:
            tokens (Sequence[str]): A sequence of string tokens to validate.
            context (str | None): Optional context for error reporting.

        Raises:
            UnknownFieldError: If a token is not found in the set of allowed fields.

        Returns:
            list[str]: A list of validated and transformed tokens.
        """
        out: list[str] = []
        seen: set[str] = set()

        for token in tokens:
            token = self.apply_alias(token)
            key = token.lower()

            if key not in self.allowed_fields:
                raise UnknownFieldError(token, self.allowed_fields, context=context)

            if not self.allow_duplicates and key in seen:
                continue

            seen.add(key)
            out.append(token)

        return out

    def apply_alias(self, token: str) -> str:
        """
        Applies alias substitution to a given token.

        This method checks if the given token matches any alias in the alias mapping.
        If a match is found (case-insensitive), it substitutes the token with the
        corresponding target value. If no match is found or the alias list is empty,
        the original token is returned unchanged.

        Args:
            token (str): The token to check against the alias mapping.

        Returns:
            str: The substituted token if an alias match is found; otherwise, the
            original token.
        """
        if not self.aliases:
            return token

        for alias, target in self.aliases.items():
            if alias.lower() == token.lower():
                return target

        return token
