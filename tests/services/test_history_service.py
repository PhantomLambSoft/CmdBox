import unittest
from unittest.mock import MagicMock

from cmdbox.models import CommandHistory
from cmdbox.services.errors import HistoryIndexError
from cmdbox.services.history_service import HistoryService


class TestHistoryService(unittest.TestCase):

    def setUp(self):
        self.repo = MagicMock()
        self.get_settings = MagicMock()
        self.mock_profile_repo = MagicMock()
        self.service = HistoryService(
            repo=self.repo,
            get_settings=self.get_settings,
            profile_repository=self.mock_profile_repo,
        )

    def _entry(self, entry_id: str = "history-id", variables_used: str | None = None):
        entry = MagicMock(spec=CommandHistory)
        entry.id = entry_id
        entry.variables_used = variables_used
        return entry

    def test_get_recent_delegates_to_repository(self):
        expected = [self._entry("1"), self._entry("2")]
        self.repo.get_recent.return_value = expected

        actual = self.service.get_recent(alias="deploy", limit=5)

        self.assertEqual(expected, actual)
        self.repo.get_recent.assert_called_once_with("deploy", 5, profile=None)

    def test_get_by_ref_with_numeric_ref_uses_index_lookup(self):
        expected = self._entry("abc123")
        self.repo.get_recent.return_value = [expected]

        actual = self.service.get_by_ref("1", alias="deploy")

        self.assertEqual(expected, actual)
        self.repo.get_recent.assert_called_once_with(
            alias="deploy", limit=1, profile=None
        )
        self.repo.get_by_id.assert_not_called()

    def test_get_by_ref_with_id_uses_repository_get_by_id(self):
        expected = self._entry("abc123")
        self.repo.get_by_id.return_value = expected

        actual = self.service.get_by_ref("abc123", alias="ignored")

        self.assertEqual(expected, actual)
        self.repo.get_by_id.assert_called_once_with("abc123", profile=None)
        self.repo.get_recent.assert_not_called()

    def test_get_by_index_raises_for_zero_index(self):
        with self.assertRaises(HistoryIndexError):
            self.service._get_by_index(0)

        self.repo.get_recent.assert_not_called()

    def test_get_by_index_raises_for_negative_index(self):
        with self.assertRaises(HistoryIndexError):
            self.service._get_by_index(-2)

        self.repo.get_recent.assert_not_called()

    def test_get_by_index_raises_when_index_exceeds_available_entries(self):
        self.repo.get_recent.return_value = [self._entry("1")]

        with self.assertRaises(HistoryIndexError):
            self.service._get_by_index(2, alias="ops")

        self.repo.get_recent.assert_called_once_with(alias="ops", limit=2, profile=None)

    def test_get_by_index_returns_matching_one_based_entry(self):
        first = self._entry("1")
        second = self._entry("2")
        self.repo.get_recent.return_value = [first, second]

        actual = self.service._get_by_index(2)

        self.assertEqual(second, actual)
        self.repo.get_recent.assert_called_once_with(alias=None, limit=2, profile=None)

    def test_get_variables_returns_none_when_variables_are_absent(self):
        entry = self._entry(variables_used=None)

        actual = self.service.get_variables(entry)

        self.assertIsNone(actual)

    def test_get_variables_parses_valid_json(self):
        entry = self._entry(variables_used='{"name": "Homer", "env": "prod"}')

        actual = self.service.get_variables(entry)

        self.assertEqual({"name": "Homer", "env": "prod"}, actual)

    def test_get_variables_raises_for_invalid_json_payload(self):
        entry = self._entry(variables_used="not-json")

        with self.assertRaises(ValueError):
            self.service.get_variables(entry)

    def test_delete_by_ref_resolves_entry_and_deletes_by_id(self):
        self.repo.get_by_id.return_value = self._entry("to-delete")
        self.repo.delete_by_id.return_value = True

        actual = self.service.delete_by_ref(
            "to-delete",
        )

        self.assertEqual(True, actual)
        self.repo.get_by_id.assert_called_once_with("to-delete", profile=None)
        self.repo.delete_by_id.assert_called_once_with("to-delete", profile=None)

    def test_delete_by_ref_with_numeric_reference_deletes_index_entry(self):
        self.repo.get_recent.return_value = [self._entry("index-entry")]
        self.repo.delete_by_id.return_value = True

        actual = self.service.delete_by_ref(
            "1",
            profile=None,
        )

        self.assertEqual(True, actual)
        self.repo.get_recent.assert_called_once_with(alias=None, limit=1, profile=None)
        self.repo.delete_by_id.assert_called_once_with("index-entry", profile=None)

    def test_clear_delegates_to_repository_with_alias(self):
        self.repo.clear.return_value = 3

        actual = self.service.clear(
            alias="deploy",
            profile=None,
        )

        self.assertEqual(3, actual)
        self.repo.clear.assert_called_once_with(alias="deploy", profile=None)

    def test_clear_delegates_to_repository_without_alias(self):
        self.repo.clear.return_value = 12

        actual = self.service.clear(
            profile=None,
        )

        self.assertEqual(12, actual)
        self.repo.clear.assert_called_once_with(alias=None, profile=None)
