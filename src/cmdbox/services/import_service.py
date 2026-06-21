import json
from dataclasses import dataclass, field
from pathlib import Path

from cmdbox.common.graph_utils import find_cycle
from cmdbox.resolve.reference_parsing import extract_references
from cmdbox.resolve.type_defs import RefKind
from cmdbox.services.command_services import CommandServices
from cmdbox.services.errors import ImportCycleError, ImportFileError
from cmdbox.services.tag_services import TagServices
from cmdbox.services.variable_services import VariableServices

_SUPPORTED_VERSIONS = {"1"}


KIND_PREFIX = {
    RefKind.COMMAND: "cmd",
    RefKind.VARIABLE: "var",
}


@dataclass
class ImportResult:
    commands_created: list[str] = field(default_factory=list)
    commands_skipped: list[str] = field(default_factory=list)
    commands_overwritten: list[str] = field(default_factory=list)
    variables_created: list[str] = field(default_factory=list)
    variables_skipped: list[str] = field(default_factory=list)
    variables_overwritten: list[str] = field(default_factory=list)
    preview: bool = False


def _label(kind: RefKind, key: str) -> str:
    return f"{KIND_PREFIX[kind]}:{key}"


def build_dependency_graph(import_data: dict) -> dict[str, list[str]]:
    """
    Builds a dependency graph from the provided import data.

    This function processes the "commands" and "variables" in the provided
    import data to construct a mapping of labels to their dependencies.
    Each label represents a command or variable, and its dependencies are
    derived from references extracted from their respective templates or values.

    Args:
        import_data (dict): A dictionary containing "commands" and
            "variables" data. Each command or variable includes an alias,
            and its references are determined by processing its template or
            value.

    Returns:
        dict[str, list[str]]: A dictionary where the keys are labels
        representing commands or variables, and the values are lists of labels
        representing their dependencies.
    """
    deps: dict[str, list[str]] = {}

    for cmd in import_data.get("commands", []):
        label = _label(RefKind.COMMAND, cmd["alias"])
        deps[label] = [
            _label(kind, key) for kind, key in extract_references(cmd["template"])
        ]

    for var in import_data.get("variables", []):
        label = _label(RefKind.VARIABLE, var["name"])
        deps[label] = [
            _label(kind, key) for kind, key in extract_references(var["value"])
        ]

    return deps


def validate_no_cycles(deps: dict[str, list[str]]) -> None:
    """
    Validates that there are no circular dependencies in the provided dependency graph.

    This function takes a dictionary where the keys represent labels and the values are lists
    of dependencies for each label. It checks if there are any cycles in the graph formed
    by the dependencies and raises an error if a cycle is detected.

    Args:
        deps (dict[str, list[str]]): A dictionary representing the dependency graph. Keys are
            the labels, and values are lists of labels each key depends on.

    Raises:
        ImportCycleError: If a circular dependency is detected in the graph.
    """

    def get_deps(label: str) -> list[str]:
        return deps.get(label, [])

    for label in deps:
        cycle = find_cycle(label, get_deps)
        if cycle is not None:
            raise ImportCycleError(cycle)


def _parse_import_file(path: str | Path) -> dict:
    try:
        text = Path(path).read_text(encoding="utf-8")
        data = json.loads(text)
    except (OSError, json.JSONDecodeError, UnicodeDecodeError) as e:
        raise ImportFileError(f"Could not read import file: {e}") from e

    if not isinstance(data, dict):
        raise ImportFileError("Import file must be a JSON object")

    version = data.get("version")
    if version not in _SUPPORTED_VERSIONS:
        supported = ", ".join(_SUPPORTED_VERSIONS)
        raise ImportFileError(
            f"Unsupported import file version: {version}. Supported versions: {supported}"
        )

    return data


