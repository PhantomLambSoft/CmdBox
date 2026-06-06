import unittest
from unittest.mock import MagicMock, patch, ANY
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
from cmdbox.services.field_selection import FieldSelectionResolver


class TestCommandHandlers(unittest.TestCase):

    DISPLAY_FIELDS = [
        "f1",
        "f2",
        "test_field_one",
        "test_field_two",
        "test_field_three",
        "test_field_four",
    ]
    SEARCH_FIELDS = [
        "sf1",
        "sf2",
        "test_field_one",
        "test_field_two",
        "test_field_three",
        "test_field_four",
    ]

    def setUp(self):
        self.mock_cmd_services = MagicMock()
        self.mock_tag_services = MagicMock()
        self.mock_console = MagicMock()
        self.mock_settings = MagicMock()
        self.get_cmd_services = lambda: self.mock_cmd_services
        self.get_tag_services = lambda: self.mock_tag_services
        self.get_console = lambda: self.mock_console
        self.get_settings = lambda: self.mock_settings
        self.get_display_field_resolver = lambda: FieldSelectionResolver(
            self.DISPLAY_FIELDS
        )
        self.get_search_field_resolver = lambda: FieldSelectionResolver(
            self.SEARCH_FIELDS
        )

    @patch("cmdbox.cli.handlers.command_handlers.render_command_created")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_alias")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_template")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_description")
    @patch("cmdbox.cli.handlers.command_handlers.get_tags_interactive")
    def test_run_add_command_interactive(
        self,
        mock_get_tags,
        mock_prompt_desc,
        mock_prompt_tmpl,
        mock_prompt_alias,
        mock_render,
    ):
        args = AddCommandArgs(
            alias=None, template=None, description=None, tags=None, interactive=True
        )
        mock_prompt_alias.return_value = "alias1"
        mock_prompt_tmpl.return_value = "tmpl1"
        mock_prompt_desc.return_value = "desc1"
        mock_get_tags.return_value = ["tag1"]
        mock_cmd = MagicMock()
        self.mock_cmd_services.create_command.return_value = mock_cmd
        mock_render.return_value = "rendered_created"

        run_add_command(
            args=args,
            get_cmd_services=self.get_cmd_services,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_cmd_services.create_command.assert_called_with(
            alias="alias1",
            template="tmpl1",
            description="desc1",
            tags=["tag1"],
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
        )
        mock_render.assert_called_once_with(mock_cmd)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.command_handlers.render_command_created")
    def test_run_add_command_non_interactive(self, mock_render):
        args = AddCommandArgs(
            alias="alias1",
            template="tmpl1",
            description="desc1",
            tags=["tag1"],
            interactive=False,
        )
        mock_cmd = MagicMock()
        self.mock_cmd_services.create_command.return_value = mock_cmd
        mock_render.return_value = "rendered_created"

        run_add_command(
            args=args,
            get_cmd_services=self.get_cmd_services,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_cmd_services.create_command.assert_called_with(
            alias="alias1",
            template="tmpl1",
            description="desc1",
            tags=["tag1"],
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
        )
        mock_render.assert_called_once_with(mock_cmd)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.command_handlers.render_command")
    def test_run_get_command(self, mock_render):
        mock_cmd = MagicMock()
        self.mock_cmd_services.get_command.return_value = mock_cmd
        mock_render.return_value = "rendered_cmd"
        run_get_command(
            alias="alias1",
            get_cmd_services=self.get_cmd_services,
            get_console=self.get_console,
        )
        self.mock_cmd_services.get_command.assert_called_with("alias1")
        mock_render.assert_called_once_with(mock_cmd)
        self.mock_console.print.assert_called_with("rendered_cmd")

    @patch("cmdbox.cli.handlers.command_handlers.render_command_updated")
    def test_run_update_command_with_fields_supplied(self, mock_render):
        mock_cmd = MagicMock()
        mock_cmd.id = 1
        mock_cmd.env = None
        self.mock_cmd_services.get_command.return_value = mock_cmd
        mock_updated_cmd = MagicMock()
        self.mock_cmd_services.get_command_by_id.return_value = mock_updated_cmd
        mock_render.return_value = "rendered_update"

        run_update_command(
            alias="alias1",
            template="new_tmpl",
            description="new_desc",
            new_alias="new_alias",
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
            set_pairs=None,
            edit_mode=False,
            edit_fields=None,
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_cmd_services.update_command.assert_called_with(
            "alias1", template="new_tmpl", description="new_desc", alias="new_alias"
        )
        mock_render.assert_called_once_with(mock_updated_cmd)
        self.mock_console.print.assert_called_with("rendered_update")

    def test_run_update_command_no_fields_edit_mode_false(self):
        with self.assertRaises(typer.BadParameter):
            run_update_command(
                alias="alias1",
                template=None,
                description=None,
                new_alias=None,
                cwd=None,
                shell=None,
                env=None,
                timeout=None,
                set_pairs=None,
                edit_mode=False,
                edit_fields=None,
                get_cmd_services=self.get_cmd_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_timeout")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_shell")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_cwd")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_alias")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_template")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_description")
    @patch("cmdbox.cli.handlers.command_handlers.render_command_updated")
    def test_run_update_command_no_fields_edit_mode_true(
        self,
        mock_render,
        mock_prompt_desc,
        mock_prompt_tmpl,
        mock_prompt_alias,
        mock_prompt_cwd,
        mock_prompt_shell,
        mock_prompt_timeout,
    ):
        mock_render.return_value = "rendered_update"
        mock_prompt_alias.return_value = "new_prompt_alias"
        mock_prompt_tmpl.return_value = "new_prompt_tmpl"
        mock_prompt_desc.return_value = "new_prompt_desc"
        mock_prompt_cwd.return_value = "new_prompt_cwd"
        mock_prompt_shell.return_value = "new_prompt_shell"
        mock_prompt_timeout.return_value = "new_prompt_timeout"

        mock_cmd = MagicMock()
        mock_cmd.id = 1
        mock_cmd.template = "old_tmpl"
        mock_cmd.description = "old_desc"
        mock_cmd.alias = "old_alias"
        mock_cmd.env = None
        mock_cmd.cwd = None
        mock_cmd.shell = None
        mock_cmd.timeout = None

        self.mock_cmd_services.get_command.return_value = mock_cmd
        mock_updated_cmd = MagicMock()
        self.mock_cmd_services.get_command_by_id.return_value = mock_updated_cmd

        run_update_command(
            alias="alias1",
            template=None,
            description=None,
            new_alias=None,
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
            set_pairs=None,
            edit_mode=True,
            edit_fields=None,
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_cmd_services.update_command.assert_called_with(
            "alias1",
            template="new_prompt_tmpl",
            description="new_prompt_desc",
            alias="new_prompt_alias",
            cwd="new_prompt_cwd",
            shell="new_prompt_shell",
            timeout="new_prompt_timeout",
        )
        mock_render.assert_called_once_with(mock_updated_cmd)
        mock_prompt_desc.assert_called_once_with(default="old_desc")
        mock_prompt_tmpl.assert_called_once_with(ANY, default="old_tmpl")
        mock_prompt_alias.assert_called_once_with(ANY, default="old_alias")
        mock_prompt_cwd.assert_called_once_with(default="")
        mock_prompt_shell.assert_called_once_with(default="")
        mock_prompt_timeout.assert_called_once_with(default="")
        self.mock_console.print.assert_called_with("rendered_update")

    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_timeout")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_shell")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_cwd")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_alias")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_template")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_description")
    @patch("cmdbox.cli.handlers.command_handlers.render_command_updated")
    def test_run_update_command_no_fields_edit_mode_true_edit_field_provided(
        self,
        mock_render,
        mock_prompt_desc,
        mock_prompt_tmpl,
        mock_prompt_alias,
        mock_prompt_cwd,
        mock_prompt_shell,
        mock_prompt_timeout,
    ):
        mock_render.return_value = "rendered_update"
        mock_prompt_alias.return_value = "new_prompt_alias"
        mock_prompt_tmpl.return_value = "new_prompt_tmpl"
        mock_prompt_desc.return_value = "new_prompt_desc"
        mock_prompt_cwd.return_value = "new_prompt_cwd"
        mock_prompt_shell.return_value = "new_prompt_shell"
        mock_prompt_timeout.return_value = "new_prompt_timeout"

        mock_cmd = MagicMock()
        mock_cmd.id = 1
        mock_cmd.template = "old_tmpl"
        mock_cmd.description = "old_desc"
        mock_cmd.alias = "old_alias"
        mock_cmd.cwd = None
        mock_cmd.shell = None
        mock_cmd.timeout = None
        mock_cmd.env = None

        self.mock_cmd_services.get_command.return_value = mock_cmd
        mock_updated_cmd = MagicMock()
        self.mock_cmd_services.get_command_by_id.return_value = mock_updated_cmd

        run_update_command(
            alias="alias1",
            template=None,
            description=None,
            new_alias=None,
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
            set_pairs=None,
            edit_mode=True,
            edit_fields="template",
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_cmd_services.update_command.assert_called_with(
            "alias1",
            template="new_prompt_tmpl",
        )
        mock_render.assert_called_once_with(mock_updated_cmd)
        mock_prompt_desc.assert_not_called()
        mock_prompt_tmpl.assert_called_once_with(ANY, default="old_tmpl")
        mock_prompt_alias.assert_not_called()
        self.mock_console.print.assert_called_with("rendered_update")

    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_timeout")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_shell")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_cwd")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_alias")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_template")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_description")
    @patch("cmdbox.cli.handlers.command_handlers.render_command_updated")
    def test_run_update_command_no_fields_edit_mode_true_multiple_edit_fields_provided(
        self,
        mock_render,
        mock_prompt_desc,
        mock_prompt_tmpl,
        mock_prompt_alias,
        mock_prompt_cwd,
        mock_prompt_shell,
        mock_prompt_timeout,
    ):
        mock_render.return_value = "rendered_update"
        mock_prompt_alias.return_value = "new_prompt_alias"
        mock_prompt_tmpl.return_value = "new_prompt_tmpl"
        mock_prompt_desc.return_value = "new_prompt_desc"
        mock_prompt_cwd.return_value = "new_prompt_cwd"
        mock_prompt_shell.return_value = "new_prompt_shell"
        mock_prompt_timeout.return_value = "new_prompt_timeout"

        mock_cmd = MagicMock()
        mock_cmd.id = 1
        mock_cmd.template = "old_tmpl"
        mock_cmd.description = "old_desc"
        mock_cmd.alias = "old_alias"
        mock_cmd.env = None
        mock_cmd.cwd = None
        mock_cmd.shell = None
        mock_cmd.timeout = None

        self.mock_cmd_services.get_command.return_value = mock_cmd
        mock_updated_cmd = MagicMock()
        self.mock_cmd_services.get_command_by_id.return_value = mock_updated_cmd

        run_update_command(
            alias="alias1",
            template=None,
            description=None,
            new_alias=None,
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
            set_pairs=None,
            edit_mode=True,
            edit_fields="template, description",
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_cmd_services.update_command.assert_called_with(
            "alias1",
            template="new_prompt_tmpl",
            description="new_prompt_desc",
        )
        mock_render.assert_called_once_with(mock_updated_cmd)
        mock_prompt_desc.assert_called_once_with(default="old_desc")
        mock_prompt_tmpl.assert_called_once_with(ANY, default="old_tmpl")
        mock_prompt_alias.assert_not_called()
        self.mock_console.print.assert_called_with("rendered_update")

    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_timeout")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_shell")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_cwd")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_alias")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_template")
    @patch("cmdbox.cli.handlers.command_handlers.prompt_for_description")
    @patch("cmdbox.cli.handlers.command_handlers.render_command_updated")
    def test_run_update_command_no_fields_edit_mode_true_edit_field_provided_as_aliases_from_settings(
        self,
        mock_render,
        mock_prompt_desc,
        mock_prompt_tmpl,
        mock_prompt_alias,
        mock_prompt_cwd,
        mock_prompt_shell,
        mock_prompt_timeout,
    ):
        mock_render.return_value = "rendered_update"
        mock_prompt_alias.return_value = "new_prompt_alias"
        mock_prompt_tmpl.return_value = "new_prompt_tmpl"
        mock_prompt_desc.return_value = "new_prompt_desc"
        mock_prompt_cwd.return_value = "new_prompt_cwd"
        mock_prompt_shell.return_value = "new_prompt_shell"
        mock_prompt_timeout.return_value = "new_prompt_timeout"

        mock_cmd = MagicMock()
        mock_cmd.id = 1
        mock_cmd.template = "old_tmpl"
        mock_cmd.description = "old_desc"
        mock_cmd.alias = "old_alias"
        mock_cmd.env = None
        mock_cmd.cwd = None
        mock_cmd.shell = None
        mock_cmd.timeout = None

        self.mock_cmd_services.get_command.return_value = mock_cmd
        mock_updated_cmd = MagicMock()
        self.mock_cmd_services.get_command_by_id.return_value = mock_updated_cmd
        self.mock_settings.field_aliases.alias_mapping = {"template": ["tpl"]}

        run_update_command(
            alias="alias1",
            template=None,
            description=None,
            new_alias=None,
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
            set_pairs=None,
            edit_mode=True,
            edit_fields="tpl",
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_cmd_services.update_command.assert_called_with(
            "alias1",
            template="new_prompt_tmpl",
        )
        mock_render.assert_called_once_with(mock_updated_cmd)
        mock_prompt_desc.assert_not_called()
        mock_prompt_tmpl.assert_called_once_with(ANY, default="old_tmpl")
        mock_prompt_alias.assert_not_called()
        self.mock_console.print.assert_called_with("rendered_update")

    def test_run_update_command_in_edit_mode_with_fields_raised_error(self):
        with self.assertRaises(typer.BadParameter):
            run_update_command(
                alias="alias1",
                template="AmarilloByMorning",
                description=None,
                new_alias=None,
                cwd=None,
                shell=None,
                env=None,
                timeout=None,
                set_pairs=None,
                edit_mode=True,
                edit_fields=None,
                get_cmd_services=self.get_cmd_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.command_handlers.render_command_updated")
    def test_run_update_command_does_not_update_if_updated_fields_are_unchanged(
        self, mock_render
    ):
        mock_cmd = MagicMock()
        mock_cmd.id = 1
        mock_cmd.alias = "alias1"
        mock_cmd.template = "old_tmpl"
        mock_cmd.description = "old_desc"
        mock_cmd.env = None
        self.mock_cmd_services.get_command.return_value = mock_cmd
        mock_updated_cmd = MagicMock()
        self.mock_cmd_services.get_command_by_id.return_value = mock_updated_cmd
        mock_render.return_value = "rendered_update"

        run_update_command(
            alias="alias1",
            template="old_tmpl",
            description="old_desc",
            new_alias="alias1",
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
            set_pairs=None,
            edit_mode=False,
            edit_fields=None,
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_cmd_services.update_command.assert_not_called()
        mock_render.assert_not_called()
        self.mock_console.info.assert_called_with("No changes detected.")

    def test_run_update_raises_error_when_field_variable_is_also_in_set(self):
        with self.assertRaises(typer.BadParameter):
            run_update_command(
                alias="alias1",
                template="WriteThisDown",
                description=None,
                new_alias=None,
                cwd=None,
                shell=None,
                env=None,
                timeout=None,
                set_pairs=["template=MarinaDelRey"],
                edit_mode=False,
                edit_fields=None,
                get_cmd_services=self.get_cmd_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    def test_run_update_raises_error_when_set_is_provided_with_edit_flag(self):
        with self.assertRaises(typer.BadParameter):
            run_update_command(
                alias="alias1",
                template=None,
                description=None,
                new_alias=None,
                cwd=None,
                shell=None,
                env=None,
                timeout=None,
                set_pairs=["description=OceanFrontProperty"],
                edit_mode=True,
                edit_fields=None,
                get_cmd_services=self.get_cmd_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.command_handlers.render_command_list")
    def test_run_list_command(self, mock_render):
        cmds = ["c1", "c2"]
        self.mock_cmd_services.list_commands.return_value = cmds
        mock_render.return_value = "rendered_list"
        run_list_command(
            limit=10,
            order="name",
            tags=["t1"],
            fields=["f1"],
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
        )
        self.mock_cmd_services.list_commands.assert_called_with(
            limit=10, order_by="name", tags=["t1"]
        )
        mock_render.assert_called_once_with(cmds, title="Commands", fields=["f1"])
        self.mock_console.print.assert_called_with("rendered_list")

    @patch("cmdbox.cli.handlers.command_handlers.render_command_list")
    def test_run_list_commands_uses_correct_defaults_from_settings(self, mock_render):
        cmds = ["c1", "c2", "c3"]
        self.mock_cmd_services.list_commands.return_value = cmds
        mock_render.return_value = "rendered_list"
        self.mock_settings.default_fields.command_output = [
            "test_field_one",
            "test_field_two",
        ]
        run_list_command(
            limit=10,
            order="name",
            tags=["t1"],
            fields=None,
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
        )
        self.mock_cmd_services.list_commands.assert_called_with(
            limit=10, order_by="name", tags=["t1"]
        )
        mock_render.assert_called_once_with(
            cmds, title="Commands", fields=["test_field_one", "test_field_two"]
        )
        self.mock_console.print.assert_called_with("rendered_list")

    @patch("cmdbox.cli.handlers.command_handlers.render_command_list")
    def test_run_list_command_uses_all_keyword_correctly(self, mock_render):
        cmds = ["c1", "c2", "c3"]
        self.mock_cmd_services.list_commands.return_value = cmds
        mock_render.return_value = "rendered_list"
        run_list_command(
            limit=10,
            order="name",
            tags=["t1"],
            fields=["all"],
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
        )
        self.mock_cmd_services.list_commands.assert_called_with(
            limit=10, order_by="name", tags=["t1"]
        )
        mock_render.assert_called_once_with(
            cmds, title="Commands", fields=self.DISPLAY_FIELDS
        )
        self.mock_console.print.assert_called_with("rendered_list")

    @patch("cmdbox.cli.handlers.command_handlers.render_command_list")
    def test_run_search_command(self, mock_render):
        cmds = ["c1"]
        self.mock_cmd_services.search.return_value = cmds
        mock_render.return_value = "rendered_search"
        run_search_command(
            term="term",
            limit=5,
            search_fields=["sf1"],
            fields=["f1"],
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
            get_search_field_resolver=self.get_search_field_resolver,
        )
        self.mock_cmd_services.search.assert_called_with(
            "term", limit=5, fields=["sf1"]
        )
        mock_render.assert_called_once_with(cmds, title="Search Results", fields=["f1"])
        self.mock_console.print.assert_called_with("rendered_search")

    @patch("cmdbox.cli.handlers.command_handlers.render_command_list")
    def test_run_search_commands_uses_correct_defaults_from_settings(self, mock_render):
        cmds = ["c1", "c2", "c3"]
        self.mock_cmd_services.search.return_value = cmds
        mock_render.return_value = "rendered_search"
        self.mock_settings.default_fields.command_output = [
            "test_field_one",
            "test_field_two",
        ]
        self.mock_settings.default_fields.command_search = [
            "test_field_three",
            "test_field_four",
        ]
        run_search_command(
            term="term",
            limit=5,
            search_fields=None,
            fields=None,
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
            get_search_field_resolver=self.get_search_field_resolver,
        )
        self.mock_cmd_services.search.assert_called_with(
            "term",
            limit=5,
            fields=["test_field_three", "test_field_four"],  # search fields
        )
        mock_render.assert_called_once_with(
            cmds,
            title="Search Results",
            fields=["test_field_one", "test_field_two"],  # output fields
        )

    @patch("cmdbox.cli.handlers.command_handlers.render_command_list")
    def test_run_search_commands_uses_all_keyword_correctly(self, mock_render):
        cmds = ["c1"]
        self.mock_cmd_services.search.return_value = cmds
        mock_render.return_value = "rendered_search"
        run_search_command(
            term="term",
            limit=5,
            search_fields=["all"],
            fields=["all"],
            get_cmd_services=self.get_cmd_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
            get_search_field_resolver=self.get_search_field_resolver,
        )
        self.mock_cmd_services.search.assert_called_with(
            "term", limit=5, fields=self.SEARCH_FIELDS
        )
        mock_render.assert_called_once_with(
            cmds, title="Search Results", fields=self.DISPLAY_FIELDS
        )
        self.mock_console.print.assert_called_with("rendered_search")

    @patch("cmdbox.cli.handlers.command_handlers.render_command_deleted")
    def test_run_delete_command_success(self, mock_render):
        mock_cmd = MagicMock()
        self.mock_cmd_services.get_command.return_value = mock_cmd
        self.mock_cmd_services.delete_command.return_value = True
        mock_render.return_value = "rendered_deleted"

        run_delete_command(
            alias="alias1",
            get_cmd_services=self.get_cmd_services,
            get_console=self.get_console,
        )

        self.mock_cmd_services.delete_command.assert_called_with("alias1")
        mock_render.assert_called_once_with(mock_cmd)
        self.mock_console.print.assert_called_with("rendered_deleted")

    def test_run_delete_command_failure(self):
        self.mock_cmd_services.delete_command.return_value = False
        run_delete_command(
            alias="alias1",
            get_cmd_services=self.get_cmd_services,
            get_console=self.get_console,
        )
        self.mock_console.error.assert_called()
