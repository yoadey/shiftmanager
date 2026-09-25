# abgeltung Specification

## Purpose
Regelt die Abgeltungsbeträge für nicht geleistete Pflichtstunden,
gestaffelt pro Fehlstunde, und deren Abrechnung.

## Requirements

### Requirement: Gestaffelte Fehlstunden-Liste (G-001)
Der Abgeltungsbetrag MUST als geordnete Liste geführt werden, deren
Index die Fehlstunde und deren Wert den Preis dieser Fehlstunde angibt.

#### Scenario: Liste konfigurieren
- **WHEN** ein Vorstandsmitglied `PUT /api/v1/settings/fee-tiers` mit einer geordneten Preisliste aufruft
- **THEN** speichert das System die Liste als aktuelle Fee-Tiers

### Requirement: Standardpreis über die Liste hinaus (G-002)
Fehlstunden über die Länge der Liste hinaus MUST mit dem Preis des
letzten Eintrags abgerechnet werden.

#### Scenario: Fehlstunde jenseits der Liste
- **WHEN** ein Mitglied mehr Fehlstunden hat als die Liste Einträge enthält
- **THEN** verwendet das System für jede darüber hinausgehende Fehlstunde den Preis des letzten Listeneintrags

### Requirement: Änderbarkeit durch Vorstand (G-003)
Die Gebührenliste MUST durch Vorstand/Admin über die
Verwaltungsoberfläche änderbar sein.

#### Scenario: Änderung durch Vorstand
- **WHEN** ein Vorstandsmitglied die Gebührenliste in der Verwaltungsoberfläche bearbeitet
- **THEN** übernimmt das System die Änderung

### Requirement: Individuelle Gebührenliste pro Mitglied (G-004)
Eine individuelle Gebührenliste pro Mitglied SHOULD möglich sein.

#### Scenario: Individuelle Liste setzen
- **WHEN** für ein Mitglied über `PUT /api/v1/settings/members/{id}/fee-tiers` eine eigene Liste hinterlegt wird
- **THEN** verwendet das System für dieses Mitglied die individuelle statt der globalen Liste

### Requirement: Jahresübersicht der Beträge (G-005)
Das System MUST eine Jahresübersicht der zu zahlenden Beträge
bereitstellen.

#### Scenario: Jahresübersicht abrufen
- **WHEN** `GET /api/v1/billing/{clubYearId}` aufgerufen wird
- **THEN** liefert das System die berechneten Abgeltungsbeträge aller betroffenen Mitglieder für dieses Vereinsjahr

### Requirement: Export als PDF und CSV (G-006)
Die Jahresübersicht MUST als PDF und als CSV exportierbar sein.

#### Scenario: PDF-Export
- **WHEN** `GET /api/v1/billing/{clubYearId}/export.pdf` aufgerufen wird
- **THEN** liefert das System die Übersicht als PDF-Datei

### Requirement: Protokollierung von Konfigurationsänderungen (G-007)
Änderungen an der Gebührenkonfiguration MUST protokolliert werden
(Nutzer, Zeitpunkt, vorheriger Wert).

#### Scenario: Änderung protokolliert
- **WHEN** die Gebührenliste geändert wird
- **THEN** legt das System einen Audit-Eintrag mit Nutzer, Zeitpunkt und vorherigem Wert an

### Requirement: Keine rückwirkende Wirkung (G-008)
Änderungen an der Gebührenkonfiguration MUST NOT rückwirkend wirken,
sondern nur auf das laufende oder zukünftige Vereinsjahr.

#### Scenario: Änderung während laufendem Vereinsjahr
- **WHEN** die Gebührenliste während eines laufenden Vereinsjahres geändert wird
- **THEN** bleiben bereits abgeschlossene Vereinsjahre bei ihrer zum jeweiligen Zeitpunkt gültigen Liste

### Requirement: Konfigurierbarer Abrechnungsmodus (G-009)
Der Abrechnungsmodus MUST konfigurierbar sein: automatische Berechnung
zum Jahresende oder manuell ausgelöste Abrechnung.

#### Scenario: Automatischer Modus
- **WHEN** der Abrechnungsmodus auf "automatisch" steht und das Vereinsjahr endet
- **THEN** berechnet das System die Abgeltung automatisch zum Jahresende

### Requirement: Vorstand-only für Billing-Endpunkte (G-010)
Alle Billing-Endpunkte und die Club-Year-CRUD-Writes MUST strikt
Vorstand/Admin vorbehalten sein.

#### Scenario: Zugriff ohne Vorstandsrolle
- **WHEN** ein Veranstaltungsleiter ohne Vorstandsrolle einen Billing-Endpunkt aufruft
- **THEN** lehnt das System die Anfrage ab
