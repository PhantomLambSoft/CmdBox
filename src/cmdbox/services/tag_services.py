from typing import Sequence

from cmdbox.models import Tag
from cmdbox.repositories.tag_repository import TagRepository


class TagServices:
    """
    Provides services for managing tags.

    The `TagServices` class encapsulates the logic for working with tags,
    offering methods to create, update, delete, and retrieve tag records.
    It also supports additional functionality such as searching and listing tags
    with sorting and limit options.

    Attributes:
        tag_repository (TagRepository): Repository for managing tag records.
    """

    def __init__(self, tag_repository: TagRepository):
        self._repo = tag_repository

    def create_tag(self, name: str, description: str | None = None) -> Tag:
        """
        Creates a new tag by storing the provided name and optional description
        into the database.

        Args:
            name (str): The name of the tag to be created.
            description (str | None): An optional description for the tag. Defaults to None.

        Returns:
            Tag: The created tag record.
        """
        return self._repo.create(name=name, description=description)

    def update_tag(self, name: str, **fields) -> Tag:
        """
        Updates an existing tag by its name with new field values.

        Retrieves the tag corresponding to the given name and updates it
        with the provided fields. The update is performed using the repository.

        Args:
            name (str): The name of the tag to update.
            **fields: Arbitrary field values to update on the tag.

        Returns:
            Tag: The updated tag object.
        """
        tag = self._repo.get_by_name(name)
        return self._repo.update(tag, **fields)

    def delete_tag(self, name: str) -> bool:
        """
        Deletes a tag by its name.

        This method removes a tag record from the repository that matches the
        provided name.

        Args:
            name: The name of the tag to delete.

        Returns:
            bool: True if the tag was deleted successfully, False otherwise.
        """
        tag = self._repo.get_by_name(name)
        return self._repo.delete(tag)

    def get_tag(self, name: str) -> Tag:
        """
        Retrieves a tag by its name.

        This method fetches a tag object associated with the given name from the
        repository.

        Args:
            name (str): The name of the tag to retrieve.

        Returns:
            Tag: The tag object associated with the given name.
        """
        return self._repo.get_by_name(name)

    def list_tags(
        self,
        order_by: str | Sequence[str] = "name",
        limit: int = 25,
    ) -> list[Tag]:
        """
        Lists tags with sorting and limit options.

        This function fetches a list of tags from the repository. The results can
        be sorted and limited based on the provided arguments.

        Args:
            order_by (str | Sequence[str]): Specifies the field(s) to sort the results by.
                Default is "name".
            limit (int, optional): The maximum number of tags to return. Default is 25.

        Returns:
            list[Tag]: A list of tags sorted according to the specified criteria.
        """
        return self._repo.list_all(order_by, limit)

    def search(
        self,
        query: str,
        fields: str | Sequence[str] | None = ("name", "description"),
    ) -> list[Tag]:
        """
        Searches for tags that match the given query across specified fields.

        This method allows you to perform a search within the data repository for tags
        that match the provided query string in the specified fields. It returns a list of
        tags that satisfy the search criteria.

        Args:
            query (str): The search term used for matching against the repository.
            fields (str | Sequence[str] | None): The fields to perform the search within. Defaults
                to ("name", "description"). If None, no specific fields are targeted.

        Returns:
            list[Tag]: A list of Tag objects that match the search query.
        """
        return self._repo.search(query, fields)
