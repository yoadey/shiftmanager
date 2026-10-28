# Design

## Context

`AuthHandler.Callback` legt unbekannte OIDC-Nutzer bereits als inaktive
Mitglieder an (`provisionMember`), die Detailansicht bietet
„Freischalten“. Die Übersicht `AdminMembers.tsx` lädt jedoch
`GET /api/v1/members` ohne `limit` und erhält damit nur die ersten 50
Einträge (Default in `MemberHandler.List`); Suche und Darstellung laufen
clientseitig auf diesen 50. „Inaktiv“ steht sowohl für neue
Registrierungen als auch für ausgetretene Mitglieder.

Rollen werden ausschließlich in der App gepflegt; der OIDC-Adapter
(`internal/adapter/oidc`) liest nur `sub`, `email` und `name` aus dem
ID-Token.

## Goals / Non-Goals

**Goals:**
- Wartende Registrierungen zuverlässig sichtbar machen, unabhängig von der
  Vereinsgröße.
- Eindeutige Unterscheidung „wartet auf Freischaltung“ vs. „ausgetreten“
  ohne Datenbank-Migration.
- Rollen optional aus dem IdP übernehmen, im Stil von Grafana, damit
  bekannte Vereinsmitglieder keine manuelle Freischaltung brauchen.

**Non-Goals:**
- Kein laufender Rollen-Sync für bereits aktive Mitglieder (die App bleibt
  führend; siehe Entscheidungen).
- Keine E-Mail-Benachrichtigung über neue Registrierungen.
- Keine Paginierung der übrigen Listen (Dashboard, manuelle Buchung,
  Veranstaltungsdetail); sie laden weiterhin alle Mitglieder, jetzt
  explizit mit `limit=0` statt implizit begrenzt auf 50.
- Kein Fix des Testmodus-Problems, dass `sqlite::memory:` pro
  DB-Verbindung eine eigene leere Datenbank erzeugt.

## Decisions

- **„Wartet“ = `is_active = false AND left_at IS NULL`** statt einer neuen
  Status-Spalte. Alle Deaktivierungspfade setzen bereits `left_at`
  (`MemberRepo.Deactivate`, `Anonymize`) — außer `PUT /members/{id}` mit
  `isActive=false`; dieser setzt jetzt ebenfalls das Austrittsdatum. So
  braucht es weder Migration noch Backfill. Altdaten, die über `PUT`
  deaktiviert wurden, erscheinen einmalig als wartend und lassen sich dort
  ablehnen. `domain.Member.IsPendingApproval()` kapselt die Regel.
- **Ablehnen = Deaktivieren.** `DeactivateMember` akzeptiert neben aktiven
  jetzt auch wartende Mitglieder. Alternative DSGVO-Löschung verworfen:
  sie ist nicht umkehrbar und ein erneuter Login würde ein frisches Konto
  anlegen.
- **Gesamtzahl im Header `X-Total-Count`**, der Body bleibt ein Array. Ein
  Envelope (`{items,total}`) hätte alle bestehenden Aufrufer (Dashboard,
  Kiosk-Suche, Event-Detail, manuelle Buchung, E2E-Tests) gebrochen.
  `MemberRepository.CountFiltered` teilt sich die WHERE-Bedingungen mit
  `List` (`applyMemberFilter`).
- **Status-Filter `approved`** (aktiv oder ausgetreten) für die
  paginierte Liste, damit wartende Registrierungen nur im eigenen
  Abschnitt oben stehen und nicht doppelt.
- **Seite und Suche in der URL** (`?seite=`, `?suche=`), konsistent mit dem
  Browser-History-Konzept der App; Seite 1 und leere Suche werden
  weggelassen. Tippen ersetzt den History-Eintrag (`replace`), ein
  Seitenwechsel erzeugt einen neuen. Suche ist um 300 ms entprellt und
  springt auf Seite 1. Seitengröße 25.
- **Hinweis-Banner** als eigene Komponente oberhalb der Screen-Routen in
  `App.tsx`, nur für `useIsVorstand()` (die Rollen, die freischalten
  dürfen — `PUT /members/{id}` ist `RequireRole(vorstand)`); auf
  `/mitglieder` ausgeblendet. Die Abfrage (`status=pending&limit=0`) läuft
  nur für diese Rollen und pollt alle 5 Minuten; Mutationen invalidieren
  `['members']` ohnehin.
- **JMESPath via `github.com/jmespath/go-jmespath`**, dieselbe Sprache wie
  Grafanas `role_attribute_path`, damit bestehende Ausdrücke übertragbar
  sind. Der Ausdruck wird beim Start kompiliert (`NewRoleMapper`); ein
  Syntaxfehler (`ErrInvalidRoleExpression`) bricht den Start ab, während
  ein Discovery-Fehler wie bisher nur den Provider deaktiviert.
- **ID-Token zuerst, dann UserInfo**, wie bei Grafana. Dafür bekommt
  `OIDCService.VerifyIDToken` die kompletten `OIDCTokens` statt nur des
  rohen ID-Tokens (Access-Token für UserInfo). UserInfo wird nur
  abgefragt, wenn ein Ausdruck konfiguriert ist und das ID-Token keine
  Rolle liefert. Die Rolle landet in `OIDCClaims.Role`.
- **Kein Erben** des Ausdrucks durch benannte Provider: Rollen aus den
  Claims eines Providers zu übernehmen ist eine Vertrauensentscheidung pro
  IdP (z. B. darf ein Google-Login keine Vereinsrolle erhalten, nur weil
  der Vereins-IdP einen Ausdruck hat).
- **Kein Rollen-Sync für aktive Mitglieder:** Die Rolle wird nur beim
  Zulassen (neu oder wartend) übernommen. Ein Überschreiben bei jedem
  Login würde in der App vergebene Rollen still zurücksetzen. Ein
  optionaler Sync (Grafana: `skip_org_role_sync`) kann später ergänzt
  werden.

## Risks / Trade-offs

- [Rollen kommen aus einem fremden System] → Nur pro Provider explizit
  aktivierbar, ausgetretene Mitglieder werden nie reaktiviert, jede
  Zulassung landet im Audit-Log (`member.oidc_role_provision`,
  `member.oidc_role_activate`).
- [`limit=0` liefert unbegrenzt viele Zeilen] → bestehendes Verhalten,
  bei Vereinsgrößen von einigen hundert Mitgliedern unkritisch.
- [Altdaten: über `PUT` deaktivierte Mitglieder ohne Austrittsdatum
  erscheinen als wartend] → einmalig über „Ablehnen“ bereinigbar; kein
  automatischer Backfill, da solche Fälle nicht von echten wartenden
  Registrierungen unterscheidbar sind.
- [SQLite `lower()` behandelt nur ASCII] → betrifft nur Tests/Testmodus;
  PostgreSQL in Produktion ist Unicode-fähig.
