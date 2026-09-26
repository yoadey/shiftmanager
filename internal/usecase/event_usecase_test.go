package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

func newEventUC() (*EventUsecase, *fakeEventRepo, *fakeShiftRepo, *fakeRegistrationRepo, *fakeAuditRepo, *fakeMemberRepo, *fakeEventAttachmentRepo) {
	events := newFakeEventRepo()
	shifts := newFakeShiftRepo()
	regs := newFakeRegistrationRepo()
	audit := newFakeAuditRepo()
	members := newFakeMemberRepo()
	attachments := newFakeEventAttachmentRepo()
	email := &fakeEmailService{}
	return NewEventUsecase(events, shifts, regs, audit, email, members, attachments), events, shifts, regs, audit, members, attachments
}

func TestCreateEvent(t *testing.T) {
	uc, _, _, _, audit, _, _ := newEventUC()
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
	uc, _, _, _, _, _, _ := newEventUC()
	_, err := uc.CreateEvent(context.Background(), uuid.New(), CreateEventInput{Name: ""})
	assert.Error(t, err)

	start := time.Now().UTC()
	_, err = uc.CreateEvent(context.Background(), uuid.New(), CreateEventInput{Name: "X", StartDate: start, EndDate: start.Add(-time.Hour)})
	assert.Error(t, err)
}

func TestPublishEvent(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
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
	uc, events, _, _, _, members, _ := newEventUC()
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
	uc, events, _, _, audit, _, _ := newEventUC()
	id := uuid.New()
	events.add(&domain.Event{ID: id, Status: domain.EventStatusDraft})
	_, err := uc.DeleteEvent(context.Background(), uuid.New(), id)
	require.NoError(t, err)
	assert.True(t, audit.has(domain.AuditActionDelete, domain.AuditEntityEvent))

	// Published events can be deleted (role enforcement happens at the HTTP layer).
	pubID := uuid.New()
	events.add(&domain.Event{ID: pubID, Status: domain.EventStatusPublished})
	_, err = uc.DeleteEvent(context.Background(), uuid.New(), pubID)
	assert.NoError(t, err)
}

func TestDeleteEvent_RemovesAttachments(t *testing.T) {
	uc, events, _, _, _, _, attachments := newEventUC()
	id := uuid.New()
	events.add(&domain.Event{ID: id, Status: domain.EventStatusDraft})
	a, err := uc.AddAttachment(context.Background(), uuid.New(), id, "flyer.png", "/uploads/flyer-abc.png", "image/png", 1024)
	require.NoError(t, err)

	deleted, err := uc.DeleteEvent(context.Background(), uuid.New(), id)
	require.NoError(t, err)
	require.Len(t, deleted, 1)
	assert.Equal(t, a.ID, deleted[0].ID)

	_, err = attachments.GetByID(context.Background(), a.ID)
	assert.ErrorIs(t, err, domain.ErrEventAttachmentNotFound)
}

func TestCreateShift(t *testing.T) {
	uc, events, _, _, audit, _, _ := newEventUC()
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
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft})
	start := time.Now().UTC()
	_, err := uc.CreateShift(context.Background(), uuid.New(), CreateShiftInput{EventID: eventID, StartAt: start, EndAt: start})
	assert.Error(t, err)
}

func TestAddAttachment(t *testing.T) {
	uc, events, _, _, audit, _, attachments := newEventUC()
	eventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft})

	a, err := uc.AddAttachment(context.Background(), uuid.New(), eventID, "flyer.png", "/uploads/flyer-abc.png", "image/png", 1024)
	require.NoError(t, err)
	assert.Equal(t, eventID, a.EventID)
	assert.Equal(t, "image/png", a.ContentType)
	assert.True(t, audit.has(domain.AuditActionCreate, domain.AuditEntityEvent))

	list, err := uc.ListAttachments(context.Background(), eventID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "flyer.png", list[0].FileName)

	stored, err := attachments.GetByID(context.Background(), a.ID)
	require.NoError(t, err)
	assert.Equal(t, a.URL, stored.URL)
}

