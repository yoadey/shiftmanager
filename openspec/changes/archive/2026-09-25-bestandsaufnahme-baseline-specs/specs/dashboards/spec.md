# Spec Delta

## Purpose

Stellt Mitgliedern und der Vereinsverwaltung Übersichten über Stunden,
Schichten und Veranstaltungen bereit.

## ADDED Requirements

### Requirement: Mitglieder-Dashboard mit Stundenkonto (D-001)
Das Mitglieder-Dashboard MUST bestätigte Stunden und Stunden inklusive
offener Reservierungen je gegen das Jahres-Stundenziel anzeigen.

#### Scenario: Stundenkonto anzeigen
- **WHEN** ein Mitglied sein Dashboard öffnet
- **THEN** zeigt das System bestätigte Stunden und Stunden inkl. Reservierungen jeweils im Verhältnis zum Jahresziel

### Requirement: Bevorstehende eigene Schichten (D-002)
Das Dashboard MUST die bevorstehenden eigenen Schichten des Mitglieds
anzeigen.

#### Scenario: Anstehende Schicht sichtbar
- **WHEN** ein Mitglied für eine zukünftige Schicht angemeldet ist
- **THEN** listet das Dashboard diese Schicht auf

### Requirement: Öffentlicher Kalender freier Schichten (D-003)
Das System MUST einen öffentlichen Kalender aller veröffentlichten
Veranstaltungen mit freien Schichten anbieten.

#### Scenario: Freie Schichten entdecken
- **WHEN** ein Mitglied die Entdecken-Ansicht öffnet
- **THEN** zeigt das System veröffentlichte Veranstaltungen mit noch freien Plätzen

### Requirement: Admin-Dashboard mit Systemstatistiken (D-004)
Ein Verwaltungs-Dashboard MUST systemweite Statistiken (Gesamtstunden,
offene Schichten, …) anzeigen und MUST ab der Rolle Veranstaltungsleiter
sichtbar sein.

#### Scenario: Veranstaltungsleiter sieht Verwaltungs-Dashboard
- **WHEN** ein Veranstaltungsleiter `GET /api/v1/stats` aufruft
- **THEN** liefert das System die systemweiten Statistiken, ohne dass Vorstandsrechte nötig sind

### Requirement: Filterung der Entdecken-Ansicht (D-005)
Die Entdecken-Ansicht MUST nach Datum, Veranstaltung und verfügbaren
Plätzen filterbar sein.

#### Scenario: Filter nach Datum
- **WHEN** ein Mitglied einen Datumsfilter anwendet
- **THEN** zeigt das System nur Veranstaltungen im gewählten Zeitraum

### Requirement: Kartenansicht mit Belegungsbalken (D-006)
Jede Veranstaltung MUST in der Kartenansicht mit einem Belegungsbalken
dargestellt werden.

#### Scenario: Belegungsbalken anzeigen
- **WHEN** eine Veranstaltung mit teilweise besetzten Schichten in der Kartenansicht erscheint
- **THEN** zeigt das System einen Balken, der den Belegungsgrad visualisiert
