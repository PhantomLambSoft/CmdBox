import json
import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import MagicMock, patch

from cmdbox.services.export_service import (
    ExportService,
    _build_document,
    _resolve_output_path,
    _serialize_command,
    _serialize_variable,
    collect_deep_commands,
    collect_deep_variables,
    flatten_template,
)


def make_command(
    alias,
    template,
    description=None,
    tags=None,
    cwd=None,
    shell=None,
    env=None,
    timeout=None,
):
    return SimpleNamespace(
        alias=alias,
        template=template,
        description=description,
        tags=[SimpleNamespace(name=t) for t in (tags or [])],
        cwd=cwd,
        shell=shell,
        env=env,
        timeout=timeout,
    )


def make_variable(name, value, tags=None):
    return SimpleNamespace(
        name=name,
        value=value,
        tags=[SimpleNamespace(name=t) for t in (tags or [])],
    )


class TestExportHelpers(unittest.TestCase):

    def test_collect_deep_commands_follows_nested_references_and_skips_missing(self):
        cmd_service = MagicMock()
        commands = {
            "deploy": make_command("deploy", "run <cmd:build>"),
            "build": make_command("build", "echo done <cmd:missing>"),
        }
        cmd_service.get_command_or_none.side_effect = lambda alias: commands.get(alias)

        result = collect_deep_commands(["deploy", "unknown"], cmd_service)

        self.assertEqual({"deploy", "build"}, set(result.keys()))

    def test_collect_deep_variables_reads_initial_names_and_command_references(self):
        var_service = MagicMock()
        commands = {
            "deploy": make_command("deploy", "x=<var:host> <var:password>"),
        }
        variables = {
            "host": make_variable("host", "<var:region>"),
            "region": make_variable("region", "us-east"),
            "user": make_variable("user", "admin"),
        }
        var_service.get_variable_or_none.side_effect = lambda name: variables.get(name)

        result = collect_deep_variables(["user"], commands, var_service)

        self.assertEqual({"host", "region", "user"}, set(result.keys()))

    def test_flatten_template_handles_escape_unknown_and_recursion(self):
        cmd_service = MagicMock()
        var_service = MagicMock()
        commands = {
            "loop": make_command("loop", "start <cmd:loop>"),
            "deploy": make_command("deploy", "to <var:host>"),
        }
        variables = {
            "host": make_variable("host", "api-<var:host>"),
        }
        cmd_service.get_command_or_none.side_effect = lambda alias: commands.get(alias)
        var_service.get_variable_or_none.side_effect = lambda name: variables.get(name)

        result = flatten_template(
            r"A \<var:x> B <cmd:deploy> C <cmd:missing> D <cmd:loop> E <broken",
            cmd_service,
            var_service,
        )

        self.assertEqual(
            r"A \<var:x> B to api-<var:host> C <cmd:missing> D start <cmd:loop> E <broken",
            result,
        )

    @patch("cmdbox.services.export_service.datetime")
    def test_build_document_has_expected_shape(self, mock_datetime):
        fixed_iso = "2024-01-01T00:00:00+00:00"
        now_obj = MagicMock()
        now_obj.isoformat.return_value = fixed_iso
        mock_datetime.now.return_value = now_obj

        result = _build_document("cmds", [{"alias": "a"}], [{"name": "v"}])

        self.assertEqual("1", result["version"])
        self.assertEqual("cmds", result["type"])
        self.assertEqual(fixed_iso, result["exported_at"])
        self.assertEqual([{"alias": "a"}], result["commands"])
        self.assertEqual([{"name": "v"}], result["variables"])

    @patch("cmdbox.services.export_service.datetime")
    @patch("cmdbox.services.export_service.Path.is_dir", return_value=True)
    def test_resolve_output_path_uses_cwd_or_directory_or_file(
        self, _mock_is_dir, mock_datetime
    ):
        mock_datetime.now.return_value.strftime.return_value = "2025-02-03"

        with patch(
            "cmdbox.services.export_service.Path.cwd", return_value=Path("C:\\work")
        ):
            default_path = _resolve_output_path(None, "cmds")

        directory_path = _resolve_output_path("C:\\exports", "cmds")

        with patch("cmdbox.services.export_service.Path.is_dir", return_value=False):
            file_path = _resolve_output_path("C:\\exports\\out.json", "cmds")

        self.assertEqual(Path("C:\\work\\cmdbox-cmds-2025-02-03.json"), default_path)
        self.assertEqual(
            Path("C:\\exports\\cmdbox-cmds-2025-02-03.json"), directory_path
        )
        self.assertEqual(Path("C:\\exports\\out.json"), file_path)

    def test_serialize_command_and_variable_handle_flatten_and_env(self):
        cmd = make_command(
            "deploy",
            "<var:host>",
            description="d",
            tags=["ops"],
            cwd="C:/repo",
            shell="pwsh",
            env='{"A": "B"}',
            timeout=30,
        )
        var = make_variable("host", "example", tags=["prod"])

        with patch(
            "cmdbox.services.export_service.flatten_template", return_value="flattened"
        ):
            serialized_cmd = _serialize_command(cmd, True, MagicMock(), MagicMock())
            serialized_var = _serialize_variable(var, True, MagicMock(), MagicMock())

        self.assertEqual("deploy", serialized_cmd["alias"])
        self.assertEqual("flattened", serialized_cmd["template"])
        self.assertEqual({"A": "B"}, serialized_cmd["env"])
        self.assertEqual(["ops"], serialized_cmd["tags"])
        self.assertEqual("host", serialized_var["name"])
        self.assertEqual("flattened", serialized_var["value"])
        self.assertEqual(["prod"], serialized_var["tags"])


