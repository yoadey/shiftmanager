# OpenSpec: Seite 08 – Veranstaltungs-Verwaltung (Admin)

**Version:** 1.0  
**Stand:** 2026-06-04  
**Status:** Draft

---

## 1. Zweck

Anlegen und Verwalten von Veranstaltungen. Eine Veranstaltung ist der übergeordnete Container für alle Bereiche und Schichten. Typische Beispiele: Stadtfest, Festival, Vereinsveranstaltung.

---

## 2. Routen

| Route | Beschreibung |
|---|---|
| `/admin/veranstaltungen` | Veranstaltungs-Übersicht |
| `/admin/veranstaltungen/neu` | Neue Veranstaltung anlegen |
| `/admin/veranstaltungen/:id` | Veranstaltungsdetails |
| `/admin/veranstaltungen/:id/bearbeiten` | Veranstaltung bearbeiten |

---

## 3. User Stories

| ID | Als... | möchte ich... | damit... |
|---|---|---|---|
| EVT-01 | Admin | eine neue Veranstaltung anlegen | Schichten und Helfer dafür verwaltet werden können |
| EVT-02 | Admin | eine Veranstaltung archivieren | vergangene Events aus der aktiven Ansicht verschwinden |
| EVT-03 | Admin | eine Veranstaltung aus einer Vorlage erstellen | ich Bereiche und Schichten aus einem früheren Event wiederverwenden kann |
| EVT-04 | Admin | den Status einer Veranstaltung steuern | ich kontrollieren kann wann Helfer sich anmelden können |
| EVT-05 | Admin | eine Veranstaltung löschen | ich Testevents entfernen kann |

---

## 4. UI-Anforderungen

### 4.1 Veranstaltungs-Übersicht (`/admin/veranstaltungen`)

- Kachelansicht mit Veranstaltungen, gruppiert nach Status (Aktiv / Entwurf / Archiviert)
- Pro Kachel:
  - Name, Datum (Start–Ende), Ort
  - Status-Badge
  - Anzahl Bereiche, Schichten, Helfer
  - Belegungsquote gesamt
  - Quick-Links: Bereiche, Schichten, Helfer
  - Aktionen: Bearbeiten, Archivieren, Löschen
- Button "Neue Veranstaltung"

### 4.2 Veranstaltung anlegen / bearbeiten

**Formular-Felder:**
- Name (Pflicht)
- Beschreibung (optional, Textarea)
- Startdatum + Startzeit (Pflicht)
- Enddatum + Endzeit (Pflicht)
- Ort/Adresse (optional)
- Zeitzone (Dropdown, default: Europe/Berlin)
- Status: Entwurf / Aktiv / Archiviert
- Anmeldung erlaubt ab (Datum, optional – bevor diesem Datum können Helfer sich nicht anmelden)
- Anmeldung erlaubt bis (Datum, optional)
- Selbstregistrierung: Ja/Nein Toggle
- Automatische Genehmigung: Ja/Nein Toggle
- Logo/Bild (optional, Upload)
- Abmelde-Frist in Stunden (Number, z. B. 4)

### 4.3 Veranstaltungsdetails (`/admin/veranstaltungen/:id`)

- Alle Stammdaten anzeigen (read-only)
- Statistiken: Helfer total, Schichten total/besetzt/offen, Belegungsquote gesamt
- Tabs: Bereiche, Schichten, Helfer
- Button "Bearbeiten"
- Button "Aus Vorlage erstellen" → Wizard: Bereiche + Schichtstruktur aus einem Archiv-Event kopieren (ohne Anmeldungen)

---

## 5. API-Endpunkte

| Methode | Pfad | Beschreibung |
|---|---|---|
| `GET` | `/api/v1/admin/events` | Alle Veranstaltungen |
| `POST` | `/api/v1/admin/events` | Neue Veranstaltung anlegen |
| `GET` | `/api/v1/admin/events/:id` | Veranstaltungsdetails |
| `PUT` | `/api/v1/admin/events/:id` | Veranstaltung bearbeiten |
| `DELETE` | `/api/v1/admin/events/:id` | Veranstaltung löschen |
| `POST` | `/api/v1/admin/events/:id/archive` | Veranstaltung archivieren |
| `POST` | `/api/v1/admin/events/:id/clone` | Veranstaltung als Vorlage klonen |
| `GET` | `/api/v1/events/active` | Aktive Veranstaltungen (für Helfer-UI) |

---

## 6. Business Rules

- Löschen ist nur für Veranstaltungen im Status "Entwurf" möglich (ohne aktive Anmeldungen).
- Aktive oder archivierte Veranstaltungen können nur archiviert, nicht gelöscht werden.
- Beim Klonen werden nur Bereiche und Schichtstruktur (Name, Zeit, Min/Max) kopiert – keine Anmeldungen.
- Wenn `Anmeldung erlaubt ab` in der Zukunft liegt: Helfer sehen die Schichten, können sich aber noch nicht anmelden (Countdown anzeigen).
- Eine Veranstaltung ohne aktiven Status erscheint nicht im Helfer-Schichtenplan.

---

## 7. Akzeptanzkriterien

- [ ] Neue Veranstaltung kann angelegt werden
- [ ] Statuswechsel (Entwurf → Aktiv → Archiviert) funktioniert
- [ ] Klonen-Funktion kopiert Bereiche und Schichten ohne Anmeldungen
- [ ] Löschen von aktiven Veranstaltungen ist gesperrt
- [ ] Zeitzone wird korrekt gespeichert und alle Zeiten werden entsprechend angezeigt
- [ ] `Anmeldung erlaubt ab` Datum sperrt korrekt die Anmeldung für Helfer
