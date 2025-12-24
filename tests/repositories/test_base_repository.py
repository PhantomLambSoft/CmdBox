import unittest

from cmdbox.models import Command
from cmdbox.repositories.base_repository import BaseRepository


class TestBaseRepository(unittest.TestCase):

    def setUp(self):
        self.repo = BaseRepository()
        self.repo.model = Command

    def test_resolve_ordering(self):
        ordering = self.repo._resolve_ordering("alias")
        self.assertEqual(Command.alias.asc(), ordering[0])

    def test_resolve_ordering_desc(self):
        ordering = self.repo._resolve_ordering("-alias")
        self.assertEqual(Command.alias.desc(), ordering[0])

    def test_resolve_ordering_multiple_fields_string(self):
        ordering = self.repo._resolve_ordering("alias, template")
        self.assertEqual(Command.alias.asc(), ordering[0])
        self.assertEqual(Command.template.asc(), ordering[1])

    def test_resolve_ordering_multiple_fields_list(self):
        ordering = self.repo._resolve_ordering(["alias", "template"])
        self.assertEqual(Command.alias.asc(), ordering[0])
        self.assertEqual(Command.template.asc(), ordering[1])

    def test_resolve_ordering_multiple_fields_mixed(self):
        ordering = self.repo._resolve_ordering(["alias", "template", "-description"])
        self.assertEqual(Command.alias.asc(), ordering[0])
        self.assertEqual(Command.template.asc(), ordering[1])
        self.assertEqual(Command.description.desc(), ordering[2])

    def test_empty_ordering_raises_exception(self):
        with self.assertRaises(ValueError):
            self.repo._resolve_ordering("")

    def test_invalid_ordering_raises_exception(self):
        with self.assertRaises(ValueError):
            self.repo._resolve_ordering("invalid_field")

    def test_get_sequence_turns_string_into_list(self):
        sequence = self.repo._get_sequence("alias,template")
        self.assertEqual(["alias", "template"], sequence)

    def test_get_sequence_turns_tuple_into_list(self):
        sequence = self.repo._get_sequence(("alias", "template"))
        self.assertEqual(["alias", "template"], sequence)

    def test_get_sequence_keeps_list_as_list(self):
        sequence = self.repo._get_sequence(["alias", "template"])
        self.assertEqual(["alias", "template"], sequence)

    def test_get_sequence_strips_whitespace_from_string(self):
        sequence = self.repo._get_sequence(" alias , template   ")
        self.assertEqual(["alias", "template"], sequence)

    def test_get_sequence_strips_whitespace_from_list(self):
        sequence = self.repo._get_sequence([" alias ", " template   "])
        self.assertEqual(["alias", "template"], sequence)
