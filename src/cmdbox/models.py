from datetime import datetime

from peewee import (
    Model,
    CharField,
    IntegerField,
    DateTimeField,
    TextField,
    ForeignKeyField,
)
from cmdbox.database import init_database


db = init_database()


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


class Command(BaseModel):
    """
    Represents a command model.

    This class provides a structure for storing and managing command-related data,
    including an alias for the command, its template, timestamps for its creation
    and last update, and a counter for how many times the command has been used.

    Attributes:
        alias (CharField): Unique identifier for the command.
        template (CharField): The associated template string for the command.
        date_created (DateTimeField): Timestamp indicating when the command was created.
        last_updated (DateTimeField): Timestamp indicating when the command was last updated.
        used (IntegerField): Counter representing how many times the command has been used.
        last_used (DateTimeField): Timestamp indicating when the command was last used.
    """

    alias = CharField(unique=True)
    template = TextField()
    description = TextField()
    date_created = DateTimeField(default=datetime.now)
    last_updated = DateTimeField(default=datetime.now)
    used = IntegerField(default=0)
    last_used = DateTimeField(null=True, default=None)


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
    description = TextField()
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

    command = ForeignKeyField(Command, backref='tags')
    tag = ForeignKeyField(Tag, backref='commands')
    date_created = DateTimeField(default=datetime.now)


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

    variable = ForeignKeyField(Variable, backref='tags')
    tag = ForeignKeyField(Tag, backref='variables')
    date_created = DateTimeField(default=datetime.now)
