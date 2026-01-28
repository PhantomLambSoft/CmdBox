import unittest
from unittest.mock import patch, MagicMock
from pathlib import Path
import typer
import os

from cmdbox.cli.handlers import init_handler
from cmdbox.init.specs import ShellSpec


class TestInitHandler(unittest.TestCase):

    @patch("cmdbox.cli.handlers.init_handler.resources.files")
    def test_load_integration_text(self, mock_files):
        # Setup
        mock_resource = MagicMock()
        mock_resource.joinpath.return_value.read_text.return_value = "snippet-content\n"
        mock_files.return_value = mock_resource

        # Execute
        result = init_handler.load_integration_text("test.sh")

        # Verify
        self.assertEqual(result, "snippet-content\n")
        mock_resource.joinpath.assert_called_once_with("test.sh")

    @patch("cmdbox.cli.handlers.init_handler.Path.mkdir")
    @patch("cmdbox.cli.handlers.init_handler.Path.exists")
    @patch("cmdbox.cli.handlers.init_handler.Path.read_text")
    @patch("cmdbox.cli.handlers.init_handler.Path.write_text")
    @patch("cmdbox.cli.handlers.init_handler.shutil.copy2")
    def test_upsert_marked_block_new_file(
        self, mock_copy, mock_write, mock_read, mock_exists, mock_mkdir
    ):
        # Setup
        mock_exists.return_value = False
        path = Path("/mock/path/.bashrc")
        block_text = "some-integration"

        # Execute
        init_handler.upsert_marked_block(path, block_text)

        # Verify
        mock_mkdir.assert_called_once_with(parents=True, exist_ok=True)
        mock_write.assert_called_once()
        written_content = mock_write.call_args[0][0]
        self.assertIn(init_handler.START_MARK, written_content)
        self.assertIn(block_text, written_content)
        self.assertIn(init_handler.END_MARK, written_content)
        mock_copy.assert_not_called()

    @patch("cmdbox.cli.handlers.init_handler.Path.mkdir")
    @patch("cmdbox.cli.handlers.init_handler.Path.exists")
    @patch("cmdbox.cli.handlers.init_handler.Path.read_text")
    @patch("cmdbox.cli.handlers.init_handler.Path.write_text")
    @patch("cmdbox.cli.handlers.init_handler.shutil.copy2")
    def test_upsert_marked_block_update_existing(
        self, mock_copy, mock_write, mock_read, mock_exists, mock_mkdir
    ):
        # Setup
        mock_exists.return_value = True
        existing_content = f"some-stuff\n{init_handler.START_MARK}\nold-content\n{init_handler.END_MARK}\nmore-stuff"
        mock_read.return_value = existing_content
        path = Path("/mock/path/.bashrc")
        new_block = "new-content"

        # Execute
        init_handler.upsert_marked_block(path, new_block)

        # Verify
        mock_copy.assert_called_once()
        mock_write.assert_called_once()
        written_content = mock_write.call_args[0][0]
        self.assertIn(new_block, written_content)
        self.assertNotIn("old-content", written_content)
        self.assertIn("some-stuff", written_content)
        self.assertIn("more-stuff", written_content)

    def test_default_paths(self):
        with patch(
            "cmdbox.cli.handlers.init_handler.Path.home",
            return_value=Path("/home/user"),
        ):
            self.assertEqual(init_handler.default_bashrc(), Path("/home/user/.bashrc"))
            self.assertEqual(init_handler.default_zshrc(), Path("/home/user/.zshrc"))
            self.assertEqual(
                init_handler.default_fish_function(),
                Path("/home/user/.config/fish/functions/cb.fish"),
            )

    @patch.dict(os.environ, {"USERPROFILE": "C:\\Users\\User"}, clear=True)
    def test_default_powershell_profile_windows(self):
        expected = (
            Path("C:\\Users\\User")
            / "Documents"
            / "PowerShell"
            / "Microsoft.PowerShell_profile.ps1"
        )
        self.assertEqual(init_handler.default_powershell_profile(), expected)

    @patch.dict(os.environ, {}, clear=True)
    @patch(
        "cmdbox.cli.handlers.init_handler.Path.home", return_value=Path("/home/user")
    )
    def test_default_powershell_profile_fallback(self, mock_home):
        expected = (
            Path("/home/user")
            / "Documents"
            / "PowerShell"
            / "Microsoft.PowerShell_profile.ps1"
        )
        self.assertEqual(init_handler.default_powershell_profile(), expected)

    @patch("cmdbox.cli.handlers.init_handler.detect_shell")
    @patch("cmdbox.cli.handlers.init_handler.load_integration_text")
    def test_run_init_command_no_install(self, mock_load, mock_detect):
        # Setup
        mock_detect.return_value = "bash"
        mock_load.return_value = "snippet-code"
        mock_console = MagicMock()
        get_console = lambda: mock_console

        # Execute
        init_handler.run_init_command(
            shell=None, install=False, get_console=get_console
        )

        # Verify
        mock_console.print.assert_called()
        self.assertIn("snippet-code", str(mock_console.print.call_args_list))

    @patch("cmdbox.cli.handlers.init_handler.upsert_marked_block")
    @patch("cmdbox.cli.handlers.init_handler.load_integration_text")
    def test_run_init_command_install_profile_block(self, mock_load, mock_upsert):
        # Setup
        mock_load.return_value = "snippet-code"
        mock_console = MagicMock()
        get_console = lambda: mock_console

        # We need to mock SHELLS or at least make sure it uses our mock for default_path_fn
        with patch.dict(
            init_handler.SHELLS,
            {
                "bash": ShellSpec(
                    "bash", "bash.sh", lambda: Path("/mock/.bashrc"), "profile_block"
                )
            },
        ):
            # Execute
            init_handler.run_init_command(
                shell="bash", install=True, get_console=get_console
            )

            # Verify
            mock_upsert.assert_called_once_with(Path("/mock/.bashrc"), "snippet-code")
            mock_console.success.assert_called_once()

    @patch("cmdbox.cli.handlers.init_handler.Path.write_text")
    @patch("cmdbox.cli.handlers.init_handler.Path.mkdir")
    @patch("cmdbox.cli.handlers.init_handler.Path.exists")
    @patch("cmdbox.cli.handlers.init_handler.load_integration_text")
    def test_run_init_command_install_write_file(
        self, mock_load, mock_exists, mock_mkdir, mock_write
    ):
        # Setup
        mock_load.return_value = "snippet-code"
        mock_exists.return_value = False
        mock_console = MagicMock()
        get_console = lambda: mock_console

        with patch.dict(
            init_handler.SHELLS,
            {
                "fish": ShellSpec(
                    "fish", "fish.fish", lambda: Path("/mock/cb.fish"), "write_file"
                )
            },
        ):
            # Execute
            init_handler.run_init_command(
                shell="fish", install=True, get_console=get_console
            )

            # Verify
            mock_write.assert_called_once_with("snippet-code", encoding="utf-8")
            mock_console.success.assert_called_once()

    def test_run_init_command_wrapper_hint(self):
        # Setup
        mock_console = MagicMock()
        get_console = lambda: mock_console

        with patch(
            "cmdbox.cli.handlers.init_handler.load_integration_text",
            return_value="snippet-code",
        ):
            with patch.dict(
                init_handler.SHELLS,
                {"cmd": ShellSpec("cmd", "cmd.bat", None, "wrapper_hint")},
            ):
                # Execute
                init_handler.run_init_command(
                    shell="cmd", install=True, get_console=get_console
                )

                # Verify
                # Should print the hint message and the snippet
                mock_console.print.assert_any_call("snippet-code", end="")
                # Check for part of the message
                self.assertTrue(
                    any(
                        "wrapper script" in str(args)
                        for args, kwargs in mock_console.print.call_args_list
                    )
                )

    def test_run_init_command_invalid_shell(self):
        mock_console = MagicMock()
        get_console = lambda: mock_console
        with self.assertRaises(typer.BadParameter):
            init_handler.run_init_command(
                shell="invalid-shell", install=False, get_console=get_console
            )

    @patch("cmdbox.cli.handlers.init_handler.detect_shell")
    def test_run_detect_shell(self, mock_detect):
        mock_detect.return_value = "zsh"
        mock_console = MagicMock()
        get_console = lambda: mock_console

        init_handler.run_detect_shell(get_console=get_console)

        mock_console.print.assert_called_with("Detected shell: zsh")

    @patch("psutil.Process")
    @patch("os.getppid")
    def test_detect_shell_process_windows_pwsh(self, mock_getppid, mock_process):
        mock_getppid.return_value = 123
        mock_parent = MagicMock()
        mock_parent.name.return_value = "pwsh.exe"
        mock_process.return_value.parent.return_value = mock_parent

        self.assertEqual(init_handler.detect_shell(), "powershell")

    @patch("psutil.Process")
    @patch("os.getppid")
    @patch.dict(os.environ, {"SHELL": "/usr/bin/zsh"}, clear=True)
    def test_detect_shell_fallback_env(self, mock_getppid, mock_process):
        mock_getppid.return_value = 123
        mock_parent = MagicMock()
        mock_parent.name.return_value = "unknown"
        mock_process.return_value.parent.return_value = mock_parent

        self.assertEqual(init_handler.detect_shell(), "zsh")

    @patch("sys.platform", "win32")
    @patch.dict(os.environ, {"PSModulePath": "some-path"}, clear=True)
    def test_detect_shell_env_windows_ps(self):
        self.assertEqual(init_handler.detect_shell_env(), "powershell")

    @patch("sys.platform", "linux")
    @patch.dict(os.environ, {"SHELL": "/bin/fish"}, clear=True)
    def test_detect_shell_env_linux_fish(self):
        self.assertEqual(init_handler.detect_shell_env(), "fish")

    @patch("sys.platform", "linux")
    @patch.dict(os.environ, {}, clear=True)
    def test_detect_shell_env_default(self):
        self.assertEqual(init_handler.detect_shell_env(), "bash")
