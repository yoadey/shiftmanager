# Spec Delta

## ADDED Requirements

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
