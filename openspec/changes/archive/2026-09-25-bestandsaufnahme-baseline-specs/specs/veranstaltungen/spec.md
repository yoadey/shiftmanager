# Spec Delta

## Purpose

Regelt das Anlegen, Verwalten und Kopieren von Veranstaltungen, denen
Schichten zugeordnet werden.

## ADDED Requirements

### Requirement: Veranstaltungen erstellen (V-001)
Veranstaltungsleiter, Vorstand und Admin MUST Veranstaltungen erstellen
können.

#### Scenario: Veranstaltung erstellen
- **WHEN** ein Veranstaltungsleiter `POST /api/v1/events` mit gültigen Feldern aufruft
- **THEN** legt das System die Veranstaltung an

#### Scenario: Mitglied ohne Berechtigung
- **WHEN** ein Mitglied ohne Veranstaltungsleiter-Rolle versucht, eine Veranstaltung zu erstellen
- **THEN** lehnt das System die Anfrage ab

### Requirement: Pflichtfelder einer Veranstaltung (V-002)
Eine Veranstaltung MUST Name, Beschreibung, Start-/Enddatum, Ort,
Kategorie und Sichtbarkeit enthalten.

#### Scenario: Anlage ohne Pflichtfeld
- **WHEN** beim Erstellen ein Pflichtfeld fehlt
- **THEN** lehnt das System die Anfrage ab

### Requirement: Mehrtägige Veranstaltungen (V-003)
Eine Veranstaltung MUST sich über mehrere Tage erstrecken können.

#### Scenario: Veranstaltung über drei Tage
- **WHEN** eine Veranstaltung mit einem Start- und Enddatum angelegt wird, das drei Kalendertage umfasst
- **THEN** legt das System für jeden der drei Tage einen eigenen Veranstaltungstag mit eigenen Schichten an

### Requirement: Tageweise Zeitplan-Ansicht (V-004)
Bei mehrtägigen Veranstaltungen MUST die Zeitplan-Ansicht jeden Tag als
eigene Zeile mit seinen Schichten darstellen.

#### Scenario: Timeline einer mehrtägigen Veranstaltung
- **WHEN** eine Veranstaltung mit mehreren Tagen in der Detailansicht geöffnet wird
- **THEN** zeigt das System jeden Tag als eigene Zeile mit den zugehörigen Schichten

### Requirement: Veranstaltungsstatus (V-005)
Eine Veranstaltung MUST einen Status aus Entwurf, Veröffentlicht,
Abgeschlossen oder Abgesagt haben.

#### Scenario: Statuswechsel
- **WHEN** ein Veranstaltungsleiter den Status einer Veranstaltung ändert
- **THEN** übernimmt das System den neuen Status

### Requirement: Sichtbarkeit nur für veröffentlichte Veranstaltungen (V-006)
Für Mitglieder MUST nur veröffentlichte Veranstaltungen sichtbar sein.

#### Scenario: Entwurf nicht sichtbar
- **WHEN** ein Mitglied `GET /api/v1/events` aufruft
- **THEN** liefert das System Veranstaltungen im Status Entwurf nicht mit

### Requirement: Wiederkehrende Veranstaltungen (V-007)
Eine Veranstaltung SHOULD als wöchentlich oder monatlich wiederkehrend
konfigurierbar sein.

#### Scenario: Wöchentliche Serie anlegen
- **WHEN** ein Veranstaltungsleiter über `POST /api/v1/events/{id}/recurrence` eine wöchentliche Wiederholung konfiguriert
- **THEN** legt das System vollständige Kopien der Veranstaltung inklusive Schichten für die Folgetermine an

#### Scenario: Monatliches Schrittmaß mit Tag-Überlauf
- **WHEN** eine monatliche Serie einen Zieltag erzeugt, den der Zielmonat nicht hat
- **THEN** begrenzt das System das Vorkommen auf den letzten Tag des Zielmonats statt in den Folgemonat zu rutschen

#### Scenario: Abgesagtes Event kann keine neue Serie beginnen
- **WHEN** versucht wird, eine Wiederholung für ein abgesagtes oder abgeschlossenes Event zu konfigurieren
- **THEN** lehnt das System dies ab

### Requirement: Bilder und Anhänge an Veranstaltungen (V-008)
Eine Veranstaltung SHOULD mit Bildern und Anhängen versehen werden
können.

#### Scenario: Anhang hochladen
- **WHEN** ein Veranstaltungsleiter über `POST /api/v1/events/{id}/attachments` eine Bilddatei innerhalb der Größen- und Typ-Allowlist hochlädt
- **THEN** speichert das System den Anhang und macht ihn über `GET /api/v1/events/{id}/attachments` abrufbar

#### Scenario: Abgelehnter Dateityp
- **WHEN** eine Datei außerhalb der Extension-/Content-Type-Allowlist (z. B. SVG) oder über dem 5-MB-Limit hochgeladen wird
- **THEN** lehnt das System den Upload ab

### Requirement: Veranstaltung kopieren (VC-001)
Veranstaltungsleiter, Vorstand und Admin MUST eine Veranstaltung
inklusive aller Schichten kopieren können.

#### Scenario: Veranstaltung kopieren
- **WHEN** ein Veranstaltungsleiter `POST /api/v1/events/{id}/copy` für eine bestehende Veranstaltung aufruft
- **THEN** erstellt das System eine neue Veranstaltung mit allen Schichten der Vorlage

### Requirement: Status und Name der Kopie (VC-002)
Eine kopierte Veranstaltung MUST den Status "Entwurf" und den Namen
"Kopie von [Original]" erhalten.

#### Scenario: Neue Kopie
- **WHEN** eine Veranstaltung kopiert wird
- **THEN** trägt die Kopie den Status Entwurf und den Namen "Kopie von [Originalname]"

### Requirement: Identische Schichtfelder in der Kopie (VC-003)
Alle Schichtfelder (Name, Zeiten, Helferanzahl, Qualifikationen) MUST in
der Kopie identisch zur Originalveranstaltung übertragen werden.

#### Scenario: Schichtdaten nach dem Kopieren
- **WHEN** eine Veranstaltung mit Schichten kopiert wird
- **THEN** stimmen Name, Zeiten, Min./Max.-Helferzahl und Qualifikation jeder kopierten Schicht mit dem Original überein

### Requirement: Kopier-Funktion in der Übersicht erreichbar (VC-004)
Die Kopier-Funktion MUST über einen Button in der
Veranstaltungsübersicht erreichbar sein.

#### Scenario: Kopieren aus der Übersicht
- **WHEN** ein Veranstaltungsleiter in der Veranstaltungsübersicht den Kopieren-Button einer Veranstaltung anklickt
- **THEN** startet das System den Kopiervorgang für diese Veranstaltung
