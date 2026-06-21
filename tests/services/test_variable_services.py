import unittest
from unittest.mock import MagicMock, patch

from cmdbox.repositories.errors import UnknownNameError
from cmdbox.services.variable_services import VariableServices
from cmdbox.models import Variable, Tag
from cmdbox.repositories.results import TagAttachResult, TagDetachResult


class TestVariableServices(unittest.TestCase):
    def setUp(self):
        self.mock_repo = MagicMock()
        self.mock_tag_repo = MagicMock()
        self.services = VariableServices(self.mock_repo, self.mock_tag_repo)

    @patch("cmdbox.services.variable_services.db")
    def test_create_variable_without_tags(self, mock_db):
        # Setup
        name = "test-var"
        value = "test-value"
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.create.return_value = expected_var

        # Execute
        result = self.services.create_variable(name, value)

        # Assert
        self.assertEqual(result, expected_var)
        self.mock_repo.create.assert_called_once_with(name=name, value=value)
        self.mock_repo.add_tags.assert_not_called()
        mock_db.atomic.assert_called_once()

    @patch("cmdbox.services.variable_services.db")
    def test_create_variable_with_tags(self, mock_db):
        # Setup
        name = "test-var"
        value = "test-value"
        tag_names = ["tag1", "tag2"]
        tags = [MagicMock(spec=Tag), MagicMock(spec=Tag)]
        expected_var = MagicMock(spec=Variable)

        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.create.return_value = expected_var

        # Execute
        result = self.services.create_variable(name, value, tags=tag_names)

        # Assert
        self.assertEqual(result, expected_var)
        self.assertEqual(self.mock_tag_repo.get_by_name.call_count, 2)
        self.mock_repo.create.assert_called_once_with(name=name, value=value)
        self.mock_repo.add_tags.assert_called_once_with(expected_var, tags)
        mock_db.atomic.assert_called_once()

    def test_update_variable(self):
        # Setup
        name = "test-var"
        var = MagicMock(spec=Variable)
        updated_var = MagicMock(spec=Variable)
        self.mock_repo.get_by_name.return_value = var
        self.mock_repo.update.return_value = updated_var

        # Execute
        result = self.services.update_variable(name, value="new")

        # Assert
        self.assertEqual(result, updated_var)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.update.assert_called_once_with(var, value="new")

    def test_delete_variable(self):
        # Setup
        name = "test-var"
        var = MagicMock(spec=Variable)
        self.mock_repo.get_by_name.return_value = var
        self.mock_repo.delete.return_value = True

        # Execute
        result = self.services.delete_variable(name)

        # Assert
        self.assertTrue(result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.delete.assert_called_once_with(var)

    def test_add_tags(self):
        # Setup
        name = "test-var"
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        var = MagicMock(spec=Variable)
        expected_result = TagAttachResult(added=["tag1"], existing=[])

        self.mock_repo.get_by_name.return_value = var
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.add_tags.return_value = expected_result

        # Execute
        result = self.services.add_tags(name, tag_names)

        # Assert
        self.assertEqual(result, expected_result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.add_tags.assert_called_once_with(var, tags)

    def test_remove_tags(self):
        # Setup
        name = "test-var"
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        var = MagicMock(spec=Variable)
        expected_result = TagDetachResult(removed=["tag1"], not_attached=[])

        self.mock_repo.get_by_name.return_value = var
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.remove_tags.return_value = expected_result

        # Execute
        result = self.services.remove_tags(name, tag_names)

        # Assert
        self.assertEqual(result, expected_result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.remove_tags.assert_called_once_with(var, tags)

    def test_get_variable(self):
        # Setup
        name = "test-var"
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.get_by_name.return_value = expected_var

        # Execute
        result = self.services.get_variable(name)

        # Assert
        self.assertEqual(result, expected_var)
        self.mock_repo.get_by_name.assert_called_once_with(name)

    def test_get_variable_or_none(self):
        name = "test-var"
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.get_by_name.return_value = expected_var

        result = self.services.get_variable_or_none(name)

        self.assertEqual(result, expected_var)
        self.mock_repo.get_by_name.assert_called_once_with(name)

    def test_get_variable_or_none_returns_none_if_not_found(self):
        name = "non-existant-test-var"
        self.mock_repo.get_by_name.side_effect = UnknownNameError(name)

        result = self.services.get_variable_or_none(name)

        self.assertEqual(result, None)
        self.mock_repo.get_by_name.assert_called_once_with(name)

    def test_get_variable_by_id(self):
        id_ = 123
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.get_by_id.return_value = expected_var

        result = self.services.get_variable_by_id(id_)

        self.assertEqual(expected_var, result)
        self.mock_repo.get_by_id.assert_called_once_with(id_)

    def test_list_variables_no_tags(self):
        # Setup
        expected_vars = [MagicMock(spec=Variable), MagicMock(spec=Variable)]
        self.mock_repo.list_all.return_value = expected_vars

        # Execute
        result = self.services.list_variables(order_by="name", limit=10)

        # Assert
        self.assertEqual(result, expected_vars)
        self.mock_repo.list_all.assert_called_once_with("name", 10)
        self.mock_repo.list_by_tag.assert_not_called()

    def test_list_variables_with_tags(self):
        # Setup
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        expected_vars = [MagicMock(spec=Variable)]
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.list_by_tag.return_value = expected_vars

        # Execute
        result = self.services.list_variables(tags=tag_names, order_by="name", limit=10)

        # Assert
        self.assertEqual(result, expected_vars)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.list_by_tag.assert_called_once_with(tags, "name", 10)
        self.mock_repo.list_all.assert_not_called()

    def test_search(self):
        # Setup
        query = "hello"
        fields = ["name"]
        expected_vars = [MagicMock(spec=Variable)]
        self.mock_repo.search.return_value = expected_vars

        # Execute
        result = self.services.search(query, fields=fields, limit=10)

        # Assert
        self.assertEqual(expected_vars, result)
        self.mock_repo.search.assert_called_once_with(query, fields=fields, limit=10)

    def test_get_tags_internal(self):
        # Setup
        tag_names = ["tag1", "tag2"]
        tags = [MagicMock(spec=Tag), MagicMock(spec=Tag)]
        self.mock_tag_repo.get_by_name.side_effect = tags

        # Execute
        result = self.services._get_tags(tag_names)

        # Assert
        self.assertEqual(result, tags)
        self.assertEqual(self.mock_tag_repo.get_by_name.call_count, 2)

    def test_get_tags_internal_none(self):
        # Execute
        result = self.services._get_tags(None)

        # Assert
        self.assertEqual(result, [])
        self.mock_tag_repo.get_by_name.assert_not_called()

    def test_get_tags_internal_empty(self):
        # Execute
        result = self.services._get_tags([])

        # Assert
        self.assertEqual(result, [])
        self.mock_tag_repo.get_by_name.assert_not_called()

    @patch("cmdbox.services.variable_services.db")
    def test_create_variable_empty_tags(self, mock_db):
        # Setup
        name = "test-var"
        value = "test-value"
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.create.return_value = expected_var

        # Execute
        result = self.services.create_variable(name, value, tags=[])

        # Assert
        self.assertEqual(result, expected_var)
        self.mock_repo.create.assert_called_once_with(name=name, value=value)
        self.mock_repo.add_tags.assert_not_called()
        mock_db.atomic.assert_called_once()
