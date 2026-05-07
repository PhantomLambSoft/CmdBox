import unittest
from unittest.mock import MagicMock, patch

from cmdbox.cli.handlers.history_handlers import (
    run_history_list,
    run_history_show,
    run_history_rerun,
    run_history_clear,
    run_rerun_last,
)
from cmdbox.runtime.executor import RunContext


class TestHistoryHandlers(unittest.TestCase):

    @patch("cmdbox.cli.handlers.history_handlers.render_history_list")
    def test_run_history_list_renders_entries(self, mock_render_history_list):
        entries = [MagicMock(), MagicMock()]
        mock_service = MagicMock()
        mock_service.get_recent.return_value = entries
        mock_console = MagicMock()
        mock_render_history_list.return_value = "rendered-list"

        run_history_list(
            alias="deploy",
            limit=10,
            get_history_service=lambda: mock_service,
            get_console=lambda: mock_console,
        )

        mock_service.get_recent.assert_called_once_with(alias="deploy", limit=10)
        mock_render_history_list.assert_called_once_with(entries)
        mock_console.print.assert_called_once_with("rendered-list")
        mock_console.info.assert_not_called()

    @patch("cmdbox.cli.handlers.history_handlers.render_history_list")
    def test_run_history_list_shows_empty_message_when_no_entries(
        self, mock_render_history_list
    ):
        mock_service = MagicMock()
        mock_service.get_recent.return_value = []
        mock_console = MagicMock()

        run_history_list(
            alias=None,
            limit=25,
            get_history_service=lambda: mock_service,
            get_console=lambda: mock_console,
        )

        mock_service.get_recent.assert_called_once_with(alias=None, limit=25)
        mock_console.info.assert_called_once_with("No history found")
        mock_console.print.assert_not_called()
        mock_render_history_list.assert_not_called()

    @patch("cmdbox.cli.handlers.history_handlers.render_history_entry")
    def test_run_history_show_loads_entry_and_variables(
        self, mock_render_history_entry
    ):
        entry = MagicMock()
        variables = {"ENV": "prod"}
        mock_service = MagicMock()
        mock_service.get_by_ref.return_value = entry
        mock_service.get_variables.return_value = variables
        mock_console = MagicMock()
        mock_render_history_entry.return_value = "rendered-entry"

        run_history_show(
            ref="12",
            get_history_service=lambda: mock_service,
            get_console=lambda: mock_console,
        )

        mock_service.get_by_ref.assert_called_once_with("12")
        mock_service.get_variables.assert_called_once_with(entry)
        mock_render_history_entry.assert_called_once_with(entry, variables)
        mock_console.print.assert_called_once_with("rendered-entry")

    def test_run_history_rerun_forwards_alias_vars_and_context(self):
        entry = MagicMock()
        entry.alias = "deploy"
        variables = {"A": "1"}
        mock_history_service = MagicMock()
        mock_history_service.get_by_ref.return_value = entry
        mock_history_service.get_variables.return_value = variables
        mock_run_service = MagicMock()

        run_history_rerun(
            ref="abc123",
            get_history_service=lambda: mock_history_service,
            get_run_service=lambda: mock_run_service,
        )

        mock_history_service.get_by_ref.assert_called_once_with("abc123")
        mock_history_service.get_variables.assert_called_once_with(entry)
        mock_run_service.run.assert_called_once()
        args, kwargs = mock_run_service.run.call_args
        self.assertEqual("deploy", args[0])
        self.assertEqual(variables, kwargs["runtime_vars"])
        self.assertIsInstance(kwargs["ctx"], RunContext)

    @patch("cmdbox.cli.handlers.history_handlers.render_history_cleared")
    @patch("cmdbox.cli.handlers.history_handlers.prompt_for_confirm")
    def test_run_history_clear_yes_true_skips_prompt_and_clears(
        self, mock_prompt_for_confirm, mock_render_history_cleared
    ):
        mock_service = MagicMock()
        mock_service.clear.return_value = 3
        mock_console = MagicMock()
        mock_render_history_cleared.return_value = "cleared"

        run_history_clear(
            alias=None,
            yes=True,
            get_history_service=lambda: mock_service,
            get_console=lambda: mock_console,
        )

        mock_prompt_for_confirm.assert_not_called()
        mock_service.clear.assert_called_once_with(alias=None)
        mock_render_history_cleared.assert_called_once_with(3, None)
        mock_console.print.assert_called_once_with("cleared")
        mock_console.info.assert_not_called()

    @patch("cmdbox.cli.handlers.history_handlers.render_history_cleared")
    @patch("cmdbox.cli.handlers.history_handlers.prompt_for_confirm")
    def test_run_history_clear_confirm_accepts_and_uses_alias_scope(
        self, mock_prompt_for_confirm, mock_render_history_cleared
    ):
        mock_prompt_for_confirm.return_value = True
        mock_service = MagicMock()
        mock_service.clear.return_value = 1
        mock_console = MagicMock()
        mock_render_history_cleared.return_value = "cleared-one"

        run_history_clear(
            alias="deploy",
            yes=False,
            get_history_service=lambda: mock_service,
            get_console=lambda: mock_console,
        )

        mock_prompt_for_confirm.assert_called_once_with(
            "Clear all history for 'deploy'?"
        )
        mock_service.clear.assert_called_once_with(alias="deploy")
        mock_render_history_cleared.assert_called_once_with(1, "deploy")
        mock_console.print.assert_called_once_with("cleared-one")
        mock_console.info.assert_not_called()

    @patch("cmdbox.cli.handlers.history_handlers.render_history_cleared")
    @patch("cmdbox.cli.handlers.history_handlers.prompt_for_confirm")
    def test_run_history_clear_confirm_declines_aborts(
        self, mock_prompt_for_confirm, mock_render_history_cleared
    ):
        mock_prompt_for_confirm.return_value = False
        mock_service = MagicMock()
        mock_console = MagicMock()

        run_history_clear(
            alias=None,
            yes=False,
            get_history_service=lambda: mock_service,
            get_console=lambda: mock_console,
        )

        mock_prompt_for_confirm.assert_called_once_with("Clear all history?")
        mock_console.info.assert_called_once_with("Aborted")
        mock_service.clear.assert_not_called()
        mock_console.print.assert_not_called()
        mock_render_history_cleared.assert_not_called()

    def test_run_history_rerun_last_no_entries(self):
        mock_service = MagicMock()
        mock_service.get_recent.return_value = []
        mock_run_service = MagicMock()
        mock_console = MagicMock()

        run_rerun_last(
            get_history_service=lambda: mock_service,
            get_run_service=lambda: mock_run_service,
            get_console=lambda: mock_console,
        )

        mock_console.info.assert_called_once_with("No command history found")
        mock_run_service.run.assert_not_called()
        mock_service.get_recent.assert_called_once_with(limit=1)

    def test_run_history_rerun_last_with_entries(self):
        entry = MagicMock()
        entry.alias = "deploy"
        variables = {"ENV": "prod"}
        mock_service = MagicMock()
        mock_run_service = MagicMock()
        mock_console = MagicMock()
        mock_service.get_recent.return_value = [entry]
        mock_service.get_variables.return_value = variables

        run_rerun_last(
            get_run_service=lambda: mock_run_service,
            get_history_service=lambda: mock_service,
            get_console=lambda: mock_console,
        )

        mock_run_service.run.assert_called_once_with("deploy", runtime_vars=variables)
        mock_service.get_recent.assert_called_once_with(limit=1)
        mock_service.get_variables.assert_called_once_with(entry)
