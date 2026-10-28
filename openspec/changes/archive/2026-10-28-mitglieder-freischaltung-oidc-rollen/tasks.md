# Tasks

## 1. Domain & Persistenz

- [x] 1.1 `domain.Member.IsPendingApproval()` (inaktiv ohne Austrittsdatum) und `domain.ParseRole()` (Rollenname case-insensitiv normalisieren)
- [x] 1.2 `port.MemberFilter.Status` (`active|pending|left|approved`) und `MemberRepository.CountFiltered`; gemeinsamer Filter `applyMemberFilter` in `member_repo.go` für `List` und `CountFiltered`; Suche zusätzlich über den vollständigen Namen
- [x] 1.3 `MemberUsecase.ListMembers` liefert Seite + Gesamtzahl; `UpdateMember` setzt bei Deaktivierung das Austrittsdatum; `DeactivateMember` akzeptiert wartende Registrierungen (Ablehnen)

## 2. API

- [x] 2.1 `GET /api/v1/members`: Parameter `status` (400 bei unbekanntem Wert), Header `X-Total-Count`; Kiosk-Suche an neue Signatur angepasst
- [x] 2.2 `api/openapi.yaml` aktualisiert (`status`, `limit`, `offset`, `active`, Header, 400) und `make generate` ausgeführt

## 3. OIDC-Rollenzuordnung

- [x] 3.1 `internal/adapter/oidc/role.go`: `RoleMapper` (JMESPath via `github.com/jmespath/go-jmespath`), `ErrInvalidRoleExpression`
- [x] 3.2 `VerifyIDToken` erhält `*OIDCTokens`, wertet den Ausdruck gegen ID-Token-Claims und ersatzweise UserInfo aus, Ergebnis in `OIDCClaims.Role`
- [x] 3.3 Konfiguration `OIDC_ROLE_ATTRIBUTE_PATH` / `OIDC_<N>_ROLE_ATTRIBUTE_PATH` (ohne Erben), `main.go` bricht bei ungültigem Ausdruck ab
- [x] 3.4 `AuthHandler.Callback`: neue Nutzer mit Rolle aktiv anlegen (`member.oidc_role_provision`), wartende Mitglieder mit Rolle aktivieren (`member.oidc_role_activate`); aktive Rollen und ausgetretene Mitglieder unverändert; Bootstrap-Admin hat Vorrang

## 4. Frontend

- [x] 4.1 `api/members.ts`: `useMembers` lädt alle (`limit=0`), neue Hooks `useMemberPage` (liest `X-Total-Count`) und `usePendingMembers`, Helfer `isPendingApproval`; `Member.left` gemappt
- [x] 4.2 `AdminMembers.tsx`: Abschnitt „Warten auf Freischaltung“ mit Freischalten/Ablehnen (Rückfrage), paginierte Liste (`status=approved`, 25 pro Seite) mit Zurück/Weiter, Seite und Suche in der URL, serverseitige entprellte Suche, Badge „Ausgetreten“
- [x] 4.3 `PendingMembersBanner.tsx` in `App.tsx`: Hinweis für Vorstand/Admin auf allen Seiten außer `/mitglieder`
- [x] 4.4 `MemberDetail.tsx`: Freischalten/Ablehnen nur für wartende Konten, „Reaktivieren“ für ausgetretene; Datumsanzeige („seit“, „registriert am“, „Deaktiviert am“) für ISO-Zeitstempel korrigiert

## 5. Deployment & Doku

- [x] 5.1 Helm-Chart: `app.oidc.roleAttributePath` in `values.yaml`, `OIDC_ROLE_ATTRIBUTE_PATH` in `configmap.yaml` (nur wenn gesetzt), Abschnitt im Chart-README
- [x] 5.2 `.env.example` und `README.md` (Umgebungsvariablen, Ablauf der Freischaltung) ergänzt

## 6. Tests

- [x] 6.1 Go-Unit-Tests: `RoleMapper` (Grafana-Stil, Keycloak-Pfad, unbekannte Rolle, Fehler, leerer Ausdruck, ungültiger Ausdruck), Config (`ROLE_ATTRIBUTE_PATH`, kein Erben), Callback (neuer Nutzer mit Rolle, wartend → aktiv, ausgetreten bleibt inaktiv, aktive Rolle bleibt), Usecase (Status-Filter, Deaktivierung setzt Austrittsdatum, Ablehnen)
- [x] 6.2 Integrationstest `TestListMembers_PaginationAndStatus` (SQLite): `limit`/`offset`, `X-Total-Count`, Volltext-Suche, alle Status-Filter, 400 bei unbekanntem Status
- [x] 6.3 Frontend-Tests: `AdminMembers.test.tsx` (Abschnitt, Freischalten, Ablehnen mit Rückfrage, `status=approved`, Seitenwechsel, Seite aus URL), `PendingMembersBanner.test.tsx` (Anzahl/Singular, ausgeblendet auf `/mitglieder`, ohne Wartende, ohne Vorstandsrolle)
- [x] 6.4 `go test ./...`, `go vet ./...`, `npm test -- --run`, `npx tsc --noEmit` grün; Oberfläche im Testmodus (Desktop 1280 px, Mobil 390 px) per Playwright-Screenshot geprüft — Banner, Abschnitt, Seitenwechsel, Badge
