import unittest

from peewee import DoesNotExist

from cmdbox.database import init_database, db
from cmdbox.exceptions import ValidationError, NameConflictError, UnknownNameError
from cmdbox.models import Tag
from cmdbox.repositories.tag_repository import TagRepository


class TestTagRepository(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        init_database(testing=True)
        db.connect()
        db.bind([Tag])
        db.create_tables([Tag])

    @classmethod
    def tearDownClass(cls):
        # Close database connection after all tests
        db.drop_tables([Tag])
        db.close()

    def setUp(self):
        Tag.delete().execute()
        self.repo = TagRepository()

    def _create_tag_group(self):
        self.tag_one = Tag.create(
            name="test1", description="test_description_e Antelope Bee"
        )
        self.tag_two = Tag.create(
            name="test2", description="test_description_d Zebra Bee"
        )
        self.tag_three = Tag.create(
            name="test3", description="test_description_c Kangaroo Bee"
        )
        self.tag_four = Tag.create(
            name="test4", description="test_description_b Bee Bee Bee"
        )
        self.tag_five = Tag.create(
            name="test5", description="test_description_a Bee Goldfish"
        )

    def test_create_tag_works(self):
        tag = self.repo.create(name="test", description="test_description")
        q_tag = Tag.get(Tag.name == "test")
        self.assertTrue(isinstance(tag, Tag))
        self.assertEqual(tag, q_tag)

    def test_create_no_name_supplied(self):
        with self.assertRaises(ValidationError):
            self.repo.create(name=None, description="test_description")

    def test_create_blank_name_supplied(self):
        with self.assertRaises(ValidationError):
            self.repo.create(name="", description="test_description")

    def test_create_duplicate_tag_name_not_allowed(self):
        Tag.create(name="test", description="test_description")
        with self.assertRaises(NameConflictError):
            self.repo.create(name="test", description="test_description2")

    def test_create_no_description_supplied(self):
        self.repo.create(name="test", description=None)

    def test_create_blank_description_supplied(self):
        self.repo.create(name="test", description="")

    def test_duplicate_tag_description_is_allowed(self):
        self.repo.create(name="test", description="test_description")
        self.repo.create(name="test2", description="test_description")

    def test_all_symbols_are_allowed_in_tag_name(self):
        self.repo.create(
            name="test-1!@#$%^&*()_+-=[]\;/.,<>?:{}|", description="test_description"
        )

    def test_unicode_characters_are_allowed_in_tag_name(self):
        self.repo.create(name="git-✨", description="test_description")

    def test_all_symbols_are_allowed_in_tag_description(self):
        self.repo.create(name="test", description="test-1!@#$%^&*()_+-=[]\;/.,<>?:{}|")

    def test_unicode_characters_are_allowed_in_tag_description(self):
        self.repo.create(name="test", description="git-✨")

    def test_whitespace_in_middle_of_tag_name_is_not_allowed(self):
        with self.assertRaises(ValidationError):
            self.repo.create(name="test test", description="test_description")

    def test_whitespace_at_beginning_and_end_of_tag_name_is_stripped(self):
        self.repo.create(name=" test", description="test_description")
        self.assertEqual("test", Tag.get(Tag.name == "test").name)

    def test_get_tag_by_name(self):
        tag = Tag.create(name="test", description="test_description")
        tag = self.repo.get_by_name("test")
        self.assertEqual(tag, tag)

    def test_get_unknown_tag_raises_exception(self):
        with self.assertRaises(UnknownNameError):
            self.repo.get_by_name("invalid_name")

    def test_tag_name_capitalisation_does_not_matter(self):
        tag = Tag.create(name="test", description="test_description")
        tag = self.repo.get_by_name("TEST")
        self.assertEqual(tag, tag)

    def test_get_by_blank_field_raises_exception(self):
        with self.assertRaises(UnknownNameError):
            self.repo.get_by_name("")

    def test_get_by_other_fields_does_not_work(self):
        Tag.create(name="test", description="test_description")
        with self.assertRaises(UnknownNameError):
            self.repo.get_by_name("test_description")

    def test_get_by_all_symbols_is_allowed(self):
        Tag.create(
            name="test-1!@#$%^&*()_+-=[]\;/.,<>?:{}|", description="test_description"
        )
        self.repo.get_by_name("test-1!@#$%^&*()_+-=[]\;/.,<>?:{}|")

    def test_get_by_unicode_characters_is_allowed(self):
        Tag.create(name="git-✨", description="test_description")
        self.repo.get_by_name(name="git-✨")

    def test_update_tag_name_works(self):
        Tag.create(name="test", description="test_description")
        self.repo.update(tag_name="test", name="new_name")
        Tag.get(Tag.name == "new_name")

    def test_update_tag_description_works(self):
        tag = Tag.create(name="test", description="test_description")
        tag = self.repo.update(tag_name="test", description="new_description")
        self.assertEqual("new_description", Tag.get(Tag.name == "test").description)
        self.assertEqual(tag, tag)

    def test_update_tag_with_blank_name_raises_exception(self):
        Tag.create(name="test", description="test_description")
        with self.assertRaises(ValidationError):
            self.repo.update(tag_name="test", name="")

    def test_update_tag_with_null_name_does_nothing(self):
        tag = Tag.create(name="test", description="test_description")
        self.repo.update(tag_name="test", name=None)
        self.assertEqual(tag, Tag.get(Tag.name == "test"))

    def test_update_tag_with_duplicate_name_raises_exception(self):
        Tag.create(name="test", description="test_description")
        Tag.create(name="test2", description="test_description2")
        with self.assertRaises(NameConflictError):
            self.repo.update(tag_name="test2", name="test")

    def test_update_tag_description_to_an_existing_tag_is_allowed(self):
        """Multiple tags can have the same description."""
        Tag.create(name="test", description="test_description")
        Tag.create(name="test2", description="test_description2")
        self.repo.update(tag_name="test2", description="test")

    def test_update_with_no_fields_throws_exception(self):
        Tag.create(name="test", description="test_description")
        with self.assertRaises(ValueError):
            self.repo.update(tag_name="test")

    def test_update_name_with_white_space_in_middle_of_name_is_not_allowed(self):
        Tag.create(name="test", description="test_description")
        with self.assertRaises(ValidationError):
            self.repo.update(tag_name="test", name="test2 test")

    def test_update_name_with_white_space_at_beginning_and_end_of_name_is_stripped(
        self,
    ):
        Tag.create(name="test", description="test_description")
        self.repo.update(tag_name="test", name=" test2 ")
        self.assertEqual("test2", Tag.get(Tag.name == "test2").name)

    def test_list_all_works(self):
        self._create_tag_group()
        tags = self.repo.list_all()
        self.assertEqual(5, len(tags))
        self.assertTrue(self.tag_one in tags)
        self.assertTrue(self.tag_two in tags)
        self.assertTrue(self.tag_three in tags)
        self.assertTrue(self.tag_four in tags)
        self.assertTrue(self.tag_five in tags)

    def test_list_all_returns_empty_list_if_no_tags(self):
        tags = self.repo.list_all()
        self.assertEqual([], tags)

    def test_list_all_ordered_by_name_by_default(self):
        self._create_tag_group()
        tags = self.repo.list_all()
        self.assertEqual(self.tag_one, tags[0])
        self.assertEqual(self.tag_two, tags[1])
        self.assertEqual(self.tag_three, tags[2])
        self.assertEqual(self.tag_four, tags[3])
        self.assertEqual(self.tag_five, tags[4])

    def test_list_all_ordered_by_description(self):
        self._create_tag_group()
        tags = self.repo.list_all(order_by="description")
        self.assertEqual(self.tag_five, tags[0])
        self.assertEqual(self.tag_four, tags[1])
        self.assertEqual(self.tag_three, tags[2])
        self.assertEqual(self.tag_two, tags[3])
        self.assertEqual(self.tag_one, tags[4])

    def test_list_by_name_desc(self):
        self._create_tag_group()
        tags = self.repo.list_all(order_by="-name")
        self.assertEqual(self.tag_five, tags[0])
        self.assertEqual(self.tag_four, tags[1])
        self.assertEqual(self.tag_three, tags[2])
        self.assertEqual(self.tag_two, tags[3])
        self.assertEqual(self.tag_one, tags[4])

    def test_list_all_limit_functions_correctly(self):
        self._create_tag_group()
        tags = self.repo.list_all(limit=2)
        self.assertEqual(2, len(tags))

    def test_limit_of_zero_returns_empty_list(self):
        self._create_tag_group()
        tags = self.repo.list_all(limit=0)
        self.assertEqual(0, len(tags))

    def test_limit_of_none_does_not_limit(self):
        self._create_tag_group()
        tags = self.repo.list_all(limit=None)
        self.assertEqual(5, len(tags))

    def test_search_empty_term_returns_empty_list(self):
        self._create_tag_group()
        tags = self.repo.search("")
        self.assertEqual(0, len(tags))

    def test_search_returns_empty_list_if_no_tags_match(self):
        tags = self.repo.search("anyname")
        self.assertEqual([], tags)

    def test_search_returns_matching_tags_on_default_fields(self):
        self._create_tag_group()
        tags = self.repo.search("test")
        self.assertEqual(5, len(tags))

        tags = self.repo.search("test2")
        self.assertEqual(self.tag_two, tags[0])

    def test_search_returns_matching_tags_on_specific_fields(self):
        self._create_tag_group()
        tags = self.repo.search("zebra", fields="description")
        self.assertEqual(1, len(tags))
        self.assertEqual(self.tag_two, tags[0])

    def test_search_returns_matching_tags_on_multiple_fields(self):
        self._create_tag_group()
        tags = self.repo.search("zebra", fields=["name", "description"])
        self.assertEqual(1, len(tags))
        self.assertEqual(self.tag_two, tags[0])

    def test_search_on_invalid_field_raises_exception(self):
        self._create_tag_group()
        with self.assertRaises(ValueError):
            self.repo.search("zebra", fields="invalid_field")

    def test_search_empty_string_returns_no_results(self):
        self._create_tag_group()
        tags = self.repo.search("")
        self.assertEqual(0, len(tags))

    def test_search_is_case_insensitive(self):
        self._create_tag_group()
        tags = self.repo.search("TEST")
        self.assertEqual(5, len(tags))

    def test_search_is_ordered_by_most_relevant(self):
        self._create_tag_group()
        tags = self.repo.search("bee", fields="description")
        self.assertEqual(self.tag_four, tags[0])
        self.assertEqual(self.tag_five, tags[1])
        self.assertEqual(self.tag_two, tags[2])
        self.assertEqual(self.tag_one, tags[3])
        self.assertEqual(self.tag_three, tags[4])

    def test_delete_tag_works(self):
        self._create_tag_group()
        self.assertEqual(5, Tag.select().count())
        self.repo.delete("test2")
        self.assertEqual(4, Tag.select().count())
        self.repo.delete("test5")
        self.assertEqual(3, Tag.select().count())
        with self.assertRaises(DoesNotExist):
            Tag.get(Tag.id == self.tag_two)
            Tag.get(Tag.id == self.tag_five)
