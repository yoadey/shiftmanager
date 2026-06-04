# OpenSpec: Seite 07 – Bereiche-Verwaltung (Admin)

**Version:** 1.0  
**Stand:** 2026-06-04  
**Status:** Draft

---

## 1. Zweck

Verwaltung der Bereiche/Stationen einer Veranstaltung. Bereiche strukturieren die Schichten thematisch oder örtlich (z. B. "Eingang", "Bar", "Bühne", "Security").

---

## 2. Routen

| Route | Beschreibung |
|---|---|
| `/admin/bereiche` | Bereiche-Übersicht |
| `/admin/bereiche/neu` | Neuen Bereich anlegen |
| `/admin/bereiche/:id/bearbeiten` | Bereich bearbeiten |

---

## 3. User Stories

| ID | Als... | möchte ich... | damit... |
|---|---|---|---|
| AREA-01 | Admin | Bereiche anlegen | ich Schichten thematisch/örtlich gruppieren kann |
| AREA-02 | Admin | Bereiche bearbeiten | ich Name, Farbe oder Beschreibung anpassen kann |
| AREA-03 | Admin | einem Bereich einen Koordinator zuweisen | eine verantwortliche Person den Bereich verwaltet |
| AREA-04 | Admin | Bereiche löschen | ich nicht benötigte Bereiche entfernen kann |
| AREA-05 | Admin | die Belegungsquote pro Bereich sehen | ich den Fortschritt des gesamten Events überblicke |

---

## 4. UI-Anforderungen

### 4.1 Bereiche-Übersicht (`/admin/bereiche`)

- Kachel- oder Listenansicht der Bereiche
- Pro Bereich:
  - Name + farbiger Kreis (Bereichsfarbe)
  - Koordinator (Name oder "Nicht zugewiesen")
  - Anzahl Schichten
  - Belegungsquote (Fortschrittsbalken: X von Y Schichtplätzen besetzt)
  - Aktionen: Bearbeiten, Löschen
- Button "Neuen Bereich anlegen"
- Veranstaltungs-Selector oben (falls mehrere Veranstaltungen)

### 4.2 Bereich anlegen / bearbeiten

**Formular-Felder:**
- Veranstaltung (Dropdown, Pflicht)
- Name (Pflicht, max. 100 Zeichen)
- Beschreibung (optional, Textarea)
- Farbe (Farbpicker, Pflicht – wird im Schichtenplan als Badge-Farbe verwendet)
- Koordinator (Suchfeld → Dropdown aus Koordinatoren/Admins, optional)
- Ort/Standort (optional, Freitext, z. B. "Zelt A, Eingang Nord")

---

## 5. API-Endpunkte

| Methode | Pfad | Beschreibung |
|---|---|---|
| `GET` | `/api/v1/admin/areas` | Alle Bereiche (gefiltert nach Event) |
| `POST` | `/api/v1/admin/areas` | Neuen Bereich anlegen |
| `GET` | `/api/v1/admin/areas/:id` | Bereich-Details |
| `PUT` | `/api/v1/admin/areas/:id` | Bereich bearbeiten |
| `DELETE` | `/api/v1/admin/areas/:id` | Bereich löschen |
| `GET` | `/api/v1/areas` | Öffentliche Bereiche-Liste (für Schichtenplan-Filter) |

---

## 6. Business Rules

- Ein Bereich kann nicht gelöscht werden wenn er noch aktive (nicht abgesagte) Schichten hat. Erst alle Schichten löschen/abbrechen.
- Der Koordinator eines Bereichs erhält automatisch die Rolle `coordinator` wenn er noch `helper` ist (mit Hinweis-Dialog).
- Bereiche sind immer einer Veranstaltung zugeordnet. Es gibt keine event-übergreifenden Bereiche (aber gleiche Namen in verschiedenen Veranstaltungen sind erlaubt).
- Die Farbe wird überall konsistent genutzt (Schichtenplan, Dashboard, Berichte).

---

## 7. Akzeptanzkriterien

- [ ] Bereiche können angelegt, bearbeitet und gelöscht werden
- [ ] Farbpicker funktioniert und die Farbe wird im Schichtenplan korrekt angezeigt
- [ ] Koordinator kann einem Bereich zugewiesen werden
- [ ] Löschen mit aktiven Schichten zeigt Fehlermeldung
- [ ] Belegungsquote wird korrekt berechnet
