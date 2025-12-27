import unittest
from unittest.mock import MagicMock, patch
from cmdbox.services.command_services import CommandServices
from cmdbox.models import Command, Tag
from cmdbox.repositories.results import TagAttachResult, TagDetachResult


class TestCommandServices(unittest.TestCase):

    def setUp(self):
        self.mock_repo = MagicMock()
        self.mock_tag_repo = MagicMock()
        self.services = CommandServices(self.mock_repo, self.mock_tag_repo)

    @patch("cmdbox.services.command_services.db")
    def test_create_command_without_tags(self, mock_db):
        # Setup
        alias = "test-cmd"
        template = "echo hello"
        description = "A test command"
        expected_cmd = MagicMock(spec=Command)
        self.mock_repo.create.return_value = expected_cmd

        # Execute
        result = self.services.create_command(alias, template, description)

        # Assert
        self.assertEqual(result, expected_cmd)
        self.mock_repo.create.assert_called_once_with(
            alias=alias, template=template, description=description
        )
        self.mock_repo.add_tags.assert_not_called()
        mock_db.atomic.assert_called_once()

    @patch("cmdbox.services.command_services.db")
    def test_create_command_with_tags(self, mock_db):
        # Setup
        alias = "test-cmd"
        template = "echo hello"
        tag_names = ["tag1", "tag2"]
        tags = [MagicMock(spec=Tag), MagicMock(spec=Tag)]
        expected_cmd = MagicMock(spec=Command)

        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.create.return_value = expected_cmd

        # Execute
        result = self.services.create_command(alias, template, tags=tag_names)

        # Assert
        self.assertEqual(result, expected_cmd)
        self.assertEqual(self.mock_tag_repo.get_by_name.call_count, 2)
        self.mock_repo.create.assert_called_once_with(
            alias=alias, template=template, description=None
        )
        self.mock_repo.add_tags.assert_called_once_with(expected_cmd, tags)
        mock_db.atomic.assert_called_once()

    def test_update_command(self):
        # Setup
        alias = "test-cmd"
        cmd = MagicMock(spec=Command)
        updated_cmd = MagicMock(spec=Command)
        self.mock_repo.get_by_alias.return_value = cmd
        self.mock_repo.update.return_value = updated_cmd

        # Execute
        result = self.services.update_command(alias, template="new")

        # Assert
        self.assertEqual(result, updated_cmd)
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
        self.mock_repo.update.assert_called_once_with(cmd, template="new")

    def test_delete_command(self):
        # Setup
        alias = "test-cmd"
        cmd = MagicMock(spec=Command)
        self.mock_repo.get_by_alias.return_value = cmd
        self.mock_repo.delete.return_value = True

        # Execute
        result = self.services.delete_command(alias)

        # Assert
        self.assertTrue(result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
        self.mock_repo.delete.assert_called_once_with(cmd)

    def test_add_tags(self):
        # Setup
        alias = "test-cmd"
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        cmd = MagicMock(spec=Command)
        expected_result = TagAttachResult(added=["tag1"], existing=[])

        self.mock_repo.get_by_alias.return_value = cmd
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.add_tags.return_value = expected_result

        # Execute
        result = self.services.add_tags(alias, tag_names)

        # Assert
        self.assertEqual(result, expected_result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.add_tags.assert_called_once_with(cmd, tags)

    def test_remove_tags(self):
        # Setup
        alias = "test-cmd"
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        cmd = MagicMock(spec=Command)
        expected_result = TagDetachResult(removed=["tag1"], not_attached=[])

        self.mock_repo.get_by_alias.return_value = cmd
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.remove_tags.return_value = expected_result

        # Execute
        result = self.services.remove_tags(alias, tag_names)

        # Assert
        self.assertEqual(result, expected_result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.remove_tags.assert_called_once_with(cmd, tags)

    def test_get_command(self):
        # Setup
        alias = "test-cmd"
        expected_cmd = MagicMock(spec=Command)
        self.mock_repo.get_by_alias.return_value = expected_cmd

        # Execute
        result = self.services.get_command(alias)

        # Assert
        self.assertEqual(result, expected_cmd)
        self.mock_repo.get_by_alias.assert_called_once_with(alias)

    def test_list_commands_no_tags(self):
        # Setup
        expected_commands = [MagicMock(spec=Command), MagicMock(spec=Command)]
        self.mock_repo.list_all.return_value = expected_commands

        # Execute
        result = self.services.list_commands(order_by="alias", limit=10)

        # Assert
        self.assertEqual(result, expected_commands)
        self.mock_repo.list_all.assert_called_once_with("alias", 10)
        self.mock_repo.list_by_tag.assert_not_called()

    def test_list_commands_with_tags(self):
        # Setup
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        expected_commands = [MagicMock(spec=Command)]
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.list_by_tag.return_value = expected_commands

        # Execute
        result = self.services.list_commands(tags=tag_names, order_by="alias", limit=10)

        # Assert
        self.assertEqual(result, expected_commands)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.list_by_tag.assert_called_once_with(tags, "alias", 10)
        self.mock_repo.list_all.assert_not_called()

    def test_search(self):
        # Setup
        query = "hello"
        fields = ["alias"]
        expected_commands = [MagicMock(spec=Command)]
        self.mock_repo.search.return_value = expected_commands

        # Execute
        result = self.services.search(query, fields)

        # Assert
        self.assertEqual(result, expected_commands)
        self.mock_repo.search.assert_called_once_with(query, fields)

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

    @patch("cmdbox.services.command_services.db")
    def test_create_command_empty_tags(self, mock_db):
        # Setup
        alias = "test-cmd"
        template = "echo hello"
        expected_cmd = MagicMock(spec=Command)
        self.mock_repo.create.return_value = expected_cmd

        # Execute
        result = self.services.create_command(alias, template, tags=[])

        # Assert
        self.assertEqual(result, expected_cmd)
        self.mock_repo.create.assert_called_once_with(
            alias=alias, template=template, description=None
        )
        self.mock_repo.add_tags.assert_not_called()
        mock_db.atomic.assert_called_once()
