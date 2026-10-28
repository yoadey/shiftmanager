package db

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/yoadey/shiftmanager/internal/domain"
	"github.com/yoadey/shiftmanager/internal/port"
)

func newTestMemberRepo(t *testing.T) *MemberRepo {
	t.Helper()
	gdb, err := Open("sqlite::memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(gdb); err != nil {
		t.Fatal(err)
	}
	return NewMemberRepo(gdb)
}

func createTestMember(t *testing.T, r *MemberRepo, first, last, email string) *domain.Member {
	t.Helper()
	m := &domain.Member{
		ID: uuid.New(), FirstName: first, LastName: last, Email: email,
		JoinedAt: time.Now().UTC(), IsActive: true, Role: domain.RoleMitglied,
	}
	if err := r.Create(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMemberRepoUpdate_PersistsGoalAndSubject(t *testing.T) {
	r := newTestMemberRepo(t)
	ctx := context.Background()
	m := createTestMember(t, r, "Max", "Muster", "max@example.org")

	goal := 12.5
	sub := "subject-1"
	m.IndividualGoalHours = &goal
	m.OIDCSubject = &sub
	if err := r.Update(ctx, m); err != nil {
		t.Fatal(err)
	}

	got, err := r.GetByID(ctx, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.IndividualGoalHours == nil || *got.IndividualGoalHours != goal {
		t.Errorf("individual goal = %v, want %v", got.IndividualGoalHours, goal)
	}
	if got.OIDCSubject == nil || *got.OIDCSubject != sub {
		t.Errorf("oidc subject = %v, want %q", got.OIDCSubject, sub)
	}
}

func TestMemberRepoList_SearchTreatsWildcardsLiterally(t *testing.T) {
	r := newTestMemberRepo(t)
	ctx := context.Background()
	createTestMember(t, r, "Anna", "Schmidt", "anna@example.org")
	under := createTestMember(t, r, "Ben", "Under", "ben_under@example.org")

	for _, search := range []string{"_", "%"} {
		got, err := r.List(ctx, port.MemberFilter{Search: search})
		if err != nil {
			t.Fatal(err)
		}
		want := 0
		if search == "_" {
			want = 1
		}
		if len(got) != want {
			t.Fatalf("search %q: got %d members, want %d", search, len(got), want)
		}
		if want == 1 && got[0].ID != under.ID {
			t.Errorf("search %q matched %s, want %s", search, got[0].Email, under.Email)
		}
	}
}
