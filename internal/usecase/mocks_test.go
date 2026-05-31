package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

// This file provides lightweight in-memory fakes implementing the port interfaces,
// used across the usecase tests. They favour deterministic behaviour and easy
// assertions over completeness.

// --- MemberRepo fake ---

type fakeMemberRepo struct {
	members map[uuid.UUID]*domain.Member
	links   []*domain.OIDCLink
}

var _ port.MemberRepository = (*fakeMemberRepo)(nil)

func newFakeMemberRepo() *fakeMemberRepo {
	return &fakeMemberRepo{members: map[uuid.UUID]*domain.Member{}}
}

func (f *fakeMemberRepo) add(m *domain.Member) *domain.Member {
	cp := *m
	f.members[m.ID] = &cp
	return &cp
}

func (f *fakeMemberRepo) Create(ctx context.Context, m *domain.Member) error {
	cp := *m
	f.members[m.ID] = &cp
	return nil
}

func (f *fakeMemberRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Member, error) {
	m, ok := f.members[id]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	cp := *m
	return &cp, nil
}

func (f *fakeMemberRepo) GetByEmail(ctx context.Context, email string) (*domain.Member, error) {
	email = strings.ToLower(email)
	for _, m := range f.members {
		if strings.ToLower(m.Email) == email {
			cp := *m
			return &cp, nil
		}
	}
	return nil, domain.ErrMemberNotFound
}

func (f *fakeMemberRepo) GetByOIDCSubject(ctx context.Context, provider, subject string) (*domain.Member, error) {
	return nil, domain.ErrMemberNotFound
}

func (f *fakeMemberRepo) List(ctx context.Context, filter port.MemberFilter) ([]*domain.Member, error) {
	var out []*domain.Member
	for _, m := range f.members {
		if filter.IsActive != nil && m.IsActive != *filter.IsActive {
			continue
		}
		cp := *m
		out = append(out, &cp)
	}
	return out, nil
}

func (f *fakeMemberRepo) Update(ctx context.Context, m *domain.Member) error {
	if _, ok := f.members[m.ID]; !ok {
		return domain.ErrMemberNotFound
	}
	cp := *m
	f.members[m.ID] = &cp
	return nil
}

func (f *fakeMemberRepo) Deactivate(ctx context.Context, id uuid.UUID, leftAt time.Time) error {
	m, ok := f.members[id]
	if !ok {
		return domain.ErrMemberNotFound
	}
	m.IsActive = false
	m.LeftAt = &leftAt
	return nil
}

func (f *fakeMemberRepo) LinkOIDC(ctx context.Context, link *domain.OIDCLink) error {
	f.links = append(f.links, link)
	return nil
}

func (f *fakeMemberRepo) GetOIDCLinks(ctx context.Context, memberID uuid.UUID) ([]*domain.OIDCLink, error) {
	var out []*domain.OIDCLink
	for _, l := range f.links {
		if l.MemberID == memberID {
			out = append(out, l)
		}
	}
	return out, nil
}

func (f *fakeMemberRepo) Count(ctx context.Context) (int, error) {
	return len(f.members), nil
}

func (f *fakeMemberRepo) CountActive(ctx context.Context) (int, error) {
	n := 0
	for _, m := range f.members {
		if m.IsActive {
			n++
		}
	}
	return n, nil
}

func (f *fakeMemberRepo) SetReminderOptOut(ctx context.Context, id uuid.UUID, optOut bool) error {
	m, ok := f.members[id]
	if !ok {
		return domain.ErrMemberNotFound
	}
	m.ReminderOptOut = optOut
	return nil
}

func (f *fakeMemberRepo) Anonymize(ctx context.Context, id uuid.UUID, leftAt time.Time) error {
	m, ok := f.members[id]
	if !ok {
		return domain.ErrMemberNotFound
	}
	m.FirstName = "Geloeschtes"
	m.LastName = "Mitglied"
	m.Email = "deleted+" + id.String() + "@invalid.local"
	m.OIDCSubject = nil
	m.IsActive = false
	m.LeftAt = &leftAt
	return nil
}

// --- EventRepo fake ---

type fakeEventRepo struct {
	events map[uuid.UUID]*domain.Event
}

