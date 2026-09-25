# ShiftManager — openspec

Dieses Verzeichnis ist die aktuelle, mit der Implementierung synchron gehaltene
Spezifikation von ShiftManager. Es ersetzt den früheren Entwurf (Stand
2026-06-04), der ein anderes Rollen- und Datenmodell (Helfer/Koordinator/Gast,
`Bereich`-Entität) beschrieb, das nie umgesetzt wurde und nicht mehr der
Realität entspricht.

**Quelle der Wahrheit:** [`project/requirements_extracted.txt`](../../project/requirements_extracted.txt)
— das vollständige Anforderungsdokument mit Anforderungs-IDs (`A-001`,
`SC-010`, `T-013`, …) und Implementierungsstatus. Die Dateien in diesem
Verzeichnis gliedern denselben Inhalt nach Funktionsbereich, mit direktem
Bezug zu Code (Endpunkte, Screens) statt nur zur Anforderungsliste.

## Aufbau

| Datei | Inhalt |
|---|---|
| [`00-general.md`](00-general.md) | Projektziel, Rollen, Datenmodell, Architektur-Überblick, Seiten-Übersicht |
| [`01-auth-mitglieder.md`](01-auth-mitglieder.md) | OIDC-Login (A-xxx), Mitgliederverwaltung (ML-xxx), Namensanzeige (NM-xxx) |
| [`02-kiosk.md`](02-kiosk.md) | Kiosk-Modus (K-xxx) |
| [`03-veranstaltungen-schichten.md`](03-veranstaltungen-schichten.md) | Veranstaltungen (V-xxx, VC-xxx), Schichten (SC-xxx) |
| [`04-stunden-abrechnung.md`](04-stunden-abrechnung.md) | Stundenziel (S-xxx), Stundenerfassung (SA-xxx), Abgeltung (G-xxx) |
| [`05-dashboard-berichte.md`](05-dashboard-berichte.md) | Dashboards (D-xxx), Berichte/Exporte, Audit-Log |
| [`06-benachrichtigungen.md`](06-benachrichtigungen.md) | E-Mail-Benachrichtigungen, Präferenzen (N-xxx) |
| [`07-branding-einstellungen.md`](07-branding-einstellungen.md) | Branding (B-xxx), Profil, Systemeinstellungen |
| [`08-technik.md`](08-technik.md) | Backend/Frontend-Architektur (T-xxx, F-xxx) |
| [`09-datenschutz-sicherheit.md`](09-datenschutz-sicherheit.md) | Datenschutz & Sicherheit (DS-xxx) |

Jede Datei listet nur Anforderungen, die **umgesetzt** sind — Implementiertes
wird hier beschrieben, nicht als offener Punkt geführt.

## `changes/`

Noch nicht umgesetzte oder neu vorgeschlagene Arbeit wird **nicht** in die
obigen Dateien gemischt, sondern als eigener Change-Vorschlag unter
`changes/` geführt, bis sie umgesetzt ist. Danach wird der entsprechende
Change gelöscht und der Inhalt in die passende Spec-Datei übernommen.

Aktuell offen:

| Change | Betrifft |
|---|---|
| [`changes/2026-09-25-veranstaltung-markdown-headerbild.md`](changes/2026-09-25-veranstaltung-markdown-headerbild.md) | Veranstaltungsbeschreibung als Markdown/WYSIWYG (`V-009`), Headerbild (`V-010`) |
| [`changes/2026-09-25-konfiguration-eigene-seite-statt-popup.md`](changes/2026-09-25-konfiguration-eigene-seite-statt-popup.md) | Neuer Grundsatz `UX-001`: Formulare mit >5 Feldern als eigene Seite statt Popup; Umbau Event-Erstellen/-Bearbeiten |
| [`changes/2026-09-25-profilkarte-navigation-glocke-entfernen.md`](changes/2026-09-25-profilkarte-navigation-glocke-entfernen.md) | Profil über Profilkarte unten links erreichbar (`NAV-001`), Glocken-Icon entfernen (`NAV-002`) |
