import json
from typing import Callable

from cmdbox.models import CommandHistory, Profile
from cmdbox.repositories.history_repository import HistoryRepository
from cmdbox.repositories.profile_repository import ProfileRepository
from cmdbox.services.errors import HistoryIndexError
from cmdbox.settings.models import Settings


class HistoryService:

    def __init__(
        self,
        repo: HistoryRepository,
        get_settings: Callable[[], Settings],
        profile_repository: ProfileRepository,
    ):
        self._repo = repo
        self._get_settings = get_settings
        self._profile_repo = profile_repository

    def resolve_profile(self, profile: str | None) -> Profile | None:
        return self._profile_repo.get_by_name(profile) if profile else None

    def get_recent(
        self,
        alias: str | None = None,
        limit: int = 20,
        profile: str | None = None,
    ) -> list[CommandHistory]:
        resolved_profile = self.resolve_profile(profile)
        return self._repo.get_recent(alias, limit, profile=resolved_profile)

    def get_by_ref(
        self,
        ref: str,
        alias: str | None = None,
        profile: str | None = None,
    ) -> CommandHistory:
        resolved_profile = self.resolve_profile(profile)
        if ref.isdigit():
            return self._get_by_index(int(ref), alias=alias, profile=resolved_profile)
        return self._repo.get_by_id(ref, profile=resolved_profile)

    def get_variables(self, entry: CommandHistory) -> dict[str, str] | None:
        if entry.variables_used is None:
            return None
        return json.loads(entry.variables_used)

    def delete_by_ref(self, ref: str, profile: str | None = None) -> bool:
        resolved_profile = self.resolve_profile(profile)
        entry = self.get_by_ref(ref, profile=profile)
        return self._repo.delete_by_id(entry.id, profile=resolved_profile)

    def clear(
        self,
        alias: str | None = None,
        profile: str | None = None,
    ) -> int:
        resolved_profile = self.resolve_profile(profile)
        return self._repo.clear(alias=alias, profile=resolved_profile)

    def _get_by_index(
        self,
        index: int,
        alias: str | None = None,
        profile: Profile | None = None,
    ) -> CommandHistory:
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
            profile (str | None): Optional profile to filter the history entries. If
                None, no profile filter is applied.

        Returns:
            CommandHistory: The command history entry corresponding to the provided
            index.

        Raises:
            HistoryIndexError: If the provided index is less than 1 or if the index
            exceeds the number of available entries.
        """
        if index < 1:
            raise HistoryIndexError(index=index)
        entries = self._repo.get_recent(alias=alias, limit=index, profile=profile)
        if index > len(entries):
            raise HistoryIndexError(index=index)
        return entries[index - 1]