class TestExportService(unittest.TestCase):
    def setUp(self):
        self.cmd_service = MagicMock()
        self.var_service = MagicMock()
        self.service = ExportService(self.cmd_service, self.var_service)

    @patch("cmdbox.services.export_service.atomic_write_text")
    @patch("cmdbox.services.export_service._resolve_output_path")
    def test_export_cmds_non_flatten_collects_deep_dependencies_and_warnings(
        self, mock_resolve_output_path, mock_atomic_write_text
    ):
        mock_resolve_output_path.return_value = Path("C:\\exports\\cmds.json")
        deploy = make_command("deploy", "run <cmd:build> <var:env>", tags=["release"])
        build = make_command("build", "echo <var:build_flag>")
        env = make_variable("env", "prod")
        build_flag = make_variable("build_flag", "-v")
        self.cmd_service.get_command_or_none.side_effect = lambda alias: {
            "deploy": deploy,
            "build": build,
        }.get(alias)
        self.var_service.get_variable_or_none.side_effect = lambda name: {
            "env": env,
            "build_flag": build_flag,
        }.get(name)

        result = self.service.export_cmds(["deploy", "missing"], flatten=False)

        self.assertEqual(["Command missing not found"], result.warnings)
        self.assertEqual(["deploy"], result.commands)
        self.assertEqual(["build"], result.transient_commands)
        self.assertEqual([], result.variables)
        self.assertEqual({"env", "build_flag"}, set(result.transient_variables))
        self.assertEqual(Path("C:\\exports\\cmds.json"), result.path)
        mock_atomic_write_text.assert_called_once()
        written_doc = json.loads(mock_atomic_write_text.call_args.args[1])
        self.assertEqual("cmds", written_doc["type"])
        self.assertEqual(2, len(written_doc["commands"]))
        self.assertEqual(2, len(written_doc["variables"]))

    @patch("cmdbox.services.export_service.atomic_write_text")
    @patch("cmdbox.services.export_service._resolve_output_path")
    def test_export_cmds_flatten_serializes_only_found_targets(
        self, mock_resolve_output_path, mock_atomic_write_text
    ):
        mock_resolve_output_path.return_value = Path("C:\\exports\\cmds-flat.json")
        self.cmd_service.get_command_or_none.side_effect = lambda alias: {
            "deploy": make_command("deploy", "<var:env>")
        }.get(alias)
        self.var_service.get_variable_or_none.side_effect = lambda name: {
            "env": make_variable("env", "prod")
        }.get(name)

        result = self.service.export_cmds(["deploy", "missing"], flatten=True)

        self.assertEqual(["deploy"], result.commands)
        self.assertEqual([], result.variables)
        self.assertEqual([], result.transient_commands)
        self.assertEqual([], result.transient_variables)
        self.assertEqual(["Command missing not found"], result.warnings)
        written_doc = json.loads(mock_atomic_write_text.call_args.args[1])
        self.assertEqual(1, len(written_doc["commands"]))
        self.assertEqual(0, len(written_doc["variables"]))

    @patch("cmdbox.services.export_service.atomic_write_text")
    def test_export_cmds_uses_tag_listing_when_aliases_are_none(
        self, mock_atomic_write_text
    ):
        self.cmd_service.list_commands.return_value = [
            make_command("by-tag", "echo hi"),
        ]
        self.cmd_service.get_command_or_none.side_effect = lambda alias: {
            "by-tag": make_command("by-tag", "echo hi")
        }.get(alias)

        result = self.service.export_cmds(None, tag="ops", flatten=True)

        self.assertEqual(["by-tag"], result.commands)
        self.cmd_service.list_commands.assert_called_once_with(tags="ops", limit=10000)
        mock_atomic_write_text.assert_called_once()

    @patch("cmdbox.services.export_service.atomic_write_text")
    def test_export_vars_flatten_uses_tag_listing_and_ignores_missing(
        self, mock_atomic_write_text
    ):
        self.var_service.list_variables.return_value = [
            make_variable("a", "1"),
            make_variable("missing", "?"),
        ]
        self.var_service.get_variable_or_none.side_effect = lambda name: {
            "a": make_variable("a", "1")
        }.get(name)

        result = self.service.export_vars(None, tag="prod", flatten=True)

        self.assertEqual(["a"], result.variables)
        self.assertEqual([], result.transient_commands)
        self.assertEqual([], result.transient_variables)
        self.assertEqual(["Variable missing not found"], result.warnings)
        self.var_service.list_variables.assert_called_once_with(
            tags="prod", limit=10000
        )
        mock_atomic_write_text.assert_called_once()

    @patch("cmdbox.services.export_service.atomic_write_text")
    def test_export_vars_uses_default_listing_when_no_names_or_tag(
        self, mock_atomic_write_text
    ):
        self.var_service.list_variables.return_value = [make_variable("x", "1")]
        self.var_service.get_variable_or_none.side_effect = lambda name: {
            "x": make_variable("x", "1")
        }.get(name)

        result = self.service.export_vars()

        self.assertEqual(["x"], result.variables)
        self.var_service.list_variables.assert_called_once_with(limit=10000)
        mock_atomic_write_text.assert_called_once()

    @patch("cmdbox.services.export_service.atomic_write_text")
    def test_export_all_serializes_everything_and_only_flattens_commands(
        self, mock_atomic_write_text
    ):
        all_cmds = [make_command("deploy", "<var:env>")]
        all_vars = [make_variable("env", "prod")]
        self.cmd_service.list_commands.return_value = all_cmds
        self.var_service.list_variables.return_value = all_vars
        self.var_service.get_variable_or_none.side_effect = lambda name: {
            "env": make_variable("env", "prod")
        }.get(name)

        result = self.service.export_all(flatten=True)

        self.assertEqual(["deploy"], result.commands)
        self.assertEqual(["env"], result.variables)
        written_doc = json.loads(mock_atomic_write_text.call_args.args[1])
        self.assertEqual("all", written_doc["type"])
        self.assertEqual("prod", written_doc["variables"][0]["value"])
        self.assertEqual("prod", written_doc["commands"][0]["template"])
