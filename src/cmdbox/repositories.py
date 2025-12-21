from typing import Sequence, Generic, TypeVar, Type

from peewee import Model, fn, IntegrityError, Node

from cmdbox.database import db
from cmdbox.domain.validators import CommandValidator, VariableValidator, TagValidator
from cmdbox.exceptions import (
    ValidationError,
    AliasConflictError,
    UnknownAliasError,
    NameConflictError,
    UnknownNameError,
    UnknownTagError,
    TagAttachError,
    TagDetachError,
)
from cmdbox.models import Command, Variable, Tag, CommandTag
from cmdbox._types import TagAttachResult, TagDetachResult


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


class CommandRepository(BaseRepository[Command]):

    model = Command

    def __init__(self, validator: CommandValidator | None = None):
        self.validator = validator or CommandValidator()

    def create(
        self,
        alias: str,
        template: str,
        description: str | None = None,
        tags: Sequence[str] | None = None,
    ) -> Command:
        """
        Validates and creates a new Command object based on provided input parameters.

        Args:
            alias (str): Unique identifier for the command to be created.
            template (str): Template string associated with the command.
            description (str | None): Optional description of the command.
            tags (Sequence[str] | None): Optional sequence of tag names to associate with the command.

        Returns:
            Command: The created Command object.

        Raises:
            AliasConflictError: If the provided alias already exists and causes a
                unique constraint violation during creation.
            IntegrityError: If any database integrity issue occurs during the creation.
        """
        alias = alias.strip() if alias else None
        self.validator.validate_create(
            alias=alias, template=template, description=description
        )
        tags_actual = self._get_tags_by_name(*tags or [])
        try:
            with db.atomic():
                cmd = Command.create(
                    alias=alias,
                    template=template,
                    description=description,
                )
                self._attach_tags(cmd, tags_actual)
                return cmd
        except IntegrityError as exc:
            if alias is not None and self._is_unique_alias_violation(exc):
                raise AliasConflictError(alias=alias) from exc
            raise

    def get_by_alias(self, alias: str) -> Command:
        """
        Retrieves a Command instance by its alias.

        This method searches for a Command instance using the given alias. The alias
        is converted to lowercase before searching. If no matching Command instance
        is found, an UnknownAliasError is raised.

        Args:
            alias (str): The alias of the Command being searched for.

        Returns:
            Command: The Command instance that matches the provided alias.

        Raises:
            UnknownAliasError: If no Command is found with the provided alias.
        """
        alias = alias.lower()
        cmd = Command.get_or_none(Command.alias == alias)
        if cmd is None:
            raise UnknownAliasError(alias=alias)
        return cmd

    def update(self, cmd_alias: str, **fields) -> Command:
        """
        Updates an existing command based on the provided alias and fields.

        This method retrieves a command by its alias and updates its fields with the
        provided values. It validates the updates, ensuring field integrity and
        uniqueness. If the command doesn't exist or an update violates constraints,
        appropriate actions are taken.

        Args:
            cmd_alias: The alias of the command to update.
            **fields: Arbitrary keyword arguments representing the fields to update.
                Supported fields include 'alias', 'template', and 'description'.

        Returns:
            Command: The updated command object if successful.

        Raises:
            ValidationError: If any of the provided fields are invalid or do not exist
                on the command.
            AliasConflictError: If the new alias conflicts with an existing command's
                alias.
            IntegrityError: If there is a general integrity constraint violation during
                the update process.
        """
        cmd = self.get_by_alias(cmd_alias)
        if not fields:
            raise ValueError("No fields provided for update.")

        # Strip whitespace from alias field if alias field is supplied
        if "alias" in fields and fields.get("alias") is not None:
            fields["alias"] = fields.get("alias").strip()

        self.validator.validate_update(
            alias=fields.get("alias", cmd_alias),
            template=fields.get("template", cmd.template),
            description=fields.get("description", None),
        )
        try:
            for key, value in fields.items():
                if not hasattr(cmd, key):
                    raise ValidationError(f"Invalid field: {key}")
                if value is not None:
                    setattr(cmd, key, value)
            cmd.save()
            return cmd
        except IntegrityError as exc:
            alias = fields.get("alias", "")
            if alias is not None and self._is_unique_alias_violation(exc):
                raise AliasConflictError(alias=alias) from exc
            raise

    def add_tags(self, alias: str, tags: Sequence[str]) -> TagAttachResult:
        """
        Attach tags to a command identified by its alias.

        This function associates tags with a specified command. If the tags already
        exist for the command, they are added to an existing list. Otherwise, new tags
        are created and linked to the command. If no tags are provided, it returns
        immediately with empty results. In case of a database integrity issue, an
        appropriate error is raised.

        Args:
            alias (str): The alias identifier of the command to which the tags are to
                be attached.
            tags (Sequence[str]): A collection of tag names to be attached to the
                command.

        Returns:
            TagAttachResult: An object containing lists of newly added tags and
                tags that already existed.

        Raises:
            TagAttachError: If there is an issue attaching the tags to the command,
                typically due to database integrity constraints.
        """
        if not tags:
            return TagAttachResult(added=[], existing=[])
        tags_actual = self._get_tags_by_name(*tags)
        command = self.get_by_alias(alias)
        try:
            with db.atomic():
                return self._attach_tags(command, tags_actual)
        except UnknownTagError:
            raise
        except IntegrityError as exc:
            raise TagAttachError("Could not attach tags to command.") from exc

    def remove_tags(self, alias: str, tags: Sequence[str]) -> TagDetachResult:
        """
        Removes tags from a command identified by the provided alias. The method first validates
        the tags to ensure they exist in the database, then attempts to remove the associations
        between the command and the respective tags. If a tag is not attached to the command, it
        is recorded in the `not_attached` list, while successfully removed tags are recorded in
        the `removed` list. Any errors during detachment raise a `TagDetachError`.

        Args:
            alias (str): The unique identifier or alias of the command from which the tags will
                be detached.
            tags (Sequence[str]): A list of tag names to be detached from the command.

        Returns:
            TagDetachResult: Object that contains two lists:
                - `removed`: A list of successfully detached tags.
                - `not_attached`: A list of tags that were not associated with the command.

        Raises:
            TagDetachError: If the detachment process encounters an issue, such as database
                integrity errors.
        """
        if not tags:
            return TagDetachResult(removed=[], not_attached=[])
        tags_actual = self._get_tags_by_name(*tags)
        command = self.get_by_alias(alias)
        removed = []
        not_attached = []
        try:
            with db.atomic():
                for tag in tags_actual:
                    deleted = (
                        CommandTag.delete()
                        .where(
                            (CommandTag.command == command) & (CommandTag.tag == tag)
                        )
                        .execute()
                    )
                    if deleted:
                        removed.append(tag.name)
                    else:
                        not_attached.append(tag.name)
        except IntegrityError as exc:
            raise TagDetachError("Could not detach tags from command.") from exc
        return TagDetachResult(removed=removed, not_attached=not_attached)

    def list_all(
        self, order_by: str | Sequence[str] = "alias", limit: int = 25
    ) -> list[Command]:
        """
        Lists all Command objects from the database, optionally ordered by specified fields.

        Args:
            order_by (str | Sequence[str]): A string or sequence of strings indicating the field(s)
                by which the Command objects should be ordered. Defaults to "name".
            limit (int): The maximum number of Command objects to return. Defaults to 25.

        Returns:
            list[Command]: A list of Command objects retrieved from the database,
                sorted based on the specified ordering criteria.
        """
        ordering = self._resolve_ordering(order_by)
        return list(Command.select().order_by(*ordering).limit(limit))

    def list_by_tag(self, *tags: str) -> list[Command]:
        pass

    def search(
        self, query: str, fields: str | Sequence[str] | None = ("alias", "description")
    ) -> list[Command]:
        """
        Searches for commands matching the given query across specified fields.

        This function performs a case-insensitive search for the query in the specified
        fields of the Command model. It calculates a relevance score for each match
        based on the occurrence count and position of the query within the fields and
        sorts the results by relevance.

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
        return self._search(query, "alias", fields)

    def delete(self, alias: str) -> bool:
        """
        Deletes the command with the specified alias.

        Args:
            alias (str): The alias of the command to delete.

        Returns:
            bool: True if the command was deleted, False otherwise.
        """
        cmd = self.get_by_alias(alias)
        if not cmd:
            return False
        cmd.delete_instance()
        return True

    def _attach_tags(self, cmd: Command, tags: Sequence[Tag]) -> TagAttachResult:
        """
        Attaches tags to the given command. If a tag is already associated with the
        command, it is grouped under existing tags; otherwise, it is added to newly
        attached tags.

        Args:
            cmd (Command): The command to which tags are to be attached.
            tags (Sequence[Tag]): A sequence of tags to be attached to the command.

        Returns:
            TagAttachResult: An object containing lists of newly added and
            previously existing tags.
        """
        added = []
        existing = []
        for tag in tags:
            cmd_tag, created = CommandTag.get_or_create(command=cmd, tag=tag)
            if created:
                added.append(tag.name)
            else:
                existing.append(tag.name)
        return TagAttachResult(added=added, existing=existing)

    def _is_unique_alias_violation(self, exc: IntegrityError) -> bool:
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
        return "UNIQUE constraint failed" in msg and f"{table}.alias" in msg

    def _is_unique_command_tag_violation(self, exc: IntegrityError) -> bool:
        msg = str(exc)
        return (
            "UNIQUE constraint failed" in msg
            and "commandtag.command_id" in msg
            and "commandtag.tag_id" in msg
        )


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


class TagRepository(BaseRepository[Tag]):

    model = Tag

    def __init__(self, validator: TagValidator | None = None):
        self.validator = validator or TagValidator()

    def create(self, name: str, description: str | None = None) -> Tag:
        name = name.strip() if name else None
        self.validator.validate_create(name=name, description=description)
        try:
            return Tag.create(name=name, description=description)
        except IntegrityError as exc:
            if name is not None and self._is_unique_name_violation(exc):
                raise NameConflictError(name=name) from exc
            raise

    def get_by_name(self, name: str) -> Tag | None:
        name = name.lower()
        tag = Tag.get_or_none(Tag.name == name)
        if tag is None:
            raise UnknownNameError(name=name)
        return tag

    def update(self, tag_name: str, **fields) -> Tag | None:
        tag = self.get_by_name(tag_name)
        if not fields:
            raise ValueError("No fields provided for update.")

        if "name" in fields and fields.get("name") is not None:
            fields["name"] = fields.get("name").strip()

        self.validator.validate_update(
            name=fields.get("name", tag.name),
            description=fields.get("description", tag.description),
        )
        try:
            for key, value in fields.items():
                if not hasattr(tag, key):
                    raise ValidationError(f"Invalid field: {key}")
                if value is not None:
                    setattr(tag, key, value)
            tag.save()
            return tag
        except IntegrityError as exc:
            name = fields.get("name", "")
            if name is not None and self._is_unique_name_violation(exc):
                raise NameConflictError(name=name) from exc
            raise

    def list_all(
        self, order_by: str | Sequence[str] = "name", limit: int = 25
    ) -> list[Tag]:
        ordering = self._resolve_ordering(order_by)
        return list(Tag.select().order_by(*ordering).limit(limit))

    def search(
        self, query: str, fields: str | Sequence[str] | None = ("name", "description")
    ) -> list[Tag]:
        return self._search(query, "name", fields)

    def delete(self, name: str) -> bool:
        tag = self.get_by_name(name)
        if not tag:
            return False
        tag.delete_instance()
        return True