func TestAddAttachment_UnknownEvent(t *testing.T) {
	uc, _, _, _, _, _, _ := newEventUC()
	_, err := uc.AddAttachment(context.Background(), uuid.New(), uuid.New(), "flyer.png", "/uploads/x.png", "image/png", 10)
	assert.ErrorIs(t, err, domain.ErrEventNotFound)
}

func TestListAttachments_UnknownEvent(t *testing.T) {
	uc, _, _, _, _, _, _ := newEventUC()
	_, err := uc.ListAttachments(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrEventNotFound)
}

func TestDeleteAttachment(t *testing.T) {
	uc, events, _, _, audit, _, _ := newEventUC()
	eventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft})
	a, err := uc.AddAttachment(context.Background(), uuid.New(), eventID, "flyer.png", "/uploads/flyer-abc.png", "image/png", 1024)
	require.NoError(t, err)

	deleted, err := uc.DeleteAttachment(context.Background(), uuid.New(), eventID, a.ID)
	require.NoError(t, err)
	assert.Equal(t, a.URL, deleted.URL)
	assert.True(t, audit.has(domain.AuditActionDelete, domain.AuditEntityEvent))

	list, err := uc.ListAttachments(context.Background(), eventID)
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestDeleteAttachment_WrongEvent(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	otherEventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft})
	events.add(&domain.Event{ID: otherEventID, Status: domain.EventStatusDraft})
	a, err := uc.AddAttachment(context.Background(), uuid.New(), eventID, "flyer.png", "/uploads/flyer-abc.png", "image/png", 1024)
	require.NoError(t, err)

	_, err = uc.DeleteAttachment(context.Background(), uuid.New(), otherEventID, a.ID)
	assert.ErrorIs(t, err, domain.ErrEventAttachmentNotFound)
}

func TestSetHeaderImage(t *testing.T) {
	uc, events, _, _, audit, _, _ := newEventUC()
	eventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft})

	e, prevURL, err := uc.SetHeaderImage(context.Background(), uuid.New(), eventID, "/uploads/event-header-abc.png")
	require.NoError(t, err)
	assert.Equal(t, "/uploads/event-header-abc.png", e.HeaderImageURL)
	assert.Empty(t, prevURL)
	assert.True(t, audit.has(domain.AuditActionUpdate, domain.AuditEntityEvent))

	stored, err := events.GetByID(context.Background(), eventID)
	require.NoError(t, err)
	assert.Equal(t, "/uploads/event-header-abc.png", stored.HeaderImageURL)
}

func TestSetHeaderImage_ReplacesExisting(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft, HeaderImageURL: "/uploads/event-header-old.png"})

	e, prevURL, err := uc.SetHeaderImage(context.Background(), uuid.New(), eventID, "/uploads/event-header-new.png")
	require.NoError(t, err)
	assert.Equal(t, "/uploads/event-header-new.png", e.HeaderImageURL)
	// The caller (handler) needs the old URL back to delete that now-orphaned
	// file from storage — the same contract ClearHeaderImage already has.
	assert.Equal(t, "/uploads/event-header-old.png", prevURL)
}

func TestSetHeaderImage_UnknownEvent(t *testing.T) {
	uc, _, _, _, _, _, _ := newEventUC()
	_, _, err := uc.SetHeaderImage(context.Background(), uuid.New(), uuid.New(), "/uploads/x.png")
	assert.ErrorIs(t, err, domain.ErrEventNotFound)
}

func TestClearHeaderImage(t *testing.T) {
	uc, events, _, _, audit, _, _ := newEventUC()
	eventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft, HeaderImageURL: "/uploads/event-header-abc.png"})

	prevURL, err := uc.ClearHeaderImage(context.Background(), uuid.New(), eventID)
	require.NoError(t, err)
	assert.Equal(t, "/uploads/event-header-abc.png", prevURL)
	assert.True(t, audit.has(domain.AuditActionUpdate, domain.AuditEntityEvent))

	stored, err := events.GetByID(context.Background(), eventID)
	require.NoError(t, err)
	assert.Empty(t, stored.HeaderImageURL)
}

