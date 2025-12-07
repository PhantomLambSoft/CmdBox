import sys
from pathlib import Path
from peewee import SqliteDatabase

DB_PATH = Path.home() / '.cmdbox' / 'cmdbox.db'
DB_PATH.parent.mkdir(parents=True, exist_ok=True)


def init_database() -> SqliteDatabase:
    """
    Initializes and creates/connects a SqliteDatabase instance based on the environment
    in which the application is running.  For a testing environment, an in-memory database
    is created.  For a real-world environment, a persistent file-based database is created.
    Returns:
        SqliteDatabase: An initialized SqliteDatabase instance.
    """
    database = SqliteDatabase(None)
    if 'unittests' in sys.modules:
        database.init(':memory:')
    else:
        database.init(DB_PATH)
    return database
