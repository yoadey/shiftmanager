# OpenSpec: Seite 05 – Schicht-Verwaltung (Admin)

**Version:** 1.0  
**Stand:** 2026-06-04  
**Status:** Draft

---

## 1. Zweck

Verwaltung aller Schichten einer Veranstaltung durch Koordinatoren und Admins. Anlegen, Bearbeiten, Löschen von Schichten sowie Verwaltung der Belegung (Helfer manuell zu- oder abweisen).

---

## 2. Routen

| Route | Beschreibung |
|---|---|
| `/admin/schichten` | Übersicht aller Schichten (Admin/Koordinator) |
| `/admin/schichten/neu` | Neue Schicht anlegen |
| `/admin/schichten/:id/bearbeiten` | Schicht bearbeiten |
| `/admin/schichten/:id/helfer` | Helfer einer Schicht verwalten |

---

## 3. User Stories

| ID | Als... | möchte ich... | damit... |
|---|---|---|---|
| SMGMT-01 | Admin | neue Schichten anlegen | Helfer sich darauf anmelden können |
| SMGMT-02 | Admin | Schichten bearbeiten | ich Fehler korrigieren oder Zeiten ändern kann |
| SMGMT-03 | Admin | Schichten löschen / abbrechen | ich nicht benötigte Schichten entfernen kann |
| SMGMT-04 | Koordinator | Schichten meines Bereichs sehen und verwalten | ich meinen Bereich selbst pflegen kann |
| SMGMT-05 | Admin | Helfer manuell zu einer Schicht hinzufügen | ich auch ohne Selbstanmeldung des Helfers planen kann |
| SMGMT-06 | Admin | Helfer aus einer Schicht entfernen | ich Änderungen direkt durchführen kann |
| SMGMT-07 | Koordinator | Helfer-Anmeldungen bestätigen oder ablehnen | ich die Kontrolle über meine Schichten habe |
| SMGMT-08 | Admin | Schichten duplizieren | ich wiederkehrende Schichten schnell erstellen kann |
| SMGMT-09 | Admin | Schichten als CSV/Excel exportieren | ich die Daten weiterverarbeiten kann |

---

## 4. UI-Anforderungen

### 4.1 Schicht-Übersicht (`/admin/schichten`)

**Filter-Toolbar:**
- Veranstaltungs-Selector
- Datumsbereich
- Bereichsfilter
- Statusfilter (Offen / Voll / Abgesagt / Alle)
- Suchfeld (Name der Schicht)

**Tabelle / Liste:**
| Spalte | Beschreibung |
|---|---|
| Name | Schichtname |
| Bereich | Farbiger Badge |
| Datum | Datum |
| Zeit | Start – Ende |
| Belegung | X / Y (Fortschrittsbalken) |
| Status | Badge (Offen/Voll/Abgesagt) |
| Aktionen | Bearbeiten, Helfer, Duplizieren, Löschen |

- Sortierung nach allen Spalten möglich
- Bulk-Aktionen: Mehrere Schichten markieren → "Abbrechen", "Löschen"
- Button "Neue Schicht" oben rechts
- Button "Exportieren" (CSV/Excel)

### 4.2 Schicht anlegen / bearbeiten

**Formular-Felder:**
- Name (Pflicht)
- Beschreibung (optional, Textarea)
- Bereich (Dropdown, Pflicht)
- Datum (Datepicker, Pflicht)
- Startzeit (Timepicker, Pflicht)
- Endzeit (Timepicker, Pflicht) – Validierung: Endzeit nach Startzeit
- Min. Helfer (Number-Input, min. 1)
- Max. Helfer (Number-Input, ≥ Min. Helfer)
- Status (Offen / Geschlossen / Abgesagt)
- Notiz für Helfer (optional, Textarea)

**Aktionen:**
- Speichern
- Abbrechen
- "Weitere Schicht mit gleichen Daten anlegen" (Checkbox beim Erstellen)

### 4.3 Helfer-Verwaltung einer Schicht (`/admin/schichten/:id/helfer`)

- Liste aller angemeldeten Helfer:
  - Name, Email, Anmeldedatum, Status (Angemeldet/Bestätigt/No-Show)
  - Aktionen: Bestätigen, No-Show markieren, Entfernen
- Helfer manuell hinzufügen: Suchfeld (nach Name/Email) → Auswahl → Hinzufügen
- Gesamtstatus der Schicht oben: X/Y Helfer besetzt

---

## 5. API-Endpunkte

| Methode | Pfad | Beschreibung |
|---|---|---|
| `GET` | `/api/v1/admin/shifts` | Alle Schichten (Admin-View, gefiltert) |
| `POST` | `/api/v1/admin/shifts` | Neue Schicht anlegen |
| `GET` | `/api/v1/admin/shifts/:id` | Schicht-Details |
| `PUT` | `/api/v1/admin/shifts/:id` | Schicht bearbeiten |
| `DELETE` | `/api/v1/admin/shifts/:id` | Schicht löschen |
| `POST` | `/api/v1/admin/shifts/:id/duplicate` | Schicht duplizieren |
| `GET` | `/api/v1/admin/shifts/:id/registrations` | Helfer der Schicht |
| `POST` | `/api/v1/admin/shifts/:id/registrations` | Helfer manuell hinzufügen |
| `PUT` | `/api/v1/admin/shifts/:id/registrations/:reg_id` | Anmeldungsstatus ändern |
| `DELETE` | `/api/v1/admin/shifts/:id/registrations/:reg_id` | Helfer entfernen |
| `GET` | `/api/v1/admin/shifts/export` | CSV/Excel-Export |

---

## 6. Business Rules

- Koordinatoren können nur Schichten in ihrem zugewiesenen Bereich verwalten.
- Admins können alle Schichten verwalten.
- Beim Löschen einer Schicht mit aktiven Anmeldungen: Warnung anzeigen + alle Helfer erhalten automatisch eine Benachrichtigungs-Email.
- Beim Abbrechen (`status = cancelled`) einer Schicht: alle Helfer werden per Email informiert.
- Endzeit muss nach Startzeit liegen (Schichten können über Mitternacht gehen: z. B. 22:00–02:00).
- Max. Helfer muss ≥ Min. Helfer sein.
- Schicht-Duplikat hat Status "Offen" und keine Anmeldungen.

---

## 7. Akzeptanzkriterien

- [ ] Koordinator sieht nur Schichten seines Bereichs
- [ ] Admin sieht alle Schichten
- [ ] Neue Schicht kann angelegt werden (alle Pflichtfelder validiert)
- [ ] Schicht kann bearbeitet werden
- [ ] Schicht-Löschung mit Anmeldungen zeigt Warnmeldung
- [ ] Helfer können manuell zu Schichten hinzugefügt werden
- [ ] Helfer können aus Schichten entfernt werden
- [ ] Anmeldungsstatus kann geändert werden (Bestätigt/No-Show)
- [ ] Schicht-Duplikat erstellt neue Schicht ohne Anmeldungen
- [ ] Export als CSV funktioniert
