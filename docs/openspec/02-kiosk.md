# 02 — Kiosk-Modus

Umgesetzt. Anforderungs-IDs: `K-001`–`K-013`.

Der Kiosk-Modus ermöglicht die Schichteintragung ohne Benutzeranmeldung. Er
ist für einen dedizierten Browser-Tab im Vereinsheim gedacht
(`frontend/src/screens/kiosk/KioskPage.tsx`, Route `/kiosk`, außerhalb der
Auth-Guard in `frontend/src/router.tsx`).

Ob Mitglieder über eine Suche auffindbar sind, steuert eine
Datenschutz-Einstellung des Administrators. Dieselbe Logik gilt für
angemeldete Mitglieder, die eine andere Person für eine Schicht eintragen
(K-013).

## E-Mail-Identifikationsfluss (bei deaktivierter Mitgliedersuche)

1. Nutzer gibt eine E-Mail-Adresse ein.
2. **Szenario A — bekannte E-Mail:** Adresse gehört einem registrierten
   Mitglied → System sendet Bestätigungslink; die Schicht gilt bis zum Klick
   als "Reserviert".
3. **Szenario B — unbekannte E-Mail:** System sendet eine Benachrichtigung
   an den Veranstaltungsorganisator zur manuellen Klärung.

Das Ergebnis der internen Prüfung (bekannt/unbekannt) wird dem
Kiosk-Nutzer selbst **nicht** angezeigt (K-005) — sonst ließe sich daraus
die Mitgliederliste erschließen.

## Anforderungen

| ID | Anforderung | Umsetzung |
|---|---|---|
| K-001 | Separate, dedizierte URL | `/kiosk` |
| K-002 | Admin konfiguriert: Mitgliedersuche erlaubt oder nur E-Mail | `AppSettings.KioskSearch` |
| K-003 | Suche über Name/Mitgliedsnummer, nur Vorname + abgekürzter Nachname sichtbar | Kiosk-Suchendpunkt, `NM-002`-Format erzwungen |
| K-004 | Bei deaktivierter Suche: ausschließlich E-Mail-Eingabe | `KioskPage.tsx` |
| K-005 | Interne Prüfung bekannt/unbekannt, Ergebnis nicht sichtbar für Nutzer | Kiosk-Handler |
| K-006 | Szenario A: Bestätigungslink, Status "Reserviert" bis Bestätigung | `domain.RegistrationStateReserved` |
| K-007 | Szenario B: Benachrichtigung an Organisator | E-Mail-Typ "Kiosk-Bestätigung" (siehe `06-benachrichtigungen.md`) |
| K-008 | Mehrere Personen in einem Kiosk-Vorgang eintragbar | `KioskPage.tsx` Mehrfach-Auswahl |
| K-009 | Unbestätigte Reservierungen verfallen automatisch, Platz wird sofort freigegeben | `internal/infrastructure/scheduler` (Cron-Job) |
| K-010 | Gültigkeitsdauer der Reservierung konfigurierbar (Standard 48 h) | `AppSettings.ReservationHours` |
| K-011 | Bestätigungslinks zeitlich begrenzt (K-010) und einmalig verwendbar | Signierter Token mit Ablauf + Verwendungs-Flag |
| K-012 | Admin kann Kiosk-Modus sperren | `AppSettings.KioskLocked`, Toggle in `AdminSettings.tsx` |
| K-013 | Gleiche Datenschutz-Logik/E-Mail-Fluss auch für angemeldete Mitglieder, die jemand anderen eintragen | Gilt sowohl im Kiosk-Flow als auch im organisatorischen "Helfer hinzufügen" (siehe `03-veranstaltungen-schichten.md`, `SC-010`) — mit dem Unterschied, dass Veranstaltungsleiter/Vorstand dort direkt Namen ohne Bestätigungslink eintragen dürfen |
