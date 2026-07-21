from datetime import datetime

from peewee import (
    Model,
    CharField,
    IntegerField,
    DateTimeField,
    TextField,
    ForeignKeyField,
)
from cmdbox.database import db


class BaseModel(Model):
    """
    BaseModel serves as the foundation for database models.

    This class extends the 'Model' class from the Peewee library, providing a
    base for all database models within the application. It centralizes the
    database connection, ensuring consistency across all derived models.

    Attributes:
        Meta (class): Configuration for the database connection, linking models
                      to the shared database instance.
    """

    class Meta:
        database = db


class Profile(BaseModel):
    """
    Represents a named profile, a container that commands and variables
    can be scoped to, and settings can be scoped to via a per-profile
    settings file.

    Attributes:
        name (CharField): The unique name of the profile.
        description (CharField): An optional description of the profile.
            Defaults to None.
        date_created (DateTimeField): The date and time when the profile
            was created. Defaults to the current date and time.
        last_used (DateTimeField): The date and time when the profile
            was last used. Defaults to None.
    """

    name = CharField(unique=True)
    description = CharField(null=True, default=None)
    date_created = DateTimeField(default=datetime.now)
    last_used = DateTimeField(null=True, default=None)


class ProfileState(BaseModel):
    """
    Represents the state of active profiles for various system configurations.

    The ProfileState class is used to track and manage the active profiles for
    commands, variables, and settings. It is designed to ensure consistency by
    keeping references to the currently active profiles within a system. This
    class extends the BaseModel to integrate seamlessly with the application's
    database logic and ORM tools.

    Attributes:
        active_command_profile (ForeignKeyField): Reference to the active command
            profile. Prevents deletion when referenced by this field.
        active_variable_profile (ForeignKeyField): Reference to the active variable
            profile. Prevents deletion when referenced by this field.
        active_settings_profile (ForeignKeyField): Reference to the active settings
            profile. Prevents deletion when referenced by this field.
    """

    active_command_profile = ForeignKeyField(Profile, backref="+", on_delete="RESTRICT")
    active_variable_profile = ForeignKeyField(
        Profile, backref="+", on_delete="RESTRICT"
    )
    active_settings_profile = ForeignKeyField(
        Profile, backref="+", on_delete="RESTRICT"
    )


class Command(BaseModel):
    """
    Represents a command model.

    This class provides a structure for storing and managing command-related data,
    including an alias for the command, its template, timestamps for its creation
    and last update, and a counter for how many times the command has been used.

    Attributes:
        alias (CharField): Unique identifier for the command.
        template (CharField): The associated template string for the command.
        description (CharField): A description of the command and what it does.
        cwd (CharField): The current working directory for the command execution.
        shell (CharField): The shell to use for command execution.
        env (TextField): Environment variables to set for the command execution.
        timeout (IntegerField): Number of seconds before the process is killed.
        date_created (DateTimeField): Timestamp indicating when the command was created.
        last_updated (DateTimeField): Timestamp indicating when the command was last updated.
        used (IntegerField): Counter representing how many times the command has been used.
        last_used (DateTimeField): Timestamp indicating when the command was last used.
    """

    alias = CharField(unique=True)
    template = TextField()
    description = TextField(null=True, default=None)
    cwd = CharField(null=True, default=None)
    shell = CharField(null=True, default=None)
    env = TextField(null=True, default=None)
    timeout = IntegerField(null=True, default=None)
    date_created = DateTimeField(default=datetime.now)
    last_updated = DateTimeField(default=datetime.now)
    used = IntegerField(default=0)
    last_used = DateTimeField(null=True, default=None)
    profile = ForeignKeyField(Profile, backref="commands", on_delete="CASCADE")


class Variable(BaseModel):
    """
    Represents a variable model.

    A variable model is a key-value pair that can be stored and recalled in commands
    and other variables.

    Attributes:
        name (CharField): The unique name of the configuration variable.
        value (CharField): The value associated with the configuration variable.
    """

    name = CharField(unique=True)
    value = CharField()
    date_created = DateTimeField(default=datetime.now)
    last_updated = DateTimeField(default=datetime.now)
    profile = ForeignKeyField(Profile, backref="variables", on_delete="CASCADE")


class Tag(BaseModel):
    """
    Represents a Tag model with attributes for tagging, description, and timestamps.

    This class is designed to store tag information as a name and description to allow
    users to organize their commands and variables into categories and to give them a
     convenient way to search for and filter them later.

    Attributes:
        name (CharField): A unique name identifier for the tag.
        description (TextField): A detailed textual description of the tag.
        date_created (DateTimeField): The timestamp indicating when the tag was created.
        last_updated (DateTimeField): The timestamp indicating the last update for the tag.
    """

    name = CharField(unique=True)
    description = TextField(null=True, default=None)
    date_created = DateTimeField(default=datetime.now)
    last_updated = DateTimeField(default=datetime.now)


class CommandTag(BaseModel):
    """
    Represents the relationship between commands and tags.

    This class is used as an intermediary table to establish a many-to-many
    relationship between a command and a tag.

    Attributes:
        command (ForeignKeyField): Refers to a command associated with a tag.
        tag (ForeignKeyField): Refers to a tag associated with a command.
        date_created (DateTimeField): DateTimeField indicating when the relationship was created.
    """

    command = ForeignKeyField(Command, backref="tags", on_delete="CASCADE")
    tag = ForeignKeyField(Tag, backref="commands", on_delete="CASCADE")
    date_created = DateTimeField(default=datetime.now)

    class Meta:
        indexes = ((("command", "tag"), True),)


class VariableTag(BaseModel):
    """
    Represents a many-to-many relationship between Variable and Tag models.

    This class serves as an intermediary table to establish a many-to-many relationship
    between a variable and a tag.

    Attributes:
        variable: (ForeignKeyField) Refers to a variable associated with a tag.
        tag: (ForeignKeyField) Refers to a tag associated with a variable.
        date_created (DateTimeField): DateTimeField indicating when the relationship was created.
    """

    variable = ForeignKeyField(Variable, backref="tags", on_delete="CASCADE")
    tag = ForeignKeyField(Tag, backref="variables", on_delete="CASCADE")
    date_created = DateTimeField(default=datetime.now)

    class Meta:
        indexes = ((("variable", "tag"), True),)


class CommandHistory(BaseModel):

    id = TextField(primary_key=True)  # uuid4().hex - 32 char, no dashes
    alias = CharField()  # alias as invoked, not a FK
    template = TextField()  # raw template at time of run
    resolved = TextField()  # fully resolved command string
    variables_used = TextField(null=True)  # JSON: {"name": "Homer"} or null
    exit_code = IntegerField(null=True)
    ran_at = DateTimeField(default=datetime.now)

    class Meta:
        table_name = "command_history"
        indexes = (
            (("alias",), False),
            (("ran_at",), False),
        )


ALL_MODELS = [
    Profile,
    ProfileState,
    Command,
    Variable,
    Tag,
    CommandTag,
    VariableTag,
    CommandHistory,
]
