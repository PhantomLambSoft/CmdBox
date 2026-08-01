import json
import unittest
from unittest.mock import MagicMock, patch

from cmdbox.exceptions import CmdboxError
from cmdbox.runtime.executor import RunContext
from cmdbox.services.run_service import RunService
from cmdbox.models import Command
from cmdbox.resolve.type_defs import ResolveResult, TraceStep, RefKind
from cmdbox.runtime.results import ExecutionResult
from cmdbox.repositories.errors import UnknownAliasError
from cmdbox.resolve.errors import ResolutionError


class TestRunService(unittest.TestCase):

    def setUp(self):
        self.mock_repo = MagicMock()
        self.mock_resolver = MagicMock()
        self.mock_executor = MagicMock()
        self.mock_profile_repo = MagicMock()
        self.service = RunService(
            repo=self.mock_repo,
            resolver=self.mock_resolver,
            executor=self.mock_executor,
            profile_repo=self.mock_profile_repo,
        )

    @patch("cmdbox.services.run_service.RunService.build_context")
    def test_run_success(self, mock_build_context):
        mock_context = MagicMock()
        mock_build_context.return_value = mock_context

        # Setup
        alias = "test-alias"
        template = "echo <variable:name>"
        resolved_text = "echo world"

        command = MagicMock(spec=Command)
        command.template = template
        command.env = None
        self.mock_repo.get_by_alias.return_value = command

        resolve_result = MagicMock(spec=ResolveResult)
        resolve_result.text = resolved_text
        resolve_result.trace = [TraceStep(RefKind.VARIABLE, "name", "world", "stored")]
        self.mock_resolver.resolve.return_value = resolve_result

        execution_result = ExecutionResult(
            command=resolved_text, exit_code=0, stdout="world\n", stderr=""
        )
        self.mock_executor.run.return_value = execution_result

        # Execute
        result = self.service.run(alias)

        # Assert
        self.assertEqual(result, execution_result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_resolver.resolve.assert_called_once_with(template, runtime_vars=None)
        self.mock_executor.run.assert_called_once_with(resolved_text, ctx=mock_context)

    @patch("cmdbox.services.run_service.RunService.build_context")
    def test_record_use_updates_command_usage(self, mock_build_context):
        mock_context = MagicMock()
        mock_build_context.return_value = mock_context

        # Setup
        alias = "test-alias"
        template = "echo <variable:name>"
        resolved_text = "echo world"

        command = MagicMock(spec=Command)
        command.template = template
        command.env = None
        self.mock_repo.get_by_alias.return_value = command

        resolve_result = MagicMock(spec=ResolveResult)
        resolve_result.text = resolved_text
        resolve_result.trace = [TraceStep(RefKind.VARIABLE, "name", "world", "stored")]
        self.mock_resolver.resolve.return_value = resolve_result

        execution_result = ExecutionResult(
            command=resolved_text, exit_code=0, stdout="world\n", stderr=""
        )
        self.mock_executor.run.return_value = execution_result

        # Execute
        result = self.service.run(alias)

        # Assert
        self.mock_repo.record_use.assert_called_once()

    @patch("cmdbox.services.run_service.RunService.build_context")
    def test_record_use_updates_when_command_fails(self, mock_build_context):
        """
        Running a stored command and encountering a failed exit code should record
        a use. CmdBox's concern is that a command was run, not whether that command
        was successful or not.
        """
        alias = "test-alias"
        execution_result = ExecutionResult(
            command="echo test", exit_code=1, stdout="", stderr="failed"
        )
        self.mock_executor.run.return_value = execution_result

        self.service.run(alias)

        self.mock_repo.record_use.assert_called_once()

    @patch("cmdbox.services.run_service.RunService.build_context")
    def test_record_use_updates_when_executor_throws_exception(
        self, mock_build_context
    ):
        """
        If an exception occurs during the execution of the command, it's use
        should not be recorded as an issue with CmdBox kept the command from
        executing at all.
        """
        alias = "test-alias"
        self.mock_executor.run.side_effect = CmdboxError()

        with self.assertRaises(CmdboxError):
            self.service.run(alias)

        self.mock_repo.record_use.assert_not_called()

    def test_preview_success(self):
        # Setup
        alias = "test-alias"
        template = "echo <variable:name>"
        resolved_text = "echo world"

        command = MagicMock(spec=Command)
        command.template = template
        command.env = None
        self.mock_repo.get_by_alias.return_value = command

        resolve_result = MagicMock(spec=ResolveResult)
        resolve_result.text = resolved_text
        resolve_result.trace = []
        self.mock_resolver.resolve.return_value = resolve_result

        # Execute
        result, ctx = self.service.preview(alias)

        # Assert
        self.assertEqual(result, resolve_result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_resolver.resolve.assert_called_once_with(template, runtime_vars=None)
        self.mock_executor.run.assert_not_called()

    def test_run_command_not_found(self):
        # Setup
        alias = "non-existent"
        self.mock_repo.get_by_alias.side_effect = UnknownAliasError(alias)

        # Execute & Assert
        with self.assertRaises(UnknownAliasError) as context:
            self.service.run(alias)

        self.assertIn(f"Alias '{alias}' not found.", str(context.exception))
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_resolver.resolve.assert_not_called()
        self.mock_executor.run.assert_not_called()

    def test_preview_command_not_found(self):
        # Setup
        alias = "non-existent"
        self.mock_repo.get_by_alias.side_effect = UnknownAliasError(alias)

        # Execute & Assert
        with self.assertRaises(UnknownAliasError) as context:
            self.service.preview(alias)

        self.assertIn(f"Alias '{alias}' not found.", str(context.exception))
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_resolver.resolve.assert_not_called()

    def test_run_resolution_failure(self):
        # Setup
        alias = "test-alias"
        template = "echo <variable:circular>"

        command = MagicMock(spec=Command)
        command.template = template
        self.mock_repo.get_by_alias.return_value = command

        self.mock_resolver.resolve.side_effect = ResolutionError("Cycle detected")

        # Execute & Assert
        with self.assertRaises(ResolutionError) as context:
            self.service.run(alias)

        self.assertEqual(str(context.exception), "Cycle detected")
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_resolver.resolve.assert_called_once_with(template, runtime_vars=None)
        self.mock_executor.run.assert_not_called()

    def test_preview_resolution_failure(self):
        # Setup
        alias = "test-alias"
        template = "echo <variable:circular>"

        command = MagicMock(spec=Command)
        command.template = template
        self.mock_repo.get_by_alias.return_value = command

        self.mock_resolver.resolve.side_effect = ResolutionError("Cycle detected")

        # Execute & Assert
        with self.assertRaises(ResolutionError) as context:
            self.service.preview(alias)

        self.assertEqual(str(context.exception), "Cycle detected")
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_resolver.resolve.assert_called_once_with(template, runtime_vars=None)

    def test_collect_missing_vars_success(self):
        alias = "test-alias"
        template = "echo <name> <title>"
        expected_missing = ["name", "title"]

        command = MagicMock(spec=Command)
        command.template = template
        self.mock_repo.get_by_alias.return_value = command
        self.mock_resolver.collect_missing_vars.return_value = expected_missing

        result = self.service.collect_missing_vars(alias)

        self.assertEqual(expected_missing, result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_resolver.collect_missing_vars.assert_called_once_with(
            template, runtime_vars=None
        )

    def test_collect_missing_vars_with_runtime_vars(self):
        alias = "test-alias"
        template = "echo <name> <title>"
        runtime_vars = {"name": "Willi"}
        expected_missing = ["title"]

        command = MagicMock(spec=Command)
        command.template = template
        self.mock_repo.get_by_alias.return_value = command
        self.mock_resolver.collect_missing_vars.return_value = expected_missing

        result = self.service.collect_missing_vars(alias, runtime_vars=runtime_vars)

        self.assertEqual(expected_missing, result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_resolver.collect_missing_vars.assert_called_once_with(
            template, runtime_vars=runtime_vars
        )

    def test_collect_missing_vars_command_not_found(self):
        alias = "non-existent"
        self.mock_repo.get_by_alias.side_effect = UnknownAliasError(alias)

        with self.assertRaises(UnknownAliasError) as context:
            self.service.collect_missing_vars(alias)

        self.assertIn(f"Alias '{alias}' not found.", str(context.exception))
        self.mock_repo.get_by_alias.assert_called_once_with(alias, profile=None)
        self.mock_resolver.collect_missing_vars.assert_not_called()


class TestBuildContext(unittest.TestCase):

    def setUp(self):
        self.service = RunService(
            repo=MagicMock(),
            resolver=MagicMock(),
            executor=MagicMock(),
            profile_repo=MagicMock(),
        )

    def _make_command(
        self,
        cwd=None,
        shell=None,
        env=None,
        timeout=None,
    ) -> MagicMock:
        """Creates a mock Command with the given execution context fields.
        env should be passed as a dict; it is serialized to JSON to match
        how it is stored on the model.
        """
        cmd = MagicMock(spec=Command)
        cmd.cwd = cwd
        cmd.shell = shell
        cmd.env = json.dumps(env) if env is not None else None
        cmd.timeout = timeout
        return cmd

    # =========================================================================
    # SECTION: None/empty cases
    # =========================================================================

    def test_no_stored_fields_no_runtime_returns_none(self):
        cmd = self._make_command()
        result = self.service.build_context(cmd, None)
        self.assertIsNone(result)

    def test_no_stored_fields_default_runtime_context_returns_none(self):
        # RunContext() has all defaults (None/False), so nothing meaningful
        # to merge — should still return None
        cmd = self._make_command()
        result = self.service.build_context(cmd, RunContext())
        self.assertIsNone(result)

    # =========================================================================
    # SECTION: cwd
    # =========================================================================

    def test_stored_cwd_used_when_runtime_is_none(self):
        cmd = self._make_command(cwd="/stored/path")
        result = self.service.build_context(cmd, None)
        self.assertIsNotNone(result)
        self.assertEqual(result.cwd, "/stored/path")

    def test_stored_cwd_used_when_runtime_cwd_is_none(self):
        cmd = self._make_command(cwd="/stored/path")
        result = self.service.build_context(cmd, RunContext())
        self.assertEqual(result.cwd, "/stored/path")

    def test_runtime_cwd_overrides_stored_cwd(self):
        cmd = self._make_command(cwd="/stored/path")
        result = self.service.build_context(cmd, RunContext(cwd="/runtime/path"))
        self.assertEqual(result.cwd, "/runtime/path")

    def test_runtime_cwd_used_when_no_stored_cwd(self):
        cmd = self._make_command()
        result = self.service.build_context(cmd, RunContext(cwd="/runtime/path"))
        self.assertEqual(result.cwd, "/runtime/path")

    # =========================================================================
    # SECTION: shell
    # =========================================================================

    def test_stored_shell_used_when_runtime_is_none(self):
        cmd = self._make_command(shell="bash")
        result = self.service.build_context(cmd, None)
        self.assertIsNotNone(result)
        self.assertEqual(result.shell, "bash")

    def test_stored_shell_used_when_runtime_shell_is_none(self):
        cmd = self._make_command(shell="bash")
        result = self.service.build_context(cmd, RunContext())
        self.assertEqual(result.shell, "bash")

    def test_runtime_shell_overrides_stored_shell(self):
        cmd = self._make_command(shell="bash")
        result = self.service.build_context(cmd, RunContext(shell="zsh"))
        self.assertEqual(result.shell, "zsh")

    def test_runtime_shell_used_when_no_stored_shell(self):
        cmd = self._make_command()
        result = self.service.build_context(cmd, RunContext(shell="zsh"))
        self.assertEqual(result.shell, "zsh")

    # =========================================================================
    # SECTION: timeout
    # =========================================================================

    def test_stored_timeout_used_when_runtime_is_none(self):
        cmd = self._make_command(timeout=30)
        result = self.service.build_context(cmd, None)
        self.assertIsNotNone(result)
        self.assertEqual(result.timeout, 30)

    def test_stored_timeout_used_when_runtime_timeout_is_none(self):
        cmd = self._make_command(timeout=30)
        result = self.service.build_context(cmd, RunContext())
        self.assertEqual(result.timeout, 30)

    def test_runtime_timeout_overrides_stored_timeout(self):
        cmd = self._make_command(timeout=30)
        result = self.service.build_context(cmd, RunContext(timeout=10))
        self.assertEqual(result.timeout, 10)

    def test_runtime_timeout_used_when_no_stored_timeout(self):
        cmd = self._make_command()
        result = self.service.build_context(cmd, RunContext(timeout=10))
        self.assertEqual(result.timeout, 10)

    # =========================================================================
    # SECTION: env
    # =========================================================================

    def test_stored_env_used_when_runtime_is_none(self):
        cmd = self._make_command(env={"FOO": "bar"})
        result = self.service.build_context(cmd, None)
        self.assertIsNotNone(result)
        self.assertEqual(result.env, {"FOO": "bar"})

    def test_stored_env_used_when_runtime_env_is_none(self):
        cmd = self._make_command(env={"FOO": "bar"})
        result = self.service.build_context(cmd, RunContext())
        self.assertEqual(result.env, {"FOO": "bar"})

    def test_runtime_env_key_overrides_stored_key(self):
        cmd = self._make_command(env={"FOO": "stored_value"})
        result = self.service.build_context(
            cmd, RunContext(env={"FOO": "runtime_value"})
        )
        self.assertEqual(result.env["FOO"], "runtime_value")

    def test_runtime_env_merge_preserves_non_overridden_stored_keys(self):
        cmd = self._make_command(env={"FOO": "stored_foo", "BAR": "stored_bar"})
        result = self.service.build_context(cmd, RunContext(env={"FOO": "runtime_foo"}))
        self.assertEqual(result.env["FOO"], "runtime_foo")
        self.assertEqual(result.env["BAR"], "stored_bar")

    def test_runtime_env_only_used_when_no_stored_env(self):
        cmd = self._make_command()
        result = self.service.build_context(cmd, RunContext(env={"FOO": "bar"}))
        self.assertEqual(result.env, {"FOO": "bar"})

    def test_both_env_none_produces_no_env_on_context(self):
        cmd = self._make_command(cwd="/some/path")
        result = self.service.build_context(cmd, RunContext())
        self.assertIsNone(result.env)

    def test_multiple_stored_and_runtime_env_keys_merged_correctly(self):
        cmd = self._make_command(env={"A": "1", "B": "2", "C": "3"})
        result = self.service.build_context(
            cmd, RunContext(env={"B": "overridden", "D": "4"})
        )
        self.assertEqual(result.env["A"], "1")
        self.assertEqual(result.env["B"], "overridden")
        self.assertEqual(result.env["C"], "3")
        self.assertEqual(result.env["D"], "4")

    # =========================================================================
    # SECTION: behavioral flags
    # =========================================================================

    def test_capture_flag_taken_from_runtime(self):
        cmd = self._make_command(cwd="/some/path")
        result = self.service.build_context(cmd, RunContext(capture=True))
        self.assertTrue(result.capture)

    def test_emit_flag_taken_from_runtime(self):
        cmd = self._make_command(cwd="/some/path")
        result = self.service.build_context(cmd, RunContext(emit=True))
        self.assertTrue(result.emit)

    def test_verbose_flag_taken_from_runtime(self):
        cmd = self._make_command(cwd="/some/path")
        result = self.service.build_context(cmd, RunContext(verbose=True))
        self.assertTrue(result.verbose)

    def test_behavioral_flags_default_to_false_when_runtime_is_none(self):
        cmd = self._make_command(cwd="/some/path")
        result = self.service.build_context(cmd, None)
        self.assertFalse(result.capture)
        self.assertFalse(result.emit)
        self.assertFalse(result.verbose)

    def test_behavioral_flags_not_sourced_from_stored_command(self):
        # Behavioral flags only come from runtime; stored command has no
        # capture/emit/verbose fields. This verifies that even if runtime
        # has all flags False, the result is still a valid context when
        # stored fields are present.
        cmd = self._make_command(cwd="/some/path", shell="bash", timeout=30)
        result = self.service.build_context(cmd, RunContext())
        self.assertFalse(result.capture)
        self.assertFalse(result.emit)
        self.assertFalse(result.verbose)

    # =========================================================================
    # SECTION: all fields together
    # =========================================================================

    def test_all_fields_stored_and_no_runtime_uses_all_stored(self):
        cmd = self._make_command(
            cwd="/stored/path",
            shell="bash",
            env={"FOO": "bar"},
            timeout=30,
        )
        result = self.service.build_context(cmd, None)
        self.assertEqual(result.cwd, "/stored/path")
        self.assertEqual(result.shell, "bash")
        self.assertEqual(result.env, {"FOO": "bar"})
        self.assertEqual(result.timeout, 30)

    def test_all_fields_runtime_overrides_all_stored(self):
        cmd = self._make_command(
            cwd="/stored/path",
            shell="bash",
            env={"FOO": "stored"},
            timeout=30,
        )
        runtime = RunContext(
            cwd="/runtime/path",
            shell="zsh",
            env={"FOO": "runtime"},
            timeout=10,
        )
        result = self.service.build_context(cmd, runtime)
        self.assertEqual(result.cwd, "/runtime/path")
        self.assertEqual(result.shell, "zsh")
        self.assertEqual(result.env["FOO"], "runtime")
        self.assertEqual(result.timeout, 10)
