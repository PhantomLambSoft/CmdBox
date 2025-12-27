import unittest

from mypy.checkpattern import self_match_type_names
from peewee import DoesNotExist

from cmdbox.database import init_database, db
from cmdbox.repositories.errors import (
    ValidationError,
    NameConflictError,
    UnknownTagError,
    UnknownNameError,
    UpdateError,
    TagAttachError,
    TagDetachError,
)
from cmdbox.models import Variable, Tag, VariableTag
from cmdbox.repositories.variable_repository import VariableRepository


class TestVariableRepository(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        init_database(testing=True)
        db.connect()
        db.bind([Variable, Tag, VariableTag])
        db.create_tables([Variable, Tag, VariableTag])

    @classmethod
    def tearDownClass(cls):
        # Close database connection after all tests
        db.drop_tables([Variable, Tag, VariableTag])
        db.close()

    def setUp(self):
        Variable.delete().execute()
        Tag.delete().execute()
        VariableTag.delete().execute()
        self.repo = VariableRepository()

    def _create_variable_group(self):
        self.var_one = Variable.create(name="test1", value="test_value_e Antelope Bee")
        self.var_two = Variable.create(name="test2", value="test_value_d Zebra Bee")
        self.var_three = Variable.create(
            name="test3", value="test_value_c Kangaroo Bee"
        )
        self.var_four = Variable.create(name="test4", value="test_value_b Bee Bee Bee")
        self.var_five = Variable.create(name="test5", value="test_value_a Bee Goldfish")

    def _tag_variable_group(self):
        self.tag_one = Tag.create(name="tag_one")
        self.tag_two = Tag.create(name="tag_two")
        self.tag_three = Tag.create(name="tag_three")
        self.var_tag_one = VariableTag.create(variable=self.var_one, tag=self.tag_one)
        self.var_tag_two = VariableTag.create(variable=self.var_two, tag=self.tag_one)
        self.var_tag_three = VariableTag.create(variable=self.var_two, tag=self.tag_two)
        self.var_tag_four = VariableTag.create(
            variable=self.var_two, tag=self.tag_three
        )
        self.var_tag_five = VariableTag.create(
            variable=self.var_three, tag=self.tag_three
        )

    # =================================================================================
    # SECTION: CREATE TESTS
    # =================================================================================

    def test_create_variable_works(self):
        var = self.repo.create(name="test", value="test_value")
        q_var = Variable.get(Variable.name == "test")
        self.assertTrue(isinstance(var, Variable))
        self.assertEqual(var, q_var)

    def test_create_no_name_supplied(self):
        with self.assertRaises(ValidationError):
            self.repo.create(name=None, value="test_value")

    def test_create_blank_name_supplied(self):
        with self.assertRaises(ValidationError):
            self.repo.create(name="", value="test_value")

    def test_create_duplicate_variable_name_not_allowed(self):
        Variable.create(name="test", value="test_value")
        with self.assertRaises(NameConflictError):
            self.repo.create(name="test", value="test_value2")

    def test_create_no_value_supplied(self):
        with self.assertRaises(ValidationError):
            self.repo.create(name="test", value=None)

    def test_create_blank_value_supplied(self):
        self.repo.create(name="test", value="")

    def test_duplicate_variable_value_is_allowed(self):
        self.repo.create(name="test", value="test_value")
        self.repo.create(name="test2", value="test_value")

    def test_all_symbols_are_allowed_in_variable_name(self):
        self.repo.create(name="test-1!@#$%^&*()_+-=[]\;/.,<>?:{}|", value="test_value")

    def test_unicode_characters_are_allowed_in_variable_name(self):
        self.repo.create(name="git-✨", value="test_value")

    def test_all_symbols_are_allowed_in_variable_value(self):
        self.repo.create(name="test", value="test-1!@#$%^&*()_+-=[]\;/.,<>?:{}|")

    def test_unicode_characters_are_allowed_in_variable_value(self):
        self.repo.create(name="test", value="git-✨")

    def test_whitespace_in_middle_of_variable_name_is_not_allowed(self):
        with self.assertRaises(ValidationError):
            self.repo.create(name="test test", value="test_value")

    def test_whitespace_at_beginning_and_end_of_variable_name_is_stripped(self):
        self.repo.create(name=" test", value="test_value")
        self.assertEqual("test", Variable.get(Variable.name == "test").name)

    def test_get_unknown_variable_raises_exception(self):
        with self.assertRaises(UnknownNameError):
            self.repo.get_by_name("invalid_name")

    def test_variable_name_capitalisation_does_not_matter(self):
        variable = Variable.create(name="test", value="test_value")
        var = self.repo.get_by_name("TEST")
        self.assertEqual(variable, var)

    # =================================================================================
    # SECTION: GET TESTS
    # =================================================================================

    def test_get_variable_by_name(self):
        variable = Variable.create(name="test", value="test_value")
        var = self.repo.get_by_name("test")
        self.assertEqual(variable, var)

    def test_get_by_blank_field_raises_exception(self):
        with self.assertRaises(UnknownNameError):
            self.repo.get_by_name("")

    def test_get_by_other_fields_does_not_work(self):
        Variable.create(name="test", value="test_value")
        with self.assertRaises(UnknownNameError):
            self.repo.get_by_name("test_value")

    def test_get_by_all_symbols_is_allowed(self):
        Variable.create(name="test-1!@#$%^&*()_+-=[]\;/.,<>?:{}|", value="test_value")
        self.repo.get_by_name("test-1!@#$%^&*()_+-=[]\;/.,<>?:{}|")

    def test_get_by_unicode_characters_is_allowed(self):
        Variable.create(name="git-✨", value="test_value")
        self.repo.get_by_name(name="git-✨")

    # =================================================================================
    # SECTION: UPDATE TESTS
    # =================================================================================

    def test_update_variable_name_works(self):
        var = Variable.create(name="test", value="test_value")
        self.repo.update(variable=var, name="new_name")
        Variable.get(Variable.name == "new_name")

    def test_update_variable_value_works(self):
        variable = Variable.create(name="test", value="test_value")
        var = self.repo.update(variable=variable, value="new_value")
        self.assertEqual("new_value", Variable.get(Variable.name == "test").value)
        self.assertEqual(variable, var)

    def test_update_variable_with_blank_name_raises_exception(self):
        var = Variable.create(name="test", value="test_value")
        with self.assertRaises(ValidationError):
            self.repo.update(variable=var, name="")

    def test_update_variable_with_null_name_does_nothing(self):
        var = Variable.create(name="test", value="test_value")
        self.repo.update(variable=var, name=None)
        self.assertEqual(var, Variable.get(Variable.name == "test"))

    def test_update_variable_with_duplicate_name_raises_exception(self):
        Variable.create(name="test", value="test_value")
        var = Variable.create(name="test2", value="test_value2")
        with self.assertRaises(NameConflictError):
            self.repo.update(variable=var, name="test")

    def test_update_variable_value_to_an_existing_variable_is_allowed(self):
        """Multiple variables can have the same value."""
        Variable.create(name="test", value="test_value")
        var = Variable.create(name="test2", value="test_value2")
        self.repo.update(variable=var, value="test")

    def test_update_with_no_fields_throws_exception(self):
        var = Variable.create(name="test", value="test_value")
        with self.assertRaises(UpdateError):
            self.repo.update(variable=var)

    def test_update_name_with_white_space_in_middle_of_name_is_not_allowed(self):
        var = Variable.create(name="test", value="test_value")
        with self.assertRaises(ValidationError):
            self.repo.update(variable=var, name="test2 test")

    def test_update_name_with_white_space_at_beginning_and_end_of_name_is_stripped(
        self,
    ):
        var = Variable.create(name="test", value="test_value")
        self.repo.update(variable=var, name=" test2 ")
        self.assertEqual("test2", Variable.get(Variable.name == "test2").name)

    # =================================================================================
    # SECTION: LIST TESTS
    # =================================================================================

    def test_list_all_works(self):
        self._create_variable_group()
        vars = self.repo.list_all()
        self.assertEqual(5, len(vars))
        self.assertTrue(self.var_one in vars)
        self.assertTrue(self.var_two in vars)
        self.assertTrue(self.var_three in vars)
        self.assertTrue(self.var_four in vars)
        self.assertTrue(self.var_five in vars)

    def test_list_all_returns_empty_list_if_no_variables(self):
        vars = self.repo.list_all()
        self.assertEqual([], vars)

    def test_list_all_ordered_by_name_by_default(self):
        self._create_variable_group()
        vars = self.repo.list_all()
        self.assertEqual(self.var_one, vars[0])
        self.assertEqual(self.var_two, vars[1])
        self.assertEqual(self.var_three, vars[2])
        self.assertEqual(self.var_four, vars[3])
        self.assertEqual(self.var_five, vars[4])

    def test_list_all_ordered_by_value(self):
        self._create_variable_group()
        vars = self.repo.list_all(order_by="value")
        self.assertEqual(self.var_five, vars[0])
        self.assertEqual(self.var_four, vars[1])
        self.assertEqual(self.var_three, vars[2])
        self.assertEqual(self.var_two, vars[3])
        self.assertEqual(self.var_one, vars[4])

    def test_list_by_name_desc(self):
        self._create_variable_group()
        vars = self.repo.list_all(order_by="-name")
        self.assertEqual(self.var_five, vars[0])
        self.assertEqual(self.var_four, vars[1])
        self.assertEqual(self.var_three, vars[2])
        self.assertEqual(self.var_two, vars[3])
        self.assertEqual(self.var_one, vars[4])

    def test_list_all_limit_functions_correctly(self):
        self._create_variable_group()
        vars = self.repo.list_all(limit=2)
        self.assertEqual(2, len(vars))

    def test_limit_of_zero_returns_empty_list(self):
        self._create_variable_group()
        vars = self.repo.list_all(limit=0)
        self.assertEqual(0, len(vars))

    def test_limit_of_none_does_not_limit(self):
        self._create_variable_group()
        vars = self.repo.list_all(limit=None)
        self.assertEqual(5, len(vars))

    # =================================================================================
    # SECTION: LIST BY TAG TESTS
    # =================================================================================

    def test_list_by_tag(self):
        self._create_variable_group()
        self._tag_variable_group()
        vars = self.repo.list_by_tag([self.tag_one])
        self.assertEqual(2, len(vars))
        self.assertTrue(self.var_one in vars)
        self.assertTrue(self.var_two in vars)

    def test_list_by_multiple_tags(self):
        self._create_variable_group()
        self._tag_variable_group()
        vars = self.repo.list_by_tag([self.tag_one, self.tag_three])
        self.assertEqual(3, len(vars))
        self.assertTrue(self.var_one in vars)
        self.assertTrue(self.var_two in vars)
        self.assertTrue(self.var_three in vars)

    def test_list_by_null_tag_returns_empty_list(self):
        self._create_variable_group()
        self._tag_variable_group()
        vars = self.repo.list_by_tag([None])
        self.assertEqual([], vars)

    def test_list_by_tag_order_by_template_changes_order(self):
        self._create_variable_group()
        self._tag_variable_group()
        vars = self.repo.list_by_tag([self.tag_one], order_by="value")
        self.assertEqual(2, len(vars))
        self.assertEqual(self.var_two, vars[0])
        self.assertEqual(self.var_one, vars[1])

    def test_list_by_tag_limits_apply(self):
        self._create_variable_group()
        self._tag_variable_group()
        vars = self.repo.list_by_tag([self.tag_one], limit=1)
        self.assertEqual(1, len(vars))
        self.assertTrue(self.var_one in vars)

    # =================================================================================
    # SECTION: SEARCH TESTS
    # =================================================================================

    def test_search_empty_term_returns_empty_list(self):
        self._create_variable_group()
        vars = self.repo.search("")
        self.assertEqual(0, len(vars))

    def test_search_returns_empty_list_if_no_variables_match(self):
        vars = self.repo.search("anyname")
        self.assertEqual([], vars)

    def test_search_returns_matching_variables_on_default_fields(self):
        self._create_variable_group()
        vars = self.repo.search("test")
        self.assertEqual(5, len(vars))

        vars = self.repo.search("test2")
        self.assertEqual(self.var_two, vars[0])

    def test_search_returns_matching_variables_on_specific_fields(self):
        self._create_variable_group()
        vars = self.repo.search("zebra", fields="value")
        self.assertEqual(1, len(vars))
        self.assertEqual(self.var_two, vars[0])

    def test_search_returns_matching_variables_on_multiple_fields(self):
        self._create_variable_group()
        vars = self.repo.search("zebra", fields=["name", "value"])
        self.assertEqual(1, len(vars))
        self.assertEqual(self.var_two, vars[0])

    def test_search_on_invalid_field_raises_exception(self):
        self._create_variable_group()
        with self.assertRaises(ValueError):
            self.repo.search("zebra", fields="invalid_field")

    def test_search_empty_string_returns_no_results(self):
        self._create_variable_group()
        vars = self.repo.search("")
        self.assertEqual(0, len(vars))

    def test_search_is_case_insensitive(self):
        self._create_variable_group()
        vars = self.repo.search("TEST")
        self.assertEqual(5, len(vars))

    def test_search_is_ordered_by_most_relevant(self):
        self._create_variable_group()
        vars = self.repo.search("bee", fields="value")
        self.assertEqual(self.var_four, vars[0])
        self.assertEqual(self.var_five, vars[1])
        self.assertEqual(self.var_two, vars[2])
        self.assertEqual(self.var_one, vars[3])
        self.assertEqual(self.var_three, vars[4])

    # =================================================================================
    # SECTION: DELETE TESTS
    # =================================================================================

    def test_delete_variable_works(self):
        self._create_variable_group()
        self.assertEqual(5, Variable.select().count())
        self.repo.delete(self.var_two)
        self.assertEqual(4, Variable.select().count())
        self.repo.delete(self.var_five)
        self.assertEqual(3, Variable.select().count())
        with self.assertRaises(DoesNotExist):
            Variable.get(Variable.id == self.var_two)
            Variable.get(Variable.id == self.var_five)


class TestVariableTagging(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        init_database(testing=True)
        db.connect()
        db.bind([Variable, Tag, VariableTag])
        db.create_tables([Variable, Tag, VariableTag])

    @classmethod
    def tearDownClass(cls):
        # Close database connection after all tests
        db.drop_tables([Variable, Tag, VariableTag])
        db.close()

    def setUp(self):
        Tag.delete().execute()
        Variable.delete().execute()
        VariableTag.delete().execute()
        self.repo = VariableRepository()

    def _create_var_tags(self):
        self.var_one = Variable.create(name="var_one", value="Value one")
        self.var_two = Variable.create(name="var_two", value="Value two")
        self.tag_one = Tag.create(name="tag_one", description="Tag One Description")
        self.tag_two = Tag.create(name="tag_two", description="Tag Two Description")
        self.var_tag_one = VariableTag.create(variable=self.var_one, tag=self.tag_one)
        self.var_tag_two = VariableTag.create(variable=self.var_two, tag=self.tag_one)
        self.var_tag_three = VariableTag.create(variable=self.var_two, tag=self.tag_two)

    def test_add_tag(self):
        var = Variable.create(name="test_var", value="test")
        tag = Tag.create(name="test_tag", description="test_description")
        results = self.repo.add_tags(variable=var, tags=[tag])
        var_tag = VariableTag.get(variable=var, tag=tag)
        self.assertTrue(isinstance(var_tag, VariableTag))
        self.assertEqual("test_tag", results.added[0])
        self.assertEqual(0, len(results.existing))

    def test_add_multiple_tags(self):
        var = Variable.create(name="test_var", value="test")
        tag1 = Tag.create(name="test_tag1", description="test_description1")
        tag2 = Tag.create(name="test_tag2", description="test_description2")
        results = self.repo.add_tags(variable=var, tags=[tag1, tag2])
        fetched_var_tag1 = VariableTag.get(variable=var, tag=tag1)
        fetched_var_tag2 = VariableTag.get(variable=var, tag=tag2)
        self.assertTrue(isinstance(fetched_var_tag1, VariableTag))
        self.assertTrue(isinstance(fetched_var_tag2, VariableTag))
        self.assertEqual("test_tag1", results.added[0])
        self.assertEqual("test_tag2", results.added[1])
        self.assertEqual(0, len(results.existing))

    def test_mixed_tagging_of_existing_and_new_works_correctly(self):
        var = Variable.create(name="test_var", value="test")
        tag1 = Tag.create(name="test_tag1", description="test_description1")
        var_tag1 = VariableTag.create(variable=var, tag=tag1)
        tag2 = Tag.create(name="test_tag2", description="test_description2")
        results = self.repo.add_tags(variable=var, tags=[tag1, tag2])
        self.assertEqual(1, len(results.added))
        self.assertEqual(1, len(results.existing))

    def test_add_tag_with_no_tags_does_nothing(self):
        var = Variable.create(name="test_var", value="test")
        results = self.repo.add_tags(variable=var, tags=[])
        self.assertEqual(0, len(results.added))
        self.assertEqual(0, len(results.existing))

    def test_add_tag_with_non_existent_tag_raises_exception(self):
        var = Variable.create(name="test_var", value="test")
        with self.assertRaises(TagAttachError):
            self.repo.add_tags(variable=var, tags=[None])

    def test_add_tag_with_non_existent_variable_name_raises_exception(self):
        tag = Tag.create(name="test_tag", description="test_description")
        with self.assertRaises(TagAttachError):
            self.repo.add_tags(variable=None, tags=[tag])

    def test_double_tagging_does_not_raise_error(self):
        var = Variable.create(name="test_var", value="test")
        tag = Tag.create(name="test_tag")
        var_tag = VariableTag.create(variable=var, tag=tag)

        results = self.repo.add_tags(variable=var, tags=[tag])
        self.assertTrue(isinstance(var_tag, VariableTag))
        self.assertEqual(0, len(results.added))
        self.assertEqual("test_tag", results.existing[0])

    def test_add_tag_is_atomic_and_no_tags_are_added_if_one_fails(self):
        var = Variable.create(name="test_var", value="test")
        tag = Tag.create(name="test_tag")
        with self.assertRaises(TagAttachError):
            self.repo.add_tags(variable=var, tags=[tag, None])
        self.assertEqual(0, VariableTag.select().count())

    def test_remove_tag(self):
        self._create_var_tags()
        var = VariableTag.get(variable=self.var_one, tag=self.tag_one)
        self.repo.remove_tags(variable=var, tags=[self.tag_one])
        with self.assertRaises(DoesNotExist):
            VariableTag.get(variable=self.var_one, tag=self.tag_one)

    def test_remove_multiple_tags(self):
        self._create_var_tags()
        result = self.repo.remove_tags(
            variable=self.var_two, tags=[self.tag_one, self.tag_two]
        )
        self.assertEqual(2, len(result.removed))
        self.assertEqual(0, len(result.not_attached))
        with self.assertRaises(DoesNotExist):
            VariableTag.get(variable=self.var_two, tag=self.tag_one)
            VariableTag.get(variable=self.var_two, tag=self.tag_two)

    def test_remove_tag_with_mix_of_existing_and_non_existing_tagged_variables(self):
        self._create_var_tags()
        result = self.repo.remove_tags(
            variable=self.var_one, tags=[self.tag_one, self.tag_two]
        )
        self.assertEqual(1, len(result.removed))
        self.assertEqual(1, len(result.not_attached))
        with self.assertRaises(DoesNotExist):
            VariableTag.get(variable=self.var_one, tag=self.tag_one)

    def test_remove_tag_with_no_tags_does_nothing(self):
        self._create_var_tags()
        result = self.repo.remove_tags(variable=self.var_one, tags=[])
        self.assertEqual(0, len(result.removed))
        self.assertEqual(0, len(result.not_attached))

    def test_remove_tag_with_non_existent_tag_raises_exception(self):
        self._create_var_tags()
        with self.assertRaises(TagDetachError):
            self.repo.remove_tags(variable=self.var_one, tags=[None])

    def test_remove_tag_with_non_existent_variable_name_raises_exception(self):
        self._create_var_tags()
        result = self.repo.remove_tags(variable=None, tags=[self.tag_one])
        self.assertEqual(0, len(result.removed))
        self.assertEqual(1, len(result.not_attached))

    def test_removing_a_tag_twice_does_not_raise_error(self):
        self._create_var_tags()
        r1 = self.repo.remove_tags(variable=self.var_two, tags=[self.tag_one])
        self.assertEqual(1, len(r1.removed))
        self.assertEqual(0, len(r1.not_attached))
        r2 = self.repo.remove_tags(variable=self.var_two, tags=[self.tag_one])
        self.assertEqual(0, len(r2.removed))
        self.assertEqual(1, len(r2.not_attached))

    def test_remove_tag_is_atomic_and_no_tags_are_removed_if_one_fails(self):
        self._create_var_tags()
        with self.assertRaises(TagDetachError):
            self.repo.remove_tags(variable=self.var_two, tags=[self.tag_one, None])
        VariableTag.get(variable=self.var_one, tag=self.tag_one)
