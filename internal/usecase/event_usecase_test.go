package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoadey/shiftmanager/internal/domain"
)

func newEventUC() (*EventUsecase, *fakeEventRepo, *fakeShiftRepo, *fakeRegistrationRepo, *fakeAuditRepo) {
	events := newFakeEventRepo()
	shifts := newFakeShiftRepo()
	regs := newFakeRegistrationRepo()
	audit := newFakeAuditRepo()
	return NewEventUsecase(events, shifts, regs, audit), events, shifts, regs, audit
}

func TestCreateEvent(t *testing.T) {
	uc, _, _, _, audit := newEventUC()
	start := time.Now().UTC()
	e, err := uc.CreateEvent(context.Background(), uuid.New(), CreateEventInput{
		Name: "Sommerfest", StartDate: start, EndDate: start.Add(8 * time.Hour),
	})
	require.NoError(t, err)
	assert.Equal(t, domain.EventStatusDraft, e.Status)
	assert.Equal(t, domain.EventVisibilityPublic, e.Visibility)
	assert.True(t, audit.has(domain.AuditActionCreate, domain.AuditEntityEvent))
}

func TestCreateEvent_Validation(t *testing.T) {
	uc, _, _, _, _ := newEventUC()
	_, err := uc.CreateEvent(context.Background(), uuid.New(), CreateEventInput{Name: ""})
	assert.Error(t, err)

	start := time.Now().UTC()
	_, err = uc.CreateEvent(context.Background(), uuid.New(), CreateEventInput{Name: "X", StartDate: start, EndDate: start.Add(-time.Hour)})
	assert.Error(t, err)
}

func TestPublishEvent(t *testing.T) {
	uc, events, _, _, _ := newEventUC()
	id := uuid.New()
	events.add(&domain.Event{ID: id, Status: domain.EventStatusDraft})
	e, err := uc.PublishEvent(context.Background(), uuid.New(), id)
	require.NoError(t, err)
	assert.Equal(t, domain.EventStatusPublished, e.Status)

	// Cannot publish again.
	_, err = uc.PublishEvent(context.Background(), uuid.New(), id)
	assert.ErrorIs(t, err, domain.ErrEventNotPublishable)
}

func TestDeleteEvent(t *testing.T) {
	uc, events, _, _, audit := newEventUC()
	id := uuid.New()
	events.add(&domain.Event{ID: id, Status: domain.EventStatusDraft})
	err := uc.DeleteEvent(context.Background(), uuid.New(), id)
	require.NoError(t, err)
	assert.True(t, audit.has(domain.AuditActionDelete, domain.AuditEntityEvent))

	// Published event cannot be deleted.
	pubID := uuid.New()
	events.add(&domain.Event{ID: pubID, Status: domain.EventStatusPublished})
	err = uc.DeleteEvent(context.Background(), uuid.New(), pubID)
	assert.Error(t, err)
}

func TestCreateShift(t *testing.T) {
	uc, events, _, _, audit := newEventUC()
	eventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft})
	start := time.Now().UTC()
	s, err := uc.CreateShift(context.Background(), uuid.New(), CreateShiftInput{
		EventID: eventID, Name: "Bar", StartAt: start, EndAt: start.Add(4 * time.Hour), MinHelpers: 2, MaxHelpers: 5,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, s.MinHelpers)
	assert.Equal(t, 5, s.MaxHelpers)
	assert.True(t, audit.has(domain.AuditActionCreate, domain.AuditEntityShift))
}

func TestCreateShift_BadTimes(t *testing.T) {
	uc, events, _, _, _ := newEventUC()
	eventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft})
	start := time.Now().UTC()
	_, err := uc.CreateShift(context.Background(), uuid.New(), CreateShiftInput{EventID: eventID, StartAt: start, EndAt: start})
	assert.Error(t, err)
}

func TestGetEventTimeline_Occupancy(t *testing.T) {
	uc, events, shifts, regs, _ := newEventUC()
	eventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusPublished})

	day := time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC)
	shiftID := uuid.New()
	shifts.add(&domain.Shift{ID: shiftID, EventID: eventID, StartAt: day, EndAt: day.Add(4 * time.Hour), Date: day, MinHelpers: 2, MaxHelpers: 3})

	regs.add(&domain.Registration{ID: uuid.New(), ShiftID: shiftID, State: domain.RegistrationStateRegistered})
	regs.add(&domain.Registration{ID: uuid.New(), ShiftID: shiftID, State: domain.RegistrationStateConfirmed})

	tl, err := uc.GetEventTimeline(context.Background(), eventID)
	require.NoError(t, err)
	require.Len(t, tl.Days, 1)
	require.Len(t, tl.Days[0].Shifts, 1)
	occ := tl.Days[0].Shifts[0].Occupancy
	assert.Equal(t, 2, occ.Registered)
	assert.Equal(t, 1, occ.Confirmed)
	assert.False(t, occ.IsUnder)
	assert.False(t, occ.IsFull)
	assert.Equal(t, domain.OccupancyBesetzt, occ.Status())
}
