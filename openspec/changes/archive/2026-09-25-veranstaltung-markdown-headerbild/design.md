# Design

## Context

Die Beschreibung ist heute ein reiner `string` in `domain.Event`
(`Description`), bearbeitet über ein einfaches `<Textarea>` in
`CreateEventFlow.tsx:220` und im Bearbeiten-Sheet von
`EventDetail.tsx:705f.`. Anhänge (`EventAttachment`) laufen bereits über
`port.MediaStorage` mit Extension-/Content-Type-Allowlist (SVG
ausgeschlossen), 5-MB-Limit und lokalem oder S3-kompatiblem Storage
(`T-013`). Diese Infrastruktur wird für das Headerbild wiederverwendet.

Das Eventformular wird im Rahmen der separaten Change
`bedienkonzept-formulare-als-seite` ohnehin von einem Sheet-Overlay zu
einer eigenen Seite umgebaut; der neue Editor und der Headerbild-Upload
sollten direkt in dieser neuen Seitenstruktur landen statt zusätzlich im
alten Sheet.

## Goals / Non-Goals

**Goals:**
- Markdown als einziges Speicherformat der Beschreibung (keine
  Zweit-Repräsentation als HTML).
- Sicheres Rendering (kein Stored-XSS über die Beschreibung).
- Headerbild nutzt dieselbe Validierung/Storage-Infrastruktur wie
  bestehende Anhänge.

**Non-Goals:**
- Kein Rich-Media-Editor über Markdown hinaus (kein eingebettetes Video,
  keine Tabellen-Editor-UI in v1).
- Keine Änderung an der bestehenden Mehrfach-Anhang-Funktion (`V-008`)
  selbst — das Headerbild ist ein separates Feld/Endpoint.

## Decisions

- **Speicherformat: Markdown-String, kein separates HTML-Feld.**
  Alternative wäre, zusätzlich gerendertes HTML serverseitig zu
  cachen — verworfen, da das eine zweite Quelle der Wahrheit schafft und
  bei jeder Sanitizing-Regeländerung neu generiert werden müsste. Das
  Rendering erfolgt clientseitig (und für E-Mail-Versand serverseitig)
  jeweils frisch aus dem Markdown.
- **WYSIWYG-Editor mit Markdown-Export statt reinem Markdown-Textfeld
  als Standard.** Zielgruppe sind Vereins-Ehrenamtliche ohne
  Markdown-Erfahrung; ein Umschalter auf den Rohtext bleibt für
  versierte Nutzer erhalten.
- **Headerbild als eigenes Feld/Endpoint, nicht als "erster Anhang"
  markiert.** Ein spezielles `IsHeader`-Flag auf einem bestehenden
  `EventAttachment` wäre implizit und fehleranfällig (z. B. beim Löschen
  des "ersten" Anhangs); ein eigenes Feld macht die Beziehung explizit.
- **Sanitizing beim Rendern, nicht beim Speichern.** Das gespeicherte
  Markdown bleibt unverändert (originalgetreu für erneutes Bearbeiten);
  die Sanitizing-Regel wird beim Rendern angewendet, sodass eine spätere
  Verschärfung der Regel ohne Datenmigration wirkt.

## Risks / Trade-offs

- **Neue Frontend-Abhängigkeit** (Markdown-WYSIWYG-Bibliothek) → erhöht
  Bundle-Größe; mitigiert durch Lazy-Loading des Editors analog zum
  bestehenden Code-Splitting (`F-009`).
- **XSS über Markdown-Edge-Cases** (z. B. `javascript:`-Links) →
  mitigiert durch eine etablierte Sanitizing-Bibliothek statt
  eigener Regex-Filterung.

## Open Questions

- Konkrete Bibliothekswahl für den WYSIWYG-Markdown-Editor (z. B. Tiptap
  mit Markdown-Serialisierung vs. eine fertige React-Markdown-WYSIWYG-
  Komponente) — beeinflusst nicht die Spec-Requirements oder die
  Task-Aufteilung, wird bei Task 2 (Editor auswählen) entschieden.
