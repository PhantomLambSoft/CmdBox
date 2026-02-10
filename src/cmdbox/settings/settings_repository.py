from pathlib import Path
from tomlkit import dumps, parse, document


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
        self.path.parent.mkdir(parents=True, exist_ok=True)
        doc = document()
        doc.update(data)
        self.path.write_text(dumps(doc), encoding="utf-8")
