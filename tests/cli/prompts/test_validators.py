import unittest
from prompt_toolkit.document import Document
from prompt_toolkit.validation import ValidationError
from cmdbox.cli.prompts.validators import (
    AliasValidator,
    TemplateValidator,
    NameValidator,
)


class TestAliasValidator(unittest.TestCase):

    def test_validate_valid(self):
        validator = AliasValidator()
        # Should not raise any error
        validator.validate(Document("myalias"))
        validator.validate(Document("my-alias_123"))

    def test_validate_empty(self):
        validator = AliasValidator()
        with self.assertRaises(ValidationError) as cm:
            validator.validate(Document(""))
        self.assertEqual(cm.exception.message, "Alias cannot be empty")

        with self.assertRaises(ValidationError) as cm:
            validator.validate(Document("   "))
        self.assertEqual(cm.exception.message, "Alias cannot be empty")

    def test_validate_spaces(self):
        validator = AliasValidator()
        with self.assertRaises(ValidationError) as cm:
            validator.validate(Document("my alias"))
        self.assertEqual(cm.exception.message, "Alias cannot contain spaces")

    def test_validate_invalid_chars(self):
        validator = AliasValidator()
        with self.assertRaises(ValidationError) as cm:
            validator.validate(Document("alias<"))
        self.assertEqual(
            cm.exception.message, "Alias cannot contain '<' or '>' characters"
        )

        with self.assertRaises(ValidationError) as cm:
            validator.validate(Document("alias>"))
        self.assertEqual(
            cm.exception.message, "Alias cannot contain '<' or '>' characters"
        )


class TestTemplateValidator(unittest.TestCase):

    def test_validate_valid(self):
        validator = TemplateValidator()
        validator.validate(Document("some template"))
        validator.validate(Document("{{var}}"))

    def test_validate_empty(self):
        validator = TemplateValidator()
        with self.assertRaises(ValidationError) as cm:
            validator.validate(Document(""))
        self.assertEqual(cm.exception.message, "Template cannot be empty")

        with self.assertRaises(ValidationError) as cm:
            validator.validate(Document("  "))
        self.assertEqual(cm.exception.message, "Template cannot be empty")


class TestNameValidator(unittest.TestCase):

    def test_validate_valid(self):
        validator = NameValidator()
        validator.validate(Document("myname"))

    def test_validate_empty(self):
        validator = NameValidator()
        with self.assertRaises(ValidationError) as cm:
            validator.validate(Document(""))
        self.assertEqual(cm.exception.message, "Name cannot be empty")

    def test_validate_spaces(self):
        validator = NameValidator()
        with self.assertRaises(ValidationError) as cm:
            validator.validate(Document("my name"))
        self.assertEqual(cm.exception.message, "Name cannot contain spaces")

    def test_validate_invalid_chars(self):
        validator = NameValidator()
        with self.assertRaises(ValidationError) as cm:
            validator.validate(Document("name<"))
        self.assertEqual(
            cm.exception.message, "Name cannot contain '<' or '>' characters"
        )

        with self.assertRaises(ValidationError) as cm:
            validator.validate(Document("name>"))
        self.assertEqual(
            cm.exception.message, "Name cannot contain '<' or '>' characters"
        )