class ImportService:

    def __init__(
        self,
        cmd_service: CommandServices,
        var_service: VariableServices,
        tag_service: TagServices,
    ):
        self._cmd_service = cmd_service
        self._var_service = var_service
        self.tag_service = tag_service
        self.result = ImportResult()

    def import_file(
        self, path: str | Path, overwrite: bool = False, preview: bool = False
    ) -> ImportResult:
        self.result.preview = preview
        data = _parse_import_file(path)

        deps = build_dependency_graph(data)
        validate_no_cycles(deps)

        cmd_items = data.get("commands", [])
        var_items = data.get("variables", [])

        cmd_actions: list[tuple[dict, str]] = []
        for cmd_data in cmd_items:
            alias = cmd_data["alias"]
            existing = self._cmd_service.get_command_or_none(alias)
            if existing is None:
                cmd_actions.append((cmd_data, "create"))
            elif overwrite:
                cmd_actions.append((cmd_data, "overwrite"))
            else:
                cmd_actions.append((cmd_data, "skip"))

        var_actions: list[tuple[dict, str]] = []
        for var_data in var_items:
            name = var_data["name"]
            existing = self._var_service.get_variable_or_none(name)
            if existing is None:
                var_actions.append((var_data, "create"))
            elif overwrite:
                var_actions.append((var_data, "overwrite"))
            else:
                var_actions.append((var_data, "skip"))

        self.classify_commands(cmd_actions)
        self.classify_variables(var_actions)

        if preview:
            return self.result

        tag_names = self._collect_tag_names(cmd_actions, var_actions)
        self._ensure_tag_exists(tag_names)

        self.handle_commands(cmd_actions)
        self.handle_variables(var_actions)

        return self.result

    def classify_commands(self, cmd_actions: list[tuple[dict, str]]):
        """
        Classifies commands based on their specified actions and updates the result object
        with the corresponding command names categorized into created, overwritten, or skipped.

        This is handled as a separate method with a separate iteration so
        that `preview` can still return a meaningful result.

        Args:
            cmd_actions (list[tuple[dict, str]]): A list of tuples, where each tuple
            contains a dictionary with command data and a string representing the
            action to be performed. The dictionary must include an 'alias' key.

        """
        for data, action in cmd_actions:
            alias = data["alias"]
            if action == "create":
                self.result.commands_created.append(alias)
            elif action == "overwrite":
                self.result.commands_overwritten.append(alias)
            elif action == "skip":
                self.result.commands_skipped.append(alias)

    def handle_commands(self, cmd_actions: list[tuple[dict, str]]):
        for cmd_data, action in cmd_actions:
            alias = cmd_data["alias"]

            if action == "skip":
                self.result.commands_skipped.append(alias)
                continue

            tags = cmd_data.get("tags", [])

            if action == "create":
                self._cmd_service.create_command(
                    alias=alias,
                    template=cmd_data["template"],
                    description=cmd_data.get("description", None),
                    tags=tags,
                    cwd=cmd_data.get("cwd", None),
                    shell=cmd_data.get("shell", None),
                    env=cmd_data.get("env", None),
                    timeout=cmd_data.get("timeout", None),
                )
                self.result.commands_created.append(alias)
            else:
                self.overwrite_existing_command(alias, cmd_data, tags)
                self.result.commands_overwritten.append(alias)

    def overwrite_existing_command(self, alias: str, cmd_data: dict, tags: list[str]):
        existing = self._cmd_service.get_command(alias)
        current_tags = {ct.tag.name for ct in existing.tags}
        new_tag = set(tags)
        self._cmd_service.update_command(
            alias=alias,
            template=cmd_data["template"],
            description=cmd_data.get("description", None),
            cwd=cmd_data.get("cwd", None),
            shell=cmd_data.get("shell", None),
            env=cmd_data.get("env", None),
            timeout=cmd_data.get("timeout", None),
        )
        tags_to_remove = list(current_tags - new_tag)
        tags_to_add = list(new_tag - current_tags)
        if tags_to_remove:
            self._cmd_service.remove_tags(alias, tags_to_remove)
        if tags_to_add:
            self._cmd_service.add_tags(alias, tags_to_add)

    def classify_variables(self, var_actions: list[tuple[dict, str]]):
        """
        Classifies variables based on the specified actions and updates the result object with
        the corresponding variable names categorized into created, overwritten, or skipped.

        This is handled as a separate method with a separate iteration so
        that `preview` can still return a meaningful result.

        Args:
            var_actions (list[tuple[dict, str]]): A list of tuples, where each tuple contains a
                dictionary with variable data and a string representing the associated action
                to be performed. The dictionary must include a 'name' key.
        """
        for data, action in var_actions:
            name = data["name"]
            if action == "create":
                self.result.variables_created.append(name)
            elif action == "overwrite":
                self.result.variables_overwritten.append(name)
            elif action == "skip":
                self.result.variables_skipped.append(name)

    def handle_variables(self, var_actions: list[tuple[dict, str]]):
        for var_data, action in var_actions:
            name = var_data["name"]

            if action == "skip":
                self.result.variables_skipped.append(name)
                continue

            tags = var_data.get("tags", [])

            if action == "create":
                self._var_service.create_variable(
                    name, value=var_data["value"], tags=tags
                )
                self.result.variables_created.append(name)
            else:
                self.overwrite_existing_variable(name, var_data, tags)
                self.result.variables_overwritten.append(name)

    def overwrite_existing_variable(self, name: str, var_data: dict, tags: list[str]):
        existing = self._var_service.get_variable(name)
        current_tags = {ct.tag.name for ct in existing.tags}
        new_tags = set(tags)
        self._var_service.update_variable(name, value=var_data["value"])
        tags_to_remove = list(current_tags - new_tags)
        tags_to_add = list(new_tags - current_tags)
        if tags_to_remove:
            self._var_service.remove_tags(name, tags_to_remove)
        if tags_to_add:
            self._var_service.add_tags(name, tags_to_add)

    def _collect_tag_names(
        self, cmd_actions: list[tuple[dict, str]], var_actions: list[tuple[dict, str]]
    ) -> set[str]:
        names: set[str] = set()
        for data, action in cmd_actions + var_actions:
            if action != "skip":
                names.update(data.get("tags", []))
        return names

    def _ensure_tag_exists(self, tag_names: set[str]) -> None:
        for name in tag_names:
            if not self.tag_service.tag_exists(name):
                self.tag_service.create_tag(name)
