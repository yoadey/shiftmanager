# OpenSpec: Seite 03 – Schichtenplan

**Version:** 1.0  
**Stand:** 2026-06-04  
**Status:** Draft

---

## 1. Zweck

Öffentliche/halböffentliche Übersichtsseite aller Schichten einer Veranstaltung. Helfer können sich von hier aus für Schichten anmelden. Kalender- und Listenansicht sind verfügbar.

---

## 2. Route

| Route | Beschreibung |
|---|---|
| `/schichten` | Schichtenplan (Standard: aktive Veranstaltung) |
| `/schichten?event=<id>` | Schichtenplan für bestimmte Veranstaltung |
| `/schichten?area=<id>` | Gefiltert auf Bereich |

---

## 3. User Stories

| ID | Als... | möchte ich... | damit... |
|---|---|---|---|
| PLAN-01 | Helfer | alle verfügbaren Schichten übersichtlich sehen | ich mich für passende Schichten anmelden kann |
| PLAN-02 | Helfer | nach Datum, Bereich oder Uhrzeit filtern | ich schnell relevante Schichten finde |
| PLAN-03 | Helfer | den Belegungsstatus einer Schicht sehen | ich weiß ob noch Plätze frei sind |
| PLAN-04 | Helfer | mich direkt aus dem Plan für eine Schicht anmelden | ich nicht zu einer anderen Seite navigieren muss |
| PLAN-05 | Gast | den Plan ohne Login einsehen | ich mich über Schichten informieren kann |
| PLAN-06 | Helfer | Schichten in einer Kalenderansicht sehen | ich Überschneidungen meiner eigenen Schichten erkenne |

---

## 4. UI-Anforderungen

### 4.1 Toolbar / Filter
- Veranstaltungs-Selector (Dropdown, falls mehrere aktiv)
- Datumsfilter: Datumsbereich-Picker
- Bereichsfilter: Multiselect-Dropdown (alle Bereiche der Veranstaltung)
- Filter: "Nur offene Schichten" (Toggle)
- Filter: "Nur meine Schichten" (nur wenn eingeloggt)
- Ansichts-Toggle: Liste / Kalender / Timeline

### 4.2 Listenansicht
- Gruppiert nach Datum
- Pro Schicht: Name, Bereich (farbiger Badge), Zeit (Start–Ende), Dauer, Belegung (X/Y Helfer), Status-Badge
- Belegungsbalken (visuell)
- Button "Anmelden" (deaktiviert wenn voll oder bereits angemeldet, ausgeblendet wenn nicht eingeloggt)
- Button "Abmelden" wenn bereits angemeldet
- Schichten die der Helfer hat: visuell hervorgehoben (z. B. Häkchen-Icon, anderer Hintergrund)

### 4.3 Kalenderansicht
- Wochenansicht (default) + Tagesansicht
- Schichten als Blöcke in der entsprechenden Zeile
- Farbe nach Bereich
- Klick auf Schicht → öffnet Schicht-Detail-Drawer (von rechts)
- Eigene Schichten anders eingefärbt (z. B. hellblau)

### 4.4 Schicht-Detail-Drawer
- Schichtname, Bereich, Beschreibung
- Zeit, Dauer
- Belegung: Liste der bereits angemeldeten Helfer (Name, nur für Koordinatoren/Admins; für Helfer: Anzahl angezeigt)
- Button "Anmelden" / "Abmelden"
- Schließen-Button (X)

### 4.5 Anmelde-Bestätigungsdialog
- Schichtdetails zusammengefasst
- Bestätigen / Abbrechen
- Nach Bestätigung: Toast "Erfolgreich angemeldet" + Schicht wird als "eigene Schicht" markiert

---

## 5. API-Endpunkte

| Methode | Pfad | Beschreibung |
|---|---|---|
| `GET` | `/api/v1/shifts` | Alle Schichten (mit Filterparametern) |
| `GET` | `/api/v1/shifts/:id` | Einzelne Schicht mit Details |
| `POST` | `/api/v1/shifts/:id/register` | Für Schicht anmelden |
| `DELETE` | `/api/v1/shifts/:id/register` | Von Schicht abmelden |
| `GET` | `/api/v1/shifts/:id/registrations` | Angemeldete Helfer (Koordinator/Admin) |

**Query-Parameter für GET /api/v1/shifts:**
```
?event_id=&area_id=&date_from=&date_to=&status=open&my_shifts=true&page=&limit=
```

---

## 6. Business Rules

- Ein Helfer kann sich nicht für zwei Schichten anmelden, die sich zeitlich überschneiden.
- Abmeldung von einer Schicht ist nur bis X Stunden vor Beginn möglich (konfigurierbar, z. B. 4 Stunden; danach nur durch Koordinator/Admin).
- Schichten mit `status = closed` oder `cancelled` können nicht belegt werden.
- Schichten mit `max_helfer` erreicht: Anmeldung nicht möglich (Button deaktiviert + Anzeige "Ausgebucht").
- Gäste (nicht eingeloggt) sehen den Plan, können sich aber nicht anmelden → "Anmelden" Button verlinkt zu `/login?next=...`

---

## 7. Akzeptanzkriterien

- [ ] Schichtenplan lädt und zeigt alle Schichten der aktiven Veranstaltung
- [ ] Filter nach Datum, Bereich und Verfügbarkeit funktionieren
- [ ] Listenansicht und Kalenderansicht sind nutzbar und korrekt
- [ ] Anmeldung über "Anmelden"-Button funktioniert mit Bestätigungsdialog
- [ ] Zeitkonflikt-Prüfung verhindert doppelte Belegung
- [ ] Voll belegte Schichten sind als "Ausgebucht" markiert und nicht anmeldbar
- [ ] Gäste sehen den Plan (read-only)
- [ ] Eigene Schichten sind visuell hervorgehoben
