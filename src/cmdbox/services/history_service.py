import json
from typing import Callable

from cmdbox.models import CommandHistory
from cmdbox.repositories.history_repository import HistoryRepository
from cmdbox.services.errors import HistoryIndexError
from cmdbox.settings.models import Settings


class HistoryService:

    def __init__(self, repo: HistoryRepository, get_settings: Callable[[], Settings]):
        self._repo = repo
        self._get_settings = get_settings

    def get_recent(
        self,
        alias: str | None = None,
        limit: int = 20,
    ) -> list[CommandHistory]:
        return self._repo.get_recent(alias, limit)

    def get_by_ref(self, ref: str, alias: str | None = None) -> CommandHistory:
        if ref.isdigit():
            return self._get_by_index(int(ref), alias=alias)
        return self._repo.get_by_id(ref)

    def get_variables(self, entry: CommandHistory) -> dict[str, str] | None:
        if entry.variables_used is None:
            return None
        return json.loads(entry.variables_used)

    def delete_by_ref(self, ref: str) -> bool:
        entry = self.get_by_ref(ref)
        return self._repo.delete_by_id(entry.id)

    def clear(self, alias: str | None = None) -> int:
        return self._repo.clear(alias=alias)

    def _get_by_index(self, index: int, alias: str | None = None) -> CommandHistory:
        """
        Fetches a specific command history entry by its index.

        This method retrieves a command history entry from the repository based on
        the passed index. If the index is out of range or invalid, a
        HistoryIndexError is raised. The method optionally allows limiting the
        retrieval to a specific alias.

        Args:
            index (int): The one-based index of the command history entry to
                retrieve. Must be greater than 0.
            alias (str | None): Optional alias to filter the history entries. If
                None, no alias filter is applied.

        Returns:
            CommandHistory: The command history entry corresponding to the provided
            index.

        Raises:
            HistoryIndexError: If the provided index is less than 1 or if the index
            exceeds the number of available entries.
        """
        if index < 1:
            raise HistoryIndexError(index=index)
        entries = self._repo.get_recent(alias=alias, limit=index)
        if index > len(entries):
            raise HistoryIndexError(index=index)
        return entries[index - 1]
