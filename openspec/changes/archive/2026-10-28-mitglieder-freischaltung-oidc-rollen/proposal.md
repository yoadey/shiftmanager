# Proposal

## Why

Unbekannte OIDC-Nutzer werden beim ersten Login bereits automatisch als
inaktives Mitglied angelegt, damit der Vorstand sie freischalten kann.
In der Praxis gehen diese Registrierungen aber unter:

- Die Mitgliederliste lädt nur die ersten 50 Einträge (Backend-Default von
  `GET /api/v1/members`, das Frontend setzt kein `limit`). Bei mehr als 50
  Mitgliedern fehlt jeder dahinter alphabetisch einsortierte Nutzer, und
  die rein clientseitige Suche findet ihn auch nicht.
- In der Liste ist nicht erkennbar, wer auf Freischaltung wartet; das
  Badge „Nicht freigeschaltet“ erscheint erst in der Detailansicht und
  unterscheidet nicht zwischen neuen Registrierungen und ausgetretenen
  Mitgliedern.
- Niemand wird darauf hingewiesen, dass eine Registrierung wartet.

Außerdem pflegen viele Vereine ihre Gruppen/Rollen bereits im
Identity-Provider. Für solche Nutzer ist die manuelle Freischaltung
überflüssig: Wie bei Grafanas `role_attribute_path` soll ein
konfigurierbarer Ausdruck die Rolle aus den OIDC-Claims ableiten und das
Mitglied direkt zulassen.

## What Changes

- Die Mitgliederliste (`GET /api/v1/members`) wird serverseitig
  paginiert, durchsucht und nach Kontostatus gefiltert (`status=active|
  pending|left|approved`); die Gesamtzahl der Treffer kommt im Header
  `X-Total-Count`. `limit=0` liefert weiterhin alle Treffer (für
  Namensauflösung und Auswahllisten).
- „Wartet auf Freischaltung“ wird ein eindeutiger Zustand: inaktiv **ohne**
  Austrittsdatum. Jede Deaktivierung setzt ein Austrittsdatum, damit
  ausgetretene Mitglieder nicht als neue Registrierungen erscheinen.
- Die Mitgliederübersicht zeigt wartende Registrierungen in einem eigenen
  Abschnitt oben mit „Freischalten“ und „Ablehnen“ (mit Rückfrage); darunter
  die paginierte Liste mit Seitenwechsel (Seite in der URL, `?seite=`) und
  Kennzeichnung „Ausgetreten“.
- Vorstand und Admin sehen oben in der App einen Hinweis, sobald
  Registrierungen auf Freischaltung warten.
- Die Detailansicht unterscheidet „wartet auf Freischaltung“
  (Freischalten/Ablehnen) und „ausgetreten“ (Reaktivieren).
- Neue Konfiguration `OIDC_ROLE_ATTRIBUTE_PATH` (bzw.
  `OIDC_<N>_ROLE_ATTRIBUTE_PATH` pro Provider) und Helm-Wert
  `app.oidc.roleAttributePath`: ein JMESPath-Ausdruck, der aus den Claims
  von ID-Token bzw. UserInfo eine Rolle ableitet. Liefert er eine gültige
  Rolle, wird ein neuer oder noch wartender Nutzer beim Login sofort mit
  dieser Rolle aktiviert.

## Capabilities

### New Capabilities
(keine)

### Modified Capabilities
- `mitgliederverwaltung`: neue Requirements für Selbstregistrierung mit
  Freischaltung (ML-011), paginierte Mitgliederliste (ML-012) und Hinweis
  auf wartende Registrierungen (ML-013); ML-001 präzisiert, dass eine
  Deaktivierung das Austrittsdatum setzt; ML-010 schließt wartende
  Registrierungen für Nutzer ohne Vorstandsrolle aus.
- `auth`: neues Requirement Rollenzuordnung über OIDC-Claims (A-006).

## Impact

- Backend: `internal/domain/member.go` (`IsPendingApproval`, `ParseRole`),
  `internal/port` (`MemberFilter.Status`, `CountFiltered`,
  `OIDCClaims.Role`, geänderte Signatur `VerifyIDToken`),
  `internal/adapter/db/member_repo.go`, `internal/usecase/member_usecase.go`,
  `internal/adapter/http/handler/{member,auth,kiosk}_handler.go`,
  `internal/adapter/oidc` (neuer `RoleMapper`), `internal/config`,
  `cmd/server/main.go`.
- API: `api/openapi.yaml` (`listMembers`: Parameter `status`, `limit`,
  `offset`, Header `X-Total-Count`; Parametername `active` statt des nie
  ausgewerteten `isActive`), generierte Typen neu erzeugt.
- Frontend: `api/members.ts` (`useMemberPage`, `usePendingMembers`),
  `screens/admin/AdminMembers.tsx`, `screens/admin/MemberDetail.tsx`,
  neue `components/layout/PendingMembersBanner.tsx`, `App.tsx`.
- Deployment: Helm-Chart (`values.yaml`, `configmap.yaml`, README),
  `.env.example`, `README.md`.
- Neue Abhängigkeit: `github.com/jmespath/go-jmespath`.
