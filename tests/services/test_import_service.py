import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import MagicMock, patch

from cmdbox.services.errors import ImportCycleError, ImportFileError
from cmdbox.services.import_service import (
    ImportResult,
    ImportService,
    _label,
    _parse_import_file,
    build_dependency_graph,
    validate_no_cycles,
)


def make_tag(name):
    return SimpleNamespace(tag=SimpleNamespace(name=name))


class TestImportHelpers(unittest.TestCase):

    def test_label_formats_command_and_variable_prefixes(self):
        from cmdbox.resolve.type_defs import RefKind

        self.assertEqual("cmd:build", _label(RefKind.COMMAND, "build"))
        self.assertEqual("var:host", _label(RefKind.VARIABLE, "host"))

    def test_build_dependency_graph_reads_command_and_variable_references(self):
        import_data = {
            "commands": [
                {"alias": "deploy", "template": "run <cmd:build> <var:env>"},
            ],
            "variables": [
                {"name": "env", "value": "prod-<var:region>"},
            ],
        }

        deps = build_dependency_graph(import_data)

        self.assertEqual(["cmd:build", "var:env"], deps["cmd:deploy"])
        self.assertEqual(["var:region"], deps["var:env"])

    def test_build_dependency_graph_handles_missing_sections(self):
        deps = build_dependency_graph({})
        self.assertEqual({}, deps)

    def test_validate_no_cycles_passes_for_acyclic_graph(self):
        deps = {
            "cmd:deploy": ["cmd:build", "var:env"],
            "cmd:build": [],
            "var:env": ["var:region"],
            "var:region": [],
        }

        validate_no_cycles(deps)

    def test_validate_no_cycles_raises_import_cycle_error_with_cycle(self):
        deps = {
            "cmd:a": ["var:x"],
            "var:x": ["cmd:a"],
        }

        with self.assertRaises(ImportCycleError) as ctx:
            validate_no_cycles(deps)

        self.assertEqual(["cmd:a", "var:x", "cmd:a"], ctx.exception.cycle)

    @patch("cmdbox.services.import_service.Path.read_text")
    def test_parse_import_file_returns_data_for_valid_version(self, mock_read_text):
        mock_read_text.return_value = (
            '{"version": "1", "commands": [], "variables": []}'
        )

        result = _parse_import_file("C:\\imports\\ok.json")

        self.assertEqual("1", result["version"])
        self.assertEqual([], result["commands"])
        self.assertEqual([], result["variables"])

    @patch("cmdbox.services.import_service.Path.read_text")
    def test_parse_import_file_raises_for_invalid_json(self, mock_read_text):
        mock_read_text.return_value = "{not-json"

        with self.assertRaises(ImportFileError) as ctx:
            _parse_import_file("C:\\imports\\bad.json")

        self.assertIn("Could not read import file", str(ctx.exception))

    @patch("cmdbox.services.import_service.Path.read_text")
    def test_parse_import_file_raises_for_non_object_root(self, mock_read_text):
        mock_read_text.return_value = '[{"version": "1"}]'

        with self.assertRaises(ImportFileError) as ctx:
            _parse_import_file("C:\\imports\\array.json")

        self.assertEqual("Import file must be a JSON object", str(ctx.exception))

    @patch("cmdbox.services.import_service.Path.read_text")
    def test_parse_import_file_raises_for_unsupported_version(self, mock_read_text):
        mock_read_text.return_value = (
            '{"version": "2", "commands": [], "variables": []}'
        )

        with self.assertRaises(ImportFileError) as ctx:
            _parse_import_file(Path("C:\\imports\\v2.json"))

        self.assertIn("Unsupported import file version: 2", str(ctx.exception))
        self.assertIn("Supported versions: 1", str(ctx.exception))