func TestClearHeaderImage_NoneSet(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft})

	prevURL, err := uc.ClearHeaderImage(context.Background(), uuid.New(), eventID)
	require.NoError(t, err)
	assert.Empty(t, prevURL)
}

func TestClearHeaderImage_UnknownEvent(t *testing.T) {
	uc, _, _, _, _, _, _ := newEventUC()
	_, err := uc.ClearHeaderImage(context.Background(), uuid.New(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrEventNotFound)
}

func TestGetEventTimeline_Occupancy(t *testing.T) {
	uc, events, shifts, regs, _, _, _ := newEventUC()
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

func TestAddMonthsClamped(t *testing.T) {
	cases := []struct {
		name string
		in   time.Time
		n    int
		want time.Time
	}{
		{"end of Jan + 1 clamps to Feb 28 (non-leap)", time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC), 1, time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)},
		{"end of Jan + 1 clamps to Feb 29 (leap year)", time.Date(2028, 1, 31, 10, 0, 0, 0, time.UTC), 1, time.Date(2028, 2, 29, 10, 0, 0, 0, time.UTC)},
		{"end of Jan + 2 lands on Mar 31 exactly", time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC), 2, time.Date(2026, 3, 31, 10, 0, 0, 0, time.UTC)},
		{"mid-month day is never clamped", time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC), 1, time.Date(2026, 2, 15, 10, 0, 0, 0, time.UTC)},
		{"December + 1 rolls over the year", time.Date(2026, 12, 31, 10, 0, 0, 0, time.UTC), 1, time.Date(2027, 1, 31, 10, 0, 0, 0, time.UTC)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := addMonthsClamped(c.in, c.n)
			assert.True(t, got.Equal(c.want), "got %v, want %v", got, c.want)
		})
	}
}

func TestGenerateRecurrence_Weekly(t *testing.T) {
	uc, events, shifts, _, audit, _, _ := newEventUC()
	eventID := uuid.New()
	start := time.Date(2026, 1, 5, 18, 0, 0, 0, time.UTC) // a Monday
	events.add(&domain.Event{ID: eventID, Name: "Training", Status: domain.EventStatusDraft, StartDate: start, EndDate: start.Add(2 * time.Hour)})
	shifts.add(&domain.Shift{ID: uuid.New(), EventID: eventID, Name: "Aufsicht", StartAt: start, EndAt: start.Add(2 * time.Hour), MinHelpers: 1, MaxHelpers: 2})

	until := start.AddDate(0, 0, 21) // 3 weeks later
	result, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyWeekly, until)
	require.NoError(t, err)
	require.Len(t, result, 4) // source + 3 weekly occurrences

	src := result[0]
	assert.Equal(t, eventID, src.ID)
	assert.Equal(t, domain.RecurrenceFrequencyWeekly, src.RecurrenceFrequency)
	require.NotNil(t, src.RecurrenceGroupID)
	require.NotNil(t, src.RecurrenceUntil)
	assert.True(t, src.RecurrenceUntil.Equal(until))

	for i, occ := range result[1:] {
		assert.NotEqual(t, eventID, occ.ID)
		assert.Equal(t, "Training", occ.Name)
		assert.Equal(t, domain.EventStatusDraft, occ.Status)
		assert.Equal(t, *src.RecurrenceGroupID, *occ.RecurrenceGroupID)
		wantStart := start.AddDate(0, 0, 7*(i+1))
		assert.True(t, occ.StartDate.Equal(wantStart), "occurrence %d start", i)
		assert.True(t, occ.EndDate.Equal(wantStart.Add(2*time.Hour)))

		occShifts, err := shifts.FindByEventID(context.Background(), occ.ID)
		require.NoError(t, err)
		require.Len(t, occShifts, 1)
		assert.Equal(t, "Aufsicht", occShifts[0].Name)
		assert.True(t, occShifts[0].StartAt.Equal(wantStart))
	}

	assert.True(t, audit.has(domain.AuditActionCreate, domain.AuditEntityEvent))
}

