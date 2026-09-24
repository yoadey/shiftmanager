# 03 — Veranstaltungen und Schichten

Umgesetzt. Anforderungs-IDs: `V-001`–`V-008`, `VC-001`–`VC-004`,
`SC-001`–`SC-010`.

## Veranstaltungen

| ID | Anforderung | Umsetzung |
|---|---|---|
| V-001 | Veranstaltungsleiter/Admin (und Vorstand, per Rollenhierarchie) erstellen Veranstaltungen | `POST /api/v1/events`, `RequireRole(RoleVeranstaltungsleiter)` |
| V-002 | Felder: Name, Beschreibung, Start-/Enddatum, Ort, Kategorie, Sichtbarkeit | `domain.Event` |
| V-003 | Mehrtägige Veranstaltungen | `Event.Days []EventDay` |
| V-004 | Zeitplan-Ansicht: mehrtägige Events tageweise untereinander, je Tag eine Zeile mit Schichten | `EventDetail.tsx` Timeline |
| V-005 | Status: Entwurf, Veröffentlicht, Abgeschlossen, Abgesagt | `domain.EventStatus*` |
| V-006 | Nur veröffentlichte Events für Mitglieder sichtbar | `GET /api/v1/events` Filter |
| V-007 | Wiederkehrende Veranstaltungen (wöchentlich/monatlich) | `POST /api/v1/events/{id}/recurrence` — vollständige Kopien inkl. Schichten; monatliches Schrittmaß begrenzt Tag-Überlauf auf den letzten Tag des Zielmonats statt in den Folgemonat zu rutschen; Verknüpfung der Vorkommen über atomares, bedingtes `MarkRecurring`-Update als letzten Schritt (Rollback bei Teilfehler); ein abgesagtes/abgeschlossenes Event kann nicht Kopf einer neuen Serie werden |
| V-008 | Bilder/Anhänge an Veranstaltungen | `POST/GET/DELETE /api/v1/events/{id}/attachments`; Validierung per Extension+Content-Type-Allowlist (SVG bewusst ausgeschlossen, Stored-XSS-Risiko), 5 MB Limit; `/uploads/*`-Server ohne Directory-Listing, `X-Content-Type-Options: nosniff` |

### Veranstaltungen kopieren

| ID | Anforderung | Umsetzung |
|---|---|---|
| VC-001 | Veranstaltungsleiter/Admin kopieren ein Event inkl. aller Schichten | `POST /api/v1/events/{id}/copy` |
| VC-002 | Kopie erhält Status "Entwurf", Name "Kopie von [Original]" | `CopyEvent`-Usecase |
| VC-003 | Alle Schichtfelder identisch übertragen (Name, Zeiten, Helferanzahl, Qualifikationen) | ebenda |
| VC-004 | Erreichbar über Button in der Veranstaltungsübersicht | `AdminEvents.tsx` |

## Schichten

| ID | Anforderung | Umsetzung |
|---|---|---|
| SC-001 | Mehrere Schichten pro Veranstaltung | `Shift` gehört zu `EventDay` |
| SC-002 | Felder: Name, Start-/Endzeit, Min./Max.-Helfer, Beschreibung | `domain.Shift` |
| SC-003 | Übersichtliche Zeitplan-/Timeline-Darstellung | `EventDetail.tsx` |
| SC-004 | Unterbesetzte Schichten optisch hervorgehoben | `calcOccupancy()`, Badge-Farbe |
| SC-005 | Vollständig belegte Schichten als "Ausgebucht" markiert | ebenda |
| SC-006 | Abmeldung nur bis zu konfigurierbarer Frist vor Schichtbeginn | `AppSettings.DeregisterDeadlineH` |
| SC-007 | Benachrichtigung an Veranstaltungsleiter bei Unterschreitung der Mindesthelferzahl | E-Mail-Typ "Unterstaffed" |
| SC-008 | Veranstaltungsleiter schließt Schichten ab, bestätigt individuell die tatsächlich geleistete Zeit | `POST /api/v1/hours/confirm`, `RequireRole(RoleVeranstaltungsleiter)` |
| SC-009 | Schichten können Qualifikationsanforderungen haben (z. B. Führerschein) | `Shift.Qualification` |
| SC-010 | Veranstaltungsleiter/Vorstand tragen beliebige Helfer ein — auch Personen ohne Benutzerkonto (nur Name, optional E-Mail) — und entfernen jede Eintragung (eigene oder fremde) wieder; Übersicht zeigt stets vollständige Namen | `POST /api/v1/shifts/{id}/add-member`, `POST /api/v1/shifts/{id}/add-guest` (`RegistrationUsecase.OrganizerAddGuest`), `PATCH/DELETE /api/v1/registrations/{id}` (`RequireRole(RoleVeranstaltungsleiter)`); UI: `ShiftRow`/`RegistrationRow` in `EventDetail.tsx`, gated über `useIsEventManager()` |

**Berechtigungsgrenze bei SC-010:** Das Zutragen/Entfernen beliebiger
Helfer ist eine **Veranstaltungsleiter+**-Berechtigung
(`RequireRole(RoleVeranstaltungsleiter)`), keine Vorstand-only-Funktion —
konsistent mit V-001/SC-008, wo derselbe Rollen-Schnitt für die
Schichtverwaltung gilt. Die frühere freie Formulierung "Vorstand kann …"
aus der ursprünglichen Anfrage ist hier entsprechend auf die tatsächliche
Backend-Rollenschwelle präzisiert.

Doppelte Gast-E-Mail auf derselben Schicht wird abgelehnt
(`domain.ErrAlreadyRegistered`), analog zur bestehenden Prüfung bei
`Register`/`OrganizerRegister`.
