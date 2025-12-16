import unittest

from cmdbox.domain.validators import (
    CommandValidator,
    CommandValidatorConfig,
    VariableValidator,
    VariableValidatorConfig,
    TagValidator,
)
from cmdbox.exceptions import ValidationError


class TestCommandValidator(unittest.TestCase):

    def setUp(self):
        # Use small limits to make boundary tests concise
        self.config = CommandValidatorConfig(
            reserved_aliases=frozenset({"help", "init", "rm"}),
            max_alias_length=5,
            max_description_length=10,
        )
        self.validator = CommandValidator(config=self.config)

    # --- validate_alias ---
    def test_alias_empty_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_alias("")
        self.assertIn("Alias cannot be empty", str(ctx.exception))

    def test_alias_whitespace_only_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_alias("   \t  ")
        self.assertIn("Alias cannot contain only whitespace", str(ctx.exception))

    def test_alias_with_internal_space_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_alias("ab cd")
        self.assertIn("Alias cannot contain spaces", str(ctx.exception))

    def test_alias_reserved_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_alias("help")
        self.assertIn("reserved", str(ctx.exception))

    def test_alias_max_length_boundary_ok(self):
        self.validator.validate_alias("a" * self.config.max_alias_length)

    def test_alias_too_long_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_alias("a" * (self.config.max_alias_length + 1))
        self.assertIn("too long", str(ctx.exception))

    def test_alias_leading_trailing_spaces_allowed(self):
        # Stripping is checked internally; should not raise
        self.validator.validate_alias("  abc  ")

    # --- validate_template ---
    def test_template_empty_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_template("")
        self.assertIn("Template cannot be empty", str(ctx.exception))

    def test_template_whitespace_only_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_template("   ")
        self.assertIn("whitespace", str(ctx.exception))

    def test_template_non_empty_ok(self):
        self.validator.validate_template(" echo hi ")

    # --- validate_description ---
    def test_description_none_ok(self):
        self.validator.validate_description(None)

    def test_description_empty_ok(self):
        self.validator.validate_description("")

    def test_description_max_length_boundary_ok(self):
        self.validator.validate_description("x" * self.config.max_description_length)

    def test_description_too_long_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_description(
                "x" * (self.config.max_description_length + 1)
            )
        self.assertIn("too long", str(ctx.exception))

    # --- validate_no_self_reference ---
    def test_template_self_reference_raises(self):
        alias = "test"
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_no_self_reference(alias, f"run <{alias}>")
        self.assertIn("self-reference", str(ctx.exception))

    def test_template_with_similar_but_not_self_reference_ok(self):
        # Not exactly <alias>
        self.validator.validate_no_self_reference("test", "run <test2>")

    # --- validate_create and validate_update integration ---
    def test_validate_create_happy_path(self):
        self.validator.validate_create(
            alias="abc", template="echo hi", description=None
        )

    def test_validate_create_raises_on_any_invalid_field(self):
        with self.assertRaises(ValidationError):
            self.validator.validate_create(alias="", template="echo hi")
        with self.assertRaises(ValidationError):
            self.validator.validate_create(alias="ok", template="")
        with self.assertRaises(ValidationError):
            self.validator.validate_create(alias="ok", template="use <ok>")

    def test_validate_update_validates_only_provided_fields(self):
        # None fields should be ignored
        self.validator.validate_update(alias=None, template=None, description=None)
        self.validator.validate_update(alias=" ok ")  # ok after stripping
        self.validator.validate_update(template="echo")
        self.validator.validate_update(
            description="a" * self.config.max_description_length
        )

    def test_validate_update_triggers_self_reference_only_when_both_present(self):
        # Should not raise because only alias provided
        self.validator.validate_update(alias="abc")
        # Should not raise because only template provided
        self.validator.validate_update(template="run <abc>")
        # Should raise when both provided and template self-references alias
        with self.assertRaises(ValidationError):
            self.validator.validate_update(alias="abc", template="run <abc>")


