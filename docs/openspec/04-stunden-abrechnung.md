# 04 — Stundenziel, Stundenerfassung, Abgeltung

Umgesetzt. Anforderungs-IDs: `S-001`–`S-006`, `SA-001`–`SA-009`,
`G-001`–`G-009`.

## Stundenziel-Konfiguration

| ID | Anforderung | Umsetzung |
|---|---|---|
| S-001 | Globales Standard-Stundenziel pro Mitglied und Vereinsjahr | `ClubYear.DefaultTargetHours` |
| S-002 | Anpassbar durch Admin/Vorstand | `PUT /api/v1/hours/club-years/{id}`, `RequireRole(RoleVorstand)` |
| S-003 | Individuelles, abweichendes Stundenziel pro Mitglied | `Member.IndividualGoalHours` |
| S-004 | Individuelles Ziel hat Vorrang vor globalem Ziel | `resolveTargetHours()` (gemeinsame Hilfsfunktion, verwendet von Hour-/Billing-/Notification-/Stats-Usecase) |
| S-005 | Ziel gilt pro Vereinsjahr; Zeitraum konfigurierbar (Standard 01.01.–31.12.) | `ClubYear.StartDate/EndDate` |
| S-006 | Übertragung überzähliger Stunden ins Folgejahr optional konfigurierbar | `ClubYear.CarryOverEnabled`; Aktivierung eines neuen Jahres über `CreateActiveClubYear` (eine atomare DB-Transaktion: Insert + Deaktivierung aller anderen Jahre); Übertrag pro Mitglied best-effort und immer auditiert, auch bei frühen Fehlern |

## Stunden-Erfassung und Bestätigung

Stunden werden erst nach Bestätigung durch Veranstaltungsleiter/Vorstand als
geleistet gewertet. Das Mitglieder-Dashboard unterscheidet zwei Werte:
**"Bestätigte Stunden"** (abgenommen und gebucht) und **"Inkl.
Reservierungen"** (bestätigt plus offene Anmeldungen) — beide nebeneinander.

| ID | Anforderung | Umsetzung |
|---|---|---|
| SA-001 | Nur vom Veranstaltungsleiter/Vorstand bestätigte Stunden zählen | `POST /api/v1/hours/confirm` |
| SA-002 | Individuelle Erfassung pro Person nach Schichtabschluss: geleistete Stunden (ggf. abweichend), Nichterscheinen, Teilleistung | `ConfirmShiftHours`-Usecase |
| SA-003 | Nichterschienene erhalten 0 Stunden, Status "Nichterschienen" | `domain.RegistrationStateNoShow` |
| SA-004 | Dashboard zeigt getrennt "Bestätigte Stunden" und "Bestätigt + Reservierungen" | `MemberDashboard.tsx` |
| SA-005 | Vorstand/Admin buchen manuell Stunden ohne Veranstaltungsbezug (Pflicht: Datum, Stunden, Mitglied, Beschreibung) | `POST /api/v1/hours/manual`, `RequireRole(RoleVorstand)`, `ManualBooking.tsx` |
| SA-006 | Manuelle Buchungen gekennzeichnet und auditiert | `HourEntry.Source = manual` |
| SA-007 | Vorstand korrigiert/storniert Einträge, Ursprungswerte bleiben im Audit-Log | `PUT/DELETE /api/v1/hours/{id}`, `RequireRole(RoleVorstand)` |
| SA-008 | Optionaler Kommentar zur Schicht-Eintragung | `Registration.Comment` |
| SA-009 | Kommentare einsehbar für Veranstaltungsleiter/Vorstand | `RegistrationRow` in `EventDetail.tsx` |

## Abgeltungsbeträge

Für nicht geleistete Stunden wird ein Abgeltungsbetrag erhoben, gestaffelt
pro Fehlstunde (erste Fehlstunde kann anders kosten als die zweite).

| ID | Anforderung | Umsetzung |
|---|---|---|
| G-001 | Geordnete Liste: Index = Fehlstunde, Wert = Preis | `FeeTier[]`, `PUT /api/v1/settings/fee-tiers` |
| G-002 | Über die Liste hinausgehende Fehlstunden: letzter Eintrag als Standardpreis | Billing-Usecase |
| G-003 | Änderbar durch Vorstand/Admin über die Verwaltungsoberfläche | `RequireRole(RoleVorstand)` |
| G-004 | Individuelle Gebührenliste pro Mitglied möglich | `GET/PUT /api/v1/settings/members/{id}/fee-tiers` |
| G-005 | Jahresübersicht der zu zahlenden Beträge | `GET /api/v1/billing/{clubYearId}` |
| G-006 | Export als PDF und CSV | `GET /api/v1/billing/{clubYearId}/export.{csv,pdf}` |
| G-007 | Änderungen an der Gebührenkonfiguration protokolliert (Nutzer, Zeitpunkt, vorheriger Wert) | Audit-Log |
| G-008 | Änderungen wirken nur auf laufendes/zukünftiges Vereinsjahr, nicht rückwirkend | Fee-Tier-Zuordnung pro `ClubYear` |
| G-009 | Abrechnungsmodus konfigurierbar: automatisch zum Jahresende oder manuell ausgelöst | `AppSettings.BillingMode` |

Alle Billing-Endpunkte (`/api/v1/billing/*`) und die Club-Year-CRUD-Writes
(`POST/PUT/DELETE /hours/club-years`) sowie die manuelle Stundenbuchung
(`/hours/manual`, `/hours/{id}` Korrektur) sind strikt **Vorstand+**
(`RequireRole(RoleVorstand)`) — eine Stufe strenger als die
Schicht-/Event-Verwaltung (Veranstaltungsleiter+). Im Frontend gated durch
`useIsVorstand()`, nicht `useIsEventManager()` (z. B. der Button "Stunden
manuell buchen" im Verwaltungs-Dashboard).
