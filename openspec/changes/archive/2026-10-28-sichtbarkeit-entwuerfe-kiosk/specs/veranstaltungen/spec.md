# Spec Delta

## MODIFIED Requirements

### Requirement: Sichtbarkeit nur für veröffentlichte Veranstaltungen (V-006)
Für Mitglieder MUST nur veröffentlichte Veranstaltungen sichtbar sein;
Entwürfe MUST NOT in Liste, Detail oder Zeitleiste erscheinen.
Abgeschlossene Veranstaltungen bleiben für die eigene Schicht-Historie
sichtbar. Der Kiosk MUST ausschließlich veröffentlichte Veranstaltungen
zeigen. Veranstaltungsleiter, Vorstand und Admin sehen alle
Veranstaltungen. Eine nicht sichtbare Veranstaltung MUST wie eine
fehlende mit 404 beantwortet werden.

#### Scenario: Entwurf nicht sichtbar
- **WHEN** ein Mitglied `GET /api/v1/events` aufruft
- **THEN** liefert das System Veranstaltungen im Status Entwurf nicht mit

#### Scenario: Entwurf direkt abgerufen
- **WHEN** ein Mitglied oder der Kiosk die Zeitleiste eines Entwurfs abruft
- **THEN** antwortet das System mit 404

#### Scenario: Veranstaltungsleiter sieht Entwürfe
- **WHEN** ein Veranstaltungsleiter `GET /api/v1/events?status=draft` aufruft
- **THEN** liefert das System die Entwürfe
