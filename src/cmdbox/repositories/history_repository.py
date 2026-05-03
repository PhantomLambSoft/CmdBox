import json
import uuid
from datetime import datetime

from cmdbox.models import CommandHistory
from cmdbox.repositories.errors import (
    UnknownHistoryEntryError,
    AmbiguousHistoryIdEntryError,
)


class HistoryRepository:

    def record(
        self,
        *,
        alias: str,
        template: str,
        resolved: str,
        variables_used: dict[str, str] | None,
        exit_code: int | None,
        limit: int | None,
    ) -> CommandHistory:
        """
        Records a command execution in the history and applies retention rules
        to maintain the history within a specified limit.

        This function creates a new entry in the command history by storing
        details about the executed command, including its alias, the template
        used, resolved result, variables involved, and its exit code. It then
        applies retention policies for the alias based on the provided limit.

        Args:
            alias (str): The alias of the command being recorded.
            template (str): The template of the command executed.
            resolved (str): The resolved command after variable substitution.
            variables_used (dict[str, str] | None): The variables used in the
                command execution. Defaults to None if no variables were used.
            exit_code (int | None): The exit code of the command. Defaults to
                None if the command has no exit code.
            limit (int | None): The retention limit for the history of the
                specified alias. If None, no limit is applied.

        Returns:
            CommandHistory: The newly created command history entry.
        """
        entry = CommandHistory.create(
            id=uuid.uuid4().hex,
            alias=alias,
            template=template,
            resolved=resolved,
            variables_used=json.dumps(variables_used) if variables_used else None,
            exit_code=exit_code,
            run_at=datetime.now(),
        )
        self._apply_retention(alias, limit)
        return entry

    def get_by_id(self, ref: str) -> CommandHistory:
        """
        Retrieves a `CommandHistory` entry by its ID or a prefix of the ID.

        This method fetches an entry from the `CommandHistory` database table using the provided
        ID or ID prefix. It ensures that one and only one match exists for the given reference.
        If no match is found, or if multiple matches are found, specific exceptions are raised.

        Args:
            ref (str): The ID or prefix of the command history entry to retrieve.

        Returns:
            CommandHistory: The command history entry corresponding to the provided reference.

        Raises:
            UnknownHistoryEntryError: If no command history entry matches the given reference.
            AmbiguousHistoryIdEntryError: If multiple command history entries match the given
                prefix, making the reference ambiguous.
        """
        matches = list(
            CommandHistory.select().where(CommandHistory.id % (ref + "%")).limit(2)
        )
        if not matches:
            raise UnknownHistoryEntryError(ref=ref)
        if len(matches) > 1:
            raise AmbiguousHistoryIdEntryError(prefix=ref)
        return matches[0]

    def get_recent(
        self, alias: str | None = None, limit: int = 25
    ) -> list[CommandHistory]:
        """
        Fetches the most recent command history records, optionally filtered by alias and limited
        to a specified number of entries.

        This method queries the CommandHistory records in descending order of their `run_at`
        timestamps, which represent the time the commands were executed. If the `alias` parameter
        is provided, the results are filtered to include only those records where the alias matches
        the provided value. The number of returned records can be limited by the `limit` parameter.
        If `limit` is 0 or negative, no results will be retrieved.

        Args:
            alias: Optional. The alias name to filter the command history records.
            limit: The maximum number of command history records to return. Defaults
                to 25. If set to 0 or a negative number, no records will be retrieved.

        Returns:
            list[CommandHistory]: A list of `CommandHistory` objects representing the
            queried command history records.
        """
        query = CommandHistory.select().order_by(CommandHistory.run_at.desc())
        if alias:
            query = query.where(CommandHistory.alias == alias)
        if limit > 0:
            query = query.limit(limit)
        return list(query)

    def delete_by_id(self, history_id: str) -> bool:
        entry = self.get_by_id(history_id)
        entry.delete_instance()
        return True

    def clear(self, alias: str | None = None) -> int:
        query = CommandHistory.delete()
        if alias:
            query = query.where(CommandHistory.alias == alias)
        return query.execute()

    def _apply_retention(self, alias: str, limit: int | None) -> None:
        """
        Applies retention logic by limiting the number of recent command history
        entries for the specified alias. Deletes older command history entries
        exceeding the given limit.

        Args:
            alias (str): The alias to filter command history entries.
            limit (int | None): The maximum number of recent entries to retain. If
                the limit is less than or equal to zero, no action is performed.
        """
        if limit <= 0:
            return
        keep_ids = [
            row.id
            for row in CommandHistory.select(CommandHistory.id)
            .where(CommandHistory.alias == alias)
            .order_by(CommandHistory.run_at.desc())
            .limit(limit)
        ]
        if len(keep_ids) >= limit:
            (
                CommandHistory.delete()
                .where(
                    (CommandHistory.alias == alias)
                    & (CommandHistory.id.not_in(keep_ids))
                )
                .execute()
            )
