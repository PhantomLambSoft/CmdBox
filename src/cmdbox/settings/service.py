from dataclasses import asdict, is_dataclass, fields

from cmdbox.settings.models import Settings
from cmdbox.settings.repository import SettingsRepository


def build_dataclass(cls, data: dict):
    """
    Builds an instance of the given dataclass type populated with the provided
    data dictionary. This function recursively handles nested dataclasses,
    ensuring that each property is assigned correctly, even if the value
    is itself a dictionary representing another dataclass.

    Args:
        cls: The dataclass type to instantiate.
        data (dict): The dictionary containing data to populate the dataclass.

    Returns:
        Any: An instance of the provided dataclass, populated with the provided data.
    """
    kwargs = {}
    for f in fields(cls):
        if f.name not in data:
            continue
        v = data[f.name]
        t = f.type
        if is_dataclass(t) and isinstance(v, dict):
            kwargs[f.name] = build_dataclass(t, v)
        else:
            kwargs[f.name] = v
    return cls(**kwargs)


class SettingsService:
    """
    Manages application settings, providing functionality to load, update, and reload
    settings dynamically.

    This class serves as a service layer between a settings repository and the application,
    allowing for the merging of default settings with those persisted in a repository.
    It ensures that the latest settings are always available and maintains the current
    state of the settings in memory.

    Attributes:
        repo (SettingsRepository): The repository interface used to persist and load settings.
        defaults (Settings): The default settings structure applied when no overrides are
            provided by the repository.
        current (Settings): The in-memory representation of the current active settings,
            updated after every load, reload, or update operation.
    """

    def __init__(self, repo: SettingsRepository, defaults: Settings | None = None):
        self._repo = repo
        self._defaults = defaults or Settings()
        self._current = self._load()

    def _load(self) -> Settings:
        raw = self._repo.load()
        merged = self._merge(asdict(self._defaults), raw)
        return build_dataclass(Settings, merged)

    def get(self) -> Settings:
        return self._current

    def reload(self) -> Settings:
        self._current = self._load()
        return self._current

    def update(self, patch: dict) -> Settings:
        """
        Updates the settings by merging the given patch with existing settings and saving the result.

        This method applies a provided patch (a dictionary) to the current settings. It first merges
        the default settings with those loaded from the repository, then applies the patch on top of
        this merged result. The updated settings are saved back to the repository, and the current
        settings are reloaded to reflect these changes.

        Args:
            patch (dict): A dictionary containing new settings to be merged with the existing settings.

        Returns:
            Settings: An updated instance of the settings after applying the patch.
        """
        current_raw = self._merge(asdict(self._defaults), self._repo.load())
        new_raw = self._merge(current_raw, patch)
        self._repo.save(new_raw)
        self._current = self._load()
        return self._current

    def _merge(self, base: dict, override: dict) -> dict:
        """
        Merges two dictionaries recursively. Values in the `override` dictionary
        replace or update the corresponding entries in the `base` dictionary. If both
        dictionaries contain a value for the same key and those values are themselves
        dictionaries, the function calls itself recursively to merge the nested
        dictionaries.

        Args:
            base (dict): The base dictionary.
            override (dict): The dictionary containing values to override or update
                the base dictionary.

        Returns:
            dict: A new dictionary that represents the merged result of the `base`
                and `override` dictionaries.
        """
        out = dict(base)
        for k, v in override.items():
            if isinstance(v, dict) and isinstance(out.get(k), dict):
                out[k] = self._merge(out[k], v)
            else:
                out[k] = v
        return out
