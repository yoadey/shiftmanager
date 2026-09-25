# Proposal

## Why

ShiftManager wird bislang in `docs/openspec/*.md` spezifiziert — einer
projekteigenen, an OpenSpec angelehnten, aber nicht tool-konformen
Konvention (freie Markdown-Tabellen statt `Requirement`/`Scenario`-Blöcke,
kein `openspec`-CLI, keine Validierung). Um echtes Spec-Driven Development
mit der OpenSpec-CLI und den `/opsx:*`-Skills nutzen zu können (Vorschläge,
Deltas, Validierung, Archivierung), muss der bereits **umgesetzte**
Funktionsumfang zunächst als Baseline unter `openspec/specs/` vorliegen.
Diese Change überträgt den Inhalt der neun nummerierten `docs/openspec/`-
Dateien 1:1 inhaltlich (keine neuen fachlichen Anforderungen) in
`Requirement`/`Scenario`-Form je Capability. `docs/openspec/` wird danach
entfernt — `openspec/` ist ab sofort die einzige Spezifikationsquelle.

## What Changes

- Führt 18 neue Capabilities unter `openspec/specs/` ein, die den
  vollständig **bereits implementierten** Funktionsumfang von
  ShiftManager beschreiben (siehe Capabilities-Liste unten).
- Jede bestehende Anforderungs-ID aus
  `project/requirements_extracted.txt` (`A-xxx`, `ML-xxx`, `NM-xxx`,
  `K-xxx`, `V-xxx`, `VC-xxx`, `SC-xxx`, `S-xxx`, `SA-xxx`, `G-xxx`,
  `D-xxx`, `N-xxx`, `B-xxx`, `T-xxx`, `F-xxx`, `DS-xxx`) taucht als
  Suffix im jeweiligen Requirement-Namen auf, z. B.
  `### Requirement: Mehrtägige Veranstaltungen (V-003)`, damit die
  Rückverfolgbarkeit zur bisherigen Dokumentation und zum
  Anforderungsdokument erhalten bleibt.
- Für die drei Bereiche ohne bisherige ID (Profil, Systemeinstellungen als
  eigenständige Capability) werden neue Präfixe `PR-` und `SE-` vergeben.
- Keine Verhaltensänderung am System — reine Dokumentationsmigration.

## Capabilities

### New Capabilities
- `auth`: OIDC-Anmeldung, Multi-Provider, JWT-Session, Bootstrap-Admin
- `mitgliederverwaltung`: Mitglieder anlegen/pflegen, CSV-Import/-Export
- `namensanzeige`: Systemweiter Namensanzeige-Modus (voll/abgekürzt) und Ausnahmen
- `kiosk`: Schichteintragung ohne Login am Vereinsheim-PC
- `veranstaltungen`: Veranstaltungen anlegen/verwalten/kopieren
- `schichten`: Schichten innerhalb einer Veranstaltung, Anmeldung, Besetzung
- `stundenziel`: Stundenziel-Konfiguration pro Mitglied/Vereinsjahr
- `stundenerfassung`: Bestätigung geleisteter Stunden, manuelle Buchung
- `abgeltung`: Abgeltungsbeträge für Fehlstunden, Jahresabrechnung
- `dashboards`: Mitglieder- und Verwaltungs-Dashboard, Entdecken-Ansicht
- `benachrichtigungen`: E-Mail-Benachrichtigungen und -Vorlagen
- `branding`: Vereinsname/Logo/Farben, Kiosk-Hintergrund
- `profil`: Eigenes Profil, Erinnerungspräferenzen, Logout
- `systemeinstellungen`: Zugriff auf die Verwaltungsoberfläche für Systemeinstellungen
- `audit-log`: Protokollierung sicherheits-/datenschutzrelevanter Aktionen
- `datenschutz-sicherheit`: DSGVO-Auskunft/-Löschung, Transportsicherheit, Token-Lebensdauer
- `technik-backend`: Backend-Plattformanforderungen (API, Auth, Rate-Limiting, Storage, …)
- `technik-frontend`: Frontend-Plattformanforderungen (SPA, Responsive, Performance, …)

### Modified Capabilities
(keine — dies ist die Erstanlage aller Capabilities)

## Impact

- Reine Dokumentation: `openspec/specs/**` (neu), `docs/openspec/**`
  (entfernt in einer Folge-Änderung nach dem Archivieren dieser Change,
  siehe Repo-`README`-Historie).
- Kein Code, keine API, keine Migration betroffen.
