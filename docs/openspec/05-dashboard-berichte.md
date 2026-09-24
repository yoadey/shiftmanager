# 05 — Dashboards, Berichte, Audit-Log

Umgesetzt. Anforderungs-IDs: `D-001`–`D-006`, `DS-009` (Audit-Aufbewahrung),
`ML-009`/`G-007` (Audit-Schreibpfade — siehe `04-stunden-abrechnung.md`).

## Dashboards

| ID | Anforderung | Umsetzung |
|---|---|---|
| D-001 | Mitglieder-Dashboard: bestätigte Stunden und Stunden inkl. offener Reservierungen, je gegen Jahres-Stundenziel | `MemberDashboard.tsx` |
| D-002 | Bevorstehende eigene Schichten auf dem Dashboard | ebenda |
| D-003 | Öffentlicher Kalender aller veröffentlichten Veranstaltungen mit freien Schichten | `MemberDiscover.tsx` |
| D-004 | Admin-Dashboard mit systemweiten Statistiken (Gesamtstunden, offene Schichten, …) | `GET /api/v1/stats`, `RequireRole(RoleVeranstaltungsleiter)`, `AdminDashboard.tsx` |
| D-005 | Filterung nach Datum, Veranstaltung, verfügbaren Plätzen | `MemberDiscover.tsx` Filter |
| D-006 | Kartenansicht pro Veranstaltung mit Belegungsbalken | `OccFill`/`calcOccupancy` |

Das Verwaltungs-Dashboard (`uebersicht`-Tab) ist ab **Veranstaltungsleiter**
sichtbar (nicht Vorstand-only) — der `/stats`-Endpunkt selbst verlangt nur
`RoleVeranstaltungsleiter`.

## Berichte & Exporte

- Jahresabrechnung als PDF/CSV: siehe `G-006` in `04-stunden-abrechnung.md`.
- Mitgliederliste als CSV: siehe `ML-008` in `01-auth-mitglieder.md`.
- DSGVO-Datenexport pro Mitglied: siehe `09-datenschutz-sicherheit.md`
  (`DS-003`).

## Audit-Log

Alle sicherheits- und datenschutzrelevanten Aktionen (Mitglieder-Änderungen,
Rollen-/Einstellungsänderungen, Stundenkorrekturen, Gebühren-Änderungen,
Branding-Änderungen) werden protokolliert.

| ID | Anforderung | Umsetzung |
|---|---|---|
| T-012 | Alle kritischen Aktionen im Audit-Log (wer/was/wann) | `domain.AuditEntry`, `internal/usecase/*` (`writeAudit`-Helfer) |
| DS-009 | Aufbewahrung mind. 2 Jahre, änderungsgeschützt | `GET /api/v1/settings/audit`, `RequireRole(RoleVorstand)`, `AuditLog.tsx` |

Der Audit-Log-Screen ist Vorstand-only (Settings-Untermenü), nicht für
Veranstaltungsleiter erreichbar.
