import unittest
from unittest.mock import MagicMock, patch
from cmdbox.cli.handlers.common_handlers import get_tags_interactive


class TestCommonHandlers(unittest.TestCase):

    @patch("cmdbox.cli.handlers.common_handlers.TagCompleter")
    @patch("cmdbox.cli.handlers.common_handlers.TagNameValidator")
    @patch("cmdbox.cli.handlers.common_handlers.prompt_for_tags")
    def test_get_tags_interactive(self, mock_prompt, mock_validator, mock_completer):
        mock_tag_services = MagicMock()
        mock_tag = MagicMock()
        mock_tag.name = "tag1"
        mock_tag_services.search.return_value = [mock_tag]
        mock_prompt.return_value = ["tag1"]

        result = get_tags_interactive(mock_tag_services)

        self.assertEqual(result, ["tag1"])

        # Verify TagCompleter callback
        args, kwargs = mock_completer.call_args
        callback = args[0]
        callback_result = callback("query")
        self.assertEqual(callback_result, ["tag1"])
        mock_tag_services.search.assert_called_with("query", fields="name")

    @patch("cmdbox.cli.handlers.common_handlers.prompt_for_tags")
    def test_get_tags_interactive_none(self, mock_prompt):
        mock_tag_services = MagicMock()
        mock_prompt.return_value = None

        result = get_tags_interactive(mock_tag_services)

        self.assertIsNone(result)
