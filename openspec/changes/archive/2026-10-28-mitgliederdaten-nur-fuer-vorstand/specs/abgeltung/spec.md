# Spec Delta

## MODIFIED Requirements

### Requirement: Sichtbarkeit der Abgeltungsprognose (G-012)
Die Abgeltungsprognose `openFeeCents` eines anderen Mitglieds MUST nur
Vorstand und Admin zugänglich sein; sie ist Teil des Stundenkontos, das
nur diese Rollen für fremde Mitglieder abrufen dürfen (SA-011).

#### Scenario: Fremdes Konto ohne Vorstandsrolle
- **WHEN** ein Mitglied ohne Vorstandsrolle `GET /api/v1/hours/{memberId}` für ein anderes Mitglied aufruft
- **THEN** lehnt das System die Anfrage mit 403 ab und liefert keine Prognose

#### Scenario: Fremdes Konto durch Vorstand
- **WHEN** ein Vorstandsmitglied `GET /api/v1/hours/{memberId}` aufruft
- **THEN** enthält die Antwort `openFeeCents` des Mitglieds
