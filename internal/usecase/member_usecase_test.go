package usecase

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

func newMemberUC() (*MemberUsecase, *fakeMemberRepo, *fakeAuditRepo) {
	members := newFakeMemberRepo()
	audit := newFakeAuditRepo()
	return NewMemberUsecase(members, audit), members, audit
}

func TestCreateMember(t *testing.T) {
	uc, _, audit := newMemberUC()
	m, err := uc.CreateMember(context.Background(), uuid.New(), CreateMemberInput{
		FirstName: "Max", LastName: "Müller", Email: "  Max@Example.COM ",
	})
	require.NoError(t, err)
	assert.Equal(t, "max@example.com", m.Email) // normalised
	assert.Equal(t, domain.RoleMitglied, m.Role)
	assert.True(t, m.IsActive)
	assert.True(t, audit.has(domain.AuditActionCreate, domain.AuditEntityMember))
}

func TestCreateMember_Validation(t *testing.T) {
	uc, _, _ := newMemberUC()
	_, err := uc.CreateMember(context.Background(), uuid.New(), CreateMemberInput{Email: ""})
	assert.Error(t, err)
	_, err = uc.CreateMember(context.Background(), uuid.New(), CreateMemberInput{Email: "a@b.de", FirstName: "Max"})
	assert.Error(t, err)
}

func TestCreateMember_EmailConflict(t *testing.T) {
	uc, members, _ := newMemberUC()
	members.add(&domain.Member{ID: uuid.New(), Email: "dup@b.de", FirstName: "A", LastName: "B", IsActive: true})
	_, err := uc.CreateMember(context.Background(), uuid.New(), CreateMemberInput{FirstName: "C", LastName: "D", Email: "dup@b.de"})
	assert.ErrorIs(t, err, domain.ErrMemberEmailConflict)
}

func TestUpdateMember(t *testing.T) {
	uc, members, audit := newMemberUC()
	id := uuid.New()
	members.add(&domain.Member{ID: id, FirstName: "Old", LastName: "Name", Email: "o@b.de", IsActive: true})
	goal := 30.0
	m, err := uc.UpdateMember(context.Background(), uuid.New(), id, UpdateMemberInput{FirstName: "New", IndividualGoalHours: &goal, Role: domain.RoleVorstand})
	require.NoError(t, err)
	assert.Equal(t, "New", m.FirstName)
	assert.Equal(t, "Name", m.LastName)
	require.NotNil(t, m.IndividualGoalHours)
	assert.Equal(t, 30.0, *m.IndividualGoalHours)
	assert.Equal(t, domain.RoleVorstand, m.Role)
	assert.True(t, audit.has(domain.AuditActionUpdate, domain.AuditEntityMember))
}

func TestDeactivateMember(t *testing.T) {
	uc, members, audit := newMemberUC()
	id := uuid.New()
	members.add(&domain.Member{ID: id, FirstName: "A", LastName: "B", Email: "a@b.de", IsActive: true})

	err := uc.DeactivateMember(context.Background(), uuid.New(), id)
	require.NoError(t, err)
	m, _ := members.GetByID(context.Background(), id)
	assert.False(t, m.IsActive)
	assert.True(t, audit.has(domain.AuditActionDeactivate, domain.AuditEntityMember))

	// Deactivating again returns inactive error.
	err = uc.DeactivateMember(context.Background(), uuid.New(), id)
	assert.ErrorIs(t, err, domain.ErrMemberInactive)
}

func TestLinkOIDC(t *testing.T) {
	uc, members, audit := newMemberUC()
	id := uuid.New()
	members.add(&domain.Member{ID: id, Email: "a@b.de", IsActive: true})
	err := uc.LinkOIDC(context.Background(), uuid.New(), id, "google", "subject-123")
	require.NoError(t, err)
	links, _ := members.GetOIDCLinks(context.Background(), id)
	assert.Len(t, links, 1)
	assert.True(t, audit.has(domain.AuditActionLinkOIDC, domain.AuditEntityMember))
}

func TestImportCSV_UpsertByEmail(t *testing.T) {
	uc, members, audit := newMemberUC()
	// Pre-existing member to be updated.
	existingID := uuid.New()
	members.add(&domain.Member{ID: existingID, FirstName: "Old", LastName: "Surname", Email: "keep@b.de", IsActive: true})

	csv := "first_name,last_name,email,individual_goal_hours\n" +
		"Neu,Person,new@b.de,40\n" +
		"Updated,Surname,keep@b.de,\n"

	result, err := uc.ImportCSV(context.Background(), uuid.New(), strings.NewReader(csv))
	require.NoError(t, err)
	assert.Len(t, result.ToCreate, 1)
	assert.Len(t, result.ToUpdate, 1)

	// New created.
	created, err := members.GetByEmail(context.Background(), "new@b.de")
	require.NoError(t, err)
	require.NotNil(t, created.IndividualGoalHours)
	assert.Equal(t, 40.0, *created.IndividualGoalHours)

	// Existing updated by email (same ID).
	updated, err := members.GetByID(context.Background(), existingID)
	require.NoError(t, err)
	assert.Equal(t, "Updated", updated.FirstName)
	assert.True(t, audit.has(domain.AuditActionImport, domain.AuditEntityMember))
}

func TestImportCSVPreview(t *testing.T) {
	uc, members, _ := newMemberUC()
	members.add(&domain.Member{ID: uuid.New(), FirstName: "Same", LastName: "Name", Email: "same@b.de", IsActive: true})

	csv := "first_name,last_name,email\n" +
		"Same,Name,same@b.de\n" + // unchanged -> skip
		"Changed,Name,same@b.de\n" + // wait, same email; treat by email lookup
		"Brand,New,fresh@b.de\n"
	_ = csv
	preview, err := uc.ImportCSVPreview(context.Background(), strings.NewReader("first_name,last_name,email\nSame,Name,same@b.de\nBrand,New,fresh@b.de\n"))
	require.NoError(t, err)
	assert.Len(t, preview.ToSkip, 1)
	assert.Len(t, preview.ToCreate, 1)
}

func TestExportCSV(t *testing.T) {
	uc, members, audit := newMemberUC()
	members.add(&domain.Member{ID: uuid.New(), FirstName: "A", LastName: "B", Email: "a@b.de", IsActive: true, Role: domain.RoleMitglied})
	data, err := uc.ExportCSV(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Contains(t, string(data), "a@b.de")
	assert.Contains(t, string(data), "first_name")
	assert.True(t, audit.has(domain.AuditActionExport, domain.AuditEntityMember))
}

func TestListMembers_FilterActive(t *testing.T) {
	uc, members, _ := newMemberUC()
	members.add(&domain.Member{ID: uuid.New(), Email: "active@b.de", IsActive: true})
	members.add(&domain.Member{ID: uuid.New(), Email: "inactive@b.de", IsActive: false})
	active := true
	list, err := uc.ListMembers(context.Background(), port.MemberFilter{IsActive: &active})
	require.NoError(t, err)
	assert.Len(t, list, 1)
}
