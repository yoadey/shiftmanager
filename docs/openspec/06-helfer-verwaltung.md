# OpenSpec: Seite 06 – Helfer-Verwaltung (Admin)

**Version:** 1.0  
**Stand:** 2026-06-04  
**Status:** Draft

---

## 1. Zweck

Verwaltung aller registrierten Helfer/Nutzer durch Admins. Benutzerverwaltung, Rollenzuweisung, Einladungen und Übersicht der Schichten pro Helfer.

---

## 2. Routen

| Route | Beschreibung |
|---|---|
| `/admin/helfer` | Helfer-Übersicht |
| `/admin/helfer/:id` | Helfer-Detailansicht |
| `/admin/helfer/einladen` | Helfer einladen |

---

## 3. User Stories

| ID | Als... | möchte ich... | damit... |
|---|---|---|---|
| HMgmt-01 | Admin | alle registrierten Helfer sehen | ich einen Überblick über meine Community habe |
| HMgmt-02 | Admin | Helfer einladen | ich gezielt Personen onboarden kann |
| HMgmt-03 | Admin | einem Helfer eine Rolle zuweisen | ich Koordinatoren ernennen kann |
| HMgmt-04 | Admin | einen Helfer deaktivieren | gesperrte Nutzer keinen Zugang mehr haben |
| HMgmt-05 | Admin | alle Schichten eines Helfers sehen | ich seinen Einsatz nachvollziehen kann |
| HMgmt-06 | Admin | ausstehende Registrierungen genehmigen/ablehnen | ich kontrollieren kann wer Zugang erhält |
| HMgmt-07 | Admin | einen Helfer löschen (DSGVO) | ich Datenlöschanfragen nachkommen kann |

---

## 4. UI-Anforderungen

### 4.1 Helfer-Übersicht (`/admin/helfer`)

**Filter-Toolbar:**
- Suchfeld (Name oder Email)
- Rollenfilter (Helfer / Koordinator / Admin)
- Statusfilter (Aktiv / Inaktiv / Ausstehend)
- Veranstaltungsfilter: "Hat sich für Schichten in Veranstaltung X angemeldet"

**Tabelle:**
| Spalte | Beschreibung |
|---|---|
| Name | Vor- und Nachname |
| Email | Email-Adresse |
| Rolle | Badge (Helfer/Koordinator/Admin) |
| Schichten | Anzahl angemeldeter Schichten |
| Registriert am | Datum |
| Status | Aktiv / Inaktiv / Ausstehend |
| Aktionen | Profil, Schichten, Rolle, Deaktivieren |

- Sortierung nach Name, Registrierungsdatum, Anzahl Schichten
- Bulk-Aktionen: E-Mail-Broadcast an ausgewählte Helfer
- Button "Helfer einladen" oben rechts
- Button "Exportieren" (CSV/Excel)

**Tab: Ausstehende Registrierungen**
- Separate Liste mit Helfern, die auf Freischaltung warten
- Aktionen: Genehmigen / Ablehnen (mit optionalem Ablehnungsgrund)

### 4.2 Helfer-Detailansicht (`/admin/helfer/:id`)

**Profil-Bereich:**
- Vor-/Nachname, Email, Telefon (bearbeitbar durch Admin)
- Rolle (Dropdown: Helfer/Koordinator/Admin)
- Status (Aktiv/Inaktiv Toggle)
- Registriert am, Letzter Login

**Schichten-Tab:**
- Liste aller Schichten des Helfers (wie "Meine Schichten" aber für einen anderen User)
- Filter: Bevorstehend / Vergangen

**Aktionsbereich:**
- "Passwort-Reset-Email senden"
- "Account deaktivieren"
- "Account löschen" (mit Bestätigungsdialog + DSGVO-Hinweis)

### 4.3 Helfer einladen (`/admin/helfer/einladen`)

**Optionen:**
1. **Einzeleinladung**: Email + Vorname + Rolle → Einladungs-Email senden
2. **Massen-Einladung**: CSV-Upload mit Spalten: Email, Vorname, Nachname, Rolle
   - Vorschau der importierten Daten vor dem Absenden
   - Fehlerreport bei ungültigen Einträgen

---

## 5. API-Endpunkte

| Methode | Pfad | Beschreibung |
|---|---|---|
| `GET` | `/api/v1/admin/users` | Alle Nutzer (gefiltert) |
| `GET` | `/api/v1/admin/users/:id` | Nutzer-Details |
| `PUT` | `/api/v1/admin/users/:id` | Nutzer bearbeiten (Rolle, Status) |
| `DELETE` | `/api/v1/admin/users/:id` | Nutzer löschen (DSGVO) |
| `POST` | `/api/v1/admin/users/invite` | Einladung senden (einzeln) |
| `POST` | `/api/v1/admin/users/invite/bulk` | Massen-Einladung (CSV) |
| `GET` | `/api/v1/admin/users/:id/shifts` | Schichten eines Nutzers |
| `POST` | `/api/v1/admin/users/:id/approve` | Ausstehende Registrierung genehmigen |
| `POST` | `/api/v1/admin/users/:id/reject` | Ausstehende Registrierung ablehnen |
| `POST` | `/api/v1/admin/users/:id/reset-password` | Passwort-Reset-Email senden |
| `GET` | `/api/v1/admin/users/export` | CSV-Export |

---

## 6. Business Rules

- Ein Admin kann anderen Admins keine niedrigere Rolle zuweisen (kein self-demotion via UI).
- Das Löschen eines Nutzers anonymisiert seine historischen Schichtdaten (DSGVO): Name wird durch "Gelöschter Nutzer" ersetzt, Email wird entfernt.
- Deaktivierte Nutzer können sich nicht einloggen, ihre bestehenden Anmeldungen bleiben aber erhalten (nur Admin kann diese bereinigen).
- Einladungs-Token läuft nach 7 Tagen ab.
- Bei Massen-Einladung: bereits registrierte Emails werden übersprungen (kein Fehler, aber Hinweis im Report).

---

## 7. Akzeptanzkriterien

- [ ] Admin sieht alle Helfer mit Suchfunktion
- [ ] Rollenfilter funktioniert korrekt
- [ ] Einzeleinladung sendet Email mit Registrierungslink
- [ ] CSV-Einladung importiert Daten und versendet Emails
- [ ] Helfer-Detailansicht zeigt alle Schichten des Helfers
- [ ] Rolle kann geändert werden
- [ ] Account deaktivieren/aktivieren funktioniert
- [ ] Datenlöschung anonymisiert Schichtdaten korrekt
- [ ] Ausstehende Registrierungen können genehmigt/abgelehnt werden
