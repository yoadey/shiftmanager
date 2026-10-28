package db

import (
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Open opens a GORM database. When dsn starts with "sqlite:" it uses the SQLite
// driver (e.g. "sqlite::memory:"); otherwise a PostgreSQL connection is opened.
func Open(dsn string) (*gorm.DB, error) {
	if !strings.HasPrefix(dsn, "sqlite:") {
		return gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	// Strip the "sqlite:" prefix so the remainder is the SQLite DSN.
	sqliteDSN := strings.TrimPrefix(dsn, "sqlite:")
	gdb, err := gorm.Open(sqlite.Open(sqliteDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if isInMemorySQLite(sqliteDSN) {
		// An in-memory database exists only inside the connection that
		// created it: every further pool connection would open a new, empty
		// database ("no such table" under concurrent requests). Pin the pool
		// to one connection that is never closed; concurrent queries queue
		// for it instead. Transactions must therefore only use their tx
		// handle, never the outer *gorm.DB, or they would deadlock.
		sqlDB, err := gdb.DB()
		if err != nil {
			return nil, err
		}
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
		sqlDB.SetConnMaxLifetime(0)
		sqlDB.SetConnMaxIdleTime(0)
	}
	return gdb, nil
}

// isInMemorySQLite reports whether a SQLite DSN names an in-memory database
// (":memory:", "file::memory:…" or a URI with mode=memory).
func isInMemorySQLite(dsn string) bool {
	return strings.Contains(dsn, ":memory:") || strings.Contains(dsn, "mode=memory")
}

// Migrate runs AutoMigrate for all registered models, then any one-off data
// migrations not yet applied. Safe to call on every startup; GORM will only
// add missing columns and tables.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&MemberModel{},
		&OIDCLinkModel{},
		&EventModel{},
		&EventAttachmentModel{},
		&ShiftModel{},
		&RegistrationModel{},
		&HourEntryModel{},
		&ClubYearModel{},
		&HourTargetModel{},
		&AuditEntryModel{},
		&SettingModel{},
		&BrandingModel{},
		&FeeTierModel{},
		&MemberFeeTierModel{},
		&EmailTemplateModel{},
		&EmailLogModel{},
		&BrandingHistoryModel{},
		&DataMigrationModel{},
	); err != nil {
		return err
	}
	return runDataMigrations(db)
}