func TestGenerateRecurrence_Monthly(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	start := time.Date(2026, 1, 31, 10, 0, 0, 0, time.UTC)
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft, StartDate: start, EndDate: start.Add(time.Hour)})

	until := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	result, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyMonthly, until)
	require.NoError(t, err)
	// Jan 31 has no equivalent in February, so that occurrence clamps to
	// Feb 28 (2026 is not a leap year) rather than overflowing into March, as
	// plain time.Time.AddDate(0, n, 0) would (Jan 31 + 1 month -> Mar 3).
	// March has a 31st, so the following occurrence lands back on it exactly
	// -- each occurrence is computed from the original start date, not
	// cumulatively from the previous (possibly clamped) one, so the Feb
	// clamp doesn't drag March off course.
	wantStarts := []time.Time{
		time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 31, 10, 0, 0, 0, time.UTC),
	}
	require.Len(t, result, 1+len(wantStarts))
	for i, want := range wantStarts {
		assert.True(t, result[i+1].StartDate.Equal(want), "occurrence %d: got %v, want %v", i, result[i+1].StartDate, want)
	}
}

func TestGenerateRecurrence_InvalidFrequency(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	start := time.Now().UTC()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft, StartDate: start, EndDate: start.Add(time.Hour)})

	_, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequency("daily"), start.AddDate(0, 0, 7))
	assert.ErrorIs(t, err, domain.ErrInvalidRecurrence)
}

func TestGenerateRecurrence_UntilBeforeStart(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	start := time.Now().UTC()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft, StartDate: start, EndDate: start.Add(time.Hour)})

	_, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyWeekly, start.Add(-time.Hour))
	assert.ErrorIs(t, err, domain.ErrInvalidRecurrence)
}

func TestGenerateRecurrence_AlreadyRecurring(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	start := time.Now().UTC()
	groupID := uuid.New()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft, StartDate: start, EndDate: start.Add(time.Hour), RecurrenceGroupID: &groupID})

	_, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyWeekly, start.AddDate(0, 0, 14))
	assert.ErrorIs(t, err, domain.ErrAlreadyRecurring)
}

func TestGenerateRecurrence_UnknownEvent(t *testing.T) {
	uc, _, _, _, _, _, _ := newEventUC()
	_, err := uc.GenerateRecurrence(context.Background(), uuid.New(), uuid.New(), domain.RecurrenceFrequencyWeekly, time.Now().UTC().AddDate(0, 0, 7))
	assert.ErrorIs(t, err, domain.ErrEventNotFound)
}

// A failure partway through generating occurrences must not leave the source
// event permanently stamped as recurring: that would make every retry fail
// with ErrAlreadyRecurring even though the series was never actually built.
func TestGenerateRecurrence_PartialFailureAllowsRetry(t *testing.T) {
	uc, events, shifts, _, audit, _, _ := newEventUC()
	eventID := uuid.New()
	start := time.Date(2026, 1, 5, 18, 0, 0, 0, time.UTC)
	events.add(&domain.Event{ID: eventID, Name: "Training", Status: domain.EventStatusDraft, StartDate: start, EndDate: start.Add(time.Hour)})
	shifts.add(&domain.Shift{ID: uuid.New(), EventID: eventID, Name: "Aufsicht", StartAt: start, EndAt: start.Add(time.Hour)})

	until := start.AddDate(0, 0, 21) // 3 weekly occurrences
	shifts.failCreateAt = 2          // fail the 2nd occurrence's shift copy
	_, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyWeekly, until)
	require.Error(t, err)

	stored, err := events.GetByID(context.Background(), eventID)
	require.NoError(t, err)
	assert.Nil(t, stored.RecurrenceGroupID, "source event must not be stamped as recurring after a partial failure")

	// The one occurrence event created before the failure must have been
	// rolled back too, not left as an orphan with no source event tracking it.
	all, err := events.List(context.Background(), port.EventFilter{})
	require.NoError(t, err)
	assert.Len(t, all, 1, "no orphaned occurrence events after a partial failure")

	// The audit log is append-only (no compensating delete entries), so a
	// rolled-back occurrence must never have gotten a "create" audit entry
	// in the first place.
	assert.Equal(t, 0, audit.countActions(string(domain.AuditActionCreate)), "no create-audit entry for a rolled-back occurrence")

	// The retry must not be rejected by the "already recurring" guard.
	shifts.failCreateAt = 0
	result, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyWeekly, until)
	require.NoError(t, err)
	assert.Len(t, result, 4)
}

