from typing import Sequence

from peewee import IntegrityError

from .base_repository import BaseRepository
from cmdbox.database import db
from cmdbox.repositories.validators import CommandValidator
from cmdbox.exceptions import (
    ValidationError,
    AliasConflictError,
    UnknownAliasError,
    UnknownTagError,
    TagAttachError,
    TagDetachError,
)
from cmdbox.models import Command, Tag, CommandTag
from cmdbox.repositories.results import TagAttachResult, TagDetachResult


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

    def list_by_tag(
        self,
        tags: Sequence[str],
        order_by: str | Sequence[str] = "alias",
        limit: int = 25,
    ) -> list[Command]:
        """
        Fetches a list of commands filtered by specified tags and ordered by specific
        criteria.

        This method retrieves a list of `Command` objects associated with the tags
        provided in the `tags` parameter. The results can be customized through
        ordering and limited in number.

        Args:
            tags (Sequence[str]): A sequence of tag names used to filter the commands.
                Only commands associated with these tags will be retrieved.
            order_by (str | Sequence[str], optional): Specifies the criteria to order
                the commands. Defaults to "alias".
            limit (int, optional): The maximum number of commands to retrieve. Defaults
                to 25.

        Returns:
            list[Command]: A list of `Command` objects that meet the specified filters
            and ordering criteria.
        """
        tags_actual = self._get_tags_by_name(*tags)
        commands = (
            Command.select()
            .join(CommandTag)
            .where(CommandTag.tag << tags_actual)
            .distinct()
        )
        ordering = self._resolve_ordering(order_by)
        return list(commands.order_by(*ordering).limit(limit))

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
