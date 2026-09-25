# Spec Delta

## Purpose

Regelt Schichten innerhalb einer Veranstaltung: Anlage, Anzeige,
Besetzung, An-/Abmeldung und Abschluss.

## ADDED Requirements

### Requirement: Mehrere Schichten pro Veranstaltung (SC-001)
Eine Veranstaltung MUST mehrere Schichten enthalten können, jede einem
Veranstaltungstag zugeordnet.

#### Scenario: Zweite Schicht am selben Tag anlegen
- **WHEN** einem Veranstaltungstag eine weitere Schicht hinzugefügt wird
- **THEN** verwaltet das System beide Schichten unabhängig voneinander

### Requirement: Pflichtfelder einer Schicht (SC-002)
Eine Schicht MUST Name, Start-/Endzeit sowie Min./Max.-Helferzahl
enthalten; eine Beschreibung ist optional.

#### Scenario: Anlage ohne Pflichtfeld
- **WHEN** beim Anlegen einer Schicht Name, Start- oder Endzeit fehlt
- **THEN** lehnt das System die Anfrage ab

### Requirement: Übersichtliche Zeitplandarstellung (SC-003)
Schichten MUST in einer übersichtlichen Zeitplan-/Timeline-Darstellung
angezeigt werden.

#### Scenario: Timeline einer Veranstaltung öffnen
- **WHEN** die Detailansicht einer Veranstaltung geöffnet wird
- **THEN** zeigt das System alle Schichten in einer Zeitplan-Darstellung

### Requirement: Hervorhebung unterbesetzter Schichten (SC-004)
Unterbesetzte Schichten MUST optisch hervorgehoben werden.

#### Scenario: Schicht unter Mindestbesetzung
- **WHEN** die Anzahl angemeldeter Helfer einer Schicht unter dem Minimum liegt
- **THEN** hebt das System die Schicht optisch als unterbesetzt hervor

### Requirement: Kennzeichnung vollständig belegter Schichten (SC-005)
Vollständig belegte Schichten MUST als "Ausgebucht" markiert werden.

#### Scenario: Schicht erreicht Maximalbesetzung
- **WHEN** die Anzahl angemeldeter Helfer das Maximum einer Schicht erreicht
- **THEN** markiert das System die Schicht als "Ausgebucht"

### Requirement: Abmeldefrist (SC-006)
Eine Abmeldung von einer Schicht MUST nur bis zu einer konfigurierbaren
Frist vor Schichtbeginn möglich sein.

#### Scenario: Abmeldung innerhalb der Frist
- **WHEN** ein Mitglied sich vor Ablauf der konfigurierten Abmeldefrist abmeldet
- **THEN** akzeptiert das System die Abmeldung

#### Scenario: Abmeldung nach Fristablauf
- **WHEN** ein Mitglied sich nach Ablauf der Abmeldefrist abzumelden versucht
- **THEN** lehnt das System die Abmeldung ab

### Requirement: Benachrichtigung bei Unterbesetzung (SC-007)
Unterschreitet eine Schicht die Mindesthelferzahl in Frist-Nähe, MUST der
Veranstaltungsleiter benachrichtigt werden.

#### Scenario: Mindesthelferzahl unterschritten
- **WHEN** eine Schicht die konfigurierte Mindesthelferzahl unterschreitet
- **THEN** sendet das System eine Benachrichtigung an den Veranstaltungsleiter

### Requirement: Schichtabschluss mit individueller Stundenbestätigung (SC-008)
Der Veranstaltungsleiter MUST Schichten abschließen und die tatsächlich
geleistete Zeit je Person individuell bestätigen können.

#### Scenario: Stunden bestätigen
- **WHEN** ein Veranstaltungsleiter `POST /api/v1/hours/confirm` für eine abgeschlossene Schicht mit individuellen Zeiten je Helfer aufruft
- **THEN** übernimmt das System die bestätigten Stunden je Person

### Requirement: Qualifikationsanforderungen an Schichten (SC-009)
Eine Schicht SHOULD eine Qualifikationsanforderung (z. B. Führerschein)
tragen können.

#### Scenario: Schicht mit Qualifikation anlegen
- **WHEN** eine Schicht mit einer Qualifikationsanforderung angelegt wird
- **THEN** speichert das System diese Anforderung an der Schicht

### Requirement: Beliebige Helfer eintragen und entfernen (SC-010)
Veranstaltungsleiter und Vorstand MUST beliebige Helfer — auch Personen
ohne Benutzerkonto (nur Name, optional E-Mail) — zu einer Schicht
eintragen und jede Eintragung (eigene oder fremde) wieder entfernen
können. Die Übersicht MUST dabei stets vollständige Namen zeigen,
unabhängig vom Namensanzeige-Modus (siehe `namensanzeige` NM-005-Ausnahme).

#### Scenario: Gast ohne Konto eintragen
- **WHEN** ein Veranstaltungsleiter über `POST /api/v1/shifts/{id}/add-guest` eine Person nur mit Namen einträgt
- **THEN** legt das System eine Registrierung ohne Benutzerkonto-Bezug an

#### Scenario: Doppelte Gast-E-Mail abgelehnt
- **WHEN** für dieselbe Schicht ein zweiter Gast mit derselben E-Mail-Adresse eingetragen werden soll
- **THEN** lehnt das System die zweite Eintragung ab

#### Scenario: Vollständiger Name in der Helferliste
- **WHEN** ein Veranstaltungsleiter die Helferliste einer Schicht ansieht, während systemweit der abgekürzte Namensmodus aktiv ist
- **THEN** zeigt das System dort trotzdem vollständige Namen
