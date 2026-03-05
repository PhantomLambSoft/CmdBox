import logging
import os
import sys
import subprocess
import tempfile
from dataclasses import dataclass
from typing import Mapping

import typer

from cmdbox.runtime.results import ExecutionResult
from cmdbox.runtime.shell import build_shell_command
from cmdbox.logging_setup.log_decorators import log_action


log = logging.getLogger(__name__)


@dataclass(frozen=True)
class RunContext:
    """
    Represents the execution context for a runnable process.

    This class defines the environment and settings under which a process is
    executed. It includes attributes such as the current working directory,
    environment variables, and capture settings. Being a frozen dataclass, the
    RunContext instances are immutable.

    Attributes:
        cwd (str | None): The current working directory for the process. If None,
            the process will inherit the working directory from the parent process.
        env (Mapping[str, str] | None): Environment variables to set for the
            process. If None, the process will inherit the environment from the
            parent process.
        capture (bool): Whether the process's output streams should be captured.
            If False, the output streams will inherit those of the parent process.
        shell (str | None): The shell to use for executing the command. If None,
            the system default shell will be used.
        emit (bool): Whether to emit the command template.  If True, the command
            template is emitted to the current terminal window to be evaluated
            in the current session.  If False, the command is executed in a
            different session using a subprocess.  Defaults to False.
        verbose (bool): Whether to output additional information alongside the
            command output. Defaults to False.
    """

    cwd: str | None = None
    env: Mapping[str, str] | None = None
    capture: bool = False
    shell: str | None = None
    emit: bool = False
    verbose: bool = False


