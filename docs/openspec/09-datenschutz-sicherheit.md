# 09 — Datenschutz und Sicherheit

Umgesetzt. Anforderungs-IDs: `DS-001`–`DS-009`.

| ID | Anforderung | Umsetzung |
|---|---|---|
| DS-001 | DSGVO-konformer Betrieb | siehe alle folgenden Punkte |
| DS-002 | Verschlüsselte Übertragung (HTTPS/TLS) | Deployment-Anforderung (Reverse Proxy/TLS-Terminierung) |
| DS-003 | Mitglieder können gespeicherte Daten abrufen (Auskunftsrecht) | `GET /api/v1/members/{id}/export-data` — eigene Daten für jedes Mitglied, fremde Daten nur für Vorstand+ (Handler prüft Eigentümerschaft/Rolle) |
| DS-004 | Mitglieder können Löschung beantragen (Recht auf Vergessen, mit Ausnahmen) | `POST /api/v1/members/{id}/gdpr-delete`, `RequireRole(RoleVorstand)` |
| DS-005 | Keine Passwortspeicherung im System (Auslagerung an OIDC-Provider) | siehe `01-auth-mitglieder.md` |
| DS-006 | JWT-Tokens mit kurzer Lebensdauer | `Config.JWTExpiration` (ENV `JWT_EXPIRATION`), konfigurierbar — **Ist-Default aktuell 24 Stunden** (`internal/config/config.go`), nicht die im ursprünglichen Anforderungsdokument genannten 15 Minuten. Das automatische Token-Refresh (A-004) mindert das Risiko eines langlebigen Tokens, macht die Diskrepanz zum dokumentierten Standard aber nicht hinfällig — bei einer sicherheitsrelevanten Überarbeitung sollte der Default an die Anforderung angeglichen oder die Anforderung bewusst aktualisiert werden. |
| DS-007 | Schutz vor SQL-Injection, XSS, CSRF über Framework-Features | GORM (parametrisierte Queries), React (Auto-Escaping), CORS-Whitelist (T-004) |
| DS-008 | Kiosk-Bestätigungslinks kryptografisch zufällig, einmalig verwendbar | siehe `K-011` in `02-kiosk.md` |
| DS-009 | Audit-Logs mind. 2 Jahre aufbewahrt, änderungsgeschützt | siehe `05-dashboard-berichte.md` |

## Hinweis zu Medien-Uploads

Logo- und Event-Anhang-Uploads (`B-004`, `V-008`) sind größen- und
typbegrenzt (siehe `03-veranstaltungen-schichten.md`, `07-branding-einstellungen.md`).
Der Speicherort ist konfigurierbar: lokales Verzeichnis/PVC (Default) oder
S3-kompatibler Object-Storage — siehe `T-013` in `08-technik.md`.
