# Spec Delta

## MODIFIED Requirements

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

## ADDED Requirements

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