var _ port.EventRepository = (*fakeEventRepo)(nil)

func newFakeEventRepo() *fakeEventRepo {
	return &fakeEventRepo{events: map[uuid.UUID]*domain.Event{}}
}

func (f *fakeEventRepo) add(e *domain.Event) {
	cp := *e
	f.events[e.ID] = &cp
}

func (f *fakeEventRepo) Create(ctx context.Context, e *domain.Event) error {
	cp := *e
	f.events[e.ID] = &cp
	return nil
}

func (f *fakeEventRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	e, ok := f.events[id]
	if !ok {
		return nil, domain.ErrEventNotFound
	}
	cp := *e
	return &cp, nil
}

func (f *fakeEventRepo) List(ctx context.Context, filter port.EventFilter) ([]*domain.Event, error) {
	var out []*domain.Event
	for _, e := range f.events {
		cp := *e
		out = append(out, &cp)
	}
	return out, nil
}

func (f *fakeEventRepo) Update(ctx context.Context, e *domain.Event) error {
	if _, ok := f.events[e.ID]; !ok {
		return domain.ErrEventNotFound
	}
	cp := *e
	f.events[e.ID] = &cp
	return nil
}

func (f *fakeEventRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := f.events[id]; !ok {
		return domain.ErrEventNotFound
	}
	delete(f.events, id)
	return nil
}

func (f *fakeEventRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.EventStatus) error {
	e, ok := f.events[id]
	if !ok {
		return domain.ErrEventNotFound
	}
	e.Status = status
	return nil
}

// --- ShiftRepo fake ---

type fakeShiftRepo struct {
	shifts map[uuid.UUID]*domain.Shift
}

var _ port.ShiftRepository = (*fakeShiftRepo)(nil)

func newFakeShiftRepo() *fakeShiftRepo {
	return &fakeShiftRepo{shifts: map[uuid.UUID]*domain.Shift{}}
}

func (f *fakeShiftRepo) add(s *domain.Shift) {
	cp := *s
	f.shifts[s.ID] = &cp
}

func (f *fakeShiftRepo) Create(ctx context.Context, s *domain.Shift) error {
	cp := *s
	f.shifts[s.ID] = &cp
	return nil
}

func (f *fakeShiftRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Shift, error) {
	s, ok := f.shifts[id]
	if !ok {
		return nil, domain.ErrShiftNotFound
	}
	cp := *s
	return &cp, nil
}

