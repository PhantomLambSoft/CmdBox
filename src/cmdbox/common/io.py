import os
import tempfile
from pathlib import Path


def atomic_write_text(dest: Path, text: str, encoding: str = "utf-8") -> None:
    """
    Performs an atomic write operation to a text file. This ensures that the file content is written
    to a temporary file first and then atomically replaces the original file to prevent partial writes
    or data corruption.

    Args:
        dest (Path): The destination path for the final output file.
        text (str): The text content to write to the file.

    """
    dest.parent.mkdir(parents=True, exist_ok=True)
    with tempfile.NamedTemporaryFile(
        mode="w",
        encoding=encoding,
        dir=str(dest.parent),
        delete=False,
        prefix=dest.name + ".",
        suffix=".tmp",
        newline="\n",
    ) as tf:
        tmp_path = Path(tf.name)
        tf.write(text)
        tf.flush()
        os.fsync(tf.fileno())
    tmp_path.replace(dest)
