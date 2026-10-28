# Spec Delta

## MODIFIED Requirements

### Requirement: Mitglieder-Dashboard mit Stundenkonto (D-001)
Das Mitglieder-Dashboard MUST bestätigte Stunden und Stunden inklusive
offener Reservierungen je gegen das Jahres-Stundenziel anzeigen. Fehlen
dem Mitglied bestätigte Stunden zum Jahresziel und ist eine
Gebührenliste konfiguriert, MUST das Stundenkonto zusätzlich den
Abgeltungsbetrag anzeigen, der fällig würde, wenn das Mitglied keine
weiteren Stunden mehr leistet (G-011).

#### Scenario: Stundenkonto anzeigen
- **WHEN** ein Mitglied sein Dashboard öffnet
- **THEN** zeigt das System bestätigte Stunden und Stunden inkl. Reservierungen jeweils im Verhältnis zum Jahresziel

#### Scenario: Abgeltungsprognose anzeigen
- **WHEN** ein Mitglied mit 20 h Jahresziel und 15 h bestätigten Stunden sein Dashboard öffnet und die Gebührenliste 10 € pro Fehlstunde vorsieht
- **THEN** zeigt das Stundenkonto, dass ohne weitere Stunden 50,00 € Abgeltung fällig würden

#### Scenario: Ziel erreicht
- **WHEN** ein Mitglied sein Jahresziel mit bestätigten Stunden erreicht hat
- **THEN** zeigt das Stundenkonto keinen Abgeltungsbetrag

### Requirement: Filterung der Entdecken-Ansicht (D-005)
Die Entdecken-Ansicht MUST nach Datum, Veranstaltung und verfügbaren
Plätzen filterbar sein. Bei jedem Öffnen der Ansicht MUST der
Datumsfilter „Von“ auf das heutige Datum gesetzt und „Bis“ leer sein,
sodass vergangene Veranstaltungen standardmäßig nicht erscheinen; das
Mitglied MUST beide Werte ändern oder leeren können.

#### Scenario: Filter nach Datum
- **WHEN** ein Mitglied einen Datumsfilter anwendet
- **THEN** zeigt das System nur Veranstaltungen im gewählten Zeitraum

#### Scenario: Standardfilter ab heute
- **WHEN** ein Mitglied die Entdecken-Ansicht öffnet
- **THEN** ist „Von“ auf heute gesetzt, „Bis“ leer, und bereits beendete Veranstaltungen werden nicht angezeigt

#### Scenario: Vergangene Veranstaltungen einblenden
- **WHEN** ein Mitglied das Feld „Von“ leert
- **THEN** zeigt das System auch vergangene veröffentlichte Veranstaltungen

## ADDED Requirements

### Requirement: Seitenweise Anzeige der Entdecken-Ansicht (D-007)
Die Entdecken-Ansicht MUST die Veranstaltungen aufsteigend nach Beginn
sortiert und seitenweise anzeigen. Die Filter (D-005) MUST serverseitig
über alle Veranstaltungen wirken, und eine Filteränderung MUST auf die
erste Seite zurückspringen.

#### Scenario: Seitenweise Anzeige
- **WHEN** mehr veröffentlichte Veranstaltungen den Filter erfüllen, als auf eine Seite passen
- **THEN** zeigt das System die erste Seite mit „Seite 1 von n“ und erlaubt das Blättern zur nächsten Seite

#### Scenario: Bis-Datum schließt den ganzen Tag ein
- **WHEN** `GET /api/v1/events?to=2026-11-01` aufgerufen wird und eine Veranstaltung am 1. November um 10:00 Uhr beginnt
- **THEN** ist diese Veranstaltung im Ergebnis enthalten

#### Scenario: Gesamtzahl im Header
- **WHEN** `GET /api/v1/events?status=published&limit=10&offset=0` aufgerufen wird
- **THEN** liefert das System höchstens zehn Veranstaltungen und die Gesamtzahl der Treffer im Header `X-Total-Count`
