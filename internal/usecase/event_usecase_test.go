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

func newEventUC() (*EventUsecase, *fakeEventRepo, *fakeShiftRepo, *fakeRegistrationRepo, *fakeAuditRepo, *fakeMemberRepo) {
	events := newFakeEventRepo()
	shifts := newFakeShiftRepo()
	regs := newFakeRegistrationRepo()
	audit := newFakeAuditRepo()
	members := newFakeMemberRepo()
	email := &fakeEmailService{}
	return NewEventUsecase(events, shifts, regs, audit, email, members), events, shifts, regs, audit, members
}

func TestCreateEvent(t *testing.T) {
	uc, _, _, _, audit, _ := newEventUC()
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
	uc, _, _, _, _, _ := newEventUC()
	_, err := uc.CreateEvent(context.Background(), uuid.New(), CreateEventInput{Name: ""})
	assert.Error(t, err)

	start := time.Now().UTC()
	_, err = uc.CreateEvent(context.Background(), uuid.New(), CreateEventInput{Name: "X", StartDate: start, EndDate: start.Add(-time.Hour)})
	assert.Error(t, err)
}

func TestPublishEvent(t *testing.T) {
	uc, events, _, _, _, _ := newEventUC()
	id := uuid.New()
	events.add(&domain.Event{ID: id, Status: domain.EventStatusDraft})
	e, err := uc.PublishEvent(context.Background(), uuid.New(), id)
	require.NoError(t, err)
	assert.Equal(t, domain.EventStatusPublished, e.Status)

	// Cannot publish again.
	_, err = uc.PublishEvent(context.Background(), uuid.New(), id)
	assert.ErrorIs(t, err, domain.ErrEventNotPublishable)
}

// PublishEvent notifies opted-in active members ("Neue Veranstaltung", KANN).
func TestPublishEvent_NotifiesOptedInMembers(t *testing.T) {
	uc, events, _, _, _, members := newEventUC()
	members.add(&domain.Member{ID: uuid.New(), Email: "opted-in@club.de", IsActive: true, NotifyNewEvents: true})
	members.add(&domain.Member{ID: uuid.New(), Email: "opted-out@club.de", IsActive: true, NotifyNewEvents: false})
	members.add(&domain.Member{ID: uuid.New(), Email: "inactive@club.de", IsActive: false, NotifyNewEvents: true})

	id := uuid.New()
	events.add(&domain.Event{ID: id, Name: "Sommerfest", Status: domain.EventStatusDraft})
	_, err := uc.PublishEvent(context.Background(), uuid.New(), id)
	require.NoError(t, err)

	email := uc.email.(*fakeEmailService)
	require.Len(t, email.sent, 1)
	assert.Equal(t, "opted-in@club.de", email.sent[0].to)
	assert.Equal(t, "template:"+domain.EmailTemplateNewEvent, email.sent[0].kind)
}

func TestDeleteEvent(t *testing.T) {
	uc, events, _, _, audit, _ := newEventUC()
	id := uuid.New()
	events.add(&domain.Event{ID: id, Status: domain.EventStatusDraft})
	err := uc.DeleteEvent(context.Background(), uuid.New(), id)
	require.NoError(t, err)
	assert.True(t, audit.has(domain.AuditActionDelete, domain.AuditEntityEvent))

	// Published events can be deleted (role enforcement happens at the HTTP layer).
	pubID := uuid.New()
	events.add(&domain.Event{ID: pubID, Status: domain.EventStatusPublished})
	err = uc.DeleteEvent(context.Background(), uuid.New(), pubID)
	assert.NoError(t, err)
}

func TestCreateShift(t *testing.T) {
	uc, events, _, _, audit, _ := newEventUC()
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
	uc, events, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft})
	start := time.Now().UTC()
	_, err := uc.CreateShift(context.Background(), uuid.New(), CreateShiftInput{EventID: eventID, StartAt: start, EndAt: start})
	assert.Error(t, err)
}

func TestGetEventTimeline_Occupancy(t *testing.T) {
	uc, events, shifts, regs, _, _ := newEventUC()
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
