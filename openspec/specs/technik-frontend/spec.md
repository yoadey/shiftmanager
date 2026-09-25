# technik-frontend Specification

## Purpose
Beschreibt die Plattformanforderungen an das React-Frontend: Stack,
Responsivität, Darstellung von Schichten und Barrierefreiheit.

## Requirements

### Requirement: React-SPA mit TypeScript (F-001)
Das Frontend MUST als Single-Page-Application mit React 18+ und
TypeScript umgesetzt sein.

#### Scenario: Build-Ausgabe
- **WHEN** das Frontend gebaut wird
- **THEN** entsteht eine SPA aus React/TypeScript-Quellcode (Vite)

### Requirement: State-Management-Trennung (F-002)
Das Frontend MUST UI-Zustand und Server-Zustand über getrennte
State-Management-Ansätze verwalten (Zustand für UI-State, TanStack Query
für Server-State).

#### Scenario: Serverdaten laden
- **WHEN** eine Ansicht Serverdaten benötigt
- **THEN** lädt und cached sie diese über TanStack Query statt über den Zustand-Store

### Requirement: Vollständig responsiv (F-003)
Das Frontend MUST vollständig responsiv und Mobile-First umgesetzt sein,
mit eigenständigen Mobile- und Desktop-Layouts.

#### Scenario: Layoutwechsel am Breakpoint
- **WHEN** die Fensterbreite die 900px-Schwelle über- bzw. unterschreitet
- **THEN** wechselt das System zwischen `DesktopShell` und `MobileShell`

### Requirement: Eigenständige Kiosk-Route (F-004)
Der Kiosk MUST als eigenständige Route mit vereinfachtem Touch-UI
umgesetzt sein.

#### Scenario: Kiosk-UI
- **WHEN** `/kiosk` aufgerufen wird
- **THEN** zeigt das System ein vereinfachtes, für Touch-Bedienung optimiertes UI ohne die reguläre Navigation

### Requirement: Interaktiver Zeitplan (F-005)
Schichten MUST in einem interaktiven, Gantt-ähnlichen Zeitplan
dargestellt werden.

#### Scenario: Zeitplan einer Veranstaltung
- **WHEN** die Detailansicht einer Veranstaltung geöffnet wird
- **THEN** zeigt das System die Schichten in einer interaktiven Zeitplan-Darstellung

### Requirement: Mehrtägige Events mit eigenem Track pro Tag (F-006)
Mehrtägige Veranstaltungen MUST tageweise untereinander mit je einem
eigenen Schicht-Track pro Tag dargestellt werden.

#### Scenario: Drei-Tage-Veranstaltung
- **WHEN** eine Veranstaltung mit drei Tagen angezeigt wird
- **THEN** zeigt das System drei untereinanderliegende Tracks, je einen pro Tag

### Requirement: Farbcodierung der Besetzung (F-007)
Die Besetzung einer Schicht MUST farblich codiert dargestellt werden:
Primärfarbe bei ausreichend Helfern, Warnfarbe bei teilweiser, Fehlerfarbe
bei kritisch unbesetzter Schicht.

#### Scenario: Kritisch unterbesetzte Schicht
- **WHEN** eine Schicht deutlich unter ihrer Mindestbesetzung liegt
- **THEN** zeigt das System sie in der Fehlerfarbe an

### Requirement: WCAG 2.1 AA (F-008)
Das Frontend MUST WCAG 2.1 AA einhalten (Kontrast, ARIA-Labels).

#### Scenario: Kontrastprüfung im Branding
- **WHEN** eine Farbkombination mit unzureichendem Kontrast konfiguriert wird
- **THEN** weist das System darauf hin (siehe `branding` B-003)

### Requirement: Ladezeiten unter 2 Sekunden (F-009)
Die Hauptansichten SHOULD in unter 2 Sekunden ladbar sein, unterstützt
durch Code-Splitting.

#### Scenario: Lazy-Loading von Screens
- **WHEN** ein Nutzer zu einem noch nicht geladenen Screen navigiert
- **THEN** lädt das System nur den Code dieses Screens nach (lazy-loaded), nicht die gesamte Anwendung neu

### Requirement: Unterstützte Browser (F-010)
Das Frontend MUST in aktuellen Versionen von Chrome, Firefox, Safari und
Edge funktionieren.

#### Scenario: Zugriff mit aktuellem Chrome
- **WHEN** ein Nutzer die Anwendung mit einer aktuellen Chrome-Version öffnet
- **THEN** funktioniert die Anwendung vollständig
