import unittest
import logging
from unittest.mock import MagicMock, patch
from pathlib import Path
from cmdbox.logging_setup.log_config import (
    LogConfig,
    _level,
    get_console_level,
    get_file_enabled,
    get_file_level,
    get_logger,
    build_log_config,
    LOGGER_NAME,
)


class TestLogConfigModule(unittest.TestCase):
    def test_level_valid_string(self):
        self.assertEqual(logging.DEBUG, _level("DEBUG"))
        self.assertEqual(logging.INFO, _level("info"))
        self.assertEqual(logging.WARNING, _level("  warning  "))
        self.assertEqual(logging.ERROR, _level("Error"))
        self.assertEqual(logging.CRITICAL, _level("CRITICAL"))

    def test_level_invalid_string_defaults_to_info(self):
        self.assertEqual(logging.INFO, _level("INVALID"))
        self.assertEqual(logging.INFO, _level(""))
        self.assertEqual(logging.INFO, _level(None))

    def test_get_console_level_debug_priority(self):
        settings = MagicMock()
        self.assertEqual(
            logging.DEBUG, get_console_level(settings, verbose=False, debug=True)
        )
        self.assertEqual(
            logging.DEBUG, get_console_level(settings, verbose=True, debug=True)
        )

    def test_get_console_level_verbose_priority_after_debug(self):
        settings = MagicMock()
        self.assertEqual(
            logging.INFO, get_console_level(settings, verbose=True, debug=False)
        )

    def test_get_console_level_from_settings(self):
        settings = MagicMock()
        settings.logging.console_level = "DEBUG"
        self.assertEqual(
            logging.DEBUG, get_console_level(settings, verbose=False, debug=False)
        )

        settings.logging.console_level = "WARNING"
        self.assertEqual(
            logging.WARNING, get_console_level(settings, verbose=False, debug=False)
        )

    def test_get_file_enabled_override(self):
        settings = MagicMock()
        self.assertTrue(get_file_enabled(settings, file_logs=True))
        self.assertFalse(get_file_enabled(settings, file_logs=False))

    def test_get_file_enabled_from_settings(self):
        settings = MagicMock()
        settings.logging.file.enabled = True
        self.assertTrue(get_file_enabled(settings, file_logs=None))

        settings.logging.file.enabled = False
        self.assertFalse(get_file_enabled(settings, file_logs=None))

    def test_get_file_level_debug_priority(self):
        settings = MagicMock()
        self.assertEqual(
            logging.DEBUG, get_file_level(settings, verbose=False, debug=True)
        )
        self.assertEqual(
            logging.DEBUG, get_file_level(settings, verbose=True, debug=True)
        )

    def test_get_file_level_verbose_priority_after_debug(self):
        settings = MagicMock()
        self.assertEqual(
            logging.INFO, get_file_level(settings, verbose=True, debug=False)
        )

    def test_get_file_level_from_settings(self):
        settings = MagicMock()
        settings.logging.file.level = "DEBUG"
        self.assertEqual(
            logging.DEBUG, get_file_level(settings, verbose=False, debug=False)
        )

        settings.logging.file.level = "WARNING"
        self.assertEqual(
            logging.WARNING, get_file_level(settings, verbose=False, debug=False)
        )

    def test_get_logger(self):
        logger = get_logger()
        self.assertEqual(LOGGER_NAME, logger.name)
        self.assertIsInstance(logger, logging.Logger)

    @patch("cmdbox.logging_setup.log_config.get_log_file_path")
    def test_build_log_config(self, mock_get_log_file_path):
        mock_path = Path("/tmp/test.log")
        mock_get_log_file_path.return_value = mock_path

        settings = MagicMock()
        settings.logging.console_level = "WARNING"
        settings.logging.file.enabled = True
        settings.logging.file.level = "ERROR"
        settings.logging.file.max_bytes = 1024
        settings.logging.file.backups = 5

        config = build_log_config(settings, verbose=False, debug=False, file_logs=None)

        self.assertIsInstance(config, LogConfig)
        self.assertEqual(logging.WARNING, config.console_level)
        self.assertTrue(config.file_enabled)
        self.assertEqual(logging.ERROR, config.file_level)
        self.assertEqual(mock_path, config.file_path)
        self.assertEqual(1024, config.max_bytes)
        self.assertEqual(5, config.backups)

    @patch("cmdbox.logging_setup.log_config.get_log_file_path")
    def test_build_log_config_with_flags(self, mock_get_log_file_path):
        mock_path = Path("/tmp/test.log")
        mock_get_log_file_path.return_value = mock_path

        settings = MagicMock()
        settings.logging.console_level = "WARNING"
        settings.logging.file.enabled = False
        settings.logging.file.level = "ERROR"
        settings.logging.file.max_bytes = 1024
        settings.logging.file.backups = 5

        # Override with debug=True, file_logs=True
        config = build_log_config(settings, verbose=False, debug=True, file_logs=True)

        self.assertEqual(logging.DEBUG, config.console_level)
        self.assertTrue(config.file_enabled)
        self.assertEqual(logging.DEBUG, config.file_level)
