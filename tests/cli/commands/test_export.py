import unittest
from unittest.mock import patch

from cmdbox.cli.commands import export


class TestExportCommand(unittest.TestCase):

    @patch("cmdbox.cli.commands.export.run_export_cmds")
    @patch("cmdbox.cli.commands.export.container")
    def test_export_cmds_defaults(self, mock_container, mock_run_export_cmds):
        export.export_cmds()

        mock_run_export_cmds.assert_called_once_with(
            aliases=None,
            tag=None,
            flatten=False,
            output=None,
            get_export_service=mock_container.get_export_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.export.run_export_cmds")
    @patch("cmdbox.cli.commands.export.container")
    def test_export_cmds_forwards_explicit_values(
        self, mock_container, mock_run_export_cmds
    ):
        aliases = ["build", "deploy"]

        export.export_cmds(
            aliases=aliases,
            tag="release",
            flatten=True,
            output="exports\\cmds.json",
        )

        mock_run_export_cmds.assert_called_once_with(
            aliases=aliases,
            tag="release",
            flatten=True,
            output="exports\\cmds.json",
            get_export_service=mock_container.get_export_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.export.run_export_cmds")
    @patch("cmdbox.cli.commands.export.container")
    def test_export_cmds_forwards_empty_aliases_list(
        self, mock_container, mock_run_export_cmds
    ):
        export.export_cmds(aliases=[])

        kwargs = mock_run_export_cmds.call_args.kwargs
        self.assertEqual([], kwargs["aliases"])
        self.assertEqual(None, kwargs["tag"])
        self.assertEqual(False, kwargs["flatten"])
        self.assertEqual(None, kwargs["output"])
        self.assertEqual(
            mock_container.get_export_service, kwargs["get_export_service"]
        )
        self.assertEqual(mock_container.get_console, kwargs["get_console"])

    @patch("cmdbox.cli.commands.export.run_export_vars")
    @patch("cmdbox.cli.commands.export.container")
    def test_export_vars_defaults(self, mock_container, mock_run_export_vars):
        export.export_vars()

        mock_run_export_vars.assert_called_once_with(
            names=None,
            tag=None,
            flatten=False,
            output=None,
            get_export_service=mock_container.get_export_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.export.run_export_vars")
    @patch("cmdbox.cli.commands.export.container")
    def test_export_vars_forwards_explicit_values(
        self, mock_container, mock_run_export_vars
    ):
        names = ["API_KEY", "ENV"]

        export.export_vars(
            names=names,
            tag="prod",
            flatten=True,
            output="vars.json",
        )

        mock_run_export_vars.assert_called_once_with(
            names=names,
            tag="prod",
            flatten=True,
            output="vars.json",
            get_export_service=mock_container.get_export_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.export.run_export_vars")
    @patch("cmdbox.cli.commands.export.container")
    def test_export_vars_forwards_empty_names_list(
        self, mock_container, mock_run_export_vars
    ):
        export.export_vars(names=[])

        kwargs = mock_run_export_vars.call_args.kwargs
        self.assertEqual([], kwargs["names"])
        self.assertEqual(None, kwargs["tag"])
        self.assertEqual(False, kwargs["flatten"])
        self.assertEqual(None, kwargs["output"])
        self.assertEqual(
            mock_container.get_export_service, kwargs["get_export_service"]
        )
        self.assertEqual(mock_container.get_console, kwargs["get_console"])

    @patch("cmdbox.cli.commands.export.run_export_all")
    @patch("cmdbox.cli.commands.export.container")
    def test_export_all_defaults(self, mock_container, mock_run_export_all):
        export.export_all()

        mock_run_export_all.assert_called_once_with(
            flatten=False,
            output=None,
            get_export_service=mock_container.get_export_service,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.export.run_export_all")
    @patch("cmdbox.cli.commands.export.container")
    def test_export_all_forwards_explicit_values(
        self, mock_container, mock_run_export_all
    ):
        export.export_all(flatten=True, output="all-export.json")

        mock_run_export_all.assert_called_once_with(
            flatten=True,
            output="all-export.json",
            get_export_service=mock_container.get_export_service,
            get_console=mock_container.get_console,
        )
