# kiosk Specification

## Purpose
Ermöglicht die Schichteintragung ohne Benutzeranmeldung an einem
dedizierten Browser-Tab im Vereinsheim.

## Requirements

### Requirement: Dedizierte Kiosk-URL (K-001)
Der Kiosk-Modus MUST unter einer separaten, dedizierten URL erreichbar
sein, außerhalb der Auth-Guard. Nutzbar ist er nur mit einem gültigen
Kiosk-Link (K-016); eine Benutzeranmeldung ist nicht nötig.

#### Scenario: Zugriff ohne Login
- **WHEN** ein Browser `/kiosk` über einen gültigen Kiosk-Link ohne gültiges JWT aufruft
- **THEN** zeigt das System den Kiosk-Screen ohne Login-Aufforderung

#### Scenario: Aufruf ohne Kiosk-Link
- **WHEN** ein Browser `/kiosk` ohne gültigen Kiosk-Token aufruft
- **THEN** zeigt das System den Hinweis, dass der Kiosk nicht freigeschaltet ist, und keine Veranstaltungen

### Requirement: Konfigurierbare Mitgliedersuche (K-002)
Der Admin MUST konfigurieren können, ob im Kiosk-Modus eine
Mitgliedersuche erlaubt ist oder ausschließlich die E-Mail-Eingabe
verwendet wird. Ist die Suche deaktiviert, MUST
`GET /api/v1/kiosk/members` mit 403 antworten, und eine Kiosk-Anmeldung
mit Mitglieds-ID MUST mit 403 abgelehnt werden.

#### Scenario: Suche deaktiviert
- **WHEN** die Mitgliedersuche deaktiviert ist
- **THEN** bietet der Kiosk ausschließlich eine E-Mail-Eingabe an (siehe K-004)

#### Scenario: Suchanfrage trotz deaktivierter Suche
- **WHEN** die Mitgliedersuche deaktiviert ist und `GET /api/v1/kiosk/members?q=max` aufgerufen wird
- **THEN** antwortet das System mit 403 und liefert keine Namen

#### Scenario: Anmeldung per Mitglieds-ID trotz deaktivierter Suche
- **WHEN** die Mitgliedersuche deaktiviert ist und `POST /api/v1/kiosk/shifts/{id}/register` mit einer Mitglieds-ID aufgerufen wird
- **THEN** lehnt das System die Anmeldung mit 403 ab

### Requirement: Suchergebnis-Format (K-003)
Ist die Suche aktiviert, MUST sie über Name oder Mitgliedsnummer erfolgen
und nur Vorname plus abgekürzten Nachnamen anzeigen. Die Antwort
(`GET /api/v1/kiosk/members`) MUST je Treffer ausschließlich ID und
diesen Namen enthalten, höchstens 20 Treffer liefern und MUST NOT über
E-Mail-Adressen suchen.

#### Scenario: Suchergebnis anzeigen
- **WHEN** ein Kiosk-Nutzer nach einem Namen sucht
- **THEN** zeigt das System Treffer nur mit Vorname und abgekürztem Nachnamen (NM-002-Format)

#### Scenario: Keine weiteren Mitgliederdaten
- **WHEN** `GET /api/v1/kiosk/members?q=max` aufgerufen wird
- **THEN** enthält die Antwort je Treffer nur `id` und `name`, insbesondere keine E-Mail-Adresse

#### Scenario: Keine Suche nach E-Mail-Adressen
- **WHEN** `GET /api/v1/kiosk/members?q=max@example.org` aufgerufen wird und diese Adresse einem Mitglied gehört
- **THEN** liefert das System keinen Treffer

### Requirement: E-Mail-Eingabe bei deaktivierter Suche (K-004)
Der Kiosk MUST die Eintragung per vollständigem Namen (K-014) und per
E-Mail-Adresse mit Anzeigenamen (K-006/K-007) anbieten; ist die
Mitgliedersuche aktiviert (K-002), MUST er sie zusätzlich anbieten.

#### Scenario: Eintragung per E-Mail
- **WHEN** ein Nutzer am Kiosk E-Mail-Adresse und Anzeigenamen eingibt
- **THEN** verarbeitet das System die Eingabe über den E-Mail-Identifikationsfluss (siehe K-006/K-007)

#### Scenario: Suche deaktiviert
- **WHEN** die Mitgliedersuche deaktiviert ist
- **THEN** bietet der Kiosk die Eintragung per Name und per E-Mail mit Anzeigenamen an, aber keine Suche

