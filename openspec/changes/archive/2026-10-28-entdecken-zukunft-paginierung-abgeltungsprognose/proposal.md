# Proposal

## Why

- Die Entdecken-Ansicht zeigt auch Veranstaltungen, die längst vorbei
  sind. Mitglieder suchen dort nach Schichten, für die sie sich noch
  anmelden können; vergangene Termine sind nur Ballast. Der Datumsfilter
  ist beim Öffnen leer.
- Die Ansicht lädt alle Veranstaltungen auf einmal (bzw. stillschweigend
  nur die ersten 50, Default von `GET /api/v1/events`) und für jede einzelne
  zusätzlich die Zeitleiste. Mit wachsendem Bestand wird das langsam und
  unübersichtlich; eine Paginierung fehlt.
- Im eigenen Dashboard sieht ein Mitglied nur, wie viele Stunden ihm zum
  Jahresziel fehlen, nicht aber, was das kostet. Wie hoch die Abgeltung
  ausfällt, wenn es keine weiteren Stunden mehr leistet, erfährt es erst
  mit der Jahresabrechnung.

## What Changes

- **Entdecken:** Bei jedem Öffnen ist „Von“ auf das heutige Datum gesetzt,
  „Bis“ ist leer. Der Nutzer kann beide Felder ändern oder leeren. Die
  Liste ist aufsteigend nach Beginn sortiert (nächste Veranstaltung zuerst)
  und paginiert (10 pro Seite, „Zurück“/„Weiter“ mit „Seite x von y“).
  Datum, Kategorie und „Freie Plätze“ werden serverseitig gefiltert, damit
  jede Seite vollständig ist.
- **`GET /api/v1/events`:** neue Query-Parameter `category`, `freeSlots`
  (nur Veranstaltungen mit mindestens einer Schicht mit freiem Platz) und
  `order` (`asc`|`desc`, Default `desc` wie bisher); Gesamtzahl der Treffer
  im Header `X-Total-Count`; `to` schließt den ganzen Tag ein (bisher fielen
  Veranstaltungen, die am Bis-Tag nach 00:00 Uhr beginnen, heraus).
- **Abgeltungsprognose:** Das Stundenkonto (`GET /api/v1/hours/me`) liefert
  zusätzlich `openFeeCents` – die Abgeltung, die nach der für das Mitglied
  geltenden Gebührenliste (G-001/G-002/G-004) für die noch fehlenden
  bestätigten Stunden fällig würde. Das Mitglieder-Dashboard zeigt diesen
  Betrag im Stundenkonto an. In der Mitglied-Detailansicht sieht der
  Vorstand denselben Wert.
- Das Stundenkonto nutzt für das Jahresziel dieselbe Auflösung wie die
  Abrechnung (individuelles Ziel des Mitglieds, S-004), damit Prognose und
  Abrechnung übereinstimmen.

## Capabilities

### New Capabilities

_Keine._

### Modified Capabilities

- `dashboards`: D-005 (Filter der Entdecken-Ansicht mit Standard ab heute),
  neue Anforderung D-007 (Paginierung); D-001 (Abgeltungsprognose im
  Stundenkonto).
- `abgeltung`: neue Anforderungen G-011 (Abgeltungsprognose für das
  eigene Konto) und G-012 (Prognose fremder Konten nur für den Vorstand).

## Impact

- Backend: `port.EventFilter`, `EventRepository.CountFiltered`,
  `EventHandler.List`, `EventUsecase.ListEvents`; `HourUsecase` erhält das
  `SettingsRepository`; gemeinsame Auflösung der Gebührenliste mit
  `BillingUsecase`.
- API: `api/openapi.yaml` (`/events`-Parameter und Header,
  `MemberHourAccount.openFeeCents`), generierte Typen.
- Frontend: `api/events.ts` (`useEventPage`), `MemberDiscover.tsx`,
  `MemberDashboard.tsx`, `MemberDetail.tsx`, `api/hours.ts`.
- Keine Datenbank-Migration.
