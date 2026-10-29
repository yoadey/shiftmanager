# Proposal

## Why

Mitglieder dürfen die Daten anderer Mitglieder nicht einsehen. Einzige
Ausnahme: In einer Veranstaltung sind die Namen der Helfer zu sehen, die
in der jeweiligen Schicht eingetragen sind – sonst nichts. Stundenkonten
anderer Mitglieder sind ausschließlich Vorstand und Admin vorbehalten.

Heute ist das an mehreren Stellen nicht so:

- `GET /api/v1/members`, `GET /api/v1/members/{id}` und sogar der
  CSV-Export `GET /api/v1/members/export` sind für jeden angemeldeten
  Nutzer lesbar (inklusive E-Mail, Rolle, individuellem Ziel).
- `GET /api/v1/hours/{memberId}` und `GET /api/v1/hours/account` liefern
  jedem angemeldeten Nutzer das Stundenkonto beliebiger Mitglieder;
  `GET /api/v1/hours/summary` (alle Konten) ist ab Veranstaltungsleiter
  lesbar.
- Die Zeitleiste einer Veranstaltung liefert zu jeder Anmeldung Gast-
  E-Mail, Kommentar, gebuchte Stunden und den Bestätigungstoken – an alle
  Mitglieder und über den Kiosk sogar ohne Anmeldung. Mit dem Token kann
  jeder über `GET /api/v1/kiosk/confirm/{token}` fremde Anmeldungen
  bestätigen.
- Die Kiosk-Mitgliedersuche (ohne Anmeldung) liefert vollständige
  Mitgliederdatensätze inklusive E-Mail, obwohl K-003 nur Vorname und
  abgekürzten Nachnamen vorsieht.

## What Changes

- **Mitgliederdaten:** Liste, Export und Einzelabruf eines fremden
  Mitglieds nur für Vorstand/Admin (sonst 403). Das eigene Mitglied bleibt
  abrufbar (Profil).
- **Namenssuche für Helferzuordnung:** Veranstaltungsleiter brauchen für
  SC-010 weiterhin eine Mitgliedersuche. Neu `GET /api/v1/members/lookup?q=`
  (ab Veranstaltungsleiter, mind. 2 Zeichen, höchstens 20 Treffer), die nur
  ID und Namen aktiver Mitglieder liefert.
- **Stundenkonten:** fremde Konten (`/hours/{memberId}`, `/hours/account`)
  und die Jahresübersicht (`/hours/summary`) nur für Vorstand/Admin
  (sonst 403). Die Abgeltungsprognose fremder Konten (G-012) ist damit
  automatisch Vorstand/Admin vorbehalten.
- **Zeitleiste:** Jede Anmeldung trägt den anzuzeigenden Namen
  (`displayName`, nach Namensmodus; Veranstaltungsleiter, Vorstand und das
  Mitglied selbst sehen den vollen Namen). Unterhalb der Rolle
  Veranstaltungsleiter (und im Kiosk) entfallen bei fremden Anmeldungen
  Gast-E-Mail, Kommentar und gebuchte Stunden. Der Bestätigungstoken wird
  nie mehr in einer API-Antwort ausgeliefert, nur per E-Mail.
- **Frontend:** Mitglieder sehen in jeder Schicht die Namen der
  eingetragenen Helfer. Die Helferliste der Veranstaltungsleiter nutzt die
  Namen aus der Zeitleiste statt der Mitgliederliste, „Helfer hinzufügen“
  sucht über die neue Namenssuche. Die manuelle Stundenbuchung ist nur
  noch für Vorstand/Admin erreichbar (die API erlaubt sie ohnehin nur
  ihnen, SA-010).
- **Kiosk-Suche** liefert nur ID und Vorname plus abgekürzten Nachnamen
  (K-003) und antwortet mit 403, wenn die Suche in den Einstellungen
  deaktiviert ist (K-002; bisher nur im Frontend ausgeblendet).

## Capabilities

### New Capabilities

_Keine._

### Modified Capabilities

- `mitgliederverwaltung`: ML-008, ML-010 geändert; ML-014 neu
  (Namenssuche für die Helferzuordnung).
- `stundenerfassung`: SA-011 neu (Sichtbarkeit der Stundenkonten).
- `abgeltung`: G-012 geändert (fremde Konten gar nicht mehr für
  Nicht-Vorstand).
- `schichten`: SC-011 neu (Helfernamen pro Schicht, sonst keine fremden
  Daten).
- `kiosk`: K-002 serverseitig durchgesetzt, K-003 präzisiert (keine
  weiteren Felder).
- `datenschutz-sicherheit`: DS-009 neu (Bestätigungstoken nie in
  API-Antworten).

## Impact

- Backend: `router.go`, `MemberHandler`, `HourHandler`, `EventHandler`/
  `ShiftHandler` (Zeitleiste), `KioskHandler`, neue Zeitleisten-Aufbereitung
  pro Betrachter, `domain.Registration` (Token nicht serialisiert).
- API: `api/openapi.yaml` (403-Antworten, `/members/lookup`,
  `displayName`), generierte Typen.
- Frontend: `EventDetail.tsx`, `api/members.ts`, `api/kiosk.ts`,
  `api/mappers.ts`, `App.tsx` (Route Stundenbuchung).
- Keine Datenbank-Migration.
