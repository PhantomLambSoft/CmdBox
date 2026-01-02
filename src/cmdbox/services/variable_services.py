from typing import Sequence

from cmdbox.models import Tag, Variable
from cmdbox.repositories.variable_repository import VariableRepository
from cmdbox.database import db
from cmdbox.repositories.results import TagAttachResult, TagDetachResult
from cmdbox.repositories.tag_repository import TagRepository


class VariableServices:
    """
    Provides services for managing variables and their associated tags.

    The `VariableServices` class encapsulates the logic for working with variables and
    tags, offering methods to create, update, delete, and retrieve variable records
    and their tag associations. It also supports additional functionality such as
    searching, tagging, and listing variables with filtering and sorting options.

    Attributes:
        variable_repository (VariableRepository): Repository for managing variable records.
        tag_repository (TagRepository): Repository for managing tag-related operations.
    """

    def __init__(
        self, variable_repository: VariableRepository, tag_repository: TagRepository
    ):
        self._repo = variable_repository
        self._tag_repo = tag_repository

    def create_variable(
        self,
        name: str,
        value: str,
        tags: list[str] | None = None,
    ) -> Variable:
        """
        Creates a new variable by storing the provided name, value, and optional tags
        into the database. If tags are provided, they are associated with the created
        variable.

        Args:
            name (str): The name of the variable to be created.
            value (str): The value associated with the variable.
            tags (list[str] | None): An optional list of tags to associate with the variable.
                Defaults to None.

        Returns:
            Variable: The created variable record.
        """
        with db.atomic():
            tags = self._get_tags(tags)
            var = self._repo.create(name=name, value=value)
            if tags:
                self._repo.add_tags(var, tags)
        return var

    def update_variable(self, name: str, **fields) -> Variable:
        """
        Updates an existing variable by its name with new field values.

        Retrieves the variable corresponding to the given name and updates it
        with the provided fields. The update is performed using the repository.

        Args:
            name (str): The name of the variable to update.
            **fields: Arbitrary field values to update on the variable.

        Returns:
            Variable: The updated variable object.
        """
        var = self._repo.get_by_name(name)
        return self._repo.update(var, **fields)

    def delete_variable(self, name: str) -> bool:
        """
        Deletes a variable by its name.

        This method removes a variable record from the repository that matches the
        provided name.

        Args:
            name: The name of the variable to delete.

        Returns:
            bool: True if the variable was deleted successfully, False otherwise.
        """
        var = self._repo.get_by_name(name)
        return self._repo.delete(var)

    def add_tags(self, name: str, tags: list[str]) -> TagAttachResult:
        """
        Adds specified tags to an existing variable name.

        The method retrieves a variable object associated with the given name and
        applies the provided tags to the variable.

        Args:
            name (str): The name identifying the variable to which tags are to be
                attached.
            tags (list[str]): A list of tags to attach to the variable.

        Returns:
            TagAttachResult: The result of the tag attachment operation, indicating
                whether the tags were successfully attached to the variable.
        """
        var = self._repo.get_by_name(name)
        tags = self._get_tags(tags)
        return self._repo.add_tags(var, tags)

    def remove_tags(self, name: str, tags: list[str]) -> TagDetachResult:
        """
        Removes specific tags associated with a variable name.

        This method retrieves a variable by its name and detaches the specified tags
        from it. The updated tag associations will be handled by the repository.

        Args:
            name: The unique identifier for the variable to update.
            tags: A list of tags to be removed from the specified variable.

        Returns:
            TagDetachResult: An object representing the result of tag detachment.
        """
        var = self._repo.get_by_name(name)
        tags = self._get_tags(tags)
        return self._repo.remove_tags(var, tags)

    def get_variable(self, name: str) -> Variable:
        """
        Retrieves a variable by its name.

        This method fetches a variable object associated with the given name from the
        repository.

        Args:
            name (str): The name of the variable to retrieve.

        Returns:
            Variable: The variable object associated with the given name.
        """
        var = self._repo.get_by_name(name)
        return var

    def get_variable_by_id(self, var_id: int) -> Variable:
        var = self._repo.get_by_id(var_id)
        return var

    def list_variables(
        self,
        order_by: str | Sequence[str] = "name",
        tags: Sequence[str] | None = None,
        limit: int = 25,
    ) -> list[Variable]:
        """
        Lists variables, optionally filtered by tags, with sorting and limit options.

        This function fetches a list of variables, either filtered by specific tags
        or returning all available variables. The results can be sorted and limited
        based on the provided arguments.

        Args:
            order_by (str | Sequence[str]): Specifies the field(s) to sort the results by.
                Default is "name".
            tags (Sequence[str], optional): A list of tags to filter the variables. If
                provided, only variables matching the tags will be included.
            limit (int, optional): The maximum number of commands to return. Default is 25.

        Returns:
            list[Variable]: A list of variables matching the provided filters and sorted
            according to the specified criteria.
        """
        if tags:
            tags = self._get_tags(tags)
            return self._repo.list_by_tag(tags, order_by, limit)
        return self._repo.list_all(order_by, limit)

    def search(
        self,
        query: str,
        fields: str | Sequence[str] | None = None,
        limit: int = 25,
    ) -> list[Variable]:
        """
        Searches for variables that match the given query across specified fields.

        This method allows you to perform a search within the data repository for variables
        that match the provided query string in the specified fields. It returns a list of
        variables that satisfy the search criteria.

        Args:
            query (str): The search term used for matching against the repository.
            fields (str | Sequence[str] | None): The fields to perform the search within. Defaults
                to ("name", "value"). If None, no specific fields are targeted.
            limit (int): The maximum number of results to return. Defaults to 25.

        Returns:
            list[Variable]: A list of Variable objects that match the search query.
        """
        if not fields:
            fields = ("name", "value")
        return self._repo.search(query, fields=fields, limit=limit)

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
