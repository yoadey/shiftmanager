# Change: Formulare mit mehr als 5 Feldern als eigene Seite statt Popup

**Status:** Vorgeschlagen — ein neuer globaler UI-Grundsatz plus zwei
konkrete Umbauten, die ihn heute schon verletzen.
**Betrifft:** [`00-general.md`](../00-general.md) (neuer globaler
Grundsatz), [`03-veranstaltungen-schichten.md`](../03-veranstaltungen-schichten.md).

## Motivation

Umfangreiche Formulare als Popup/Sheet-Overlay (`components/ui/Sheet.tsx`)
sind auf kleinen wie großen Bildschirmen unübersichtlich: kein Platz für
Kontext, Klick außerhalb schließt versehentlich, keine Navigation
sichtbar. Für kurze Eingaben oder Bestätigungen ist das Sheet-Muster
weiterhin passend; für umfangreiche Konfigurationsformulare nicht.

## Neuer globaler Grundsatz

| ID | Anforderung |
|---|---|
| UX-001 | Formulare/Konfigurationsdialoge mit **mehr als 5 Eingabefeldern** werden als eigene Seite (eigene Route, volle App-Navigation sichtbar, kein Overlay/Backdrop) umgesetzt statt als Popup/Sheet/Modal. Die `Sheet`-Komponente (`variant="dialog"`, `variant="full"` sowie der Standard-Bottom-Sheet) bleibt reserviert für Bestätigungen, Kurz-Eingaben und Formulare mit ≤5 Feldern. |

## Ist-Zustand-Audit (Sheet-basierte Formulare)

Alle `<Sheet>`-Verwendungen in `frontend/src/screens/**` wurden auf
Feldanzahl geprüft:

| Screen | Datei/Zeile | Felder | Einordnung |
|---|---|---|---|
| Veranstaltung erstellen | `screens/admin/CreateEventFlow.tsx` (`Sheet variant="full"`) | Name, Beschreibung, Ort, Kategorie, Start, Ende, Status/Sichtbarkeit + dynamische Schichtliste | **>5 → Verstoß, umbauen** |
| Veranstaltung bearbeiten | `screens/events/EventDetail.tsx:699` (`Sheet variant="full"`) | Name, Beschreibung, Ort, Kategorie, Von, Bis, Status + Schichten | **>5 → Verstoß, umbauen** |
| Schicht anlegen/bearbeiten | `screens/events/EventDetail.tsx:571` (`ShiftForm`, aktuell Inline-Karte innerhalb des Bearbeiten-Sheets) | Bezeichnung, Datum, Von, Bis, Min, Max, Qualifikation | **>5 → Verstoß, umbauen** |
| Mitglied bearbeiten | `screens/admin/MemberDetail.tsx:176` (`Sheet`) | Vorname, Nachname, E-Mail, Eintrittsdatum | ≤5 → konform, bleibt Popup |
| Serie konfigurieren | `screens/admin/AdminEvents.tsx:60` (`Sheet variant="dialog"`) | Wiederholung, Bis einschließlich | ≤5 → konform |
| Helfer/Gast hinzufügen | `screens/events/EventDetail.tsx:284,470` (`Sheet`) | Suche/Name, E-Mail (+ optionaler Kommentar) | ≤5 → konform |
| Löschantrag (Profil) | `screens/member/MemberProfile.tsx:162` (`Sheet variant="dialog"`) | reine Bestätigung, keine Eingabefelder | konform |
| Systemeinstellungen | `screens/admin/AdminSettings.tsx` | — bereits eigene Seite/Route (`/settings`), kein Sheet | bereits konform, keine Änderung nötig |

Fachlich betrifft der Umbau `V-001`–`V-005` (Eckdaten-Formular) und
`SC-002` (Schicht-Felder) nur auf der Präsentationsebene — die
inhaltlichen Feld-Anforderungen selbst ändern sich nicht.

## Design-Entscheidungen (Vorschlag)

1. Routing existiert für Events bereits: `routes.eventNeu` (`/events/neu`)
   und `routes.eventBearbeiten(id)` (`/events/:id/bearbeiten`),
   siehe `frontend/src/routes.ts:28,30`. Diese Pfade rendern aktuell
   `<Sheet variant="full">` als Overlay über der aktuellen Seite (Klick
   außerhalb schließt, eigener X-Button, kein Sidebar/Nav sichtbar —
   siehe `components/ui/Sheet.tsx`). Der Umbau ändert nur das *Rendering*
   auf diesen bereits existierenden Routen: vollwertige Seite innerhalb
   `DesktopShell`/`MobileShell` (Navigation bleibt sichtbar, kein
   Overlay/Backdrop, kein „Klick außerhalb schließt“), mit explizitem
   „Zurück“/„Abbrechen“ statt X-Button.
2. Das Schichtformular (`ShiftForm`) wird Teil derselben neuen
   Event-Seite (Abschnitt „Schichten“), nicht als zusätzliches separates
   Popup.
3. Die `Sheet`-Komponente selbst bleibt für weiterhin zulässige Fälle
   (≤5 Felder, Bestätigungen) unverändert bestehen.
4. UX-001 gilt projektweit für zukünftige Formulare; bereits konforme
   Screens (`AdminSettings`, `MemberDetail`, etc.) werden nicht
   angefasst.

## Tasks

- [ ] `00-general.md`: Abschnitt „Globale Anforderungen“ um `UX-001` ergänzen
- [ ] Frontend: neue Seiten-Komponente für Veranstaltung erstellen/bearbeiten (ersetzt den Sheet in `CreateEventFlow.tsx` und den Bearbeiten-Sheet in `EventDetail.tsx`), gerendert unter den bestehenden Routen `routes.eventNeu`/`routes.eventBearbeiten`
- [ ] Frontend: `ShiftForm` in die neue Seite integrieren statt als Inline-Popup im alten Sheet
- [ ] Frontend: echte Seiten-Navigation (kein Overlay-Backdrop, Browser-Zurück funktioniert)
- [ ] Frontend: alte Sheet-Aufrufe für Event-Create/-Edit entfernen
- [ ] Frontend-Tests anpassen (ggf. auf Sheet-DOM-Struktur geschrieben)
- [ ] E2E-Tests (Playwright) für Event-Erstellung/-Bearbeitung auf neue Seitenstruktur anpassen
- [ ] Nach Umsetzung: `UX-001` dauerhaft in [`00-general.md`](../00-general.md), Rest in [`03-veranstaltungen-schichten.md`](../03-veranstaltungen-schichten.md) übernehmen; diese Datei löschen (siehe [`README.md`](../README.md))