// If the final step (stamping the source event as recurring) fails after all
// occurrences were already created, those occurrences must be rolled back
// too — otherwise they're permanently orphaned (no source event ever points
// at their RecurrenceGroupID) and a retry creates a second, independent batch.
func TestGenerateRecurrence_SourceUpdateFailureAllowsRetry(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	start := time.Date(2026, 1, 5, 18, 0, 0, 0, time.UTC)
	events.add(&domain.Event{ID: eventID, Name: "Training", Status: domain.EventStatusDraft, StartDate: start, EndDate: start.Add(time.Hour)})

	until := start.AddDate(0, 0, 21) // 3 weekly occurrences
	events.failNextMarkRecurring = true
	_, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyWeekly, until)
	require.Error(t, err)

	all, err := events.List(context.Background(), port.EventFilter{})
	require.NoError(t, err)
	assert.Len(t, all, 1, "no orphaned occurrence events after the final source-event update fails")

	result, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyWeekly, until)
	require.NoError(t, err)
	assert.Len(t, result, 4)
}

// Simulates the race where a concurrent call wins between this call's
// initial in-memory check (which necessarily still sees the event as not
// recurring, or this call wouldn't have gotten this far) and its final,
// atomic MarkRecurring: the atomic call itself — not the earlier read — is
// what must reject the loser, and the loser's already-created occurrences
// must be rolled back rather than left as orphans nothing points to.
func TestGenerateRecurrence_LosesRaceAtFinalStamp(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	start := time.Date(2026, 1, 5, 18, 0, 0, 0, time.UTC)
	events.add(&domain.Event{ID: eventID, Name: "Training", Status: domain.EventStatusDraft, StartDate: start, EndDate: start.Add(time.Hour)})

	until := start.AddDate(0, 0, 21) // 3 weekly occurrences
	events.forceNextMarkRecurringLost = true
	_, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyWeekly, until)
	assert.ErrorIs(t, err, domain.ErrAlreadyRecurring)

	stored, err := events.GetByID(context.Background(), eventID)
	require.NoError(t, err)
	assert.Nil(t, stored.RecurrenceGroupID)

	all, err := events.List(context.Background(), port.EventFilter{})
	require.NoError(t, err)
	assert.Len(t, all, 1, "the loser's occurrences must be rolled back, not left as orphans")
}

func TestGenerateRecurrence_RejectsCancelledEvent(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	start := time.Now().UTC()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusCancelled, StartDate: start, EndDate: start.Add(time.Hour)})

	_, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyWeekly, start.AddDate(0, 0, 7))
	assert.ErrorIs(t, err, domain.ErrInvalidRecurrence)
}

func TestGenerateRecurrence_RejectsCompletedEvent(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	start := time.Now().UTC()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusCompleted, StartDate: start, EndDate: start.Add(time.Hour)})

	_, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyWeekly, start.AddDate(0, 0, 7))
	assert.ErrorIs(t, err, domain.ErrInvalidRecurrence)
}

func TestGenerateRecurrence_TooManyOccurrences(t *testing.T) {
	uc, events, _, _, _, _, _ := newEventUC()
	eventID := uuid.New()
	start := time.Now().UTC()
	events.add(&domain.Event{ID: eventID, Status: domain.EventStatusDraft, StartDate: start, EndDate: start.Add(time.Hour)})

	// Weekly for 3 years is well beyond the 104-occurrence cap.
	_, err := uc.GenerateRecurrence(context.Background(), uuid.New(), eventID, domain.RecurrenceFrequencyWeekly, start.AddDate(3, 0, 0))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "too large")
}
