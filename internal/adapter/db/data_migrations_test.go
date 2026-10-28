package db

import (
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"gorm.io/gorm"
)

func TestBackfillMemberLeftAt(t *testing.T) {
	gdb, err := Open("sqlite::memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := Migrate(gdb); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// Simulate a database that predates the data migration.
	if err := gdb.Where("name = ?", "010_backfill_member_left_at").Delete(&DataMigrationModel{}).Error; err != nil {
		t.Fatal(err)
	}

	left := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	legacy := addMember(t, gdb, false, nil)        // deactivated via PUT before ML-011
	pending := addMember(t, gdb, false, nil)       // self-registered, untouched since
	editedPending := addMember(t, gdb, false, nil) // self-registered, later edited/deactivated
	active := addMember(t, gdb, true, nil)         // active members are never touched
	alreadyLeft := addMember(t, gdb, false, &left) // keeps its original leave date
	addAudit(t, gdb, pending, "member.autoregister")
	addAudit(t, gdb, editedPending, "member.autoregister")
	addAudit(t, gdb, editedPending, domain.AuditActionUpdate)

	if err := Migrate(gdb); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	assertLeftAt(t, gdb, legacy, true, "legacy deactivation gets a leave date")
	assertLeftAt(t, gdb, pending, false, "untouched self-registration stays pending")
	assertLeftAt(t, gdb, editedPending, true, "self-registration with later changes counts as left")
	assertLeftAt(t, gdb, active, false, "active member is untouched")
	var m MemberModel
	if err := gdb.First(&m, "id = ?", alreadyLeft).Error; err != nil {
		t.Fatal(err)
	}
	if m.LeftAt == nil || !m.LeftAt.Equal(left) {
		t.Errorf("existing leave date changed: %v", m.LeftAt)
	}

	// Runs once: a registration that later gets audit entries other than
	// autoregister (e.g. an admin edits the pending account) must not be
	// flipped to "left" on the next startup.
	addAudit(t, gdb, pending, domain.AuditActionUpdate)
	if err := Migrate(gdb); err != nil {
		t.Fatalf("migrate again: %v", err)
	}
	assertLeftAt(t, gdb, pending, false, "data migration must not re-run")
}

// The golang-migrate SQL file and the statement Migrate runs must not drift.
func TestBackfillMemberLeftAt_MatchesSQLMigration(t *testing.T) {
	raw, err := os.ReadFile("../../../migrations/010_backfill_member_left_at.up.sql")
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, l := range strings.Split(string(raw), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(l), "--") {
			lines = append(lines, l)
		}
	}
	ws := regexp.MustCompile(`\s+`)
	norm := func(s string) string { return strings.TrimSpace(ws.ReplaceAllString(s, " ")) }
	if norm(strings.Join(lines, "\n")) != norm(backfillMemberLeftAtSQL) {
		t.Error("migrations/010_backfill_member_left_at.up.sql differs from backfillMemberLeftAtSQL")
	}
}

func addMember(t *testing.T, gdb *gorm.DB, active bool, leftAt *time.Time) string {
	t.Helper()
	id := uuid.New()
	m := toMemberModel(&domain.Member{
		ID: id, FirstName: "F", LastName: id.String()[:8], Email: id.String() + "@x.de",
		JoinedAt: time.Now().UTC(), IsActive: active, LeftAt: leftAt, Role: domain.RoleMitglied,
	})
	if err := gdb.Create(&m).Error; err != nil {
		t.Fatal(err)
	}
	// GORM skips zero-value bools on Create when the column has a default.
	if err := gdb.Model(&MemberModel{}).Where("id = ?", id.String()).Update("is_active", active).Error; err != nil {
		t.Fatal(err)
	}
	return id.String()
}

func addAudit(t *testing.T, gdb *gorm.DB, memberID, action string) {
	t.Helper()
	e := AuditEntryModel{ID: uuid.NewString(), Action: action, Entity: "member", EntityID: memberID, ChangedAt: time.Now().UTC()}
	if err := gdb.Create(&e).Error; err != nil {
		t.Fatal(err)
	}
}

func assertLeftAt(t *testing.T, gdb *gorm.DB, id string, want bool, msg string) {
	t.Helper()
	var m MemberModel
	if err := gdb.First(&m, "id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
	if (m.LeftAt != nil) != want {
		t.Errorf("%s: left_at = %v", msg, m.LeftAt)
	}
}
