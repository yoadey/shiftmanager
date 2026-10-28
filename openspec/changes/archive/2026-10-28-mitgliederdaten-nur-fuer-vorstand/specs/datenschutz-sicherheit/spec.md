# Spec Delta

## ADDED Requirements

### Requirement: Bestätigungstoken nur per E-Mail (DS-009)
Der Token eines Bestätigungslinks (K-011) MUST ausschließlich per E-Mail
an den Empfänger gehen und MUST NOT in einer API-Antwort enthalten sein,
insbesondere nicht in Zeitleisten oder Anmeldebestätigungen.

#### Scenario: Zeitleiste ohne Token
- **WHEN** die Zeitleiste einer Veranstaltung mit einer unbestätigten Kiosk-Anmeldung abgerufen wird, ob angemeldet oder über den Kiosk
- **THEN** enthält die Antwort keinen Bestätigungstoken
