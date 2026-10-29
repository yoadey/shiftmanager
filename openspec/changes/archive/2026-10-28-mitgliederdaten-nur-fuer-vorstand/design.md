# Design

## Context

Namen in Schichtlisten löst das Frontend heute über `useMembers()` (alle
Mitglieder) auf – allerdings nur in der Organisator-Ansicht; Mitglieder
sehen bislang gar keine Namen. Die Zeitleiste liefert die rohen
`domain.Registration`-Datensätze an jeden Betrachter, auch an den
anonymen Kiosk.

## Decisions

- **Rollen-Gates im Router** statt im Handler, wo das Recht nicht vom
  Datensatz abhängt (`/members`, `/members/export`, `/hours/summary`).
  Für „eigener Datensatz oder Vorstand“ (`/members/{id}`,
  `/hours/{memberId}`, `/hours/account`) prüft der Handler und antwortet
  mit 403.
- **Namen kommen aus der Zeitleiste.** Der Server ergänzt jede Anmeldung
  um `displayName`, aufgelöst über `Member.DisplayNameFor` (Namensmodus,
  eigener Name und ab Vorstand voll). Veranstaltungsleiter sehen in der
  Helferliste nach SC-010 ebenfalls volle Namen. Damit braucht kein
  Bildschirm mehr die vollständige Mitgliederliste für Namen.
- **Datensparsame Zeitleiste pro Betrachter.** Ab Veranstaltungsleiter
  bleibt alles wie bisher (Kommentare SA-009, Gast-E-Mail, gebuchte
  Stunden). Darunter und im Kiosk werden bei fremden Anmeldungen
  `guestEmail`, `comment`, `bookedHours` und `reservedUntil` geleert; eine
  reine E-Mail-Gastanmeldung erscheint als „Gast“. Die eigene Anmeldung
  bleibt vollständig.
- **Token nie serialisieren** (`json:"-"`): Er wird nur für den
  E-Mail-Link gebraucht (`SendKioskConfirmation`), kein Client liest ihn.
- **Namenssuche statt Mitgliederliste für Veranstaltungsleiter.** Eigene
  schlanke Antwort `{id, name}` (aktive, freigeschaltete Mitglieder,
  mindestens 2 Zeichen, max. 20 Treffer) statt eines rollenabhängig
  beschnittenen `Member`. Dieselbe Form nutzt die Kiosk-Suche, dort mit
  abgekürztem Nachnamen (K-003).
- **Manuelle Buchung** im Frontend nur noch für Vorstand: Die API lehnt
  Veranstaltungsleiter ohnehin ab (SA-010); die Seite bräuchte die nun
  gesperrte Mitgliederliste.

## Risks / Trade-offs

- Veranstaltungsleiter verlieren den Blick auf E-Mail-Adressen und Rollen
  der Mitglieder; für die Helferzuordnung genügt der Name.
- Die Namenssuche zeigt Veranstaltungsleitern Namen auch außerhalb einer
  Schicht – unvermeidlich, um Helfer eintragen zu können (SC-010), aber
  auf Namen und ID beschränkt.
- Clients, die `confirmationToken` aus Antworten gelesen hätten, gibt es
  nicht (geprüft: Frontend nutzt nur den Link aus der E-Mail).
