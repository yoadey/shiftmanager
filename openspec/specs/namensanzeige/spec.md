# namensanzeige Specification

## Purpose
Steuert systemweit, wie Mitgliedernamen angezeigt werden, um den
Datenschutz der Mitglieder zu wahren.

## Requirements

### Requirement: Systemweiter Namensmodus (NM-001)
Vorstand und Admin MUST systemweit zwischen "Vollständiger Name" und
"Abgekürzter Name" wählen können.

#### Scenario: Namensmodus umstellen
- **WHEN** ein Vorstandsmitglied den Namensmodus über `PUT /api/v1/settings` ändert
- **THEN** übernimmt das System den neuen Modus systemweit

### Requirement: Format des abgekürzten Namens (NM-002)
Im abgekürzten Modus MUST der Nachname auf den ersten Buchstaben plus
Punkt reduziert werden (z. B. "Maximilian M.").

#### Scenario: Anzeige im abgekürzten Modus
- **WHEN** der Namensmodus "Abgekürzter Name" aktiv ist und ein Name angezeigt wird
- **THEN** zeigt das System Vorname plus abgekürzten Nachnamen

### Requirement: Vollständiger Name ab Vorstand (NM-003)
Rollen ab Vorstand aufwärts (Vorstand, Admin) MUST immer den vollständigen
Namen sehen, unabhängig vom Systemmodus. Ein Veranstaltungsleiter zählt
NOT dazu und sieht weiterhin abgekürzte Namen außerhalb der in SC-010
beschriebenen Ausnahme.

#### Scenario: Vorstand sieht vollständigen Namen
- **WHEN** der Systemmodus "Abgekürzter Name" aktiv ist und ein Vorstandsmitglied eine Namensliste ansieht
- **THEN** zeigt das System vollständige Namen

#### Scenario: Veranstaltungsleiter sieht weiterhin abgekürzte Namen
- **WHEN** der Systemmodus "Abgekürzter Name" aktiv ist und ein Veranstaltungsleiter eine Namensliste außerhalb der Helferliste (siehe SC-010) ansieht
- **THEN** zeigt das System abgekürzte Namen

### Requirement: Standardwert des Namensmodus (NM-004)
Der Standardwert des Namensmodus MUST "Abgekürzter Name" sein.

#### Scenario: Neuinstallation
- **WHEN** die Systemeinstellungen noch nicht angepasst wurden
- **THEN** zeigt das System Namen im abgekürzten Format an

### Requirement: Geltungsbereich der Namensanzeige (NM-005)
Der Namensmodus MUST für alle Ansichten gelten — Schichtlisten,
Kiosk-Suche, Dashboards, E-Mail-Texte an Dritte — mit der in SC-010
dokumentierten Ausnahme der Helfer-/Registrierungsliste.

#### Scenario: Konsistente Anzeige über Ansichten hinweg
- **WHEN** der Systemmodus "Abgekürzter Name" aktiv ist
- **THEN** zeigt das System in Schichtlisten, Kiosk-Suche, Dashboards und Mitteilungen an Dritte durchgängig abgekürzte Namen, außer in der Helferliste einer Schicht (SC-010)

### Requirement: Eigener Name immer vollständig (NM-006)
Ein Mitglied MUST seinen eigenen Namen immer vollständig sehen,
unabhängig vom Systemmodus.

#### Scenario: Eigenes Profil ansehen
- **WHEN** ein Mitglied im abgekürzten Modus sein eigenes Profil ansieht
- **THEN** zeigt das System seinen eigenen Namen vollständig an

### Requirement: Protokollierung von Modusänderungen (NM-007)
Änderungen am Namensmodus MUST im Audit-Log protokolliert werden.

#### Scenario: Modusänderung protokolliert
- **WHEN** der Namensmodus geändert wird
- **THEN** legt das System einen Audit-Eintrag an
