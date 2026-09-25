# Design

## Context

`frontend/src/components/ui/Sheet.tsx` rendert jede Variante
(`dialog`, `full`, Standard) als Overlay: `sm-overlay`-Div mit
`onClick={onClose}` (Klick außerhalb schließt) und einem eigenen
X-Button, ohne die reguläre App-Navigation. `routes.eventNeu`
(`/events/neu`) und `routes.eventBearbeiten(id)`
(`/events/:id/bearbeiten`) existieren in `frontend/src/routes.ts`
bereits als eigene Pfade — sie rendern heute nur zufällig ein Sheet
statt einer Seite; die Routen selbst müssen nicht geändert werden.

Ein Audit aller `<Sheet>`-Formulare in `frontend/src/screens/**` (siehe
proposal.md) zeigt, dass nur die Veranstaltungs-Formulare (Erstellen,
Bearbeiten, das darin verschachtelte `ShiftForm`) mehr als 5 Felder
haben; alle anderen bleiben unverändert.

## Goals / Non-Goals

**Goals:**
- Einen klaren, testbaren Schwellenwert (>5 Felder) für "eigene Seite
  statt Popup" festschreiben.
- Die Veranstaltungs-Formulare konkret auf diesen Grundsatz umstellen.

**Non-Goals:**
- Keine Änderung an Formularen mit ≤5 Feldern (Mitglied bearbeiten,
  Serie konfigurieren, Helfer/Gast hinzufügen, Löschantrag) — sie bleiben
  Popups.
- Keine Einführung einer neuen Routing-Bibliothek — die Zielrouten
  existieren bereits.

## Decisions

- **Grundsatz als eigene Capability `bedienkonzept`, nicht als Delta auf
  `veranstaltungen`.** Die fachlichen Requirements V-001–V-008 ändern
  sich nicht (dieselben Felder, dieselbe Validierung) — nur die
  Präsentationsebene. Eine eigene Capability macht den Grundsatz
  wiederverwendbar für künftige Formulare, ohne ihn an eine einzelne
  fachliche Capability zu binden.
- **`ShiftForm` wird Teil der neuen Event-Seite, kein eigenes Popup.**
  Es hat selbst 7 Felder und würde als eigenständiges Sheet denselben
  Verstoß nur verschieben statt beheben.
- **Bestehende Routen wiederverwenden, nur das Rendering ändern.** Damit
  bleiben Deep-Links und Tests, die auf `routes.eventNeu`/
  `routes.eventBearbeiten` verweisen, gültig.

## Risks / Trade-offs

- **Verlust des Overlay-Kontexts** (man sieht die Liste im Hintergrund
  nicht mehr während der Bearbeitung) → gewollter Trade-off laut
  Anforderung; mitigiert durch einen expliziten "Abbrechen"/"Zurück",
  der zur Liste zurückführt.
- **Zusätzliche Navigation nötig** (Seite statt Inline-Sheet) →
  akzeptiert, da der Schwellenwert bewusst nur Formulare mit hoher
  Feldzahl betrifft, bei denen der Kontextverlust gegenüber der
  Übersichtlichkeit überwiegt.
