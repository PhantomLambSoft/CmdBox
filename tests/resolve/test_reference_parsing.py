import unittest

from cmdbox.resolve.reference_parsing import (
    extract_references,
    parse_kind_and_key,
    read_angle_token,
)
from cmdbox.resolve.type_defs import RefKind


class TestReferenceParsing(unittest.TestCase):

    def test_read_angle_token_reads_simple_token(self):
        token_inner, raw_token, next_i = read_angle_token("<var:name> trailing", 0)

        self.assertEqual("var:name", token_inner)
        self.assertEqual("var:name", raw_token)
        self.assertEqual(10, next_i)

    def test_read_angle_token_respects_start_index(self):
        token_inner, raw_token, next_i = read_angle_token("x<cmd:build>z", 1)

        self.assertEqual("cmd:build", token_inner)
        self.assertEqual("cmd:build", raw_token)
        self.assertEqual(12, next_i)

    def test_read_angle_token_interprets_supported_escapes_and_strips(self):
        token_inner, raw_token, next_i = read_angle_token("<  a\\<b\\>c\\\\d  >", 0)

        self.assertEqual("a<b>c\\d", token_inner)
        self.assertEqual("  a\\<b\\>c\\\\d  ", raw_token)
        self.assertEqual(16, next_i)

    def test_read_angle_token_keeps_non_special_backslash(self):
        token_inner, raw_token, next_i = read_angle_token(r"<a\nb>", 0)

        self.assertEqual(r"a\nb", token_inner)
        self.assertEqual(r"a\nb", raw_token)
        self.assertEqual(6, next_i)

    def test_read_angle_token_returns_none_when_closing_bracket_missing(self):
        token_inner, raw_token, next_i = read_angle_token("<var:name", 0)

        self.assertIsNone(token_inner)
        self.assertEqual("<var:name", raw_token)
        self.assertEqual(9, next_i)

    def test_parse_kind_and_key_defaults_to_variable_without_separator(self):
        result = parse_kind_and_key("  user.name  ")

        self.assertEqual((RefKind.VARIABLE, "  user.name  "), result)

    def test_parse_kind_and_key_parses_command_prefix(self):
        result = parse_kind_and_key("cmd:  build:image  ")

        self.assertEqual((RefKind.COMMAND, "build:image"), result)

    def test_parse_kind_and_key_parses_variable_prefix(self):
        result = parse_kind_and_key("var:  app.port  ")

        self.assertEqual((RefKind.VARIABLE, "app.port"), result)

    def test_parse_kind_and_key_treats_unknown_prefix_as_plain_variable_reference(self):
        result = parse_kind_and_key("env: HOME")

        self.assertEqual((RefKind.VARIABLE, "env: HOME"), result)

    def test_parse_kind_and_key_unknown_when_prefix_missing_before_colon(self):
        result = parse_kind_and_key(":value")

        self.assertEqual((RefKind.VARIABLE, ":value"), result)

    def test_extract_references_returns_empty_for_plain_text(self):
        result = extract_references("hello world")

        self.assertEqual([], result)

    def test_extract_references_extracts_mixed_reference_kinds_in_order(self):
        template = "run <cmd:build> using <var:tag> and <name>"

        result = extract_references(template)

        self.assertEqual(
            [
                (RefKind.COMMAND, "build"),
                (RefKind.VARIABLE, "tag"),
                (RefKind.VARIABLE, "name"),
            ],
            result,
        )

    def test_extract_references_ignores_escaped_opening_bracket(self):
        result = extract_references(r"\<cmd:build>")

        self.assertEqual([], result)

    def test_extract_references_ignores_unclosed_reference(self):
        result = extract_references("prefix <cmd:build")

        self.assertEqual([], result)

    def test_extract_references_skips_empty_or_whitespace_tokens(self):
        result = extract_references("<> <   > <var:ok>")

        self.assertEqual([(RefKind.VARIABLE, "ok")], result)

    def test_extract_references_allows_escaped_delimiters_inside_token(self):
        result = extract_references(r"<var: a\<b\>c >")

        self.assertEqual([(RefKind.VARIABLE, "a<b>c")], result)

    def test_extract_references_treats_unknown_prefix_as_variable_with_original_token(
        self,
    ):
        result = extract_references("<repo: main>")

        self.assertEqual([(RefKind.VARIABLE, "repo: main")], result)
