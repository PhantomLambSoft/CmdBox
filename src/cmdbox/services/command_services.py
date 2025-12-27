from typing import Sequence

from cmdbox.models import Tag, Command
from cmdbox.repositories.command_repository import CommandRepository
from cmdbox.database import db
from cmdbox.repositories.results import TagAttachResult, TagDetachResult
from cmdbox.repositories.tag_repository import TagRepository


class CommandServices:
    """
    Provides services for managing commands and their associated tags.

    The `CommandServices` class encapsulates the logic for working with commands and
    tags, offering methods to create, update, delete, and retrieve command records
    and their tag associations. It also supports additional functionality such as
    searching, tagging, and listing commands with filtering and sorting options.

    Attributes:
        command_repository (CommandRepository): Repository for managing command records.
        tag_repository (TagRepository): Repository for managing tag-related operations.
    """

    def __init__(
        self, command_repository: CommandRepository, tag_repository: TagRepository
    ):
        self._repo = command_repository
        self._tag_repo = tag_repository

    def create_command(
        self,
        alias: str,
        template: str,
        description: str | None = None,
        tags: list[str] | None = None,
    ) -> Command:
        """
        Creates a new command by storing the provided alias, template, and optional description
        and tags into the database. If tags are provided, they are associated with the created
        command.

        Args:
            alias (str): The alias of the command to be created.
            template (str): The template associated with the command.
            description (str | None): An optional description for the command. Defaults to None.
            tags (list[str] | None): An optional list of tags to associate with the command.
                Defaults to None.

        Returns:
            Command: The created command record.
        """
        with db.atomic():
            tags = self._get_tags(tags)
            cmd = self._repo.create(
                alias=alias, template=template, description=description
            )
            if tags:
                self._repo.add_tags(cmd, tags)
        return cmd

    def update_command(self, alias: str, **fields) -> Command:
        """
        Updates an existing command by its alias with new field values.

        Retrieves the command corresponding to the given alias and updates it
        with the provided fields. The update is performed using the repository.

        Args:
            alias (str): The alias of the command to update.
            **fields: Arbitrary field values to update on the command.

        Returns:
            Command: The updated command object.
        """
        cmd = self._repo.get_by_alias(alias)
        return self._repo.update(cmd, **fields)

    def delete_command(self, alias: str) -> bool:
        """
        Deletes a command by its alias.

        This method removes a command record from the repository that matches the
        provided alias.

        Args:
            alias: The alias of the command to delete.

        Returns:
            bool: True if the command was deleted successfully, False otherwise.
        """
        cmd = self._repo.get_by_alias(alias)
        return self._repo.delete(cmd)

    def add_tags(self, alias: str, tags: list[str]) -> TagAttachResult:
        """
        Adds specified tags to an existing alias.

        The method retrieves a command object associated with the given alias and
        applies the provided tags to the command.

        Args:
            alias (str): The alias identifying the command to which tags are to be
                attached.
            tags (list[str]): A list of tags to attach to the command.

        Returns:
            TagAttachResult: The result of the tag attachment operation, indicating
                whether the tags were successfully attached to the command.
        """
        cmd = self._repo.get_by_alias(alias)
        tags = self._get_tags(tags)
        return self._repo.add_tags(cmd, tags)

    def remove_tags(self, alias: str, tags: list[str]) -> TagDetachResult:
        """
        Removes specific tags associated with a command alias.

        This method retrieves a command by its alias and detaches the specified tags
        from it. The updated tag associations will be handled by the repository.

        Args:
            alias: The unique identifier for the command to update.
            tags: A list of tags to be removed from the specified command.

        Returns:
            TagDetachResult: An object representing the result of tag detachment.
        """
        cmd = self._repo.get_by_alias(alias)
        tags = self._get_tags(tags)
        return self._repo.remove_tags(cmd, tags)

    def get_command(self, alias: str) -> Command:
        """
        Retrieves a command by its alias.

        This method fetches a command object associated with the given alias from the
        repository.

        Args:
            alias (str): The alias of the command to retrieve.

        Returns:
            Command: The command object associated with the given alias.
        """
        cmd = self._repo.get_by_alias(alias)
        return cmd

    def list_commands(
        self,
        order_by: str | Sequence[str] = "alias",
        tags: Sequence[str] | None = None,
        limit: int = 25,
    ) -> list[Command]:
        """
        Lists commands, optionally filtered by tags, with sorting and limit options.

        This function fetches a list of commands, either filtered by specific tags
        or returning all available commands. The results can be sorted and limited
        based on the provided arguments.

        Args:
            order_by (str | Sequence[str]): Specifies the field(s) to sort the results by.
                Default is "alias".
            tags (Sequence[str], optional): A list of tags to filter the commands. If
                provided, only commands matching the tags will be included.
            limit (int, optional): The maximum number of commands to return. Default is 25.

        Returns:
            list[Command]: A list of commands matching the provided filters and sorted
            according to the specified criteria.
        """
        if tags:
            tags = self._get_tags(tags)
            return self._repo.list_by_tag(tags, order_by, limit)
        return self._repo.list_all(order_by, limit)

    def search(
        self,
        query: str,
        fields: str | Sequence[str] | None = ("alias", "description"),
    ) -> list[Command]:
        """
        Searches for commands that match the given query across specified fields.

        This method allows you to perform a search within the data repository for commands
        that match the provided query string in the specified fields. It returns a list of
        commands that satisfy the search criteria.

        Args:
            query (str): The search term used for matching against the repository.
            fields (str | Sequence[str] | None): The fields to perform the search within. Defaults
                to ("alias", "description"). If None, no specific fields are targeted.

        Returns:
            list[Command]: A list of Command objects that match the search query.
        """
        return self._repo.search(query, fields)

    def _get_tags(self, tags: Sequence[str] | None) -> list[Tag]:
        """
        Fetches a list of Tag objects based on the provided tag names.

        This method takes a list of tag names as input and retrieves the corresponding
        Tag objects from the tag repository. The resulting list of Tag objects is then
        returned.

        Args:
            tags (list[str] | None): A list of string representing the names of the tags to
                retrieve.

        Returns:
            list[Tag]: A list of Tag objects corresponding to the specified tag names.
        """
        if tags is None:
            return []
        ret_tags: list[Tag] = []
        for name in tags:
            tag = self._tag_repo.get_by_name(name)
            ret_tags.append(tag)
        return ret_tags
