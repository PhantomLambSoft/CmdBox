from pathlib import Path
from tomlkit import dumps, parse, document

from cmdbox.common.io import atomic_write_text


class SettingsRepository:
    """
    Handles loading and saving settings to a file.

    Provides functionality to persist and retrieve application settings from a file
    on the filesystem. The settings are managed as a dictionary and stored in a
    structured format.

    Attributes:
        path (Path): Path to the file where settings are stored.
    """

    def __init__(self, path: Path):
        self.path = path

    def load(self) -> dict:
        if not self.path.exists():
            return {}
        doc = parse(self.path.read_text(encoding="utf-8"))
        return doc.unwrap()

    def save(self, data: dict) -> None:
        doc = document()
        doc.update(data)
        atomic_write_text(self.path, dumps(doc), encoding="utf-8")

    def dict_to_text(self, data: dict) -> str:
        doc = document()
        doc.update(data)
        return dumps(doc)
