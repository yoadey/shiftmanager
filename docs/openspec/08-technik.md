# 08 — Technische Anforderungen

Umgesetzt. Anforderungs-IDs: `T-001`–`T-013`, `F-001`–`F-010`.

## Backend (Go)

| ID | Anforderung | Umsetzung |
|---|---|---|
| T-001 | RESTful JSON-API unter `/api/v1` | `internal/adapter/http/router.go` |
| T-002 | Vollständige OpenAPI-3.0-Dokumentation (generiert) | `api/openapi.yaml`, oapi-codegen |
| T-003 | API-Auth über Bearer-JWT | `middleware.JWTAuth` |
| T-004 | CORS konfigurierbar (Origin-Whitelist) | `internal/config` |
| T-005 | Rate-Limiting pro IP für öffentliche Endpunkte (inkl. Kiosk) | `internal/adapter/http/middleware/ratelimit.go` (Token-Bucket pro IP), deaktiviert in `TEST_MODE` |
| T-006 | Typsicherer ORM/Query-Builder | GORM |
| T-007 | Migrationstool für Datenbankmigrationen | golang-migrate (`migrations/*.sql`, nur PostgreSQL-Produktion) |
| T-008 | Konfiguration über Umgebungsvariablen/.env | `internal/config` |
| T-009 | Strukturiertes Logging (JSON), konfigurierbare Log-Levels | `internal/infrastructure/logger` (zerolog) |
| T-010 | Health-Check-Endpunkte | `/health`, `/readyz` |
| T-011 | Hintergrundtasks als eigene Goroutinen (E-Mail-Versand, Reservierungsablauf, Jahresabrechnung) | `internal/infrastructure/scheduler` |
| T-012 | Audit-Log aller kritischen Aktionen | siehe `05-dashboard-berichte.md` |
| T-013 | Speicherort für hochgeladene Medien (Logo, Event-Anhänge) statt lokalem Verzeichnis/PVC auch auf S3-kompatiblen Object-Storage konfigurierbar, ohne Migrationsscript für Bestandsdateien | `port.MediaStorage`-Interface (`Put`/`Delete`), zwei Adapter: `internal/adapter/storage/local` (Default, unverändertes Verhalten) und `internal/adapter/storage/s3` (AWS SDK v2, kompatibel zu jedem S3-API-Objektspeicher über `S3_ENDPOINT`/`S3_FORCE_PATH_STYLE`, z. B. MinIO/Hetzner). Gewählt über `MEDIA_STORAGE=local\|s3`; `upload.go`, `event_handler.go`, `settings_handler.go` gehen ausschließlich über das Interface. `/uploads/*` (router.go) bleibt nur für `local` gemountet. **Betriebshinweis:** wie beim lokalen `/uploads/*`-Pfad sind Logo und Event-Anhänge als öffentlich, unter einer nicht erratbaren URL, ohne Auth-Prüfung gedacht — der `s3`-Adapter setzt bewusst **kein** Objekt-ACL beim Upload (viele Buckets haben ACLs unter "bucket owner enforced" deaktiviert, dort würde ein ACL-Aufruf schlicht fehlschlagen). Stattdessen muss der **Bucket selbst** (per Bucket-Policy oder vorgeschaltetem öffentlichem CDN) für Lesezugriff freigegeben sein, bevor `MEDIA_STORAGE=s3` produktiv genutzt wird — sonst liefern alle Logo-/Anhang-URLs einen 403, obwohl der Upload selbst erfolgreich war. |

## Frontend (React)

| ID | Anforderung | Umsetzung |
|---|---|---|
| F-001 | SPA mit React 18+ und TypeScript | `frontend/` (Vite) |
| F-002 | State-Management: Zustand/React Query | Zustand (UI-State) + TanStack Query (Server-State) |
| F-003 | Vollständig responsiv, Mobile-First | `MobileShell.tsx`/`DesktopShell.tsx`, Breakpoint ≥ 900px |
| F-004 | Kiosk als eigenständige Route mit vereinfachtem Touch-UI | `/kiosk` |
| F-005 | Interaktiver, Gantt-ähnlicher Zeitplan für Schichten | `EventDetail.tsx` Timeline |
| F-006 | Mehrtägige Events tageweise untereinander, je Tag eigener Schicht-Track | ebenda |
| F-007 | Farbcodierung: Primärfarbe = genug Helfer, Warnfarbe = teilweise, Fehlerfarbe = kritisch unbesetzt | `calcOccupancy()`, `OccBadge`/`OccFill` |
| F-008 | WCAG 2.1 AA | Kontrast-Check (B-003), ARIA-Labels |
| F-009 | Ladezeiten < 2 s für Hauptansichten | Code-Splitting (lazy-loaded Screens in `App.tsx`) |
| F-010 | Browser-Support: aktuelle Chrome/Firefox/Safari/Edge | — |

## Architektur-Hinweis

ShiftManager ist ein **modularer Monolith**: das Go-Backend bettet das
kompilierte React-SPA zur Buildzeit via `embed.FS` ein
(`internal/infrastructure/static`) — es gibt keinen separaten Static-Server
und kein eigenständiges Node.js-Backend. Der frühere Entwurf, der
Node.js/Express/Fastify/Next.js als Backend-Option vorschlug, ist damit
gegenstandslos und wurde aus der Spezifikation entfernt.
