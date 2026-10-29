# Design

## Context

Kiosk-Endpunkte (`/api/v1/kiosk/*`) sind öffentlich und nur
rate-limitiert; gesperrt werden kann der Kiosk systemweit (K-012). Eine
Kiosk-Anmeldung mit E-Mail wird als Gast-Registrierung (`guest_email`)
mit Bestätigungstoken gespeichert; `ConfirmByLink` setzt den Status auf
`confirmed`. Die Rolle `kiosk` existiert im Rollenmodell mit Rang 0, ein
JWT mit dieser Rolle käme aber durch `JWTAuth` und damit an alle
authentifizierten Routen ohne Rollen-Gate (z. B. `GET /events`).

## Decisions

- **Eigener Kiosk-Token statt JWT.** Ein Kiosk-Link enthält ein zufälliges
  Token (32 Byte, base64url). Gespeichert wird nur sein SHA-256-Hash in
  `kiosk_links` (Bezeichnung, Ersteller, Ablauf, Widerruf). Die
  Kiosk-Routen prüfen den Header `X-Kiosk-Token` gegen aktive, nicht
  abgelaufene Links; ohne gültiges Token 401. Der Token gilt nur für den
  Kiosk, nicht für die übrigen API-Routen, und ist einzeln widerrufbar.
  Das Klartext-Token wird nur einmal bei der Erzeugung angezeigt.
- **Token im Gerät.** Die Kiosk-Seite übernimmt `?token=` aus dem Link in
  den Local Storage des Geräts und entfernt ihn aus der Adresszeile; alle
  Kiosk-Anfragen senden ihn als Header. Ein widerrufener oder
  abgelaufener Token führt zum Hinweis „Kiosk nicht freigeschaltet“.
- **Gültigkeit pro Link.** Die Gültigkeit wird bei der Erzeugung in Tagen
  gewählt (1–3650, Vorschlag 365) – „einstellbar, wie lange“ ohne
  zusätzliche globale Einstellung.
- **Eine Registrierung, drei Wege.** Freitext-Name und E-Mail+Name
  erzeugen eine Registrierung ohne `member_id`, mit `guest_name` (dem
  eingegebenen Namen) und `self_registered = true`. Bei bekannter
  E-Mail merkt sich `pending_member_id` das Mitglied; die
  Bestätigungs-E-Mail geht an dessen Adresse. Mitgliedersuche (falls
  aktiv) bleibt wie bisher eine direkte Mitglieds-Registrierung.
- **Bestätigen = Zuordnen.** `ConfirmByLink` ordnet bei gesetztem
  `pending_member_id` zu (setzt `member_id`, leert Gastfelder und Token,
  Status bleibt). Nur reine Gast-Registrierungen behalten das bisherige
  Verhalten. Zuordnen durch Veranstaltungsleiter+ läuft über dieselbe
  Funktion (`AssignRegistration`) und informiert das Mitglied per E-Mail
  (Vorlage `kiosk-assigned`); bei Bestätigung per Link entfällt diese
  E-Mail, das Mitglied hat ja selbst bestätigt.
- **Namensähnlichkeit.** `domain.NameSimilarity` normalisiert (Kleinbuchstaben,
  ä→ae, ö→oe, ü→ue, ß→ss, gängige Akzente entfernt, Satzzeichen
  entfernt) und nimmt das Maximum aus Levenshtein-Ähnlichkeit des
  ganzen Namens, der sortierten Namensteile (vertauschte Vor-/Nachnamen)
  und einer Teil-für-Teil-Zuordnung, bei der ein Namensteil auch als
  Präfix passt („Max“ ↔ „Maximilian“). Vorschläge: aktive Mitglieder mit
  Wert ≥ 0,5, die besten fünf; das vorgemerkte Mitglied
  (`pending_member_id`) steht immer zuerst. Die Antwort enthält nur ID,
  Namen und Wert (wie ML-014).
- **Anzeige.** Die Zeitleiste zeigt für nicht zugeordnete Eintragungen
  `guest_name` nach Namensmodus (SC-011). Organisatoren erhalten
  zusätzlich `selfRegistered`, `pendingMemberId` und
  `pendingMemberName`; für Mitglieder und Kiosk werden diese entfernt.
- **Bestätigungslink.** `BASE_URL` + `/shifts/confirm/{token}` (Seite der
  App, die `/api/v1/kiosk/confirm/{token}` aufruft).

## Risks / Trade-offs

- Bestehende Kiosk-Geräte funktionieren erst nach Aufruf eines neuen
  Links wieder.
- Wer den Kiosk-Link hat, kann ihn weitergeben; deshalb Widerruf und
  Ablauf je Link sowie Audit-Einträge bei Erzeugung/Widerruf.
- Die Vorschläge laden alle aktiven Mitglieder in den Speicher; bei
  Vereinsgröße (hunderte bis wenige tausend) unkritisch.
