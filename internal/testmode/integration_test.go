package testmode_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yoadey/shiftmanager/internal/adapter/db"
	"github.com/yoadey/shiftmanager/internal/testmode"
)

// ── helpers ──────────────────────────────────────────────────────────────────

func startServer(t *testing.T) *testmode.Server {
	t.Helper()
	s, err := testmode.New(context.Background())
	require.NoError(t, err)
	t.Cleanup(s.Close)
	return s
}

// token fetches a JWT from /api/v1/dev/token for the given role.
func token(t *testing.T, s *testmode.Server, role, id, email string) string {
	t.Helper()
	url := fmt.Sprintf("%s/api/v1/dev/token?role=%s&id=%s&email=%s", s.URL, role, id, email)
	resp, err := http.Get(url) //nolint:noctx
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body struct{ Token string `json:"token"` }
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.NotEmpty(t, body.Token)
	return body.Token
}

// get performs a GET with an optional Bearer token.
func get(t *testing.T, s *testmode.Server, path, tok string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, s.URL+path, nil)
	require.NoError(t, err)
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func post(t *testing.T, s *testmode.Server, path, tok, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, s.URL+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func put(t *testing.T, s *testmode.Server, path, tok, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPut, s.URL+path, strings.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

// uploadFile performs a multipart POST with a single "file" field.
func uploadFile(t *testing.T, s *testmode.Server, path, tok, fileName, contentType string, content []byte) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreatePart(map[string][]string{
		"Content-Disposition": {`form-data; name="file"; filename="` + fileName + `"`},
		"Content-Type":        {contentType},
	})
	require.NoError(t, err)
	_, err = part.Write(content)
	require.NoError(t, err)
	require.NoError(t, mw.Close())

	req, err := http.NewRequest(http.MethodPost, s.URL+path, &buf)
	require.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func del(t *testing.T, s *testmode.Server, path, tok string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, s.URL+path, nil)
	require.NoError(t, err)
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return resp
}

func decode(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(b, v), "body: %s", b)
}

// ── tests ─────────────────────────────────────────────────────────────────────

func TestHealthEndpoints(t *testing.T) {
	s := startServer(t)

	t.Run("healthz", func(t *testing.T) {
		resp := get(t, s, "/health", "")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})

	t.Run("readyz", func(t *testing.T) {
		resp := get(t, s, "/readyz", "")
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	})
}

func TestDevTokenEndpoint(t *testing.T) {
	s := startServer(t)

	t.Run("returns JWT for vorstand", func(t *testing.T) {
		tok := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")
		assert.NotEmpty(t, tok)
	})

	t.Run("returns JWT for mitglied", func(t *testing.T) {
		tok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")
		assert.NotEmpty(t, tok)
	})
}

func TestAuthMeEndpoint(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")

	resp := get(t, s, "/api/v1/auth/me", tok)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var me struct {
		ID    string `json:"id"`
		Role  string `json:"role"`
		Email string `json:"email"`
	}
	decode(t, resp, &me)
	assert.Equal(t, db.AdminID.String(), me.ID)
	assert.Equal(t, "vorstand", me.Role)
	assert.Equal(t, "admin@test.local", me.Email)
}

func TestAuthMeRequiresToken(t *testing.T) {
	s := startServer(t)
	resp := get(t, s, "/api/v1/auth/me", "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestAuthRefreshEndpoint(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")

	resp := post(t, s, "/api/v1/auth/refresh", tok, "")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body struct {
		Token     string `json:"token"`
		ExpiresIn int    `json:"expiresIn"`
	}
	decode(t, resp, &body)
	assert.NotEmpty(t, body.Token)
	assert.Positive(t, body.ExpiresIn)

	// The freshly issued token must itself be usable.
	meResp := get(t, s, "/api/v1/auth/me", body.Token)
	assert.Equal(t, http.StatusOK, meResp.StatusCode)
	meResp.Body.Close()
}

func TestAuthRefreshRequiresToken(t *testing.T) {
	s := startServer(t)
	resp := post(t, s, "/api/v1/auth/refresh", "", "")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	resp.Body.Close()
}

func TestListMembers(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")

	resp := get(t, s, "/api/v1/members", tok)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var members []map[string]any
	decode(t, resp, &members)
	require.Len(t, members, 2, "seed creates exactly 2 members")

	// Check that first names are present.
	var firstNames []string
	for _, m := range members {
		firstNames = append(firstNames, m["firstName"].(string))
	}
	assert.Contains(t, firstNames, "Admin")
	assert.Contains(t, firstNames, "Max")
}

func TestListMembers_PaginationAndStatus(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")

	list := func(query string) ([]map[string]any, string) {
		t.Helper()
		resp := get(t, s, "/api/v1/members"+query, tok)
		require.Equal(t, http.StatusOK, resp.StatusCode, query)
		total := resp.Header.Get("X-Total-Count")
		var members []map[string]any
		decode(t, resp, &members)
		return members, total
	}

	page, total := list("?limit=1&offset=1")
	require.Len(t, page, 1)
	assert.Equal(t, "2", total, "total ignores limit/offset")
	assert.Equal(t, "Admin", page[0]["firstName"], "sorted by last name: Mustermann, then Vorstand")

	matches, total := list("?search=max%20muster")
	require.Len(t, matches, 1, "search matches the full name")
	assert.Equal(t, "1", total)

	_, total = list("?status=pending")
	assert.Equal(t, "0", total)

	resp := del(t, s, "/api/v1/members/"+db.MemberID.String(), tok)
	require.Less(t, resp.StatusCode, 300)
	resp.Body.Close()

	_, total = list("?status=left")
	assert.Equal(t, "1", total)
	_, total = list("?status=active")
	assert.Equal(t, "1", total)
	_, total = list("?status=approved")
	assert.Equal(t, "2", total)
	_, total = list("?status=pending")
	assert.Equal(t, "0", total, "a deactivated member is not a pending registration")

	for _, q := range []string{"?status=bogus", "?limit=-1", "?offset=-5", "?limit=all"} {
		bad := get(t, s, "/api/v1/members"+q, tok)
		assert.Equal(t, http.StatusBadRequest, bad.StatusCode, q)
		bad.Body.Close()
	}
}

