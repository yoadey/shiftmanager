# 01 — Authentifizierung, Mitgliederverwaltung, Namensanzeige

Umgesetzt. Anforderungs-IDs: `A-001`–`A-004`, `ML-001`–`ML-009`,
`NM-001`–`NM-007`. Offen: `A-005` (siehe [`changes/A-005-multi-oidc-provider.md`](changes/A-005-multi-oidc-provider.md)).

---

## Authentifizierung (OIDC)

Die Anmeldung erfolgt ausschließlich über OpenID Connect; es werden keine
Passwörter im System gespeichert.

| ID | Anforderung | Umsetzung |
|---|---|---|
| A-001 | Konfigurierbarer OIDC-Provider (Keycloak, Auth0, Azure AD, …) | `internal/adapter/oidc`, `go-oidc` v3 |
| A-002 | Client-ID/-Secret/Discovery-URL über Umgebungsvariablen | `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL` |
| A-003 | Internes JWT-Session-Token nach OIDC-Login | `internal/adapter/http/handler/auth_handler.go`, `JWT_SECRET` |
| A-004 | Automatischer Token-Refresh während aktiver Session | `frontend/src/hooks/useTokenRefresh.ts` (Soft-Refresh vor Ablauf) + `POST /api/v1/auth/refresh` |

**Ablauf:** Frontend → `GET /api/v1/auth/login` → OIDC-Provider →
`GET /api/v1/auth/callback` (Code-Exchange, JWT-Ausstellung) →
Weiterleitung an `LOGIN_REDIRECT_URL` mit Token im URL-Fragment →
`CallbackPage` speichert das JWT; alle folgenden Requests senden
`Authorization: Bearer <token>`.

`BOOTSTRAP_ADMIN_EMAIL` befördert beim ersten Login automatisch den
angegebenen Nutzer zu Admin (einmaliger Bootstrap).

`TEST_MODE=true` deaktiviert Rate-Limiting und stellt
`/api/v1/dev/token` bereit, um Test-JWTs mit beliebiger Rolle
auszustellen — nie in Produktion aktiv.

---

## Mitgliederverwaltung

Die Mitgliederliste wird unabhängig vom OIDC-Provider gepflegt und kann
manuell oder automatisch mit einem OIDC-Benutzer verknüpft werden, sodass
Stunden auch für Mitglieder ohne aktives OIDC-Konto erfasst werden können.

| ID | Anforderung | Umsetzung |
|---|---|---|
| ML-001 | Vorstand/Admin legen Mitglieder an, bearbeiten, deaktivieren | `POST/PUT/DELETE /api/v1/members` (Vorstand-only), `AdminMembers.tsx` |
| ML-002 | Pflichtfelder: Vorname, Nachname, E-Mail, Eintrittsdatum, optional Austrittsdatum | `domain.Member` |
| ML-003 | Eintrittsdatum default = Anlagezeitpunkt, manuell überschreibbar | `Member.JoinedAt` |
| ML-004 | Verknüpfung mit OIDC-Benutzer automatisch (E-Mail-Abgleich beim ersten Login) oder manuell durch Admin | `auth_handler.go` Login-Flow |
| ML-005 | Mitglieder ohne OIDC-Verknüpfung nutzbar in Kiosk und manueller Stundenbuchung | `Member.OIDCSubject` optional |
| ML-006 | CSV-Import, Abgleich per E-Mail (Update bestehender, Anlage neuer Einträge) | `POST /api/v1/members/import` |
| ML-007 | Vorschau der Änderungen vor Import-Bestätigung | Import-Preview-Flow im Frontend |
| ML-008 | CSV-Export der Mitgliederliste | `GET /api/v1/members/export` |
| ML-009 | Änderungen im Audit-Log protokolliert | `domain.AuditEntityMember` mit den generischen `AuditAction{Create,Update,Deactivate,Import,Export}`-Konstanten (`internal/domain/audit.go`) |

`GET /api/v1/members` (Liste) ist für jeden authentifizierten Nutzer lesbar
(z. B. für Namen-Anzeige in Schichtlisten); Schreiboperationen sind
Vorstand-only (`middleware.RequireRole(domain.RoleVorstand)`).

---

## Namensanzeige und Datenschutz (NM)

Um den Datenschutz der Mitglieder zu wahren, kann der Vorstand einstellen,
wie Namen im System angezeigt werden.

| ID | Anforderung | Umsetzung |
|---|---|---|
| NM-001 | Vorstand/Admin wählen systemweit "Vollständiger Name" oder "Abgekürzter Name" | `AppSettings.NameMode`, `PUT /api/v1/settings` |
| NM-002 | Abgekürzter Modus: Nachname → erster Buchstabe + Punkt ("Maximilian M.") | `useNameFormat.ts` (`formatName`) |
| NM-003 | Rollen **ab Vorstand aufwärts** sehen immer den vollständigen Namen, unabhängig vom Systemmodus | Backend: `domain.DisplayNameFor` (`HasRole(viewerRole, RoleVorstand)`); Frontend: `useIsVorstand()` (bewusst **ohne** Veranstaltungsleiter — siehe Rollen-Tabelle in `00-general.md`) |
| NM-004 | Standardwert: "Abgekürzter Name" | `AppSettings` Default |
| NM-005 | Gilt für alle Ansichten: Schichtlisten, Kiosk-Suche, Dashboards, E-Mail-Texte an Dritte — **mit einer dokumentierten Ausnahme** (siehe unten, SC-010) | `useNameFormat()` durchgängig verwendet |
| NM-006 | Mitglieder sehen ihren eigenen Namen immer vollständig | `useNameFormat.ts`, Eigenabgleich per `member.id === user.id` |
| NM-007 | Änderungen am Namensmodus im Audit-Log | `AppSettings`-Update-Audit |

**Ausnahme (SC-010):** In der Helfer-/Registrierungsliste einer Schicht
(`RegistrationRow`, `EventDetail.tsx`), die ab **Veranstaltungsleiter**
sichtbar ist, wird der Name **immer** vollständig angezeigt
(`formatName(member, { viewerFull: true })`) — unabhängig vom Systemmodus
und unabhängig davon, ob der Betrachter Vorstand ist. Das ist kein Fehler,
sondern die explizite Anforderung aus `SC-010` ("Übersicht zeigt stets den
vollständigen Namen", siehe `03-veranstaltungen-schichten.md`) und geht
damit über NM-003 hinaus, statt es zu unterlaufen: Ein Veranstaltungsleiter
sieht abgekürzte Namen überall außer in genau dieser Helferliste.

**Wichtig für zukünftige Änderungen:** NM-003 ist strikt auf `RoleVorstand`
und höher begrenzt (Level ≥ 3), **nicht** auf "jede Rolle über Mitglied".
Ein Veranstaltungsleiter (Level 2) sieht weiterhin abgekürzte Namen. Das
wurde in dieser Migration bewusst so belassen, nachdem eine frühere
Zwischenversion des Frontend-Refactors dies fälschlich gelockert hatte
(siehe Code-Review-Historie zu `useIsVorstand`/`useIsEventManager`).
