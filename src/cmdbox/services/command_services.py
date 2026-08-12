from typing import Sequence

from cmdbox.models import Tag, Command, Profile
from cmdbox.repositories.command_repository import CommandRepository
from cmdbox.database import db
from cmdbox.repositories.errors import UnknownAliasError
from cmdbox.repositories.profile_repository import ProfileRepository
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
        profile_repository (ProfileRepository): Repository for managing profile-related operations.
    """

    def __init__(
        self,
        command_repository: CommandRepository,
        tag_repository: TagRepository,
        profile_repository: ProfileRepository,
    ):
        self._repo = command_repository
        self._tag_repo = tag_repository
        self._profile_repo = profile_repository

    def resolve_profile(self, profile: str | None) -> Profile | None:
        return self._profile_repo.get_by_name(profile) if profile else None

    def create_command(
        self,
        alias: str,
        template: str,
        description: str | None = None,
        tags: list[str] | None = None,
        cwd: str | None = None,
        shell: str | None = None,
        env: dict[str, str] | None = None,
        timeout: int | None = None,
        profile: str | None = None,
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
            cwd (str | None): Optional working directory to run the command from. Defaults to None.
            shell (str | None): Optional shell to use when running the command. Defaults to None.
            env (dict[str, str] | None): Optional environment variables to set when running
                the command. Defaults to None.
            timeout (int | None): Optional maximum number of seconds before the process is
                killed. Defaults to None.
            profile (str | None): Optional profile to associate with the command. Defaults to None.

        Returns:
            Command: The created command record.
        """
        profile = self.resolve_profile(profile)
        with db.atomic():
            tags = self._get_tags(tags)
            cmd = self._repo.create(
                alias=alias,
                template=template,
                description=description,
                cwd=cwd,
                shell=shell,
                env=env,
                timeout=timeout,
                profile=profile,
            )
            if tags:
                self._repo.add_tags(cmd, tags)

        command = self._repo.get_by_id(cmd.id)

        return command

    def update_command(self, current_alias: str, **fields) -> Command:
        """
        Updates an existing command by its alias with new field values.

        Retrieves the command corresponding to the given alias and updates it
        with the provided fields. The update is performed using the repository.

        Args:
            current_alias (str): The alias of the command to update.
            **fields: Arbitrary field values to update on the command.

        Returns:
            Command: The updated command object.
        """
        cmd = self._repo.get_by_alias(current_alias)
        return self._repo.update(cmd, **fields)

    def delete_command(self, alias_: str, profile: str | None = None) -> bool:
        """
        Deletes a command by its alias.

        This method removes a command record from the repository that matches the
        provided alias.

        Args:
            alias_: The alias of the command to delete.
            profile (str | None): Optional profile associated with the command. Defaults to None.

        Returns:
            bool: True if the command was deleted successfully, False otherwise.
        """
        profile = self.resolve_profile(profile)
        cmd = self._repo.get_by_alias(alias_, profile=profile)
        return self._repo.delete(cmd)

    def add_tags(
        self, alias: str, tags: list[str], profile: str | None = None
    ) -> TagAttachResult:
        """
        Adds specified tags to an existing alias.

        The method retrieves a command object associated with the given alias and
        applies the provided tags to the command.

        Args:
            alias (str): The alias identifying the command to which tags are to be
                attached.
            tags (list[str]): A list of tags to attach to the command.
            profile (str | None): Optional profile associated with the command. Defaults to None.

        Returns:
            TagAttachResult: The result of the tag attachment operation, indicating
                whether the tags were successfully attached to the command.
        """
        profile = self.resolve_profile(profile)
        cmd = self._repo.get_by_alias(alias, profile=profile)
        tags = self._get_tags(tags)
        return self._repo.add_tags(cmd, tags)

    def remove_tags(
        self, alias: str, tags: list[str], profile: str | None = None
    ) -> TagDetachResult:
        """
        Removes specific tags associated with a command alias.

        This method retrieves a command by its alias and detaches the specified tags
        from it. The updated tag associations will be handled by the repository.

        Args:
            alias (str): The unique identifier for the command to update.
            tags (list[str]): A list of tags to be removed from the specified command.
            profile (str | None): Optional profile associated with the command. Defaults to None.

        Returns:
            TagDetachResult: An object representing the result of tag detachment.
        """
        profile = self.resolve_profile(profile)
        cmd = self._repo.get_by_alias(alias, profile=profile)
        tags = self._get_tags(tags)
        return self._repo.remove_tags(cmd, tags)

    def get_command(self, alias: str, profile: str | None = None) -> Command:
        """
        Retrieves a command by its alias.

        This method fetches a command object associated with the given alias from the
        repository.

        Args:
            alias (str): The alias of the command to retrieve.
            profile (str | None): Optional profile associated with the command. Defaults to None.

        Returns:
            Command: The command object associated with the given alias.
        """
        profile = self.resolve_profile(profile)
        cmd = self._repo.get_by_alias(alias, profile=profile)
        return cmd

    def get_command_or_none(
        self, alias: str, profile: str | None = None
    ) -> Command | None:
        """
        Gets the command associated with the given alias or returns None if the alias
        is not recognized.

        This method attempts to retrieve a command by its alias. If the alias is not
        found, or if an error is raised due to an unknown alias, the method will return
        None instead of raising the error.

        Args:
            alias (str): The alias of the command to retrieve.
            profile (str | None): Optional profile associated with the command. Defaults to None.

        Returns:
            Command | None: The command associated with the alias if found, otherwise
            None.
        """
        try:
            return self.get_command(alias, profile=profile)
        except UnknownAliasError:
            return None

    def get_command_by_id(self, cmd_id: int) -> Command:
        """
        Retrieves a command by its unique identifier from the repository.

        This method accesses the repository to fetch a command based on the provided
        command ID. It assumes that the command ID corresponds to a valid identifier
        within the repository.

        Args:
            cmd_id (int): The unique identifier of the command to be retrieved.

        Returns:
            The command object associated with the given command ID, or None if no
            matching command is found.
        """
        cmd = self._repo.get_by_id(cmd_id)
        return cmd

    def list_commands(
        self,
        order_by: str | Sequence[str] = "alias",
        tags: Sequence[str] | None = None,
        limit: int = 25,
        profile: str | None = None,
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
            profile (str | None): Optional profile associated with the commands. Defaults to None.

        Returns:
            list[Command]: A list of commands matching the provided filters and sorted
            according to the specified criteria.
        """
        profile = self.resolve_profile(profile)
        if tags:
            tags = self._get_tags(tags)
            return self._repo.list_by_tag(tags, order_by, limit, profile=profile)
        return self._repo.list_all(order_by, limit, profile=profile)

    def search(
        self,
        query: str,
        fields: str | Sequence[str] | None = None,
        limit: int = 25,
        profile: str | None = None,
    ) -> list[Command]:
        """
        Searches for commands that match the given query across specified fields.

        This method allows you to perform a search within the data repository for commands
        that match the provided query string in the specified fields. It returns a list of
        commands that satisfy the search criteria.

        Args:
            query (str): The search term used for matching against the repository.
            fields (str | Sequence[str] | None): The fields to perform the search within. Defaults
                to ("alias", "template", "description"). If None, no specific fields are targeted.
            limit (int): The maximum number of results to return. Defaults to 25.
            profile (str | None): The profile to use for the search. Defaults to None.

        Returns:
            list[Command]: A list of Command objects that match the search query.
        """
        if not fields:
            fields = ("alias", "template", "description")
        profile = self.resolve_profile(profile)
        return self._repo.search(query, fields=fields, limit=limit, profile=profile)

    def move_command(
        self, alias: str, target_profile: str, profile: str | None = None
    ) -> Command:
        source_profile = self.resolve_profile(profile)
        cmd = self._repo.get_by_alias(alias, profile=source_profile)
        target_profile = self._profile_repo.get_by_name(target_profile)
        return self._repo.update(cmd, profile=target_profile)

    def copy_command(
        self,
        alias: str,
        target_profile: str,
        new_alias: str | None = None,
        profile: str | None = None,
    ) -> Command:
        import json

        source_profile = self.resolve_profile(profile)
        source = self._repo.get_by_alias(alias, profile=source_profile)
        target_profile = self._profile_repo.get_by_name(target_profile)
        source_tags = [ct.tag for ct in source.tags]

        with db.atomic():
            copy = self._repo.create(
                alias=new_alias or source.alias,
                template=source.template,
                description=source.description,
                cwd=source.cwd,
                shell=source.shell,
                env=json.loads(source.env) if source.env else None,
                timeout=source.timeout,
                profile=target_profile,
            )
            if source_tags:
                self._repo.add_tags(copy, source_tags)

        return self._repo.get_by_id(copy.id, profile=target_profile)

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
