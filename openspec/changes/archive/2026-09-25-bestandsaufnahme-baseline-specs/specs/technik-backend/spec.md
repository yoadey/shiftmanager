# Spec Delta

## Purpose

Beschreibt die Plattformanforderungen an das Go-Backend: API-Kontrakt,
Authentifizierung, Betriebsfähigkeit und Speicher-Infrastruktur.

## ADDED Requirements

### Requirement: RESTful JSON-API unter /api/v1 (T-001)
Das Backend MUST eine RESTful-JSON-API unter dem Präfix `/api/v1`
bereitstellen.

#### Scenario: API-Aufruf
- **WHEN** ein Client einen beliebigen API-Endpunkt aufruft
- **THEN** liegt dessen Pfad unter `/api/v1` und die Antwort ist JSON

### Requirement: Vollständige OpenAPI-Dokumentation (T-002)
Die API MUST vollständig per OpenAPI 3.0 dokumentiert sein, generiert aus
einer einzigen Quelle.

#### Scenario: Codegen aus der Spezifikation
- **WHEN** `api/openapi.yaml` geändert wird und `make generate` ausgeführt wird
- **THEN** aktualisiert das System sowohl die Backend-Typen als auch das Frontend-SDK konsistent aus dieser einen Quelle

### Requirement: Bearer-JWT-Authentifizierung (T-003)
Der API-Zugriff MUST über Bearer-JWT authentifiziert werden, außer bei
explizit öffentlichen Endpunkten (Login, Kiosk, öffentlicher
Schicht-Bestätigungslink).

#### Scenario: Zugriff ohne Token
- **WHEN** ein geschützter Endpunkt ohne gültiges `Authorization: Bearer`-Token aufgerufen wird
- **THEN** lehnt das System die Anfrage mit 401 ab

### Requirement: Konfigurierbares CORS (T-004)
CORS MUST über eine konfigurierbare Origin-Whitelist steuerbar sein.

#### Scenario: Fremde Origin
- **WHEN** eine nicht in der Whitelist enthaltene Origin eine Anfrage mit Preflight stellt
- **THEN** verweigert das System die Anfrage über die CORS-Header

### Requirement: Rate-Limiting öffentlicher Endpunkte (T-005)
Öffentliche Endpunkte (inklusive Kiosk) MUST pro IP-Adresse
ratenbegrenzt sein; in `TEST_MODE` MUST diese Begrenzung deaktiviert
sein.

#### Scenario: Zu viele Anfragen von einer IP
- **WHEN** eine IP-Adresse die konfigurierte Anfragerate für einen öffentlichen Endpunkt überschreitet
- **THEN** lehnt das System weitere Anfragen dieser IP vorübergehend ab

### Requirement: Typsicherer Datenbankzugriff (T-006)
Datenbankzugriffe MUST über einen typsicheren ORM/Query-Builder
erfolgen.

#### Scenario: Kein rohes SQL für Nutzereingaben
- **WHEN** eine Datenbankabfrage mit Nutzereingaben ausgeführt wird
- **THEN** erfolgt sie über den ORM mit parametrisierten Queries

### Requirement: Migrationstool für Produktionsdatenbank (T-007)
Für die PostgreSQL-Produktionsdatenbank MUST ein Migrationstool
verwendet werden.

#### Scenario: Schemaänderung in Produktion
- **WHEN** ein Datenbankschema-Update ausgerollt wird
- **THEN** erfolgt es über eine versionierte Migrationsdatei, nicht über manuelle Schemaänderungen

### Requirement: Konfiguration über Umgebungsvariablen (T-008)
Die Konfiguration MUST über Umgebungsvariablen bzw. eine `.env`-Datei
erfolgen.

#### Scenario: Serverstart ohne Konfigurationsdatei im Code
- **WHEN** der Server startet
- **THEN** liest er seine Konfiguration ausschließlich aus Umgebungsvariablen, nicht aus fest einkompilierten Werten

### Requirement: Strukturiertes Logging (T-009)
Das Backend MUST strukturiertes (JSON-)Logging mit konfigurierbaren
Log-Levels bereitstellen.

#### Scenario: Log-Level ändern
- **WHEN** das konfigurierte Log-Level geändert wird
- **THEN** passt das System die Ausführlichkeit der JSON-Logausgabe entsprechend an

### Requirement: Health-Check-Endpunkte (T-010)
Das Backend MUST Health-Check-Endpunkte bereitstellen.

#### Scenario: Liveness-/Readiness-Abfrage
- **WHEN** `/health` oder `/readyz` aufgerufen wird
- **THEN** liefert das System den aktuellen Betriebsstatus

### Requirement: Hintergrundtasks als eigene Prozesse (T-011)
E-Mail-Versand, Reservierungsablauf und Jahresabrechnung MUST als eigene
Hintergrundtasks (Goroutinen/Cron-Jobs) laufen, unabhängig vom
Request-Handling.

#### Scenario: Reservierungsablauf im Hintergrund
- **WHEN** eine Reservierung ihre Gültigkeitsdauer überschreitet
- **THEN** hebt ein Hintergrundtask die Reservierung auf, ohne dass ein Nutzer-Request dafür nötig ist

### Requirement: Konfigurierbarer Medien-Speicherort (T-013)
Der Speicherort für hochgeladene Medien (Logo, Event-Anhänge) MUST
zwischen lokalem Verzeichnis/PVC und S3-kompatiblem Object-Storage
konfigurierbar sein, ohne dass ein Migrationsscript für Bestandsdateien
nötig ist.

#### Scenario: Wechsel auf S3-kompatiblen Storage
- **WHEN** `MEDIA_STORAGE=s3` mit gültigen `S3_ENDPOINT`-Zugangsdaten konfiguriert ist
- **THEN** speichert das System neue Uploads über den S3-Adapter statt lokal

#### Scenario: Öffentlicher Lesezugriff ohne Auth
- **WHEN** Logo oder Event-Anhang unter ihrer nicht erratbaren URL abgerufen werden
- **THEN** liefert das System sie ohne Auth-Prüfung aus (weder lokal noch über S3 wird ein privates ACL erzwungen — der Bucket bzw. Pfad muss stattdessen selbst für Lesezugriff freigegeben sein)
