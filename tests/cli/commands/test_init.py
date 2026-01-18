import unittest
from unittest.mock import patch
from cmdbox.cli.commands import init


class TestInitCommand(unittest.TestCase):

    @patch("cmdbox.cli.commands.init.run_init_command")
    @patch("cmdbox.cli.commands.init.container")
    def test_init_command(self, mock_container, mock_run_init):
        # Setup
        shell = "bash"
        install = True
        path = "/tmp/bashrc"

        # Execute
        init.init(shell=shell, install=install, path=path)

        # Verify
        mock_run_init.assert_called_once_with(
            shell=shell,
            install=install,
            path=path,
            get_console=mock_container.get_console,
        )

    @patch("cmdbox.cli.commands.init.run_detect_shell")
    @patch("cmdbox.cli.commands.init.container")
    def test_detect_shell_command(self, mock_container, mock_run_detect):
        # Execute
        init.detect_shell()

        # Verify
        mock_run_detect.assert_called_once_with(get_console=mock_container.get_console)
