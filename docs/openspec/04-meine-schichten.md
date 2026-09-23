# OpenSpec: Seite 04 – Meine Schichten

**Version:** 1.0  
**Stand:** 2026-06-04  
**Status:** Draft

---

## 1. Zweck

Persönliche Seite des Helfers zur Verwaltung seiner eigenen Schichtanmeldungen. Zeigt vergangene, aktuelle und zukünftige Schichten des eingeloggten Nutzers.

---

## 2. Route

| Route | Beschreibung |
|---|---|
| `/meine-schichten` | Persönlicher Schichtplan |

---

## 3. User Stories

| ID | Als... | möchte ich... | damit... |
|---|---|---|---|
| MY-01 | Helfer | alle meine angemeldeten Schichten sehen | ich meinen Einsatzplan kenne |
| MY-02 | Helfer | mich von einer Schicht abmelden können | ich bei Verhinderung stornieren kann |
| MY-03 | Helfer | meine Schichten exportieren/ausdrucken | ich sie offline verfügbar habe |
| MY-04 | Helfer | vergangene Schichten als "Verlauf" sehen | ich meine geleisteten Stunden nachvollziehen kann |
| MY-05 | Helfer | den Status meiner Anmeldung sehen | ich weiß ob meine Anmeldung bestätigt ist |

---

## 4. UI-Anforderungen

### 4.1 Tab-Navigation
- Tab: **Bevorstehend** (zukünftige Schichten)
- Tab: **Vergangen** (bereits absolvierte Schichten)

### 4.2 Schicht-Karte (bevorstehend)
- Schichtname + Bereich (farbiger Badge)
- Datum (z. B. "Samstag, 12. Juli 2025")
- Uhrzeit Start–Ende, Dauer
- Veranstaltungsname
- Status-Badge: "Angemeldet" (gelb) / "Bestätigt" (grün) / "Abgesagt" (rot)
- Button "Abmelden" (nur wenn Abmeldefrist noch nicht abgelaufen)
- Wenn Abmeldefrist abgelaufen: Hinweis "Abmeldung nicht mehr möglich – bitte Koordinator kontaktieren"

### 4.3 Schicht-Karte (vergangen)
- Wie oben, aber ohne "Abmelden"-Button
- Status: "Absolviert" / "No-Show" (wenn Koordinator das markiert hat)
- Keine Aktionsbuttons

### 4.4 Zusammenfassung oben
- Gesamtstunden (bevorstehend + absolviert)
- Anzahl Schichten bevorstehend
- Anzahl Schichten absolviert

### 4.5 Export
- Button "Als PDF exportieren" → generiert Schichtplan als druckfreundliches PDF
- Button "iCal exportieren" → `.ics`-Datei für Kalender-Import (Google Calendar, Apple Kalender etc.)

---

## 5. API-Endpunkte

| Methode | Pfad | Beschreibung |
|---|---|---|
| `GET` | `/api/v1/users/me/shifts` | Eigene Schichten (gefiltert: `?status=upcoming\|past`) |
| `DELETE` | `/api/v1/shifts/:id/register` | Von Schicht abmelden |
| `GET` | `/api/v1/users/me/shifts/export/pdf` | PDF-Export |
| `GET` | `/api/v1/users/me/shifts/export/ical` | iCal-Export |

---

## 6. Business Rules

- Abmeldung nur möglich bis zur konfigurierten Frist (z. B. 4 Stunden vor Schichtbeginn).
- Wenn Abmeldung nicht mehr möglich: Info-Text mit Kontaktmöglichkeit zum Koordinator.
- Stornierte Schichten (durch Admin) bleiben mit Status "Abgesagt" in der Liste sichtbar.
- PDF-Export enthält: Helfer-Name, Veranstaltung, Schicht-Liste sortiert nach Datum.
- iCal-Export erstellt einen Kalendereintrag pro Schicht mit korrekten Zeiten und Ort.

---

## 7. Akzeptanzkriterien

- [ ] Helfer sieht alle seine Schichten (bevorstehend + vergangen)
- [ ] Tab-Wechsel zwischen Bevorstehend/Vergangen funktioniert
- [ ] "Abmelden"-Button erscheint nur wenn Abmeldefrist noch nicht abgelaufen
- [ ] Abmeldung funktioniert mit Bestätigungsdialog
- [ ] Stundensumme wird korrekt berechnet und angezeigt
- [ ] PDF-Export ist generierbar und korrekt formatiert
- [ ] iCal-Export erzeugt gültige .ics-Datei
