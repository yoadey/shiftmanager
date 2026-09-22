# OpenSpec: Seite 01 – Login & Registrierung

**Version:** 1.0  
**Stand:** 2026-06-04  
**Status:** Draft

---

## 1. Zweck

Authentifizierung der Nutzer (Login, Logout) sowie Onboarding neuer Helfer (Registrierung, Email-Verifizierung, Passwort-Reset).

---

## 2. Seiten & Routen

| Route | Beschreibung |
|---|---|
| `/login` | Login-Formular |
| `/register` | Registrierungsformular (nur wenn Selbstregistrierung aktiv) |
| `/verify-email?token=...` | Email-Adresse bestätigen |
| `/reset-password` | Passwort-Reset anfordern |
| `/reset-password/confirm?token=...` | Neues Passwort setzen |

---

## 3. User Stories

| ID | Als... | möchte ich... | damit... |
|---|---|---|---|
| AUTH-01 | Helfer | mich mit Email + Passwort einloggen | ich Zugang zu meinem Konto habe |
| AUTH-02 | Helfer | mich registrieren | ich ein neues Konto erstellen kann |
| AUTH-03 | Helfer | mein Passwort zurücksetzen | ich bei Verlust wieder Zugang bekomme |
| AUTH-04 | Helfer | angemeldet bleiben können | ich nicht bei jedem Besuch erneut einloggen muss |
| AUTH-05 | Admin | mich ausloggen | meine Session sicher beendet wird |
| AUTH-06 | Admin | Selbstregistrierung deaktivieren | nur eingeladene Helfer Zugang erhalten |

---

## 4. UI-Anforderungen

### 4.1 Login-Seite (`/login`)
- Logo/Titel der Anwendung prominent oben
- Formular: Email-Feld, Passwort-Feld (mit Sichtbarkeits-Toggle), Absenden-Button
- Checkbox "Angemeldet bleiben" (30-Tage-Session)
- Link "Passwort vergessen?" → `/reset-password`
- Link "Noch kein Konto? Registrieren" → `/register` (nur wenn Selbstregistrierung aktiv)
- Bei Fehler: Inline-Fehlermeldung (nicht "Email oder Passwort falsch" einzeln, sondern generisch zum Schutz vor User-Enumeration)

### 4.2 Registrierungs-Seite (`/register`)
- Felder: Vorname, Nachname, Email, Passwort, Passwort bestätigen
- Passwort-Stärke-Anzeige
- Checkbox: "Ich akzeptiere die Datenschutzerklärung" (Pflichtfeld)
- Nach Absenden: Hinweis "Bestätigungs-Email wurde gesendet"
- Link "Bereits registriert? Anmelden" → `/login`

### 4.3 Passwort-Reset-Seite (`/reset-password`)
- Nur ein Feld: Email
- Hinweis: "Falls ein Konto mit dieser Email existiert, erhalten Sie eine Email."
- Keine Unterscheidung ob Email existiert (Schutz vor Enumeration)

### 4.4 Neues Passwort setzen (`/reset-password/confirm`)
- Felder: Neues Passwort, Passwort bestätigen
- Token-Validierung im Hintergrund; bei ungültigem/abgelaufenem Token → Fehlermeldung + Link zu `/reset-password`

---

## 5. Validierungsregeln

| Feld | Regeln |
|---|---|
| Email | Gültige Email-Adresse, max. 255 Zeichen |
| Passwort (neu) | Min. 8 Zeichen, mindestens 1 Großbuchstabe, 1 Zahl |
| Vorname / Nachname | Pflichtfeld, 1–100 Zeichen, keine Sonderzeichen außer `-` und ` ` |

---

## 6. API-Endpunkte

| Methode | Pfad | Beschreibung |
|---|---|---|
| `POST` | `/api/v1/auth/login` | Login, gibt Token zurück |
| `POST` | `/api/v1/auth/logout` | Session invalidieren |
| `POST` | `/api/v1/auth/register` | Neuen Helfer registrieren |
| `POST` | `/api/v1/auth/verify-email` | Email-Token bestätigen |
| `POST` | `/api/v1/auth/reset-password` | Reset-Email anfordern |
| `POST` | `/api/v1/auth/reset-password/confirm` | Neues Passwort setzen |
| `GET` | `/api/v1/auth/me` | Aktuellen User laden |

---

## 7. Business Rules

- Nach 5 fehlgeschlagenen Logins innerhalb 10 Minuten: Account für 15 Minuten sperren (Rate Limiting).
- Email-Verifizierungs-Token läuft nach 24 Stunden ab.
- Passwort-Reset-Token läuft nach 1 Stunde ab und ist Einmal-Use.
- Passwörter werden mit `bcrypt` (cost ≥ 12) gespeichert.
- Bei Selbstregistrierung muss Admin neue Helfer ggf. freischalten (konfigurierbar: auto-approve oder manual-approve).

---

## 8. Akzeptanzkriterien

- [ ] Login mit gültigen Daten → Redirect zu `/dashboard`
- [ ] Login mit falschen Daten → generische Fehlermeldung, kein Hinweis welches Feld falsch
- [ ] Registrierung → Bestätigungs-Email wird versendet
- [ ] Email-Verifikations-Link bestätigt Email und loggt User ein
- [ ] Passwort-Reset-Flow komplett funktionsfähig (End-to-End)
- [ ] "Angemeldet bleiben" Checkbox funktioniert
- [ ] Geschützte Routen leiten unauthentifizierte Nutzer zu `/login` um (mit `?next=` Parameter)
- [ ] Nach Login mit `?next=` Parameter → Redirect zur ursprünglichen Seite