class Executor:

    @log_action(__name__, "run_executor_run")
    def run(self, command: str, ctx: RunContext = RunContext()) -> ExecutionResult:
        """
        Executes a shell command in a subprocess, capturing the output and exit code.

        This method takes a shell command as a string and runs it in a subprocess. It
        allows the caller to specify the working directory, environment variables, and
        whether or not to capture the output through the provided context. The result
        of the execution is returned encapsulated in an `ExecutionResult`.

        Args:
            command (str): The shell command to be executed.
            ctx (RunContext, optional): An instance of `RunContext` that provides
                additional execution context such as the working directory,
                environment variables, and capture preferences. Defaults to a new
                `RunContext()` instance.

        Returns:
            ExecutionResult: An object containing the executed command, the exit code,
                and the captured standard output and error streams.
        """
        log.info(
            "Executing command: mode=%s, multiline=%s, capture=%s, shell=%s, cmd_len=%s, cwd_set=%s, env_override=%s",
            "emit" if ctx.emit else "subprocess",
            self.is_multiline(command),
            ctx.capture,
            ctx.shell,
            len(command),
            ctx.cwd is not None,
            ctx.env is not None,
        )

        if ctx.emit:
            self.emit_command(command)
            return None  # Just a safeguard, this should not return if emit is True

        env = os.environ.copy()
        if ctx.env:
            env.update(dict(ctx.env))

        if self.is_multiline(command):
            return self.run_multiline_as_script(command, ctx=ctx, env=env)

        popen_args = build_shell_command(command, preferred_shell=ctx.shell)
        completed = subprocess.run(
            popen_args,
            cwd=ctx.cwd,
            text=True,
            env=env,
            capture_output=ctx.capture,
        )

        log.info("Command completed with exit code: %s", completed.returncode)

        return ExecutionResult(
            command=command,
            exit_code=completed.returncode,
            stdout=completed.stdout or "",
            stderr=completed.stderr or "",
        )

    @staticmethod
    def emit_command(command: str) -> None:
        """
        Emits a formatted command to standard output and terminates the process
        with a success exit code.

        This method takes a string command as input, appends a newline character,
        and writes it to the standard output. It ensures the command is formatted
        with a single trailing newline before being emitted. Once executed, the
        method forcefully exits the process with an exit code of 0.

        Args:
            command (str): The input command that needs to be written to standard
                output. It should be a valid string representation of the command
                to execute.
        """
        cmd = command.strip("\n") + "\n"
        sys.stdout.write(cmd)
        raise typer.Exit(code=0)

    @staticmethod
    def is_multiline(command: str) -> bool:
        """
        Determines if the given command is multiline. A command is considered multiline
        if it contains at least one newline character after stripping leading and trailing
        newlines.

        Args:
            command (str): The command string to check.

        Returns:
            bool: True if the command contains at least one newline character after
            trimming leading and trailing newline characters, False otherwise.
        """
        return "\n" in command.strip("\n")

    @log_action(__name__, "run_executor_run_multiline_as_script")
    def run_multiline_as_script(
        self, command: str, ctx: RunContext, env: dict[str, str]
    ) -> ExecutionResult:
        """
        Executes a multiline command as a script in the context of a specified shell environment.

        This method creates a temporary script file containing the provided multiline command and
        executes it using subprocess. The command is normalized for consistent behavior across
        platforms, and an optional shell-specific header is prepended based on the execution context.

        Args:
            command (str): The multiline command to execute as a script.
            ctx (RunContext): Context about the execution, including shell type, working directory,
                and whether to capture output.
            env (dict[str, str]): A dictionary of environment variables to set during script execution.

        Returns:
            ExecutionResult: An object containing the executed command, its exit code, and captured
            standard output and error streams.
        """
        shell = (ctx.shell or "").lower()
        suffix = self.script_suffix_for_shell(shell)
        script_path = None

        if shell == "default":
            log.debug("exec multiline as default shell, suffix=%s", suffix)
        else:
            log.debug("exec multiline as shell=%s, suffix=%s", shell, suffix)

        # Normalize newlines so contents is consistent across platforms
        script_body = command.replace("\r\n", "\n").rstrip("\n") + "\n"

        try:
            with tempfile.NamedTemporaryFile(
                mode="w",
                encoding="utf-8",
                delete=False,
                suffix=suffix,
            ) as file:
                script_path = file.name
                header = self.script_header_for_shell(shell)
                if header:
                    file.write(header)
                file.write(script_body)

            popen_args = self.build_script_exec_args(script_path, shell=shell)
            log.debug(
                "exec multiline args_count=%s header=%s", len(popen_args), bool(header)
            )

            completed = subprocess.run(
                popen_args,
                cwd=ctx.cwd,
                text=True,
                env=env,
                capture_output=ctx.capture,
            )

            log.info("Command completed with exit code: %s", completed.returncode)

            return ExecutionResult(
                command=command,
                exit_code=completed.returncode,
                stdout=completed.stdout or "",
                stderr=completed.stderr or "",
            )
        finally:
            if script_path:
                try:
                    os.remove(script_path)
                except OSError:
                    pass

    @staticmethod
    def script_suffix_for_shell(shell: str) -> str:
        """
        Determines the appropriate script file suffix for a given shell.

        This function inspects the input shell name or its characteristics and returns
        the corresponding script file suffix based on the shell's type.

        Args:
            shell (str): The name of the shell or its identifier.

        Returns:
            str: The appropriate script file suffix for the given shell.
        """
        if "cmd" in shell or shell == "cmd.exe":
            return ".cmd"
        if "powershell" in shell or shell == "pwsh":
            return ".ps1"
        if "fish" in shell:
            return ".fish"
        # Default to bash
        return ".sh"

    @staticmethod
    def script_header_for_shell(shell: str) -> str:
        """
        Generates a script header for the given shell type.

        This method determines the appropriate script header based on the
        specified shell string. Supported shells include bash, zsh, and fish.
        If the shell type isn't recognized or no header is required, it returns
        an empty string.

        Args:
            shell (str): The name of the shell for which the script header is
                to be generated.

        Returns:
            str: The script header string corresponding to the provided shell,
                or an empty string if no header is required.
        """
        if "bash" in shell:
            return "#!/usr/bin/env bash\n"
        if "zsh" in shell:
            return "#!/usr/bin/env zsh\n"
        if "fish" in shell:
            return "#!/usr/bin/env fish\n"
        # Other types do not require a header
        return ""

    @staticmethod
    def build_script_exec_args(script_path: str, shell: str) -> list[str]:
        """
        Assembles a list of arguments to execute a given script based on the specified shell.

        This method generates the appropriate command and arguments to execute
        a script, tailored to the shell type provided. It ensures compatibility with
        various shell environments such as cmd, PowerShell, bash, zsh, and fish.

        Args:
            script_path (str): The file path to the script that needs to be executed.
            shell (str): The name or identifier of the shell environment used for script execution.

        Returns:
            list[str]: A list of arguments that, when executed, run the specified script in the given shell.
        """
        if "cmd" in shell or shell == "cmd.exe":
            return ["cmd.exe", "/d", "/s", "/c", script_path]

        if "powershell" in shell or shell == "pwsh":
            exe = "pwsh" if "pwsh" in shell else "powershell"
            return [
                exe,
                "-NoProfile",
                "-NonInteractive",
                "-ExecutionPolicy",
                "Bypass",
                "-File",
                script_path,
            ]

        if "fish" in shell:
            return ["fish", script_path]

        if "zsh" in shell:
            return ["zsh", script_path]

        if "bash" in shell:
            return ["bash", script_path]

        return ["sh", script_path]
