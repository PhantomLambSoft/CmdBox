import unittest
from datetime import datetime, timedelta

from cmdbox.database import db, ensure_schema, get_db
from cmdbox.models import ALL_MODELS, CommandHistory
from cmdbox.repositories.errors import (
    AmbiguousHistoryIdEntryError,
    UnknownHistoryEntryError,
)
from cmdbox.repositories.history_repository import HistoryRepository


class TestHistoryRepository(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        get_db(testing=True)
        ensure_schema()

    @classmethod
    def tearDownClass(cls):
        db.drop_tables(ALL_MODELS)
        db.close()

    def setUp(self):
        CommandHistory.delete().execute()
        self.repo = HistoryRepository()

    def _create_entry(
        self,
        history_id: str,
        alias: str,
        ran_at: datetime,
        template: str = "echo hi",
        resolved: str = "echo hi",
    ) -> CommandHistory:
        return CommandHistory.create(
            id=history_id,
            alias=alias,
            template=template,
            resolved=resolved,
            variables_used=None,
            exit_code=0,
            ran_at=ran_at,
        )

    def test_record_creates_history_entry_with_serialized_variables(self):
        entry = self.repo.record(
            alias="deploy",
            template="echo <name>",
            resolved="echo Homer",
            variables_used={"name": "Homer"},
            exit_code=0,
            limit=10,
        )

        persisted = CommandHistory.get_by_id(entry.id)
        self.assertEqual("deploy", persisted.alias)
        self.assertEqual('{"name": "Homer"}', persisted.variables_used)
        self.assertEqual(1, CommandHistory.select().count())

    def test_record_stores_none_for_empty_or_missing_variables(self):
        empty_vars_entry = self.repo.record(
            alias="deploy",
            template="echo",
            resolved="echo",
            variables_used={},
            exit_code=None,
            limit=10,
        )
        none_vars_entry = self.repo.record(
            alias="deploy",
            template="echo",
            resolved="echo",
            variables_used=None,
            exit_code=None,
            limit=10,
        )

        persisted_empty = CommandHistory.get_by_id(empty_vars_entry.id)
        persisted_none = CommandHistory.get_by_id(none_vars_entry.id)
        self.assertIsNone(persisted_empty.variables_used)
        self.assertIsNone(persisted_none.variables_used)

    def test_record_with_none_limit_keeps_entries_and_does_not_raise(self):
        first = self.repo.record(
            alias="deploy",
            template="echo 1",
            resolved="echo 1",
            variables_used=None,
            exit_code=0,
            limit=None,
        )
        second = self.repo.record(
            alias="deploy",
            template="echo 2",
            resolved="echo 2",
            variables_used=None,
            exit_code=0,
            limit=None,
        )

        ids = {row.id for row in CommandHistory.select()}
        self.assertEqual({first.id, second.id}, ids)

    def test_get_by_id_returns_exact_match(self):
        entry = self._create_entry(
            history_id="abc1234567890",
            alias="build",
            ran_at=datetime.now(),
        )

        actual = self.repo.get_by_id(entry.id)

        self.assertEqual(entry.id, actual.id)

    def test_get_by_id_returns_single_prefix_match(self):
        entry = self._create_entry(
            history_id="ff0011223344",
            alias="build",
            ran_at=datetime.now(),
        )

        actual = self.repo.get_by_id("ff00")

        self.assertEqual(entry.id, actual.id)

    def test_get_by_id_raises_unknown_error_when_not_found(self):
        with self.assertRaises(UnknownHistoryEntryError):
            self.repo.get_by_id("missing")

    def test_get_by_id_raises_ambiguous_error_for_multiple_prefix_matches(self):
        now = datetime.now()
        self._create_entry("aa001111", "build", now)
        self._create_entry("aa002222", "build", now + timedelta(seconds=1))

        with self.assertRaises(AmbiguousHistoryIdEntryError):
            self.repo.get_by_id("aa")

    def test_get_recent_returns_entries_in_descending_ran_at_order(self):
        now = datetime.now()
        older = self._create_entry("old", "build", now - timedelta(minutes=2))
        newest = self._create_entry("new", "build", now)
        middle = self._create_entry("mid", "build", now - timedelta(minutes=1))

        actual = self.repo.get_recent()

        self.assertEqual([newest.id, middle.id, older.id], [row.id for row in actual])

    def test_get_recent_applies_alias_filter(self):
        now = datetime.now()
        self._create_entry("deploy-1", "deploy", now)
        self._create_entry("build-1", "build", now + timedelta(seconds=1))

        actual = self.repo.get_recent(alias="deploy")

        self.assertEqual(["deploy-1"], [row.id for row in actual])

    def test_get_recent_applies_positive_limit(self):
        now = datetime.now()
        self._create_entry("1", "deploy", now - timedelta(minutes=2))
        second = self._create_entry("2", "deploy", now - timedelta(minutes=1))
        third = self._create_entry("3", "deploy", now)

        actual = self.repo.get_recent(limit=2)

        self.assertEqual([third.id, second.id], [row.id for row in actual])

    def test_get_recent_with_zero_or_negative_limit_returns_all(self):
        now = datetime.now()
        first = self._create_entry("1", "deploy", now - timedelta(minutes=1))
        second = self._create_entry("2", "deploy", now)

        zero_limit = self.repo.get_recent(limit=0)
        negative_limit = self.repo.get_recent(limit=-5)

        self.assertEqual([second.id, first.id], [row.id for row in zero_limit])
        self.assertEqual([second.id, first.id], [row.id for row in negative_limit])

    def test_delete_by_id_deletes_entry_and_returns_true(self):
        entry = self._create_entry("to-delete", "deploy", datetime.now())

        actual = self.repo.delete_by_id(entry.id)

        self.assertEqual(True, actual)
        self.assertEqual(0, CommandHistory.select().count())

    def test_delete_by_id_raises_unknown_when_entry_does_not_exist(self):
        with self.assertRaises(UnknownHistoryEntryError):
            self.repo.delete_by_id("not-there")

    def test_clear_without_alias_deletes_all_entries(self):
        now = datetime.now()
        self._create_entry("deploy-1", "deploy", now)
        self._create_entry("build-1", "build", now + timedelta(seconds=1))

        deleted_count = self.repo.clear()

        self.assertEqual(2, deleted_count)
        self.assertEqual(0, CommandHistory.select().count())

    def test_clear_with_alias_deletes_only_matching_entries(self):
        now = datetime.now()
        self._create_entry("deploy-1", "deploy", now)
        self._create_entry("build-1", "build", now + timedelta(seconds=1))

        deleted_count = self.repo.clear(alias="deploy")

        self.assertEqual(1, deleted_count)
        self.assertEqual(["build-1"], [row.id for row in CommandHistory.select()])

    def test_apply_retention_keeps_only_latest_n_for_alias(self):
        now = datetime.now()
        oldest = self._create_entry("keep-1", "deploy", now - timedelta(minutes=3))
        middle = self._create_entry("keep-2", "deploy", now - timedelta(minutes=2))
        newest = self._create_entry("keep-3", "deploy", now - timedelta(minutes=1))
        self._create_entry("other-alias", "build", now)

        self.repo._apply_retention(alias="deploy", limit=2)

        self.assertEqual(
            {middle.id, newest.id, "other-alias"},
            {row.id for row in CommandHistory.select()},
        )
        self.assertEqual(
            0, CommandHistory.select().where(CommandHistory.id == oldest.id).count()
        )

    def test_apply_retention_ignores_zero_negative_and_none_limits(self):
        now = datetime.now()
        first = self._create_entry("a", "deploy", now - timedelta(minutes=1))
        second = self._create_entry("b", "deploy", now)

        self.repo._apply_retention(alias="deploy", limit=0)
        self.repo._apply_retention(alias="deploy", limit=-1)
        self.repo._apply_retention(alias="deploy", limit=None)

        self.assertEqual(
            {first.id, second.id}, {row.id for row in CommandHistory.select()}
        )
