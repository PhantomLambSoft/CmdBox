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
        allow_duplicates (bool): A flag indicating whether duplicate fields are permitted in the
            final resolved list. Defaults to False.
        all_token (str): A special token used to indicate that all allowed fields should be selected.
            Defaults to "all".
    """

    allowed_fields: list[str]
    allow_duplicates: bool = False
    all_token: str = "all"

    @property
    def _allowed_fields(self):
        """
        Property method that retrieves a list of allowed fields in lowercase.

        This method iterates through the fields defined in the `allowed_fields`
        attribute, converts each field to lowercase, and returns the updated list.

        Returns:
            list: A list of strings representing the fields converted to lowercase.
        """
        return [x.lower() for x in self.allowed_fields]

    def resolve(
        self,
        raw: Sequence[str] | None,
        *,
        default_fields: Sequence[str] | None = None,
        aliases: Mapping[str, str] | None = None,
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
            aliases: (Mapping[str, str] | None, optional): A dictionary mapping field aliases to
                a corresponding field name. If provided, aliases will be applied to the raw input.
            context (str | None, optional): An optional string providing additional contextual
                information for validation or error handling.

        Returns:
            list[str]: A list of validated and resolved field names, taking into account allowed
            fields, special tokens, and fallback mechanisms.

        Raises:
            EmptyFieldSelectionError: If `raw` contains no valid field tokens or is equivalent
            to an empty selection within the specified context.
        """
        if not self._allowed_fields:
            return []

        if raw is None:
            fields = (
                list(default_fields)
                if default_fields is not None
                else self._allowed_fields
            )
            return self.validate(fields, context)

        raw = [x.lower() for x in raw]
        tokens = [self.apply_alias(x, aliases) for x in raw if x.strip()]

        if not tokens:
            raise EmptyFieldSelectionError(context=context)

        if any(x.lower() == self.all_token.lower() for x in tokens):
            return self._allowed_fields

        return self.validate(tokens, context)

    def validate(self, tokens: Sequence[str], context: str | None) -> list[str]:
        """
        Validates a sequence of tokens based on allowed fields and duplicate rules.

        This method processes the provided tokens to ensure they are within the set of
        allowed fields. Optionally, it filters out duplicate tokens based on the
        `allow_duplicates` rule. If any token does not match the allowed fields, an
        `UnknownFieldError` is raised.

        Args:
            tokens (Sequence[str]): A sequence of tokens to validate.
            context (str | None): An optional context string providing additional
                information for debug or error messages.

        Returns:
            list[str]: A list of validated tokens, preserving the original order and
            optionally excluding duplicates.

        Raises:
            UnknownFieldError: If a token is not in the allowed fields or violates
            validation rules.
        """
        out: list[str] = []
        seen: set[str] = set()

        for token in tokens:
            key = token.lower()

            if key not in self._allowed_fields:
                raise UnknownFieldError(token, self._allowed_fields, context=context)

            if not self.allow_duplicates and key in seen:
                continue

            seen.add(key)
            out.append(token)

        return out

    def apply_alias(self, token: str, aliases: Mapping[str, str] | None) -> str:
        """
        Processes a token and applies alias transformations if a matching alias is found.

        This method checks if a given token matches any of the keys in the provided alias
        mappings. If a match is found, the token is replaced with the associated value
        from the mapping. If no match is found or no aliases are provided, the original
        token is returned unmodified.

        Args:
            token (str): The token to be processed and potentially transformed.
            aliases (Mapping[str, str] | None): A dictionary where keys represent alias
                tokens, and values are the replacement tokens. If None is provided, no
                alias transformations will be applied.

        Returns:
            str: The transformed token if a matching alias was found, or the original
            token otherwise.
        """
        if not aliases:
            return token

        for alias, target in aliases.items():
            if alias.lower() == token.lower():
                return target

        return token
