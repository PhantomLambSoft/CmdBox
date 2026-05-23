import unittest
from unittest.mock import MagicMock
from cmdbox.services.run_service import RunService
from cmdbox.models import Command
from cmdbox.resolve.type_defs import ResolveResult
from cmdbox.runtime.results import ExecutionResult
from cmdbox.repositories.errors import UnknownAliasError
from cmdbox.resolve.errors import ResolutionError


class TestRunService(unittest.TestCase):

    def setUp(self):
        self.mock_repo = MagicMock()
        self.mock_resolver = MagicMock()
        self.mock_executor = MagicMock()
        self.service = RunService(
            repo=self.mock_repo,
            resolver=self.mock_resolver,
            executor=self.mock_executor,
        )

    def test_run_success(self):
        # Setup
        alias = "test-alias"
        template = "echo <variable:name>"
        resolved_text = "echo world"

        command = MagicMock(spec=Command)
        command.template = template
        self.mock_repo.get_by_alias.return_value = command

        resolve_result = MagicMock(spec=ResolveResult)
        resolve_result.text = resolved_text
        self.mock_resolver.resolve.return_value = resolve_result

        execution_result = ExecutionResult(
            command=resolved_text, exit_code=0, stdout="world\n", stderr=""
        )
        self.mock_executor.run.return_value = execution_result

        # Execute
        result = self.service.run(alias)

        # Assert
        self.assertEqual(result, execution_result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
        self.mock_resolver.resolve.assert_called_once_with(template, runtime_vars=None)
        self.mock_executor.run.assert_called_once_with(resolved_text, ctx=None)

    def test_preview_success(self):
        # Setup
        alias = "test-alias"
        template = "echo <variable:name>"
        resolved_text = "echo world"

        command = MagicMock(spec=Command)
        command.template = template
        self.mock_repo.get_by_alias.return_value = command

        resolve_result = MagicMock(spec=ResolveResult)
        resolve_result.text = resolved_text
        resolve_result.trace = []
        self.mock_resolver.resolve.return_value = resolve_result

        # Execute
        result = self.service.preview(alias)

        # Assert
        self.assertEqual(result, resolve_result)
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
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
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
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
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
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
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
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
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
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
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
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
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
        self.mock_resolver.collect_missing_vars.assert_called_once_with(
            template, runtime_vars=runtime_vars
        )

    def test_collect_missing_vars_command_not_found(self):
        alias = "non-existent"
        self.mock_repo.get_by_alias.side_effect = UnknownAliasError(alias)

        with self.assertRaises(UnknownAliasError) as context:
            self.service.collect_missing_vars(alias)

        self.assertIn(f"Alias '{alias}' not found.", str(context.exception))
        self.mock_repo.get_by_alias.assert_called_once_with(alias)
        self.mock_resolver.collect_missing_vars.assert_not_called()
