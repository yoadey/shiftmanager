package db

import (
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DataMigrationModel records one-off data migrations that Migrate has
// applied. AutoMigrate only evolves the schema; data fix-ups that must run
// exactly once (re-running them later could change rows written since) are
// tracked here.
type DataMigrationModel struct {
	Name      string    `gorm:"column:name;type:text;primaryKey"`
	AppliedAt time.Time `gorm:"column:applied_at"`
}

func (DataMigrationModel) TableName() string { return "data_migrations" }

// dataMigration is a one-off SQL statement, run once per database in order.
type dataMigration struct {
	name string
	sql  string
}

// backfillMemberLeftAtSQL mirrors the UPDATE in
// migrations/010_backfill_member_left_at.up.sql (golang-migrate path); see
// there for the rationale. That file records the run in data_migrations under
// the same name, so whichever path runs first, the other skips it. Kept
// portable so it also runs on SQLite.
const backfillMemberLeftAtSQL = `UPDATE members
SET left_at = CURRENT_TIMESTAMP
WHERE NOT EXISTS (SELECT 1 FROM data_migrations WHERE name = '010_backfill_member_left_at')
  AND is_active = false
  AND left_at IS NULL
  AND NOT (
    EXISTS (
      SELECT 1 FROM audit_log a
      WHERE a.entity = 'member'
        AND a.entity_id = CAST(members.id AS TEXT)
        AND a.action = 'member.autoregister'
    )
    AND NOT EXISTS (
      SELECT 1 FROM audit_log a
      WHERE a.entity = 'member'
        AND a.entity_id = CAST(members.id AS TEXT)
        AND a.action <> 'member.autoregister'
    )
  );`

var dataMigrations = []dataMigration{
	// ML-011: deactivations via PUT used to leave left_at empty; such members
	// must not be mistaken for pending self-registrations.
	{name: "010_backfill_member_left_at", sql: backfillMemberLeftAtSQL},
}

// runDataMigrations applies every data migration not yet recorded. Each runs
// in its own transaction together with its record; concurrent replicas may
// both apply one (the statements are safe to repeat at the same moment), and
// the record insert ignores the duplicate.
func runDataMigrations(gdb *gorm.DB) error {
	for _, m := range dataMigrations {
		var n int64
		if err := gdb.Model(&DataMigrationModel{}).Where("name = ?", m.name).Count(&n).Error; err != nil {
			return fmt.Errorf("data migration %s: %w", m.name, err)
		}
		if n > 0 {
			continue
		}
		err := gdb.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(m.sql).Error; err != nil {
				return err
			}
			return tx.Clauses(clause.OnConflict{DoNothing: true}).
				Create(&DataMigrationModel{Name: m.name, AppliedAt: time.Now().UTC()}).Error
		})
		if err != nil {
			return fmt.Errorf("data migration %s: %w", m.name, err)
		}
	}
	return nil
}
