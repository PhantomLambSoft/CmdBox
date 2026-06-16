import json
from dataclasses import dataclass, field
from datetime import datetime, timezone
from pathlib import Path

from cmdbox.resolve.reference_parsing import (
    extract_references,
    read_angle_token,
    parse_kind_and_key,
)
from cmdbox.resolve.type_defs import RefKind
from cmdbox.services.command_services import CommandServices
from cmdbox.models import Command, Variable
from cmdbox.services.variable_services import VariableServices
from cmdbox.common.io import atomic_write_text


_EXPORT_LIMIT = 10_000


@dataclass
class ExportResult:
    path: Path
    commands: list[str] = field(default_factory=list)
    variables: list[str] = field(default_factory=list)
    transient_commands: list[str] = field(default_factory=list)
    transient_variables: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)


def collect_deep_commands(
    aliases: list[str], cmd_service: CommandServices
) -> dict[str, Command]:
    collected: dict[str, Command] = {}
    stack = list(aliases)

    while stack:
        alias = stack.pop()
        if alias in collected:
            continue
        cmd = cmd_service.get_command_or_none(alias)
        if cmd is None:
            continue  # reference to a command that doesn't exist. Skip silently
        collected[alias] = cmd

        for kind, key in extract_references(cmd.template):
            if kind == RefKind.COMMAND and key not in collected:
                stack.append(key)

    return collected


def collect_deep_variables(
    names: list[str], commands: dict[str, Command], var_service: VariableServices
) -> dict[str, Variable]:
    collected: dict[str, Variable] = {}
    stack = list(names)

    for cmd in commands.values():
        for kind, key in extract_references(cmd.template):
            if kind == RefKind.VARIABLE and key not in collected:
                stack.append(key)

    while stack:
        name = stack.pop()
        if name in collected:
            continue
        var = var_service.get_variable_or_none(name)
        if var is None:
            continue
        collected[name] = var

        for kind, key in extract_references(var.value):
            if kind == RefKind.VARIABLE and key not in collected:
                stack.append(key)

    return collected


def flatten_template(
    template: str,
    cmd_service: CommandServices,
    var_service: VariableServices,
    _seen: frozenset[str] = frozenset(),
) -> str:
    out: list[str] = []
    i = 0
    while i < len(template):
        ch = template[i]
        if ch == "\\":
            out.append(template[i : i + 2])
            i += 2
            continue
        if ch == "<":
            token_inner, raw_token, next_i = read_angle_token(template, i)
            if token_inner is None:
                out.append("<")
                i += 1
                continue

            kind, key = parse_kind_and_key(token_inner)
            label = f"{'cmd' if kind == RefKind.COMMAND else 'var'}:{key}"

            if label in _seen:
                out.append(f"<{raw_token}>")
                i = next_i
                continue

            if kind == RefKind.COMMAND:
                rec = cmd_service.get_command_or_none(key)
                if rec is None:
                    out.append(f"<{raw_token}>")
                else:
                    out.append(
                        flatten_template(
                            rec.template, cmd_service, var_service, _seen | {label}
                        )
                    )

            else:
                rec = var_service.get_variable_or_none(key)
                if rec is None:
                    out.append(f"<{raw_token}>")
                else:
                    out.append(
                        flatten_template(
                            rec.value, cmd_service, var_service, _seen | {label}
                        )
                    )

            i = next_i
            continue
        out.append(ch)
        i += 1
    return "".join(out)


def _serialize_command(
    cmd: Command,
    flatten: bool,
    cmd_service: CommandServices,
    var_service: VariableServices,
) -> dict:
    template = (
        flatten_template(cmd.template, cmd_service, var_service)
        if flatten
        else cmd.template
    )
    return {
        "alias": cmd.alias,
        "template": template,
        "description": cmd.description,
        "tags": [tag.name for tag in cmd.tags],
        "cwd": cmd.cwd,
        "shell": cmd.shell,
        "env": json.loads(cmd.env) if cmd.env else None,
        "timeout": cmd.timeout,
    }


def _serialize_variable(
    var: Variable,
    flatten: bool,
    cmd_service: CommandServices,
    var_service: VariableServices,
) -> dict:
    value = (
        flatten_template(var.value, cmd_service, var_service) if flatten else var.value
    )
    return {
        "name": var.name,
        "value": value,
        "tags": [tag.name for tag in var.tags],
    }


def _build_document(
    type_label: str,
    commands: list[dict],
    variables: list[dict],
) -> dict:
    return {
        "version": "1",
        "type": type_label,
        "exported_at": datetime.now(timezone.utc).isoformat(),
        "commands": commands,
        "variables": variables,
    }


def _resolve_output_path(output_path: str | None, type_label: str) -> Path:
    date_str = datetime.now().strftime("%Y-%m-%d")
    filename = f"cmdbox-{type_label}-{date_str}.json"
    if output_path is None:
        return Path.cwd() / filename
    p = Path(output_path)
    if p.is_dir():
        return p / filename
    return p


