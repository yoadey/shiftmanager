# 00 — Allgemeine Spezifikation

**Stand:** 2026-09-24 · Quelle: `project/requirements_extracted.txt` (v1.5)

---

## 1. Projektziel

ShiftManager ist eine webbasierte Plattform zur digitalen Verwaltung von
Mitglieder-Helferschichten im Vereinsbetrieb (z. B. TSC Schwarz-Gelb Aachen).
Sie löst folgende Kernprobleme:

- Manuelle, fehleranfällige Schichtplanung per Tabelle
- Unklarer Überblick über belegte/freie Schichten
- Zeitaufwändige Abrechnung nicht geleisteter Stunden
- Kein zentraler Anlaufpunkt für Mitglieder

Lösung: digitale Selbstanmeldung für Schichten, Echtzeitanzeige freier und
belegter Schichten, automatische Stundenverfolgung und Abrechnung, sowie ein
Kiosk-Modus für den Vereinsheim-Computer.

---

## 2. Rollen

Fünf Rollen mit aufsteigender Berechtigung (`internal/domain/member.go`,
`RoleLevel`). Jede höhere Rolle hat mindestens die Rechte der niedrigeren:

| Rolle | Level | Beschreibung | Kernberechtigungen |
|---|---|---|---|
| **Kiosk-Nutzer** | – | Anonymer Zugang im Vereinsheim, kein Login | Mitglieder suchen (wenn aktiviert), Schichteintragung mit E-Mail-Verifizierung |
| **Mitglied** | 1 | Normales Vereinsmitglied | Eigenes Stundenkonto einsehen, sich für Schichten an-/abmelden, Kommentar hinterlegen, Erinnerungspräferenzen verwalten |
| **Veranstaltungsleiter** | 2 | Erstellt und verwaltet Veranstaltungen | Veranstaltungen/Schichten erstellen, bearbeiten, löschen; Teilnehmerlisten einsehen/exportieren; Anwesenheit pro Person bestätigen; beliebige Helfer (auch ohne Konto) einer Schicht zu-/abmelden |
| **Vorstand** | 3 | Vereinsführung mit erweitertem Datenzugriff | Alles wie Veranstaltungsleiter, zusätzlich: Jahresauswertung aller Mitglieder, individuelle Stundenkorrektur, Stundenziel/Abgeltungspreise konfigurieren, Datenschutzeinstellungen, manuelle Stundenbuchung, Mitgliederverwaltung, Abrechnung |
| **Admin** | 4 | Technische und organisatorische Gesamtverantwortung | Vollzugriff, OIDC-Konfiguration, Audit-Log-Einsicht, Datenbankexport |

Die Rollenprüfung erfolgt ausschließlich serverseitig
(`middleware.RequireRole`, `internal/adapter/http/router.go`); das Frontend
blendet nur passende Menüpunkte ein/aus (`useIsVorstand`, `useIsEventManager`
in `frontend/src/hooks/`), verlässt sich für die eigentliche Durchsetzung
aber auf die Backend-Antwort.

Navigation: Es gibt **keinen** manuellen Umschalter zwischen einer
"Mitglieder"- und einer "Vorstands"-Ansicht. Jede Rolle sieht eine einzige,
nach der echten Rolle zusammengesetzte Menüliste — die Mitglieder-Tabs plus,
für Veranstaltungsleiter aufwärts, zusätzliche Verwaltungs-Tabs (mit
Vorstand-only-Einträgen als weitere Erweiterung derselben Liste), visuell
durch eine Trennlinie/Sektionsüberschrift abgesetzt.

---

## 3. Kern-Entitäten (Datenmodell)

Siehe `internal/domain/*.go` für die vollständigen Go-Structs; `GORM
AutoMigrate` erzeugt daraus das Schema (SQLite in Tests, PostgreSQL in
Produktion via `migrations/*.sql`).

