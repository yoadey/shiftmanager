# OpenSpec: Shiftmanager – Allgemeine Spezifikation

**Version:** 1.0  
**Stand:** 2026-06-04  
**Status:** Draft

---

## 1. Projektziel

Der **Shiftmanager** ist eine Webanwendung zur Verwaltung von Helferschichten bei Veranstaltungen (z. B. Festivals, Messen, Charity-Events). Das System ermöglicht es Organisatoren, Schichten anzulegen und zu koordinieren, sowie Helfern, sich für Schichten anzumelden und ihren persönlichen Einsatzplan einzusehen.

---

## 2. Zielgruppen (Rollen)

| Rolle | Beschreibung |
|---|---|
| **Helfer** | Freiwillige, die sich für Schichten anmelden. Benötigen Login. |
| **Koordinator** | Verwaltet Schichten und Helfer für einen bestimmten Bereich/Station. |
| **Admin** | Hat vollen Zugriff auf alle Funktionen, Veranstaltungen und Einstellungen. |
| **Gast** | Kann öffentlichen Schichtenplan einsehen (optional, konfigurierbar). |

---

## 3. Kern-Entitäten (Datenmodell)

### Veranstaltung (`Event`)
```
id, name, beschreibung, start_datum, end_datum, ort, status (draft|active|archived)
```

### Bereich (`Area`)
```
id, event_id, name, beschreibung, farbe, koordinator_id
```

### Schicht (`Shift`)
```
id, area_id, event_id, name, beschreibung, start_time, end_time,
min_helfer, max_helfer, status (open|full|closed|cancelled)
```

### Helfer (`User`)
```
id, email, vorname, nachname, telefon, rolle (helper|coordinator|admin),
registriert_am, letzter_login, aktiv
```

### Schicht-Anmeldung (`ShiftRegistration`)
```
id, shift_id, user_id, angemeldet_am, status (registered|confirmed|cancelled|no-show),
notiz, bestätigt_von (user_id)
```

---

## 4. Tech-Stack (Empfehlung)

| Schicht | Technologie |
|---|---|
| Frontend | React + TypeScript + Vite |
| UI-Bibliothek | Tailwind CSS + shadcn/ui |
| State Management | Zustand oder React Query |
| Backend | Node.js + Express / Fastify **oder** Next.js (Full-Stack) |
| Datenbank | PostgreSQL |
| Auth | JWT + HTTP-only Cookies oder NextAuth |
| Email | Nodemailer / Resend |
| Deployment | Docker + Docker Compose |

---

## 5. Seiten-Übersicht

| # | Seite | Pfad | Rollen |
|---|---|---|---|
| 01 | Login / Registrierung | `/login`, `/register`, `/reset-password` | Alle |
| 02 | Dashboard | `/dashboard` | Helfer, Koordinator, Admin |
| 03 | Schichtenplan | `/schichten` | Alle (Gast read-only) |
| 04 | Meine Schichten | `/meine-schichten` | Helfer |
| 05 | Schicht-Verwaltung (Admin) | `/admin/schichten` | Koordinator, Admin |
| 06 | Helfer-Verwaltung (Admin) | `/admin/helfer` | Admin |
| 07 | Bereiche-Verwaltung | `/admin/bereiche` | Admin |
| 08 | Veranstaltungs-Verwaltung | `/admin/veranstaltungen` | Admin |
| 09 | Berichte & Export | `/admin/berichte` | Koordinator, Admin |
| 10 | Profil & Einstellungen | `/profil` | Helfer, Koordinator, Admin |

---

## 6. Globale Anforderungen

### 6.1 Authentifizierung & Autorisierung
- Alle Seiten außer Login, Registrierung und optionalem öffentlichem Schichtenplan erfordern Login.
- Rollenbasierter Zugriff: Frontend-Routen und API-Endpunkte prüfen die Rolle.
- Sessions laufen nach 8 Stunden ab (konfigurierbar). "Angemeldet bleiben" verlängert auf 30 Tage.

### 6.2 Responsive Design
- Mobile-first: alle Seiten müssen auf Smartphones (≥ 375px) vollständig nutzbar sein.
- Tablet (≥ 768px) und Desktop (≥ 1024px) sind ebenfalls vollständig unterstützt.

### 6.3 Mehrsprachigkeit (i18n)
- Initiale Sprache: Deutsch (de-DE).
- Architektur muss i18n von Anfang an unterstützen (z. B. `react-i18next`), damit weitere Sprachen ergänzt werden können.

### 6.4 Barrierefreiheit (a11y)
- WCAG 2.1 AA als Mindestziel.
- Korrekte ARIA-Labels, Tastaturnavigation, ausreichende Kontraste.

### 6.5 Fehlerbehandlung
- Alle Formulare zeigen Inline-Validierungsfehler.
- API-Fehler werden als Toast-Notifications angezeigt.
- Bei 401 wird automatisch zum Login umgeleitet.

### 6.6 Email-Benachrichtigungen
- Anmeldebestätigung nach Schicht-Anmeldung.
- Erinnerungs-Email 24 Stunden vor Schichtbeginn.
- Benachrichtigung bei Schichtabsage durch Admin.

### 6.7 Zeitzone
- Alle Zeiten werden in der Zeitzone der Veranstaltung gespeichert und angezeigt.
- Server speichert Zeiten in UTC, Anzeige erfolgt lokalisiert.

---

## 7. API-Konventionen

- REST-API unter `/api/v1/`
- JSON für Request/Response-Bodies
- HTTP-Statuscodes semantisch korrekt (200, 201, 400, 401, 403, 404, 422, 500)
- Fehlermeldungen: `{ "error": "code", "message": "Beschreibung", "details": {} }`
- Pagination: `?page=1&limit=20` mit Response `{ data: [], total, page, limit }`
- Authentifizierung: Bearer Token im Authorization-Header oder HTTP-only Cookie

---

## 8. Nicht-funktionale Anforderungen

| Kriterium | Ziel |
|---|---|
| Ladezeit (initial) | < 3s auf 3G |
| API-Antwortzeit | < 200ms p95 |
| Verfügbarkeit | ≥ 99% |
| Datenschutz | DSGVO-konform (EU), Datenlöschung auf Anfrage |
| Sicherheit | OWASP Top 10, HTTPS only, CSRF-Schutz |
