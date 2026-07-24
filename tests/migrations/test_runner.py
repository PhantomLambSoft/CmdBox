import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import MagicMock, call, patch

from cmdbox.migrations import runner
from cmdbox.migrations.errors import MigrationError


class TestMigrationRunner(unittest.TestCase):

    @patch("cmdbox.migrations.runner.migrate")
    @patch("cmdbox.migrations.runner.datetime")
    @patch("cmdbox.migrations.runner.load_migrations")
    @patch("cmdbox.migrations.runner.set_user_version")
    @patch("cmdbox.migrations.runner.get_current_version")
    @patch("cmdbox.migrations.runner.get_user_version")
    @patch("cmdbox.migrations.runner.is_empty_database")
    @patch("cmdbox.migrations.runner.SqliteDatabase")
    def test_ensure_migrated_stamps_fresh_database(
        self,
        mock_sqlite_db,
        mock_is_empty_database,
        mock_get_user_version,
        mock_get_current_version,
        mock_set_user_version,
        mock_load_migrations,
        mock_datetime,
        mock_migrate,
    ):
        db = MagicMock()
        mock_sqlite_db.return_value = db
        mock_get_user_version.return_value = 0
        mock_is_empty_database.return_value = True
        mock_get_current_version.return_value = 4
        mock_datetime.now.return_value.strftime.return_value = "20260101_120000"

        runner.ensure_migrated("db.sqlite")

        mock_is_empty_database.assert_called_once_with(db)
        mock_sqlite_db.assert_called_once_with("db.sqlite")
        self.assertEqual(2, db.connect.call_count)
        self.assertEqual(2, db.close.call_count)
        mock_set_user_version.assert_called_once_with(db, 4)
        mock_load_migrations.assert_not_called()
        mock_migrate.assert_not_called()

    @patch("cmdbox.migrations.runner.migrate")
    @patch("cmdbox.migrations.runner.load_migrations")
    @patch("cmdbox.migrations.runner.get_current_version")
    @patch("cmdbox.migrations.runner.get_user_version")
    @patch("cmdbox.migrations.runner.SqliteDatabase")
    def test_ensure_migrated_noop_when_versions_match(
        self,
        mock_sqlite_db,
        mock_get_user_version,
        mock_get_current_version,
        mock_load_migrations,
        mock_migrate,
    ):
        db = MagicMock()
        mock_sqlite_db.return_value = db
        mock_get_user_version.return_value = 3
        mock_get_current_version.return_value = 3

        runner.ensure_migrated("db.sqlite")

        self.assertEqual(1, db.connect.call_count)
        self.assertEqual(1, db.close.call_count)
        mock_load_migrations.assert_not_called()
        mock_migrate.assert_not_called()

    @patch("cmdbox.migrations.runner.get_current_version")
    @patch("cmdbox.migrations.runner.get_user_version")
    @patch("cmdbox.migrations.runner.SqliteDatabase")
    def test_ensure_migrated_raises_when_database_is_ahead(
        self,
        mock_sqlite_db,
        mock_get_user_version,
        mock_get_current_version,
    ):
        db = MagicMock()
        mock_sqlite_db.return_value = db
        mock_get_user_version.return_value = 8
        mock_get_current_version.return_value = 5

        with self.assertRaises(MigrationError) as context:
            runner.ensure_migrated("db.sqlite")

        self.assertEqual(
            "Database version 8 is ahead of the application (v5). Please update CmdBox.",
            str(context.exception),
        )
        self.assertEqual(1, db.connect.call_count)
        self.assertEqual(1, db.close.call_count)

    @patch("cmdbox.migrations.runner.migrate")
    @patch("cmdbox.migrations.runner.datetime")
    @patch("cmdbox.migrations.runner.load_migrations")
    @patch("cmdbox.migrations.runner.get_current_version")
    @patch("cmdbox.migrations.runner.get_user_version")
    @patch("cmdbox.migrations.runner.SqliteDatabase")
    def test_ensure_migrated_runs_all_required_steps(
        self,
        mock_sqlite_db,
        mock_get_user_version,
        mock_get_current_version,
        mock_load_migrations,
        mock_datetime,
        mock_migrate,
    ):
        db = MagicMock()
        mock_sqlite_db.return_value = db
        mock_get_user_version.return_value = 1
        mock_get_current_version.return_value = 4
        mock_load_migrations.return_value = {
            1: MagicMock(),
            2: MagicMock(),
            3: MagicMock(),
        }
        mock_datetime.now.return_value.strftime.return_value = "20260101_131500"

        runner.ensure_migrated("C:\\tmp\\db.sqlite")

        expected_path = Path("C:\\tmp\\db.sqlite")
        self.assertEqual(
            [
                call(
                    1,
                    expected_path,
                    mock_load_migrations.return_value,
                    "20260101_131500",
                ),
                call(
                    2,
                    expected_path,
                    mock_load_migrations.return_value,
                    "20260101_131500",
                ),
                call(
                    3,
                    expected_path,
                    mock_load_migrations.return_value,
                    "20260101_131500",
                ),
            ],
            mock_migrate.call_args_list,
        )

    def test_migrate_raises_for_missing_migration(self):
        with self.assertRaises(MigrationError) as context:
            runner.migrate(
                5, Path("db.sqlite"), migrations={}, timestamp="20260101_100000"
            )

        self.assertEqual(
            "No migration found for v5 to v6. The installation may be corrupt.",
            str(context.exception),
        )

    @patch("cmdbox.migrations.runner.backup_db")
    @patch("cmdbox.migrations.runner.set_user_version")
    @patch("cmdbox.migrations.runner.SqliteDatabase")
    @patch("cmdbox.migrations.runner.Path.rename")
    def test_migrate_executes_migration_and_replaces_database(
        self, mock_rename, mock_sqlite_db, mock_set_user_version, mock_backup_db
    ):
        original_db = Path("C:\\tmp\\commands.sqlite")
        migration_fn = MagicMock()
        new_db = MagicMock()
        mock_sqlite_db.return_value = new_db

        runner.migrate(
            1,
            original_db,
            migrations={2: migration_fn},
            timestamp="20260101_142500",
        )

        expected_new_path = original_db.parent / "commands.sqlite.new"
        migration_fn.assert_called_once_with(str(original_db), str(expected_new_path))
        mock_sqlite_db.assert_called_once_with(str(expected_new_path))
        new_db.connect.assert_called_once_with()
        mock_set_user_version.assert_called_once_with(new_db, 2)
        new_db.close.assert_called_once_with()
        mock_backup_db.assert_called_once_with(original_db, 1, "20260101_142500")
        mock_rename.assert_called_once_with(original_db)

    @patch("cmdbox.migrations.runner.importlib.import_module")
    @patch("cmdbox.migrations.runner.pkgutil.iter_modules")
    def test_load_migrations_loads_only_valid_version_modules(
        self, mock_iter_modules, mock_import_module
    ):
        mock_iter_modules.return_value = [
            (None, "m1", False),
            (None, "bad", False),
            (None, "m02", False),
            (None, "m2x", False),
        ]
        mod1 = SimpleNamespace(VERSION=1, migrate=MagicMock())
        mod2 = SimpleNamespace(VERSION=2, migrate=MagicMock())
        mock_import_module.side_effect = [mod1, mod2]

        migrations = runner.load_migrations()

        self.assertEqual({1: mod1.migrate, 2: mod2.migrate}, migrations)
        self.assertEqual(
            [
                call("cmdbox.migrations.versions.m1"),
                call("cmdbox.migrations.versions.m02"),
            ],
            mock_import_module.call_args_list,
        )

    @patch("cmdbox.migrations.runner.load_migrations")
    def test_get_current_version_uses_highest_migration_version(
        self, mock_load_migrations
    ):
        mock_load_migrations.return_value = {
            3: MagicMock(),
            1: MagicMock(),
            7: MagicMock(),
        }

        version = runner.get_current_version()

        self.assertEqual(7, version)

    @patch("cmdbox.migrations.runner.load_migrations")
    def test_get_current_version_defaults_to_one_when_no_migrations(
        self, mock_load_migrations
    ):
        mock_load_migrations.return_value = {}

        version = runner.get_current_version()

        self.assertEqual(1, version)

    def test_get_user_version_reads_pragma_value(self):
        db = MagicMock()
        cursor = MagicMock()
        cursor.fetchone.return_value = (11,)
        db.execute_sql.return_value = cursor

        version = runner.get_user_version(db)

        self.assertEqual(11, version)
        db.execute_sql.assert_called_once_with("PRAGMA user_version")

    def test_set_user_version_writes_pragma_value(self):
        db = MagicMock()

        runner.set_user_version(db, 9)

        db.execute_sql.assert_called_once_with("PRAGMA user_version = 9")

    def test_backup_db_creates_backup_file_and_returns_path(self):
        db_path = MagicMock()
        db_path.stem = "state"
        backup_dir = MagicMock()
        bak_path = MagicMock()
        db_path.parent.__truediv__.return_value = backup_dir
        backup_dir.__truediv__.return_value = bak_path

        result = runner.backup_db(db_path, 4, "20260101_153000")

        db_path.parent.__truediv__.assert_called_once_with("backups")
        backup_dir.mkdir.assert_called_once_with(exist_ok=True)
        backup_dir.__truediv__.assert_called_once_with("state_v4_20260101_153000.bak")
        db_path.rename.assert_called_once_with(bak_path)
        self.assertEqual(bak_path, result)

    def test_backup_db_works_when_backup_directory_already_exists(self):
        db_path = MagicMock()
        db_path.stem = "state"
        backup_dir = MagicMock()
        bak_path = MagicMock()
        db_path.parent.__truediv__.return_value = backup_dir
        backup_dir.__truediv__.return_value = bak_path

        result = runner.backup_db(db_path, 5, "20260101_160100")

        db_path.parent.__truediv__.assert_called_once_with("backups")
        backup_dir.mkdir.assert_called_once_with(exist_ok=True)
        backup_dir.__truediv__.assert_called_once_with("state_v5_20260101_160100.bak")
        db_path.rename.assert_called_once_with(bak_path)
        self.assertEqual(bak_path, result)

    @patch("cmdbox.migrations.runner.migrate")
    @patch("cmdbox.migrations.runner.datetime")
    @patch("cmdbox.migrations.runner.load_migrations")
    @patch("cmdbox.migrations.runner.get_current_version")
    @patch("cmdbox.migrations.runner.get_user_version")
    @patch("cmdbox.migrations.runner.is_empty_database")
    @patch("cmdbox.migrations.runner.SqliteDatabase")
    def test_ensure_migrated_treats_unstamped_existing_database_as_v1(
        self,
        mock_sqlite_db,
        mock_is_empty_database,
        mock_get_user_version,
        mock_get_current_version,
        mock_load_migrations,
        mock_datetime,
        mock_migrate,
    ):
        db = MagicMock()
        mock_sqlite_db.return_value = db
        mock_get_user_version.return_value = 0
        mock_is_empty_database.return_value = False
        mock_get_current_version.return_value = 2
        mock_load_migrations.return_value = {1: MagicMock()}
        mock_datetime.now.return_value.strftime.return_value = "20260101_170000"

        runner.ensure_migrated("db.sqlite")

        mock_is_empty_database.assert_called_once_with(db)
        mock_migrate.assert_called_once_with(
            1, Path("db.sqlite"), mock_load_migrations.return_value, "20260101_170000"
        )

    def test_is_empty_database_returns_true_when_no_user_tables(self):
        db = MagicMock()
        cursor = MagicMock()
        cursor.fetchone.return_value = (0,)
        db.execute_sql.return_value = cursor

        result = runner.is_empty_database(db)

        self.assertTrue(result)
        db.execute_sql.assert_called_once_with(
            "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'"
        )

    def test_is_empty_database_returns_false_when_user_tables_exist(self):
        db = MagicMock()
        cursor = MagicMock()
        cursor.fetchone.return_value = (3,)
        db.execute_sql.return_value = cursor

        result = runner.is_empty_database(db)

        self.assertFalse(result)

    @patch("cmdbox.migrations.runner.importlib.import_module")
    @patch("cmdbox.migrations.runner.pkgutil.iter_modules")
    def test_load_migrations_scans_versions_subpackage_not_top_level(
        self, mock_iter_modules, mock_import_module
    ):
        # Deliberately not mocking the versions package itself: this is the
        # exact bug that shipped, iter_modules was called against the top-level
        # migrations package instead of migrations.versions, and a fully mocked
        # test can't distinguish "scanned the right path" from "scanned nothing."
        import cmdbox.migrations.versions as versions_pkg

        mock_iter_modules.return_value = []

        runner.load_migrations()

        scanned_path = mock_iter_modules.call_args[0][0]
        self.assertEqual(list(versions_pkg.__path__), list(scanned_path))

    @patch("cmdbox.migrations.runner.backup_db")
    @patch("cmdbox.migrations.runner.set_user_version")
    @patch("cmdbox.migrations.runner.SqliteDatabase")
    @patch("cmdbox.migrations.runner.Path.rename")
    @patch("cmdbox.migrations.runner.Path.unlink")
    @patch("cmdbox.migrations.runner.Path.exists")
    def test_migrate_removes_stale_new_file_before_migrating(
        self,
        mock_exists,
        mock_unlink,
        mock_rename,
        mock_sqlite_db,
        mock_set_user_version,
        mock_backup_db,
    ):
        original_db = Path("C:\\tmp\\commands.sqlite")
        migration_fn = MagicMock()
        new_db = MagicMock()
        mock_sqlite_db.return_value = new_db
        mock_exists.return_value = True

        runner.migrate(
            1,
            original_db,
            migrations={2: migration_fn},
            timestamp="20260101_142500",
        )

        mock_unlink.assert_called_once_with()
        migration_fn.assert_called_once()

    @patch("cmdbox.migrations.runner.backup_db")
    @patch("cmdbox.migrations.runner.set_user_version")
    @patch("cmdbox.migrations.runner.SqliteDatabase")
    @patch("cmdbox.migrations.runner.Path.rename")
    @patch("cmdbox.migrations.runner.Path.unlink")
    @patch("cmdbox.migrations.runner.Path.exists")
    def test_migrate_skips_unlink_when_no_stale_new_file(
        self,
        mock_exists,
        mock_unlink,
        mock_rename,
        mock_sqlite_db,
        mock_set_user_version,
        mock_backup_db,
    ):
        original_db = Path("C:\\tmp\\commands.sqlite")
        migration_fn = MagicMock()
        new_db = MagicMock()
        mock_sqlite_db.return_value = new_db
        mock_exists.return_value = False

        runner.migrate(
            1,
            original_db,
            migrations={2: migration_fn},
            timestamp="20260101_142500",
        )

        mock_unlink.assert_not_called()
