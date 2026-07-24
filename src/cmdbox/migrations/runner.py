import importlib
import logging
import pkgutil
from pathlib import Path
from datetime import datetime

from peewee import SqliteDatabase

from cmdbox.migrations.errors import MigrationError

log = logging.getLogger(__name__)


def ensure_migrated(db_path: str) -> None:
    """
    Ensures the database is migrated to the current version. Connects to the SQLite database,
    checks the user version, and applies any necessary migrations to bring the database
    to the current version. If the database version is ahead of the current application
    version, an error is raised.

    A `user_version` of 0 is ambiguous on its own: SQLite reports 0 for both a brand-new,
    genuinely empty database and an existing pre-migration-system database that was never
    stamped. These are distinguished by checking whether any user tables exist. A fresh
    database is stamped directly at the current version with no migrations applied. An
    existing, unstamped database is treated as v1 and migrated forward normally.

    Args:
        db_path: A string representing the path to the SQLite database file.

    Raises:
        MigrationError: If the database version is greater than the application's
            current version, indicating the application is outdated and needs updating.
    """
    db = SqliteDatabase(db_path)
    db.connect()
    version = get_user_version(db)
    fresh = is_empty_database(db)
    db.close()

    current_version = get_current_version()

    if version == 0 and fresh:
        db.connect()
        set_user_version(db, current_version)
        db.close()
        log.debug("Fresh database stamped at v%d", current_version)
        return

    if version == 0 and not fresh:
        # Existing database predates the migration system, treat it as the baseline.
        version = 1
        log.debug("Unstamped existing database detected, treating as v1")

    if version == current_version:
        return

    if version > current_version:
        raise MigrationError(
            f"Database version {version} is ahead of the application (v{current_version}). Please update CmdBox."
        )

    migrations = load_migrations()
    path = Path(db_path)

    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")

    for v in range(version, current_version):
        migrate(v, path, migrations, timestamp)


def migrate(version: int, path: Path, migrations: dict, timestamp: str) -> None:
    """
    Migrates a database from one version to the next using the provided migration mapping.

    Args:
        version (int): The current version of the database.
        path (Path): The file path of the database to migrate.
        migrations (dict): A dictionary where keys are the version a migration module
            produces (its `VERSION` constant) and values are callable migration functions.
            Each function must accept the current database path and the path to the new
            database as string arguments.
        timestamp (str): A timestamp used for creating backups.

    Raises:
        MigrationError: If no migration function exists that produces `version + 1`.
    """
    target = version + 1

    if target not in migrations:
        raise MigrationError(
            f"No migration found for v{version} to v{target}. The installation may be corrupt."
        )

    new_path = path.parent / (path.name + ".new")
    if new_path.exists():
        log.debug("removing stale .new file from previous migration attempt: %s", new_path)
        new_path.unlink()

    log.info("Migrating database v%d -> v%d...", version, target)
    migrations[target](str(path), str(new_path))

    new_db = SqliteDatabase(str(new_path))
    new_db.connect()
    set_user_version(new_db, target)
    new_db.close()

    backup_db(path, version, timestamp)
    new_path.rename(path)

    log.info("Migration to v%d complete.", target)


def load_migrations() -> dict:
    import cmdbox.migrations.versions as migrations_pkg

    migrations = {}
    for _, name, _ in pkgutil.iter_modules(migrations_pkg.__path__):
        if not (name.startswith("m") and name[1:].isdigit()):
            continue
        module = importlib.import_module(f"cmdbox.migrations.versions.{name}")
        migrations[module.VERSION] = module.migrate
    return migrations


def get_current_version() -> int:
    migrations = load_migrations()
    return max(migrations.keys(), default=1)


def get_user_version(db: SqliteDatabase) -> int:
    return db.execute_sql("PRAGMA user_version").fetchone()[0]


def set_user_version(db: SqliteDatabase, version: int) -> None:
    db.execute_sql(f"PRAGMA user_version = {version}")


def is_empty_database(db: SqliteDatabase) -> bool:
    """
    Determines whether a database has no user-defined tables. Used to distinguish a
    genuinely fresh database from an existing one that was never stamped with a
    user_version, both of which report a user_version of 0.

    Args:
        db (SqliteDatabase): An open connection to the database to inspect.

    Returns:
        bool: True if no user tables exist, False otherwise.
    """
    cursor = db.execute_sql(
        "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
    )
    return cursor.fetchone()[0] == 0


def backup_db(db_path: Path, version: int, timestamp: str) -> Path:
    backup_dir = db_path.parent / "backups"
    backup_dir.mkdir(exist_ok=True)
    bak_path = backup_dir / f"{db_path.stem}_v{version}_{timestamp}.bak"
    db_path.rename(bak_path)
    log.debug("Backed up v%d database to %s", version, bak_path.name)
    return bak_path
