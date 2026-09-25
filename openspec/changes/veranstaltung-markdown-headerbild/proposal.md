# Proposal

## Why

Die Veranstaltungsbeschreibung ist aktuell reiner Klartext in einem
einfachen `<Textarea>` (`frontend/src/screens/admin/CreateEventFlow.tsx:220`,
`frontend/src/screens/events/EventDetail.tsx:705f.`). Veranstaltungsleiter
können damit keine Formatierung (Listen, Fett/Kursiv, Links, Absätze)
hinterlegen. Zudem gibt es kein dediziertes Headerbild für eine
Veranstaltung — nur die allgemeinen Mehrfach-Anhänge aus `V-008`
(`POST/GET/DELETE /api/v1/events/{id}/attachments`), die nicht als
Kopfbild ausgezeichnet sind und in der Detailansicht nicht prominent oben
angezeigt werden.

## What Changes

- Die Veranstaltungsbeschreibung wird als Markdown gespeichert und
  standardmäßig über einen WYSIWYG-Editor bearbeitet, mit Umschalter auf
  rohen Markdown-Quelltext.
- Ein Bild kann als eigener Veranstaltungs-Header hinterlegt werden,
  getrennt von den bestehenden Mehrfach-Anhängen (`V-008`).
- Beide neuen Requirements werden der bestehenden Capability
  `veranstaltungen` hinzugefügt (kein neues Verhalten bei bestehenden
  Requirements `V-001`–`V-008`, `VC-001`–`VC-004`).

## Capabilities

### New Capabilities
(keine)

### Modified Capabilities
- `veranstaltungen`: neue Requirements `V-009` (Markdown-Beschreibung mit
  WYSIWYG-Editor) und `V-010` (Veranstaltungs-Headerbild)

## Impact

- Backend: `internal/domain` (`Event`-Feld für Headerbild-Referenz),
  neuer/erweiterter Endpoint für Headerbild-Upload/-Löschung (Validierung
  wie `V-008`/`B-004` über `port.MediaStorage`), `api/openapi.yaml` +
  `make generate`.
- Frontend: neue Markdown-WYSIWYG-Bibliothek, Beschreibungsfeld in
  `CreateEventFlow.tsx`/der neuen Event-Seite (siehe die separate Change
  `bedienkonzept-formulare-als-seite`), Markdown-Rendering überall, wo die
  Beschreibung angezeigt wird, Headerbild-Anzeige in `EventDetail.tsx`.
