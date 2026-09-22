# OpenSpec: Seite 10 – Profil & Einstellungen

**Version:** 1.0  
**Stand:** 2026-06-04  
**Status:** Draft

---

## 1. Zweck

Persönliche Profilseite für alle Nutzer. Verwaltung eigener Accountdaten, Passwortänderung und Benachrichtigungs-Einstellungen.

---

## 2. Routen

| Route | Beschreibung |
|---|---|
| `/profil` | Persönliches Profil |
| `/profil/sicherheit` | Passwort & Sicherheit |
| `/profil/benachrichtigungen` | Benachrichtigungs-Einstellungen |

---

## 3. User Stories

| ID | Als... | möchte ich... | damit... |
|---|---|---|---|
| PROF-01 | Helfer | meine persönlichen Daten einsehen und ändern | ich meine Kontaktdaten aktuell halte |
| PROF-02 | Helfer | mein Passwort ändern | ich mein Konto absichere |
| PROF-03 | Helfer | Email-Benachrichtigungen konfigurieren | ich nur relevante Emails erhalte |
| PROF-04 | Helfer | meinen Account löschen | ich das System bei Bedarf vollständig verlassen kann |

---

## 4. UI-Anforderungen

### 4.1 Profil-Tab (`/profil`)

**Formular-Felder:**
- Vorname (Pflicht)
- Nachname (Pflicht)
- Email (Pflicht; Änderung erfordert Email-Verifikation)
- Telefonnummer (optional)
- Profilbild (optional, Upload, max. 2 MB, JPEG/PNG)

**Nur-Lesen:**
- Rolle (z. B. "Helfer") – wird durch Admin gesetzt
- Mitglied seit (Datum)
- Letzter Login (Datum/Uhrzeit)

**Aktionen:**
- Speichern
- Account löschen (gefahrenreiche Aktion, rot markiert, mit Bestätigungsdialog: "Bitte Passwort eingeben um zu bestätigen")

### 4.2 Sicherheits-Tab (`/profil/sicherheit`)

**Passwort ändern:**
- Aktuelles Passwort (Pflicht)
- Neues Passwort (Pflicht, mit Stärke-Anzeige)
- Passwort bestätigen (Pflicht)

**Aktive Sessions (optional, nice-to-have):**
- Liste aktiver Sessions: Gerät, Browser, Datum, IP-Region
- Button "Alle anderen Sessions beenden"

### 4.3 Benachrichtigungs-Tab (`/profil/benachrichtigungen`)

**Email-Benachrichtigungen (Toggle je Typ):**
- Anmeldebestätigung bei neuer Schicht-Anmeldung (Standard: An)
- Erinnerung 24h vor Schichtbeginn (Standard: An)
- Erinnerung 2h vor Schichtbeginn (Standard: Aus)
- Schicht-Absage durch Admin (Standard: An, nicht deaktivierbar)
- Schicht-Bestätigung durch Koordinator (Standard: An)
- Wöchentliche Zusammenfassung meiner Schichten (Standard: Aus)

---

## 5. API-Endpunkte

| Methode | Pfad | Beschreibung |
|---|---|---|
| `GET` | `/api/v1/users/me` | Eigenes Profil laden |
| `PUT` | `/api/v1/users/me` | Profil aktualisieren |
| `PUT` | `/api/v1/users/me/password` | Passwort ändern |
| `DELETE` | `/api/v1/users/me` | Account löschen (DSGVO) |
| `GET` | `/api/v1/users/me/notifications` | Benachrichtigungs-Einstellungen |
| `PUT` | `/api/v1/users/me/notifications` | Benachrichtigungs-Einstellungen speichern |
| `POST` | `/api/v1/users/me/avatar` | Profilbild hochladen |
| `DELETE` | `/api/v1/users/me/sessions` | Alle anderen Sessions beenden |

---

## 6. Business Rules

- Email-Änderung: neue Email muss verifiziert werden bevor sie aktiv wird; alte Email bleibt bis Bestätigung gültig.
- Account-Löschung anonymisiert alle persönlichen Daten gemäß DSGVO (wie in Seite 06 beschrieben).
- Benachrichtigungs-Typ "Schicht-Absage durch Admin" kann nicht deaktiviert werden (wichtige Information).
- Passwort-Änderung invalidiert alle anderen Sessions (außer die aktuelle).

---

## 7. Akzeptanzkriterien

- [ ] Profildaten können bearbeitet und gespeichert werden
- [ ] Email-Änderung triggert Verifizierungs-Email an neue Adresse
- [ ] Passwort kann geändert werden; andere Sessions werden invalidiert
- [ ] Benachrichtigungs-Einstellungen werden korrekt gespeichert und wirksam
- [ ] Account-Löschung mit Passwort-Bestätigung anonymisiert Daten
- [ ] Profilbild-Upload funktioniert mit Größen- und Formatvalidierung
