package testmode_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

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

func TestUnauthorizedAccessDenied(t *testing.T) {
	s := startServer(t)

	// Member role cannot access manual booking.
	memberTok := token(t, s, "mitglied", db.MemberID.String(), "max@test.local")
	body := fmt.Sprintf(`{"memberId": "%s", "hours": 1, "desc": "test"}`, db.MemberID.String())
	resp := post(t, s, "/api/v1/hours/manual", memberTok, body)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	resp.Body.Close()
}
