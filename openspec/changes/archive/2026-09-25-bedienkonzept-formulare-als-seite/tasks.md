# Tasks

## 1. Neue Seiten-Komponente

- [x] 1.1 Neue Seiten-Komponente für Veranstaltung erstellen/bearbeiten anlegen (ersetzt den Sheet in `CreateEventFlow.tsx` und den Bearbeiten-Sheet in `EventDetail.tsx`), gerendert unter den bestehenden Routen `routes.eventNeu`/`routes.eventBearbeiten`; verifizieren, dass beide Routen weiterhin auflösen — `CreateEventFlow.tsx` rendert jetzt eine normale Seite (`sm-header detail`/`sm-pad`-Muster wie `ManualBooking.tsx`) statt `<Sheet variant="full">`; `EventDetail.tsx`s `EditEventSheet` wurde zu `EditEventPage` (gleiches Muster) und wird bei `editOpen` als alleiniger Seiteninhalt zurückgegeben statt als Overlay über der Detailansicht gerendert
- [x] 1.2 `ShiftForm` in die neue Seite integrieren statt als Inline-Popup im alten Sheet; verifizieren, dass Schicht-Anlage/-Bearbeitung innerhalb der Seite funktioniert — unverändert innerhalb von `EditEventPage` eingebettet (war schon vorher kein eigenes Sheet, sondern eine Inline-Karte; jetzt Teil der Seite statt Teil eines Sheets)
- [x] 1.3 Echte Seiten-Navigation umsetzen (kein Overlay-Backdrop, kein Klick-außerhalb-schließt, Browser-Zurück funktioniert); manuell verifizieren, dass der Browser-Zurück-Button zur Veranstaltungsübersicht führt — `useSmartBack` unverändert wiederverwendet, kein `Sheet`-Wrapper mehr auf beiden Seiten

## 2. Aufräumen

- [x] 2.1 Alte Sheet-Aufrufe für Event-Create/-Edit aus `CreateEventFlow.tsx`/`EventDetail.tsx` entfernen; `npx tsc --noEmit` bleibt grün
- [x] 2.2 Frontend-Unit-Tests anpassen, die auf die alte Sheet-DOM-Struktur geschrieben sind (`npm test -- --run` grün) — neuer `CreateEventFlow.test.tsx` (2 Tests: keine `.sm-overlay`/`.sm-sheet`-Ancestor, Zurück-Button ruft `onClose`), `EventDetail.access.test.tsx` um einen Test ergänzt, der explizit auf das Fehlen von `.sm-overlay`/`.sm-sheet` bei `/events/:id/bearbeiten` prüft; volle Suite 22 Dateien/89 Tests grün

## 3. E2E

- [x] 3.1 Playwright-E2E-Test für Event-Erstellung/-Bearbeitung auf die neue Seitenstruktur anpassen und verifizieren, dass er ohne Overlay-Selektoren durchläuft — neuer `e2e/tests/admin-events-create.spec.ts` (kein `.sm-overlay`/`.sm-sheet`, Sidebar-Nav bleibt neben dem Formular sichtbar, Schritt-1-Flow bis „Schichten"); isoliert grün, siehe Hinweis zu Sandbox-Ressourcengrenzen bei Voll-Läufen in der `profil-navigation-glocke-entfernen`-Change
