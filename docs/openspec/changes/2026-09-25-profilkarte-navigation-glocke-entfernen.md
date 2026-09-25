# Change: Profil über Profilkarte (Sidebar) erreichbar, Glocken-Icon entfernen

**Status:** Vorgeschlagen, nicht umgesetzt.
**Betrifft:** [`00-general.md`](../00-general.md) (Seiten-Übersicht/Navigation),
[`07-branding-einstellungen.md`](../07-branding-einstellungen.md) (Abschnitt
„Profil (Mitglied)“).

## Motivation

Aktuell gibt es zwei Wege zum eigenen Profil: den regulären Tab `profil`
in der Navigationsliste (`App.tsx:48`) und ein Glocken-Icon oben rechts im
Mitglieder-Dashboard (`MemberDashboard.tsx:202-209`), das trotz
Glocken-Symbol (suggeriert Benachrichtigungen) direkt zu `routes.profil`
verlinkt und einen roten Punkt anzeigt, der an keinen echten
Benachrichtigungs-Status gekoppelt ist — irreführend. Gleichzeitig
existiert in der Desktop-Sidebar bereits eine Profilkarte
(`DesktopShell.tsx:88-102`, `dt-side-foot`/`dt-user`) mit Avatar, Namen
und Rolle unten links im Menü — sie ist aber aktuell nicht klickbar.

## Vorgeschlagene Anforderungen

| ID | Anforderung | Hinweise zur Umsetzung |
|---|---|---|
| NAV-001 | Das eigene Profil ist über die Profilkarte unten links im Navigationsmenü erreichbar (Klick navigiert zu `routes.profil`). | `DesktopShell.tsx`: `dt-user`/`dt-side-foot`-Block klickbar machen (Button/Link), aktiver Zustand wenn `location.pathname === routes.profil`. |
| NAV-002 | Das Glocken-Icon im Mitglieder-Dashboard, das zum Profil verlinkt, entfällt vollständig (Icon, Klick-Handler, Benachrichtigungs-Punkt). | `MemberDashboard.tsx:201-209` entfernen. |

## Ist-Zustand (Code-Referenzen)

- `frontend/src/screens/member/MemberDashboard.tsx:202-209` — klickbarer
  Kreis-Button mit `<Icon name="bell" />`, `onClick={() =>
  navigate(routes.profil)}`, dazu ein statischer roter
  Benachrichtigungs-Punkt ohne echten Benachrichtigungs-State dahinter.
- `frontend/src/components/layout/DesktopShell.tsx:88-102` —
  `dt-side-foot`/`dt-user`-Block (Avatar, Name, Rolle), aktuell reines
  `div` ohne `onClick`.
- `frontend/src/components/layout/MobileShell.tsx` — keine
  Sidebar/Fußzeile mit Profilkarte vorhanden; Profil ist auf Mobile
  bislang nur über den `profil`-Tab in der unteren Navigationsleiste
  erreichbar.
- `frontend/src/App.tsx:48` — `{ key: 'profil', label: 'Profil', icon:
  'user' }` als regulärer Nav-Tab (`NavTab[]`).
- `frontend/src/components/ui/Icon.tsx:17` — Icon-Definition `bell`
  (ggf. sonst ungenutzt, siehe Tasks).

## Design-Entscheidungen (Vorschlag)

1. Desktop: `dt-user`-Block in `DesktopShell.tsx` wird zu einem
   Button/Link mit `onClick={() => navigate(routes.profil)}`, visueller
   Hover-/Active-Zustand analog zu den `dt-navitem`-Einträgen.
2. Glocke: ersatzlos entfernen — kein Benachrichtigungscenter im Scope
   dieser Change.
3. Da die Profilkarte künftig der primäre Desktop-Zugang ist, entfällt
   dort Redundanz zum `profil`-Tab in der Nav-Liste (siehe offene Frage
   unten).

## Offene Fragen (vor Umsetzung zu klären)

- Bleibt der `profil`-Tab in der **mobilen** Bottom-Navigation bestehen
  (da dort keine Sidebar-Profilkarte existiert), oder wird für Mobile ein
  eigenes Profilkarten-Element eingeführt (z. B. fest verankert im
  „Mehr“-Sheet oder oberhalb der Content-Fläche)?
- Bleibt der `profil`-Tab auf **Desktop** zusätzlich in der Nav-Liste
  bestehen (doppelter Zugang zur Profilkarte), oder wird er dort
  entfernt, sobald die Profilkarte klickbar ist?

## Tasks

- [ ] `DesktopShell.tsx`: Profilkarte (`dt-user`) klickbar machen, Navigation zu `routes.profil`, aktiver Zustand
- [ ] Offene Fragen oben klären (Produktentscheidung)
- [ ] Ggf. Mobile-Äquivalent der Profilkarte ergänzen (`MobileShell.tsx`)
- [ ] Ggf. `profil`-Eintrag aus der regulären Tab-Liste entfernen (`App.tsx`), je nach Entscheidung
- [ ] `MemberDashboard.tsx`: Glocken-Button (Icon, Klick-Handler, Benachrichtigungs-Punkt) entfernen
- [ ] Prüfen, ob Icon-Eintrag `bell` noch anderswo verwendet wird; falls nicht, aus `Icon.tsx` entfernen
- [ ] Frontend-Tests/Snapshots anpassen, die auf das Glocken-Icon oder die alte Profil-Navigation referenzieren
- [ ] Nach Umsetzung: Inhalt in [`00-general.md`](../00-general.md) (Navigation) und [`07-branding-einstellungen.md`](../07-branding-einstellungen.md) (Profil-Abschnitt) übernehmen, diese Datei löschen (siehe [`README.md`](../README.md))
