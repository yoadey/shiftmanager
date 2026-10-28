# Spec Delta

## Purpose

Hält projektweite Bedienkonzept-Grundsätze fest, die unabhängig von
einer einzelnen Capability gelten (aktuell: wann ein Formular eine eigene
Seite statt eines Popups ist).

## ADDED Requirements

### Requirement: Formulare mit mehr als 5 Feldern als eigene Seite (UX-001)
Ein Formular oder Konfigurationsdialog mit mehr als 5 Eingabefeldern MUST
als eigene Seite umgesetzt werden — eigene Route, volle App-Navigation
sichtbar, kein Overlay/Backdrop, kein Schließen durch Klick außerhalb.
Ein Popup/Sheet/Modal-Overlay MUST auf Bestätigungen, kurze Eingaben und
Formulare mit höchstens 5 Feldern beschränkt bleiben.

#### Scenario: Formular mit mehr als 5 Feldern
- **WHEN** ein neues oder bestehendes Formular mehr als 5 Eingabefelder hat
- **THEN** wird es als eigene Seite mit eigener Route umgesetzt, nicht als Sheet/Modal-Overlay

#### Scenario: Formular mit höchstens 5 Feldern
- **WHEN** ein Formular höchstens 5 Eingabefelder hat oder eine reine Bestätigung ist
- **THEN** darf es als Popup/Sheet umgesetzt werden

#### Scenario: Veranstaltung erstellen/bearbeiten als Seite
- **WHEN** ein Veranstaltungsleiter eine Veranstaltung anlegt oder bearbeitet (Name, Beschreibung, Ort, Kategorie, Start-/Enddatum, Status plus Schichtenliste — mehr als 5 Felder)
- **THEN** öffnet das System dafür eine eigene Seite unter `routes.eventNeu`/`routes.eventBearbeiten(id)`, keine Sheet-Overlay
