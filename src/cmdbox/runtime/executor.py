import os
import subprocess
from dataclasses import dataclass
from typing import Mapping

from cmdbox.runtime.results import ExecutionResult
from cmdbox.runtime.shell import build_shell_command


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
    """

    cwd: str | None = None
    env: Mapping[str, str] | None = None
    capture: bool = False


class Executor:

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
        popen_args = build_shell_command(command)

        env = os.environ.copy()
        if ctx.env:
            env.update(dict(ctx.env))

        completed = subprocess.run(
            popen_args,
            cwd=ctx.cwd,
            text=True,
            env=env,
            capture_output=ctx.capture,
        )
        return ExecutionResult(
            command=command,
            exit_code=completed.returncode,
            stdout=completed.stdout or "",
            stderr=completed.stderr or "",
        )