class TestImportService(unittest.TestCase):
    def setUp(self):
        self.cmd_service = MagicMock()
        self.var_service = MagicMock()
        self.tag_service = MagicMock()
        self.profile_repo = MagicMock()
        self.service = ImportService(
            self.cmd_service, self.var_service, self.tag_service, self.profile_repo
        )

    @patch("cmdbox.services.import_service.validate_no_cycles")
    @patch("cmdbox.services.import_service.build_dependency_graph")
    @patch("cmdbox.services.import_service._parse_import_file")
    def test_import_file_preview_routes_actions_without_writes(
        self, mock_parse, mock_build_dependency_graph, mock_validate_no_cycles
    ):
        mock_parse.return_value = {
            "version": "1",
            "commands": [
                {"alias": "create-cmd", "template": "echo create"},
                {"alias": "skip-cmd", "template": "echo skip"},
                {"alias": "overwrite-cmd", "template": "echo overwrite"},
            ],
            "variables": [
                {"name": "create-var", "value": "1"},
                {"name": "skip-var", "value": "2"},
                {"name": "overwrite-var", "value": "3"},
            ],
        }
        self.cmd_service.get_command_or_none.side_effect = lambda alias, profile=None: {
            "skip-cmd": object(),
            "overwrite-cmd": object(),
        }.get(alias)
        self.var_service.get_variable_or_none.side_effect = lambda name, profile=None: {
            "skip-var": object(),
            "overwrite-var": object(),
        }.get(name)

        result = self.service.import_file(
            "C:\\imports\\preview.json", overwrite=True, preview=True
        )

        self.assertIsInstance(result, ImportResult)
        self.assertTrue(result.preview)
        self.assertEqual(["create-cmd"], result.commands_created)
        self.assertEqual([], result.commands_skipped)
        self.assertEqual(["skip-cmd", "overwrite-cmd"], result.commands_overwritten)
        self.assertEqual(["create-var"], result.variables_created)
        self.assertEqual([], result.variables_skipped)
        self.assertEqual(["skip-var", "overwrite-var"], result.variables_overwritten)
        self.cmd_service.create_command.assert_not_called()
        self.cmd_service.update_command.assert_not_called()
        self.var_service.create_variable.assert_not_called()
        self.var_service.update_variable.assert_not_called()
        mock_build_dependency_graph.assert_called_once_with(mock_parse.return_value)
        mock_validate_no_cycles.assert_called_once()

    @patch("cmdbox.services.import_service.validate_no_cycles")
    @patch("cmdbox.services.import_service.build_dependency_graph")
    @patch("cmdbox.services.import_service._parse_import_file")
    def test_import_file_non_preview_handles_create_skip_and_overwrite(
        self, mock_parse, _mock_build_dependency_graph, _mock_validate_no_cycles
    ):
        mock_parse.return_value = {
            "version": "1",
            "commands": [
                {
                    "alias": "create-cmd",
                    "template": "echo c",
                    "tags": ["ops"],
                    "description": "desc",
                    "cwd": "C:/repo",
                    "shell": "pwsh",
                    "env": {"A": "B"},
                    "timeout": 10,
                },
                {"alias": "skip-cmd", "template": "echo s", "tags": ["same"]},
                {
                    "alias": "overwrite-cmd",
                    "template": "echo o",
                    "tags": ["new", "shared"],
                },
            ],
            "variables": [
                {"name": "create-var", "value": "c", "tags": ["prod"]},
                {"name": "skip-var", "value": "s", "tags": ["same"]},
                {
                    "name": "overwrite-var",
                    "value": "o",
                    "tags": ["new", "shared"],
                },
            ],
        }
        self.cmd_service.get_command_or_none.side_effect = lambda alias, profile=None: {
            "skip-cmd": object(),
            "overwrite-cmd": object(),
        }.get(alias)
        self.var_service.get_variable_or_none.side_effect = lambda name, profile=None: {
            "skip-var": object(),
            "overwrite-var": object(),
        }.get(name)
        self.cmd_service.get_command.return_value = SimpleNamespace(
            tags=[make_tag("old"), make_tag("shared")]
        )
        self.var_service.get_variable.return_value = SimpleNamespace(
            tags=[make_tag("old"), make_tag("shared")]
        )

        result = self.service.import_file(
            "C:\\imports\\apply.json", overwrite=False, preview=False
        )

        self.assertEqual(["create-cmd"], result.commands_created)
        self.assertEqual(["skip-cmd", "overwrite-cmd"], result.commands_skipped)
        self.assertEqual([], result.commands_overwritten)
        self.assertEqual(["create-var"], result.variables_created)
        self.assertEqual(["skip-var", "overwrite-var"], result.variables_skipped)
        self.assertEqual([], result.variables_overwritten)
        self.cmd_service.create_command.assert_called_once_with(
            alias="create-cmd",
            template="echo c",
            description="desc",
            tags=["ops"],
            cwd="C:/repo",
            shell="pwsh",
            env={"A": "B"},
            timeout=10,
            profile=None,
        )
        self.var_service.create_variable.assert_called_once_with(
            "create-var", value="c", tags=["prod"], profile=None
        )

    def test_handle_commands_overwrite_updates_and_reconciles_tags(self):
        self.cmd_service.get_command.return_value = SimpleNamespace(
            tags=[make_tag("old"), make_tag("shared")]
        )

        self.service.handle_commands(
            [
                (
                    {
                        "alias": "deploy",
                        "template": "echo",
                        "description": "d",
                        "cwd": "C:/repo",
                        "shell": "pwsh",
                        "env": {"A": "B"},
                        "timeout": 20,
                        "tags": ["new", "shared"],
                    },
                    "overwrite",
                )
            ]
        )

        self.cmd_service.update_command.assert_called_once_with(
            current_alias="deploy",
            template="echo",
            description="d",
            cwd="C:/repo",
            shell="pwsh",
            env={"A": "B"},
            timeout=20,
            profile=None,
        )
        self.cmd_service.remove_tags.assert_called_once_with(
            "deploy", ["old"], profile=None
        )
        self.cmd_service.add_tags.assert_called_once_with(
            "deploy", ["new"], profile=None
        )

    def test_handle_commands_overwrite_skips_tag_mutations_when_unchanged(self):
        self.cmd_service.get_command.return_value = SimpleNamespace(
            tags=[make_tag("ops"), make_tag("shared")]
        )

        self.service.handle_commands(
            [
                (
                    {
                        "alias": "deploy",
                        "template": "echo",
                        "tags": ["shared", "ops"],
                    },
                    "overwrite",
                )
            ]
        )

        self.cmd_service.remove_tags.assert_not_called()
        self.cmd_service.add_tags.assert_not_called()

    def test_handle_variables_overwrite_updates_and_reconciles_tags(self):
        self.var_service.get_variable.return_value = SimpleNamespace(
            tags=[make_tag("old"), make_tag("shared")]
        )

        self.service.handle_variables(
            [
                (
                    {"name": "host", "value": "prod", "tags": ["new", "shared"]},
                    "overwrite",
                )
            ]
        )

        self.var_service.update_variable.assert_called_once_with(
            "host", value="prod", profile=None
        )
        self.var_service.remove_tags.assert_called_once_with(
            "host", ["old"], profile=None
        )
        self.var_service.add_tags.assert_called_once_with("host", ["new"], profile=None)

    def test_handle_variables_skip_and_create_paths(self):
        self.service.handle_variables(
            [
                ({"name": "skip", "value": "x"}, "skip"),
                ({"name": "create", "value": "y", "tags": ["prod"]}, "create"),
            ]
        )

        self.var_service.create_variable.assert_called_once_with(
            "create", value="y", tags=["prod"], profile=None
        )

    @patch("cmdbox.services.import_service.validate_no_cycles")
    @patch("cmdbox.services.import_service.build_dependency_graph")
    @patch("cmdbox.services.import_service._parse_import_file")
    def test_import_file_with_profile(
        self, mock_parse, mock_build_dependency_graph, mock_validate_no_cycles
    ):
        profile = "test-profile"
        mock_parse.return_value = {
            "version": "1",
            "commands": [{"alias": "cmd1", "template": "echo 1"}],
            "variables": [{"name": "var1", "value": "1"}],
        }
        self.cmd_service.get_command_or_none.return_value = None
        self.var_service.get_variable_or_none.return_value = None

        self.service.import_file("path", profile=profile)

        self.cmd_service.get_command_or_none.assert_called_with("cmd1", profile=profile)
        self.var_service.get_variable_or_none.assert_called_with(
            "var1", profile=profile
        )
        self.cmd_service.create_command.assert_called_with(
            alias="cmd1",
            template="echo 1",
            description=None,
            tags=[],
            cwd=None,
            shell=None,
            env=None,
            timeout=None,
            profile=profile,
        )
        self.var_service.create_variable.assert_called_with(
            "var1", value="1", tags=[], profile=profile
        )
