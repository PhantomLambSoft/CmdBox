package database

import (
	"database/sql"
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/PhantomLambSoft/CmdBox/internal/database/migrations"
)

// Connect initializes and returns a gorm.DB instance connected to the specified SQLite database path.
// It ensures the database file exists and foreign key support is enabled.
// Returns an error if the connection or configuration fails.
func Connect(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("opening database at %s: %w", dbPath, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting underlying sql.DB: %w", err)
	}

	if _, err := sqlDB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	return db, nil
}

// UnderlyingSQLDB retrieves the underlying *sql.DB object from a given *gorm.DB instance and returns it.
func UnderlyingSQLDB(db *gorm.DB) (*sql.DB, error) {
	return db.DB()
}

// Bootstrap initializes the database connection, runs schema migrations, and returns a gorm.DB instance or an error.
func Bootstrap(dbPath string) (*gorm.DB, error) {
	db, err := Connect(dbPath)
	if err != nil {
		return nil, err
	}

	sqlDB, err := UnderlyingSQLDB(db)
	if err != nil {
		return nil, err
	}

	if err := migrations.Run(sqlDB); err != nil {
		return nil, fmt.Errorf("bootstrapping schema for %s: %w", dbPath, err)
	}

	return db, nil
}
