# Proposal

## Why

Aktuell gibt es zwei Wege zum eigenen Profil: den regulären `profil`-Tab
in der Navigationsliste (`frontend/src/App.tsx:48`) und ein Glocken-Icon
oben rechts im Mitglieder-Dashboard
(`frontend/src/screens/member/MemberDashboard.tsx:202-209`), das trotz
Glocken-Symbol (suggeriert Benachrichtigungen) direkt zu `routes.profil`
verlinkt und einen statischen roten Punkt anzeigt, der an keinen echten
Benachrichtigungs-Status gekoppelt ist — irreführend. Gleichzeitig
existiert in der Desktop-Sidebar bereits eine Profilkarte
(`frontend/src/components/layout/DesktopShell.tsx:88-102`,
`dt-side-foot`/`dt-user`) mit Avatar, Namen und Rolle unten links im
Menü — sie ist aber aktuell nicht klickbar.

## What Changes

- Die Profilkarte unten links in der Desktop-Sidebar wird klickbar und
  zum primären Desktop-Zugang zum eigenen Profil.
- Das Glocken-Icon im Mitglieder-Dashboard, das zum Profil verlinkt,
  entfällt vollständig (Icon, Klick-Handler, Benachrichtigungs-Punkt).
- Der `profil`-Tab bleibt in der **mobilen** Bottom-Navigation bestehen
  (dort gibt es keine Sidebar-Profilkarte); auf **Desktop** wird er aus
  der Sidebar-Navigationsliste entfernt, da die Profilkarte diesen
  Zugang redundant macht.

## Capabilities

### New Capabilities
(keine)

### Modified Capabilities
- `profil`: `PR-005` (Erreichbarkeit des Profils) geändert — Desktop über
  klickbare Profilkarte statt Tab + Glocke, Mobile weiterhin über den
  `profil`-Tab

## Impact

- Frontend: `frontend/src/components/layout/DesktopShell.tsx` (Profilkarte
  klickbar, Tab-Liste für die Sidebar um `profil` gefiltert),
  `frontend/src/screens/member/MemberDashboard.tsx` (Glocken-Button
  entfernt), `frontend/src/components/ui/Icon.tsx` (Eintrag `bell` bleibt
  nur, falls anderswo noch verwendet — sonst entfernen).
- Kein Backend-Impact.
