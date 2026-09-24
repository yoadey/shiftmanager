# 07 — Branding, Profil, Systemeinstellungen

Umgesetzt. Anforderungs-IDs: `B-001`–`B-008`.

## Branding

Das System ist für den Einsatz durch verschiedene Vereine ausgelegt:
Primärfarbe, Akzentfarbe und Logo sind vollständig konfigurierbar.

| ID | Anforderung | Umsetzung |
|---|---|---|
| B-001 | Admin konfiguriert Primär-/Akzentfarbe fürs gesamte Frontend | `PUT /api/v1/settings/branding`, `RequireRole(RoleVorstand)` |
| B-002 | Farben als CSS-Custom-Properties, wirken auf Navigation, Buttons, Balken, Badges systemweit | `App.tsx` (`--primary`, `--on-primary`, …) |
| B-003 | Kontrastverhältnis bleibt WCAG-AA-konform (≥ 4,5:1); Warnung bei Unterschreitung | `pickOn()`-Helfer in `App.tsx`, Kontrast-Check in `AdminSettings.tsx` |
| B-004 | Logo-Upload (PNG, SVG; empfohlen ≥ 200×200 px) | `POST /api/v1/settings/logo`, gespeichert über `port.MediaStorage` (lokal oder S3-kompatibel, siehe `T-013` in `08-technik.md`) |
| B-005 | Logo in Seitennavigation (Desktop) und Kiosk-Header | `DesktopShell.tsx`, `MobileShell.tsx`, `KioskPage.tsx` |
| B-006 | Vereinsname als Text konfigurierbar, Fallback ohne Logo | `Branding.ClubName`, `clubInitials()`-Fallback |
| B-007 | Kiosk: zusätzlich Hintergrundbild/-farbe konfigurierbar | `Branding.KioskBackground` |
| B-008 | Farb-/Logo-Konfiguration versioniert, zurücksetzbar | `GET /api/v1/settings/branding/history`, `POST /api/v1/settings/branding/rollback/{id}` |

## Profil (Mitglied)

`MemberProfile.tsx` — eigenes Profil, Namensanzeige-Kontext (NM-006),
Erinnerungspräferenzen (N-001), Logout. Erreichbar über den `profil`-Tab
(alle Rollen) sowie über den Profil-Shortcut im Verwaltungs-Dashboard.

## Systemeinstellungen (Vorstand+)

`AdminSettings.tsx` — Bündelt, was pro Anforderungsdokument Vorstand-only
ist: Namensmodus (NM-001), Reservierungsdauer/Abmeldefrist (K-010, SC-006),
Kiosk-Sperre (K-012), Branding (B-xxx), Fee-Tiers (G-001–G-004),
E-Mail-Vorlagen/-Protokoll (N-004), Audit-Log-Zugriff (DS-009). Alle
zugehörigen Schreib-Endpunkte sind `RequireRole(RoleVorstand)`; die reinen
Lese-Endpunkte (`GET /settings`, `/settings/branding`, `/settings/fee-tiers`)
sind für jeden authentifizierten Nutzer offen.
