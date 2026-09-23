# OpenSpec: Seite 09 – Berichte & Export (Admin)

**Version:** 1.0  
**Stand:** 2026-06-04  
**Status:** Draft

---

## 1. Zweck

Reporting-Seite für Koordinatoren und Admins. Auswertung der Schichtbelegung, Helfer-Statistiken und Export der Daten in verschiedenen Formaten.

---

## 2. Route

| Route | Beschreibung |
|---|---|
| `/admin/berichte` | Berichte & Export |

---

## 3. User Stories

| ID | Als... | möchte ich... | damit... |
|---|---|---|---|
| RPT-01 | Admin | eine Übersicht der Belegungsquoten aller Bereiche sehen | ich Engpässe identifizieren kann |
| RPT-02 | Admin | alle Schichten mit Helfer-Zuordnung als Excel exportieren | ich die Daten für andere Tools nutzen kann |
| RPT-03 | Koordinator | eine Anwesenheitsliste meines Bereichs exportieren | ich sie am Veranstaltungstag ausdrucken kann |
| RPT-04 | Admin | eine Helfer-Übersicht mit Stundenzahl exportieren | ich Einsatzzeiten nachweisen kann |
| RPT-05 | Admin | visuelle Auswertungen sehen | ich Trends und Muster erkennen kann |

---

## 4. UI-Anforderungen

### 4.1 Statistik-Übersicht

**Belegungsübersicht:**
- Tabelle: Bereich | Schichten gesamt | Schichten voll | Schichten offen | Belegungsquote %
- Balkendiagramm: Belegungsquote pro Bereich
- Summenzeile gesamt

**Helfer-Statistiken:**
- Gesamtzahl Helfer
- Durchschnittliche Schichten pro Helfer
- Top 10 Helfer (nach Anzahl Schichten / Stunden)
- Helfer ohne Schicht-Anmeldungen

**Timeline-Ansicht:**
- Heatmap: Wie viele Helfer sind zu welchem Zeitpunkt eingeplant?
- Zeigt potenzielle Unterbesetzungs-Zeiträume visuell

### 4.2 Export-Bereich

**Verfügbare Exports:**
| Export | Format | Inhalt |
|---|---|---|
| Schichtplan komplett | Excel / CSV / PDF | Alle Schichten mit Helfer-Namen |
| Anwesenheitsliste | PDF | Pro Bereich: Schicht, Zeit, Helfer-Namen mit Unterschriftsfeld |
| Helfer-Stunden | Excel / CSV | Helfer-Name, Schichten, Gesamtstunden |
| Schichten ohne Besetzung | Excel / CSV | Alle Schichten unter min_helfer |
| Alle Anmeldungen | Excel / CSV | Raw-Export aller ShiftRegistrations |

**Export-Dialog:**
- Veranstaltungs-Selector
- Bereichsfilter (optional)
- Datumsbereich (optional)
- Format-Auswahl
- Button "Exportieren"

---

## 5. API-Endpunkte

| Methode | Pfad | Beschreibung |
|---|---|---|
| `GET` | `/api/v1/admin/reports/coverage` | Belegungsquoten-Report |
| `GET` | `/api/v1/admin/reports/helpers-stats` | Helfer-Statistiken |
| `GET` | `/api/v1/admin/reports/timeline` | Timeline/Heatmap-Daten |
| `GET` | `/api/v1/admin/reports/export/shift-plan` | Schichtplan-Export |
| `GET` | `/api/v1/admin/reports/export/attendance-list` | Anwesenheitslisten-Export |
| `GET` | `/api/v1/admin/reports/export/helper-hours` | Helfer-Stunden-Export |
| `GET` | `/api/v1/admin/reports/export/uncovered-shifts` | Schichten ohne Besetzung |

---

## 6. Business Rules

- Koordinatoren sehen nur Daten ihres eigenen Bereichs.
- Admins sehen alle Bereiche und Veranstaltungen.
- PDF-Exports sind für den Druck optimiert (A4, schwarz-weiß kompatibel).
- Anwesenheitsliste enthält kein Foto oder sensible Daten, nur Name und Unterschriftsfeld.
- Export-Requests können bei großen Datenmengen asynchron verarbeitet werden (Polling oder SSE für Fortschrittsanzeige).

---

## 7. Akzeptanzkriterien

- [ ] Belegungsübersicht zeigt korrekte Werte für alle Bereiche
- [ ] Diagramme rendern korrekt
- [ ] Schichtplan-Export (Excel/CSV/PDF) funktioniert und enthält korrekte Daten
- [ ] Anwesenheitsliste ist druckoptimiert (PDF)
- [ ] Helfer-Stunden werden korrekt summiert
- [ ] Koordinator sieht nur seinen Bereich
- [ ] Export-Filter (Bereich, Datum) schränken Daten korrekt ein
