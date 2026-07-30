import unittest
from unittest.mock import MagicMock, patch, ANY
import typer
from cmdbox.cli.handlers.profile_handler import (
    AddProfileArgs,
    run_add_profile,
    run_get_profile,
    run_update_profile,
    run_list_profiles,
    run_delete_profile,
    run_switch_command_profile,
    run_switch_variable_profile,
    run_switch_settings_profile,
    run_switch_profile,
    run_profile_status,
)


class TestProfileHandler(unittest.TestCase):

    def setUp(self):
        self.mock_profile_services = MagicMock()
        self.mock_console = MagicMock()
        self.mock_settings = MagicMock()
        self.get_profile_services = lambda: self.mock_profile_services
        self.get_console = lambda: self.mock_console
        self.get_settings = lambda: self.mock_settings

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_created")
    @patch("cmdbox.cli.handlers.profile_handler.prompt_for_name")
    @patch("cmdbox.cli.handlers.profile_handler.prompt_for_description")
    def test_run_add_profile_interactive(
        self, mock_prompt_desc, mock_prompt_name, mock_render
    ):
        args = AddProfileArgs(name=None, description=None, interactive=True)
        mock_prompt_name.return_value = "profile1"
        mock_prompt_desc.return_value = "desc1"
        mock_profile = MagicMock()
        self.mock_profile_services.create_profile.return_value = mock_profile
        mock_render.return_value = "rendered_created"

        run_add_profile(
            args=args,
            get_profile_services=self.get_profile_services,
            get_console=self.get_console,
        )

        self.mock_profile_services.create_profile.assert_called_with(
            name="profile1", description="desc1"
        )
        mock_render.assert_called_once_with(mock_profile)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_created")
    def test_run_add_profile_non_interactive(self, mock_render):
        args = AddProfileArgs(name="profile1", description="desc1", interactive=False)
        mock_profile = MagicMock()
        self.mock_profile_services.create_profile.return_value = mock_profile
        mock_render.return_value = "rendered_created"

        run_add_profile(
            args=args,
            get_profile_services=self.get_profile_services,
            get_console=self.get_console,
        )

        self.mock_profile_services.create_profile.assert_called_with(
            name="profile1", description="desc1"
        )
        mock_render.assert_called_once_with(mock_profile)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.profile_handler.render_profile")
    def test_run_get_profile(self, mock_render):
        mock_profile = MagicMock()
        self.mock_profile_services.get_profile.return_value = mock_profile
        mock_render.return_value = "rendered_profile"
        run_get_profile(
            name="profile1",
            get_profile_services=self.get_profile_services,
            get_console=self.get_console,
        )
        self.mock_profile_services.get_profile.assert_called_with(name="profile1")
        mock_render.assert_called_once_with(mock_profile)
        self.mock_console.print.assert_called_with("rendered_profile")

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_updated")
    def test_run_update_profile_with_fields_supplied(self, mock_render):
        mock_profile = MagicMock()
        mock_profile.name = "profile1"
        mock_profile.description = "old_desc"
        self.mock_profile_services.get_profile.side_effect = [mock_profile, MagicMock()]
        mock_render.return_value = "rendered_updated"

        run_update_profile(
            name="profile1",
            description="new_desc",
            new_name="new_name",
            set_pairs=[],
            edit_mode=False,
            edit_fields=None,
            get_profile_services=self.get_profile_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_profile_services.update_profile.assert_called_with(
            "profile1", description="new_desc", name="new_name"
        )
        self.mock_console.print.assert_called_with("rendered_updated")

    def test_run_update_profile_no_fields_edit_mode_false(self):
        mock_profile = MagicMock()
        self.mock_profile_services.get_profile.return_value = mock_profile
        with self.assertRaises(typer.BadParameter):
            run_update_profile(
                name="profile1",
                description=None,
                new_name=None,
                set_pairs=[],
                edit_mode=False,
                edit_fields=None,
                get_profile_services=self.get_profile_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.profile_handler.prompt_for_description")
    @patch("cmdbox.cli.handlers.profile_handler.prompt_for_name")
    @patch("cmdbox.cli.handlers.profile_handler.render_profile_updated")
    def test_run_update_profile_with_fields_edit_mode_true(
        self,
        mock_render,
        mock_prompt_name,
        mock_prompt_desc,
    ):
        mock_render.return_value = "rendered_updated"
        mock_prompt_name.return_value = "new_prompt_name"
        mock_prompt_desc.return_value = "new_prompt_desc"

        mock_profile = MagicMock()
        mock_profile.name = "profile1"
        mock_profile.description = "old_desc"
        self.mock_profile_services.get_profile.side_effect = [mock_profile, MagicMock()]

        run_update_profile(
            name="profile1",
            description=None,
            new_name=None,
            set_pairs=[],
            edit_mode=True,
            edit_fields=None,
            get_profile_services=self.get_profile_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_profile_services.update_profile.assert_called_with(
            "profile1", description="new_prompt_desc", name="new_prompt_name"
        )
        mock_prompt_name.assert_called_once_with(ANY, default="profile1")
        mock_prompt_desc.assert_called_once_with(default="old_desc")
        self.mock_console.print.assert_called_with("rendered_updated")

    @patch("cmdbox.cli.handlers.profile_handler.prompt_for_description")
    @patch("cmdbox.cli.handlers.profile_handler.prompt_for_name")
    @patch("cmdbox.cli.handlers.profile_handler.render_profile_updated")
    def test_run_update_profile_with_fields_edit_mode_true_edit_fields_provided(
        self,
        mock_render,
        mock_prompt_name,
        mock_prompt_desc,
    ):
        mock_render.return_value = "rendered_updated"
        mock_prompt_desc.return_value = "new_prompt_desc"

        mock_profile = MagicMock()
        mock_profile.name = "profile1"
        mock_profile.description = "old_desc"
        self.mock_profile_services.get_profile.side_effect = [mock_profile, MagicMock()]

        run_update_profile(
            name="profile1",
            description=None,
            new_name=None,
            set_pairs=[],
            edit_mode=True,
            edit_fields="description",
            get_profile_services=self.get_profile_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_profile_services.update_profile.assert_called_with(
            "profile1", description="new_prompt_desc"
        )
        mock_prompt_name.assert_not_called()
        mock_prompt_desc.assert_called_once_with(default="old_desc")
        self.mock_console.print.assert_called_with("rendered_updated")

    def test_run_update_profile_in_edit_mode_with_fields_raises_error(self):
        mock_profile = MagicMock()
        self.mock_profile_services.get_profile.return_value = mock_profile
        with self.assertRaises(typer.BadParameter):
            run_update_profile(
                name="profile1",
                description="some desc",
                new_name=None,
                set_pairs=[],
                edit_mode=True,
                edit_fields=None,
                get_profile_services=self.get_profile_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_updated")
    def test_run_update_profile_does_not_update_if_fields_are_unchanged(
        self, mock_render
    ):
        mock_profile = MagicMock()
        mock_profile.name = "profile1"
        mock_profile.description = "old_desc"
        self.mock_profile_services.get_profile.return_value = mock_profile

        run_update_profile(
            name="profile1",
            description="old_desc",
            new_name="profile1",
            set_pairs=[],
            edit_mode=False,
            edit_fields=None,
            get_profile_services=self.get_profile_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_profile_services.update_profile.assert_not_called()
        mock_render.assert_not_called()
        self.mock_console.info.assert_called_with("No changes detected.")

    def test_run_update_profile_raises_error_when_set_is_provided_with_edit_flag(self):
        mock_profile = MagicMock()
        self.mock_profile_services.get_profile.return_value = mock_profile
        with self.assertRaises(typer.BadParameter):
            run_update_profile(
                name="profile1",
                description=None,
                new_name=None,
                set_pairs=["description=new"],
                edit_mode=True,
                edit_fields=None,
                get_profile_services=self.get_profile_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_list")
    def test_run_list_profiles(self, mock_render):
        profiles = ["p1", "p2"]
        self.mock_profile_services.list_profiles.return_value = profiles
        mock_render.return_value = "rendered_list"
        self.mock_settings.default_fields.tag_list_limit = 10
        run_list_profiles(
            limit=None,
            page=None,
            order_by=None,
            fields=["f1"],
            get_profile_services=self.get_profile_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )
        self.mock_profile_services.list_profiles.assert_called_with(
            limit=10, order_by="name"
        )
        mock_render.assert_called_once_with(profiles, title="Profiles", fields=["f1"])
        self.mock_console.print_paged.assert_called_with(
            "rendered_list", row_count=2, force=None
        )

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_deleted")
    def test_run_delete_profile_success(self, mock_render):
        mock_profile = MagicMock()
        self.mock_profile_services.get_profile.return_value = mock_profile
        self.mock_profile_services.delete_profile.return_value = True
        mock_render.return_value = "rendered_deleted"

        run_delete_profile(
            name="profile1",
            force=False,
            get_profile_services=self.get_profile_services,
            get_console=self.get_console,
        )

        self.mock_profile_services.delete_profile.assert_called_with(
            "profile1", force=False
        )
        mock_render.assert_called_once_with(mock_profile)
        self.mock_console.print.assert_called_with("rendered_deleted")

    def test_run_delete_profile_failure(self):
        mock_profile = MagicMock()
        self.mock_profile_services.get_profile.return_value = mock_profile
        self.mock_profile_services.delete_profile.return_value = False
        run_delete_profile(
            name="profile1",
            force=False,
            get_profile_services=self.get_profile_services,
            get_console=self.get_console,
        )
        self.mock_console.error.assert_called_with(
            "Failed to delete profile 'profile1'."
        )

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_switched")
    def test_run_switch_command_profile(self, mock_render):
        mock_profile = MagicMock()
        self.mock_profile_services.get_profile.return_value = mock_profile
        mock_render.return_value = "rendered_switched"
        run_switch_command_profile(
            name="profile1",
            get_profile_services=self.get_profile_services,
            get_console=self.get_console,
        )
        self.mock_profile_services.switch_command_profile.assert_called_with("profile1")
        mock_render.assert_called_once_with(mock_profile, scope="command")
        self.mock_console.print.assert_called_with("rendered_switched")

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_switched")
    def test_run_switch_variable_profile(self, mock_render):
        mock_profile = MagicMock()
        self.mock_profile_services.get_profile.return_value = mock_profile
        mock_render.return_value = "rendered_switched"
        run_switch_variable_profile(
            name="profile1",
            get_profile_services=self.get_profile_services,
            get_console=self.get_console,
        )
        self.mock_profile_services.switch_variable_profile.assert_called_with(
            "profile1"
        )
        mock_render.assert_called_once_with(mock_profile, scope="variable")
        self.mock_console.print.assert_called_with("rendered_switched")

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_switched")
    def test_run_switch_settings_profile(self, mock_render):
        mock_profile = MagicMock()
        self.mock_profile_services.get_profile.return_value = mock_profile
        mock_render.return_value = "rendered_switched"
        run_switch_settings_profile(
            name="profile1",
            get_profile_services=self.get_profile_services,
            get_console=self.get_console,
        )
        self.mock_profile_services.switch_settings_profile.assert_called_with(
            "profile1"
        )
        mock_render.assert_called_once_with(mock_profile, scope="settings")
        self.mock_console.print.assert_called_with("rendered_switched")

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_switched")
    def test_run_switch_profile_all(self, mock_render):
        mock_profile = MagicMock()
        self.mock_profile_services.get_profile.return_value = mock_profile
        mock_render.return_value = "rendered_switched"
        run_switch_profile(
            name="profile1",
            cmd=False,
            var=False,
            settings=False,
            get_profile_services=self.get_profile_services,
            get_console=self.get_console,
        )
        self.mock_profile_services.switch_profile.assert_called_with("profile1")
        mock_render.assert_called_once_with(
            mock_profile, scope="command, variable, and settings"
        )
        self.mock_console.print.assert_called_with("rendered_switched")

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_switched")
    def test_run_switch_profile_partial(self, mock_render):
        mock_profile = MagicMock()
        self.mock_profile_services.get_profile.return_value = mock_profile
        mock_render.return_value = "rendered_switched"
        run_switch_profile(
            name="profile1",
            cmd=True,
            var=False,
            settings=True,
            get_profile_services=self.get_profile_services,
            get_console=self.get_console,
        )
        self.mock_profile_services.switch_command_profile.assert_called_with("profile1")
        self.mock_profile_services.switch_settings_profile.assert_called_with(
            "profile1"
        )
        self.mock_profile_services.switch_variable_profile.assert_not_called()
        mock_render.assert_called_once_with(mock_profile, scope="command, settings")
        self.mock_console.print.assert_called_with("rendered_switched")

    @patch("cmdbox.cli.handlers.profile_handler.render_profile_status")
    def test_run_profile_status(self, mock_render):
        mock_status = MagicMock()
        self.mock_profile_services.get_status.return_value = mock_status
        mock_render.return_value = "rendered_status"
        run_profile_status(
            get_profile_services=self.get_profile_services,
            get_console=self.get_console,
        )
        self.mock_profile_services.get_status.assert_called_once()
        mock_render.assert_called_once_with(mock_status)
        self.mock_console.print.assert_called_with("rendered_status")
