from datetime import datetime

from peewee import Model, CharField, IntegerField, DateTimeField, TextField
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
    last_used = DateTimeField()


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
