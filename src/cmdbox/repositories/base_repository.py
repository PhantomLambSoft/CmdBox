from typing import Sequence, Generic, TypeVar, Type

from peewee import Model, fn, IntegrityError, Node

from cmdbox.exceptions import UnknownTagError
from cmdbox.models import Tag

M = TypeVar("M", bound=Model)


class BaseRepository(Generic[M]):
    model: Type[M]

    def _search(
        self,
        query: str,
        secondary_ordering: str,
        fields: str | Sequence[str] | None = None,
    ) -> list[M]:
        """
        Searches for commands matching the given query across specified fields.

        This function performs a case-insensitive search for the query in the specified
        fields of the Command model. It calculates a relevance score for each match
        based on the position of the query within the fields and sorts the results by
        relevance.  Relevance is determined by a weighted score with the highest weight
        given to most occurrences and the second part of the score given to how early
        the search term occurs in the field.

        Args:
            query (str): The search query to match in the specified fields.
            fields (str | Sequence[str] | None): The fields to search within. By
                default, searches within "name" and "description". Can be a single field
                name as a string or a sequence of field names.

        Returns:
            list[Command]: A list of Command objects matching the search query, sorted
                by relevance.

        Raises:
            ValueError: If any provided field does not exist on the Command model.
        """
        if not query:
            return []
        if isinstance(fields, str):
            fields = self._get_sequence(fields)
        if len(fields) == 0:
            return []
        or_clauses = []
        relevance_parts = []
        query_lower = query.lower()
        for field_name in fields:
            if not hasattr(self.model, field_name) or field_name == "":
                raise ValueError(f"Invalid field: {field_name}")
            field = getattr(self.model, field_name)
            or_clauses.append(fn.LOWER(field).contains(query_lower))
            first_pos = fn.INSTR(fn.lower(field), query_lower)
            occurrences = (
                fn.LENGTH(field) - fn.LENGTH(fn.REPLACE(field, query_lower, ""))
            ) / len(query_lower)
            score = (occurrences * 1000) - first_pos
            relevance_parts.append(score)

        condition = or_clauses[0]
        for clause in or_clauses[1:]:
            condition |= clause

        if len(relevance_parts) > 1:
            relevance = fn.MIN(*relevance_parts).alias("relevance")
        else:
            relevance = relevance_parts[0].alias("relevance")

        query_obj = (
            self.model.select(self.model, relevance)
            .where(condition)
            .order_by(
                relevance.desc(),
                getattr(self.model, secondary_ordering),
            )
        )
        return list(query_obj)

    def _resolve_order_token(self, token: str) -> Node:
        """
        Resolves the ordering token for determining the sorting order of a field.

        This method takes an ordering token, identifies whether the sorting is
        ascending or descending, and ensures the token corresponds to a valid
        model attribute for sorting.

        Args:
            token (str): The ordering token indicating the field for sorting. A
                token starting with '-' indicates descending sort, otherwise it
                is ascending. The token must correspond to a valid attribute
                in `self.model`.

        Returns:
            Model: An object representing the ordering of the specified field
                in ascending or descending order.

        Raises:
            ValueError: If the token is empty or if it references an invalid
                field in the model.
        """
        if not token:
            raise ValueError("Empty order_by token.")
        desc = token.startswith("-")
        field_name = token[1:] if desc else token
        try:
            field = getattr(self.model, field_name)
        except AttributeError:
            raise ValueError(f"Invalid order_by field: {token}")
        return field.desc() if desc else field.asc()

    def _resolve_ordering(self, order_by: str | Sequence[str]) -> Sequence[Node]:
        """
        Resolves ordering tokens based on the specified order_by input.

        This method processes the input order_by parameter, which could either
        be a single string of comma-separated tokens or a sequence of strings.
        It ensures the tokens are parsed and resolved appropriately, returning a
        sequence of resolved models.

        Args:
            order_by: A string containing comma-separated tokens, or a sequence
                of strings representing ordering keys.

        Returns:
            Sequence[Model]: A sequence of resolved models based on the specified
            ordering tokens.

        Raises:
            ValueError: If the order_by sequence is empty.
        """
        tokens = self._get_sequence(order_by)
        if not tokens:
            raise ValueError("Empty order_by sequence.")
        return [self._resolve_order_token(token) for token in tokens]

    def _get_sequence(self, items: str | Sequence[str]) -> list[str]:
        """
        Processes a string or sequence of strings into a list of stripped strings.

        If the input is a single string, it is split by commas, and whitespace around each
        split item is stripped. If the input is already a sequence of strings, it is converted
        into a list without any modifications.

        Args:
            items (str | Sequence[str]): A string to be split into a list of strings or an
                existing sequence of strings.

        Returns:
            list[str]: A list of processed strings.

        Raises:
            ValueError: If the input is neither a string nor a sequence of strings.
        """
        if isinstance(items, str):
            return [item.strip() for item in items.split(",") if item.strip()]
        else:
            return [item.strip() for item in items if item.strip()]

    def _is_unique_name_violation(self, exc: IntegrityError) -> bool:
        """
        Checks if the exception indicates a unique constraint violation.

        This method analyzes the provided IntegrityError to determine if the
        error was caused by a UNIQUE constraint violation on a field of the
        database table associated with the model.

        Args:
            exc (IntegrityError): The IntegrityError exception to be checked.

        Returns:
            bool: True if the exception indicates a unique constraint violation;
            otherwise, False.
        """
        msg = str(exc)
        table = self.model._meta.table_name
        return "UNIQUE constraint failed" in msg and f"{table}.name" in msg

    def _get_tags_by_name(self, *tags: str) -> list[Tag]:
        """
        Retrieves tags by their names and raises an error if any tag names do not exist.

        Args:
            *tags (str): Variable length argument list containing tag names to be retrieved.

        Returns:
            list[Tag]: A list of Tag objects matching the provided tag names.

        Raises:
            UnknownTagError: If one or more tag names provided in the arguments do not exist.
        """
        if not tags:
            return []
        queryset = Tag.select().where(Tag.name << tags)
        tags_by_name = {t.name for t in queryset}
        missing = [name for name in tags if name not in tags_by_name]
        if missing:
            missing_names = ", ".join(missing)
            raise UnknownTagError(tag_name=missing_names)
        return list(queryset)