### Requirement: Verdecktes Prüfungsergebnis (K-005)
Das Ergebnis der internen Prüfung, ob eine eingegebene E-Mail einem
Mitglied zugeordnet ist, MUST NOT dem Kiosk-Nutzer angezeigt werden.

#### Scenario: Keine Rückschlüsse auf Mitgliederliste
- **WHEN** ein Kiosk-Nutzer eine unbekannte E-Mail-Adresse eingibt
- **THEN** zeigt das System keine Information an, die erkennen lässt, ob die Adresse bekannt oder unbekannt war

### Requirement: E-Mail-Fluss für bekannte Adressen (K-006)
Ist die eingegebene E-Mail einem aktiven Mitglied zugeordnet, MUST das
System an diese Adresse eine Bestätigungsaufforderung mit Link senden
und die Eintragung für dieses Mitglied vormerken. Bestätigt das Mitglied
über den Link oder bestätigt ein Veranstaltungsleiter, Vorstand oder
Admin die Vormerkung, MUST das System die Eintragung dem Mitglied
zuordnen, ohne ihren Anmeldestatus zu ändern. Bis dahin MUST die Schicht
den am Kiosk eingegebenen Anzeigenamen gemäß Namensmodus zeigen.

#### Scenario: Bekannte E-Mail
- **WHEN** am Kiosk die E-Mail-Adresse eines aktiven Mitglieds mit dem Anzeigenamen „Maxi“ eingegeben wird
- **THEN** sendet das System eine Bestätigungsaufforderung an diese Adresse, und die Schicht zeigt bis zur Bestätigung „Maxi“

#### Scenario: Bestätigung über den Link
- **WHEN** das Mitglied den Bestätigungslink öffnet
- **THEN** ist die Eintragung dem Mitglied zugeordnet, die Schicht zeigt seinen Namen, und der Anmeldestatus ist unverändert „angemeldet“

#### Scenario: Bestätigung durch Veranstaltungsleiter
- **WHEN** ein Veranstaltungsleiter die vorgemerkte Zuordnung in der Helferliste bestätigt
- **THEN** ist die Eintragung dem Mitglied zugeordnet

### Requirement: E-Mail-Fluss für unbekannte Adressen (K-007)
Ist die eingegebene E-Mail keinem aktiven Mitglied zugeordnet, MUST das
System die Eintragung als Gast mit dem eingegebenen Anzeigenamen
speichern, einen Bestätigungslink an die Adresse senden und die
Eintragung in der Helferliste als nicht zugeordnet kennzeichnen, damit
Veranstaltungsleiter sie zuordnen können (K-015).

#### Scenario: Unbekannte E-Mail
- **WHEN** eine eingegebene E-Mail keinem Mitglied entspricht
- **THEN** speichert das System eine Gast-Eintragung mit dem Anzeigenamen, sendet den Bestätigungslink an die Adresse und kennzeichnet die Eintragung in der Helferliste als „Zuordnung offen“

### Requirement: Mehrfacheintragung in einem Kiosk-Vorgang (K-008)
Der Kiosk MUST das Eintragen mehrerer Personen in einem
zusammenhängenden Vorgang erlauben.

#### Scenario: Mehrere Personen eintragen
- **WHEN** ein Kiosk-Nutzer mehrere Personen für dieselbe Schicht auswählt
- **THEN** trägt das System alle ausgewählten Personen ein

### Requirement: Automatisches Verfallen unbestätigter Reservierungen (K-009)
Unbestätigte Reservierungen MUST automatisch verfallen und den
freigewordenen Platz sofort wieder freigeben.

#### Scenario: Reservierung läuft ab
- **WHEN** eine Reservierung die konfigurierte Gültigkeitsdauer (K-010) überschreitet, ohne bestätigt zu werden
- **THEN** hebt das System die Reservierung auf und gibt den Platz sofort frei

### Requirement: Konfigurierbare Reservierungsdauer (K-010)
Die Gültigkeitsdauer einer Reservierung MUST konfigurierbar sein
(Standard 48 Stunden).

#### Scenario: Angepasste Reservierungsdauer
- **WHEN** der Admin die Reservierungsdauer ändert
- **THEN** wendet das System den neuen Wert auf neue Reservierungen an

### Requirement: Zeitlich begrenzte, einmalige Bestätigungslinks (K-011)
Bestätigungslinks MUST zeitlich begrenzt (entsprechend K-010) und nur
einmal verwendbar sein.

#### Scenario: Bereits verwendeter Link
- **WHEN** ein Bestätigungslink ein zweites Mal aufgerufen wird
- **THEN** lehnt das System die erneute Bestätigung ab

