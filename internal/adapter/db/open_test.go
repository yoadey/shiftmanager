package db

import (
	"testing"
	"time"

	"gorm.io/gorm"
)

// An in-memory SQLite database lives inside a single connection; a second
// pool connection would open a fresh, empty database ("no such table").
// Concurrent requests (TEST_MODE, integration tests) must therefore all see
// the migrated schema.
func TestOpen_InMemorySQLiteSharedAcrossConcurrentUse(t *testing.T) {
	gdb, err := Open("sqlite::memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := Migrate(gdb); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	inTx := make(chan struct{})
	release := make(chan struct{})
	txDone := make(chan error, 1)
	go func() {
		// Hold the pool's connection busy, forcing any concurrent query to
		// either wait for it or open another connection.
		txDone <- gdb.Transaction(func(tx *gorm.DB) error {
			close(inTx)
			<-release
			return tx.Model(&MemberModel{}).Count(new(int64)).Error
		})
	}()
	<-inTx

	queryDone := make(chan error, 1)
	go func() {
		var n int64
		queryDone <- gdb.Model(&MemberModel{}).Count(&n).Error
	}()

	time.Sleep(50 * time.Millisecond)
	close(release)

	for name, ch := range map[string]chan error{"transaction": txDone, "concurrent query": queryDone} {
		select {
		case err := <-ch:
			if err != nil {
				t.Errorf("%s: %v", name, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("%s did not finish", name)
		}
	}
}
