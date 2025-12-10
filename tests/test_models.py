import unittest
from datetime import datetime

from cmdbox.database import db, init_database
from cmdbox.models import Command, Variable, Tag, CommandTag, VariableTag


class TestCommandModel(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        init_database(testing=True)
        db.connect()
        db.bind([Command])
        db.create_tables([Command])

    @classmethod
    def tearDownClass(cls):
        # Close database connection after all tests
        db.drop_tables([Command])
        db.close()

    def setUp(self):
        # Clean the table for each test
        Command.delete().execute()

    def test_create_command(self):
        """Test creating a new Command instance."""
        command = Command.create(
            alias="deploy",
            template="echo Deploying",
            description="Deployment command",
        )
        cmd = Command.get(alias=command.alias)
        self.assertEqual(Command.select().count(), 1)
        self.assertEqual(cmd.template, command.template)
        self.assertEqual(cmd.description, command.description)

    def test_unique_alias(self):
        """Test that 'alias' field must be unique."""
        Command.create(
            alias="build",
            template="echo Building",
            description="Build command",
            last_used=datetime.now(),
        )
        with self.assertRaises(Exception):
            Command.create(
                alias="build",
                template="echo Build again",
                description="Duplicate alias",
                last_used=datetime.now(),
            )

    def test_update_command_usage(self):
        """Test updating the usage count and last used timestamp."""
        init_date = datetime(2025, 12, 6)
        command = Command.create(
            alias="test",
            template="pytest",
            description="Run tests",
            used=5,
            last_used=init_date,
        )
        command.used += 1
        command.last_used = datetime.now()
        command.save()

        updated_command = Command.get(Command.alias == "test")
        self.assertEqual(updated_command.used, 6)
        self.assertNotEqual(updated_command.last_used, init_date)

    def test_update_unique_alias(self):
        """Test that 'alias' uniqueness is enforced on update."""
        Command.create(
            alias="build",
            template="echo Building",
            description="Build command",
            last_used=datetime.now(),
        )
        test_command = Command.create(
            alias="run",
            template="echo run command",
            description="Run command",
            last_used=datetime.now(),
        )
        with self.assertRaises(Exception):
            test_command.alias = "build"
            test_command.save()

    def test_delete_command(self):
        """Test deleting a command."""
        command = Command.create(
            alias="remove-me",
            template="echo Removing",
            description="Temporary command",
            last_used=datetime.now(),
        )
        self.assertEqual(Command.select().count(), 1)
        Command.delete().where(Command.id == command.id).execute()
        self.assertEqual(Command.select().count(), 0)


class TestVariableModule(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        init_database(testing=True)
        db.connect()
        db.bind([Variable])
        db.create_tables([Variable])

    @classmethod
    def tearDownClass(cls):
        # Close database connection after all tests
        db.close()

    def setUp(self):
        # Clean the table for each test
        Variable.delete().execute()

    def test_create_variable(self):
        """Test creating a new variable instance."""
        variable = Variable.create(
            name="test_variable",
            value="test_value",
        )
        var = Variable.get(name=variable.name)
        self.assertEqual(Variable.select().count(), 1)
        self.assertEqual(var.value, variable.value)

    def test_unique_name(self):
        """Test that the 'name' field must be unique."""
        Variable.create(
            name="test_variable",
            value="test_value",
        )
        with self.assertRaises(Exception):
            Variable.create(
                name="test_variable",
                value="test_value_2",
            )

    def test_update_variable(self):
        """Test updating a variable."""
        var = Variable.create(
            name="test_variable",
            value="test_value",
        )
        var.value = "new_value"
        var.save()
        updated_var = Variable.get(Variable.name == "test_variable")
        self.assertEqual(updated_var.value, "new_value")

    def test_update_unique_name(self):
        """Test that 'name' uniqueness is enforced on update."""
        Variable.create(
            name="test_variable",
            value="test_value",
        )
        var = Variable.create(
            name="new_name",
            value="new_value",
        )
        with self.assertRaises(Exception):
            var.name = "test_variable"
            var.save()

    def test_delete_variable(self):
        """Test deleting a variable."""
        var = Variable.create(
            name="test_variable",
            value="test_value",
        )
        self.assertEqual(Variable.select().count(), 1)
        Variable.delete().where(Variable.id == var.id).execute()
        self.assertEqual(Variable.select().count(), 0)


class TestTagModule(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        init_database(testing=True)
        db.connect()
        db.bind([Tag])
        db.create_tables([Tag])

    @classmethod
    def tearDownClass(cls):
        db.close()

    def setUp(self):
        Tag.delete().execute()

    def test_create_tag(self):
        tag = Tag.create(name="test_tag", description="Test tag")
        self.assertEqual(Tag.select().count(), 1)
        self.assertEqual(tag.name, "test_tag")
        self.assertEqual(tag.description, "Test tag")

    def test_unique_name(self):
        Tag.create(name="test_tag", description="Test tag")
        with self.assertRaises(Exception):
            Tag.create(name="test_tag", description="Test tag 2")

    def test_update_tag(self):
        tag = Tag.create(name="test_tag", description="Test tag")
        tag.description = "Updated tag"
        tag.save()
        updated_tag = Tag.get(Tag.name == "test_tag")
        self.assertEqual(updated_tag.description, "Updated tag")

    def test_tag_update_unique_name(self):
        Tag.create(name="test_tag", description="Test tag")
        tag = Tag.create(name="new_tag", description="New tag")
        with self.assertRaises(Exception):
            tag.name = "test_tag"
            tag.save()

    def test_delete_tag(self):
        tag = Tag.create(name="test_tag", description="Test tag")
        self.assertEqual(Tag.select().count(), 1)
        Tag.delete().where(Tag.id == tag.id).execute()
        self.assertEqual(Tag.select().count(), 0)


class TestCommandTagModule(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        init_database(testing=True)
        db.connect()
        db.bind([CommandTag, Command, Tag])
        db.create_tables([CommandTag, Command, Tag])

    @classmethod
    def tearDownClass(cls):
        db.close()

    def setUp(self):
        CommandTag.delete().execute()
        Command.delete().execute()
        Tag.delete().execute()
        self.command = Command.create(
            alias="test", template="echo test", description="Test command"
        )
        self.tag = Tag.create(name="test_tag", description="Test tag")

    def test_create_command_tag(self):
        CommandTag.create(command=self.command, tag=self.tag)
        self.assertEqual(CommandTag.select().count(), 1)

    def test_delete_command_tag(self):
        CommandTag.create(command=self.command, tag=self.tag)
        self.assertEqual(CommandTag.select().count(), 1)
        CommandTag.delete().where(
            (CommandTag.command == self.command) & (CommandTag.tag == self.tag)
        ).execute()
        self.assertEqual(CommandTag.select().count(), 0)

    def test_command_tag_has_date_created(self):
        command_tag = CommandTag.create(command=self.command, tag=self.tag)
        self.assertIsNotNone(command_tag.date_created)


class TestVariableTagModule(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        init_database(testing=True)
        db.connect()
        db.bind([VariableTag, Variable, Tag])
        db.create_tables([VariableTag, Variable, Tag])

    @classmethod
    def tearDownClass(cls):
        db.close()

    def setUp(self):
        VariableTag.delete().execute()
        Variable.delete().execute()
        Tag.delete().execute()
        self.variable = Variable.create(name="test_variable", value="test_value")
        self.tag = Tag.create(name="test_tag", description="Test tag")

    def test_create_variable_tag(self):
        VariableTag.create(variable=self.variable, tag=self.tag)
        self.assertEqual(VariableTag.select().count(), 1)

    def test_delete_variable_tag(self):
        VariableTag.create(variable=self.variable, tag=self.tag)
        self.assertEqual(VariableTag.select().count(), 1)
        VariableTag.delete().where(
            (VariableTag.variable == self.variable) & (VariableTag.tag == self.tag)
        ).execute()
        self.assertEqual(VariableTag.select().count(), 0)

    def test_variable_tag_has_date_created(self):
        variable_tag = VariableTag.create(variable=self.variable, tag=self.tag)
        self.assertIsNotNone(variable_tag.date_created)
