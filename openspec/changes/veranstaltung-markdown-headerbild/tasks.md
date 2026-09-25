# Tasks

## 1. API-Vertrag

- [ ] 1.1 `api/openapi.yaml`: `Event`-Schema um Headerbild-Feld (URL bzw. Attachment-Relation) erweitern; `description` bleibt `string` (Markdown) und verifizieren mit `make generate`, dass Backend-Typen und Frontend-SDK synchron neu generiert werden
- [ ] 1.2 Neuen Endpoint (oder erweiterten Attachment-Endpoint mit `isHeader`) für Headerbild-Upload/-Löschung in der Spec ergänzen

## 2. Backend

- [ ] 2.1 `domain.Event`-Feld für Headerbild-Referenz ergänzen, GORM AutoMigrate verifizieren (`go test ./internal/testmode/...` läuft weiter grün)
- [ ] 2.2 Endpoint(s) für Headerbild-Upload/-Löschung implementieren, Validierung wie `V-008`/`B-004` (Extension-/Content-Type-Allowlist, kein SVG, Größenlimit) über `port.MediaStorage` und mit Usecase-Tests (`internal/usecase/mocks_test.go`-Fakes) abdecken, die ein abgelehntes SVG/übergroßes Bild verifizieren
- [ ] 2.3 Integrationstest in `internal/testmode/integration_test.go` ergänzen: Headerbild hochladen und in der Event-Antwort verifizieren

## 3. Frontend — Markdown-Editor

- [ ] 3.1 Markdown-WYSIWYG-Bibliothek auswählen, installieren und Bundle-Impact mit Lazy-Loading (analog `F-009`) verifizieren
- [ ] 3.2 Beschreibungsfeld im Eventformular durch den WYSIWYG-Editor ersetzen, inkl. Umschalter auf Markdown-Quelltext; Komponententest verifiziert Roundtrip WYSIWYG → Markdown → WYSIWYG ohne Datenverlust
- [ ] 3.3 Markdown-Rendering (mit Sanitizing) überall ergänzen, wo die Beschreibung angezeigt wird (`EventDetail.tsx`, Mitglieder-Dashboard-Timeline); Test verifiziert, dass eingebettetes `<script>` nicht ausgeführt wird

## 4. Frontend — Headerbild

- [ ] 4.1 Headerbild-Upload-Steuerelement im Eventformular ergänzen
- [ ] 4.2 Headerbild-Anzeige im Kopfbereich von `EventDetail.tsx` ergänzen; Test verifiziert, dass ohne Headerbild kein Platzhalter-Fehler auftritt

## 5. E2E

- [ ] 5.1 Playwright-Test ergänzen/anpassen: Veranstaltung mit formatierter Markdown-Beschreibung und Headerbild anlegen, in der Detailansicht verifizieren, dass Formatierung gerendert und Headerbild angezeigt wird
