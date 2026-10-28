# stundenziel Specification

## Purpose
Regelt, wie viele Helferstunden ein Mitglied pro Vereinsjahr leisten
muss und wie dieses Ziel konfiguriert wird.

## Requirements

### Requirement: Globales Standard-Stundenziel (S-001)
Das System MUST ein globales Standard-Stundenziel pro Mitglied und
Vereinsjahr führen.

#### Scenario: Neues Vereinsjahr mit Standardziel
- **WHEN** ein neues Vereinsjahr angelegt wird
- **THEN** gilt für alle Mitglieder ohne individuelles Ziel das globale Standard-Stundenziel

### Requirement: Anpassbarkeit durch Vorstand (S-002)
Das globale Stundenziel MUST durch Admin/Vorstand anpassbar sein.

#### Scenario: Ziel ändern
- **WHEN** ein Vorstandsmitglied `PUT /api/v1/hours/club-years/{id}` mit einem neuen Zielwert aufruft
- **THEN** übernimmt das System den neuen Wert für dieses Vereinsjahr

### Requirement: Individuelles Stundenziel (S-003)
Ein Mitglied SHOULD ein individuelles, vom globalen Ziel abweichendes
Stundenziel haben können.

#### Scenario: Individuelles Ziel setzen
- **WHEN** für ein Mitglied ein individuelles Ziel hinterlegt wird
- **THEN** speichert das System diesen Wert am Mitglied

### Requirement: Vorrang des individuellen Ziels (S-004)
Ist ein individuelles Ziel gesetzt, MUST es Vorrang vor dem globalen Ziel
haben.

#### Scenario: Individuelles Ziel überschreibt globales Ziel
- **WHEN** ein Mitglied ein individuelles Ziel hat und der Fortschritt berechnet wird
- **THEN** verwendet das System das individuelle statt des globalen Ziels

### Requirement: Ziel pro Vereinsjahr (S-005)
Das Stundenziel MUST pro Vereinsjahr gelten; der Zeitraum eines
Vereinsjahres MUST konfigurierbar sein (Standard 01.01.–31.12.).

#### Scenario: Abweichender Vereinsjahr-Zeitraum
- **WHEN** ein Vereinsjahr mit abweichendem Start-/Enddatum angelegt wird
- **THEN** verwendet das System diesen Zeitraum für die Zielberechnung

### Requirement: Optionaler Stundenübertrag (S-006)
Die Übertragung überzähliger Stunden ins Folgejahr SHOULD optional
konfigurierbar sein.

#### Scenario: Übertrag aktiviert
- **WHEN** ein neues Vereinsjahr mit aktiviertem Stundenübertrag angelegt wird
- **THEN** überträgt das System die überzähligen Stunden jedes Mitglieds aus dem Vorjahr, best-effort und auditiert pro Mitglied
