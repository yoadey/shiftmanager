# Spec Delta

## ADDED Requirements

### Requirement: Abgeltungsprognose für das eigene Konto (G-011)
Das Stundenkonto eines Mitglieds (`GET /api/v1/hours/me`) MUST den
Abgeltungsbetrag `openFeeCents` enthalten, der für das aktive
Vereinsjahr fällig würde, wenn das Mitglied keine weiteren Stunden mehr
leistet: die Differenz aus Jahresziel und bestätigten Stunden, bewertet
nach der für das Mitglied geltenden Gebührenliste (G-001, G-002, G-004)
wie in der Jahresabrechnung (G-005). Ist keine Gebührenliste
konfiguriert, MUST das Feld fehlen.

#### Scenario: Prognose im eigenen Konto
- **WHEN** ein Mitglied mit 20 h Ziel, 17,5 h bestätigten Stunden und der Gebührenliste [10 €, 15 €] `GET /api/v1/hours/me` aufruft
- **THEN** enthält die Antwort `openFeeCents` = 3250 (1 h × 10 € + 1,5 h × 15 €)

#### Scenario: Ziel erreicht
- **WHEN** ein Mitglied mit erreichtem Jahresziel `GET /api/v1/hours/me` aufruft
- **THEN** enthält die Antwort `openFeeCents` = 0

### Requirement: Sichtbarkeit der Abgeltungsprognose (G-012)
Für das Stundenkonto eines anderen Mitglieds
(`GET /api/v1/hours/{memberId}`) MUST das System `openFeeCents` nur
Vorstand und Admin liefern.

#### Scenario: Fremdes Konto ohne Vorstandsrolle
- **WHEN** ein Mitglied ohne Vorstandsrolle `GET /api/v1/hours/{memberId}` für ein anderes Mitglied aufruft
- **THEN** enthält die Antwort kein `openFeeCents`

#### Scenario: Fremdes Konto durch Vorstand
- **WHEN** ein Vorstandsmitglied `GET /api/v1/hours/{memberId}` aufruft
- **THEN** enthält die Antwort `openFeeCents` des Mitglieds
