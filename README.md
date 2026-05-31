# ShiftManager

ShiftManager ist eine webbasierte Plattform zur digitalen Verwaltung von Mitglieder-Helferschichten im Vereinsbetrieb, entwickelt für TSC Schwarz-Gelb Aachen. Das System ersetzt die manuelle, fehleranfällige Schichtplanung per Tabelle durch eine selbsterklärende Webanwendung: Mitglieder melden sich direkt für Schichten an, der Vereinsvorstand behält jederzeit den Überblick über Belegung und Stundenkonto, und am Jahresende erstellt das System automatisch die Abrechnung nicht geleisteter Stunden.

## Features

- **Mitgliederverwaltung** — Anlegen, Bearbeiten und Deaktivieren von Mitgliedern unabhängig vom OIDC-System; CSV-Import und -Export; automatische Verknüpfung mit OIDC-Konto per E-Mail-Abgleich
- **OIDC-Authentifizierung** — Keine Passwörter im System; Anbindung an beliebige OIDC-Provider (Keycloak, Auth0, Azure AD); internes JWT nach Login
- **Schicht-Anmeldung** — Selbstständige An- und Abmeldung für Schichten mit konfigurierter Frist; optionaler Kommentar zur Eintragung
- **Veranstaltungsverwaltung** — Mehrschichtige Veranstaltungen (auch mehrtägig); Status-Workflow Entwurf → Veröffentlicht → Abgeschlossen → Abgesagt
- **Interaktiver Schichtzeitplan** — Gantt-ähnliche Timeline; Farbcodierung nach Belegungsgrad (ausreichend / teilweise / kritisch unbesetzt)
- **Stundenverfolgung** — Dashboard mit getrennter Anzeige bestätigter Stunden und Stunden inkl. Reservierungen gegen das Jahres-Stundenziel
- **Stunden-Bestätigung** — Individuelle Bestätigung und Korrektur pro Person nach Schichtabschluss durch den Veranstaltungsleiter; manuelle Stundenbuchungen durch den Vorstand
- **Jahresabrechnung** — Konfigurierbare Abgeltungsbeträge pro Fehlstunde; automatische Berechnung am Jahresende; Export als PDF und CSV
- **Kiosk-Modus** — Eigenständige URL ohne Anmeldung für den Vereinsheim-PC; Touch-optimiertes UI; E-Mail-Identifikationsfluss mit Bestätigungslink; datenschutzkonforme Mitgliedersuche konfigurierbar
- **Automatische Benachrichtigungen** — E-Mail bei Anmeldung, Abmeldung, Stornierung, 1-Woche- und 24h-Erinnerung, Jahresabrechnung; alle Templates über die Admin-Oberfläche anpassbar
- **Vereinsbranding** — Primär- und Akzentfarbe, Logo-Upload (PNG/SVG), Vereinsname als Fallback; WCAG AA-Konformitätsprüfung
- **Datenschutzkonfiguration** — Namensanzeige systemweit umschaltbar (vollständig / abgekürzt); rollenabhängige Ausnahmen; DSGVO-konforme Auskunfts- und Löschrechte
- **Audit-Log** — Alle sicherheits- und datenschutzrelevanten Aktionen werden unveränderbar protokolliert
- **Rollenmodell** — Fünf Rollen mit abgestuften Rechten (siehe [Rollen](#rollen))
- **Responsive Design** — Mobile-First, WCAG 2.1 AA, aktuelle Browser (Chrome, Firefox, Safari, Edge)

## Architecture

ShiftManager folgt dem Ansatz des **modularen Monolithen**: Das Go-Backend und das React-Frontend werden als ein einziges Deployment-Artefakt ausgeliefert. Die statischen Frontend-Assets werden zur Build-Zeit via `embed.FS` in das Go-Binary eingebettet — es gibt keinen separaten statischen File-Server oder CDN-Eintrag.

Das Backend ist nach **Clean Architecture / Hexagonal Architecture** strukturiert:

```
cmd/server/         — Einstiegspunkt, Dependency Wiring
internal/
  domain/           — Entities, Repository-Interfaces, Domain Services
  usecase/          — Anwendungsfälle (Business Logic)
  port/             — HTTP-Handler (chi), Request/Response-DTOs
  adapter/          — Datenbankadapter (pgx), OIDC-Adapter, Mail-Adapter
  infrastructure/   — Logger, Scheduler, Static-File-Embed
migrations/         — golang-migrate SQL-Migrationsdateien
frontend/           — React-SPA (TypeScript, Vite, TanStack Query, Zustand)
```

Alle Kubernetes-Manifeste liegen unter `deploy/kubernetes/`. Die CI/CD-Pipeline baut ein einziges Docker-Image (`ghcr.io/yoadey/shiftmanager`) und deployt es auf einem k3s-Cluster via GitHub Actions.

Die vollständigen funktionalen und nicht-funktionalen Anforderungen finden sich in [`project/requirements_extracted.txt`](project/requirements_extracted.txt).

## Tech Stack

| Schicht | Technologie |
|---|---|
| **Backend** | Go 1.22, [chi](https://github.com/go-chi/chi) (HTTP Router), [pgx v5](https://github.com/jackc/pgx) (PostgreSQL), [go-oidc v3](https://github.com/coreos/go-oidc), [golang-jwt](https://github.com/golang-jwt/jwt), [zerolog](https://github.com/rs/zerolog) |
| **Frontend** | React 18, Vite, TypeScript, [TanStack Query](https://tanstack.com/query/latest), [Zustand](https://zustand-demo.pmnd.rs/), React Router 6, Axios |
| **Datenbank** | PostgreSQL 16 (primär), Redis 7 (optional, Sessions & Caching) |
| **Authentifizierung** | OpenID Connect via go-oidc (Keycloak, Auth0, Azure AD, ...) |
| **Deployment** | Docker, Docker Compose, Kubernetes/k3s, Traefik, cert-manager (Let's Encrypt) |
| **CI/CD** | GitHub Actions, ghcr.io (GitHub Container Registry) |

## Quick Start (Docker Compose)

```bash
# 1. Konfiguration erstellen
cp .env.example .env
# .env bearbeiten und mind. DATABASE_URL, OIDC_* und JWT_SECRET anpassen

# 2. Datenbankmigrationen einmalig ausfuehren
docker compose --profile migrate up migrate

# 3. Stack starten (App + PostgreSQL)
docker compose up

# Mit Redis:
docker compose --profile full up
```

Die Anwendung ist danach unter **http://localhost:8080** erreichbar.

## Development Setup

**Voraussetzungen:** Go 1.22+, Node 20+, PostgreSQL 16, (optional) Redis 7

```bash
# Repository klonen
git clone https://github.com/yoadey/shiftmanager.git
cd shiftmanager

# ---- Backend ----
cp .env.example .env
# .env bearbeiten
source .env  # oder: export $(cat .env | xargs)

go mod download

# Migrationen ausfuehren
migrate -path migrations -database "$DATABASE_URL" up

# Backend starten
go run ./cmd/server
# Laeuft auf http://localhost:8080

# ---- Frontend (separates Terminal) ----
cd frontend
npm install
npm run dev
# Dev-Server auf http://localhost:5173
# /api/* wird automatisch zu localhost:8080 proxied (vite.config.ts)
```

Im Entwicklungsmodus werden Frontend-Aenderungen mit Hot Module Replacement sofort angezeigt. Die Go-API muss separat gestartet (und bei Aenderungen neu kompiliert) werden — `air` empfiehlt sich als Live-Reload-Tool.

## Database Migrations

Migrationen werden mit [golang-migrate](https://github.com/golang-migrate/migrate) verwaltet. Die SQL-Dateien liegen in `migrations/`.

```bash
# Alle ausstehenden Migrationen ausfuehren
migrate -path migrations -database "$DATABASE_URL" up

# Eine Migration zurueckrollen
migrate -path migrations -database "$DATABASE_URL" down 1

# Aktuelle Version pruefen
migrate -path migrations -database "$DATABASE_URL" version

# Neue Migration erstellen
migrate create -ext sql -dir migrations -seq <name>
```

In Kubernetes uebernimmt ein Init-Container (in `deploy/kubernetes/deployment.yaml`) die Migrationen automatisch vor jedem App-Start.

## Configuration

Alle Konfigurationsparameter werden ueber Umgebungsvariablen gesteuert. Kopiere `.env.example` nach `.env` und passe die Werte an.

| Variable | Beschreibung | Standard |
|---|---|---|
| `PORT` | HTTP-Port des Servers | `8080` |
| `DATABASE_URL` | PostgreSQL Connection String | _(erforderlich)_ |
| `REDIS_URL` | Redis Connection String | `""` (in-memory Fallback) |
| `OIDC_ISSUER` | Issuer-URL des OIDC-Providers | _(erforderlich)_ |
| `OIDC_CLIENT_ID` | OIDC Client ID | _(erforderlich)_ |
| `OIDC_CLIENT_SECRET` | OIDC Client Secret | _(erforderlich)_ |
| `OIDC_REDIRECT_URL` | Backend-Callback-URL; muss beim IdP **verbatim** als Redirect-URI registriert sein (inkl. `/api/v1`-Prefix), z. B. `http://localhost:8080/api/v1/auth/callback` | _(erforderlich)_ |
| `LOGIN_REDIRECT_URL` | SPA-Route, auf die der Callback nach erfolgreichem Login weiterleitet (Token im URL-Fragment) | `/auth/callback` |
| `BOOTSTRAP_ADMIN_EMAIL` | Mitglied mit dieser E-Mail wird beim Login automatisch angelegt/verknuepft und als aktiver Administrator freigeschaltet (Erst-Admin-Bootstrap) | `""` |
| `JWT_SECRET` | Signierungsgeheimnis fuer interne JWTs (min. 32 Zeichen) | _(erforderlich)_ |
| `SMTP_HOST` | SMTP-Servername | _(erforderlich fuer E-Mail)_ |
| `SMTP_PORT` | SMTP-Port (587 fuer STARTTLS) | `587` |
| `SMTP_USER` | SMTP-Benutzername | _(erforderlich fuer E-Mail)_ |
| `SMTP_PASS` | SMTP-Passwort | _(erforderlich fuer E-Mail)_ |
| `SMTP_FROM` | Absenderadresse mit Displayname | _(erforderlich fuer E-Mail)_ |
| `BASE_URL` | Oeffentliche URL der Anwendung (fuer Links in E-Mails) | `http://localhost:8080` |
| `LOG_LEVEL` | Loglevel: `trace` `debug` `info` `warn` `error` | `info` |
| `GOMEMLIMIT` | Go Runtime Memory Limit | `200MiB` |
| `CLUB_NAME` | Vereinsname als Fallback vor DB-Branding | `TSC Schwarz-Gelb Aachen` |

### Erster Login & Mitglieder-Onboarding

Die Anmeldung laeuft ausschliesslich ueber OIDC. Der Ablauf:

1. **OIDC-Redirect korrekt setzen:** `OIDC_REDIRECT_URL` muss auf den Backend-Callback `…/api/v1/auth/callback` zeigen und exakt so beim IdP registriert sein. Nach erfolgreichem Code-Tausch leitet das Backend mit dem JWT im URL-Fragment auf `LOGIN_REDIRECT_URL` (Standard `/auth/callback`) weiter; die SPA speichert das Token und meldet den Nutzer an.
2. **Ersten Admin anlegen (Bootstrap):** `BOOTSTRAP_ADMIN_EMAIL` auf die E-Mail des Vorstands setzen. Beim ersten Login mit dieser Identitaet wird das Mitglied automatisch angelegt (bzw. ein bestehendes verknuepft), als `admin` gesetzt und sofort freigeschaltet. Danach kann der Wert wieder entfernt werden.
3. **Weitere Mitglieder:** Unbekannte OIDC-Nutzer werden beim Login automatisch als Mitglied **registriert, aber nicht freigeschaltet** (`is_active = false`). Sie sehen einen Hinweis „Warten auf Freischaltung“. Ein Administrator/Vorstand aktiviert sie anschliessend in der Mitgliederverwaltung. Alternativ koennen Mitglieder vorab per CSV-Import oder manuell angelegt werden; die OIDC-Verknuepfung erfolgt dann beim ersten Login automatisch per E-Mail-Abgleich.

## Deployment (Kubernetes)

### Helm (empfohlen)

Das vollständige Helm-Chart liegt unter [`deploy/helm/shiftmanager/`](deploy/helm/shiftmanager/)
und bringt optional PostgreSQL und Redis als Subcharts mit (Bitnami), inkl.
Migrations-Init-Container, Uploads-PVC, Ingress, HPA, PDB, NetworkPolicy, optionalem
`pg_dump`-Backup-CronJob und `helm test`.

```bash
helm dependency update ./deploy/helm/shiftmanager

helm upgrade --install shiftmanager ./deploy/helm/shiftmanager \
  -n shiftmanager --create-namespace \
  --set app.jwtSecret="$(openssl rand -hex 32)" \
  --set postgresql.auth.password="$(openssl rand -hex 16)" \
  --set app.bootstrapAdminEmail="vorstand@mein-verein.de" \
  --set app.baseUrl="https://schichten.mein-verein.de" \
  --set app.oidc.issuer="https://id.mein-verein.de/realms/verein" \
  --set app.oidc.clientId="shiftmanager" \
  --set app.oidc.clientSecret="••••" \
  --set app.oidc.redirectUrl="https://schichten.mein-verein.de/api/v1/auth/callback"
```

Externe DB/Redis statt Subcharts: `--set postgresql.enabled=false --set database.external.url=…`
bzw. `--set redis.enabled=false`. Details und ein Produktionsbeispiel siehe
[`deploy/helm/shiftmanager/README.md`](deploy/helm/shiftmanager/README.md).

### Rohe Manifeste (Alternative)

Die Kubernetes-Manifeste unter `deploy/kubernetes/` sind fuer k3s mit Traefik und cert-manager ausgelegt.

### Voraussetzungen

- k3s-Cluster mit Traefik (standardmaessig enthalten)
- cert-manager mit einem `ClusterIssuer` namens `letsencrypt-prod`
- `kubectl` mit Zugriff auf den Cluster

### Manuelles Deployment

```bash
# 1. Secrets konfigurieren
#    NIEMALS secret.yaml mit echten Werten committen!
#    Empfohlen: Sealed Secrets oder External Secrets Operator
#    Alternativ: Werte direkt mit kubectl setzen
kubectl create namespace shiftmanager
kubectl create secret generic shiftmanager-secret \
  --namespace shiftmanager \
  --from-literal=DATABASE_URL="postgres://..." \
  --from-literal=JWT_SECRET="..." \
  # ... weitere Keys

# 2. Domain eintragen
#    In deploy/kubernetes/ingress.yaml "shiftmanager.example.com" ersetzen.
#    In deploy/kubernetes/configmap.yaml BASE_URL anpassen.

# 3. Alle Manifeste anwenden (ohne secret.yaml wenn manuell erstellt)
kubectl apply -f deploy/kubernetes/namespace.yaml
kubectl apply -f deploy/kubernetes/configmap.yaml
kubectl apply -f deploy/kubernetes/postgres-statefulset.yaml
kubectl apply -f deploy/kubernetes/deployment.yaml
kubectl apply -f deploy/kubernetes/service.yaml
kubectl apply -f deploy/kubernetes/ingress.yaml
kubectl apply -f deploy/kubernetes/hpa.yaml
kubectl apply -f deploy/kubernetes/pdb.yaml
kubectl apply -f deploy/kubernetes/networkpolicy.yaml
kubectl apply -f deploy/kubernetes/cronjob-backup.yaml

# Optional: Redis
kubectl apply -f deploy/kubernetes/redis-deployment.yaml
```

### GitHub Actions (automatisch)

Das Workflow-Paar fuehrt folgendes aus:

1. **CI** (`.github/workflows/ci.yml`) — bei jedem Push und PR: Backend-Lint (golangci-lint + go vet), Backend-Tests mit Race-Detector, Frontend-Build (faengt TypeScript-Fehler ab). Bei Push auf `main`: Docker-Image bauen und nach `ghcr.io/yoadey/shiftmanager` pushen.

2. **Deploy** (`.github/workflows/deploy.yml`) — bei Push auf `main` nach erfolgreichem CI: Image-Tag im Deployment-Manifest aktualisieren, alle Manifeste per `kubectl apply` einfahren, Rollout abwarten.

Benoetigter Repository-Secret: `KUBECONFIG` (base64-kodierte kubeconfig).

### Datenbankbackup

Ein taeglicher CronJob (`cronjob-backup.yaml`) sichert die Datenbank per `pg_dump` auf ein PVC. Fuer die Ablage in S3 oder einem anderen Object-Store ist ein Stub mit Anleitung im Manifest enthalten.

## Roles

| Rolle | Beschreibung | Kernrechte |
|---|---|---|
| **admin** | Technische und organisatorische Gesamtverantwortung | Vollzugriff; Benutzerverwaltung; Systemkonfiguration; Datenbankexport; Audit-Log; OIDC-Konfiguration |
| **vorstand** | Vereinsfuehrung mit erweitertem Datenzugriff | Alles wie Veranstaltungsleiter; Jahresauswertung aller Mitglieder; individuelle Stundenkorrektur; Stundenziel- und Abgeltungspreiskonfiguration; Datenschutzeinstellungen; manuelle Stundenbuchung |
| **veranstaltungsleiter** | Erstellt und verwaltet Veranstaltungen | Veranstaltungen und Schichten anlegen/bearbeiten/loeschen; Teilnehmerlisten einsehen und exportieren; individuelle Anwesenheitsbestaetigung |
| **mitglied** | Normales Vereinsmitglied | Eigenes Stundenkonto einsehen; Schichten an-/abmelden; Kommentar hinterlegen; Erinnerungspraeferenzen verwalten |
| **kiosk** | Anonymer Zugang im Vereinsheim | Mitglieder suchen (wenn aktiviert); Schichteintragung mit E-Mail-Verifizierung; keine Anmeldung erforderlich |

## Design Prototype

Das Verzeichnis `project/` enthaelt den interaktiven HTML-Prototypen, der mit Claude Design erstellt wurde. Oeffne `project/index.html` im Browser, um die gesamte UI zu erkunden — ohne laufendes Backend. Der Prototyp dient als Design-Referenz fuer die Frontend-Implementierung.

## License

Proprietaer — TSC Schwarz-Gelb Aachen e.V. Alle Rechte vorbehalten.
