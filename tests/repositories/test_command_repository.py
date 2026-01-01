import unittest

from peewee import DoesNotExist

from cmdbox.database import db, get_db
from cmdbox.repositories.errors import (
    ValidationError,
    AliasConflictError,
    UnknownAliasError,
    TagAttachError,
    TagDetachError,
    UpdateError,
    UnknownCommandError,
)
from cmdbox.models import Command, Tag, CommandTag, ALL_MODELS
from cmdbox.repositories.command_repository import CommandRepository
from cmdbox.database import ensure_schema


class TestCommandRepository(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        get_db(testing=True)
        ensure_schema()

    @classmethod
    def tearDownClass(cls):
        # Close database connection after all tests
        db.drop_tables(ALL_MODELS)
        db.close()

    def setUp(self):
        Command.delete().execute()
        Tag.delete().execute()
        CommandTag.delete().execute()
        self.repo = CommandRepository()

    def _create_command_group(self):
        self.cmd_one = Command.create(
            alias="test", template="echo e", description="Orangutan aligator"
        )
        self.cmd_two = Command.create(
            alias="test2", template="echo d", description="Harpsichord aligator"
        )
        self.cmd_three = Command.create(
            alias="test3", template="echo c", description="Aligator aligator"
        )
        self.cmd_four = Command.create(
            alias="test4", template="echo b", description="Zebra aligator"
        )
        self.cmd_five = Command.create(
            alias="test5", template="echo a", description="Length aligator gator"
        )

    def _tag_command_group(self):
        self.tag_one = Tag.create(name="tag_one")
        self.tag_two = Tag.create(name="tag_two")
        self.tag_three = Tag.create(name="tag_three")
        self.cmd_tag_one = CommandTag.create(command=self.cmd_one, tag=self.tag_one)
        self.cmd_tag_two = CommandTag.create(command=self.cmd_two, tag=self.tag_one)
        self.cmd_tag_three = CommandTag.create(command=self.cmd_two, tag=self.tag_two)
        self.cmd_tag_four = CommandTag.create(command=self.cmd_two, tag=self.tag_three)
        self.cmd_tag_five = CommandTag.create(
            command=self.cmd_three, tag=self.tag_three
        )

    # =================================================================================
    # SECTION: CREATE TESTS
    # =================================================================================

    def test_create(self):
        command = self.repo.create(
            alias="test", template="echo test", description="Test command"
        )
        self.assertTrue(isinstance(command, Command))
        self.assertEqual(1, Command.select().count())

    def test_create_no_alias_supplied(self):
        """Command should not be created without an alias."""
        with self.assertRaises(ValidationError):
            self.repo.create(
                alias=None, template="echo test", description="Test command"
            )
        self.assertEqual(0, Command.select().count())

    def test_create_duplicate_alias_raises_exception(self):
        """Command should not be created if an alias already exists."""
        Command.create(alias="test", template="echo test", description="Test command")
        with self.assertRaises(AliasConflictError):
            self.repo.create(
                alias="test", template="echo test", description="Test command"
            )

    def test_create_no_template_supplied(self):
        """Command should not be created without a template."""
        with self.assertRaises(ValidationError):
            self.repo.create(alias="test", template=None, description="Test command")
        self.assertEqual(0, Command.select().count())

    def test_create_no_description_supplied(self):
        """Command should be created without a description if none is supplied."""
        command = self.repo.create(alias="test", template="echo test")
        self.assertTrue(isinstance(command, Command))
        self.assertEqual(1, Command.select().count())

    def test_duplicate_template_is_allowed(self):
        """Commands are allowed to have duplicate templates."""
        self.repo.create(alias="test", template="echo test", description="Test command")
        Command.create(alias="test2", template="echo test", description="Test command")

    def test_duplicate_description_is_allowed(self):
        """Commands are allowed to have duplicate descriptions."""
        self.repo.create(alias="test", template="echo test", description="Test command")
        self.repo.create(
            alias="test2", template="echo test2", description="Test command"
        )

    def test_all_symbols_and_numbers_are_allowed_in_alias(self):
        """Commands are allowed to have symbols and numbers in their aliases."""
        self.repo.create(
            alias="test-1!@#$%^&*()_+-=[]\;/.,<>?:{}|",
            template="echo test",
            description="Test command",
        )

    def test_unicode_characters_are_allowed_in_alias(self):
        """Commands are allowed to have unicode characters in their aliases."""
        self.repo.create(
            alias="git-✨", template="echo test", description="Test command"
        )

    def test_all_symbols_are_allowed_in_template(self):
        """Commands are allowed to have weird characters in their templates."""
        self.repo.create(
            alias="test",
            template="echo test!@#$%^&*()_+-=[]\;/.,<>?:{}",
            description="Test command",
        )

    def test_unicode_is_allowed_in_template(self):
        self.repo.create(alias="test", template="git-✨")

    def test_create_with_white_space_in_middle_of_alias_is_not_allowed(self):
        """Aliases with whitespace should be throw validation error."""
        with self.assertRaises(ValidationError):
            self.repo.create(alias="test test", template="echo test")

    def test_create_with_white_space_at_beginning_and_end_of_alias_is_stripped_and_allowed(
        self,
    ):
        """Aliases with whitespace should be throw validation error."""
        command = self.repo.create(alias=" test ", template="echo test")
        self.assertTrue(isinstance(command, Command))
        self.assertEqual("test", command.alias)

    # =================================================================================
    # SECTION: GET TESTS
    # =================================================================================

    def test_get_command_by_alias(self):
        """Test retrieving a command by its alias."""
        command = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        self.assertEqual(command, self.repo.get_by_alias(alias="test"))

    def test_get_by_alias_raises_exception_if_not_found(self):
        """None should be returned if no command is found with the given alias."""
        Command.create(alias="test", template="echo test", description="Test command")
        with self.assertRaises(UnknownAliasError):
            self.assertIsNone(self.repo.get_by_alias(alias="test_two"))

    def test_alias_capitalization_does_not_affect_result(self):
        """Command aliases should be case-insensitive."""
        command = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        self.assertEqual(command, self.repo.get_by_alias(alias="TesT"))

    def test_get_by_blank_alias_raises_exception(self):
        """A blank alias should be treated as an unknown alias."""
        Command.create(alias="test", template="echo test", description="Test command")
        with self.assertRaises(UnknownAliasError):
            self.assertIsNone(self.repo.get_by_alias(alias=""))

    def test_get_by_other_fields_does_not_work(self):
        """Commands should only be fetched by alias"""
        template = "echo test"
        Command.create(alias="test", template=template, description="Test command")
        with self.assertRaises(UnknownAliasError):
            self.assertIsNone(self.repo.get_by_alias(alias=template))

    def test_get_command_with_all_symbols_and_numbers_is_allowed(self):
        Command.create(
            alias="test-1!@#$%^&*()_+-=[]\;/.,<>?:{}|",
            template="echo test",
            description="Test command",
        )
        self.repo.get_by_alias(alias="test-1!@#$%^&*()_+-=[]\;/.,<>?:{}|")

    def test_get_command_allows_unicode_query(self):
        Command.create(alias="git-✨", template="echo test", description="Test command")
        self.repo.get_by_alias(alias="git-✨")

    def test_get_command_by_id(self):
        cmd = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        command = self.repo.get_by_id(cmd.id)
        self.assertEqual(cmd, command)

    def test_get_command_by_id_raises_error_for_nonexistent_id(self):
        Command.create(alias="test", template="echo test", description="Test command")
        with self.assertRaises(UnknownCommandError):
            self.repo.get_by_id(cmd_id=999)

    # =================================================================================
    # SECTION: UPDATE TESTS
    # =================================================================================

    def test_update_command_alias(self):
        """Test updating the alias of a command."""
        command = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        cmd = self.repo.update(command=command, alias="new_test")
        self.assertEqual("new_test", Command.get(Command.alias == "new_test").alias)
        self.assertEqual(command, cmd)

    def test_update_alias_to_duplicate_not_allowed(self):
        """Updating a command alias to a duplicate should raise an exception."""
        cmd = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        Command.create(
            alias="new_test", template="echo test", description="Test command"
        )
        with self.assertRaises(AliasConflictError):
            self.repo.update(command=cmd, alias="new_test")

    def test_update_alias_to_itself_does_not_throw_exception(self):
        """Updating a command alias to itself should not raise an exception and should not change the command."""
        cmd = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        self.repo.update(command=cmd, alias="test")
        self.assertIsNotNone(Command.get(Command.alias == "test"))

    def test_update_command_template_works(self):
        """Test updating the template of a command."""
        command = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        cmd = self.repo.update(command=command, template="echo new test")
        self.assertEqual("echo new test", Command.get(Command.alias == "test").template)
        self.assertEqual(command, cmd)

    def test_updating_with_unknown_alias_raises_exception(self):
        """Updating a command with an unknown alias should raise an exception."""
        with self.assertRaises(UpdateError):
            self.repo.update(command=None, template="echo new test")

    def test_updating_template_to_blank_string_raises_exception(self):
        """Updating a command template to an empty string should raise an exception."""
        cmd = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        with self.assertRaises(ValidationError):
            self.repo.update(command=cmd, template="")

    def test_updating_template_to_existing_template_is_allowed(self):
        """
        Updating a command template to an existing template should not raise an exception.
        Commands are allowed to have duplicate templates.
        """
        cmd = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        self.repo.update(command=cmd, template="echo new test")
        self.assertEqual("echo new test", Command.get(Command.alias == "test").template)

    def test_updating_command_description_works(self):
        """Test updating the description of a command."""
        command = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        cmd = self.repo.update(command=command, description="New description")
        self.assertEqual(
            "New description", Command.get(Command.alias == "test").description
        )
        self.assertEqual(command, cmd)

    def test_updating_command_to_remove_description_is_allowed(self):
        """Updating a command to remove its description should not raise an exception."""
        cmd = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        self.repo.update(command=cmd, description="")
        self.assertEqual("", Command.get(Command.alias == "test").description)

    def test_update_with_no_fields_throws_exception(self):
        cmd = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        with self.assertRaises(UpdateError):
            self.repo.update(command=cmd)

    def test_update_with_white_space_in_middle_of_alias_is_not_allowed(self):
        """Aliases with whitespace should be throw validation error."""
        cmd = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        with self.assertRaises(ValidationError):
            self.repo.update(command=cmd, alias="test2 test")

    def test_update_with_white_space_at_beginning_and_end_of_alias_is_stripped_and_allowed(
        self,
    ):
        """Aliases with whitespace should be throw validation error."""
        cmd = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        self.repo.update(command=cmd, alias=" test2 ")
        self.assertEqual("test2", Command.get(Command.id == cmd.id).alias)

    # =================================================================================
    # SECTION: LIST TESTS
    # =================================================================================

    def test_list_all_works(self):
        """Test listing all commands."""
        self._create_command_group()
        commands = self.repo.list_all()
        self.assertEqual(5, len(commands))
        self.assertTrue(self.cmd_one in commands)
        self.assertTrue(self.cmd_two in commands)
        self.assertTrue(self.cmd_three in commands)
        self.assertTrue(self.cmd_four in commands)
        self.assertTrue(self.cmd_five in commands)

    def test_list_all_returns_empty_list_if_no_commands(self):
        """Test that an empty list is returned if no commands exist."""
        commands = self.repo.list_all()
        self.assertEqual([], commands)

    def test_list_all_ordered_by_alias(self):
        """Test that commands are returned in alphabetical order by alias."""
        self._create_command_group()
        commands = self.repo.list_all()
        self.assertEqual(self.cmd_one, commands[0])
        self.assertEqual(self.cmd_two, commands[1])
        self.assertEqual(self.cmd_three, commands[2])
        self.assertEqual(self.cmd_four, commands[3])
        self.assertEqual(self.cmd_five, commands[4])

    def test_list_ordered_by_template(self):
        """Test that commands are returned in alphabetical order by template."""
        self._create_command_group()
        commands = self.repo.list_all(order_by="template")
        self.assertEqual(self.cmd_five, commands[0])
        self.assertEqual(self.cmd_four, commands[1])
        self.assertEqual(self.cmd_three, commands[2])
        self.assertEqual(self.cmd_two, commands[3])
        self.assertEqual(self.cmd_one, commands[4])

    def test_list_ordered_by_description(self):
        """Test that commands are returned in alphabetical order by description."""
        self._create_command_group()
        commands = self.repo.list_all(order_by="description")
        self.assertEqual(self.cmd_three, commands[0])
        self.assertEqual(self.cmd_two, commands[1])
        self.assertEqual(self.cmd_five, commands[2])
        self.assertEqual(self.cmd_one, commands[3])
        self.assertEqual(self.cmd_four, commands[4])

    def test_list_ordered_by_alias_descending(self):
        """Test that commands are returned in descending order by alias."""
        self._create_command_group()
        commands = self.repo.list_all(order_by="-alias")
        self.assertEqual(self.cmd_five, commands[0])
        self.assertEqual(self.cmd_four, commands[1])
        self.assertEqual(self.cmd_three, commands[2])
        self.assertEqual(self.cmd_two, commands[3])
        self.assertEqual(self.cmd_one, commands[4])

    def test_list_all_limit_functions_correctly(self):
        self._create_command_group()
        commands = self.repo.list_all(limit=2)
        self.assertEqual(2, len(commands))

    def test_list_none_limit_does_not_limit(self):
        self._create_command_group()
        commands = self.repo.list_all(limit=None)
        self.assertEqual(5, len(commands))

    def test_that_limit_of_zero_returns_empty_list(self):
        self._create_command_group()
        commands = self.repo.list_all(limit=0)
        self.assertEqual(0, len(commands))

    # =================================================================================
    # SECTION: LIST BY TAG TESTS
    # =================================================================================

    def test_list_by_tag(self):
        self._create_command_group()
        self._tag_command_group()
        cmds = self.repo.list_by_tag([self.tag_one])
        self.assertEqual(2, len(cmds))
        self.assertTrue(self.cmd_one in cmds)
        self.assertTrue(self.cmd_two in cmds)

    def test_list_by_multiple_tags(self):
        self._create_command_group()
        self._tag_command_group()
        cmds = self.repo.list_by_tag([self.tag_one, self.tag_three])
        self.assertEqual(3, len(cmds))
        self.assertTrue(self.cmd_one in cmds)
        self.assertTrue(self.cmd_two in cmds)
        self.assertTrue(self.cmd_three in cmds)

    def test_list_by_null_tag_returns_empty_list(self):
        self._create_command_group()
        cmds = self.repo.list_by_tag([None])
        self.assertEqual([], cmds)

    def test_list_by_tag_order_by_template_changes_order(self):
        self._create_command_group()
        self._tag_command_group()
        cmds = self.repo.list_by_tag(
            [self.tag_one, self.tag_three], order_by="template"
        )
        self.assertEqual(3, len(cmds))
        self.assertEqual(self.cmd_three, cmds[0])
        self.assertEqual(self.cmd_two, cmds[1])
        self.assertEqual(self.cmd_one, cmds[2])

    def test_list_by_tag_limits_apply(self):
        self._create_command_group()
        self._tag_command_group()
        cmds = self.repo.list_by_tag([self.tag_one, self.tag_three], limit=2)
        self.assertEqual(2, len(cmds))

    # =================================================================================
    # SECTION: SEARCH TESTS
    # =================================================================================

    def test_search_empty_term_returns_empty_list(self):
        self._create_command_group()
        commands = self.repo.search("")
        self.assertEqual(0, len(commands))

    def test_search_returns_empty_list_if_no_commands_match(self):
        commands = self.repo.search("anyname")
        self.assertEqual([], commands)

    def test_search_returns_matching_commands_on_default_fields(self):
        self._create_command_group()
        commands = self.repo.search("test")
        self.assertEqual(5, len(commands))

        commands = self.repo.search("test2")
        self.assertEqual(self.cmd_two, commands[0])

        commands = self.repo.search("Zebra")
        self.assertEqual(self.cmd_four, commands[0])

    def test_search_returns_matching_commands_on_specific_fields(self):
        self._create_command_group()
        commands = self.repo.search("echo", fields="template")
        self.assertEqual(5, len(commands))

        commands = self.repo.search("echo", fields="description")
        self.assertEqual(0, len(commands))

    def test_search_returns_matching_commands_on_multiple_fields(self):
        self._create_command_group()
        commands = self.repo.search("echo", fields=["template", "description"])
        self.assertEqual(5, len(commands))

    def test_search_returns_matching_commands_on_multiple_fields_supplied_as_list(self):
        self._create_command_group()
        commands = self.repo.search("echo", fields=["template", "description"])
        self.assertEqual(5, len(commands))

    def test_search_on_invalid_field_raises_exception(self):
        self._create_command_group()
        with self.assertRaises(ValueError):
            self.repo.search("echo", fields="invalid_field")

    def test_search_on_empty_str_returns_no_results(self):
        self._create_command_group()
        commands = self.repo.search("")
        self.assertEqual(0, len(commands))

    def test_search_is_case_insensitive(self):
        self._create_command_group()
        commands = self.repo.search("TEST")
        self.assertEqual(5, len(commands))

    def test_search_is_ordered_by_most_relevant(self):
        self._create_command_group()
        commands = self.repo.search("gator", fields="description")
        self.assertEqual(self.cmd_three, commands[0])
        self.assertEqual(self.cmd_five, commands[1])
        self.assertEqual(self.cmd_four, commands[2])
        self.assertEqual(self.cmd_one, commands[3])
        self.assertEqual(self.cmd_two, commands[4])

    def test_search_limit_is_applied(self):
        self._create_command_group()
        commands = self.repo.search("echo", fields=["template", "description"], limit=3)
        self.assertEqual(3, len(commands))

    # =================================================================================
    # SECTION: DELETE TESTS
    # =================================================================================

    def testa_delete_functions_correctly(self):
        self._create_command_group()
        self.assertEqual(5, Command.select().count())
        self.repo.delete(self.cmd_two)
        self.assertEqual(4, Command.select().count())
        self.repo.delete(self.cmd_five)
        self.assertEqual(3, Command.select().count())
        with self.assertRaises(DoesNotExist):
            Command.get(Command.id == self.cmd_two)
            Command.get(Command.id == self.cmd_five)


class TestCommandTagging(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        get_db(testing=True)
        ensure_schema()

    @classmethod
    def tearDownClass(cls):
        # Close database connection after all tests
        db.drop_tables(ALL_MODELS)
        db.close()

    def setUp(self):
        Tag.delete().execute()
        Command.delete().execute()
        CommandTag.delete().execute()
        self.repo = CommandRepository()

    def _create_cmd_tags(self):
        self.cmd_one = Command.create(alias="cmd_one", template="echo Command one")
        self.cmd_two = Command.create(alias="cmd_two", template="echo Command two")
        self.tag_one = Tag.create(name="tag_one", description="Tag One Description")
        self.tag_two = Tag.create(name="tag_two", description="Tag Two Description")
        self.cmd_tag_one = CommandTag.create(command=self.cmd_one, tag=self.tag_one)
        self.cmd_tag_two = CommandTag.create(command=self.cmd_two, tag=self.tag_one)
        self.cmd_tag_three = CommandTag.create(command=self.cmd_two, tag=self.tag_two)

    def test_add_tag(self):
        cmd = Command.create(alias="test_cmd", template="echo test")
        tag = Tag.create(name="test_tag", description="test_description")
        results = self.repo.add_tags(command=cmd, tags=[tag])
        cmd_tag = CommandTag.get(command=cmd, tag=tag)
        self.assertTrue(isinstance(cmd_tag, CommandTag))
        self.assertEqual("test_tag", results.added[0])
        self.assertEqual(0, len(results.existing))

    def test_add_multiple_tags(self):
        cmd = Command.create(alias="test_cmd", template="echo test")
        tag1 = Tag.create(name="test_tag1", description="test_description1")
        tag2 = Tag.create(name="test_tag2", description="test_description2")
        results = self.repo.add_tags(command=cmd, tags=[tag1, tag2])
        fetched_cmd_tag1 = CommandTag.get(command=cmd, tag=tag1)
        fetched_cmd_tag2 = CommandTag.get(command=cmd, tag=tag2)
        self.assertTrue(isinstance(fetched_cmd_tag1, CommandTag))
        self.assertTrue(isinstance(fetched_cmd_tag2, CommandTag))
        self.assertEqual("test_tag1", results.added[0])
        self.assertEqual("test_tag2", results.added[1])
        self.assertEqual(0, len(results.existing))

    def test_mixed_tagging_of_existing_and_new_works_correctly(self):
        cmd = Command.create(alias="test_cmd", template="echo test")
        tag1 = Tag.create(name="test_tag1", description="test_description1")
        cmd_tag1 = CommandTag.create(command=cmd, tag=tag1)
        tag2 = Tag.create(name="test_tag2", description="test_description2")
        results = self.repo.add_tags(command=cmd, tags=[tag1, tag2])
        self.assertEqual(1, len(results.added))
        self.assertEqual(1, len(results.existing))

    def test_add_tag_with_no_tags_does_nothing(self):
        cmd = Command.create(alias="test_cmd", template="echo test")
        results = self.repo.add_tags(command=cmd, tags=[])
        self.assertEqual(0, len(results.added))
        self.assertEqual(0, len(results.existing))

    def test_add_tag_with_non_existent_tag_raises_exception(self):
        cmd = Command.create(alias="test_cmd", template="echo test")
        with self.assertRaises(TagAttachError):
            self.repo.add_tags(command=cmd, tags=[None])

    def test_add_tag_with_non_existent_command_alias_raises_exception(self):
        tag = Tag.create(name="test_tag", description="test_description")
        with self.assertRaises(TagAttachError):
            self.repo.add_tags(command=None, tags=[tag])

    def test_double_tagging_does_not_raise_error(self):
        cmd = Command.create(alias="test_cmd", template="echo test")
        tag = Tag.create(name="test_tag")
        cmd_tag = CommandTag.create(command=cmd, tag=tag)

        results = self.repo.add_tags(command=cmd, tags=[tag])
        self.assertTrue(isinstance(cmd_tag, CommandTag))
        self.assertEqual(0, len(results.added))
        self.assertEqual("test_tag", results.existing[0])

    def test_add_tag_is_atomic_and_no_tags_are_added_if_one_fails(self):
        cmd = Command.create(alias="test_cmd", template="echo test")
        tag = Tag.create(name="test_tag")
        with self.assertRaises(TagAttachError):
            self.repo.add_tags(command=cmd, tags=[tag, None])
        self.assertEqual(0, CommandTag.select().count())

    def test_remove_tag(self):
        self._create_cmd_tags()
        cmd = CommandTag.get(command=self.cmd_one, tag=self.tag_one)
        self.repo.remove_tags(command=cmd, tags=[self.tag_one])
        with self.assertRaises(DoesNotExist):
            CommandTag.get(command=self.cmd_one, tag=self.tag_one)

    def test_remove_multiple_tags(self):
        self._create_cmd_tags()
        result = self.repo.remove_tags(
            command=self.cmd_two, tags=[self.tag_one, self.tag_two]
        )
        self.assertEqual(2, len(result.removed))
        self.assertEqual(0, len(result.not_attached))
        with self.assertRaises(DoesNotExist):
            CommandTag.get(command=self.cmd_two, tag=self.tag_one)
            CommandTag.get(command=self.cmd_two, tag=self.tag_two)

    def test_remove_tag_with_mix_of_existing_and_non_existing_tagged_commands(self):
        self._create_cmd_tags()
        result = self.repo.remove_tags(
            command=self.cmd_one, tags=[self.tag_one, self.tag_two]
        )
        self.assertEqual(1, len(result.removed))
        self.assertEqual(1, len(result.not_attached))
        with self.assertRaises(DoesNotExist):
            CommandTag.get(command=self.cmd_one, tag=self.tag_one)

    def test_remove_tag_with_no_tags_does_nothing(self):
        self._create_cmd_tags()
        result = self.repo.remove_tags(command=self.cmd_one, tags=[])
        self.assertEqual(0, len(result.removed))
        self.assertEqual(0, len(result.not_attached))

    def test_remove_tag_with_non_existent_tag_raises_exception(self):
        self._create_cmd_tags()
        with self.assertRaises(TagDetachError):
            self.repo.remove_tags(command=self.cmd_one, tags=[None])

    def test_remove_tag_with_non_existent_command_alias_does_not_raise_exception(self):
        self._create_cmd_tags()
        result = self.repo.remove_tags(command=None, tags=[self.tag_one])
        self.assertEqual(0, len(result.removed))
        self.assertEqual(1, len(result.not_attached))

    def test_removing_a_tag_twice_does_not_raise_error(self):
        self._create_cmd_tags()
        r1 = self.repo.remove_tags(command=self.cmd_two, tags=[self.tag_one])
        self.assertEqual(1, len(r1.removed))
        self.assertEqual(0, len(r1.not_attached))
        r2 = self.repo.remove_tags(command=self.cmd_two, tags=[self.tag_one])
        self.assertEqual(0, len(r2.removed))
        self.assertEqual(1, len(r2.not_attached))

    def test_remove_tag_is_atomic_and_no_tags_are_removed_if_one_fails(self):
        self._create_cmd_tags()
        with self.assertRaises(TagDetachError):
            self.repo.remove_tags(command=self.cmd_two, tags=[self.tag_one, None])
        CommandTag.get(command=self.cmd_one, tag=self.tag_one)
