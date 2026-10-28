# systemeinstellungen Specification

## Purpose
Bündelt die Verwaltungsoberfläche für systemweite Konfiguration, die per
Anforderungsdokument Vorstand-only ist (Namensmodus, Reservierungsdauer/
Abmeldefrist, Kiosk-Sperre, Branding, Fee-Tiers, E-Mail-Vorlagen/-Protokoll,
Audit-Log-Zugriff), und regelt deren Zugriffsschutz.

## Requirements

### Requirement: Zugriffsschutz der Systemeinstellungen (SE-001)
Alle schreibenden Endpunkte der Systemeinstellungen (Namensmodus,
Reservierungsdauer/Abmeldefrist, Kiosk-Sperre, Branding, Fee-Tiers,
E-Mail-Vorlagen, Audit-Log-Einsicht) MUST Vorstand/Admin vorbehalten
sein; die zugehörigen Lese-Endpunkte (`GET /settings`,
`/settings/branding`, `/settings/fee-tiers`) MUST für jeden
authentifizierten Nutzer offen sein.

#### Scenario: Schreibzugriff ohne Vorstandsrolle
- **WHEN** ein Veranstaltungsleiter ohne Vorstandsrolle `PUT /api/v1/settings` aufruft
- **THEN** lehnt das System die Anfrage ab

#### Scenario: Lesezugriff für jedes Mitglied
- **WHEN** ein authentifiziertes Mitglied `GET /api/v1/settings` aufruft
- **THEN** liefert das System die aktuellen Einstellungen

### Requirement: Systemeinstellungen als eigene Seite (SE-002)
Die Systemeinstellungen MUST als eigene Seite/Route (`/settings`)
erreichbar sein, nicht als Popup.

#### Scenario: Aufruf der Einstellungsseite
- **WHEN** ein Vorstandsmitglied den `settings`-Tab anklickt
- **THEN** navigiert das System zur eigenständigen Einstellungsseite
