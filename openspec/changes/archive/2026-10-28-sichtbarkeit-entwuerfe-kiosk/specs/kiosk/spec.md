# Spec Delta

## MODIFIED Requirements

### Requirement: Konfigurierbare Mitgliedersuche (K-002)
Der Admin MUST konfigurieren können, ob im Kiosk-Modus eine
Mitgliedersuche erlaubt ist oder ausschließlich die E-Mail-Eingabe
verwendet wird. Ist die Suche deaktiviert, MUST
`GET /api/v1/kiosk/members` mit 403 antworten, und eine Kiosk-Anmeldung
mit Mitglieds-ID MUST mit 403 abgelehnt werden.

#### Scenario: Suche deaktiviert
- **WHEN** die Mitgliedersuche deaktiviert ist
- **THEN** bietet der Kiosk ausschließlich eine E-Mail-Eingabe an (siehe K-004)

#### Scenario: Suchanfrage trotz deaktivierter Suche
- **WHEN** die Mitgliedersuche deaktiviert ist und `GET /api/v1/kiosk/members?q=max` aufgerufen wird
- **THEN** antwortet das System mit 403 und liefert keine Namen

#### Scenario: Anmeldung per Mitglieds-ID trotz deaktivierter Suche
- **WHEN** die Mitgliedersuche deaktiviert ist und `POST /api/v1/kiosk/shifts/{id}/register` mit einer Mitglieds-ID aufgerufen wird
- **THEN** lehnt das System die Anmeldung mit 403 ab
