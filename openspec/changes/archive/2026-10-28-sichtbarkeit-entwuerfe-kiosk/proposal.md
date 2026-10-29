# Proposal

## Why

Das Code-Review des PR #15 hat drei Lücken gefunden, die die neuen
Datenschutzregeln (nur Helfernamen pro Schicht) umgehen:

- Mitglieder konnten Entwürfe über `GET /api/v1/events?status=draft` und
  `GET /api/v1/events/{id}` lesen – samt der Helfernamen. V-006 erlaubt
  Mitgliedern nur veröffentlichte Veranstaltungen. Der Kiosk lieferte die
  Zeitleiste beliebiger Veranstaltungen ohne Anmeldung.
- Die Zeitleiste enthielt bei fremden Anmeldungen weiterhin die
  Mitglieds-ID. Damit konnte jeder über den Kiosk ohne Anmeldung andere
  Mitglieder für Schichten eintragen.
- `POST /api/v1/kiosk/shifts/{id}/register` nahm eine Mitglieds-ID auch
  an, wenn die Mitgliedersuche abgeschaltet war (K-002: dann nur E-Mail).

## What Changes

- Sichtbarkeit (V-006): Veranstaltungsleiter sehen alle Veranstaltungen,
  Mitglieder alle außer Entwürfen (abgeschlossene bleiben für die
  Schicht-Historie sichtbar), der Kiosk nur veröffentlichte. Gilt für
  Liste, Detail und Zeitleiste; nicht sichtbare Veranstaltungen ergeben
  404.
- Die Zeitleiste liefert unterhalb der Rolle Veranstaltungsleiter und im
  Kiosk keine Mitglieds-ID fremder Anmeldungen (SC-011).
- Der Kiosk nimmt eine Mitglieds-ID nur bei aktivierter Mitgliedersuche
  an, sonst 403 (K-002).

## Capabilities

### New Capabilities

_Keine._

### Modified Capabilities

- `veranstaltungen`: V-006.
- `schichten`: SC-011.
- `kiosk`: K-002.

## Impact

- Backend: `EventUsecase` (`EventVisibleTo`, `GetEventTimelineFor`),
  `EventHandler.List` (`ExcludeDrafts`), `KioskHandler.Register`.
