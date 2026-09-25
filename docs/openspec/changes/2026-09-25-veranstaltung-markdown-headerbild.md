# Change: Veranstaltungsbeschreibung als Markdown (WYSIWYG-Editor) & Headerbild

**Status:** Vorgeschlagen, nicht umgesetzt.
**Betrifft:** [`03-veranstaltungen-schichten.md`](../03-veranstaltungen-schichten.md) (`V-xxx`).

## Motivation

Die Beschreibung einer Veranstaltung ist aktuell reiner Klartext in einem
einfachen `<Textarea>` (`CreateEventFlow.tsx:220`,
`EventDetail.tsx:705f.`). Veranstaltungsleiter können damit keine
Formatierung (Listen, Fett/Kursiv, Links, Absätze) hinterlegen. Zudem gibt
es kein dediziertes Headerbild für eine Veranstaltung — nur die
allgemeinen Mehrfach-Anhänge aus `V-008`
(`POST/GET/DELETE /api/v1/events/{id}/attachments`), die nicht als
Kopfbild ausgezeichnet sind und in der Detailansicht nicht prominent oben
angezeigt werden.

## Vorgeschlagene Anforderungen

| ID | Anforderung | Hinweise zur Umsetzung |
|---|---|---|
| V-009 | Die Veranstaltungsbeschreibung wird als Markdown gespeichert. Bearbeitung erfolgt standardmäßig über einen WYSIWYG-Editor; Nutzer können jederzeit auf eine Rohtext-/Markdown-Ansicht umschalten. | `domain.Event.Description` bleibt `string` (Markdown-Inhalt, kein separates HTML-Feld — vermeidet Sync-Probleme zwischen WYSIWYG-HTML und Markdown). Anzeige über Markdown-Renderer mit HTML-Sanitizing (kein raw HTML aus dem Markdown zulassen). |
| V-010 | Ein Bild kann als Veranstaltungs-Header hinterlegt werden, getrennt von den bestehenden Mehrfach-Anhängen aus `V-008`. | Neues Feld/Relation an `Event` (z. B. `HeaderImageURL` oder ein `EventAttachment` mit `IsHeader bool`), Anzeige oben in `EventDetail.tsx`. Gleiche Validierung wie `V-008`/`B-004`: Extension+Content-Type-Allowlist, SVG bewusst ausgeschlossen (Stored-XSS-Risiko), Größenlimit, Speicherung über `port.MediaStorage`. |

## Ist-Zustand (Code-Referenzen)

- `internal/domain` — `Event.Description` ist ein einfacher `string`, kein Markdown-Handling, kein Rendering-Schutz nötig, da bisher nur Klartext.
- `frontend/src/screens/admin/CreateEventFlow.tsx:220` — `<Textarea>` für die Beschreibung beim Erstellen.
- `frontend/src/screens/events/EventDetail.tsx:705f.` — `<Textarea>` für die Beschreibung im Bearbeiten-Sheet.
- `V-008` (`03-veranstaltungen-schichten.md`) — bereits vorhandene Anhangs-Infrastruktur (`port.MediaStorage`, Allowlist, 5-MB-Limit) kann für das Headerbild wiederverwendet werden.

## Design-Entscheidungen (Vorschlag)

1. Speicherformat bleibt ein einfacher String mit Markdown-Inhalt — Single Source of Truth.
2. Editor: WYSIWYG als Standardansicht (z. B. Tiptap mit Markdown-Serialisierung oder ein vergleichbarer React-Markdown-WYSIWYG-Editor — Bibliothekswahl im Rahmen der Umsetzung final festlegen), mit einem Umschalter „Markdown-Quelltext bearbeiten“ für die rohe Textform.
3. Rendering überall dort, wo die Beschreibung angezeigt wird (`EventDetail.tsx`, Mitglieder-Dashboard-Timeline, ggf. Kiosk), über einen Markdown-Renderer mit HTML-Sanitizing.
4. Headerbild nutzt dieselbe `port.MediaStorage`-Validierung wie `V-008`/`B-004` (kein SVG, Größenlimit, `/uploads/*` ohne Directory-Listing, `X-Content-Type-Options: nosniff`).
5. `api/openapi.yaml` entsprechend erweitern, danach `make generate` (Backend-Typen **und** Frontend-SDK synchron halten, siehe `CLAUDE.md` → „Code Generation“).

## Abhängigkeit / Synergie

Das Eventformular (`CreateEventFlow.tsx`, Bearbeiten-Sheet in
`EventDetail.tsx`) hat bereits mehr als 5 Felder und wird im Rahmen der
separaten Change
[„Konfiguration als eigene Seite statt Popup“](2026-09-25-konfiguration-eigene-seite-statt-popup.md)
ohnehin zu einer eigenen Seite umgebaut. Der neue Editor und der
Headerbild-Upload sollten direkt in der neuen Seitenstruktur umgesetzt
werden, nicht zusätzlich im bestehenden Sheet.

## Tasks

- [ ] `api/openapi.yaml`: `Event`-Schema um Headerbild-Feld (URL bzw. Attachment-Relation) erweitern; `description` bleibt `string` (Markdown)
- [ ] `make generate` (Backend-Typen + Frontend-SDK)
- [ ] Backend: `domain.Event`-Feld für Headerbild-Referenz + GORM AutoMigrate
- [ ] Backend: Endpoint(s) für Headerbild-Upload/-Löschung (Validierung wie `V-008`/`B-004`)
- [ ] Backend: Tests (Usecase + Handler)
- [ ] Frontend: Markdown-WYSIWYG-Bibliothek auswählen und einbinden
- [ ] Frontend: Beschreibungsfeld im Eventformular durch WYSIWYG-Editor ersetzen, inkl. Umschalter auf Markdown-Quelltext
- [ ] Frontend: Headerbild-Upload-Steuerelement im Eventformular
- [ ] Frontend: Markdown-Rendering (sanitized) überall, wo die Beschreibung angezeigt wird
- [ ] Frontend: Headerbild-Anzeige im Kopfbereich von `EventDetail.tsx`
- [ ] Frontend-Tests (Rendering/Editor-Roundtrip)
- [ ] Nach Umsetzung: Inhalt in [`03-veranstaltungen-schichten.md`](../03-veranstaltungen-schichten.md) übernehmen, diese Datei löschen (siehe [`README.md`](../README.md))
