import unittest
from unittest.mock import patch
from cmdbox.cli.commands import profile_crud
from cmdbox.cli.handlers import profile_handler


class TestProfileCrud(unittest.TestCase):

    @patch("cmdbox.cli.commands.profile_crud.profile_handler.run_add_profile")
    @patch("cmdbox.cli.commands.profile_crud.container")
    def test_add(self, mock_container, mock_run_add_profile):
        name = "test-profile"
        description = "test description"
        interactive = False

        profile_crud.add(name=name, description=description, interactive=interactive)

        mock_run_add_profile.assert_called_once()
        args = mock_run_add_profile.call_args[1]["args"]
        self.assertIsInstance(args, profile_handler.AddProfileArgs)
        self.assertEqual(name, args.name)
        self.assertEqual(description, args.description)
        self.assertEqual(interactive, args.interactive)
        self.assertEqual(
            mock_run_add_profile.call_args[1]["get_profile_services"],
            mock_container.get_profile_service,
        )
        self.assertEqual(
            mock_run_add_profile.call_args[1]["get_console"], mock_container.get_console
        )

    @patch("cmdbox.cli.commands.profile_crud.profile_handler.run_get_profile")
    @patch("cmdbox.cli.commands.profile_crud.container")
    def test_get(self, mock_container, mock_run_get_profile):
        name = "test-profile"

        profile_crud.get(name=name)

        mock_run_get_profile.assert_called_once_with(
            name=name,
            get_profile_services=mock_container.get_profile_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.profile_crud.profile_handler.run_update_profile")
    @patch("cmdbox.cli.commands.profile_crud.container")
    def test_update(self, mock_container, mock_run_update_profile):
        name = "test-profile"
        description = "new description"
        new_name = "new-name"
        set_pairs = ["key1=val1"]
        edit_mode = False
        edit_fields = "name,description"

        profile_crud.update(
            name=name,
            description=description,
            new_name=new_name,
            set_=set_pairs,
            edit_mode=edit_mode,
            edit_fields=edit_fields,
        )

        mock_run_update_profile.assert_called_once_with(
            name=name,
            description=description,
            new_name=new_name,
            set_pairs=set_pairs,
            edit_mode=edit_mode,
            edit_fields=edit_fields,
            get_profile_services=mock_container.get_profile_service,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.profile_crud.profile_handler.run_list_profiles")
    @patch("cmdbox.cli.commands.profile_crud.container")
    def test_list_profiles(self, mock_container, mock_run_list_profiles):
        order = "name"
        limit = 10
        page = True
        fields = ["name", "description"]

        profile_crud.list_profiles(order=order, limit=limit, page=page, fields=fields)

        mock_run_list_profiles.assert_called_once_with(
            limit=limit,
            page=page,
            order_by=order,
            fields=fields,
            get_profile_services=mock_container.get_profile_service,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
            get_display_field_resolver=mock_container.get_profile_display_field_resolver,
        )

    @patch("cmdbox.cli.commands.profile_crud.profile_handler.run_search_profiles")
    @patch("cmdbox.cli.commands.profile_crud.container")
    def test_search_profiles(self, mock_container, mock_run_search_profiles):
        term = "profile-search-term"
        limit = 15
        search_fields = ["description"]
        fields = ["name"]

        profile_crud.search(
            term=term, limit=limit, search_fields=search_fields, fields=fields
        )

        mock_run_search_profiles.assert_called_once_with(
            term=term,
            limit=limit,
            page=None,
            search_fields=search_fields,
            fields=fields,
            get_profile_services=mock_container.get_profile_service,
            get_settings=mock_container.get_settings,
            get_console=mock_container.get_console,
            get_display_field_resolver=mock_container.get_profile_display_field_resolver,
            get_search_field_resolver=mock_container.get_profile_search_field_resolver,
        )

    @patch("cmdbox.cli.commands.profile_crud.profile_handler.run_delete_profile")
    @patch("cmdbox.cli.commands.profile_crud.container")
    def test_delete(self, mock_container, mock_run_delete_profile):
        name = "test-profile"
        force = True

        profile_crud.delete(name=name, force=force)

        mock_run_delete_profile.assert_called_once_with(
            name=name,
            force=force,
            get_profile_services=mock_container.get_profile_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.profile_crud.profile_handler.run_profile_status")
    @patch("cmdbox.cli.commands.profile_crud.container")
    def test_status(self, mock_container, mock_run_profile_status):
        profile_crud.status()

        mock_run_profile_status.assert_called_once_with(
            get_profile_services=mock_container.get_profile_service,
            get_console=mock_container.get_console,
        )

    @patch(
        "cmdbox.cli.commands.profile_crud.profile_handler.run_switch_command_profile"
    )
    @patch("cmdbox.cli.commands.profile_crud.container")
    def test_switch_cmd(self, mock_container, mock_run_switch):
        name = "test-profile"

        profile_crud.switch_cmd(name=name)

        mock_run_switch.assert_called_once_with(
            name=name,
            get_profile_services=mock_container.get_profile_service,
            get_console=mock_container.get_console,
        )

    @patch(
        "cmdbox.cli.commands.profile_crud.profile_handler.run_switch_variable_profile"
    )
    @patch("cmdbox.cli.commands.profile_crud.container")
    def test_switch_var(self, mock_container, mock_run_switch):
        name = "test-profile"

        profile_crud.switch_var(name=name)

        mock_run_switch.assert_called_once_with(
            name=name,
            get_profile_services=mock_container.get_profile_service,
            get_console=mock_container.get_console,
        )

    @patch(
        "cmdbox.cli.commands.profile_crud.profile_handler.run_switch_settings_profile"
    )
    @patch("cmdbox.cli.commands.profile_crud.container")
    def test_switch_settings(self, mock_container, mock_run_switch):
        name = "test-profile"

        profile_crud.switch_settings(name=name)

        mock_run_switch.assert_called_once_with(
            name=name,
            get_profile_services=mock_container.get_profile_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.profile_crud.profile_handler.run_switch_profile")
    @patch("cmdbox.cli.commands.profile_crud.container")
    def test_switch(self, mock_container, mock_run_switch):
        name = "test-profile"
        cmd = True
        var = False
        settings = True

        profile_crud.switch(name=name, cmd=cmd, var=var, settings=settings)

        mock_run_switch.assert_called_once_with(
            name=name,
            cmd=cmd,
            var=var,
            settings=settings,
            get_profile_services=mock_container.get_profile_service,
            get_console=mock_container.get_console,
        )
