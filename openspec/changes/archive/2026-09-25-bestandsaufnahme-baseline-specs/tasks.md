# Tasks

## 1. Capability-Deltas schreiben

- [x] 1.1 Alle 18 Capabilities aus `docs/openspec/00-09-*.md` als `specs/<capability>/spec.md` mit `ADDED Requirements` übertragen und verifizieren, dass jede bestehende Anforderungs-ID aus `project/requirements_extracted.txt` in einem Requirement-Namen auftaucht
- [x] 1.2 Für die bisher ID-losen Bereiche Profil und Systemeinstellungen die neuen Präfixe `PR-`/`SE-` konsistent vergeben

## 2. Validieren und archivieren

- [x] 2.1 `openspec validate bestandsaufnahme-baseline-specs --strict` ausführen und Befunde beheben, bis der Change gültig ist
- [x] 2.2 `openspec archive bestandsaufnahme-baseline-specs --yes` ausführen und verifizieren, dass alle 18 `openspec/specs/<capability>/spec.md`-Dateien angelegt wurden

## 3. Alte Dokumentation ablösen

- [x] 3.1 `docs/openspec/` entfernen, nachdem die Inhalte vollständig unter `openspec/specs/` verfügbar sind
- [x] 3.2 `CLAUDE.md` um einen kurzen Verweis auf den OpenSpec-Workflow (`openspec/`, `/opsx:*`-Skills) ergänzen
