import unittest
from unittest.mock import MagicMock
from pathlib import Path

from cmdbox.models import Profile, ProfileState
from cmdbox.services.profile_services import ProfileServices, ProfileStatus


class TestProfileStatus(unittest.TestCase):

    def test_linked_true(self):
        status = ProfileStatus(
            command_profile="default",
            variable_profile="default",
            settings_profile="default",
        )
        self.assertTrue(status.linked)

    def test_linked_false(self):
        status = ProfileStatus(
            command_profile="default",
            variable_profile="custom",
            settings_profile="default",
        )
        self.assertFalse(status.linked)


class TestProfileServices(unittest.TestCase):

    def setUp(self):
        self.mock_repo = MagicMock()
        self.services = ProfileServices(self.mock_repo)

    def test_create_profile(self):
        name = "test-profile"
        description = "test description"
        expected_profile = MagicMock(spec=Profile)
        self.mock_repo.create.return_value = expected_profile

        result = self.services.create_profile(name, description)

        self.assertEqual(expected_profile, result)
        self.mock_repo.create.assert_called_once_with(name, description)

    def test_update_profile(self):
        name = "test-profile"
        profile = MagicMock(spec=Profile)
        updated_profile = MagicMock(spec=Profile)
        self.mock_repo.get_by_name.return_value = profile
        self.mock_repo.update.return_value = updated_profile

        result = self.services.update_profile(name, description="new description")

        self.assertEqual(updated_profile, result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.update.assert_called_once_with(
            profile, description="new description"
        )

    def test_delete_profile(self):
        name = "test-profile"
        profile = MagicMock(spec=Profile)
        self.mock_repo.get_by_name.return_value = profile
        self.mock_repo.delete.return_value = True

        result = self.services.delete_profile(name)

        self.assertTrue(result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.delete.assert_called_once_with(profile, force=False)

    def test_delete_profile_force(self):
        name = "test-profile"
        profile = MagicMock(spec=Profile)
        self.mock_repo.get_by_name.return_value = profile
        self.mock_repo.delete.return_value = True

        result = self.services.delete_profile(name, force=True)

        self.assertTrue(result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.delete.assert_called_once_with(profile, force=True)

    def test_get_profile(self):
        name = "test-profile"
        expected_profile = MagicMock(spec=Profile)
        self.mock_repo.get_by_name.return_value = expected_profile

        result = self.services.get_profile(name)

        self.assertEqual(expected_profile, result)
        self.mock_repo.get_by_name.assert_called_once_with(name)

    def test_list_profiles(self):
        expected_profiles = [MagicMock(spec=Profile), MagicMock(spec=Profile)]
        self.mock_repo.list_all.return_value = expected_profiles

        result = self.services.list_profiles(order_by="name", limit=10)

        self.assertEqual(expected_profiles, result)
        self.mock_repo.list_all.assert_called_once_with(order_by="name", limit=10)

    def test_switch_profile(self):
        name = "test-profile"
        profile = MagicMock(spec=Profile)
        expected_state = MagicMock(spec=ProfileState)
        self.mock_repo.get_by_name.return_value = profile
        self.mock_repo.set_active_settings_profile.return_value = expected_state

        result = self.services.switch_profile(name)

        self.assertEqual(expected_state, result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.set_active_command_profile.assert_called_once_with(profile)
        self.mock_repo.set_active_variable_profile.assert_called_once_with(profile)
        self.mock_repo.set_active_settings_profile.assert_called_once_with(profile)

    def test_switch_command_profile(self):
        name = "test-profile"
        profile = MagicMock(spec=Profile)
        expected_state = MagicMock(spec=ProfileState)
        self.mock_repo.get_by_name.return_value = profile
        self.mock_repo.set_active_command_profile.return_value = expected_state

        result = self.services.switch_command_profile(name)

        self.assertEqual(expected_state, result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.set_active_command_profile.assert_called_once_with(profile)

    def test_switch_variable_profile(self):
        name = "test-profile"
        profile = MagicMock(spec=Profile)
        expected_state = MagicMock(spec=ProfileState)
        self.mock_repo.get_by_name.return_value = profile
        self.mock_repo.set_active_variable_profile.return_value = expected_state

        result = self.services.switch_variable_profile(name)

        self.assertEqual(expected_state, result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.set_active_variable_profile.assert_called_once_with(profile)

    def test_switch_settings_profile(self):
        name = "test-profile"
        profile = MagicMock(spec=Profile)
        expected_state = MagicMock(spec=ProfileState)
        self.mock_repo.get_by_name.return_value = profile
        self.mock_repo.set_active_settings_profile.return_value = expected_state

        result = self.services.switch_settings_profile(name)

        self.assertEqual(expected_state, result)
        self.mock_repo.get_by_name.assert_called_once_with(name)
        self.mock_repo.set_active_settings_profile.assert_called_once_with(profile)

    def test_get_status(self):
        state = MagicMock()
        state.active_command_profile = "cmd"
        state.active_variable_profile = "var"
        state.active_settings_profile = "settings"
        self.mock_repo.get_state.return_value = state

        result = self.services.get_status()

        self.assertIsInstance(result, ProfileStatus)
        self.assertEqual("cmd", result.command_profile)
        self.assertEqual("var", result.variable_profile)
        self.assertEqual("settings", result.settings_profile)

    def test_resolve_settings_path_default(self):
        app_data_dir = Path("/app/data")
        result = self.services.resolve_settings_path(app_data_dir, "default")
        self.assertEqual(app_data_dir / "config.toml", result)

    def test_resolve_settings_path_custom(self):
        app_data_dir = Path("/app/data")
        result = self.services.resolve_settings_path(app_data_dir, "custom")
        self.assertEqual(app_data_dir / "custom_config.toml", result)

    def test_resolve_settings_path_from_state(self):
        app_data_dir = Path("/app/data")
        state = MagicMock()
        state.active_settings_profile.name = "active"
        self.mock_repo.get_state.return_value = state

        result = self.services.resolve_settings_path(app_data_dir)

        self.assertEqual(app_data_dir / "active_config.toml", result)
        self.mock_repo.get_state.assert_called_once()
