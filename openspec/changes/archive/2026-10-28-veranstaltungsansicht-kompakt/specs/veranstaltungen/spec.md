# Spec Delta

## MODIFIED Requirements

### Requirement: Bilder und Anhänge an Veranstaltungen (V-008)
Eine Veranstaltung SHOULD mit Bildern und Anhängen versehen werden
können. Hochladen und Entfernen von Anhängen MUST auf der
Bearbeiten-Seite der Veranstaltung erfolgen. Die Veranstaltungsansicht
MUST den Bereich „Anhänge“ nur anzeigen, wenn Anhänge vorhanden sind,
und dort keine Bearbeitung anbieten.

#### Scenario: Anhang hochladen
- **WHEN** ein Veranstaltungsleiter über `POST /api/v1/events/{id}/attachments` eine Bilddatei innerhalb der Größen- und Typ-Allowlist hochlädt
- **THEN** speichert das System den Anhang und macht ihn über `GET /api/v1/events/{id}/attachments` abrufbar

#### Scenario: Abgelehnter Dateityp
- **WHEN** eine Datei außerhalb der Extension-/Content-Type-Allowlist (z. B. SVG) oder über dem 5-MB-Limit hochgeladen wird
- **THEN** lehnt das System den Upload ab

#### Scenario: Veranstaltung ohne Anhänge
- **WHEN** ein Veranstaltungsleiter eine Veranstaltung ohne Anhänge ansieht
- **THEN** zeigt die Ansicht keinen Bereich „Anhänge“

#### Scenario: Anhänge auf der Bearbeiten-Seite verwalten
- **WHEN** ein Veranstaltungsleiter die Bearbeiten-Seite einer Veranstaltung öffnet
- **THEN** kann er dort Anhänge hochladen und entfernen, während die Veranstaltungsansicht vorhandene Anhänge nur anzeigt

## ADDED Requirements

### Requirement: Kompakte Veranstaltungsansicht (V-011)
Die Veranstaltungsansicht MUST die Schichten ohne vorgeschaltete
Übersichtskacheln (Anzahl Schichten, freie Plätze, Tage) zeigen. Die
Beschreibung MUST als eigener, mit „Beschreibung“ überschriebener Bereich
abgesetzt sein; ist sie länger als die ersten Zeilen, MUST sie
standardmäßig eingeklappt sein (der Rest blendet aus) und sich auf- und
wieder zuklappen lassen. Im Kopf jeder Schicht
MUST der Titel neben Uhrzeit und Dauer stehen, wenn die Breite reicht,
und sonst darunter umbrechen.

#### Scenario: Keine Übersichtskacheln
- **WHEN** ein Mitglied eine Veranstaltung mit zwei Schichten öffnet
- **THEN** folgen auf Termin, Ort und Beschreibung direkt die Schichten, ohne Kacheln für Schichten, freie Plätze und Tage

#### Scenario: Lange Beschreibung einklappen
- **WHEN** eine Veranstaltung eine Beschreibung hat, die über die Vorschau hinausgeht
- **THEN** zeigt die Ansicht im abgesetzten Bereich „Beschreibung“ zunächst nur die ersten Zeilen, die nach unten ausblenden, mit „Mehr anzeigen“; danach den vollständigen Text mit „Weniger anzeigen“, womit sie wieder einklappt

#### Scenario: Titel neben der Uhrzeit
- **WHEN** die Veranstaltung auf einem breiten Bildschirm angezeigt wird
- **THEN** steht der Schichttitel in derselben Zeile wie Uhrzeit und Dauer