### Requirement: Kiosk-Sperre (K-012)
Der Admin MUST den Kiosk-Modus systemweit sperren können.

#### Scenario: Gesperrter Kiosk
- **WHEN** der Kiosk-Modus gesperrt ist und `/kiosk` aufgerufen wird
- **THEN** zeigt das System an, dass der Kiosk-Modus deaktiviert ist, statt die Eintragung zu erlauben

### Requirement: Gleiche Datenschutzlogik bei Fremdeintragung (K-013)
Meldet ein angemeldetes Mitglied eine andere Person zu einer Schicht an,
MUST dieselbe Datenschutz-Logik und derselbe E-Mail-Fluss wie im
Kiosk-Modus gelten.

#### Scenario: Mitglied trägt eine andere Person ein
- **WHEN** ein angemeldetes Mitglied eine E-Mail-Adresse für eine andere Person zur Anmeldung eingibt
- **THEN** wendet das System denselben Bekannt/Unbekannt-Fluss wie im Kiosk-Modus an

### Requirement: Eintragung per vollständigem Namen (K-014)
Am Kiosk MUST man sich mit seinem vollständigen Namen als Freitext für
eine Schicht eintragen können. Das System MUST die Eintragung ohne
Mitgliedszuordnung speichern, in der Schicht den eingegebenen Namen
gemäß Namensmodus zeigen und sie für Veranstaltungsleiter als
„Zuordnung offen“ kennzeichnen.

#### Scenario: Eintragung per Name
- **WHEN** am Kiosk „Maximilian Mustermann“ für eine Schicht eingegeben wird und der Namensmodus „Abgekürzter Name“ aktiv ist
- **THEN** zeigt die Schicht „Maximilian M.“, und Veranstaltungsleiter sehen die Eintragung als „Zuordnung offen“

### Requirement: Zuordnung mit Namensvorschlägen (K-015)
Veranstaltungsleiter, Vorstand und Admin MUST eine nicht zugeordnete
Kiosk- oder Gast-Eintragung einem aktiven Mitglied zuordnen können. Das
System MUST dazu Mitglieder nach Namensähnlichkeit vorschlagen, auch
ohne exakten Treffer, und ein für die Eintragung vorgemerktes Mitglied
(K-006) zuerst nennen. Nach der Zuordnung MUST das System das Mitglied per
E-Mail informieren. Ist das Mitglied bereits in der Schicht eingetragen,
MUST die Zuordnung abgelehnt werden.

#### Scenario: Ähnlicher Name wird vorgeschlagen
- **WHEN** ein Veranstaltungsleiter für die Eintragung „Max Muller“ Vorschläge abruft und das Mitglied „Maximilian Müller“ aktiv ist
- **THEN** enthält die Vorschlagsliste „Maximilian Müller“

#### Scenario: Zuordnen und informieren
- **WHEN** ein Veranstaltungsleiter die Eintragung einem Mitglied zuordnet
- **THEN** zeigt die Schicht das Mitglied, und das System sendet ihm eine E-Mail über die Eintragung

#### Scenario: Mitglied schon eingetragen
- **WHEN** die Eintragung einem Mitglied zugeordnet werden soll, das bereits in derselben Schicht eingetragen ist
- **THEN** lehnt das System die Zuordnung ab

### Requirement: Kiosk-Link mit Token (K-016)
Vorstand und Admin MUST Kiosk-Links erzeugen können, die einen Token mit
einer bei der Erzeugung festgelegten Gültigkeit in Tagen enthalten. Sie
MUST die aktiven Links sehen und einzeln widerrufen können. Alle
Kiosk-Endpunkte außer dem Bestätigungslink aus der E-Mail MUST einen
gültigen, nicht abgelaufenen und nicht widerrufenen Token verlangen und
ohne ihn mit 401 antworten. Das System MUST nur einen Hash des Tokens
speichern und Erzeugung sowie Widerruf protokollieren.

#### Scenario: Link erzeugen
- **WHEN** ein Vorstandsmitglied einen Kiosk-Link mit 90 Tagen Gültigkeit erzeugt
- **THEN** zeigt das System den Link mit Token einmalig an, und der Kiosk ist damit 90 Tage nutzbar

#### Scenario: Widerrufener Link
- **WHEN** ein Kiosk-Link widerrufen wurde und der Kiosk mit seinem Token Veranstaltungen abruft
- **THEN** antwortet das System mit 401

#### Scenario: Bestätigungslink ohne Token
- **WHEN** ein Mitglied den Bestätigungslink aus der E-Mail öffnet
- **THEN** verarbeitet das System die Bestätigung ohne Kiosk-Token
