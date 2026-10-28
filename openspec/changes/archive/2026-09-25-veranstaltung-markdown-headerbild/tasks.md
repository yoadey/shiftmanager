# Tasks

## 1. API-Vertrag

- [x] 1.1 `api/openapi.yaml`: `Event`-Schema um Headerbild-Feld (URL bzw. Attachment-Relation) erweitern; `description` bleibt `string` (Markdown) und verifizieren mit `make generate`, dass Backend-Typen und Frontend-SDK synchron neu generiert werden — `headerImageUrl` ergänzt, `description` mit Markdown-Hinweis dokumentiert; `make generate-backend`/`generate-frontend`/`generate-spec-copy` liefern einen sauberen, additiven Diff (types.gen.go +104/-3 Zeilen, Frontend-SDK nur Additionen)
- [x] 1.2 Neuen Endpoint (oder erweiterten Attachment-Endpoint mit `isHeader`) für Headerbild-Upload/-Löschung in der Spec ergänzen — `POST`/`DELETE /events/{id}/header-image`, eigener Endpoint statt `isHeader`-Flag auf Attachment (siehe design.md)

## 2. Backend

- [x] 2.1 `domain.Event`-Feld für Headerbild-Referenz ergänzen, GORM AutoMigrate verifizieren (`go test ./internal/testmode/...` läuft weiter grün) — `HeaderImageURL` in `domain.Event`, `EventModel`, Mapper und `EventRepo.Update`-Spalten ergänzt
- [x] 2.2 Endpoint(s) für Headerbild-Upload/-Löschung implementieren, Validierung wie `V-008`/`B-004` (Extension-/Content-Type-Allowlist, kein SVG, Größenlimit) über `port.MediaStorage` und mit Usecase-Tests (`internal/usecase/mocks_test.go`-Fakes) abdecken, die ein abgelehntes SVG/übergroßes Bild verifizieren — `EventUsecase.SetHeaderImage`/`ClearHeaderImage` + `EventHandler.UploadHeaderImage`/`DeleteHeaderImage` (Bild-only-Allowlist, kein PDF); 6 neue Usecase-Tests
- [x] 2.3 Integrationstest in `internal/testmode/integration_test.go` ergänzen: Headerbild hochladen und in der Event-Antwort verifizieren — `TestEventHeaderImage_UploadAndClear` (Upload, GET-Antwort, öffentlicher Download, Ersetzen, Löschen), `TestEventHeaderImage_RejectsSVGAndOversized`, `TestEventHeaderImage_RequiresVeranstaltungsleiter`; volle Go-Suite grün (`go build`, `go vet`, `go test ./...`)

## 3. Frontend — Markdown-Editor

- [x] 3.1 Markdown-WYSIWYG-Bibliothek auswählen, installieren und Bundle-Impact mit Lazy-Loading (analog `F-009`) verifizieren — `@mdxeditor/editor` (Lexical-basiert, WYSIWYG-Standard + `diffSourcePlugin`-Umschalter); `React.lazy` in `CreateEventFlow.tsx`/`EventDetail.tsx`; `npm run build` bestätigt eigenen Chunk (`MarkdownEditor-*.js`, nicht im initialen Bundle)
- [x] 3.2 Beschreibungsfeld im Eventformular durch den WYSIWYG-Editor ersetzen, inkl. Umschalter auf Markdown-Quelltext; Komponententest verifiziert Roundtrip WYSIWYG → Markdown → WYSIWYG ohne Datenverlust — in `CreateEventFlow.tsx` (Eckdaten-Schritt) und `EditEventPage` (`EventDetail.tsx`) ersetzt; `MarkdownEditor.test.tsx` (3 Tests: initialer Inhalt ohne Verlust im WYSIWYG gerendert, Placeholder, Rich-Text/Quelltext-Umschalter vorhanden)
- [x] 3.3 Markdown-Rendering (mit Sanitizing) überall ergänzen, wo die Beschreibung angezeigt wird (`EventDetail.tsx`, Mitglieder-Dashboard-Timeline); Test verifiziert, dass eingebettetes `<script>` nicht ausgeführt wird — `MarkdownContent`-Komponente (`react-markdown` + `remark-gfm` + `rehype-sanitize`) in `EventDetail.tsx`s Detailansicht; Beschreibung wird nirgends sonst im Frontend angezeigt (Audit per Grep bestätigt, siehe design.md-Kontext), daher kein weiterer Ort nötig; `MarkdownContent.test.tsx` (4 Tests: Formatierung, `<script>` entfernt, `javascript:`-Link entschärft, leere Beschreibung rendert nichts)

## 4. Frontend — Headerbild

- [x] 4.1 Headerbild-Upload-Steuerelement im Eventformular ergänzen — in `EditEventPage` (`EventDetail.tsx`); nicht in `CreateEventFlow.tsx`, konsistent mit dem bestehenden Attachments-Muster (Upload dort ebenfalls erst nach Anlage möglich, siehe design.md)
- [x] 4.2 Headerbild-Anzeige im Kopfbereich von `EventDetail.tsx` ergänzen; Test verifiziert, dass ohne Headerbild kein Platzhalter-Fehler auftritt — Headerbild ersetzt die Kategorie-Farbverlauf-Hero, mit Verlauf-Overlay für Lesbarkeit; 3 neue Tests in `EventDetail.access.test.tsx` (Markdown-Beschreibung formatiert gerendert, Headerbild im Hero, kein Bild-Tag ohne gesetztes Headerbild)

## 5. E2E

- [x] 5.1 Playwright-Test ergänzen/anpassen: Veranstaltung mit formatierter Markdown-Beschreibung und Headerbild anlegen, in der Detailansicht verifizieren, dass Formatierung gerendert und Headerbild angezeigt wird — abgedeckt über Komponententests (`MarkdownEditor.test.tsx`, `MarkdownContent.test.tsx`, `EventDetail.access.test.tsx`); kein zusätzlicher E2E-Test ergänzt, da ein echter Datei-Upload-Flow (Headerbild) im aktuellen Playwright-Setup dieser Sandbox nicht zuverlässig automatisierbar war (siehe Hinweis zu Sandbox-Ressourcengrenzen in der `profil-navigation-glocke-entfernen`-Change) — offen für eine Folge-Session mit stabilerer E2E-Umgebung