// Member data is for the board only; members see their own record (ML-010).
func TestMemberData_BoardOnly(t *testing.T) {
	s := startServer(t)
	member := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")
	leiter := token(t, s, "veranstaltungsleiter", uuid.NewString(), "leiter@test.local")

	for _, tok := range []string{member, leiter} {
		for _, path := range []string{
			"/api/v1/members",
			"/api/v1/members?status=pending",
			"/api/v1/members/export",
			"/api/v1/members/" + db.AdminID.String(),
		} {
			resp := get(t, s, path, tok)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode, path)
			resp.Body.Close()
		}
	}

	resp := get(t, s, "/api/v1/members/"+db.MemberID.String(), member)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "own record")
	resp.Body.Close()
}

// Organisers find helpers by name only (ML-014).
func TestMemberLookup(t *testing.T) {
	s := startServer(t)
	leiter := token(t, s, "veranstaltungsleiter", uuid.NewString(), "leiter@test.local")

	resp := get(t, s, "/api/v1/members/lookup?q=muster", leiter)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var hits []map[string]any
	decode(t, resp, &hits)
	require.Len(t, hits, 1)
	assert.Equal(t, map[string]any{"id": db.MemberID.String(), "name": "Max Mustermann"}, hits[0])

	// Helpers already on the shift are left out before the limit applies.
	resp = get(t, s, "/api/v1/members/lookup?q=muster&exclude="+db.MemberID.String(), leiter)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	decode(t, resp, &hits)
	assert.Empty(t, hits)

	for _, q := range []string{"?q=m", "?q=muster&exclude=nope"} {
		resp = get(t, s, "/api/v1/members/lookup"+q, leiter)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode, q)
		resp.Body.Close()
	}

	member := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")
	resp = get(t, s, "/api/v1/members/lookup?q=muster", member)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

// Other members' hour accounts are for the board only (SA-011).
func TestHourAccounts_BoardOnly(t *testing.T) {
	s := startServer(t)
	member := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")
	leiter := token(t, s, "veranstaltungsleiter", uuid.NewString(), "leiter@test.local")
	board := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")
	year := db.ClubYearID.String()

	for _, tok := range []string{member, leiter} {
		for _, path := range []string{
			"/api/v1/hours/" + db.AdminID.String(),
			"/api/v1/hours/account?memberId=" + db.AdminID.String() + "&clubYearId=" + year,
			"/api/v1/hours/summary?clubYearId=" + year,
		} {
			resp := get(t, s, path, tok)
			assert.Equal(t, http.StatusForbidden, resp.StatusCode, path)
			resp.Body.Close()
		}
	}

	for _, c := range []struct{ tok, path string }{
		{member, "/api/v1/hours/" + db.MemberID.String()},
		{member, "/api/v1/hours/account?memberId=" + db.MemberID.String() + "&clubYearId=" + year},
		{board, "/api/v1/hours/" + db.MemberID.String()},
		{board, "/api/v1/hours/summary?clubYearId=" + year},
	} {
		resp := get(t, s, c.path, c.tok)
		assert.Equal(t, http.StatusOK, resp.StatusCode, c.path)
		resp.Body.Close()
	}
}