func (f *fakeShiftRepo) FindByEventID(ctx context.Context, eventID uuid.UUID) ([]*domain.Shift, error) {
	var out []*domain.Shift
	for _, s := range f.shifts {
		if s.EventID == eventID {
			cp := *s
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakeShiftRepo) FindShiftsStartingBetween(ctx context.Context, from, to time.Time) ([]*domain.Shift, error) {
	var out []*domain.Shift
	for _, s := range f.shifts {
		if !s.StartAt.Before(from) && s.StartAt.Before(to) {
			cp := *s
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakeShiftRepo) FindUpcomingShifts(ctx context.Context, after time.Time) ([]*domain.Shift, error) {
	var out []*domain.Shift
	for _, s := range f.shifts {
		if !s.StartAt.Before(after) {
			cp := *s
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakeShiftRepo) Update(ctx context.Context, s *domain.Shift) error {
	if _, ok := f.shifts[s.ID]; !ok {
		return domain.ErrShiftNotFound
	}
	cp := *s
	f.shifts[s.ID] = &cp
	return nil
}

func (f *fakeShiftRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := f.shifts[id]; !ok {
		return domain.ErrShiftNotFound
	}
	delete(f.shifts, id)
	return nil
}

// --- RegistrationRepo fake ---

type fakeRegistrationRepo struct {
	regs map[uuid.UUID]*domain.Registration
}

var _ port.RegistrationRepository = (*fakeRegistrationRepo)(nil)

func newFakeRegistrationRepo() *fakeRegistrationRepo {
	return &fakeRegistrationRepo{regs: map[uuid.UUID]*domain.Registration{}}
}

func (f *fakeRegistrationRepo) add(r *domain.Registration) {
	cp := *r
	f.regs[r.ID] = &cp
}

func (f *fakeRegistrationRepo) Create(ctx context.Context, r *domain.Registration) error {
	cp := *r
	f.regs[r.ID] = &cp
	return nil
}

func (f *fakeRegistrationRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Registration, error) {
	r, ok := f.regs[id]
	if !ok {
		return nil, domain.ErrRegistrationNotFound
	}
	cp := *r
	return &cp, nil
}

func (f *fakeRegistrationRepo) GetByToken(ctx context.Context, token uuid.UUID) (*domain.Registration, error) {
	for _, r := range f.regs {
		if r.ConfirmationToken != nil && *r.ConfirmationToken == token {
			cp := *r
			return &cp, nil
		}
	}
	return nil, domain.ErrRegistrationNotFound
}

func (f *fakeRegistrationRepo) FindByShiftID(ctx context.Context, shiftID uuid.UUID) ([]*domain.Registration, error) {
	var out []*domain.Registration
	for _, r := range f.regs {
		if r.ShiftID == shiftID {
			cp := *r
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakeRegistrationRepo) FindByMemberID(ctx context.Context, memberID uuid.UUID) ([]*domain.Registration, error) {
	var out []*domain.Registration
	for _, r := range f.regs {
		if r.MemberID != nil && *r.MemberID == memberID {
			cp := *r
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakeRegistrationRepo) FindByMemberAndShift(ctx context.Context, memberID, shiftID uuid.UUID) (*domain.Registration, error) {
	for _, r := range f.regs {
		if r.MemberID != nil && *r.MemberID == memberID && r.ShiftID == shiftID {
			cp := *r
			return &cp, nil
		}
	}
	return nil, domain.ErrRegistrationNotFound
}

func (f *fakeRegistrationRepo) FindByGuestEmailAndShift(ctx context.Context, guestEmail string, shiftID uuid.UUID) (*domain.Registration, error) {
	guestEmail = strings.ToLower(guestEmail)
	for _, r := range f.regs {
		if r.GuestEmail != nil && strings.ToLower(*r.GuestEmail) == guestEmail && r.ShiftID == shiftID {
			cp := *r
			return &cp, nil
		}
	}
	return nil, domain.ErrRegistrationNotFound
}

func (f *fakeRegistrationRepo) CountActiveByShift(ctx context.Context, shiftID uuid.UUID) (int, error) {
	n := 0
	for _, r := range f.regs {
		if r.ShiftID == shiftID &&
			(r.State == domain.RegistrationStateRegistered || r.State == domain.RegistrationStateConfirmed) {
			n++
		}
	}
	return n, nil
}

func (f *fakeRegistrationRepo) Update(ctx context.Context, r *domain.Registration) error {
	if _, ok := f.regs[r.ID]; !ok {
		return domain.ErrRegistrationNotFound
	}
	cp := *r
	f.regs[r.ID] = &cp
	return nil
}

func (f *fakeRegistrationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if _, ok := f.regs[id]; !ok {
		return domain.ErrRegistrationNotFound
	}
	delete(f.regs, id)
	return nil
}

func (f *fakeRegistrationRepo) ListUnconfirmedExpiredReservations(ctx context.Context, before time.Time) ([]*domain.Registration, error) {
	var out []*domain.Registration
	for _, r := range f.regs {
		if r.State == domain.RegistrationStateReserved && r.ReservedUntil != nil && r.ReservedUntil.Before(before) {
			cp := *r
			out = append(out, &cp)
		}
	}
	return out, nil
}

// --- HourRepo fake ---

type fakeHourRepo struct {
	entries  map[uuid.UUID]*domain.HourEntry
	years    map[uuid.UUID]*domain.ClubYear
	targets  map[string]*domain.HourTarget // key memberID|yearID
	activeYr *domain.ClubYear
}

var _ port.HourRepository = (*fakeHourRepo)(nil)

func newFakeHourRepo() *fakeHourRepo {
	return &fakeHourRepo{
		entries: map[uuid.UUID]*domain.HourEntry{},
		years:   map[uuid.UUID]*domain.ClubYear{},
		targets: map[string]*domain.HourTarget{},
	}
}

func targetKey(memberID, yearID uuid.UUID) string {
	return memberID.String() + "|" + yearID.String()
}

func (f *fakeHourRepo) addYear(y *domain.ClubYear) {
	cp := *y
	f.years[y.ID] = &cp
	if y.IsActive {
		f.activeYr = &cp
	}
}

func (f *fakeHourRepo) CreateEntry(ctx context.Context, e *domain.HourEntry) error {
	cp := *e
	f.entries[e.ID] = &cp
	return nil
}

func (f *fakeHourRepo) GetEntryByID(ctx context.Context, id uuid.UUID) (*domain.HourEntry, error) {
	e, ok := f.entries[id]
	if !ok {
		return nil, domain.ErrHourEntryNotFound
	}
	cp := *e
	return &cp, nil
}

func (f *fakeHourRepo) FindEntriesByMemberAndYear(ctx context.Context, memberID, clubYearID uuid.UUID) ([]*domain.HourEntry, error) {
	var out []*domain.HourEntry
	for _, e := range f.entries {
		if e.MemberID == memberID && e.ClubYearID == clubYearID {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakeHourRepo) FindEntriesByYear(ctx context.Context, clubYearID uuid.UUID) ([]*domain.HourEntry, error) {
	var out []*domain.HourEntry
	for _, e := range f.entries {
		if e.ClubYearID == clubYearID {
			cp := *e
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (f *fakeHourRepo) UpdateEntry(ctx context.Context, e *domain.HourEntry) error {
	if _, ok := f.entries[e.ID]; !ok {
		return domain.ErrHourEntryNotFound
	}
	cp := *e
	f.entries[e.ID] = &cp
	return nil
}

func (f *fakeHourRepo) DeleteEntry(ctx context.Context, id uuid.UUID) error {
	if _, ok := f.entries[id]; !ok {
		return domain.ErrHourEntryNotFound
	}
	delete(f.entries, id)
	return nil
}

func (f *fakeHourRepo) CreateClubYear(ctx context.Context, y *domain.ClubYear) error {
	f.addYear(y)
	return nil
}

func (f *fakeHourRepo) GetActiveClubYear(ctx context.Context) (*domain.ClubYear, error) {
	if f.activeYr == nil {
		return nil, domain.ErrClubYearNotFound
	}
	cp := *f.activeYr
	return &cp, nil
}

func (f *fakeHourRepo) GetClubYearByID(ctx context.Context, id uuid.UUID) (*domain.ClubYear, error) {
	y, ok := f.years[id]
	if !ok {
		return nil, domain.ErrClubYearNotFound
	}
	cp := *y
	return &cp, nil
}

func (f *fakeHourRepo) ListClubYears(ctx context.Context) ([]*domain.ClubYear, error) {
	var out []*domain.ClubYear
	for _, y := range f.years {
		cp := *y
		out = append(out, &cp)
	}
	return out, nil
}

func (f *fakeHourRepo) GetHourTarget(ctx context.Context, memberID, clubYearID uuid.UUID) (*domain.HourTarget, error) {
	t, ok := f.targets[targetKey(memberID, clubYearID)]
	if !ok {
		return nil, nil
	}
	cp := *t
	return &cp, nil
}

func (f *fakeHourRepo) UpsertHourTarget(ctx context.Context, t *domain.HourTarget) error {
	cp := *t
	f.targets[targetKey(t.MemberID, t.ClubYearID)] = &cp
	return nil
}

// --- AuditRepo fake ---

type fakeAuditRepo struct {
	entries []*domain.AuditEntry
}

var _ port.AuditRepository = (*fakeAuditRepo)(nil)

func newFakeAuditRepo() *fakeAuditRepo {
	return &fakeAuditRepo{}
}

func (f *fakeAuditRepo) Insert(ctx context.Context, e *domain.AuditEntry) error {
	f.entries = append(f.entries, e)
	return nil
}

func (f *fakeAuditRepo) List(ctx context.Context, filter port.AuditFilter) ([]*domain.AuditEntry, error) {
	var out []*domain.AuditEntry
	for _, e := range f.entries {
		if filter.Entity != "" && e.Entity != filter.Entity {
			continue
		}
		out = append(out, e)
	}
	return out, nil
}

// countActions returns how many audit entries have the given action.
func (f *fakeAuditRepo) countActions(action string) int {
	n := 0
	for _, e := range f.entries {
		if e.Action == action {
			n++
		}
	}
	return n
}

// hasEntity returns true if any audit entry has the given entity and action.
func (f *fakeAuditRepo) has(action, entity string) bool {
	for _, e := range f.entries {
		if e.Action == action && e.Entity == entity {
			return true
		}
	}
	return false
}

// --- SettingsRepo fake ---

type fakeSettingsRepo struct {
	kv          map[string]string
	branding    *domain.BrandingConfig
	tiers       map[uuid.UUID][]*domain.FeeTier
	memberTiers map[string][]*domain.FeeTier // key memberID|yearID
}

var _ port.SettingsRepository = (*fakeSettingsRepo)(nil)

func newFakeSettingsRepo() *fakeSettingsRepo {
	return &fakeSettingsRepo{
		kv:          map[string]string{},
		tiers:       map[uuid.UUID][]*domain.FeeTier{},
		memberTiers: map[string][]*domain.FeeTier{},
	}
}

func (f *fakeSettingsRepo) GetSetting(ctx context.Context, key string) (string, error) {
	v, ok := f.kv[key]
	if !ok {
		return "", domain.ErrClubYearNotFound // any not-found sentinel; callers fall back to defaults
	}
	return v, nil
}

func (f *fakeSettingsRepo) SetSetting(ctx context.Context, key, value string) error {
	f.kv[key] = value
	return nil
}

func (f *fakeSettingsRepo) GetAllSettings(ctx context.Context) (map[string]string, error) {
	out := map[string]string{}
	for k, v := range f.kv {
		out[k] = v
	}
	return out, nil
}

func (f *fakeSettingsRepo) GetBranding(ctx context.Context) (*domain.BrandingConfig, error) {
	if f.branding == nil {
		b := domain.DefaultBranding()
		return &b, nil
	}
	cp := *f.branding
	return &cp, nil
}

func (f *fakeSettingsRepo) UpdateBranding(ctx context.Context, b *domain.BrandingConfig) error {
	cp := *b
	f.branding = &cp
	return nil
}

func (f *fakeSettingsRepo) GetFeeTiers(ctx context.Context, clubYearID uuid.UUID) ([]*domain.FeeTier, error) {
	return f.tiers[clubYearID], nil
}

func (f *fakeSettingsRepo) ReplaceFeeTiers(ctx context.Context, clubYearID uuid.UUID, tiers []*domain.FeeTier) error {
	f.tiers[clubYearID] = tiers
	return nil
}

func (f *fakeSettingsRepo) GetMemberFeeTiers(ctx context.Context, memberID, clubYearID uuid.UUID) ([]*domain.FeeTier, error) {
	return f.memberTiers[memberID.String()+"|"+clubYearID.String()], nil
}

func (f *fakeSettingsRepo) ReplaceMemberFeeTiers(ctx context.Context, memberID, clubYearID uuid.UUID, tiers []*domain.FeeTier) error {
	f.memberTiers[memberID.String()+"|"+clubYearID.String()] = tiers
	return nil
}

// --- EmailService fake ---

type sentEmail struct {
	kind      string
	to        string
	daysUntil int
}

type fakeEmailService struct {
	sent []sentEmail
}

var _ port.EmailService = (*fakeEmailService)(nil)

func (f *fakeEmailService) SendConfirmation(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event) error {
	f.sent = append(f.sent, sentEmail{kind: "confirmation", to: to})
	return nil
}

func (f *fakeEmailService) SendReminder(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event, daysUntil int) error {
	f.sent = append(f.sent, sentEmail{kind: "reminder", to: to, daysUntil: daysUntil})
	return nil
}

func (f *fakeEmailService) SendCancellation(ctx context.Context, to string, reg *domain.Registration, shift *domain.Shift, event *domain.Event) error {
	f.sent = append(f.sent, sentEmail{kind: "cancellation", to: to})
	return nil
}

func (f *fakeEmailService) SendKioskConfirmation(ctx context.Context, to string, confirmURL string, shift *domain.Shift, event *domain.Event) error {
	f.sent = append(f.sent, sentEmail{kind: "kiosk", to: to})
	return nil
}

func (f *fakeEmailService) SendMissingHoursWarning(ctx context.Context, to string, member *domain.Member, missingHours float64, year *domain.ClubYear) error {
	f.sent = append(f.sent, sentEmail{kind: "missing_hours", to: to})
	return nil
}

func (f *fakeEmailService) SendYearBilling(ctx context.Context, to string, member *domain.Member, missingHours float64, amountCents int, year *domain.ClubYear) error {
	f.sent = append(f.sent, sentEmail{kind: "year_billing", to: to})
	return nil
}

func (f *fakeEmailService) SendUnderstaffedNotice(ctx context.Context, to string, shift *domain.Shift, event *domain.Event) error {
	f.sent = append(f.sent, sentEmail{kind: "understaffed", to: to})
	return nil
}

func (f *fakeEmailService) SendByTemplate(ctx context.Context, to, templateName string, data map[string]any) error {
	f.sent = append(f.sent, sentEmail{kind: "template:" + templateName, to: to})
	return nil
}

// --- EmailTemplateRepo fake ---

type fakeEmailTemplateRepo struct {
	templates map[string]*domain.EmailTemplate
}

var _ port.EmailTemplateRepository = (*fakeEmailTemplateRepo)(nil)

func newFakeEmailTemplateRepo() *fakeEmailTemplateRepo {
	return &fakeEmailTemplateRepo{templates: map[string]*domain.EmailTemplate{}}
}

func (f *fakeEmailTemplateRepo) ListTemplates(ctx context.Context) ([]*domain.EmailTemplate, error) {
	var out []*domain.EmailTemplate
	for _, t := range f.templates {
		cp := *t
		out = append(out, &cp)
	}
	return out, nil
}

func (f *fakeEmailTemplateRepo) GetTemplate(ctx context.Context, name string) (*domain.EmailTemplate, error) {
	t, ok := f.templates[name]
	if !ok {
		return nil, domain.ErrMemberNotFound
	}
	cp := *t
	return &cp, nil
}

func (f *fakeEmailTemplateRepo) UpsertTemplate(ctx context.Context, t *domain.EmailTemplate) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	cp := *t
	f.templates[t.Name] = &cp
	return nil
}

// --- EmailLogRepo fake ---

type fakeEmailLogRepo struct {
	entries []*domain.EmailLogEntry
}

var _ port.EmailLogRepository = (*fakeEmailLogRepo)(nil)

func newFakeEmailLogRepo() *fakeEmailLogRepo {
	return &fakeEmailLogRepo{}
}

func (f *fakeEmailLogRepo) Insert(ctx context.Context, e *domain.EmailLogEntry) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	cp := *e
	f.entries = append(f.entries, &cp)
	return nil
}

func (f *fakeEmailLogRepo) List(ctx context.Context, limit, offset int) ([]*domain.EmailLogEntry, error) {
	return f.entries, nil
}

func (f *fakeEmailLogRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.EmailLogEntry, error) {
	for _, e := range f.entries {
		if e.ID == id {
			cp := *e
			return &cp, nil
		}
	}
	return nil, domain.ErrMemberNotFound
}

func (f *fakeEmailService) countKind(kind string) int {
	n := 0
	for _, e := range f.sent {
		if e.kind == kind {
			n++
		}
	}
	return n
}

// --- CacheService fake ---

type fakeCache struct {
	store map[string]string
}

var _ port.CacheService = (*fakeCache)(nil)

func newFakeCache() *fakeCache { return &fakeCache{store: map[string]string{}} }

func (f *fakeCache) Get(ctx context.Context, key string) (string, error) {
	v, ok := f.store[key]
	if !ok {
		return "", port.ErrCacheMiss
	}
	return v, nil
}

func (f *fakeCache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	f.store[key] = value
	return nil
}

func (f *fakeCache) Delete(ctx context.Context, key string) error {
	delete(f.store, key)
	return nil
}

func (f *fakeCache) Flush(ctx context.Context) error {
	f.store = map[string]string{}
	return nil
}
