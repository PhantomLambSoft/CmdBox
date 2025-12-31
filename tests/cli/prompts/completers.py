import unittest
from prompt_toolkit.document import Document
from prompt_toolkit.completion import CompleteEvent
from cmdbox.cli.prompts.completers import (
    TagCompleter,
    TagCompleterConfig,
    _normalize_tag,
    _split_csv_like,
    _simple_fuzzy_score,
)


class TestHelperFunctions(unittest.TestCase):

    def test_normalize_tag(self):
        self.assertEqual(_normalize_tag("  Tag1  "), "tag1")
        self.assertEqual(_normalize_tag("TAG2"), "tag2")
        self.assertEqual(_normalize_tag("tag3"), "tag3")

    def test_split_csv_like(self):
        # No comma
        self.assertEqual(_split_csv_like("tag1"), ("", "tag1", 0))
        # Single comma, no space
        self.assertEqual(_split_csv_like("tag1,tag2"), ("tag1,", "tag2", 5))
        # Single comma with space
        self.assertEqual(_split_csv_like("tag1, tag2"), ("tag1, ", "tag2", 6))
        # Multiple commas
        self.assertEqual(
            _split_csv_like("tag1, tag2, tag3"), ("tag1, tag2, ", "tag3", 12)
        )
        # Trailing comma
        self.assertEqual(_split_csv_like("tag1, "), ("tag1, ", "", 6))

    def test_simple_fuzzy_score(self):
        # Exact match at start
        self.assertEqual(_simple_fuzzy_score("abc", "abcdef"), 100)
        # Substring match
        self.assertEqual(_simple_fuzzy_score("bcd", "abcdef"), 75)
        # All characters present in order
        self.assertEqual(_simple_fuzzy_score("adf", "abcdef"), 55)
        # No match
        self.assertEqual(_simple_fuzzy_score("xyz", "abcdef"), 0)
        # Empty needle
        self.assertEqual(_simple_fuzzy_score("", "abcdef"), 1)


class TestTagCompleter(unittest.TestCase):

    def test_get_completions_basic(self):
        get_tags = lambda: ["python", "rust", "cpp"]
        completer = TagCompleter(get_tags)

        # Match 'py'
        doc = Document("py")
        completions = list(completer.get_completions(doc, CompleteEvent()))
        self.assertEqual(len(completions), 1)
        self.assertEqual(completions[0].text, "python")
        self.assertEqual(completions[0].start_position, -2)

    def test_get_completions_csv(self):
        get_tags = lambda: ["python", "rust", "cpp", "javascript"]
        completer = TagCompleter(get_tags)

        # Already has 'python', typing 'r'
        doc = Document("python, r")
        completions = list(completer.get_completions(doc, CompleteEvent()))
        # Both 'rust' (score 100) and 'javascript' (score 75) should be suggested
        self.assertEqual(len(completions), 2)
        self.assertEqual(completions[0].text, "rust")
        self.assertEqual(completions[1].text, "javascript")
        self.assertEqual(completions[0].start_position, -1)

    def test_get_completions_exclude_already_used(self):
        get_tags = lambda: ["python", "rust"]
        completer = TagCompleter(get_tags)

        # Already has 'python', typing 'p' - should not suggest 'python' again
        doc = Document("python, p")
        completions = list(completer.get_completions(doc, CompleteEvent()))
        self.assertEqual(len(completions), 0)

    def test_case_sensitivity(self):
        get_tags = lambda: ["Python", "Rust"]

        # Case insensitive (default)
        completer_ci = TagCompleter(get_tags, TagCompleterConfig(case_insensitive=True))
        doc = Document("py")
        completions = list(completer_ci.get_completions(doc, CompleteEvent()))
        self.assertEqual(len(completions), 1)
        self.assertEqual(completions[0].text, "Python")

        # Case sensitive
        completer_cs = TagCompleter(
            get_tags, TagCompleterConfig(case_insensitive=False)
        )
        doc = Document("py")
        completions = list(completer_cs.get_completions(doc, CompleteEvent()))
        self.assertEqual(len(completions), 0)

        doc2 = Document("Py")
        completions2 = list(completer_cs.get_completions(doc2, CompleteEvent()))
        self.assertEqual(len(completions2), 1)
        self.assertEqual(completions2[0].text, "Python")

    def test_max_results(self):
        get_tags = lambda: [f"tag{i}" for i in range(20)]
        completer = TagCompleter(get_tags, TagCompleterConfig(max_results=5))

        doc = Document("t")
        completions = list(completer.get_completions(doc, CompleteEvent()))
        self.assertEqual(len(completions), 5)

    def test_min_score(self):
        get_tags = lambda: ["abc", "def"]
        # abc will score 100 for 'a', def will score 0 (or low)
        completer = TagCompleter(get_tags, TagCompleterConfig(min_score=50))

        doc = Document("a")
        completions = list(completer.get_completions(doc, CompleteEvent()))
        self.assertEqual(len(completions), 1)
        self.assertEqual(completions[0].text, "abc")