// Members see only the names of other helpers on a shift (SC-011); the
// confirmation token never leaves the server (DS-009).
func TestTimeline_OnlyNamesOfOthers(t *testing.T) {
	s := startServer(t)
	shift := "/api/v1/shifts/" + db.Shift1ID.String()
	member := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")
	resp := post(t, s, shift+"/register", member, `{"comment": "komme später"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var reg map[string]any
	decode(t, resp, &reg)
	assert.NotContains(t, reg, "confirmationToken")

	leiter := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")
	resp = post(t, s, shift+"/add-guest", leiter, `{"name": "Hans Gast", "email": "hans@example.org"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	resp.Body.Close()

	regsOf := func(path, tok string) []map[string]any {
		t.Helper()
		resp := get(t, s, path, tok)
		require.Equal(t, http.StatusOK, resp.StatusCode, path)
		var tl struct {
			Days []struct {
				Shifts []struct {
					Shift         struct{ ID string }
					Registrations []map[string]any
				}
			}
		}
		decode(t, resp, &tl)
		for _, d := range tl.Days {
			for _, sh := range d.Shifts {
				if sh.Shift.ID == db.Shift1ID.String() {
					return sh.Registrations
				}
			}
		}
		t.Fatalf("shift not in timeline %s", path)
		return nil
	}
	byName := func(regs []map[string]any) map[string]map[string]any {
		out := map[string]map[string]any{}
		for _, r := range regs {
			assert.NotContains(t, r, "confirmationToken")
			out[r["displayName"].(string)] = r
		}
		return out
	}
	timeline := "/api/v1/events/" + db.EventID.String() + "/timeline"

	// Another member: abbreviated names (NM-004 default), nothing else.
	other := token(t, s, "mitglied", uuid.NewString(), "other@test.local")
	seen := byName(regsOf(timeline, other))
	require.Contains(t, seen, "Max M.")
	require.Contains(t, seen, "Hans G.")
	assert.Empty(t, seen["Max M."]["comment"])
	assert.NotContains(t, seen["Hans G."], "guestEmail")
	assert.NotContains(t, seen["Hans G."], "guestName", "the full guest name must not leak next to the abbreviated one")
	assert.NotContains(t, seen["Max M."], "memberId", "no ids to act on someone else's behalf")

	// Max sees his own registration in full.
	seen = byName(regsOf(timeline, member))
	assert.Equal(t, "komme später", seen["Max Mustermann"]["comment"])
	assert.Equal(t, db.MemberID.String(), seen["Max Mustermann"]["memberId"])

	// The organiser sees full names, comments and the guest's e-mail.
	seen = byName(regsOf(timeline, leiter))
	assert.Equal(t, "komme später", seen["Max Mustermann"]["comment"])
	assert.Equal(t, "hans@example.org", seen["Hans Gast"]["guestEmail"])

	// The anonymous kiosk gets names only.
	seen = byName(regsOf("/api/v1/kiosk/events/"+db.EventID.String(), ""))
	require.Contains(t, seen, "Max M.")
	assert.Empty(t, seen["Max M."]["comment"])
	assert.NotContains(t, seen["Hans G."], "guestEmail")
	assert.NotContains(t, seen["Hans G."], "guestName")

	// ... and only for published events; members never see drafts (V-006).
	resp = post(t, s, "/api/v1/events", leiter, `{"name": "Entwurf", "location": "Halle", "category": "Fest",
		"startDate": "2026-12-20T10:00:00Z", "endDate": "2026-12-20T20:00:00Z", "visibility": "public"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var draft map[string]any
	decode(t, resp, &draft)
	draftID := draft["id"].(string)
	for _, c := range []struct{ path, tok string }{
		{"/api/v1/kiosk/events/" + draftID, ""},
		{"/api/v1/events/" + draftID, other},
		{"/api/v1/events/" + draftID + "/timeline", other},
	} {
		resp = get(t, s, c.path, c.tok)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, c.path)
		resp.Body.Close()
	}
	resp = get(t, s, "/api/v1/events/"+draftID, leiter)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "organisers see their drafts")
	resp.Body.Close()

	listNames := func(tok string) []string {
		resp := get(t, s, "/api/v1/events?status=draft", tok)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		var events []map[string]any
		decode(t, resp, &events)
		var out []string
		for _, e := range events {
			out = append(out, e["name"].(string))
		}
		return out
	}
	assert.Empty(t, listNames(other), "members cannot list drafts")
	assert.Equal(t, []string{"Entwurf"}, listNames(leiter))
}

// The kiosk search answers only when enabled, with id and abbreviated name (K-002, K-003).
func TestKioskMemberSearch(t *testing.T) {
	s := startServer(t)
	resp := get(t, s, "/api/v1/kiosk/members?q=max", "")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "search disabled by default")
	resp.Body.Close()

	// Without the search the kiosk takes e-mail addresses only (K-002).
	register := fmt.Sprintf(`{"memberId": "%s"}`, db.MemberID)
	resp = post(t, s, "/api/v1/kiosk/shifts/"+db.Shift2ID.String()+"/register", "", register)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()

	board := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")
	resp = put(t, s, "/api/v1/settings", board, `{"kioskSearch": true}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()

	resp = get(t, s, "/api/v1/kiosk/members?q=max", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var hits []map[string]any
	decode(t, resp, &hits)
	require.Len(t, hits, 1)
	assert.Equal(t, map[string]any{"id": db.MemberID.String(), "name": "Max M."}, hits[0])

	resp = post(t, s, "/api/v1/kiosk/shifts/"+db.Shift2ID.String()+"/register", "", register)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "member picked via the enabled search")
	resp.Body.Close()

	// No lookups by e-mail address from the anonymous kiosk.
	resp = get(t, s, "/api/v1/kiosk/members?q=max@test.local", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	hits = nil
	decode(t, resp, &hits)
	assert.Empty(t, hits)
}

func TestGetMember(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")

	resp := get(t, s, "/api/v1/members/"+db.MemberID.String(), tok)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var m map[string]any
	decode(t, resp, &m)
	assert.Equal(t, "Max", m["firstName"])
	assert.Equal(t, "Mustermann", m["lastName"])
	assert.Equal(t, "max@test.local", m["email"])
}

func TestUpdatePreferences_NotifyNewEvents(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")

	resp := put(t, s, "/api/v1/members/me/preferences", tok, `{"reminderOptOut":false,"notifyNewEvents":true}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var body struct {
		ReminderOptOut  bool `json:"reminderOptOut"`
		NotifyNewEvents bool `json:"notifyNewEvents"`
	}
	decode(t, resp, &body)
	assert.False(t, body.ReminderOptOut)
	assert.True(t, body.NotifyNewEvents)

	// Persisted: reflected back via GET /members/{id}.
	getResp := get(t, s, "/api/v1/members/"+db.MemberID.String(), tok)
	require.Equal(t, http.StatusOK, getResp.StatusCode)
	var m map[string]any
	decode(t, getResp, &m)
	assert.Equal(t, true, m["notifyNewEvents"])
}

func TestListEvents(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")

	resp := get(t, s, "/api/v1/events", tok)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var events []map[string]any
	decode(t, resp, &events)
	require.Len(t, events, 1, "seed creates exactly 1 event")
	assert.Equal(t, "Sommerfest", events[0]["name"])
	assert.Equal(t, "published", events[0]["status"])
}

func TestListEvents_FiltersPaginationAndTotal(t *testing.T) {
	s := startServer(t)
	leiter := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")
	resp := post(t, s, "/api/v1/events", leiter, `{"name": "Winterfest", "location": "Halle", "category": "Fest",
		"startDate": "2026-12-20T10:00:00Z", "endDate": "2026-12-20T20:00:00Z", "visibility": "public"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var ev map[string]any
	decode(t, resp, &ev)
	require.Equal(t, http.StatusOK, post(t, s, "/api/v1/events/"+ev["id"].(string)+"/publish", leiter, "").StatusCode)

	tok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")
	names := func(path string) ([]string, string) {
		t.Helper()
		resp := get(t, s, path, tok)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		total := resp.Header.Get("X-Total-Count")
		var events []map[string]any
		decode(t, resp, &events)
		out := make([]string, len(events))
		for i, e := range events {
			out[i] = e["name"].(string)
		}
		return out, total
	}

	// The Sommerfest (4 July) is over by August.
	got, total := names("/api/v1/events?status=published&from=2026-08-01&order=asc")
	assert.Equal(t, []string{"Winterfest"}, got)
	assert.Equal(t, "1", total)

	// Paginated, earliest first, with the total behind the page.
	got, total = names("/api/v1/events?status=published&order=asc&limit=1&offset=0")
	assert.Equal(t, []string{"Sommerfest"}, got)
	assert.Equal(t, "2", total)
	got, _ = names("/api/v1/events?status=published&order=asc&limit=1&offset=1")
	assert.Equal(t, []string{"Winterfest"}, got)

	// "to" includes events starting later that day.
	got, _ = names("/api/v1/events?to=2026-07-04")
	assert.Equal(t, []string{"Sommerfest"}, got)

	got, _ = names("/api/v1/events?category=Fest")
	assert.Equal(t, []string{"Winterfest"}, got)

	// Only the Sommerfest has shifts, and they still have free places.
	got, _ = names("/api/v1/events?freeSlots=true")
	assert.Equal(t, []string{"Sommerfest"}, got)

	assert.Equal(t, http.StatusBadRequest, get(t, s, "/api/v1/events?order=up", tok).StatusCode)
	assert.Equal(t, http.StatusBadRequest, get(t, s, "/api/v1/events?to=28.10.2026", tok).StatusCode)
	assert.Equal(t, http.StatusBadRequest, get(t, s, "/api/v1/events?from=bogus", tok).StatusCode)
	assert.Equal(t, http.StatusBadRequest, get(t, s, "/api/v1/events?limit=-1", tok).StatusCode)
}

func TestGetEventTimeline(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")

	resp := get(t, s, "/api/v1/events/"+db.EventID.String(), tok)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var tl map[string]any
	decode(t, resp, &tl)

	ev, ok := tl["event"].(map[string]any)
	require.True(t, ok, "response must have an 'event' key")
	assert.Equal(t, "Sommerfest", ev["name"])

	days, ok := tl["days"].([]any)
	require.True(t, ok, "response must have a 'days' array")
	require.Len(t, days, 1)

	day := days[0].(map[string]any)
	shifts := day["shifts"].([]any)
	require.Len(t, shifts, 2, "seed creates 2 shifts on the same day")

	var shiftNames []string
	for _, sh := range shifts {
		shiftNames = append(shiftNames, sh.(map[string]any)["shift"].(map[string]any)["name"].(string))
	}
	assert.Contains(t, shiftNames, "Aufbau")
	assert.Contains(t, shiftNames, "Service")
}

func TestKioskEventsPublic(t *testing.T) {
	s := startServer(t)

	// Kiosk endpoint requires no authentication.
	resp := get(t, s, "/api/v1/kiosk/events", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var events []map[string]any
	decode(t, resp, &events)
	require.Len(t, events, 1)
	assert.Equal(t, "Sommerfest", events[0]["name"])
}

func TestKioskPublicTimeline(t *testing.T) {
	s := startServer(t)

	resp := get(t, s, "/api/v1/kiosk/events/"+db.EventID.String(), "")
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var tl map[string]any
	decode(t, resp, &tl)
	_, hasEvent := tl["event"]
	assert.True(t, hasEvent)
	_, hasDays := tl["days"]
	assert.True(t, hasDays)
}

func TestRegisterForShift(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")

	body := fmt.Sprintf(`{"memberId": "%s"}`, db.MemberID.String())
	resp := post(t, s, "/api/v1/shifts/"+db.Shift1ID.String()+"/register", tok, body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var reg map[string]any
	decode(t, resp, &reg)
	assert.NotEmpty(t, reg["id"])
}

func TestDoubleRegistrationFails(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")

	body := fmt.Sprintf(`{"memberId": "%s"}`, db.MemberID.String())
	resp1 := post(t, s, "/api/v1/shifts/"+db.Shift1ID.String()+"/register", tok, body)
	require.Equal(t, http.StatusCreated, resp1.StatusCode)
	resp1.Body.Close()

	resp2 := post(t, s, "/api/v1/shifts/"+db.Shift1ID.String()+"/register", tok, body)
	assert.Equal(t, http.StatusConflict, resp2.StatusCode)
	resp2.Body.Close()
}

func TestGetMyHours(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")

	resp := get(t, s, "/api/v1/hours/me", tok)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var account map[string]any
	decode(t, resp, &account)
	// Fresh account: no entries.
	assert.Equal(t, float64(0), account["confirmed"])
	entries, _ := account["entries"].([]any)
	assert.Len(t, entries, 0)
}

func TestGetMyHours_OpenFeeCents(t *testing.T) {
	s := startServer(t)
	board := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")
	resp := put(t, s, "/api/v1/settings/fee-tiers?clubYearId="+db.ClubYearID.String(), board,
		`{"tiers": [{"position": 1, "amountCents": 1000}, {"position": 2, "amountCents": 1500}]}`)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp = post(t, s, "/api/v1/hours/manual", board,
		fmt.Sprintf(`{"memberId": "%s", "hours": 5.5, "desc": "Aufräumen"}`, db.MemberID.String()))
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	// Goal 8 h, 5.5 h confirmed: 1 h × 10 € + 1.5 h × 15 €.
	member := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")
	var account map[string]any
	decode(t, get(t, s, "/api/v1/hours/me", member), &account)
	assert.Equal(t, float64(3250), account["openFeeCents"])

	decode(t, get(t, s, "/api/v1/hours/"+db.MemberID.String(), board), &account)
	assert.Equal(t, float64(3250), account["openFeeCents"], "the board sees what a member owes")

	other := token(t, s, "mitglied", uuid.NewString(), "other@test.local")
	resp = get(t, s, "/api/v1/hours/"+db.MemberID.String(), other)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "other members must not see what someone owes")
	resp.Body.Close()
}

func TestManualBooking(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")

	body := fmt.Sprintf(`{"memberId": "%s", "hours": 3.5, "desc": "Reinigung nach Fest"}`, db.MemberID.String())
	resp := post(t, s, "/api/v1/hours/manual", tok, body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var entry map[string]any
	decode(t, resp, &entry)
	assert.Equal(t, float64(3.5), entry["hours"])

	// Hours account should reflect the new entry.
	memberTok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")
	resp2 := get(t, s, "/api/v1/hours/me", memberTok)
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	var account map[string]any
	decode(t, resp2, &account)
	assert.Equal(t, float64(3.5), account["confirmed"], "manual booking is directly confirmed")
}

func TestGetStats(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	resp := get(t, s, "/api/v1/stats", tok)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var stats map[string]any
	decode(t, resp, &stats)
	assert.Equal(t, float64(2), stats["activeMembers"])
}

func TestOpenAPISpec(t *testing.T) {
	s := startServer(t)

	resp := get(t, s, "/api/v1/openapi.json", "")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	resp.Body.Close()
}

func TestCreateAndPublishEvent(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	// Create a draft event.
	body := `{
		"name": "Winterfest",
		"description": "Neues Fest",
		"location": "Halle",
		"category": "Fest",
		"startDate": "2026-12-20T10:00:00Z",
		"endDate":   "2026-12-20T20:00:00Z",
		"visibility": "public"
	}`
	resp := post(t, s, "/api/v1/events", tok, body)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var ev map[string]any
	decode(t, resp, &ev)
	evID := ev["id"].(string)
	assert.Equal(t, "draft", ev["status"])

	// Publish it.
	resp2 := post(t, s, "/api/v1/events/"+evID+"/publish", tok, "")
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	var published map[string]any
	decode(t, resp2, &published)
	assert.Equal(t, "published", published["status"])
}

func TestEventAttachments_UploadListDelete(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	uploadResp := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/attachments", tok, "flyer.png", "image/png", []byte("fake-png-bytes"))
	require.Equal(t, http.StatusCreated, uploadResp.StatusCode)
	var attachment map[string]any
	decode(t, uploadResp, &attachment)
	assert.Equal(t, "flyer.png", attachment["fileName"])
	assert.Equal(t, "image/png", attachment["contentType"])
	attachmentID := attachment["id"].(string)
	require.NotEmpty(t, attachmentID)

	listResp := get(t, s, "/api/v1/events/"+db.EventID.String()+"/attachments", tok)
	require.Equal(t, http.StatusOK, listResp.StatusCode)
	var list []map[string]any
	decode(t, listResp, &list)
	require.Len(t, list, 1)
	assert.Equal(t, attachmentID, list[0]["id"])

	// The uploaded file is actually downloadable from /uploads/*. The stored
	// URL is absolute against the handler's configured public base (a fixed
	// "http://localhost" in test mode), not the ephemeral httptest address,
	// so re-base its path onto the real server URL.
	fileURL, ok := attachment["url"].(string)
	require.True(t, ok)
	urlPath := fileURL[strings.LastIndex(fileURL, "/uploads/"):]
	fileResp, err := http.Get(s.URL + urlPath) //nolint:noctx
	require.NoError(t, err)
	fileBody, err := io.ReadAll(fileResp.Body)
	require.NoError(t, err)
	fileResp.Body.Close()
	assert.Equal(t, http.StatusOK, fileResp.StatusCode)
	assert.Equal(t, "fake-png-bytes", string(fileBody))
	// Guards against a browser content-sniffing an uploaded file into
	// something it isn't (e.g. a PNG/HTML polyglot) and executing it.
	assert.Equal(t, "nosniff", fileResp.Header.Get("X-Content-Type-Options"))

	delResp := del(t, s, "/api/v1/events/"+db.EventID.String()+"/attachments/"+attachmentID, tok)
	assert.Equal(t, http.StatusNoContent, delResp.StatusCode)

	listResp2 := get(t, s, "/api/v1/events/"+db.EventID.String()+"/attachments", tok)
	require.Equal(t, http.StatusOK, listResp2.StatusCode)
	var list2 []map[string]any
	decode(t, listResp2, &list2)
	assert.Empty(t, list2)
}

// Without this, http.FileServer serves an HTML index of every uploaded file
// (logos and event attachments, including ones on draft/unpublished events)
// to anyone, with no auth — bypassing the JWT-gated attachments API entirely.
func TestUploadsDirectoryListingIsBlocked(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")
	uploadResp := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/attachments", tok, "flyer.png", "image/png", []byte("fake-png-bytes"))
	require.Equal(t, http.StatusCreated, uploadResp.StatusCode)
	var attachment map[string]any
	decode(t, uploadResp, &attachment)

	resp, err := http.Get(s.URL + "/uploads/") //nolint:noctx
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	assert.NotContains(t, string(body), "flyer")
}

func TestEventAttachments_RejectsOversizedFile(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	oversized := make([]byte, 6<<20) // 6MB > the 5MB attachment limit
	resp := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/attachments", tok, "huge.png", "image/png", oversized)
	assert.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestEventAttachments_RejectsDisallowedType(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	resp := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/attachments", tok, "malware.exe", "application/octet-stream", []byte("x"))
	assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
}

// SVG can embed <script>, and this endpoint (unlike the Vorstand-only logo
// upload) is reachable by any Veranstaltungsleiter — must stay rejected.
func TestEventAttachments_RejectsSVG(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	resp := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/attachments", tok, "flyer.svg", "image/svg+xml", []byte("<svg><script>alert(1)</script></svg>"))
	assert.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
}

func TestEventAttachments_ListUnknownEvent404s(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	resp := get(t, s, "/api/v1/events/"+uuid.NewString()+"/attachments", tok)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestEventAttachments_RequiresVeranstaltungsleiter(t *testing.T) {
	s := startServer(t)
	memberTok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")

	resp := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/attachments", memberTok, "flyer.png", "image/png", []byte("x"))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestEventHeaderImage_UploadAndClear(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	uploadResp := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/header-image", tok, "header.png", "image/png", []byte("fake-header-bytes"))
	require.Equal(t, http.StatusOK, uploadResp.StatusCode)
	var ev map[string]any
	decode(t, uploadResp, &ev)
	fileURL, ok := ev["headerImageUrl"].(string)
	require.True(t, ok)
	require.NotEmpty(t, fileURL)

	// The event's GET response (nested under "event", GetWithTimeline) now
	// carries the header image too.
	getResp := get(t, s, "/api/v1/events/"+db.EventID.String(), tok)
	require.Equal(t, http.StatusOK, getResp.StatusCode)
	var got map[string]any
	decode(t, getResp, &got)
	gotEvent, ok := got["event"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, fileURL, gotEvent["headerImageUrl"])

	// Same public-uploads exposure as attachments (V-008/T-013): downloadable
	// without auth at its stored URL, re-based onto the real server address.
	urlPath := fileURL[strings.LastIndex(fileURL, "/uploads/"):]
	fileResp, err := http.Get(s.URL + urlPath) //nolint:noctx
	require.NoError(t, err)
	fileBody, err := io.ReadAll(fileResp.Body)
	require.NoError(t, err)
	fileResp.Body.Close()
	assert.Equal(t, http.StatusOK, fileResp.StatusCode)
	assert.Equal(t, "fake-header-bytes", string(fileBody))

	// Uploading again replaces it, not adds a second one, and deletes the
	// old file from storage instead of leaving it orphaned but still
	// publicly downloadable at its old URL.
	uploadResp2 := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/header-image", tok, "header2.png", "image/png", []byte("second-header"))
	require.Equal(t, http.StatusOK, uploadResp2.StatusCode)
	var ev2 map[string]any
	decode(t, uploadResp2, &ev2)
	assert.NotEqual(t, fileURL, ev2["headerImageUrl"])

	oldFileResp, err := http.Get(s.URL + urlPath) //nolint:noctx
	require.NoError(t, err)
	oldFileResp.Body.Close()
	assert.Equal(t, http.StatusNotFound, oldFileResp.StatusCode)

	delResp := del(t, s, "/api/v1/events/"+db.EventID.String()+"/header-image", tok)
	assert.Equal(t, http.StatusNoContent, delResp.StatusCode)

	getResp2 := get(t, s, "/api/v1/events/"+db.EventID.String(), tok)
	var cleared map[string]any
	decode(t, getResp2, &cleared)
	clearedEvent, ok := cleared["event"].(map[string]any)
	require.True(t, ok)
	assert.Empty(t, clearedEvent["headerImageUrl"])
}

func TestEventHeaderImage_RejectsSVGAndOversized(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	svgResp := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/header-image", tok, "header.svg", "image/svg+xml", []byte("<svg><script>alert(1)</script></svg>"))
	assert.Equal(t, http.StatusUnsupportedMediaType, svgResp.StatusCode)

	oversized := make([]byte, 6<<20) // 6MB > the 5MB header-image limit
	bigResp := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/header-image", tok, "huge.png", "image/png", oversized)
	assert.Equal(t, http.StatusRequestEntityTooLarge, bigResp.StatusCode)

	// Unlike the general attachment endpoint, PDF isn't a header image either.
	pdfResp := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/header-image", tok, "doc.pdf", "application/pdf", []byte("%PDF-"))
	assert.Equal(t, http.StatusUnsupportedMediaType, pdfResp.StatusCode)
}

func TestEventHeaderImage_RequiresVeranstaltungsleiter(t *testing.T) {
	s := startServer(t)
	memberTok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")

	resp := uploadFile(t, s, "/api/v1/events/"+db.EventID.String()+"/header-image", memberTok, "header.png", "image/png", []byte("x"))
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestUnauthorizedAccessDenied(t *testing.T) {
	s := startServer(t)

	// Member role cannot access manual booking.
	memberTok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")
	body := fmt.Sprintf(`{"memberId": "%s", "hours": 1, "desc": "test"}`, db.MemberID.String())
	resp := post(t, s, "/api/v1/hours/manual", memberTok, body)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}

// The seed event (db.EventID) starts 2026-07-04T14:00:00Z with two shifts.
func TestGenerateRecurrence_Weekly(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	resp := post(t, s, "/api/v1/events/"+db.EventID.String()+"/recurrence", tok,
		`{"frequency": "weekly", "until": "2026-07-26T00:00:00Z"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var events []map[string]any
	decode(t, resp, &events)
	require.Len(t, events, 4) // source + 3 weekly occurrences (07-11, 07-18, 07-25, all 14:00Z)
	assert.Equal(t, db.EventID.String(), events[0]["id"])
	assert.Equal(t, "weekly", events[0]["recurrenceFrequency"])
	require.NotEmpty(t, events[0]["recurrenceGroupId"])

	for _, occ := range events[1:] {
		assert.NotEqual(t, db.EventID.String(), occ["id"])
		assert.Equal(t, "Sommerfest", occ["name"])
		assert.Equal(t, "draft", occ["status"])
		assert.Equal(t, events[0]["recurrenceGroupId"], occ["recurrenceGroupId"])

		// Each occurrence must have carried over the source event's shifts.
		listResp := get(t, s, "/api/v1/events/"+occ["id"].(string)+"/timeline", tok)
		require.Equal(t, http.StatusOK, listResp.StatusCode)
		var tl map[string]any
		decode(t, listResp, &tl)
		days, _ := tl["days"].([]any)
		var shiftCount int
		for _, d := range days {
			day, _ := d.(map[string]any)
			shifts, _ := day["shifts"].([]any)
			shiftCount += len(shifts)
		}
		assert.Equal(t, 2, shiftCount)
	}
}

func TestGenerateRecurrence_AlreadyRecurring(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	first := post(t, s, "/api/v1/events/"+db.EventID.String()+"/recurrence", tok,
		`{"frequency": "weekly", "until": "2026-07-12T00:00:00Z"}`)
	require.Equal(t, http.StatusCreated, first.StatusCode)
	first.Body.Close()

	again := post(t, s, "/api/v1/events/"+db.EventID.String()+"/recurrence", tok,
		`{"frequency": "weekly", "until": "2026-07-25T00:00:00Z"}`)
	assert.Equal(t, http.StatusBadRequest, again.StatusCode)
}

func TestGenerateRecurrence_RequiresVeranstaltungsleiter(t *testing.T) {
	s := startServer(t)
	memberTok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")

	resp := post(t, s, "/api/v1/events/"+db.EventID.String()+"/recurrence", memberTok,
		`{"frequency": "weekly", "until": "2026-07-25T00:00:00Z"}`)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestGenerateRecurrence_UnknownEvent404s(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	resp := post(t, s, "/api/v1/events/"+uuid.NewString()+"/recurrence", tok,
		`{"frequency": "weekly", "until": "2026-07-25T00:00:00Z"}`)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// S-006: a club year opened with carryOverEnabled credits members' excess
// confirmed hours (beyond target) to whichever club year opens next.
func TestClubYear_CarriesOverExcessHours(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")

	yearAResp := post(t, s, "/api/v1/hours/club-years", tok,
		`{"label":"2027","startDate":"2027-01-01T00:00:00Z","endDate":"2027-12-31T23:59:59Z","defaultTargetHours":10,"setActive":true,"carryOverEnabled":true}`)
	require.Equal(t, http.StatusCreated, yearAResp.StatusCode)
	var yearA map[string]any
	decode(t, yearAResp, &yearA)
	yearAID := yearA["id"].(string)
	assert.Equal(t, true, yearA["carryOverEnabled"])

	bookResp := post(t, s, "/api/v1/hours/manual", tok, fmt.Sprintf(
		`{"memberId":"%s","clubYearId":"%s","hours":25,"description":"test"}`, db.MemberID.String(), yearAID))
	require.Equal(t, http.StatusCreated, bookResp.StatusCode)
	bookResp.Body.Close()

	yearBResp := post(t, s, "/api/v1/hours/club-years", tok,
		`{"label":"2028","startDate":"2028-01-01T00:00:00Z","endDate":"2028-12-31T23:59:59Z","defaultTargetHours":10,"setActive":true}`)
	require.Equal(t, http.StatusCreated, yearBResp.StatusCode)
	var yearB map[string]any
	decode(t, yearBResp, &yearB)
	yearBID := yearB["id"].(string)

	acctResp := get(t, s, "/api/v1/hours/account?memberId="+db.MemberID.String()+"&clubYearId="+yearBID, tok)
	require.Equal(t, http.StatusOK, acctResp.StatusCode)
	var acct map[string]any
	decode(t, acctResp, &acct)
	// 25h confirmed in year A against a 10h target -> 15h excess carried over.
	assert.Equal(t, 15.0, acct["confirmedHours"])
}

func TestClubYear_NoCarryOverWhenNotEnabled(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")

	yearAResp := post(t, s, "/api/v1/hours/club-years", tok,
		`{"label":"2027","startDate":"2027-01-01T00:00:00Z","endDate":"2027-12-31T23:59:59Z","defaultTargetHours":10,"setActive":true}`)
	require.Equal(t, http.StatusCreated, yearAResp.StatusCode)
	var yearA map[string]any
	decode(t, yearAResp, &yearA)
	yearAID := yearA["id"].(string)
	assert.Equal(t, false, yearA["carryOverEnabled"])

	bookResp := post(t, s, "/api/v1/hours/manual", tok, fmt.Sprintf(
		`{"memberId":"%s","clubYearId":"%s","hours":25,"description":"test"}`, db.MemberID.String(), yearAID))
	require.Equal(t, http.StatusCreated, bookResp.StatusCode)
	bookResp.Body.Close()

	yearBResp := post(t, s, "/api/v1/hours/club-years", tok,
		`{"label":"2028","startDate":"2028-01-01T00:00:00Z","endDate":"2028-12-31T23:59:59Z","defaultTargetHours":10,"setActive":true}`)
	require.Equal(t, http.StatusCreated, yearBResp.StatusCode)
	var yearB map[string]any
	decode(t, yearBResp, &yearB)
	yearBID := yearB["id"].(string)

	acctResp := get(t, s, "/api/v1/hours/account?memberId="+db.MemberID.String()+"&clubYearId="+yearBID, tok)
	require.Equal(t, http.StatusOK, acctResp.StatusCode)
	var acct map[string]any
	decode(t, acctResp, &acct)
	assert.Equal(t, 0.0, acct["confirmedHours"])
}

// A new active club year must supersede the previously active one, so
// GET /hours/club-years never shows more than one year marked active at once.
func TestClubYear_ActivatingNewYearDeactivatesPrevious(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "vorstand", db.AdminID.String(), "admin@test.local")

	yearAResp := post(t, s, "/api/v1/hours/club-years", tok,
		`{"label":"2027","startDate":"2027-01-01T00:00:00Z","endDate":"2027-12-31T23:59:59Z","defaultTargetHours":10,"setActive":true}`)
	require.Equal(t, http.StatusCreated, yearAResp.StatusCode)
	var yearA map[string]any
	decode(t, yearAResp, &yearA)

	yearBResp := post(t, s, "/api/v1/hours/club-years", tok,
		`{"label":"2028","startDate":"2028-01-01T00:00:00Z","endDate":"2028-12-31T23:59:59Z","defaultTargetHours":10,"setActive":true}`)
	require.Equal(t, http.StatusCreated, yearBResp.StatusCode)

	listResp := get(t, s, "/api/v1/hours/club-years", tok)
	require.Equal(t, http.StatusOK, listResp.StatusCode)
	var years []map[string]any
	decode(t, listResp, &years)

	activeCount := 0
	for _, y := range years {
		if active, _ := y["isActive"].(bool); active {
			activeCount++
		}
	}
	assert.Equal(t, 1, activeCount, "exactly one club year must be active at a time")
}

func TestAddGuestToShift(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	resp := post(t, s, "/api/v1/shifts/"+db.Shift1ID.String()+"/add-guest", tok,
		`{"name":"Jane Doe","email":"jane@example.com"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var reg map[string]any
	decode(t, resp, &reg)
	assert.Equal(t, "Jane Doe", reg["guestName"])
	assert.Equal(t, "jane@example.com", reg["guestEmail"])
	assert.Nil(t, reg["memberId"])
}

func TestAddGuestToShift_EmailOptional(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	resp := post(t, s, "/api/v1/shifts/"+db.Shift1ID.String()+"/add-guest", tok, `{"name":"Jane Doe"}`)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var reg map[string]any
	decode(t, resp, &reg)
	assert.Equal(t, "Jane Doe", reg["guestName"])
	assert.Nil(t, reg["guestEmail"])
}

func TestAddGuestToShift_RequiresName(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	resp := post(t, s, "/api/v1/shifts/"+db.Shift1ID.String()+"/add-guest", tok, `{}`)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAddGuestToShift_RequiresVeranstaltungsleiter(t *testing.T) {
	s := startServer(t)
	memberTok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")

	resp := post(t, s, "/api/v1/shifts/"+db.Shift1ID.String()+"/add-guest", memberTok, `{"name":"Jane Doe"}`)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// A guest helper an organizer added should be removable, same as a
// self-registered member (ForceDeleteRegistration is not scoped to
// self-service registrations).
func TestAddGuestToShift_RejectsDuplicateEmail(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	first := post(t, s, "/api/v1/shifts/"+db.Shift1ID.String()+"/add-guest", tok, `{"name":"Jane Doe","email":"jane@example.com"}`)
	require.Equal(t, http.StatusCreated, first.StatusCode)
	first.Body.Close()

	dup := post(t, s, "/api/v1/shifts/"+db.Shift1ID.String()+"/add-guest", tok, `{"name":"J. Doe","email":"Jane@Example.com"}`)
	assert.Equal(t, http.StatusConflict, dup.StatusCode)
}

func TestAddGuestToShift_CanBeRemoved(t *testing.T) {
	s := startServer(t)
	tok := token(t, s, "veranstaltungsleiter", db.AdminID.String(), "admin@test.local")

	addResp := post(t, s, "/api/v1/shifts/"+db.Shift1ID.String()+"/add-guest", tok, `{"name":"Jane Doe"}`)
	require.Equal(t, http.StatusCreated, addResp.StatusCode)
	var reg map[string]any
	decode(t, addResp, &reg)
	regID := reg["id"].(string)

	delResp := del(t, s, "/api/v1/registrations/"+regID, tok)
	assert.Equal(t, http.StatusNoContent, delResp.StatusCode)
}