class ExportService:

    def __init__(self, cmd_service: CommandServices, var_service: VariableServices):
        self._cmd_service = cmd_service
        self._var_service = var_service

    def export_cmds(
        self,
        aliases: list[str],
        tag: str | None = None,
        flatten: bool = False,
        output_path: str | None = None,
    ) -> ExportResult:
        result = ExportResult(path=_resolve_output_path(output_path, "cmds"))

        if aliases:
            target_aliases = aliases
        elif tag:
            target_aliases = [
                c.alias
                for c in self._cmd_service.list_commands(tags=tag, limit=_EXPORT_LIMIT)
            ]
        else:
            target_aliases = [
                c.alias for c in self._cmd_service.list_commands(limit=_EXPORT_LIMIT)
            ]

        if flatten:
            serialized_cmds = []
            for alias in target_aliases:
                cmd = self._cmd_service.get_command_or_none(alias)
                if cmd is None:
                    result.warnings.append(f"Command {alias} not found")
                    continue
                serialized_cmds.append(
                    _serialize_command(
                        cmd,
                        flatten=True,
                        cmd_service=self._cmd_service,
                        var_service=self._var_service,
                    )
                )
                result.commands.append(alias)
            serialized_vars = []
        else:
            collected_cmds = collect_deep_commands(target_aliases, self._cmd_service)
            collected_vars = collect_deep_variables(
                [], collected_cmds, self._var_service
            )

            target_set = set(target_aliases)
            for alias in target_aliases:
                if alias not in collected_cmds:
                    result.warnings.append(f"Command {alias} not found")

            serialized_cmds = [
                _serialize_command(
                    cmd,
                    flatten=False,
                    cmd_service=self._cmd_service,
                    var_service=self._var_service,
                )
                for cmd in collected_cmds.values()
            ]
            serialized_vars = [
                _serialize_variable(
                    var,
                    flatten=False,
                    cmd_service=self._cmd_service,
                    var_service=self._var_service,
                )
                for var in collected_vars.values()
            ]
            result.commands = [x for x in collected_cmds if x in target_set]
            result.transient_commands = [
                x for x in collected_cmds if x not in target_set
            ]
            result.transient_variables = list(collected_vars.keys())

        doc = _build_document("cmds", serialized_cmds, serialized_vars)
        atomic_write_text(result.path, json.dumps(doc, indent=2))
        return result

    def export_vars(
        self,
        names: list[str] | None = None,
        tag: str | None = None,
        flatten: bool = False,
        output_path: str | None = None,
    ) -> ExportResult:
        result = ExportResult(path=_resolve_output_path(output_path, "vars"))

        if names:
            target_names = names
        elif tag:
            target_names = [
                v.name
                for v in self._var_service.list_variables(tags=tag, limit=_EXPORT_LIMIT)
            ]
        else:
            target_names = [
                v.name for v in self._var_service.list_variables(limit=_EXPORT_LIMIT)
            ]

        if flatten:
            serialized_vars = []
            for name in target_names:
                var = self._var_service.get_variable_or_none(name)
                if var is None:
                    result.warnings.append(f"Variable {name} not found")
                    continue
                serialized_vars.append(
                    _serialize_variable(
                        var,
                        flatten=True,
                        cmd_service=self._cmd_service,
                        var_service=self._var_service,
                    )
                )
                result.variables.append(name)
        else:
            collected_vars = collect_deep_variables(target_names, {}, self._var_service)

            target_names = set(target_names)
            for name in target_names:
                if name not in collected_vars:
                    result.warnings.append(f"Variable {name} not found")

            serialized_vars = [
                _serialize_variable(
                    var,
                    flatten=False,
                    cmd_service=self._cmd_service,
                    var_service=self._var_service,
                )
                for var in collected_vars.values()
            ]

            result.variables = [x for x in collected_vars if x in target_names]
            result.transient_variables = [
                x for x in collected_vars if x not in target_names
            ]

        doc = _build_document("vars", [], serialized_vars)
        atomic_write_text(result.path, json.dumps(doc, indent=2))
        return result

    def export_all(
        self, flatten: bool = False, output_path: str | None = None
    ) -> ExportResult:
        result = ExportResult(path=_resolve_output_path(output_path, "all"))

        all_cmds = self._cmd_service.list_commands(limit=_EXPORT_LIMIT)
        all_vars = self._var_service.list_variables(limit=_EXPORT_LIMIT)

        serialized_cmds = [
            _serialize_command(
                cmd,
                flatten=flatten,
                cmd_service=self._cmd_service,
                var_service=self._var_service,
            )
            for cmd in all_cmds
        ]
        serialized_vars = [
            _serialize_variable(
                var,
                flatten=False,
                cmd_service=self._cmd_service,
                var_service=self._var_service,
            )
            for var in all_vars
        ]
        result.commands = [cmd.alias for cmd in all_cmds]
        result.variables = [var.name for var in all_vars]

        doc = _build_document("all", serialized_cmds, serialized_vars)
        atomic_write_text(result.path, json.dumps(doc, indent=2))
        return result
