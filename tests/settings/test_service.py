import unittest
from unittest.mock import MagicMock
from dataclasses import dataclass
from cmdbox.settings.service import SettingsService, build_dataclass
from cmdbox.settings.models import Settings
from cmdbox.settings.repository import SettingsRepository


class TestBuildDataclass(unittest.TestCase):

    def test_build_simple_dataclass(self):
        @dataclass(frozen=True)
        class Simple:
            name: str
            age: int = 20

        data = {"name": "test", "age": 30}
        instance = build_dataclass(Simple, data)
        self.assertEqual(instance.name, "test")
        self.assertEqual(instance.age, 30)

    def test_build_nested_dataclass(self):
        data = {
            "ui": {"use_color": False},
            "execution_settings": {"default_shell": "pwsh"},
        }
        instance = build_dataclass(Settings, data)
        self.assertFalse(instance.ui.use_color)
        self.assertEqual(instance.execution_settings.default_shell, "pwsh")
        # Check defaults are preserved for missing fields in nested
        self.assertEqual(instance.ui.colors.success, "green")


class TestSettingsService(unittest.TestCase):

    def setUp(self):
        self.mock_repo = MagicMock(spec=SettingsRepository)
        self.mock_repo.load.return_value = {}
        self.service = SettingsService(self.mock_repo)

    def test_initial_load(self):
        # Service calls _load in __init__
        self.mock_repo.load.assert_called()
        self.assertEqual(self.service.get(), Settings())

    def test_load_with_overrides(self):
        self.mock_repo.load.return_value = {"ui": {"use_color": False}}
        service = SettingsService(self.mock_repo)
        self.assertFalse(service.get().ui.use_color)

    def test_reload(self):
        self.mock_repo.load.return_value = {"ui": {"use_color": False}}
        self.service.reload()
        self.assertFalse(self.service.get().ui.use_color)
        self.assertEqual(
            self.mock_repo.load.call_count, 2
        )  # once in init, once in reload

    def test_update(self):
        # mock_repo.load will be called:
        # 1. in __init__ (returns {})
        # 2. in update() to get current_raw (returns {"ui": {"use_color": True}})
        # 3. in update() -> _load() (returns {"ui": {"use_color": False}})
        self.mock_repo.load.side_effect = [
            {},
            {"ui": {"use_color": True}},
            {"ui": {"use_color": False}},
        ]
        service = SettingsService(self.mock_repo)

        patch = {"ui": {"use_color": False}}
        updated_settings = service.update(patch)

        # Verify save was called
        self.mock_repo.save.assert_called_once()
        saved_data = self.mock_repo.save.call_args[0][0]

        # Check specific values in saved data
        self.assertFalse(saved_data["ui"]["use_color"])

        self.assertFalse(updated_settings.ui.use_color)
        self.assertFalse(service.get().ui.use_color)

    def test_merge_simple(self):
        base = {"a": 1, "b": 2}
        override = {"b": 3, "c": 4}
        result = self.service._merge(base, override)
        self.assertEqual(result, {"a": 1, "b": 3, "c": 4})

    def test_merge_recursive(self):
        base = {"ui": {"use_color": True, "other": 1}}
        override = {"ui": {"use_color": False}}
        result = self.service._merge(base, override)
        self.assertEqual(result, {"ui": {"use_color": False, "other": 1}})
