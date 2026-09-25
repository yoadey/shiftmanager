# Tasks

## 1. Neue Seiten-Komponente

- [ ] 1.1 Neue Seiten-Komponente für Veranstaltung erstellen/bearbeiten anlegen (ersetzt den Sheet in `CreateEventFlow.tsx` und den Bearbeiten-Sheet in `EventDetail.tsx`), gerendert unter den bestehenden Routen `routes.eventNeu`/`routes.eventBearbeiten`; verifizieren, dass beide Routen weiterhin auflösen
- [ ] 1.2 `ShiftForm` in die neue Seite integrieren statt als Inline-Popup im alten Sheet; verifizieren, dass Schicht-Anlage/-Bearbeitung innerhalb der Seite funktioniert
- [ ] 1.3 Echte Seiten-Navigation umsetzen (kein Overlay-Backdrop, kein Klick-außerhalb-schließt, Browser-Zurück funktioniert); manuell verifizieren, dass der Browser-Zurück-Button zur Veranstaltungsübersicht führt

## 2. Aufräumen

- [ ] 2.1 Alte Sheet-Aufrufe für Event-Create/-Edit aus `CreateEventFlow.tsx`/`EventDetail.tsx` entfernen; `npx tsc --noEmit` bleibt grün
- [ ] 2.2 Frontend-Unit-Tests anpassen, die auf die alte Sheet-DOM-Struktur geschrieben sind (`npm test -- --run` grün)

## 3. E2E

- [ ] 3.1 Playwright-E2E-Test für Event-Erstellung/-Bearbeitung auf die neue Seitenstruktur anpassen und verifizieren, dass er ohne Overlay-Selektoren durchläuft
