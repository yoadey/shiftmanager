# Proposal

## Why

Die Detailansicht einer Veranstaltung verschenkt viel Platz, bevor die
Schichten kommen:

- Der Bereich „Anhänge“ steht auch ohne Anhänge da, nur um für
  Veranstaltungsleiter ein „+“ anzubieten.
- Drei große Kacheln (Schichten, Freie Plätze, Tage) wiederholen, was die
  Zeitleiste ohnehin zeigt.
- Lange Beschreibungen schieben die Schichten nach unten.
- In jeder Schicht stehen Uhrzeit und Titel untereinander, auch wenn
  daneben Platz wäre.
- Veranstaltungsleiter sehen die eingetragenen Helfer erst nach Klick auf
  „x eingetragen“ – eine Zahl, die der Belegungsbalken direkt darüber
  schon zeigt. Auch „Helfer hinzufügen“ liegt hinter diesem Klick.

## What Changes

- Anhänge: In der Veranstaltungsansicht erscheint der Bereich nur, wenn
  Anhänge vorhanden sind, und nur zum Ansehen. Hochladen und Entfernen
  geschieht auf der Bearbeiten-Seite (neben dem Headerbild).
- Die Übersichtskacheln (Schichten, Freie Plätze, Tage) entfallen.
- Die Beschreibung steht in einem eigenen, mit „Beschreibung“
  überschriebenen Bereich und ist standardmäßig eingeklappt: Nur die
  ersten Zeilen sind sichtbar, der Rest blendet aus; „Mehr anzeigen“/
  „Weniger anzeigen“ klappt auf und zu.
- Schicht-Kopf: Der Titel steht neben Uhrzeit und Dauer, wenn die Breite
  reicht, sonst darunter.
- Helferliste: Veranstaltungsleiter sehen die eingetragenen Helfer jeder
  Schicht direkt (ohne Aufklappen); die Zeile „x eingetragen“ entfällt.
  „Helfer hinzufügen“ ist in jeder Schicht direkt sichtbar und
  hervorgehoben, solange Plätze frei sind. Auf breiten Bildschirmen
  stehen die Helfer nebeneinander.

## Capabilities

### New Capabilities

_Keine._

### Modified Capabilities

- `veranstaltungen`: V-008 geändert (Anhänge nur auf der Bearbeiten-Seite
  verwalten, in der Ansicht nur bei vorhandenen Anhängen); V-011 neu
  (kompakte Veranstaltungsansicht).
- `schichten`: SC-012 neu (Helferliste und „Helfer hinzufügen“ direkt in
  der Schicht).

## Impact

- Frontend: `screens/events/EventDetail.tsx` (Ansicht, Bearbeiten-Seite,
  `ShiftRow`), zugehörige Tests. Keine API- oder Datenbankänderung.
