# auth Specification

## Purpose
Regelt die Anmeldung am System ausschließlich über OpenID Connect (OIDC) —
es werden keine Passwörter im System selbst gespeichert.

## Requirements

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

### Requirement: Rollenzuordnung über OIDC-Claims (A-006)
Das System MUST pro OIDC-Provider einen optionalen JMESPath-Ausdruck
(`OIDC_ROLE_ATTRIBUTE_PATH` für den Einzelprovider,
`OIDC_<N>_ROLE_ATTRIBUTE_PATH` für benannte Provider aus `OIDC_PROVIDERS`;
im Helm-Chart `app.oidc.roleAttributePath`) unterstützen, der — analog zu
Grafanas `role_attribute_path` — aus den Claims eine Anwendungsrolle
ableitet. Der Ausdruck MUST zuerst gegen die Claims des ID-Tokens und,
wenn er dort keine Rolle liefert, gegen die Antwort des
UserInfo-Endpoints ausgewertet werden. Als Rolle MUST nur ein Ergebnis
gelten, das (ohne Beachtung der Groß-/Kleinschreibung) einer der Rollen
`kiosk`, `mitglied`, `veranstaltungsleiter`, `vorstand` oder `admin`
entspricht; jedes andere Ergebnis (null, unbekannter Name, kein String,
Auswertungsfehler) bedeutet „keine Rolle“.

Liefert der Ausdruck eine Rolle, MUST das System einen bisher unbekannten
Nutzer als aktives Mitglied mit dieser Rolle anlegen und ein Mitglied,
das auf Freischaltung wartet (ML-011), mit dieser Rolle aktivieren — in
beiden Fällen wird der Nutzer direkt eingeloggt und der Vorgang im
Audit-Log protokolliert. Die Rolle bereits aktiver Mitglieder MUST NOT
überschrieben werden (sie wird in der App gepflegt), und ausgetretene
Mitglieder MUST NOT über die Rollenzuordnung reaktiviert werden. Die
Bootstrap-Admin-Regel (A-005) hat Vorrang. Benannte Provider MUST NOT
den Ausdruck des Einzelproviders erben. Ein syntaktisch ungültiger
Ausdruck MUST den Serverstart mit einer Fehlermeldung abbrechen.

#### Scenario: Neuer Nutzer mit Rolle aus der Gruppe
- **WHEN** `OIDC_ROLE_ATTRIBUTE_PATH` auf `contains(groups[*], 'sm-leitung') && 'veranstaltungsleiter' || 'mitglied'` gesetzt ist und sich ein unbekannter Nutzer mit dem Claim `groups: ["sm-leitung"]` anmeldet
- **THEN** legt das System ihn als aktives Mitglied mit Rolle `veranstaltungsleiter` an und loggt ihn ein

#### Scenario: Rolle nur in UserInfo
- **WHEN** das ID-Token keinen passenden Claim enthält, die UserInfo-Antwort aber schon
- **THEN** verwendet das System die aus der UserInfo-Antwort abgeleitete Rolle

#### Scenario: Ausdruck liefert keine Rolle
- **WHEN** der Ausdruck für einen unbekannten Nutzer `null`, eine unbekannte Rolle wie `Viewer` oder einen Fehler liefert
- **THEN** legt das System ihn wie ohne Rollenzuordnung als wartende Registrierung an (ML-011)

#### Scenario: Wartende Registrierung erhält später eine Rolle
- **WHEN** sich ein Mitglied, das auf Freischaltung wartet, anmeldet und der Ausdruck inzwischen eine Rolle liefert
- **THEN** aktiviert das System das Mitglied mit dieser Rolle und loggt es ein

#### Scenario: Aktives Mitglied behält seine Rolle
- **WHEN** sich ein aktives Mitglied mit Rolle `vorstand` anmeldet und der Ausdruck `mitglied` liefert
- **THEN** bleibt die Rolle `vorstand` unverändert

#### Scenario: Ausgetretenes Mitglied bleibt ausgeschlossen
- **WHEN** sich ein ausgetretenes Mitglied anmeldet und der Ausdruck eine Rolle liefert
- **THEN** bleibt das Mitglied inaktiv und der Login wird als „inaktiv“ abgewiesen

#### Scenario: Ungültiger Ausdruck
- **WHEN** der Server mit einem syntaktisch ungültigen `OIDC_ROLE_ATTRIBUTE_PATH` startet
- **THEN** bricht der Start mit einer Fehlermeldung ab, statt den Provider stillschweigend zu deaktivieren
