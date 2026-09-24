# 06 — Benachrichtigungen (E-Mail)

Umgesetzt. Anforderungs-IDs: `N-001`–`N-004`, plus die E-Mail-Typen aus
Abschnitt 4 des Anforderungsdokuments.

Alle E-Mail-Vorlagen sind über die Administrationsoberfläche anpassbar
(`GET/PUT /api/v1/settings/email-templates/{name}`, `EmailTemplates.tsx`,
Vorstand-only).

## E-Mail-Typen

| Typ | Auslöser | Empfänger | Priorität |
|---|---|---|---|
| Kiosk-Bestätigung | Eintragung im Kiosk-Modus | Eingetragene E-Mail | MUSS |
| Schicht-Bestätigung | Erfolgreiche Anmeldung | Mitglied | MUSS |
| Schicht-Abmeldung | Abmeldung von Schicht | Mitglied | MUSS |
| Erinnerung (1 Woche) | 7 Tage vor Schichtbeginn | Alle Helfer | SOLL |
| Erinnerung (1 Tag) | 24 h vor Schichtbeginn | Alle Helfer | SOLL |
| Schicht abgesagt | Organisator sagt ab | Alle Helfer | MUSS |
| Jahresabrechnung | Ende Vereinsjahr | Mitglieder mit Fehlstunden | MUSS |
| Warnung Fehlstunden | X Wochen vor Jahresende | Mitglieder unter Ziel | SOLL |
| Neue Veranstaltung | Veranstaltung veröffentlicht | Alle Mitglieder (Opt-in) | KANN |
| Unterschreitung Minimum | Schicht unter Mindesthelfer | Veranstaltungsleiter | SOLL |
| Stunden bestätigt | Organisator schließt Schicht ab | Mitglied | SOLL |

Versand und Terminierung laufen als eigene Go-Goroutinen/Cron-Jobs
(`internal/infrastructure/scheduler`, siehe `T-011` in `08-technik.md`).

## Anforderungen

| ID | Anforderung | Umsetzung |
|---|---|---|
| N-001 | Opt-out für Erinnerungsmails (außer Pflicht-Mails) | `MemberPreferences` (generierter API-Typ), `PUT /api/v1/members/me/preferences` |
| N-002 | Versandzeitpunkt für Erinnerungen global konfigurierbar | `AppSettings.ReminderHourOfDay`, `AppSettings.ReminderLeadWeeks` |
| N-003 | Versand über SMTP mit TLS/STARTTLS | `internal/adapter/email` |
| N-004 | Fehlgeschlagene E-Mails protokolliert, manuell erneut versendbar | `GET /api/v1/settings/email-log`, `POST /api/v1/settings/email-log/{id}/resend`, `EmailLog.tsx` |
