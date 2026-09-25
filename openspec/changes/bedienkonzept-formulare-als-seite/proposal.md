# Proposal

## Why

Umfangreiche Formulare als Popup/Sheet-Overlay
(`frontend/src/components/ui/Sheet.tsx`) sind unübersichtlich: kein Platz
für Kontext, Klick außerhalb schließt versehentlich, keine Navigation
sichtbar. Ein Audit aller `<Sheet>`-Verwendungen in
`frontend/src/screens/**` zeigt drei konkrete Verstöße gegen dieses
Prinzip (siehe Impact): Veranstaltung erstellen (`CreateEventFlow.tsx`,
`Sheet variant="full"`, 7 Felder + dynamische Schichtliste), Veranstaltung
bearbeiten (`EventDetail.tsx:699`, `Sheet variant="full"`, 7 Felder +
Schichten) und das darin verschachtelte Schichtformular (`ShiftForm`,
`EventDetail.tsx:571`, 7 Felder). Alle anderen `Sheet`-Formulare im
Projekt (Mitglied bearbeiten, Serie konfigurieren, Helfer/Gast
hinzufügen, Löschantrag) haben höchstens 5 Felder oder sind reine
Bestätigungen und bleiben unverändert.

## What Changes

- Neuer globaler UI-Grundsatz: Formulare mit mehr als 5 Eingabefeldern
  werden als eigene Seite (eigene Route, volle App-Navigation sichtbar,
  kein Overlay/Backdrop) umgesetzt statt als Popup/Sheet/Modal.
- Konkrete Umsetzung: Veranstaltung erstellen/bearbeiten (inkl.
  Schichtformular) wird von einem Sheet-Overlay zu einer echten Seite
  unter den bereits bestehenden Routen `routes.eventNeu`
  (`/events/neu`) und `routes.eventBearbeiten(id)`
  (`/events/:id/bearbeiten`) umgebaut.

## Capabilities

### New Capabilities
- `bedienkonzept`: projektweite Bedienkonzept-Grundsätze (aktuell: wann
  eine eigene Seite statt eines Popups verwendet wird)

### Modified Capabilities
(keine — `veranstaltungen`/`schichten` ändern sich fachlich nicht, nur
die Präsentationsebene; siehe design.md)

## Impact

- Frontend: `frontend/src/screens/admin/CreateEventFlow.tsx` (Sheet
  entfällt, wird zu einer Seite), `frontend/src/screens/events/EventDetail.tsx`
  (Bearbeiten-Sheet ab Zeile 699 und `ShiftForm` ab Zeile 571 werden Teil
  derselben neuen Seite), `frontend/src/routes.ts` (Routen bleiben
  bestehen, nur das Rendering ändert sich).
- Kein Backend-Impact (reine Präsentationsänderung).
