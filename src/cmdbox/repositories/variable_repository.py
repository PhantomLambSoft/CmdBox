from typing import Sequence

from peewee import IntegrityError

from .base_repository import BaseRepository
from .errors import (
    ValidationError,
    NameConflictError,
    UnknownNameError,
    UnknownTagError,
    TagAttachError,
    TagDetachError,
)
from .validators import VariableValidator
from .results import TagAttachResult, TagDetachResult
from cmdbox.database import db
from cmdbox.models import Variable, Tag, VariableTag


class VariableRepository(BaseRepository[Variable]):
    model = Variable

    def __init__(self, validator: VariableValidator | None = None):
        self.validator = validator or VariableValidator()

    def create(self, name: str, value: str) -> Variable:
        """
        Validates and creates a new Variable object based on provided input parameters.

        Args:
            name (str): Unique identifier for the variable to be created.
            value (str): The value that will be subbed for the name when executing.

        Returns:
            Variable: The created Variable object.

        Raises:
            NameConflictError: If a variable with the provided name already exists
                and causes a unique constraint violation during creation.
            IntegrityError: If any database integrity issue occurs during the creation.
        """
        name = name.strip() if name else None
        self.validator.validate_create(name=name, value=value)
        try:
            return Variable.create(name=name, value=value)
        except IntegrityError as exc:
            if name is not None and self._is_unique_name_violation(exc):
                raise NameConflictError(name=name) from exc
            raise

    def get_by_name(self, name: str) -> Variable:
        """
        Retrieves a variable instance by its name.

        Converts the provided name to lowercase, searches for a variable with the
        matching name in the database, and retrieves it. This function raises an error
        if no variable with the specified name is found.

        Args:
            name (str): The name of the variable to retrieve.

        Returns:
            Variable: The variable object if found.

        Raises:
            UnknownNameError: If no variable with the specified name exists.
        """
        name = name.lower()
        var = Variable.get_or_none(Variable.name == name)
        if var is None:
            raise UnknownNameError(name=name)
        return var

    def update(self, var_name: str, **fields) -> Variable:
        """
        Updates an existing variable based on the provided name and fields.

        This method retrieves a variable by its name and updates its fields with the
        provided values. It validates the updates, ensuring field integrity and
        uniqueness. If the variable doesn't exist or an update violates constraints,
        appropriate actions are taken.

        Args:
            var_name: The name of the variable to update.
            **fields: Arbitrary keyword arguments representing the fields to update.
                Supported fields include 'name' and 'value'.

        Returns:
            Variable: The updated variable object if successful, or None if the
                variable is not found.

        Raises:
            ValidationError: If any of the provided fields are invalid or do not exist
                on a variable.
            NameConflictError: If the new name conflicts with an existing variable's
                name.
            IntegrityError: If there is a general integrity constraint violation during
                the update process.
        """
        var = self.get_by_name(var_name)
        if not fields:
            raise ValueError("No fields provided for update.")

        if "name" in fields and fields.get("name") is not None:
            fields["name"] = fields.get("name").strip()

        self.validator.validate_update(
            name=fields.get("name", var_name), value=fields.get("value", var.value)
        )

        try:
            for key, value in fields.items():
                if not hasattr(var, key):
                    raise ValidationError(f"Invalid field: {key}")
                if value is not None:
                    setattr(var, key, value)
            var.save()
            return var
        except IntegrityError as exc:
            name = fields.get("name", "")
            if name is not None and self._is_unique_name_violation(exc):
                raise NameConflictError(name=name) from exc
            raise

    def add_tags(self, name: str, tags: Sequence[str]) -> TagAttachResult:
        """
        Attach tags to a variable identified by its name.

        This function associates tags with a specified variable. If the tags already
        exist for the variable, they are added to an existing list. Otherwise, new tags
        are created and linked to the variable. If no tags are provided, it returns
        immediately with empty results. In case of a database integrity issue, an
        appropriate error is raised.

        Args:
            name (str): The name identifier of the variable to which the tags are to
                be attached.
            tags (Sequence[str]): A collection of tag names to be attached to the
                variable.

        Returns:
            TagAttachResult: An object containing lists of newly added tags and
                tags that already existed.

        Raises:
            TagAttachError: If there is an issue attaching the tags to the variable,
                typically due to database integrity constraints.
        """
        if not tags:
            return TagAttachResult(added=[], existing=[])
        tags_actual = self._get_tags_by_name(*tags)
        var = self.get_by_name(name)
        try:
            with db.atomic():
                added = []
                existing = []
                for tag in tags_actual:
                    var_tag, created = VariableTag.get_or_create(variable=var, tag=tag)
                    if created:
                        added.append(tag.name)
                    else:
                        existing.append(tag.name)
                return TagAttachResult(added=added, existing=existing)
        except UnknownTagError:
            raise
        except IntegrityError as exc:
            raise TagAttachError("Could not attach tags to variable.") from exc

    def remove_tags(self, name: str, tags: Sequence[str]) -> TagDetachResult:
        """
        Removes tags from a variable identified by the provided name. The method first validates
        the tags to ensure they exist in the database, then attempts to remove the associations
        between the variable and the respective tags. If a tag is not attached to the variable, it
        is recorded in the `not_attached` list, while successfully removed tags are recorded in
        the `removed` list. Any errors during detachment raise a `TagDetachError`.

        Args:
            name (str): The unique identifier or name of the variable from which the tags will
                be detached.
            tags (Sequence[str]): A list of tag names to be detached from the variable.

        Returns:
            TagDetachResult: Object that contains two lists:
                - `removed`: A list of successfully detached tags.
                - `not_attached`: A list of tags that were not associated with the variable.

        Raises:
            TagDetachError: If the detachment process encounters an issue, such as database
                integrity errors.
        """
        if not tags:
            return TagDetachResult(removed=[], not_attached=[])
        tags_actual = self._get_tags_by_name(*tags)
        variable = self.get_by_name(name)
        removed = []
        not_attached = []
        try:
            with db.atomic():
                for tag in tags_actual:
                    deleted = (
                        VariableTag.delete()
                        .where(
                            (VariableTag.variable == variable)
                            & (VariableTag.tag == tag)
                        )
                        .execute()
                    )
                    if deleted:
                        removed.append(tag.name)
                    else:
                        not_attached.append(tag.name)
        except IntegrityError as exc:
            raise TagDetachError("Could not detach tags from variable.") from exc
        return TagDetachResult(removed=removed, not_attached=not_attached)

    def list_all(
        self, order_by: str | Sequence[str] = "name", limit: int = 25
    ) -> list[Variable]:
        """
        Lists all variables from the database, optionally ordered by specified fields.

        Args:
            order_by (str | Sequence[str]): A string or sequence of strings indicating the field(s)
                by which the variables should be ordered. Defaults to "name".
            limit (int): The maximum number of variables to return. Defaults to 25.

        Returns:
            List[Variable]: A list of Variable objects retrieved from the database,
                sorted based on the specified ordering criteria.
        """
        ordering = self._resolve_ordering(order_by)
        return list(Variable.select().order_by(*ordering).limit(limit))

    def list_by_tag(
        self,
        tags: Sequence[str],
        order_by: str | Sequence[str] = "name",
        limit: int = 25,
    ) -> list[Variable]:
        """
        Fetches a list of variables filtered by specified tags and ordered by specific
        criteria.

        This method retrieves a list of `Variable` objects associated with the tags
        provided in the `tags` argument. The results are optionally ordered based on the
        `order_by` field and limited by the `limit` argument.

        Args:
            tags (Sequence[str]): A list of tag names to filter variables by.
            order_by (str | Sequence[str]): Criteria to order the resulting variable
                list. Defaults to "name".
            limit (int): The maximum number of variables to return. Defaults to 25.

        Returns:
            list[Variable]: A list of `Variable` objects matching the tags and ordered as
                requested.

        Raises:
            UnknownTagError: If one or more provided tags do not exist in the database.
        """
        tags_actual = self._get_tags_by_name(*tags)
        ordering = self._resolve_ordering(order_by)
        return list(
            Variable.select()
            .join(VariableTag)
            .where(VariableTag.tag << tags_actual)
            .order_by(*ordering)
            .distinct()
            .limit(limit)
        )

    def search(
        self, query: str, fields: str | Sequence[str] | None = "name"
    ) -> list[Variable]:
        """
        Searches for variables matching the given query across specified fields.

        This function performs a case-insensitive search for the query in the specified
        fields of the Variable model. It calculates a relevance score for each match
        based on the occurrence count and position of the query within the fields and
        sorts the results by relevance.
        Args:
            query (str): The search query to match in the specified fields.
            fields (str | Sequence[str] | None): The fields to search within. By
                default, searches within "name". Can be a single field name as a string
                or a sequence of field names.

        Returns:
            list[Variable]: A list of Variable objects matching the search query, sorted
                by relevance.

        Raises:
            ValueError: If any provided field does not exist on the Variable model.
        """
        return self._search(query, "name", fields)

    def delete(self, name: str) -> bool:
        """
        Deletes the variable with the specified name.

        Args:
            name (str): The name of the variable to delete.

        Returns:
            bool: True if the variable was deleted, False otherwise.
        """
        var = self.get_by_name(name)
        if not var:
            return False
        var.delete_instance()
        return True

    def _is_unique_variable_tag_violation(self, exc: IntegrityError) -> bool:
        msg = str(exc)
        return (
            "UNIQUE constraint failed" in msg
            and "variabletag.variable_id" in msg
            and "variabletag.tag_id" in msg
        )
