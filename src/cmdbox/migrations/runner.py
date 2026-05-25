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

    Args:
        db_path: A string representing the path to the SQLite database file.

    Raises:
        MigrationError: If the database version is greater than the application's
            current version, indicating the application is outdated and needs updating.
    """
    db = SqliteDatabase(db_path)
    db.connect()
    version = get_user_version(db)
    db.close()

    current_version = get_current_version()

    if version == 0:
        db.connect()
        set_user_version(db, current_version)
        db.close()
        log.debug("Fresh database stamped at v%d", current_version)
        return

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
        migrations (dict): A dictionary where keys are version numbers and values are callable
            migration functions. Each function must accept the current database path and the
            path to the new database as string arguments.
        timestamp (str): A timestamp used for creating backups.

    Raises:
        MigrationError: If no migration function exists for the current version.
    """
    if version not in migrations:
        raise MigrationError(
            f"No migration found for v{version} to v{version + 1}. The installation may be corrupt."
        )

    new_path = path.parent / (path.name + ".new")

    log.info("Migrating database v%d -> v%d...", version, version + 1)
    migrations[version](str(path), str(new_path))

    new_db = SqliteDatabase(str(new_path))
    new_db.connect()
    set_user_version(new_db, version + 1)
    new_db.close()

    backup_db(path, version, timestamp)
    new_path.rename(path)

    log.info("Migration to v%d complete.", version + 1)


def load_migrations() -> dict:
    import cmdbox.migrations as migrations_pkg

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


def backup_db(db_path: Path, version: int, timestamp: str) -> Path:
    backup_dir = db_path.parent / "backups"
    backup_dir.mkdir(exist_ok=True)
    bak_path = backup_dir / f"{db_path.stem}_v{version}_{timestamp}.bak"
    db_path.rename(bak_path)
    log.debug("Backed up v%d database to %s", version, bak_path.name)
    return bak_path
