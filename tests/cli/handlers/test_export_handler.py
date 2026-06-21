import unittest
from pathlib import Path
from unittest.mock import MagicMock, call, patch

from cmdbox.cli.handlers.export_handler import (
    run_export_all,
    run_export_cmds,
    run_export_vars,
)
from cmdbox.services.export_service import ExportResult


class TestExportHandler(unittest.TestCase):

    def setUp(self):
        self.mock_export_service = MagicMock()
        self.mock_console = MagicMock()
        self.get_export_service = lambda: self.mock_export_service
        self.get_console = lambda: self.mock_console

    @patch("cmdbox.cli.handlers.export_handler.render_export_result")
    def test_run_export_cmds_calls_service_and_prints_result_no_warnings(
        self, mock_render
    ):
        result = ExportResult(
            path=Path("out.json"),
            commands=["build"],
            variables=[],
            transient_commands=[],
            transient_variables=[],
            warnings=[],
        )
        self.mock_export_service.export_cmds.return_value = result
        mock_render.return_value = "rendered"

        run_export_cmds(
            aliases=["build"],
            tag="release",
            flatten=True,
            output="out.json",
            get_export_service=self.get_export_service,
            get_console=self.get_console,
        )

        self.mock_export_service.export_cmds.assert_called_once_with(
            aliases=["build"],
            tag="release",
            flatten=True,
            output_path="out.json",
        )
        mock_render.assert_called_once_with(result)
        self.mock_console.warning.assert_not_called()
        self.mock_console.print.assert_called_once_with("rendered")

    @patch("cmdbox.cli.handlers.export_handler.render_export_result")
    def test_run_export_cmds_prints_all_warnings_in_order(self, mock_render):
        result = ExportResult(
            path=Path("out.json"),
            warnings=["missing cmd:x", "missing var:y"],
        )
        self.mock_export_service.export_cmds.return_value = result
        mock_render.return_value = "rendered"

        run_export_cmds(
            aliases=None,
            tag=None,
            flatten=False,
            output=None,
            get_export_service=self.get_export_service,
            get_console=self.get_console,
        )

        self.mock_console.warning.assert_has_calls(
            [call("missing cmd:x"), call("missing var:y")]
        )
        self.assertEqual(2, self.mock_console.warning.call_count)
        self.mock_console.print.assert_called_once_with("rendered")

    @patch("cmdbox.cli.handlers.export_handler.render_export_result")
    def test_run_export_vars_calls_service_and_prints_result(self, mock_render):
        result = ExportResult(path=Path("vars.json"), warnings=[])
        self.mock_export_service.export_vars.return_value = result
        mock_render.return_value = "rendered_vars"

        run_export_vars(
            names=["API_KEY"],
            tag="prod",
            flatten=True,
            output="vars.json",
            get_export_service=self.get_export_service,
            get_console=self.get_console,
        )

        self.mock_export_service.export_vars.assert_called_once_with(
            names=["API_KEY"],
            tag="prod",
            flatten=True,
            output_path="vars.json",
        )
        mock_render.assert_called_once_with(result)
        self.mock_console.warning.assert_not_called()
        self.mock_console.print.assert_called_once_with("rendered_vars")

    @patch("cmdbox.cli.handlers.export_handler.render_export_result")
    def test_run_export_vars_handles_none_inputs(self, mock_render):
        result = ExportResult(path=Path("vars.json"), warnings=["warn"])
        self.mock_export_service.export_vars.return_value = result
        mock_render.return_value = "rendered_vars"

        run_export_vars(
            names=None,
            tag=None,
            flatten=False,
            output=None,
            get_export_service=self.get_export_service,
            get_console=self.get_console,
        )

        self.mock_export_service.export_vars.assert_called_once_with(
            names=None,
            tag=None,
            flatten=False,
            output_path=None,
        )
        self.mock_console.warning.assert_called_once_with("warn")
        self.mock_console.print.assert_called_once_with("rendered_vars")

    @patch("cmdbox.cli.handlers.export_handler.render_export_result")
    def test_run_export_all_calls_service_and_prints_result(self, mock_render):
        result = ExportResult(path=Path("all.json"), warnings=[])
        self.mock_export_service.export_all.return_value = result
        mock_render.return_value = "rendered_all"

        run_export_all(
            flatten=True,
            output="all.json",
            get_export_service=self.get_export_service,
            get_console=self.get_console,
        )

        self.mock_export_service.export_all.assert_called_once_with(
            flatten=True,
            output_path="all.json",
        )
        mock_render.assert_called_once_with(result)
        self.mock_console.warning.assert_not_called()
        self.mock_console.print.assert_called_once_with("rendered_all")

    @patch("cmdbox.cli.handlers.export_handler.render_export_result")
    def test_run_export_all_prints_multiple_warnings_before_result(self, mock_render):
        result = ExportResult(
            path=Path("all.json"),
            warnings=["warn-1", "warn-2", "warn-3"],
        )
        self.mock_export_service.export_all.return_value = result
        mock_render.return_value = "rendered_all"

        run_export_all(
            flatten=False,
            output=None,
            get_export_service=self.get_export_service,
            get_console=self.get_console,
        )

        self.mock_console.warning.assert_has_calls(
            [call("warn-1"), call("warn-2"), call("warn-3")]
        )
        self.assertEqual(3, self.mock_console.warning.call_count)
        self.mock_console.print.assert_called_once_with("rendered_all")
