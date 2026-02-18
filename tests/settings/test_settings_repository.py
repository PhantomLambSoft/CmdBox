import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch
from cmdbox.settings.settings_repository import SettingsRepository


class TestSettingsRepository(unittest.TestCase):

    def setUp(self):
        self.settings_path = MagicMock(spec=Path)
        self.repo = SettingsRepository(self.settings_path)

    def test_load_non_existent_file(self):
        self.settings_path.exists.return_value = False
        result = self.repo.load()
        self.assertEqual({}, result)
        self.settings_path.exists.assert_called_once()

    @patch("cmdbox.settings.settings_repository.atomic_write_text")
    def test_save(self, mock_atomic_write):
        data = {
            "ui": {"use_color": False},
            "execution_settings": {"default_shell": "bash"},
        }

        self.repo.save(data)

        mock_atomic_write.assert_called_once()
        args, kwargs = mock_atomic_write.call_args
        self.assertEqual(self.settings_path, args[0])
        saved_content = args[1]
        self.assertIn("use_color = false", saved_content)
        self.assertIn('default_shell = "bash"', saved_content)
        self.assertEqual("utf-8", kwargs.get("encoding"))

    @patch("cmdbox.settings.settings_repository.atomic_write_text")
    def test_save_and_load(self, mock_atomic_write):
        data = {
            "ui": {"use_color": False},
            "execution_settings": {"default_shell": "bash"},
        }

        # Mock save content capture
        def side_effect(path, content, encoding="utf-8"):
            self.settings_path.read_text.return_value = content

        mock_atomic_write.side_effect = side_effect

        self.repo.save(data)

        # Verify load
        self.settings_path.exists.return_value = True
        loaded_data = self.repo.load()
        self.assertEqual(data, loaded_data)

    def test_dict_to_text(self):
        data = {"a": 1, "b": {"c": 2}}
        text = self.repo.dict_to_text(data)
        self.assertIn("a = 1", text)
        self.assertIn("[b]\nc = 2", text)

    @patch("cmdbox.settings.settings_repository.atomic_write_text")
    def test_save_overwrites_existing_file(self, mock_atomic_write):
        self.repo.save({"a": 1})
        self.repo.save({"b": 2})

        self.assertEqual(2, mock_atomic_write.call_count)

        # To check the content of the second call
        args, _ = mock_atomic_write.call_args_list[1]
        saved_content = args[1]

        self.settings_path.exists.return_value = True
        self.settings_path.read_text.return_value = saved_content
        loaded = self.repo.load()
        self.assertEqual({"b": 2}, loaded)
