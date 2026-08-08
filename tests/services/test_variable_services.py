import unittest
from unittest.mock import MagicMock, patch

from cmdbox.repositories.errors import UnknownNameError
from cmdbox.services.variable_services import VariableServices
from cmdbox.models import Variable, Tag, Profile
from cmdbox.repositories.results import TagAttachResult, TagDetachResult


class TestVariableServices(unittest.TestCase):

    def setUp(self):
        self.mock_repo = MagicMock()
        self.mock_tag_repo = MagicMock()
        self.mock_profile_repo = MagicMock()
        self.services = VariableServices(
            self.mock_repo, self.mock_tag_repo, self.mock_profile_repo
        )

    @patch("cmdbox.services.variable_services.db")
    def test_create_variable_without_tags(self, mock_db):
        name = "test-var"
        value = "test-value"
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.create.return_value = expected_var

        result = self.services.create_variable(name, value)

        self.assertEqual(expected_var, result)
        self.mock_repo.create.assert_called_once_with(
            name=name, value=value, profile=None
        )
        self.mock_repo.add_tags.assert_not_called()
        mock_db.atomic.assert_called_once()

    @patch("cmdbox.services.variable_services.db")
    def test_create_variable_with_tags(self, mock_db):
        name = "test-var"
        value = "test-value"
        tag_names = ["tag1", "tag2"]
        tags = [MagicMock(spec=Tag), MagicMock(spec=Tag)]
        expected_var = MagicMock(spec=Variable)

        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.create.return_value = expected_var

        result = self.services.create_variable(name, value, tags=tag_names)

        self.assertEqual(expected_var, result)
        self.assertEqual(2, self.mock_tag_repo.get_by_name.call_count)
        self.mock_repo.create.assert_called_once_with(
            name=name, value=value, profile=None
        )
        self.mock_repo.add_tags.assert_called_once_with(expected_var, tags)
        mock_db.atomic.assert_called_once()

    @patch("cmdbox.services.variable_services.db")
    def test_create_variable_with_different_profile(self, mock_db):
        name = "test-var"
        value = "test-value"
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.create.return_value = expected_var
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        result = self.services.create_variable(name, value, profile=mock_profile)

        self.assertEqual(expected_var, result)
        self.mock_repo.create.assert_called_once_with(
            name=name, value=value, profile=mock_profile
        )
        self.mock_repo.add_tags.assert_not_called()
        mock_db.atomic.assert_called_once()

    def test_update_variable(self):
        name = "test-var"
        var = MagicMock(spec=Variable)
        updated_var = MagicMock(spec=Variable)
        self.mock_repo.get_by_name.return_value = var
        self.mock_repo.update.return_value = updated_var

        result = self.services.update_variable(name, value="new")

        self.assertEqual(updated_var, result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.update.assert_called_once_with(var, value="new")

    def test_delete_variable(self):
        name = "test-var"
        var = MagicMock(spec=Variable)
        self.mock_repo.get_by_name.return_value = var
        self.mock_repo.delete.return_value = True

        result = self.services.delete_variable(name)

        self.assertTrue(result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=None)
        self.mock_repo.delete.assert_called_once_with(var)

    def test_delete_variable_on_different_profile(self):
        name = "test-var"
        var = MagicMock(spec=Variable)
        self.mock_repo.get_by_name.return_value = var
        self.mock_repo.delete.return_value = True
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        result = self.services.delete_variable(name, profile=mock_profile)

        self.assertTrue(result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=mock_profile)
        self.mock_repo.delete.assert_called_once_with(var)

    def test_add_tags(self):
        name = "test-var"
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        var = MagicMock(spec=Variable)
        expected_result = TagAttachResult(added=["tag1"], existing=[])

        self.mock_repo.get_by_name.return_value = var
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.add_tags.return_value = expected_result

        result = self.services.add_tags(name, tag_names)

        self.assertEqual(expected_result, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=None)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.add_tags.assert_called_once_with(var, tags)

    def test_add_tags_from_different_profile(self):
        name = "test-var"
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        var = MagicMock(spec=Variable)
        expected_result = TagAttachResult(added=["tag1"], existing=[])

        self.mock_repo.get_by_name.return_value = var
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.add_tags.return_value = expected_result
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        result = self.services.add_tags(name, tag_names, profile=mock_profile)

        self.assertEqual(expected_result, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=mock_profile)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.add_tags.assert_called_once_with(var, tags)

    def test_remove_tags(self):
        name = "test-var"
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        var = MagicMock(spec=Variable)
        expected_result = TagDetachResult(removed=["tag1"], not_attached=[])

        self.mock_repo.get_by_name.return_value = var
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.remove_tags.return_value = expected_result

        result = self.services.remove_tags(name, tag_names)

        self.assertEqual(expected_result, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=None)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.remove_tags.assert_called_once_with(var, tags)

    def test_remove_tags_from_different_profile(self):
        name = "test-var"
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        var = MagicMock(spec=Variable)
        expected_result = TagDetachResult(removed=["tag1"], not_attached=[])

        self.mock_repo.get_by_name.return_value = var
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.remove_tags.return_value = expected_result
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        result = self.services.remove_tags(name, tag_names, profile=mock_profile)

        self.assertEqual(expected_result, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=mock_profile)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.remove_tags.assert_called_once_with(var, tags)

    def test_get_variable(self):
        name = "test-var"
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.get_by_name.return_value = expected_var

        result = self.services.get_variable(name)

        self.assertEqual(expected_var, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=None)

    def test_get_variable_from_different_profile(self):
        name = "test-var"
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.get_by_name.return_value = expected_var
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        result = self.services.get_variable(name, profile=mock_profile)

        self.assertEqual(expected_var, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=mock_profile)

    def test_get_variable_or_none(self):
        name = "test-var"
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.get_by_name.return_value = expected_var

        result = self.services.get_variable_or_none(name)

        self.assertEqual(expected_var, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=None)

    def test_get_variable_or_none_returns_existing_variable_from_different_profile(
        self,
    ):
        name = "test-var"
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.get_by_name.return_value = expected_var
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        result = self.services.get_variable_or_none(name, profile=mock_profile)

        self.assertEqual(expected_var, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=mock_profile)

    def test_get_variable_or_none_returns_none_if_not_found(self):
        name = "non-existant-test-var"
        self.mock_repo.get_by_name.side_effect = UnknownNameError(name)

        result = self.services.get_variable_or_none(name)

        self.assertEqual(None, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=None)

    def test_get_variable_or_none_returns_none_if_not_found_on_different_profile(self):
        name = "non-existant-test-var"
        self.mock_repo.get_by_name.side_effect = UnknownNameError(name)
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        result = self.services.get_variable_or_none(name, profile=mock_profile)

        self.assertEqual(None, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=mock_profile)

    def test_get_variable_by_id(self):
        id_ = 123
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.get_by_id.return_value = expected_var

        result = self.services.get_variable_by_id(id_)

        self.assertEqual(expected_var, result)
        self.mock_repo.get_by_id.assert_called_once_with(id_)

    def test_list_variables_no_tags(self):
        expected_vars = [MagicMock(spec=Variable), MagicMock(spec=Variable)]
        self.mock_repo.list_all.return_value = expected_vars

        result = self.services.list_variables(order_by="name", limit=10)

        self.assertEqual(expected_vars, result)
        self.mock_repo.list_all.assert_called_once_with("name", 10, profile=None)
        self.mock_repo.list_by_tag.assert_not_called()

    def test_list_variables_no_tags_from_different_profile(self):
        expected_vars = [MagicMock(spec=Variable), MagicMock(spec=Variable)]
        self.mock_repo.list_all.return_value = expected_vars
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        result = self.services.list_variables(
            order_by="name", limit=10, profile=mock_profile
        )

        self.assertEqual(expected_vars, result)
        self.mock_repo.list_all.assert_called_once_with(
            "name", 10, profile=mock_profile
        )
        self.mock_repo.list_by_tag.assert_not_called()

    def test_list_variables_with_tags(self):
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        expected_vars = [MagicMock(spec=Variable)]
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.list_by_tag.return_value = expected_vars

        result = self.services.list_variables(tags=tag_names, order_by="name", limit=10)

        self.assertEqual(expected_vars, result)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.list_by_tag.assert_called_once_with(
            tags, "name", 10, profile=None
        )
        self.mock_repo.list_all.assert_not_called()

    def test_list_variables_with_tags_from_different_profile(self):
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        expected_vars = [MagicMock(spec=Variable)]
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.list_by_tag.return_value = expected_vars
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        result = self.services.list_variables(
            tags=tag_names, order_by="name", limit=10, profile=mock_profile
        )

        self.assertEqual(expected_vars, result)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.list_by_tag.assert_called_once_with(
            tags, "name", 10, profile=mock_profile
        )
        self.mock_repo.list_all.assert_not_called()

    def test_search(self):
        query = "hello"
        fields = ["name"]
        expected_vars = [MagicMock(spec=Variable)]
        self.mock_repo.search.return_value = expected_vars

        result = self.services.search(query, fields=fields, limit=10)

        self.assertEqual(expected_vars, result)
        self.mock_repo.search.assert_called_once_with(
            query, fields=fields, limit=10, profile=None
        )

    def test_search_in_different_profile(self):
        query = "hello"
        fields = ["name"]
        expected_vars = [MagicMock(spec=Variable)]
        self.mock_repo.search.return_value = expected_vars
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        result = self.services.search(
            query, fields=fields, limit=10, profile=mock_profile
        )

        self.assertEqual(expected_vars, result)
        self.mock_repo.search.assert_called_once_with(
            query, fields=fields, limit=10, profile=mock_profile
        )

    def test_move_variable(self):
        name = "test-var"
        target_profile_name = "target-profile"
        source_profile = MagicMock(spec=Profile)
        target_profile = MagicMock(spec=Profile)
        var = MagicMock(spec=Variable)
        updated_var = MagicMock(spec=Variable)

        self.mock_profile_repo.get_by_name.side_effect = [
            source_profile,
            target_profile,
        ]
        self.mock_repo.get_by_name.return_value = var
        self.mock_repo.update.return_value = updated_var

        result = self.services.move_variable(
            name, target_profile_name, profile="source-profile"
        )

        self.assertEqual(updated_var, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=source_profile)
        self.mock_profile_repo.get_by_name.assert_any_call("source-profile")
        self.mock_profile_repo.get_by_name.assert_any_call(target_profile_name)
        self.mock_repo.update.assert_called_once_with(var, profile=target_profile)

    @patch("cmdbox.services.variable_services.db")
    def test_copy_variable(self, mock_db):
        name = "test-var"
        target_profile_name = "target-profile"
        source_profile = MagicMock(spec=Profile)
        target_profile = MagicMock(spec=Profile)
        source_tag = MagicMock(spec=Tag)
        source_var = MagicMock(spec=Variable)
        source_var.name = name
        source_var.value = "test-value"
        source_var.tags = [MagicMock(tag=source_tag)]

        copy_var = MagicMock(spec=Variable)
        copy_var.id = 456
        expected_var = MagicMock(spec=Variable)

        self.mock_profile_repo.get_by_name.side_effect = [
            source_profile,
            target_profile,
        ]
        self.mock_repo.get_by_name.return_value = source_var
        self.mock_repo.create.return_value = copy_var
        self.mock_repo.get_by_id.return_value = expected_var

        result = self.services.copy_variable(
            name, target_profile_name, profile="source-profile"
        )

        self.assertEqual(expected_var, result)
        self.mock_repo.get_by_name.assert_called_once_with(name, profile=source_profile)
        self.mock_repo.create.assert_called_once_with(
            name=name, value=source_var.value, profile=target_profile
        )
        self.mock_repo.add_tags.assert_called_once_with(copy_var, [source_tag])
        self.mock_repo.get_by_id.assert_called_once_with(copy_var.id)
        mock_db.atomic.assert_called_once()

    def test_get_tags_internal(self):
        tag_names = ["tag1", "tag2"]
        tags = [MagicMock(spec=Tag), MagicMock(spec=Tag)]
        self.mock_tag_repo.get_by_name.side_effect = tags

        result = self.services._get_tags(tag_names)

        self.assertEqual(tags, result)
        self.assertEqual(2, self.mock_tag_repo.get_by_name.call_count)

    def test_get_tags_internal_none(self):
        result = self.services._get_tags(None)

        self.assertEqual([], result)
        self.mock_tag_repo.get_by_name.assert_not_called()

    def test_get_tags_internal_empty(self):
        result = self.services._get_tags([])

        self.assertEqual([], result)
        self.mock_tag_repo.get_by_name.assert_not_called()

    @patch("cmdbox.services.variable_services.db")
    def test_create_variable_empty_tags(self, mock_db):
        name = "test-var"
        value = "test-value"
        expected_var = MagicMock(spec=Variable)
        self.mock_repo.create.return_value = expected_var

        result = self.services.create_variable(name, value, tags=[])

        self.assertEqual(expected_var, result)
        self.mock_repo.create.assert_called_once_with(
            name=name, value=value, profile=None
        )
        self.mock_repo.add_tags.assert_not_called()
        mock_db.atomic.assert_called_once()
