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
	var dialector gorm.Dialector
	if strings.HasPrefix(dsn, "sqlite:") {
		// Strip the "sqlite:" prefix so the remainder is the SQLite DSN.
		sqliteDSN := strings.TrimPrefix(dsn, "sqlite:")
		dialector = sqlite.Open(sqliteDSN)
	} else {
		dialector = postgres.Open(dsn)
	}
	return gorm.Open(dialector, &gorm.Config{})
}

// Migrate runs AutoMigrate for all registered models. Safe to call on every
// startup; GORM will only add missing columns and tables.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&MemberModel{},
		&OIDCLinkModel{},
		&EventModel{},
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
	)
}
