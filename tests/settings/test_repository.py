import unittest
from pathlib import Path
from unittest.mock import MagicMock
from cmdbox.settings.repository import SettingsRepository


class TestSettingsRepository(unittest.TestCase):

    def setUp(self):
        self.settings_path = MagicMock(spec=Path)
        # Mocking the / operator for nested paths
        self.settings_path.__truediv__.return_value = self.settings_path
        self.repo = SettingsRepository(self.settings_path)

    def test_load_non_existent_file(self):
        # Should return an empty dict if file doesn't exist
        self.settings_path.exists.return_value = False
        result = self.repo.load()
        self.assertEqual(result, {})
        self.settings_path.exists.assert_called_once()

    def test_save_and_load(self):
        data = {
            "ui": {"use_color": False},
            "execution_settings": {"default_shell": "bash"},
        }

        # Mock read_text to return the saved data in TOML format
        # We need to simulate how tomlkit.dumps works or just check that write_text was called with some string
        self.repo.save(data)

        # Verify write_text was called
        self.settings_path.write_text.assert_called_once()
        args, kwargs = self.settings_path.write_text.call_args
        saved_content = args[0]

        # Verify load
        self.settings_path.exists.return_value = True
        self.settings_path.read_text.return_value = saved_content

        loaded_data = self.repo.load()
        self.assertEqual(loaded_data, data)

    def test_save_creates_parent_directories(self):
        parent_mock = MagicMock(spec=Path)
        self.settings_path.parent = parent_mock

        data = {"key": "value"}
        self.repo.save(data)

        parent_mock.mkdir.assert_called_once_with(parents=True, exist_ok=True)
        self.settings_path.write_text.assert_called_once()

    def test_save_overwrites_existing_file(self):
        self.repo.save({"a": 1})
        self.repo.save({"b": 2})

        # We can check that write_text was called twice
        self.assertEqual(self.settings_path.write_text.call_count, 2)

        # To check the content of the second call
        args, _ = self.settings_path.write_text.call_args_list[1]
        saved_content = args[0]

        self.settings_path.exists.return_value = True
        self.settings_path.read_text.return_value = saved_content
        loaded = self.repo.load()
        self.assertEqual(loaded, {"b": 2})