class TestVariableValidator(unittest.TestCase):

    def setUp(self):
        self.config = VariableValidatorConfig(
            reserved_names=frozenset({"help", "init", "rm"}),
            max_name_length=5,
        )
        self.validator = VariableValidator(config=self.config)

    # --- validate_name ---
    def test_name_empty_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_name("")
        self.assertIn("cannot be empty", str(ctx.exception))

    def test_name_whitespace_only_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_name("   ")
        self.assertIn("only whitespace", str(ctx.exception))

    def test_name_with_internal_space_raises(self):
        with self.assertRaises(ValidationError):
            self.validator.validate_name("ab cd")

    def test_name_reserved_raises(self):
        with self.assertRaises(ValidationError):
            self.validator.validate_name("help")

    def test_name_max_length_boundary_ok(self):
        self.validator.validate_name("a" * self.config.max_name_length)

    def test_name_too_long_raises(self):
        with self.assertRaises(ValidationError):
            self.validator.validate_name("a" * (self.config.max_name_length + 1))

    def test_name_leading_trailing_spaces_allowed(self):
        self.validator.validate_name("  abc  ")

    # --- validate_value ---
    def test_value_none_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_value(None)
        self.assertIn("cannot be None", str(ctx.exception))

    def test_value_empty_string_ok(self):
        self.validator.validate_value("")

    # --- validate_create/update ---
    def test_validate_create_happy_path(self):
        self.validator.validate_create(name="abc", value="val")

    def test_validate_update_validates_only_provided_fields(self):
        self.validator.validate_update(name=None, value=None)
        self.validator.validate_update(name=" abc ", value=None)
        self.validator.validate_update(name=None, value="")

    # --- validate_no_self_reference (standalone method) ---
    def test_value_self_reference_raises(self):
        with self.assertRaises(ValidationError) as ctx:
            self.validator.validate_no_self_reference("name", "<name> value")
        self.assertIn("self-reference", str(ctx.exception))

    def test_value_similar_but_not_self_reference_ok(self):
        self.validator.validate_no_self_reference("name", "<name2> value")


class TestTagValidator(unittest.TestCase):

    def setUp(self):
        # Note: TagValidator currently reuses a VariableValidatorConfig with an extra
        # max_description_length field.
        self.config = VariableValidatorConfig(
            reserved_names=frozenset({"help", "init", "rm"}),
            max_name_length=5,
            max_description_length=8,
        )
        self.validator = TagValidator(config=self.config)

    def test_name_rules_match_variable_validator(self):
        with self.assertRaises(ValidationError):
            self.validator.validate_name("")
        with self.assertRaises(ValidationError):
            self.validator.validate_name("   ")
        with self.assertRaises(ValidationError):
            self.validator.validate_name("ab cd")
        with self.assertRaises(ValidationError):
            self.validator.validate_name("help")
        # boundary ok
        self.validator.validate_name("a" * self.config.max_name_length)
        # too long raises
        with self.assertRaises(ValidationError):
            self.validator.validate_name("a" * (self.config.max_name_length + 1))

    def test_description_none_or_empty_ok(self):
        self.validator.validate_description(None)
        self.validator.validate_description("")

    def test_description_max_length_boundary_ok(self):
        self.validator.validate_description("x" * self.config.max_description_length)

    def test_description_too_long_raises(self):
        with self.assertRaises(ValidationError):
            self.validator.validate_description(
                "x" * (self.config.max_description_length + 1)
            )

    def test_validate_create_and_update(self):
        # happy paths
        self.validator.validate_create(name="tag", description=None)
        self.validator.validate_create(name="tag", description="desc")
        self.validator.validate_update(name=None, description=None)
        # invalids
        with self.assertRaises(ValidationError):
            self.validator.validate_create(name="", description="desc")
        with self.assertRaises(ValidationError):
            self.validator.validate_update(name="  ", description=None)
