import unittest
from unittest.mock import MagicMock, patch
import typer
from cmdbox.cli.handlers.command_handlers import (
    AddCommandArgs,
    run_add_command,
    run_get_command,
    run_update_command,
    run_list_command,
    run_search_command,
    run_delete_command,
)


class TestCommandHandlers(unittest.TestCase):

    def setUp(self):
        self.mock_cmd_services = MagicMock()
        self.mock_tag_services = MagicMock()
        self.mock_console = MagicMock()
        self.get_cmd_services = lambda: self.mock_cmd_services
        self.get_tag_services = lambda: self.mock_tag_services
        self.get_console = lambda: self.mock_console

    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_alias")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_template")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_description")
    @patch("cmdbox.cli.handlers.command_handlers.get_tags_interactive")
    def test_run_add_command_interactive(
        self, mock_get_tags, mock_prompt_desc, mock_prompt_tmpl, mock_prompt_alias
    ):
        args = AddCommandArgs(
            alias=None, template=None, description=None, tags=None, interactive=True
        )
        mock_prompt_alias.return_value = "alias1"
        mock_prompt_tmpl.return_value = "tmpl1"
        mock_prompt_desc.return_value = "desc1"
        mock_get_tags.return_value = ["tag1"]
        self.mock_cmd_services.create_command.return_value = MagicMock()

        run_add_command(
            args=args,
            get_cmd_services=self.get_cmd_services,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_cmd_services.create_command.assert_called_with(
            alias="alias1", template="tmpl1", description="desc1", tags=["tag1"]
        )
        self.mock_console.success.assert_called()

    def test_run_add_command_non_interactive(self):
        args = AddCommandArgs(
            alias="alias1",
            template="tmpl1",
            description="desc1",
            tags=["tag1"],
            interactive=False,
        )
        self.mock_cmd_services.create_command.return_value = MagicMock()

        run_add_command(
            args=args,
            get_cmd_services=self.get_cmd_services,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_cmd_services.create_command.assert_called_with(
            alias="alias1", template="tmpl1", description="desc1", tags=["tag1"]
        )

    def test_run_get_command(self):
        self.mock_cmd_services.get_command.return_value = "fake_cmd"
        run_get_command(
            alias="alias1",
            get_cmd_services=self.get_cmd_services,
            get_console=self.get_console,
        )
        self.mock_cmd_services.get_command.assert_called_with("alias1")
        self.mock_console.print_command.assert_called_with("fake_cmd")

    def test_run_update_command(self):
        mock_cmd = MagicMock()
        mock_cmd.id = 1
        self.mock_cmd_services.get_command.return_value = mock_cmd
        self.mock_cmd_services.get_command_by_id.return_value = "updated_cmd"

        run_update_command(
            alias="alias1",
            template="new_tmpl",
            description="new_desc",
            new_alias="new_alias",
            set_pairs=None,
            get_cmd_services=self.get_cmd_services,
            get_console=self.get_console,
        )

        self.mock_cmd_services.update_command.assert_called_with(
            "alias1", template="new_tmpl", description="new_desc", alias="new_alias"
        )
        self.mock_console.success.assert_called()
        self.mock_console.print_command.assert_called_with("updated_cmd")

    def test_run_update_command_no_fields(self):
        with self.assertRaises(typer.BadParameter):
            run_update_command(
                alias="alias1",
                template=None,
                description=None,
                new_alias=None,
                set_pairs=None,
                get_cmd_services=self.get_cmd_services,
                get_console=self.get_console,
            )

    def test_run_list_command(self):
        self.mock_cmd_services.list_commands.return_value = ["c1", "c2"]
        run_list_command(
            limit=10,
            order="name",
            tags=["t1"],
            fields=["f1"],
            get_cmd_services=self.get_cmd_services,
            get_console=self.get_console,
        )
        self.mock_cmd_services.list_commands.assert_called_with(
            limit=10, order_by="name", tags=["t1"]
        )
        self.mock_console.print_command_list.assert_called_with(
            ["c1", "c2"], output_fields=["f1"]
        )

    def test_run_search_command(self):
        self.mock_cmd_services.search_commands.return_value = ["c1"]
        run_search_command(
            term="term",
            limit=5,
            search_fields=["sf1"],
            fields=["f1"],
            get_cmd_services=self.get_cmd_services,
            get_console=self.get_console,
        )
        self.mock_cmd_services.search_commands.assert_called_with(
            "term", limit=5, fields=["sf1"]
        )
        self.mock_console.print_command_list.assert_called_with(
            ["c1"], output_fields=["f1"]
        )

    def test_run_delete_command_success(self):
        mock_cmd = MagicMock()
        self.mock_cmd_services.get_command.return_value = mock_cmd
        self.mock_cmd_services.delete_command.return_value = True

        run_delete_command(
            alias="alias1",
            get_cmd_services=self.get_cmd_services,
            get_console=self.get_console,
        )

        self.mock_cmd_services.delete_command.assert_called_with("alias1")
        self.mock_console.success.assert_called()
        self.mock_console.print_command.assert_called_with(mock_cmd)

    def test_run_delete_command_failure(self):
        self.mock_cmd_services.delete_command.return_value = False
        run_delete_command(
            alias="alias1",
            get_cmd_services=self.get_cmd_services,
            get_console=self.get_console,
        )
        self.mock_console.error.assert_called()
