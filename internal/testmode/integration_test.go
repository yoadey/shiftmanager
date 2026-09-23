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
