import unittest
from unittest.mock import patch
from cmdbox.cli.commands import command_crud
from cmdbox.cli.handlers import command_handlers


class TestCommandCrud(unittest.TestCase):

    @patch("cmdbox.cli.commands.command_crud.command_handlers.run_add_command")
    @patch("cmdbox.cli.commands.command_crud.container")
    def test_add(self, mock_container, mock_run_add_command):
        # Setup
        alias = "test-alias"
        template = "echo hello"
        description = "test description"
        tags = ["tag1", "tag2"]
        interactive = False

        # Execute
        command_crud.add(
            alias=alias,
            template=template,
            description=description,
            tags=tags,
            interactive=interactive,
        )

        # Verify
        mock_run_add_command.assert_called_once()
        args = mock_run_add_command.call_args[1]["args"]
        self.assertIsInstance(args, command_handlers.AddCommandArgs)
        self.assertEqual(args.alias, alias)
        self.assertEqual(args.template, template)
        self.assertEqual(args.description, description)
        self.assertEqual(args.tags, tags)
        self.assertEqual(args.interactive, interactive)

        self.assertEqual(
            mock_run_add_command.call_args[1]["get_cmd_services"],
            mock_container.get_command_services,
        )
        self.assertEqual(
            mock_run_add_command.call_args[1]["get_tag_services"],
            mock_container.get_tag_services,
        )
        self.assertEqual(
            mock_run_add_command.call_args[1]["get_console"], mock_container.get_console
        )

    @patch("cmdbox.cli.commands.command_crud.command_handlers.run_get_command")
    @patch("cmdbox.cli.commands.command_crud.container")
    def test_get(self, mock_container, mock_run_get_command):
        # Setup
        alias = "test-alias"

        # Execute
        command_crud.get(alias=alias)

        # Verify
        mock_run_get_command.assert_called_once_with(
            alias=alias,
            profile=None,
            get_cmd_services=mock_container.get_command_services,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.command_crud.command_handlers.run_update_command")
    @patch("cmdbox.cli.commands.command_crud.container")
    def test_update(self, mock_container, mock_run_update_command):
        # Setup
        alias = "test-alias"
        template = "new template"
        description = "new description"
        new_alias = "new-alias"
        set_pairs = ["key1=val1"]

        # Execute
        command_crud.update(
            alias=alias,
            template=template,
            description=description,
            new_alias=new_alias,
            set_=set_pairs,
            edit_mode=False,
            edit_fields="template,description",
        )

        # Verify
        mock_run_update_command.assert_called_once_with(
            alias=alias,
            template=template,
            description=description,
            new_alias=new_alias,
            cwd=None,
            clear_cwd=False,
            shell=None,
            clear_shell=False,
            env=None,
            clear_env=False,
            timeout=None,
            clear_timeout=False,
            set_pairs=set_pairs,
            edit_mode=False,
            edit_fields="template,description",
            get_cmd_services=mock_container.get_command_services,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.command_crud.command_handlers.run_list_command")
    @patch("cmdbox.cli.commands.command_crud.container")
    def test_list_cmds(self, mock_container, mock_run_list_command):
        # Setup
        order = "alias"
        tags = ["tag1"]
        limit = 5
        fields = ["alias", "template"]

        # Execute
        command_crud.list_cmds(order=order, tags=tags, limit=limit, fields=fields)

        # Verify
        mock_run_list_command.assert_called_once_with(
            limit=limit,
            page=None,
            order=order,
            tags=tags,
            fields=fields,
            profile=None,
            get_cmd_services=mock_container.get_command_services,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
            get_display_field_resolver=mock_container.get_command_display_field_resolver,
        )

    @patch("cmdbox.cli.commands.command_crud.command_handlers.run_search_command")
    @patch("cmdbox.cli.commands.command_crud.container")
    def test_search(self, mock_container, mock_run_search_command):
        # Setup
        term = "search-term"
        limit = 20
        search_fields = ["description"]
        fields = ["alias"]

        # Execute
        command_crud.search(
            term=term, limit=limit, search_fields=search_fields, fields=fields
        )

        # Verify
        mock_run_search_command.assert_called_once_with(
            term=term,
            limit=limit,
            page=None,
            search_fields=search_fields,
            fields=fields,
            profile=None,
            get_cmd_services=mock_container.get_command_services,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
            get_display_field_resolver=mock_container.get_command_display_field_resolver,
            get_search_field_resolver=mock_container.get_command_search_field_resolver,
        )

    @patch("cmdbox.cli.commands.command_crud.command_handlers.run_delete_command")
    @patch("cmdbox.cli.commands.command_crud.container")
    def test_delete(self, mock_container, mock_run_delete_command):
        # Setup
        alias = "test-alias"

        # Execute
        command_crud.delete(alias=alias)

        # Verify
        mock_run_delete_command.assert_called_once_with(
            alias=alias,
            profile=None,
            get_cmd_services=mock_container.get_command_services,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.command_crud.command_handlers.run_attach_tags")
    @patch("cmdbox.cli.commands.command_crud.container")
    def test_add_tags(self, mock_container, mock_run_attach_tags):
        # Setup
        alias = "test-alias"
        tags = ["tag1", "tag2"]

        # Execute
        command_crud.add_tags(alias=alias, tags=tags)

        # Verify
        mock_run_attach_tags.assert_called_once_with(
            alias=alias,
            tag_names=tags,
            profile=None,
            get_cmd_services=mock_container.get_command_services,
            get_tag_services=mock_container.get_tag_services,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.command_crud.command_handlers.run_detach_tags")
    @patch("cmdbox.cli.commands.command_crud.container")
    def test_remove_tags(self, mock_container, mock_run_detach_tags):
        # Setup
        alias = "test-alias"
        tags = ["tag1", "tag2"]

        # Execute
        command_crud.remove_tags(alias=alias, tags=tags)

        # Verify
        mock_run_detach_tags.assert_called_once_with(
            alias=alias,
            tag_names=tags,
            profile=None,
            get_cmd_services=mock_container.get_command_services,
            get_tag_services=mock_container.get_tag_services,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.command_crud.command_handlers.run_move_command")
    @patch("cmdbox.cli.commands.command_crud.container")
    def test_move(self, mock_container, mock_run_move_command):
        # Setup
        alias = "test-alias"
        target_profile = "target-profile"

        # Execute
        command_crud.move(alias=alias, target_profile=target_profile)

        # Verify
        mock_run_move_command.assert_called_once_with(
            alias=alias,
            target_profile=target_profile,
            profile=None,
            get_cmd_services=mock_container.get_command_services,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.command_crud.command_handlers.run_copy_command")
    @patch("cmdbox.cli.commands.command_crud.container")
    def test_copy(self, mock_container, mock_run_copy_command):
        # Setup
        alias = "test-alias"
        target_profile = "target-profile"
        new_alias = "new-alias"

        # Execute
        command_crud.copy(
            alias=alias, target_profile=target_profile, new_alias=new_alias
        )

        # Verify
        mock_run_copy_command.assert_called_once_with(
            alias=alias,
            target_profile=target_profile,
            new_alias=new_alias,
            profile=None,
            get_cmd_services=mock_container.get_command_services,
            get_console=mock_container.get_console,
        )
