import unittest
from unittest.mock import MagicMock, patch

from cmdbox.repositories.errors import UnknownAliasError
from cmdbox.services.command_services import CommandServices
from cmdbox.models import Command, Tag, ALL_MODELS, Profile
from cmdbox.repositories.results import TagAttachResult, TagDetachResult
from cmdbox.database import get_db, ensure_schema, db


class TestCommandServices(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        get_db(testing=True)
        ensure_schema()

    @classmethod
    def tearDownClass(cls):
        db.drop_tables(ALL_MODELS)
        db.close()

    def setUp(self):
        self.mock_repo = MagicMock()
        self.mock_tag_repo = MagicMock()
        self.mock_profile_repo = MagicMock()
        self.services = CommandServices(
            self.mock_repo, self.mock_tag_repo, self.mock_profile_repo
        )
        get_db(testing=True)
        ensure_schema()

    @patch("cmdbox.services.command_services.db")
    def test_create_command_without_tags(self, mock_db):
        # Setup
        alias = "test-cmd"
        template = "echo hello"
        description = "A test command"
        expected_cmd = MagicMock(spec=Command)
        self.mock_repo.create.return_value = expected_cmd
        self.mock_repo.get_by_id.return_value = expected_cmd

        # Execute
        result = self.services.create_command(alias, template, description)

        # Assert
        self.assertEqual(expected_cmd, result)
        self.mock_repo.create.assert_called_once_with(
            alias=alias,
            template=template,
            description=description,
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
            profile=None,
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
        self.mock_repo.get_by_id.return_value = expected_cmd

        # Execute
        result = self.services.create_command(alias, template, tags=tag_names)

        # Assert
        self.assertEqual(expected_cmd, result)
        self.assertEqual(self.mock_tag_repo.get_by_name.call_count, 2)
        self.mock_repo.create.assert_called_once_with(
            alias=alias,
            template=template,
            description=None,
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
            profile=None,
        )
        self.mock_repo.add_tags.assert_called_once_with(expected_cmd, tags)
        mock_db.atomic.assert_called_once()

    @patch("cmdbox.services.command_services.db")
    def test_create_command_with_different_profile(self, mock_db):
        # Setup
        alias = "test-cmd"
        template = "echo hello"
        description = "A test command"
        expected_cmd = MagicMock(spec=Command)
        self.mock_repo.create.return_value = expected_cmd
        self.mock_repo.get_by_id.return_value = expected_cmd
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        # Execute
        result = self.services.create_command(
            alias, template, description, profile=mock_profile
        )

        # Assert
        self.assertEqual(expected_cmd, result)
        self.mock_repo.create.assert_called_once_with(
            alias=alias,
            template=template,
            description=description,
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
            profile=mock_profile,
        )
        self.mock_repo.add_tags.assert_not_called()
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
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_repo.delete.assert_called_once_with(cmd)

    def test_delete_command_on_different_profile(self):
        # Setup
        alias = "test-cmd"
        cmd = MagicMock(spec=Command)
        self.mock_repo.get_by_alias.return_value = cmd
        self.mock_repo.delete.return_value = True
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        # Execute
        result = self.services.delete_command(alias, profile=mock_profile)

        # Assert
        self.assertTrue(result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=mock_profile)
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
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.add_tags.assert_called_once_with(cmd, tags)

    def test_add_tags_from_different_profile(self):
        # Setup
        alias = "test-cmd"
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        cmd = MagicMock(spec=Command)
        expected_result = TagAttachResult(added=["tag1"], existing=[])

        self.mock_repo.get_by_alias.return_value = cmd
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.add_tags.return_value = expected_result
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        # Execute
        result = self.services.add_tags(alias, tag_names, profile=mock_profile)

        # Assert
        self.assertEqual(result, expected_result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=mock_profile)
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
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.remove_tags.assert_called_once_with(cmd, tags)

    def test_remove_tags_from_different_profile(self):
        # Setup
        alias = "test-cmd"
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        cmd = MagicMock(spec=Command)
        expected_result = TagDetachResult(removed=["tag1"], not_attached=[])

        self.mock_repo.get_by_alias.return_value = cmd
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.remove_tags.return_value = expected_result
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        # Execute
        result = self.services.remove_tags(alias, tag_names, profile=mock_profile)

        # Assert
        self.assertEqual(result, expected_result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=mock_profile)
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
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)

    def test_get_command_from_different_profile(self):
        # Setup
        alias = "test-cmd"
        expected_cmd = MagicMock(spec=Command)
        self.mock_repo.get_by_alias.return_value = expected_cmd
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        # Execute
        result = self.services.get_command(alias, profile=mock_profile)

        # Assert
        self.assertEqual(result, expected_cmd)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=mock_profile)

    def tets_get_command_or_none_returns_existing_command(self):
        alias = "test-cmd"
        expected_cmd = MagicMock(spec=Command)
        self.mock_repo.get_by_alias.return_value = expected_cmd

        # Execute
        result = self.services.get_command_or_none(alias)

        # Assert
        self.assertEqual(result, expected_cmd)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)

    def tets_get_command_or_none_returns_existing_command_from_different_profile(self):
        alias = "test-cmd"
        expected_cmd = MagicMock(spec=Command)
        self.mock_repo.get_by_alias.return_value = expected_cmd
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        # Execute
        result = self.services.get_command_or_none(alias, profile=mock_profile)

        # Assert
        self.assertEqual(result, expected_cmd)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=mock_profile)

    def test_get_command_or_none_returns_none_if_not_found(self):
        alias = "non-existant-test-cmd"
        self.mock_repo.get_by_alias.side_effect = UnknownAliasError(alias)

        # Execute
        result = self.services.get_command_or_none(alias)

        # Assert
        self.assertEqual(result, None)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)

    def test_get_command_or_none_returns_none_if_not_found_on_different_profile(self):
        alias = "non-existant-test-cmd"
        self.mock_repo.get_by_alias.side_effect = UnknownAliasError(alias)
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        # Execute
        result = self.services.get_command_or_none(alias, profile=mock_profile)

        # Assert
        self.assertEqual(result, None)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=mock_profile)

    def test_get_command_by_id(self):
        id_ = 123
        expected_cmd = MagicMock(spec=Command)
        self.mock_repo.get_by_id.return_value = expected_cmd

        result = self.services.get_command_by_id(id_)

        self.assertEqual(expected_cmd, result)
        self.mock_repo.get_by_id.assert_called_once_with(id_)

    def test_list_commands_no_tags(self):
        # Setup
        expected_commands = [MagicMock(spec=Command), MagicMock(spec=Command)]
        self.mock_repo.list_all.return_value = expected_commands

        # Execute
        result = self.services.list_commands(order_by="alias", limit=10)

        # Assert
        self.assertEqual(result, expected_commands)
        self.mock_repo.list_all.assert_called_once_with("alias", 10, profile=None)
        self.mock_repo.list_by_tag.assert_not_called()

    def test_list_commands_no_tags_from_different_profile(self):
        # Setup
        expected_commands = [MagicMock(spec=Command), MagicMock(spec=Command)]
        self.mock_repo.list_all.return_value = expected_commands
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        # Execute
        result = self.services.list_commands(
            order_by="alias", limit=10, profile=mock_profile
        )

        # Assert
        self.assertEqual(result, expected_commands)
        self.mock_repo.list_all.assert_called_once_with(
            "alias", 10, profile=mock_profile
        )
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
        self.mock_repo.list_by_tag.assert_called_once_with(
            tags, "alias", 10, profile=None
        )
        self.mock_repo.list_all.assert_not_called()

    def test_list_commands_with_tags_from_different_profile(self):
        # Setup
        tag_names = ["tag1"]
        tags = [MagicMock(spec=Tag)]
        expected_commands = [MagicMock(spec=Command)]
        self.mock_tag_repo.get_by_name.side_effect = tags
        self.mock_repo.list_by_tag.return_value = expected_commands
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        # Execute
        result = self.services.list_commands(
            tags=tag_names, order_by="alias", limit=10, profile=mock_profile
        )

        # Assert
        self.assertEqual(result, expected_commands)
        self.mock_tag_repo.get_by_name.assert_called_once_with("tag1")
        self.mock_repo.list_by_tag.assert_called_once_with(
            tags, "alias", 10, profile=mock_profile
        )
        self.mock_repo.list_all.assert_not_called()

    def test_search(self):
        # Setup
        query = "hello"
        fields = ["alias"]
        expected_commands = [MagicMock(spec=Command)]
        self.mock_repo.search.return_value = expected_commands

        # Execute
        result = self.services.search(query, fields, limit=20)

        # Assert
        self.assertEqual(expected_commands, result)
        self.mock_repo.search.assert_called_once_with(
            query, fields=fields, limit=20, profile=None
        )

    def test_search_in_different_profile(self):
        # Setup
        query = "hello"
        fields = ["alias"]
        expected_commands = [MagicMock(spec=Command)]
        self.mock_repo.search.return_value = expected_commands
        mock_profile = MagicMock(spec=Profile(id=10))
        self.mock_profile_repo.get_by_name.return_value = mock_profile

        # Execute
        result = self.services.search(query, fields, limit=20, profile=mock_profile)

        # Assert
        self.assertEqual(expected_commands, result)
        self.mock_repo.search.assert_called_once_with(
            query, fields=fields, limit=20, profile=mock_profile
        )

    def test_move_command(self):
        alias = "test-cmd"
        target_profile_name = "target-profile"
        source_profile = MagicMock(spec=Profile)
        target_profile = MagicMock(spec=Profile)
        cmd = MagicMock(spec=Command)
        updated_cmd = MagicMock(spec=Command)

        self.mock_profile_repo.get_by_name.side_effect = [
            source_profile,
            target_profile,
        ]
        self.mock_repo.get_by_alias.return_value = cmd
        self.mock_repo.update.return_value = updated_cmd

        result = self.services.move_command(
            alias, target_profile_name, profile="source-profile"
        )

        self.assertEqual(updated_cmd, result)
        self.mock_repo.get_by_alias.assert_called_once_with(
            alias, profile=source_profile
        )
        self.mock_profile_repo.get_by_name.assert_any_call("source-profile")
        self.mock_profile_repo.get_by_name.assert_any_call(target_profile_name)
        self.mock_repo.update.assert_called_once_with(cmd, profile=target_profile)

    @patch("cmdbox.services.command_services.db")
    def test_copy_command(self, mock_db):
        alias = "test-cmd"
        target_profile_name = "target-profile"
        source_profile = MagicMock(spec=Profile)
        target_profile = MagicMock(spec=Profile)
        source_tag = MagicMock(spec=Tag)
        source_cmd = MagicMock(spec=Command)
        source_cmd.alias = alias
        source_cmd.template = "echo hello"
        source_cmd.description = "desc"
        source_cmd.cwd = "/tmp"
        source_cmd.shell = "/bin/bash"
        source_cmd.env = '{"KEY": "VALUE"}'
        source_cmd.timeout = 30
        source_cmd.tags = [MagicMock(tag=source_tag)]

        copy_cmd = MagicMock(spec=Command)
        copy_cmd.id = 456
        expected_cmd = MagicMock(spec=Command)

        self.mock_profile_repo.get_by_name.side_effect = [
            source_profile,
            target_profile,
        ]
        self.mock_repo.get_by_alias.return_value = source_cmd
        self.mock_repo.create.return_value = copy_cmd
        self.mock_repo.get_by_id.return_value = expected_cmd

        result = self.services.copy_command(
            alias, target_profile_name, profile="source-profile"
        )

        self.assertEqual(expected_cmd, result)
        self.mock_repo.get_by_alias.assert_called_once_with(
            alias, profile=source_profile
        )
        self.mock_repo.create.assert_called_once_with(
            alias=alias,
            template=source_cmd.template,
            description=source_cmd.description,
            cwd=source_cmd.cwd,
            shell=source_cmd.shell,
            env={"KEY": "VALUE"},
            timeout=source_cmd.timeout,
            profile=target_profile,
        )
        self.mock_repo.add_tags.assert_called_once_with(copy_cmd, [source_tag])
        self.mock_repo.get_by_id.assert_called_once_with(copy_cmd.id)
        mock_db.atomic.assert_called_once()

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
        self.mock_repo.get_by_id.return_value = expected_cmd

        # Execute
        result = self.services.create_command(alias, template, tags=[])

        # Assert
        self.assertEqual(expected_cmd, result)
        self.mock_repo.create.assert_called_once_with(
            alias=alias,
            template=template,
            description=None,
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
            profile=None,
        )
        self.mock_repo.add_tags.assert_not_called()
        mock_db.atomic.assert_called_once()
