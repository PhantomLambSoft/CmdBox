import os
import shlex
import shutil
import subprocess
import tempfile
from pathlib import Path

from prompt_toolkit import Application
from prompt_toolkit.key_binding import KeyBindings
from prompt_toolkit.layout import Layout
from prompt_toolkit.styles import Style
from prompt_toolkit.widgets import TextArea, Frame


class EditCanceled(Exception):
    pass


def edit_text_fullscreen(initial_text: str, title: str = "Edit") -> str:
    """
    Opens a fullscreen text editor in the terminal for the user to edit the
    provided text.

    This function creates an interactive fullscreen text editor with options for
    saving or canceling edits. The user can press `Ctrl+S` to save changes,
    `Ctrl+Q` or `Esc` to cancel editing, and directly work with a text area that
    supports multiline editing, scrolling, and line numbers.

    Args:
        initial_text (str): The initial text content to be edited by the user.
        title (str, optional): The title of the editor window. Defaults to "Edit".

    Returns:
        str: The edited text after the user saves and exits the editor.
    """
    kb = KeyBindings()

    text_area = TextArea(
        text=initial_text,
        multiline=True,
        scrollbar=True,
        line_numbers=True,
        wrap_lines=False,
    )

    @kb.add("c-s")
    def _save(event):
        event.app.exit(result=text_area.text)

    @kb.add("c-q")
    @kb.add("escape")
    def _cancel(event):
        event.app.exit(exception=EditCanceled())

    root = Frame(text_area, title=f"{title} (Ctrl+S to save, Esc to cancel)")
    app = Application(
        layout=Layout(root),
        key_bindings=kb,
        full_screen=True,
        mouse_support=True,
        style=Style.from_dict({}),
    )

    return app.run()


def edit_text_in_editor(
    initial_text: str, suffix: str = ".txt", title_hint: str | None = None
) -> str:
    """
    Opens the given initial text in a temporary file for editing using a text editor
    resolved by the system. The final text, after editing, is returned as a string.

    Args:
        initial_text: The initial text content to populate the temporary file with.
        suffix: The file suffix/extension to use for the temporary file. Defaults to ".txt".
        title_hint: Optional hint for the editor window title.

    Returns:
        Edited content as a string after modifications in the editor.

    Raises:
        RuntimeError: If the editor cannot be found or launched properly.
    """
    editor_cmd = resolve_editor()

    with tempfile.TemporaryDirectory(prefix="cb_edit_") as td:
        path = Path(td) / f"edit{suffix}"
        path.write_text(initial_text, encoding="utf-8")
        cmd = [*editor_cmd, str(path)]

        env = os.environ.copy()
        if title_hint:
            env["CB_EDIT_TITLE"] = title_hint

        try:
            subprocess.run(cmd, check=False, env=env)
        except FileNotFoundError as e:
            raise RuntimeError(f"Unable to find editor: {editor_cmd}") from e

        return path.read_text(encoding="utf-8")


def resolve_editor() -> list[str]:
    """
    Determines the command used to open a text editor based on user environment
    or available system defaults.

    If an editor is specified in the `VISUAL` or `EDITOR` environment variables,
    this value is used. Otherwise, the function falls back to platform-specific
    default editors or searches for common command-line editors available on the
    system.

    Returns:
        list[str]: A list representing the command to open a text editor.
        For example, it may contain the editor's executable name or path, along
        with any required arguments.
    """
    editor = os.environ.get("VISUAL") or os.environ.get("EDITOR")
    if editor:
        return shlex.split(editor)

    if os.name == "nt":
        return ["notepad"]

    for candidate in ["nano", "vim", "vi"]:
        if shutil.which(candidate) is not None:
            return [candidate]

    return ["vi"]
