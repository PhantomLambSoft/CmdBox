import re
import shutil
from pathlib import Path
from importlib import resources


START_MARK = "# >>> cmdbox shell integration >>>"
END_MARK = "# <<< cmdbox shell integration <<<"


def load_integration_text(filename: str) -> str:
    """
    Loads a text file from the 'cmdbox.init.integrations' resource directory, reads its contents, and
    returns the text stripped of trailing whitespace with a newline character appended.

    Args:
        filename (str): The name of the file to be loaded and read.

    Returns:
        str: The contents of the specified file as a string, with trailing whitespace removed
             and a newline character appended at the end.
    """
    return (
        resources.files("cmdbox.init.integrations")
        .joinpath(filename)
        .read_text(encoding="utf-8")
        .rstrip()
        + "\n"
    )


def upsert_marked_block(profile_path: Path, block_text: str) -> None:
    """
    Updates or inserts a marked text block into a given file at the specified path.
    If the marked block already exists in the file, it is replaced with the provided
    new content. Otherwise, the block is added to the end of the file. Additionally,
    a backup of the original file is created if it already exists.

    Args:
        profile_path (Path): The path to the file where the marked block will be
            updated or inserted.
        block_text (str): The content to be added or updated as part of the marked
            block.

    """
    profile_path.parent.mkdir(parents=True, exist_ok=True)
    existing = profile_path.read_text(encoding="utf-8") if profile_path.exists() else ""

    marked = f"{START_MARK}\n{block_text.rstrip()}\n{END_MARK}\n"

    pattern = re.compile(
        re.escape(START_MARK) + r".*?" + re.escape(END_MARK) + r"\n?",
        flags=re.DOTALL,
    )

    if pattern.search(existing):
        new_text = pattern.sub(marked, existing)
    else:
        sep = "\n" if existing and not existing.endswith("\n") else ""
        new_text = existing + sep + marked

    if profile_path.exists():
        backup = profile_path.with_suffix(profile_path.suffix + ".bak")
        shutil.copy2(profile_path, backup)

    profile_path.write_text(new_text, encoding="utf-8")
