from __future__ import annotations

from dataclasses import dataclass

from cmdbox.exceptions import ValidationError


@dataclass(frozen=True)
class CommandValidatorConfig:
    """Configuration for command validation rules."""

    reserved_aliases: frozenset[str] = frozenset(
        {
            "help",
            "init",
            "list",
            "ls",
            "add",
            "rm",
            "delete",
        }
    )
    max_alias_length: int = 100
    max_description_length: int = 1000


class CommandValidator:

    def __init__(self, config: CommandValidatorConfig | None = None):
        self.config = config or CommandValidatorConfig()

    def validate_create(
        self,
        *,
        alias: str,
        template: str,
        description: str | None = None,
    ) -> None:
        """
        Validates the parameters for creating a Command. Ensures that the provided alias,
        template, and description meet the required conditions and do not contain
        self-referencing patterns.

        Args:
            alias (str): The alias to validate.
            template (str): The template to validate.
            description (str | None): The optional description to validate.
        """
        self.validate_alias(alias)
        self.validate_template(template)
        self.validate_description(description)
        self.validate_no_self_reference(alias, template)

    def validate_update(
        self,
        *,
        alias: str | None = None,
        template: str | None = None,
        description: str | None = None,
    ):
        """
        Validates various update parameters, ensuring they meet the required conditions
        and constraints. This method can validate a combination of alias, template,
        and description, applying specific checks for each provided parameter.

        Args:
            alias (str | None): An optional alias to validate. If provided, it will
                be checked using the `self.validate_alias` method.
            template (str | None): An optional template to validate. If provided, it
                will be checked using the `self.validate_template` method.
            description (str | None): An optional description to validate. If
                provided, it will be checked using the `self.validate_description`
                method.
        """
        if alias is not None:
            self.validate_alias(alias)
        if template is not None:
            self.validate_template(template)
        if description is not None:
            self.validate_description(description)
        if alias is not None and template is not None:
            self.validate_no_self_reference(alias, template)

    def validate_alias(self, alias: str) -> None:
        if not alias:
            raise ValidationError("Alias cannot be empty.")

        stripped = alias.strip()
        if not stripped:
            raise ValidationError("Alias cannot contain only whitespace.")

        if " " in stripped:
            raise ValidationError("Alias cannot contain spaces.")

        if stripped in self.config.reserved_aliases:
            raise ValidationError(f"Alias '{stripped}' is reserved.")

        if len(stripped) > self.config.max_alias_length:
            raise ValidationError(
                f"Alias '{stripped}' is too long. Maximum length is {self.config.max_alias_length}."
            )

    def validate_template(self, template: str) -> None:
        if not template:
            raise ValidationError("Template cannot be empty.")

        stripped = template.strip()
        if not stripped:
            raise ValidationError("Template cannot contain only whitespace.")

    def validate_description(self, description: str | None) -> None:
        if not description:
            return
        if len(description) > self.config.max_description_length:
            raise ValidationError(
                f"Description is too long. Maximum length is {self.config.max_description_length}."
            )

    def validate_no_self_reference(self, alias: str, template: str) -> None:
        self_reference = f"<{alias}>"
        if self_reference in template:
            raise ValidationError(
                f"Template cannot contain self-reference: {self_reference}"
            )


@dataclass(frozen=True)
class VariableValidatorConfig:
    """Configuration for variable validation rules."""

    reserved_names: frozenset[str] = frozenset(
        {
            "help",
            "init",
            "list",
            "ls",
            "add",
            "rm",
            "delete",
        }
    )
    max_name_length: int = 100


class VariableValidator:

    def __init__(self, config: VariableValidatorConfig | None = None):
        self.config = config or VariableValidatorConfig()

    def validate_create(self, name: str, value: str) -> None:
        self.validate_name(name)
        self.validate_value(value)

    def validate_update(self, name: str | None, value: str | None) -> None:
        if name is not None:
            self.validate_name(name)
        if value is not None:
            self.validate_value(value)

    def validate_name(self, name: str) -> None:
        if not name:
            raise ValidationError("Variable name cannot be empty.")

        stripped = name.strip()
        if not stripped:
            raise ValidationError("Variable name cannot contain only whitespace.")

        if " " in stripped:
            raise ValidationError("Variable name cannot contain spaces.")

        if stripped in self.config.reserved_names:
            raise ValidationError(f"Variable name '{stripped}' is reserved.")

        if len(stripped) > self.config.max_name_length:
            raise ValidationError(
                f"Variable name '{stripped}' is too long. Maximum length is {self.config.max_name_length}."
            )

    def validate_value(self, value: str) -> None:
        if value is None:
            raise ValidationError("Variable value cannot be None.")

    def validate_no_self_reference(self, name: str, value: str) -> None:
        self_reference = f"<{name}>"
        if self_reference in value:
            raise ValidationError(
                f"Variable value cannot contain self-reference: {self_reference}"
            )


@dataclass(frozen=True)
class VariableValidatorConfig:
    """Configuration for variable validation rules."""

    reserved_names: frozenset[str] = frozenset(
        {
            "help",
            "init",
            "list",
            "ls",
            "add",
            "rm",
            "delete",
        }
    )
    max_name_length: int = 100
    max_description_length: int = 1000


class TagValidator:

    def __init__(self, config: VariableValidatorConfig | None = None):
        self.config = config or VariableValidatorConfig()

    def validate_create(self, name: str, description: str | None) -> None:
        self.validate_name(name)
        self.validate_description(description)

    def validate_update(self, name: str | None, description: str | None) -> None:
        if name is not None:
            self.validate_name(name)
        if description is not None:
            self.validate_description(description)

    def validate_name(self, name: str) -> None:
        if not name:
            raise ValidationError("Variable name cannot be empty.")

        stripped = name.strip()
        if not stripped:
            raise ValidationError("Variable name cannot contain only whitespace.")

        if " " in stripped:
            raise ValidationError("Variable name cannot contain spaces.")

        if stripped in self.config.reserved_names:
            raise ValidationError(f"Variable name '{stripped}' is reserved.")

        if len(stripped) > self.config.max_name_length:
            raise ValidationError(
                f"Variable name '{stripped}' is too long. Maximum length is {self.config.max_name_length}."
            )

    def validate_description(self, description: str | None) -> None:
        if not description:
            return
        if len(description) > self.config.max_description_length:
            raise ValidationError(
                f"Description is too long. Maximum length is {self.config.max_description_length}."
            )
