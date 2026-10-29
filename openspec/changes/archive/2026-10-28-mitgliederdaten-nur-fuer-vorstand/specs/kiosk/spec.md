# Spec Delta

## MODIFIED Requirements

### Requirement: Konfigurierbare Mitgliedersuche (K-002)
Der Admin MUST konfigurieren können, ob im Kiosk-Modus eine
Mitgliedersuche erlaubt ist oder ausschließlich die E-Mail-Eingabe
verwendet wird. Ist die Suche deaktiviert, MUST
`GET /api/v1/kiosk/members` mit 403 antworten.

#### Scenario: Suche deaktiviert
- **WHEN** die Mitgliedersuche deaktiviert ist
- **THEN** bietet der Kiosk ausschließlich eine E-Mail-Eingabe an (siehe K-004)

#### Scenario: Suchanfrage trotz deaktivierter Suche
- **WHEN** die Mitgliedersuche deaktiviert ist und `GET /api/v1/kiosk/members?q=max` aufgerufen wird
- **THEN** antwortet das System mit 403 und liefert keine Namen

### Requirement: Suchergebnis-Format (K-003)
Ist die Suche aktiviert, MUST sie über Name oder Mitgliedsnummer erfolgen
und nur Vorname plus abgekürzten Nachnamen anzeigen. Die Antwort
(`GET /api/v1/kiosk/members`) MUST je Treffer ausschließlich ID und
diesen Namen enthalten, höchstens 20 Treffer liefern und MUST NOT über
E-Mail-Adressen suchen.

#### Scenario: Suchergebnis anzeigen
- **WHEN** ein Kiosk-Nutzer nach einem Namen sucht
- **THEN** zeigt das System Treffer nur mit Vorname und abgekürztem Nachnamen (NM-002-Format)

#### Scenario: Keine weiteren Mitgliederdaten
- **WHEN** `GET /api/v1/kiosk/members?q=max` aufgerufen wird
- **THEN** enthält die Antwort je Treffer nur `id` und `name`, insbesondere keine E-Mail-Adresse

#### Scenario: Keine Suche nach E-Mail-Adressen
- **WHEN** `GET /api/v1/kiosk/members?q=max@example.org` aufgerufen wird und diese Adresse einem Mitglied gehört
- **THEN** liefert das System keinen Treffer
