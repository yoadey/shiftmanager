# Design

## Context

`DesktopShell.tsx` und `MobileShell.tsx` teilen sich dieselbe `tabs:
NavTab[]`-Liste, die `App.tsx:48` mit dem Eintrag `{ key: 'profil',
label: 'Profil', icon: 'user' }` zusammenstellt und unverändert an beide
Shells durchreicht. Die Desktop-Sidebar hat bereits einen
`dt-side-foot`/`dt-user`-Block (Avatar, Name, Rolle) unten links, aber
als reines `div` ohne `onClick`. Der Glocken-Button im
Mitglieder-Dashboard (`MemberDashboard.tsx:202-209`) navigiert per
`onClick={() => navigate(routes.profil)}` und zeigt einen statischen
roten Punkt, der an keinen echten Benachrichtigungs-Zustand gekoppelt
ist.

## Goals / Non-Goals

**Goals:**
- Desktop: Profilkarte wird der alleinige, klare Zugang zum Profil in
  der Sidebar.
- Mobile bleibt funktionsfähig, ohne einen Sidebar-Ersatz nachbauen zu
  müssen.
- Irreführendes Glocken-Symbol verschwindet ersatzlos.

**Non-Goals:**
- Kein echtes Benachrichtigungscenter (die Glocke hatte ohnehin keine
  echte Benachrichtigungsfunktion dahinter).
- Keine Änderung an `MobileShell.tsx`s Struktur über das Beibehalten des
  bestehenden `profil`-Tabs hinaus.

## Decisions

- **`profil`-Tab bleibt in der gemeinsamen `tabs`-Liste, wird aber nur in
  `DesktopShell.tsx` beim Rendern der Sidebar-Navigation herausgefiltert**
  (`tabs.filter(t => t.key !== 'profil')` für die `dt-nav`-Liste), statt
  ihn global aus `App.tsx` zu entfernen. Alternative wäre ein
  shell-spezifisches `tabs`-Array gewesen — verworfen, da `App.tsx`
  sonst zwei parallele Listen pflegen müsste, die leicht auseinanderlaufen.
  Die Route `/profil` und der `MemberProfile`-Screen selbst ändern sich
  nicht — nur, welche Shell einen Tab-Button dafür zeigt.
- **`dt-user`-Block wird ein `<button>`/klickbares Element** mit
  `onClick={() => navigate(routes.profil)}`, aktivem Zustand analog zu
  `dt-navitem` (`location.pathname === routes.profil`).
- **Glocke ersatzlos entfernen**, kein Platzhalter-Icon an ihrer Stelle.

## Risks / Trade-offs

- **Mobile und Desktop haben dauerhaft unterschiedliche Profil-Zugänge**
  (Tab vs. Sidebar-Karte) → akzeptiert, da beide Layouts ohnehin
  eigenständige Shells sind (siehe `technik-frontend` F-003) und ein
  erzwungen einheitlicher Zugang mehr Komplexität als Nutzen brächte.