| Entität | Zweck |
|---|---|
| `Member` | Vereinsmitglied (Vorname, Nachname, E-Mail, Rolle, Eintritts-/Austrittsdatum, optionale OIDC-Verknüpfung, individuelles Stundenziel) |
| `Event` | Veranstaltung (Name, Beschreibung, Zeitraum, Ort, Kategorie, Status: `entwurf`/`veröffentlicht`/`abgeschlossen`/`abgesagt`, optionale Anhänge, optionale Wiederholungsregel) |
| `Shift` | Schicht innerhalb eines Event-Tages (Name, Zeitfenster, Min./Max.-Helfer, optionale Qualifikation) |
| `Registration` | Zu-/Absage eines Mitglieds oder Gasts (`GuestName`/`GuestEmail` bei Personen ohne Konto) zu einer Schicht; Status inkl. Reservierung, Bestätigung, Nichterscheinen |
| `HourEntry` | Gebuchte Stunden (aus Schichtbestätigung oder manueller Buchung), einem `ClubYear` zugeordnet |
| `ClubYear` | Abrechnungsjahr mit Zeitraum, Ziel-Stundenkonfiguration, optionalem Stundenübertrag ins Folgejahr |
| `FeeTier` | Abgeltungsbetrag pro Fehlstunden-Index (global oder pro Mitglied) |
| `EventAttachment` | Bild/PDF-Anhang an einer Veranstaltung |
| `AuditEntry` | Protokolleintrag zu jeder sicherheits-/datenschutzrelevanten Aktion |
| `EmailTemplate` / `EmailLogEntry` | Anpassbare Mail-Vorlagen und Versandprotokoll |
| `BrandingConfig` (+ Historie) | Vereinsname, Logo, Farben, Kiosk-Hintergrund |
| `AppSettings` | Systemweite Konfiguration (Namensmodus, Reservierungsdauer, Abmeldefrist, Kiosk-Sperre, …) |

Es gibt **keine** `Bereich`/`Area`-Entität — Schichten hängen direkt an
einem Event-Tag, nicht an einer zusätzlichen Organisationsebene. Ein
früherer Entwurf sah das vor; das wurde nie umgesetzt und ist gestrichen.

---

## 4. Architektur & Tech-Stack

Modularer Monolith: das Go-Backend bettet das kompilierte React-SPA zur
Buildzeit via `embed.FS` ein; es gibt keinen separaten Static-Server. Siehe
`CLAUDE.md` für die vollständige Verzeichnisstruktur.

| Schicht | Technologie |
|---|---|
| Backend | Go, Hexagonal/Clean Architecture (`domain` → `port` → `usecase` → `adapter`), REST-API |
| Router | chi |
| ORM | GORM (PostgreSQL produktiv, SQLite in-memory für Tests) |
| Migrationen | golang-migrate (nur PostgreSQL-Produktion; GORM AutoMigrate in Tests) |
| Frontend | React + TypeScript, Vite |
| State Management | Zustand (UI-Zustand) + TanStack Query (Server-Zustand) |
| Auth | OIDC (go-oidc v3) → internes JWT |
| E-Mail | SMTP (TLS/STARTTLS) |
| API-Doku | OpenAPI 3.0 (`api/openapi.yaml`), Codegen für Go-Typen (oapi-codegen) und TS-SDK (`@hey-api/openapi-ts`) |
| Cache | In-Memory, optional Redis |
| Deployment | Docker |

---

## 5. Seiten-Übersicht

Reale Tabs/Screens (`frontend/src/App.tsx`), nicht das frühere,
nie umgesetzte Seitenkonzept:

| Tab-Schlüssel | Screen | Sichtbar ab Rolle |
|---|---|---|
| `start` | Mitglieder-Dashboard | Mitglied |
| `entdecken` | Schichten entdecken (öffentlicher Zeitplan) | Mitglied |
| `schichten` | Meine Schichten | Mitglied |
| `profil` | Profil & Einstellungen | Mitglied |
| `uebersicht` | Verwaltungs-Dashboard | Veranstaltungsleiter |
| `events` | Veranstaltungsverwaltung | Veranstaltungsleiter |
| `mitglieder` | Mitgliederverwaltung | Vorstand |
| `abrechnungen` | Abrechnung | Vorstand |
| `settings` | Systemeinstellungen | Vorstand |

