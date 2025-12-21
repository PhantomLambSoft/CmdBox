import unittest

from peewee import DoesNotExist

from cmdbox.database import db, init_database
from cmdbox.exceptions import (
    ValidationError,
    AliasConflictError,
    UnknownAliasError,
    NameConflictError,
    UnknownNameError,
    UnknownTagError,
    CmdboxError,
    TagAttachError,
)
from cmdbox.repositories import (
    CommandRepository,
    VariableRepository,
    TagRepository,
    BaseRepository,
)
from cmdbox.models import Command, Variable, Tag, CommandTag, VariableTag


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


class TestCommandRepository(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        init_database(testing=True)
        db.connect()
        db.bind([Command, Tag, CommandTag])
        db.create_tables([Command, Tag, CommandTag])

    @classmethod
    def tearDownClass(cls):
        # Close database connection after all tests
        db.drop_tables([Command, Tag, CommandTag])
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

    def test_get_command_by_alias(self):
        """Test retrieving a command by its alias."""
        command = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        self.assertEqual(command, self.repo.get_by_alias(alias="test"))

    def test_create_with_tags(self):
        tag = Tag.create(name="tag_one")
        cmd = self.repo.create(alias="test", template="echo test", tags=["tag_one"])
        self.assertTrue(isinstance(cmd, Command))
        Command.select()
        CommandTag.get(CommandTag.command == cmd, CommandTag.tag == tag)

    def test_create_with_tags_multiple(self):
        tag_one = Tag.create(name="tag_one")
        tag_two = Tag.create(name="tag_two")
        cmd = self.repo.create(
            alias="test", template="echo test", tags=["tag_one", "tag_two"]
        )
        self.assertTrue(isinstance(cmd, Command))
        command = Command.select()
        CommandTag.get(CommandTag.command == command, CommandTag.tag == tag_one)
        CommandTag.get(CommandTag.command == command, CommandTag.tag == tag_two)

    def test_create_with_tags_duplicate_does_not_raise_error(self):
        tag = Tag.create(name="tag_one")
        cmd = self.repo.create(
            alias="test", template="echo test", tags=["tag_one", "tag_one"]
        )
        self.assertTrue(isinstance(cmd, Command))
        Command.select()
        CommandTag.get(CommandTag.command == cmd, CommandTag.tag == tag)

    def test_create_with_tags_non_existent_tag_is_atomic_and_raises_error(self):
        with self.assertRaises(UnknownTagError):
            self.repo.create(alias="test", template="echo test", tags=["tag_one"])
        with self.assertRaises(DoesNotExist):
            Command.get(Command.alias == "test")

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

    def test_update_command_alias(self):
        """Test updating the alias of a command."""
        command = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        cmd = self.repo.update(cmd_alias=command.alias, alias="new_test")
        self.assertEqual("new_test", Command.get(Command.alias == "new_test").alias)
        self.assertEqual(command, cmd)

    def test_update_alias_to_duplicate_not_allowed(self):
        """Updating a command alias to a duplicate should raise an exception."""
        Command.create(alias="test", template="echo test", description="Test command")
        Command.create(
            alias="new_test", template="echo test", description="Test command"
        )
        with self.assertRaises(AliasConflictError):
            self.repo.update(cmd_alias="test", alias="new_test")

    def test_update_alias_to_itself_does_not_throw_exception(self):
        """Updating a command alias to itself should not raise an exception and should not change the command."""
        Command.create(alias="test", template="echo test", description="Test command")
        self.repo.update(cmd_alias="test", alias="test")
        self.assertIsNotNone(Command.get(Command.alias == "test"))

    def test_update_command_template_works(self):
        """Test updating the template of a command."""
        command = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        cmd = self.repo.update(cmd_alias="test", template="echo new test")
        self.assertEqual("echo new test", Command.get(Command.alias == "test").template)
        self.assertEqual(command, cmd)

    def test_updating_with_unknown_alias_raises_exception(self):
        """Updating a command with an unknown alias should raise an exception."""
        with self.assertRaises(UnknownAliasError):
            self.repo.update(cmd_alias="invalid_name", template="echo new test")

    def test_updating_template_to_blank_string_raises_exception(self):
        """Updating a command template to an empty string should raise an exception."""
        Command.create(alias="test", template="echo test", description="Test command")
        with self.assertRaises(ValidationError):
            self.repo.update(cmd_alias="test", template="")

    def test_updating_template_to_existing_template_is_allowed(self):
        """
        Updating a command template to an existing template should not raise an exception.
        Commands are allowed to have duplicate templates.
        """
        Command.create(alias="test", template="echo test", description="Test command")
        self.repo.update(cmd_alias="test", template="echo new test")
        self.assertEqual("echo new test", Command.get(Command.alias == "test").template)

    def test_updating_command_description_works(self):
        """Test updating the description of a command."""
        command = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        cmd = self.repo.update(cmd_alias="test", description="New description")
        self.assertEqual(
            "New description", Command.get(Command.alias == "test").description
        )
        self.assertEqual(command, cmd)

    def test_updating_command_to_remove_description_is_allowed(self):
        """Updating a command to remove its description should not raise an exception."""
        Command.create(alias="test", template="echo test", description="Test command")
        self.repo.update(cmd_alias="test", description="")
        self.assertEqual("", Command.get(Command.alias == "test").description)

    def test_update_with_no_fields_throws_exception(self):
        Command.create(alias="test", template="echo test", description="Test command")
        with self.assertRaises(ValueError):
            self.repo.update(cmd_alias="test")

    def test_update_with_white_space_in_middle_of_alias_is_not_allowed(self):
        """Aliases with whitespace should be throw validation error."""
        Command.create(alias="test", template="echo test", description="Test command")
        with self.assertRaises(ValidationError):
            self.repo.update(cmd_alias="test", alias="test2 test")

    def test_update_with_white_space_at_beginning_and_end_of_alias_is_stripped_and_allowed(
        self,
    ):
        """Aliases with whitespace should be throw validation error."""
        cmd = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        self.repo.update(cmd_alias="test", alias=" test2 ")
        self.assertEqual("test2", Command.get(Command.id == cmd.id).alias)

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

    def test_list_by_tag(self):
        self._create_command_group()
        self._tag_command_group()
        cmds = self.repo.list_by_tag(["tag_one"])
        self.assertEqual(2, len(cmds))
        self.assertTrue(self.cmd_one in cmds)
        self.assertTrue(self.cmd_two in cmds)

    def test_list_by_multiple_tags(self):
        self._create_command_group()
        self._tag_command_group()
        cmds = self.repo.list_by_tag(["tag_one", "tag_three"])
        self.assertEqual(3, len(cmds))
        self.assertTrue(self.cmd_one in cmds)
        self.assertTrue(self.cmd_two in cmds)
        self.assertTrue(self.cmd_three in cmds)

    def test_list_by_nonexistent_tag_raises_error(self):
        self._create_command_group()
        with self.assertRaises(UnknownTagError):
            self.repo.list_by_tag(["invalid_tag"])

    def test_list_by_tag_order_by_template_changes_order(self):
        self._create_command_group()
        self._tag_command_group()
        cmds = self.repo.list_by_tag(["tag_one", "tag_three"], order_by="template")
        self.assertEqual(3, len(cmds))
        self.assertEqual(self.cmd_three, cmds[0])
        self.assertEqual(self.cmd_two, cmds[1])
        self.assertEqual(self.cmd_one, cmds[2])

    def test_list_by_tag_limits_apply(self):
        self._create_command_group()
        self._tag_command_group()
        cmds = self.repo.list_by_tag(["tag_one", "tag_three"], limit=2)
        self.assertEqual(2, len(cmds))

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

    def testa_delete_functions_correctly(self):
        self._create_command_group()
        self.assertEqual(5, Command.select().count())
        self.repo.delete(self.cmd_two.alias)
        self.assertEqual(4, Command.select().count())
        self.repo.delete(self.cmd_five.alias)
        self.assertEqual(3, Command.select().count())
        with self.assertRaises(DoesNotExist):
            Command.get(Command.id == self.cmd_two)
            Command.get(Command.id == self.cmd_five)

    def test_delete_nonexistent_command_raises_exception(self):
        cmd = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        with self.assertRaises(UnknownAliasError):
            self.repo.delete("invlaid_name")
        Command.get(Command.id == cmd.id)


class TestVariableRepository(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        init_database(testing=True)
        db.connect()
        db.bind([Variable])
        db.create_tables([Variable])

    @classmethod
    def tearDownClass(cls):
        # Close database connection after all tests
        db.drop_tables([Variable])
        db.close()

    def setUp(self):
        Variable.delete().execute()
        self.repo = VariableRepository()

    def _create_variable_group(self):
        self.var_one = Variable.create(name="test1", value="test_value_e Antelope Bee")
        self.var_two = Variable.create(name="test2", value="test_value_d Zebra Bee")
        self.var_three = Variable.create(
            name="test3", value="test_value_c Kangaroo Bee"
        )
        self.var_four = Variable.create(name="test4", value="test_value_b Bee Bee Bee")
        self.var_five = Variable.create(name="test5", value="test_value_a Bee Goldfish")

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

    def test_get_variable_by_name(self):
        variable = Variable.create(name="test", value="test_value")
        var = self.repo.get_by_name("test")
        self.assertEqual(variable, var)

    def test_get_unknown_variable_raises_exception(self):
        with self.assertRaises(UnknownNameError):
            self.repo.get_by_name("invalid_name")

    def test_variable_name_capitalisation_does_not_matter(self):
        variable = Variable.create(name="test", value="test_value")
        var = self.repo.get_by_name("TEST")
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

    def test_update_variable_name_works(self):
        Variable.create(name="test", value="test_value")
        self.repo.update(var_name="test", name="new_name")
        Variable.get(Variable.name == "new_name")

    def test_update_variable_value_works(self):
        variable = Variable.create(name="test", value="test_value")
        var = self.repo.update(var_name="test", value="new_value")
        self.assertEqual("new_value", Variable.get(Variable.name == "test").value)
        self.assertEqual(variable, var)

    def test_update_variable_with_blank_name_raises_exception(self):
        Variable.create(name="test", value="test_value")
        with self.assertRaises(ValidationError):
            self.repo.update(var_name="test", name="")

    def test_update_variable_with_null_name_does_nothing(self):
        var = Variable.create(name="test", value="test_value")
        self.repo.update(var_name="test", name=None)
        self.assertEqual(var, Variable.get(Variable.name == "test"))

    def test_update_variable_with_duplicate_name_raises_exception(self):
        Variable.create(name="test", value="test_value")
        Variable.create(name="test2", value="test_value2")
        with self.assertRaises(NameConflictError):
            self.repo.update(var_name="test2", name="test")

    def test_update_variable_value_to_an_existing_variable_is_allowed(self):
        """Multiple variables can have the same value."""
        Variable.create(name="test", value="test_value")
        Variable.create(name="test2", value="test_value2")
        self.repo.update(var_name="test2", value="test")

    def test_update_with_no_fields_throws_exception(self):
        Variable.create(name="test", value="test_value")
        with self.assertRaises(ValueError):
            self.repo.update(var_name="test")

    def test_update_name_with_white_space_in_middle_of_name_is_not_allowed(self):
        Variable.create(name="test", value="test_value")
        with self.assertRaises(ValidationError):
            self.repo.update(var_name="test", name="test2 test")

    def test_update_name_with_white_space_at_beginning_and_end_of_name_is_stripped(
        self,
    ):
        Variable.create(name="test", value="test_value")
        self.repo.update(var_name="test", name=" test2 ")
        self.assertEqual("test2", Variable.get(Variable.name == "test2").name)

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

    def test_delete_variable_works(self):
        self._create_variable_group()
        self.assertEqual(5, Variable.select().count())
        self.repo.delete("test2")
        self.assertEqual(4, Variable.select().count())
        self.repo.delete("test5")
        self.assertEqual(3, Variable.select().count())
        with self.assertRaises(DoesNotExist):
            Variable.get(Variable.id == self.var_two)
            Variable.get(Variable.id == self.var_five)


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


class TestCommandTagging(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        init_database(testing=True)
        db.connect()
        db.bind([Command, Tag, CommandTag])
        db.create_tables([Command, Tag, CommandTag])

    @classmethod
    def tearDownClass(cls):
        # Close database connection after all tests
        db.drop_tables([Command, Tag, CommandTag])
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
        results = self.repo.add_tags(alias="test_cmd", tags=["test_tag"])
        cmd_tag = CommandTag.get(command=cmd, tag=tag)
        self.assertTrue(isinstance(cmd_tag, CommandTag))
        self.assertEqual("test_tag", results.added[0])
        self.assertEqual(0, len(results.existing))

    def test_add_multiple_tags(self):
        cmd = Command.create(alias="test_cmd", template="echo test")
        tag1 = Tag.create(name="test_tag1", description="test_description1")
        tag2 = Tag.create(name="test_tag2", description="test_description2")
        results = self.repo.add_tags(alias="test_cmd", tags=["test_tag1", "test_tag2"])
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
        results = self.repo.add_tags(alias="test_cmd", tags=["test_tag1", "test_tag2"])
        self.assertEqual(1, len(results.added))
        self.assertEqual(1, len(results.existing))

    def test_add_tag_with_no_tags_does_nothing(self):
        cmd = Command.create(alias="test_cmd", template="echo test")
        results = self.repo.add_tags(alias="test_cmd", tags=[])
        self.assertEqual(0, len(results.added))
        self.assertEqual(0, len(results.existing))

    def test_add_tag_with_non_existent_tag_raises_exception(self):
        Command.create(alias="test_cmd", template="echo test")
        with self.assertRaises(UnknownTagError):
            self.repo.add_tags(alias="test_cmd", tags=["invalid_tag"])

    def test_add_tag_with_non_existent_command_alias_raises_exception(self):
        Tag.create(name="test_tag", description="test_description")
        with self.assertRaises(UnknownAliasError):
            self.repo.add_tags(alias="invalid_alias", tags=["test_tag"])

    def test_double_tagging_does_not_raise_error(self):
        cmd = Command.create(alias="test_cmd", template="echo test")
        tag = Tag.create(name="test_tag")
        cmd_tag = CommandTag.create(command=cmd, tag=tag)

        results = self.repo.add_tags(alias="test_cmd", tags=["test_tag"])
        self.assertTrue(isinstance(cmd_tag, CommandTag))
        self.assertEqual(0, len(results.added))
        self.assertEqual("test_tag", results.existing[0])

    def test_add_tag_is_atomic_and_no_tags_are_added_if_one_fails(self):
        cmd = Command.create(alias="test_cmd", template="echo test")
        tag = Tag.create(name="test_tag")
        with self.assertRaises(UnknownTagError):
            self.repo.add_tags(alias="test_cmd", tags=["test_tag", "invalid_tag"])
        self.assertEqual(0, CommandTag.select().count())

    def test_remove_tag(self):
        self._create_cmd_tags()
        CommandTag.get(command=self.cmd_one, tag=self.tag_one)
        self.repo.remove_tags(alias="cmd_one", tags=["tag_one"])
        with self.assertRaises(DoesNotExist):
            CommandTag.get(command=self.cmd_one, tag=self.tag_one)

    def test_remove_multiple_tags(self):
        self._create_cmd_tags()
        result = self.repo.remove_tags(alias="cmd_two", tags=["tag_one", "tag_two"])
        self.assertEqual(2, len(result.removed))
        self.assertEqual(0, len(result.not_attached))
        with self.assertRaises(DoesNotExist):
            CommandTag.get(command=self.cmd_two, tag=self.tag_one)
            CommandTag.get(command=self.cmd_two, tag=self.tag_two)

    def test_remove_tag_with_mix_of_existing_and_non_existing_tagged_commands(self):
        self._create_cmd_tags()
        result = self.repo.remove_tags(alias="cmd_one", tags=["tag_one", "tag_two"])
        self.assertEqual(1, len(result.removed))
        self.assertEqual(1, len(result.not_attached))
        with self.assertRaises(DoesNotExist):
            CommandTag.get(command=self.cmd_one, tag=self.tag_one)

    def test_remove_tag_with_no_tags_does_nothing(self):
        self._create_cmd_tags()
        result = self.repo.remove_tags(alias="cmd_one", tags=[])
        self.assertEqual(0, len(result.removed))
        self.assertEqual(0, len(result.not_attached))

    def test_remove_tag_with_non_existent_tag_raises_exception(self):
        self._create_cmd_tags()
        with self.assertRaises(UnknownTagError):
            self.repo.remove_tags(alias="cmd_one", tags=["invalid_tag"])

    def test_remove_tag_with_non_existent_command_alias_raises_exception(self):
        self._create_cmd_tags()
        with self.assertRaises(UnknownAliasError):
            self.repo.remove_tags(alias="invalid_alias", tags=["tag_one"])

    def test_removing_a_tag_twice_does_not_raise_error(self):
        self._create_cmd_tags()
        r1 = self.repo.remove_tags(alias="cmd_two", tags=["tag_one"])
        self.assertEqual(1, len(r1.removed))
        self.assertEqual(0, len(r1.not_attached))
        r2 = self.repo.remove_tags(alias="cmd_two", tags=["tag_one"])
        self.assertEqual(0, len(r2.removed))
        self.assertEqual(1, len(r2.not_attached))

    def test_remove_tag_is_atomic_and_no_tags_are_removed_if_one_fails(self):
        self._create_cmd_tags()
        with self.assertRaises(UnknownTagError):
            self.repo.remove_tags(alias="cmd_two", tags=["tag_one", "invalid_tag"])
        CommandTag.get(command=self.cmd_one, tag=self.tag_one)


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
        results = self.repo.add_tags(name="test_var", tags=["test_tag"])
        var_tag = VariableTag.get(variable=var, tag=tag)
        self.assertTrue(isinstance(var_tag, VariableTag))
        self.assertEqual("test_tag", results.added[0])
        self.assertEqual(0, len(results.existing))

    def test_add_multiple_tags(self):
        var = Variable.create(name="test_var", value="test")
        tag1 = Tag.create(name="test_tag1", description="test_description1")
        tag2 = Tag.create(name="test_tag2", description="test_description2")
        results = self.repo.add_tags(name="test_var", tags=["test_tag1", "test_tag2"])
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
        results = self.repo.add_tags(name="test_var", tags=["test_tag1", "test_tag2"])
        self.assertEqual(1, len(results.added))
        self.assertEqual(1, len(results.existing))

    def test_add_tag_with_no_tags_does_nothing(self):
        var = Variable.create(name="test_var", value="test")
        results = self.repo.add_tags(name="test_var", tags=[])
        self.assertEqual(0, len(results.added))
        self.assertEqual(0, len(results.existing))

    def test_add_tag_with_non_existent_tag_raises_exception(self):
        Variable.create(name="test_var", value="test")
        with self.assertRaises(UnknownTagError):
            self.repo.add_tags(name="test_var", tags=["invalid_tag"])

    def test_add_tag_with_non_existent_variable_name_raises_exception(self):
        Tag.create(name="test_tag", description="test_description")
        with self.assertRaises(UnknownNameError):
            self.repo.add_tags(name="invalid_name", tags=["test_tag"])

    def test_double_tagging_does_not_raise_error(self):
        var = Variable.create(name="test_var", value="test")
        tag = Tag.create(name="test_tag")
        var_tag = VariableTag.create(variable=var, tag=tag)

        results = self.repo.add_tags(name="test_var", tags=["test_tag"])
        self.assertTrue(isinstance(var_tag, VariableTag))
        self.assertEqual(0, len(results.added))
        self.assertEqual("test_tag", results.existing[0])

    def test_add_tag_is_atomic_and_no_tags_are_added_if_one_fails(self):
        var = Variable.create(name="test_var", value="test")
        tag = Tag.create(name="test_tag")
        with self.assertRaises(UnknownTagError):
            self.repo.add_tags(name="test_var", tags=["test_tag", "invalid_tag"])
        self.assertEqual(0, VariableTag.select().count())

    def test_remove_tag(self):
        self._create_var_tags()
        VariableTag.get(variable=self.var_one, tag=self.tag_one)
        self.repo.remove_tags(name="var_one", tags=["tag_one"])
        with self.assertRaises(DoesNotExist):
            VariableTag.get(variable=self.var_one, tag=self.tag_one)

    def test_remove_multiple_tags(self):
        self._create_var_tags()
        result = self.repo.remove_tags(name="var_two", tags=["tag_one", "tag_two"])
        self.assertEqual(2, len(result.removed))
        self.assertEqual(0, len(result.not_attached))
        with self.assertRaises(DoesNotExist):
            VariableTag.get(variable=self.var_two, tag=self.tag_one)
            VariableTag.get(variable=self.var_two, tag=self.tag_two)

    def test_remove_tag_with_mix_of_existing_and_non_existing_tagged_variables(self):
        self._create_var_tags()
        result = self.repo.remove_tags(name="var_one", tags=["tag_one", "tag_two"])
        self.assertEqual(1, len(result.removed))
        self.assertEqual(1, len(result.not_attached))
        with self.assertRaises(DoesNotExist):
            VariableTag.get(variable=self.var_one, tag=self.tag_one)

    def test_remove_tag_with_no_tags_does_nothing(self):
        self._create_var_tags()
        result = self.repo.remove_tags(name="var_one", tags=[])
        self.assertEqual(0, len(result.removed))
        self.assertEqual(0, len(result.not_attached))

    def test_remove_tag_with_non_existent_tag_raises_exception(self):
        self._create_var_tags()
        with self.assertRaises(UnknownTagError):
            self.repo.remove_tags(name="var_one", tags=["invalid_tag"])

    def test_remove_tag_with_non_existent_variable_name_raises_exception(self):
        self._create_var_tags()
        with self.assertRaises(UnknownNameError):
            self.repo.remove_tags(name="invalid_name", tags=["tag_one"])

    def test_removing_a_tag_twice_does_not_raise_error(self):
        self._create_var_tags()
        r1 = self.repo.remove_tags(name="var_two", tags=["tag_one"])
        self.assertEqual(1, len(r1.removed))
        self.assertEqual(0, len(r1.not_attached))
        r2 = self.repo.remove_tags(name="var_two", tags=["tag_one"])
        self.assertEqual(0, len(r2.removed))
        self.assertEqual(1, len(r2.not_attached))

    def test_remove_tag_is_atomic_and_no_tags_are_removed_if_one_fails(self):
        self._create_var_tags()
        with self.assertRaises(UnknownTagError):
            self.repo.remove_tags(name="var_two", tags=["tag_one", "invalid_tag"])
        VariableTag.get(variable=self.var_one, tag=self.tag_one)
