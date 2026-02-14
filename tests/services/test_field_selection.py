import unittest
from cmdbox.services.field_selection import FieldSelectionResolver
from cmdbox.services.errors import EmptyFieldSelectionError, UnknownFieldError


class TestFieldSelectionResolver(unittest.TestCase):

    def test_resolve_none_returns_allowed_fields(self):
        resolver = FieldSelectionResolver(allowed_fields=["id", "name", "desc"])
        result = resolver.resolve(None)
        self.assertEqual(["id", "name", "desc"], result)

    def test_resolve_none_returns_default_fields(self):
        resolver = FieldSelectionResolver(allowed_fields=["id", "name", "desc"])
        result = resolver.resolve(None, default_fields=["id", "name"])
        self.assertEqual(["id", "name"], result)

    def test_resolve_empty_list_raises_error(self):
        resolver = FieldSelectionResolver(allowed_fields=["id", "name"])
        with self.assertRaises(EmptyFieldSelectionError):
            resolver.resolve([])

    def test_resolve_whitespace_only_raises_error(self):
        resolver = FieldSelectionResolver(allowed_fields=["id", "name"])
        with self.assertRaises(EmptyFieldSelectionError):
            resolver.resolve(["  ", "\t"])

    def test_resolve_all_token_returns_all_allowed_fields(self):
        resolver = FieldSelectionResolver(
            allowed_fields=["id", "name"], all_token="everything"
        )
        result = resolver.resolve(["everything"])
        self.assertEqual(["id", "name"], result)

    def test_resolve_all_token_case_insensitive(self):
        resolver = FieldSelectionResolver(
            allowed_fields=["id", "name"], all_token="ALL"
        )
        result = resolver.resolve(["all"])
        self.assertEqual(["id", "name"], result)

    def test_resolve_valid_fields(self):
        resolver = FieldSelectionResolver(allowed_fields=["id", "name", "desc"])
        result = resolver.resolve(["ID", "name"])
        self.assertEqual(["ID", "name"], result)

    def test_resolve_unknown_field_raises_error(self):
        resolver = FieldSelectionResolver(allowed_fields=["id", "name"])
        with self.assertRaises(UnknownFieldError) as cm:
            resolver.resolve(["id", "invalid"])
        self.assertEqual("invalid", cm.exception.unknown)
        self.assertEqual(["id", "name"], cm.exception.allowed)

    def test_resolve_unknown_field_with_context(self):
        resolver = FieldSelectionResolver(allowed_fields=["id"])
        with self.assertRaises(UnknownFieldError) as cm:
            resolver.resolve(["invalid"], context="test-context")
        self.assertEqual("test-context", cm.exception.context)
        self.assertIn("(test-context)", str(cm.exception))

    def test_duplicates_removed_by_default(self):
        resolver = FieldSelectionResolver(allowed_fields=["id", "name"])
        result = resolver.resolve(["id", "ID", "name", "id"])
        self.assertEqual(["id", "name"], result)

    def test_duplicates_allowed(self):
        resolver = FieldSelectionResolver(
            allowed_fields=["id", "name"], allow_duplicates=True
        )
        result = resolver.resolve(["id", "ID", "name", "id"])
        self.assertEqual(["id", "ID", "name", "id"], result)

    def test_alias_substitution(self):
        aliases = {"n": "name", "i": "id"}
        resolver = FieldSelectionResolver(allowed_fields=["id", "name"])
        result = resolver.resolve(["i", "N"], aliases=aliases)
        self.assertEqual(["id", "name"], result)

    def test_alias_case_insensitivity(self):
        aliases = {"NAME_ALIAS": "name"}
        resolver = FieldSelectionResolver(allowed_fields=["name"])
        result = resolver.resolve(["name_alias"], aliases=aliases)
        self.assertEqual(["name"], result)

    def test_resolve_empty_allowed_fields(self):
        resolver = FieldSelectionResolver(allowed_fields=[])
        result = resolver.resolve(["any"])
        self.assertEqual([], result)

    def test_resolve_none_with_empty_allowed_fields(self):
        resolver = FieldSelectionResolver(allowed_fields=[])
        result = resolver.resolve(None)
        self.assertEqual([], result)

    def test_apply_alias_no_aliases(self):
        resolver = FieldSelectionResolver(allowed_fields=["id"])
        self.assertEqual("id", resolver.apply_alias("id", None))

    def test_apply_alias_no_match(self):
        resolver = FieldSelectionResolver(allowed_fields=["id"])
        self.assertEqual("other", resolver.apply_alias("other", {"a": "id"}))

    def test_validate_preserves_case_of_allowed_fields_matches(self):
        resolver = FieldSelectionResolver(allowed_fields=["id", "name"])
        result = resolver.validate(["ID", "Name"], None)
        self.assertEqual(["ID", "Name"], result)
