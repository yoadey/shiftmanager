# Spec Delta

## Purpose

Protokolliert sicherheits- und datenschutzrelevante Aktionen im System
nachvollziehbar und manipulationsgeschützt.

## ADDED Requirements

### Requirement: Protokollierung kritischer Aktionen (T-012)
Alle sicherheits- und datenschutzrelevanten Aktionen (Mitglieder-
Änderungen, Rollen-/Einstellungsänderungen, Stundenkorrekturen,
Gebühren-Änderungen, Branding-Änderungen) MUST im Audit-Log mit
Nutzer, Aktion und Zeitpunkt protokolliert werden.

#### Scenario: Rollenänderung protokolliert
- **WHEN** die Rolle eines Mitglieds geändert wird
- **THEN** legt das System einen Audit-Eintrag mit Nutzer, Aktion und Zeitpunkt an

### Requirement: Aufbewahrung und Änderungsschutz (DS-009)
Audit-Log-Einträge MUST mindestens zwei Jahre aufbewahrt und
änderungsgeschützt (nicht nachträglich editierbar) gespeichert werden.

#### Scenario: Zugriff auf das Audit-Log
- **WHEN** ein Vorstandsmitglied `GET /api/v1/settings/audit` aufruft
- **THEN** liefert das System die protokollierten Einträge der letzten mindestens zwei Jahre

#### Scenario: Kein Veranstaltungsleiter-Zugriff
- **WHEN** ein Veranstaltungsleiter ohne Vorstandsrolle versucht, das Audit-Log aufzurufen
- **THEN** lehnt das System die Anfrage ab
