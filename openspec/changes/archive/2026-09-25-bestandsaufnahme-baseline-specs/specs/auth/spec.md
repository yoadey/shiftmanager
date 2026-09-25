# Spec Delta

## Purpose

Regelt die Anmeldung am System ausschließlich über OpenID Connect (OIDC) —
es werden keine Passwörter im System selbst gespeichert.

## ADDED Requirements

### Requirement: Konfigurierbarer OIDC-Provider (A-001)
Das System MUST die Anmeldung über einen konfigurierbaren OIDC-Provider
(z. B. Keycloak, Auth0, Azure AD) ermöglichen.

#### Scenario: Login über OIDC-Provider
- **WHEN** ein Nutzer sich anmeldet
- **THEN** leitet das System an den konfigurierten OIDC-Provider weiter und akzeptiert dessen Login-Antwort

### Requirement: Provider-Konfiguration über Umgebungsvariablen (A-002)
Das System MUST Client-ID, Client-Secret und Discovery-URL des
OIDC-Providers über Umgebungsvariablen entgegennehmen.

#### Scenario: Serverstart mit OIDC-Konfiguration
- **WHEN** der Server mit `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET` und `OIDC_REDIRECT_URL` gestartet wird
- **THEN** verwendet er diese Werte für den OIDC-Handshake

### Requirement: Internes JWT nach Login (A-003)
Das System MUST nach erfolgreichem OIDC-Login ein internes,
selbst-signiertes JWT als Session-Token ausstellen.

#### Scenario: JWT nach Callback
- **WHEN** der OIDC-Callback erfolgreich abgeschlossen wird
- **THEN** stellt das System ein JWT aus und der Client sendet es fortan als `Authorization: Bearer <token>`

### Requirement: Automatischer Token-Refresh (A-004)
Das System SHOULD das JWT während einer aktiven Session automatisch
erneuern, bevor es abläuft.

#### Scenario: Soft-Refresh vor Ablauf
- **WHEN** das JWT einer aktiven Session sich dem Ablaufzeitpunkt nähert
- **THEN** fordert das Frontend über `POST /api/v1/auth/refresh` ein neues Token an, ohne dass der Nutzer sich erneut anmelden muss

### Requirement: Mehrere gleichzeitige OIDC-Provider (A-005)
Das System MUST mehrere gleichzeitig konfigurierte OIDC-Provider
unterstützen (z. B. Vereins-SSO und Google), zwischen denen der Nutzer auf
der Login-Seite wählt.

#### Scenario: Provider-Auswahl beim Login
- **WHEN** mehrere Provider über `OIDC_PROVIDERS` konfiguriert sind und ein Nutzer `GET /api/v1/auth/providers` aufruft
- **THEN** liefert das System die Liste der konfigurierten Provider (Name/Label, keine Secrets) und der Nutzer kann über `GET /api/v1/auth/login?provider=<name>` einen davon wählen

#### Scenario: Provider-Zuordnung bleibt bis zum Callback erhalten
- **WHEN** ein Nutzer einen Provider gewählt hat und zum OIDC-Provider weitergeleitet wird
- **THEN** trägt das System die Provider-Auswahl serverseitig im `oidc_state`-Cookie (zusammen mit dem CSRF-State-Token) bis zum Callback weiter, nicht über einen client-kontrollierten Query-Parameter auf dem Callback selbst

#### Scenario: Bootstrap-Admin
- **WHEN** sich der in `BOOTSTRAP_ADMIN_EMAIL` konfigurierte Nutzer zum allerersten Mal anmeldet
- **THEN** befördert das System ihn einmalig automatisch zur Rolle Admin

#### Scenario: Test-Modus stellt Dev-Token aus
- **WHEN** `TEST_MODE=true` gesetzt ist
- **THEN** deaktiviert das System Rate-Limiting und stellt `/api/v1/dev/token` zum Ausstellen von Test-JWTs mit beliebiger Rolle bereit; dieser Endpunkt MUST NOT in Produktion aktiv sein
