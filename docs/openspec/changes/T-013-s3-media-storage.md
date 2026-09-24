# Change: T-013 — S3-kompatibler Object-Storage für Medien

**Status:** offen · **Anforderung:** `T-013` (KANN) · **Quelle:**
Nutzeranfrage vom 2026-09-23, festgehalten in `project/requirements_extracted.txt` 5.2

## Problem

Logo-Uploads (`B-004`) und Event-Anhänge (`V-008`) werden aktuell
ausschließlich auf ein lokales Verzeichnis geschrieben, konfiguriert über
`UPLOAD_DIR` (Default `./uploads`, siehe `internal/config/config.go`):

- `internal/adapter/http/handler/upload.go` (`receiveUpload`): schreibt via
  `os.Create`/`io.Copy` direkt ins lokale Dateisystem — gemeinsam genutzt
  von `event_handler.go` (Anhänge) und `settings_handler.go` (Logo).
- `internal/adapter/http/router.go`: `/uploads/*` wird per
  `http.FileServer(http.Dir(cfg.UploadDir))` ausgeliefert.
- Löschpfade (`event_handler.go`, `settings_handler.go`): `os.Remove(...)`.

In einer Kubernetes-Bereitstellung erfordert das ein PersistentVolumeClaim
mit `ReadWriteMany`, was je nach Storage-Klasse nicht überall verfügbar ist.
Ein S3-kompatibler Object-Storage (AWS S3, MinIO, Hetzner Object Storage, …)
soll als Alternative konfigurierbar sein — **ohne** Migrationsscript für
bereits vorhandene lokale Dateien (laut Nutzeranfrage explizit nicht
gefordert).

## Vorgeschlagene Lösung

1. **Abstraktion:** neues Port-Interface `port.MediaStorage` mit
   `Put(ctx, key string, r io.Reader, contentType string) (url string, err error)`
   und `Delete(ctx, key string) error`, das `receiveUpload` und die
   Lösch-Pfade in `event_handler.go`/`settings_handler.go` statt direkter
   `os.*`-Aufrufe verwenden.
2. **Adapter:**
   - `internal/adapter/storage/local` — heutiges Verhalten (Default,
     rückwärtskompatibel zu `UPLOAD_DIR`).
   - `internal/adapter/storage/s3` — via AWS SDK v2 (`aws-sdk-go-v2/service/s3`),
     kompatibel zu jedem S3-API-Objektspeicher über konfigurierbaren
     `Endpoint` (MinIO, Hetzner, …) + `ForcePathStyle`.
3. **Konfiguration** (analog zum bestehenden `.env`-Muster):
   `MEDIA_STORAGE=local|s3` (Default `local`), sowie bei `s3`:
   `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY_ID`,
   `S3_SECRET_ACCESS_KEY`, optional `S3_FORCE_PATH_STYLE`.
4. **Public URLs:** bei `local` bleibt `/uploads/*` wie bisher; bei `s3`
   entweder öffentliche Bucket-URLs oder (falls der Bucket privat bleiben
   soll) zeitlich begrenzte presigned GET-URLs — Entscheidung hängt davon
   ab, ob Event-Anhänge/Logos öffentlich lesbar sein dürfen (aktuell sind
   sie es über `/uploads/*` ohne Auth-Prüfung).
5. **Wiring:** `cmd/server/main.go` wählt den Adapter anhand
   `MEDIA_STORAGE` und injiziert ihn in die Handler, analog zum
   bestehenden `db.Open()`-Dispatcher (SQLite vs. PostgreSQL anhand
   DSN-Präfix).

## Betroffene Bereiche

- `internal/port` (neues Interface)
- `internal/adapter/storage/{local,s3}` (neu)
- `internal/adapter/http/handler/upload.go`, `event_handler.go`,
  `settings_handler.go`
- `internal/config`, `cmd/server/main.go`
- `docs/openspec/03-veranstaltungen-schichten.md`,
  `docs/openspec/07-branding-einstellungen.md` (nach Umsetzung aktualisieren)

## Migration

Kein Migrationsscript für bestehende lokale Dateien (Nutzervorgabe). Ein
Wechsel von `local` auf `s3` beginnt mit leerem Bucket; bereits
hochgeladene Logos/Anhänge müssten bei Bedarf manuell nachgeladen werden.
