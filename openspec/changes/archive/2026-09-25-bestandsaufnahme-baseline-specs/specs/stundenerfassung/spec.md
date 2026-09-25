# Spec Delta

## Purpose

Regelt, wie geleistete Helferstunden erfasst, bestätigt und manuell
gebucht werden. Erst bestätigte Stunden zählen als geleistet.

## ADDED Requirements

### Requirement: Nur bestätigte Stunden zählen (SA-001)
Nur vom Veranstaltungsleiter oder Vorstand bestätigte Stunden MUST als
geleistet gezählt werden.

#### Scenario: Unbestätigte Anmeldung zählt nicht als geleistet
- **WHEN** ein Mitglied für eine Schicht angemeldet, aber die Schicht noch nicht abgeschlossen ist
- **THEN** zählt das System diese Stunden nicht zu den bestätigten Stunden, sondern nur zu "inkl. Reservierungen"

### Requirement: Individuelle Erfassung nach Schichtabschluss (SA-002)
Nach Schichtabschluss MUST je Person individuell erfasst werden:
geleistete Stunden (ggf. abweichend von der geplanten Zeit),
Nichterscheinen oder Teilleistung.

#### Scenario: Abweichende Stunden bestätigen
- **WHEN** ein Veranstaltungsleiter für eine Person weniger Stunden bestätigt als geplant
- **THEN** übernimmt das System die abweichende, bestätigte Stundenzahl

### Requirement: Nichterscheinen (SA-003)
Nichterschienene Personen MUST 0 Stunden und den Status "Nichterschienen"
erhalten.

#### Scenario: Person als Nichterschienen markieren
- **WHEN** ein Veranstaltungsleiter eine Person als nicht erschienen markiert
- **THEN** bucht das System 0 Stunden und setzt den Status "Nichterschienen"

### Requirement: Getrennte Anzeige bestätigter und reservierter Stunden (SA-004)
Das Dashboard MUST "Bestätigte Stunden" und "Bestätigt + Reservierungen"
getrennt anzeigen.

#### Scenario: Dashboard mit offenen Reservierungen
- **WHEN** ein Mitglied bestätigte Stunden und eine offene Schichtanmeldung hat
- **THEN** zeigt das Dashboard beide Werte nebeneinander

### Requirement: Manuelle Stundenbuchung (SA-005)
Vorstand und Admin MUST Stunden ohne Veranstaltungsbezug manuell buchen
können; Pflichtfelder sind Datum, Stunden, Mitglied und Beschreibung.

#### Scenario: Manuelle Buchung
- **WHEN** ein Vorstandsmitglied `POST /api/v1/hours/manual` mit Datum, Stunden, Mitglied und Beschreibung aufruft
- **THEN** bucht das System die Stunden für das angegebene Mitglied

### Requirement: Kennzeichnung manueller Buchungen (SA-006)
Manuelle Buchungen MUST als solche gekennzeichnet und auditiert werden.

#### Scenario: Herkunft einer Buchung
- **WHEN** eine manuelle Buchung gespeichert wird
- **THEN** kennzeichnet das System den Eintrag mit Quelle "manuell" und legt einen Audit-Eintrag an

### Requirement: Korrektur und Stornierung durch Vorstand (SA-007)
Der Vorstand MUST bestehende Einträge korrigieren oder stornieren
können; die Ursprungswerte MUST im Audit-Log erhalten bleiben.

#### Scenario: Eintrag korrigieren
- **WHEN** ein Vorstandsmitglied `PUT /api/v1/hours/{id}` mit einem geänderten Wert aufruft
- **THEN** übernimmt das System den neuen Wert und protokolliert den ursprünglichen Wert im Audit-Log

### Requirement: Optionaler Kommentar zur Eintragung (SA-008)
Bei der Schicht-Eintragung SHOULD ein optionaler Kommentar hinterlegbar
sein.

#### Scenario: Kommentar bei Anmeldung
- **WHEN** ein Mitglied sich für eine Schicht anmeldet und einen Kommentar angibt
- **THEN** speichert das System den Kommentar an der Registrierung

### Requirement: Sichtbarkeit von Kommentaren (SA-009)
Kommentare MUST für Veranstaltungsleiter und Vorstand einsehbar sein.

#### Scenario: Kommentar in der Helferliste
- **WHEN** ein Veranstaltungsleiter die Helferliste einer Schicht mit kommentierten Anmeldungen öffnet
- **THEN** zeigt das System die Kommentare an

### Requirement: Vorstand-only für Buchung/Korrektur (SA-010)
Manuelle Stundenbuchung und Korrektur/Stornierung MUST strikt Vorstand
und Admin vorbehalten sein — eine Stufe strenger als die
Schicht-/Event-Verwaltung (Veranstaltungsleiter aufwärts).

#### Scenario: Veranstaltungsleiter ohne Buchungsrecht
- **WHEN** ein Veranstaltungsleiter ohne Vorstandsrolle versucht, Stunden manuell zu buchen
- **THEN** lehnt das System die Anfrage ab
