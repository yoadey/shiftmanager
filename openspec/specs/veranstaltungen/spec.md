# veranstaltungen Specification

## Purpose
Regelt das Anlegen, Verwalten und Kopieren von Veranstaltungen, denen
Schichten zugeordnet werden.

## Requirements

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
Für Mitglieder MUST nur veröffentlichte Veranstaltungen sichtbar sein;
Entwürfe MUST NOT in Liste, Detail, Zeitleiste oder Anhängen erscheinen.
Abgeschlossene und abgesagte Veranstaltungen bleiben sichtbar, damit
Mitglieder ihre eigenen Anmeldungen und deren Ausgang nachvollziehen
können. Der Kiosk MUST ausschließlich veröffentlichte Veranstaltungen
zeigen. Veranstaltungsleiter, Vorstand und Admin sehen alle
Veranstaltungen. Eine nicht sichtbare Veranstaltung MUST wie eine
fehlende mit 404 beantwortet werden.

#### Scenario: Entwurf nicht sichtbar
- **WHEN** ein Mitglied `GET /api/v1/events` aufruft
- **THEN** liefert das System Veranstaltungen im Status Entwurf nicht mit

#### Scenario: Entwurf direkt abgerufen
- **WHEN** ein Mitglied oder der Kiosk die Zeitleiste eines Entwurfs abruft
- **THEN** antwortet das System mit 404

#### Scenario: Veranstaltungsleiter sieht Entwürfe
- **WHEN** ein Veranstaltungsleiter `GET /api/v1/events?status=draft` aufruft
- **THEN** liefert das System die Entwürfe

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

### Requirement: Markdown-Beschreibung mit WYSIWYG-Editor (V-009)
Die Veranstaltungsbeschreibung MUST als Markdown gespeichert werden.
Die Bearbeitung MUST standardmäßig über einen WYSIWYG-Editor erfolgen;
Nutzer MUST jederzeit auf eine Rohtext-/Markdown-Ansicht umschalten
können. Beim Anzeigen MUST das gespeicherte Markdown gerendert werden,
ohne dass eingebettetes rohes HTML ausgeführt wird (kein Stored-XSS über
die Beschreibung).

#### Scenario: Bearbeitung im WYSIWYG-Editor
- **WHEN** ein Veranstaltungsleiter eine Veranstaltung anlegt oder bearbeitet
- **THEN** zeigt das System die Beschreibung standardmäßig in einem WYSIWYG-Editor an, dessen Formatierungen (Listen, Fett/Kursiv, Links, Absätze) als Markdown gespeichert werden

#### Scenario: Umschalten auf Markdown-Quelltext
- **WHEN** ein Veranstaltungsleiter im Editor auf "Markdown bearbeiten" umschaltet
- **THEN** zeigt das System den rohen Markdown-Quelltext zur direkten Bearbeitung an und übernimmt Änderungen daran zurück in den WYSIWYG-Editor

#### Scenario: Sicheres Rendering
- **WHEN** eine gespeicherte Beschreibung Markdown mit eingebettetem HTML (z. B. `<script>`) enthält
- **THEN** rendert das System das Markdown, ohne das eingebettete HTML auszuführen

### Requirement: Veranstaltungs-Headerbild (V-010)
Ein Bild SHOULD als Veranstaltungs-Header hinterlegt werden können,
getrennt von den bestehenden Mehrfach-Anhängen aus `V-008`. Das
Headerbild MUST derselben Validierung wie Anhänge aus `V-008`
unterliegen (Extension-/Content-Type-Allowlist, kein SVG, Größenlimit).

#### Scenario: Headerbild hochladen
- **WHEN** ein Veranstaltungsleiter ein gültiges Bild innerhalb der Allowlist und des Größenlimits als Headerbild hochlädt
- **THEN** speichert das System es getrennt von den regulären Anhängen und zeigt es oben in der Veranstaltungs-Detailansicht an

#### Scenario: Abgelehntes Headerbild
- **WHEN** eine Datei außerhalb der Allowlist (z. B. SVG) oder über dem Größenlimit als Headerbild hochgeladen wird
- **THEN** lehnt das System den Upload ab, wie bei einem regulären Anhang (`V-008`)

#### Scenario: Veranstaltung ohne Headerbild
- **WHEN** eine Veranstaltung kein Headerbild hat
- **THEN** zeigt die Detailansicht keinen Kopfbild-Bereich (kein Platzhalter-Fehler)

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
