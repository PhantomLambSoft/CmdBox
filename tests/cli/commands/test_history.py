import unittest
from unittest.mock import patch

from cmdbox.cli.commands import history


class TestHistoryCommand(unittest.TestCase):

    @patch("cmdbox.cli.commands.history.run_history_list")
    @patch("cmdbox.cli.commands.history.container")
    def test_history_list_defaults(self, mock_container, mock_run_history_list):
        history.history_list()

        mock_run_history_list.assert_called_once_with(
            alias=None,
            limit=25,
            profile=None,
            get_history_service=mock_container.get_history_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.history.run_history_list")
    @patch("cmdbox.cli.commands.history.container")
    def test_history_list_with_alias_and_limit(
        self, mock_container, mock_run_history_list
    ):
        history.history_list(alias="deploy", limit=5)

        mock_run_history_list.assert_called_once_with(
            alias="deploy",
            limit=5,
            profile=None,
            get_history_service=mock_container.get_history_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.history.run_history_list")
    @patch("cmdbox.cli.commands.history.container")
    def test_history_list_with_zero_limit(self, mock_container, mock_run_history_list):
        history.history_list(limit=0)

        mock_run_history_list.assert_called_once_with(
            alias=None,
            limit=0,
            profile=None,
            get_history_service=mock_container.get_history_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.history.run_history_show")
    @patch("cmdbox.cli.commands.history.container")
    def test_history_show_forwards_ref(self, mock_container, mock_run_history_show):
        history.history_show(ref="12")

        mock_run_history_show.assert_called_once_with(
            ref="12",
            profile=None,
            get_history_service=mock_container.get_history_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.history.run_history_rerun")
    @patch("cmdbox.cli.commands.history.container")
    def test_history_rerun_forwards_ref(self, mock_container, mock_run_history_rerun):
        history.history_rerun(ref="abc123")

        mock_run_history_rerun.assert_called_once_with(
            ref="abc123",
            profile=None,
            get_history_service=mock_container.get_history_service,
            get_run_service=mock_container.get_run_service,
        )

    @patch("cmdbox.cli.commands.history.run_history_clear")
    @patch("cmdbox.cli.commands.history.container")
    def test_history_clear_defaults(self, mock_container, mock_run_history_clear):
        history.history_clear()

        mock_run_history_clear.assert_called_once_with(
            alias=None,
            yes=False,
            profile=None,
            get_history_service=mock_container.get_history_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.history.run_history_clear")
    @patch("cmdbox.cli.commands.history.container")
    def test_history_clear_with_alias_and_yes(
        self, mock_container, mock_run_history_clear
    ):
        history.history_clear(alias="deploy", yes=True)

        mock_run_history_clear.assert_called_once_with(
            alias="deploy",
            yes=True,
            profile=None,
            get_history_service=mock_container.get_history_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.history.run_history_show")
    @patch("cmdbox.cli.commands.history.container")
    def test_history_show_forwards_id_prefix(
        self, mock_container, mock_run_history_show
    ):
        ref = "a1b2c3"
        history.history_show(ref=ref)

        kwargs = mock_run_history_show.call_args.kwargs
        self.assertEqual(ref, kwargs["ref"])
        self.assertEqual(
            mock_container.get_history_service, kwargs["get_history_service"]
        )
        self.assertEqual(mock_container.get_console, kwargs["get_console"])
