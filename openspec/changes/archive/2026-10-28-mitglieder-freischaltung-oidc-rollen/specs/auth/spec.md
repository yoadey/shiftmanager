# Spec Delta

## ADDED Requirements

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
