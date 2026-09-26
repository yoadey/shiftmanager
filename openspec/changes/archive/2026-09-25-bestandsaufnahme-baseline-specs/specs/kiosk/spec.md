# Spec Delta

## Purpose

Ermöglicht die Schichteintragung ohne Benutzeranmeldung an einem
dedizierten Browser-Tab im Vereinsheim.

## ADDED Requirements

### Requirement: Dedizierte Kiosk-URL (K-001)
Der Kiosk-Modus MUST unter einer separaten, dedizierten URL erreichbar
sein, außerhalb der Auth-Guard.

#### Scenario: Zugriff ohne Login
- **WHEN** ein Browser `/kiosk` ohne gültiges JWT aufruft
- **THEN** zeigt das System den Kiosk-Screen ohne Login-Aufforderung

### Requirement: Konfigurierbare Mitgliedersuche (K-002)
Der Admin MUST konfigurieren können, ob im Kiosk-Modus eine
Mitgliedersuche erlaubt ist oder ausschließlich die E-Mail-Eingabe
verwendet wird.

#### Scenario: Suche deaktiviert
- **WHEN** die Mitgliedersuche deaktiviert ist
- **THEN** bietet der Kiosk ausschließlich eine E-Mail-Eingabe an (siehe K-004)

### Requirement: Suchergebnis-Format (K-003)
Ist die Suche aktiviert, MUST sie über Name oder Mitgliedsnummer erfolgen
und nur Vorname plus abgekürzten Nachnamen anzeigen.

#### Scenario: Suchergebnis anzeigen
- **WHEN** ein Kiosk-Nutzer nach einem Namen sucht
- **THEN** zeigt das System Treffer nur mit Vorname und abgekürztem Nachnamen (NM-002-Format)

### Requirement: E-Mail-Eingabe bei deaktivierter Suche (K-004)
Bei deaktivierter Suche MUST der Kiosk ausschließlich eine
E-Mail-Eingabe anbieten.

#### Scenario: Eintragung per E-Mail
- **WHEN** die Suche deaktiviert ist und ein Nutzer eine E-Mail-Adresse eingibt
- **THEN** verarbeitet das System die Eingabe über den E-Mail-Identifikationsfluss (siehe K-006/K-007)

### Requirement: Verdecktes Prüfungsergebnis (K-005)
Das Ergebnis der internen Prüfung, ob eine eingegebene E-Mail einem
Mitglied zugeordnet ist, MUST NOT dem Kiosk-Nutzer angezeigt werden.

#### Scenario: Keine Rückschlüsse auf Mitgliederliste
- **WHEN** ein Kiosk-Nutzer eine unbekannte E-Mail-Adresse eingibt
- **THEN** zeigt das System keine Information an, die erkennen lässt, ob die Adresse bekannt oder unbekannt war

### Requirement: E-Mail-Fluss für bekannte Adressen (K-006)
Ist die eingegebene E-Mail einem registrierten Mitglied zugeordnet, MUST
das System einen Bestätigungslink senden; die Schicht gilt bis zum Klick
als "Reserviert".

#### Scenario: Bekannte E-Mail
- **WHEN** eine eingegebene E-Mail einem registrierten Mitglied entspricht
- **THEN** sendet das System einen Bestätigungslink und setzt den Registrierungsstatus auf "Reserviert"

### Requirement: E-Mail-Fluss für unbekannte Adressen (K-007)
Ist die eingegebene E-Mail keinem Mitglied zugeordnet, MUST das System
den Veranstaltungsorganisator zur manuellen Klärung benachrichtigen.

#### Scenario: Unbekannte E-Mail
- **WHEN** eine eingegebene E-Mail keinem Mitglied entspricht
- **THEN** benachrichtigt das System den Organisator zur manuellen Klärung

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