Zusätzlich: `kiosk` (eigenständige, nicht authentifizierte Route),
`EventDetail` (Drill-down aus mehreren Tabs), `MemberDetail`,
`ManualBooking`, `AuditLog`, `EmailTemplates`, `EmailLog` (aus den
Verwaltungs-Screens erreichbar).

---

## 6. Globale Anforderungen

- **Auth:** alle Seiten außer Login, OIDC-Callback, Kiosk und öffentlicher
  Schicht-Bestätigungslink erfordern ein gültiges JWT.
- **Responsive:** eigenständige Mobile- (`MobileShell`) und Desktop-Shell
  (`DesktopShell`), Umschaltung bei ≥ 900px Breite.
- **Fehlerbehandlung:** `ErrorBoundary` pro Screen, API-Fehler als
  Toast-Notification, 401 → Redirect zum Login.
- **E-Mail-Benachrichtigungen:** siehe [`06-benachrichtigungen.md`](06-benachrichtigungen.md).
- **Zeitzone:** Server speichert in UTC; Anzeige lokalisiert.
- **Audit-Log:** alle kritischen Aktionen protokolliert (T-012, DS-009).

---

## 7. API-Konventionen

- REST-API unter `/api/v1/`, JSON-Bodies.
- Authentifizierung: `Authorization: Bearer <JWT>`.
- Spezifikation: `api/openapi.yaml` — einzige Quelle der Wahrheit für
  generierte Backend-Typen und Frontend-SDK (siehe `CLAUDE.md` →
  "Code Generation"). Einige ältere Endpunkt-Familien (z. B. `/events/{id}/copy`,
  `/hours/club-years`, `/registrations/*`) sind aus historischen Gründen
  nicht in der OpenAPI-Spec dokumentiert; das ist ein bekannter, akzeptierter
  Rückstand und kein Bug.

---

## 8. Nicht-funktionale Anforderungen

| Kriterium | Ziel |
|---|---|
| Verfügbarkeit | 99,5 % (exkl. geplante Wartung) |
| Skalierung | bis 2.000 aktive Mitglieder, bis 500 gleichzeitige Verbindungen |
| Code-Coverage | ≥ 70 % (Unit-Tests) |
| Backup | täglich, 30 Tage Aufbewahrung, RTO 4 Stunden |

---

## 9. Out-of-Scope (Version 1.x)

- Buchhaltungsintegration (SEPA-Lastschrift, automatische Überweisung)
- Native Mobile-App (nur responsive Webapp)
- Chatmodul/Instant-Messaging
- Inventar-/Ressourcenverwaltung
- Vollständige Mitgliederverwaltung mit Satzungslogik (Beitragsverwaltung)
- Gamification/Punktesystem

---

## 10. Glossar

| Begriff | Definition |
|---|---|
| Helferschicht | Zeitlich definierter Aufgabenblock einer Veranstaltung |
| Stundenziel | Anzahl Stunden, die ein Mitglied im Vereinsjahr leisten muss |
| Bestätigte Stunden | Nach Schichtabschluss vom Veranstaltungsleiter individuell abgenommene Stunden |
| Reservierung | Schicht-Eintragung, noch nicht per E-Mail bestätigt oder abgenommen |
| Kiosk-Modus | Zugang ohne Anmeldung für einen fest installierten PC im Vereinsheim |
| Vereinsjahr / Club Year | Abrechnungsjahr für Stundenziele (Standard 01.01.–31.12., konfigurierbar) |
| Abgeltungsbetrag | Monetarisierter Ausgleich für nicht geleistete Pflichtstunden |
