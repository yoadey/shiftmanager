# branding Specification

## Purpose
Erlaubt es, das Erscheinungsbild des Systems (Farben, Logo, Vereinsname)
pro Verein zu konfigurieren, da ShiftManager für den Einsatz durch
verschiedene Vereine ausgelegt ist.

## Requirements

### Requirement: Primär- und Akzentfarbe konfigurierbar (B-001)
Der Admin MUST Primär- und Akzentfarbe fürs gesamte Frontend
konfigurieren können.

#### Scenario: Farben ändern
- **WHEN** ein Admin `PUT /api/v1/settings/branding` mit neuen Farbwerten aufruft
- **THEN** übernimmt das System die neuen Farben systemweit

### Requirement: Farben wirken systemweit (B-002)
Die konfigurierten Farben MUST auf Navigation, Buttons, Balken und
Badges systemweit wirken.

#### Scenario: Farbwechsel wirkt überall
- **WHEN** die Primärfarbe geändert wird
- **THEN** übernehmen Navigation, Buttons, Balken und Badges im gesamten Frontend die neue Farbe

### Requirement: Kontrastprüfung (B-003)
Das Kontrastverhältnis der gewählten Farben MUST WCAG-AA-konform bleiben
(≥ 4,5:1); bei Unterschreitung MUST eine Warnung angezeigt werden.

#### Scenario: Zu geringer Kontrast
- **WHEN** ein Admin eine Farbkombination mit einem Kontrastverhältnis unter 4,5:1 wählt
- **THEN** zeigt das System eine Warnung an

### Requirement: Logo-Upload (B-004)
Der Admin MUST ein Logo (PNG oder SVG, empfohlen ≥ 200×200 px) hochladen
können.

#### Scenario: Logo hochladen
- **WHEN** ein Admin über `POST /api/v1/settings/logo` eine gültige Bilddatei hochlädt
- **THEN** verwendet das System dieses Bild fortan als Logo

### Requirement: Logo in Navigation und Kiosk (B-005)
Das Logo MUST in der Seitennavigation (Desktop) und im Kiosk-Header
angezeigt werden.

#### Scenario: Logo im Kiosk
- **WHEN** ein Logo konfiguriert ist und der Kiosk-Screen geöffnet wird
- **THEN** zeigt das System das Logo im Kiosk-Header

### Requirement: Konfigurierbarer Vereinsname mit Fallback (B-006)
Der Vereinsname MUST als Text konfigurierbar sein; ohne Logo MUST ein
Fallback (Initialen) angezeigt werden.

#### Scenario: Fallback ohne Logo
- **WHEN** kein Logo hinterlegt ist
- **THEN** zeigt das System die Initialen des Vereinsnamens als Platzhalter

### Requirement: Kiosk-Hintergrund (B-007)
Für den Kiosk SHOULD zusätzlich ein Hintergrundbild oder eine
Hintergrundfarbe konfigurierbar sein.

#### Scenario: Kiosk-Hintergrund setzen
- **WHEN** ein Admin ein Kiosk-Hintergrundbild konfiguriert
- **THEN** zeigt der Kiosk-Screen dieses Hintergrundbild

### Requirement: Versionierung und Rücksetzbarkeit (B-008)
Die Farb-/Logo-Konfiguration MUST versioniert und auf einen früheren
Stand zurücksetzbar sein.

#### Scenario: Auf vorherige Version zurücksetzen
- **WHEN** ein Admin `POST /api/v1/settings/branding/rollback/{id}` mit einer früheren Versions-ID aufruft
- **THEN** stellt das System diesen früheren Stand wieder her
