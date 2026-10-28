package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
	"gorm.io/gorm"
)

func newTestEventDB(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := Open("sqlite::memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(gdb); err != nil {
		t.Fatal(err)
	}
	return gdb
}

func createTestEvent(t *testing.T, r *EventRepo, name, category string, start time.Time) *domain.Event {
	t.Helper()
	e := &domain.Event{
		ID: uuid.New(), Name: name, Category: category,
		StartDate: start, EndDate: start.Add(3 * time.Hour),
		Status: domain.EventStatusPublished, Visibility: domain.EventVisibilityPublic,
	}
	if err := r.Create(context.Background(), e); err != nil {
		t.Fatal(err)
	}
	return e
}

func eventNames(events []*domain.Event) []string {
	out := make([]string, len(events))
	for i, e := range events {
		out[i] = e.Name
	}
	return out
}

func TestEventRepoList_FiltersAndOrder(t *testing.T) {
	gdb := newTestEventDB(t)
	r := NewEventRepo(gdb)
	ctx := context.Background()
	day := func(d int) time.Time { return time.Date(2026, 11, d, 10, 0, 0, 0, time.UTC) }

	createTestEvent(t, r, "past", "Turnier", day(1))
	createTestEvent(t, r, "a", "Turnier", day(5))
	createTestEvent(t, r, "b", "Vereinsleben", day(6))
	createTestEvent(t, r, "c", "Turnier", day(7))

	from := time.Date(2026, 11, 5, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC) // exclusive
	got, err := r.List(ctx, port.EventFilter{FromDate: &from, ToDate: &to, Ascending: true})
	if err != nil {
		t.Fatal(err)
	}
	if names := eventNames(got); len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Errorf("from/to ascending = %v, want [a b]", names)
	}

	turnier, err := r.List(ctx, port.EventFilter{Category: "Turnier"})
	if err != nil {
		t.Fatal(err)
	}
	if names := eventNames(turnier); len(names) != 3 || names[0] != "c" || names[2] != "past" {
		t.Errorf("category descending = %v, want [c a past]", names)
	}

	page, err := r.List(ctx, port.EventFilter{FromDate: &from, Ascending: true, Limit: 2, Offset: 2})
	if err != nil {
		t.Fatal(err)
	}
	if names := eventNames(page); len(names) != 1 || names[0] != "c" {
		t.Errorf("second page = %v, want [c]", names)
	}
	n, err := r.CountFiltered(ctx, port.EventFilter{FromDate: &from, Limit: 2, Offset: 2})
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("count = %d, want 3", n)
	}
}

func TestEventRepoList_FreeSlots(t *testing.T) {
	gdb := newTestEventDB(t)
	r := NewEventRepo(gdb)
	shifts := NewShiftRepo(gdb)
	regs := NewRegistrationRepo(gdb)
	ctx := context.Background()
	start := time.Date(2026, 11, 5, 10, 0, 0, 0, time.UTC)

	addShift := func(e *domain.Event, max int, states ...domain.RegistrationState) {
		t.Helper()
		s := &domain.Shift{ID: uuid.New(), EventID: e.ID, Name: "s", StartAt: start, EndAt: start.Add(2 * time.Hour), MinHelpers: 1, MaxHelpers: max, Date: start}
		if err := shifts.Create(ctx, s); err != nil {
			t.Fatal(err)
		}
		for _, st := range states {
			mid := uuid.New()
			if err := regs.Create(ctx, &domain.Registration{ID: uuid.New(), ShiftID: s.ID, MemberID: &mid, State: st}); err != nil {
				t.Fatal(err)
			}
		}
	}

	full := createTestEvent(t, r, "full", "Turnier", start)
	addShift(full, 2, domain.RegistrationStateRegistered, domain.RegistrationStateConfirmed)
	open := createTestEvent(t, r, "open", "Turnier", start)
	addShift(open, 2, domain.RegistrationStateConfirmed)
	addShift(open, 1, domain.RegistrationStateRegistered)
	// A waiting-list reservation does not take one of the places.
	waitlist := createTestEvent(t, r, "waitlist", "Turnier", start)
	addShift(waitlist, 1, domain.RegistrationStateReserved)
	createTestEvent(t, r, "noshifts", "Turnier", start)

	got, err := r.List(ctx, port.EventFilter{FreeSlots: true, Ascending: true})
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, n := range eventNames(got) {
		names[n] = true
	}
	if len(names) != 2 || !names["open"] || !names["waitlist"] {
		t.Errorf("free slots = %v, want open and waitlist", eventNames(got))
	}
	n, err := r.CountFiltered(ctx, port.EventFilter{FreeSlots: true})
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Errorf("count = %d, want 2", n)
	}
}
