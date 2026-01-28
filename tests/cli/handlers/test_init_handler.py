import unittest
from unittest.mock import patch, MagicMock
from pathlib import Path
import typer
import os

from cmdbox.cli.handlers import init_handler
from cmdbox.init.specs import ShellSpec
from cmdbox.init import io, detect, specs


class TestInitHandler(unittest.TestCase):

    @patch("cmdbox.init.io.resources.files")
    def test_load_integration_text(self, mock_files):
        # Setup
        mock_resource = MagicMock()
        mock_resource.joinpath.return_value.read_text.return_value = "snippet-content\n"
        mock_files.return_value = mock_resource

        # Execute
        result = io.load_integration_text("test.sh")

        # Verify
        self.assertEqual(result, "snippet-content\n")
        mock_resource.joinpath.assert_called_once_with("test.sh")

    @patch("cmdbox.init.io.Path.mkdir")
    @patch("cmdbox.init.io.Path.exists")
    @patch("cmdbox.init.io.Path.read_text")
    @patch("cmdbox.init.io.Path.write_text")
    @patch("cmdbox.init.io.shutil.copy2")
    def test_upsert_marked_block_new_file(
        self, mock_copy, mock_write, mock_read, mock_exists, mock_mkdir
    ):
        # Setup
        mock_exists.return_value = False
        path = Path("/mock/path/.bashrc")
        block_text = "some-integration"

        # Execute
        io.upsert_marked_block(path, block_text)

        # Verify
        mock_mkdir.assert_called_once_with(parents=True, exist_ok=True)
        mock_write.assert_called_once()
        written_content = mock_write.call_args[0][0]
        self.assertIn(io.START_MARK, written_content)
        self.assertIn(block_text, written_content)
        self.assertIn(io.END_MARK, written_content)
        mock_copy.assert_not_called()

    @patch("cmdbox.init.io.Path.mkdir")
    @patch("cmdbox.init.io.Path.exists")
    @patch("cmdbox.init.io.Path.read_text")
    @patch("cmdbox.init.io.Path.write_text")
    @patch("cmdbox.init.io.shutil.copy2")
    def test_upsert_marked_block_update_existing(
        self, mock_copy, mock_write, mock_read, mock_exists, mock_mkdir
    ):
        # Setup
        mock_exists.return_value = True
        existing_content = (
            f"some-stuff\n{io.START_MARK}\nold-content\n{io.END_MARK}\nmore-stuff"
        )
        mock_read.return_value = existing_content
        path = Path("/mock/path/.bashrc")
        new_block = "new-content"

        # Execute
        io.upsert_marked_block(path, new_block)

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
            "cmdbox.init.specs.Path.home",
            return_value=Path("/home/user"),
        ):
            self.assertEqual(specs.default_bashrc(), Path("/home/user/.bashrc"))
            self.assertEqual(specs.default_zshrc(), Path("/home/user/.zshrc"))
            self.assertEqual(
                specs.default_fish_function(),
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
        self.assertEqual(specs.default_powershell_profile(), expected)

    @patch.dict(os.environ, {}, clear=True)
    @patch("cmdbox.init.specs.Path.home", return_value=Path("/home/user"))
    def test_default_powershell_profile_fallback(self, mock_home):
        expected = (
            Path("/home/user")
            / "Documents"
            / "PowerShell"
            / "Microsoft.PowerShell_profile.ps1"
        )
        self.assertEqual(specs.default_powershell_profile(), expected)

    @patch("cmdbox.cli.handlers.init_handler.render_install_instructions")
    @patch("cmdbox.cli.handlers.init_handler.detect_shell")
    @patch("cmdbox.cli.handlers.init_handler.load_integration_text")
    def test_run_init_command_no_install(self, mock_load, mock_detect, mock_render):
        # Setup
        mock_detect.return_value = "bash"
        mock_load.return_value = "snippet-code"
        mock_render.return_value = "rendered_instructions"
        mock_console = MagicMock()
        get_console = lambda: mock_console

        # Execute
        init_handler.run_init_command(
            shell=None, install=False, get_console=get_console
        )

        # Verify
        mock_render.assert_called_once()
        mock_console.print.assert_called_with("rendered_instructions")

    @patch("cmdbox.cli.handlers.init_handler.render_install_success")
    @patch("cmdbox.cli.handlers.init_handler.upsert_marked_block")
    @patch("cmdbox.cli.handlers.init_handler.load_integration_text")
    def test_run_init_command_install_profile_block(
        self, mock_load, mock_upsert, mock_render
    ):
        # Setup
        mock_load.return_value = "snippet-code"
        mock_render.return_value = "rendered_success"
        mock_console = MagicMock()
        get_console = lambda: mock_console

        # We need to mock SHELLS or at least make sure it uses our mock for default_path_fn
        with patch.dict(
            specs.SHELLS,
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
            mock_render.assert_called_once()
            mock_console.print.assert_called_once_with("rendered_success")

    @patch("cmdbox.cli.handlers.init_handler.render_install_success")
    @patch("cmdbox.cli.handlers.init_handler.Path.write_text")
    @patch("cmdbox.cli.handlers.init_handler.Path.mkdir")
    @patch("cmdbox.cli.handlers.init_handler.Path.exists")
    @patch("cmdbox.cli.handlers.init_handler.load_integration_text")
    def test_run_init_command_install_write_file(
        self, mock_load, mock_exists, mock_mkdir, mock_write, mock_render
    ):
        # Setup
        mock_load.return_value = "snippet-code"
        mock_exists.return_value = False
        mock_render.return_value = "rendered_success"
        mock_console = MagicMock()
        get_console = lambda: mock_console

        with patch.dict(
            specs.SHELLS,
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
            mock_render.assert_called_once()
            mock_console.print.assert_called_once_with("rendered_success")

    @patch("cmdbox.cli.handlers.init_handler.render_install_instructions")
    def test_run_init_command_wrapper_hint(self, mock_render):
        # Setup
        mock_console = MagicMock()
        get_console = lambda: mock_console
        mock_render.return_value = "rendered_hint"

        with patch(
            "cmdbox.cli.handlers.init_handler.load_integration_text",
            return_value="snippet-code",
        ):
            with patch.dict(
                specs.SHELLS,
                {"cmd": ShellSpec("cmd", "cmd.bat", None, "wrapper_hint")},
            ):
                # Execute
                init_handler.run_init_command(
                    shell="cmd", install=True, get_console=get_console
                )

                # Verify
                mock_render.assert_called_once()
                mock_console.print.assert_called_with("rendered_hint")

    def test_run_init_command_invalid_shell(self):
        mock_console = MagicMock()
        get_console = lambda: mock_console
        with self.assertRaises(typer.BadParameter):
            init_handler.run_init_command(
                shell="invalid-shell", install=False, get_console=get_console
            )

    @patch("cmdbox.cli.handlers.init_handler.render_shell_output")
    @patch("cmdbox.cli.handlers.init_handler.detect_shell")
    def test_run_detect_shell(self, mock_detect, mock_render):
        mock_detect.return_value = "zsh"
        mock_render.return_value = "rendered_shell"
        mock_console = MagicMock()
        get_console = lambda: mock_console

        init_handler.run_detect_shell(get_console=get_console)

        mock_render.assert_called_with("zsh")
        mock_console.print.assert_called_with("rendered_shell")

    @patch("cmdbox.init.detect.psutil.Process")
    @patch("cmdbox.init.detect.os.getppid")
    def test_detect_shell_process_windows_pwsh(self, mock_getppid, mock_process):
        mock_getppid.return_value = 123
        mock_parent = MagicMock()
        mock_parent.name.return_value = "pwsh.exe"
        mock_process.return_value.parent.return_value = mock_parent

        self.assertEqual(detect.detect_shell(), "powershell")

    @patch("cmdbox.init.detect.psutil.Process")
    @patch("cmdbox.init.detect.os.getppid")
    @patch.dict(os.environ, {"SHELL": "/usr/bin/zsh"}, clear=True)
    def test_detect_shell_fallback_env(self, mock_getppid, mock_process):
        mock_getppid.return_value = 123
        mock_parent = MagicMock()
        mock_parent.name.return_value = "unknown"
        mock_process.return_value.parent.return_value = mock_parent

        self.assertEqual(detect.detect_shell(), "zsh")

    @patch("cmdbox.init.detect.sys.platform", "win32")
    @patch.dict(os.environ, {"PSModulePath": "some-path"}, clear=True)
    def test_detect_shell_env_windows_ps(self):
        self.assertEqual(detect.detect_shell_env(), "powershell")

    @patch("cmdbox.init.detect.sys.platform", "linux")
    @patch.dict(os.environ, {"SHELL": "/bin/fish"}, clear=True)
    def test_detect_shell_env_linux_fish(self):
        self.assertEqual(detect.detect_shell_env(), "fish")

    @patch("cmdbox.init.detect.sys.platform", "linux")
    @patch.dict(os.environ, {}, clear=True)
    def test_detect_shell_env_default(self):
        self.assertEqual(detect.detect_shell_env(), "bash")
