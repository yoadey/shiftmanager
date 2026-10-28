# Spec Delta

## MODIFIED Requirements

### Requirement: CSV-Export (ML-008)
Das System MUST einen CSV-Export der Mitgliederliste bereitstellen. Der
Export MUST Vorstand und Admin vorbehalten sein.

#### Scenario: Export anfordern
- **WHEN** ein Vorstandsmitglied `GET /api/v1/members/export` aufruft
- **THEN** liefert das System eine CSV-Datei mit der aktuellen Mitgliederliste

#### Scenario: Export ohne Vorstandsrolle
- **WHEN** ein Mitglied oder Veranstaltungsleiter ohne Vorstandsrolle `GET /api/v1/members/export` aufruft
- **THEN** lehnt das System die Anfrage mit 403 ab

### Requirement: Leseberechtigung der Mitgliederliste (ML-010)
Die Mitgliederliste (`GET /api/v1/members`) und die Datensätze anderer
Mitglieder (`GET /api/v1/members/{id}`) MUST ausschließlich Vorstand und
Admin lesen dürfen; anderen Nutzern MUST das System mit 403 antworten.
Den eigenen Datensatz MUST jedes Mitglied abrufen können.
Schreiboperationen sind Vorstand und Admin vorbehalten.

#### Scenario: Lesezugriff durch einfaches Mitglied
- **WHEN** ein authentifiziertes Mitglied ohne Vorstandsrolle `GET /api/v1/members` aufruft
- **THEN** lehnt das System die Anfrage mit 403 ab

#### Scenario: Wartende Registrierungen ohne Vorstandsrolle
- **WHEN** ein authentifiziertes Mitglied ohne Vorstandsrolle `GET /api/v1/members?status=pending` aufruft
- **THEN** lehnt das System die Anfrage mit 403 ab

#### Scenario: Fremdes Mitglied ohne Vorstandsrolle
- **WHEN** ein Mitglied ohne Vorstandsrolle `GET /api/v1/members/{id}` für ein anderes Mitglied aufruft
- **THEN** lehnt das System die Anfrage mit 403 ab

#### Scenario: Eigener Datensatz
- **WHEN** ein Mitglied `GET /api/v1/members/{id}` mit seiner eigenen ID aufruft
- **THEN** liefert das System seinen Datensatz

#### Scenario: Vorstand liest die Liste
- **WHEN** ein Vorstandsmitglied `GET /api/v1/members` aufruft
- **THEN** liefert das System die Liste einschließlich wartender Registrierungen

## ADDED Requirements

### Requirement: Namenssuche für die Helferzuordnung (ML-014)
Damit Veranstaltungsleiter Mitglieder einer Schicht zuordnen können
(SC-010), MUST `GET /api/v1/members/lookup?q=` ab der Rolle
Veranstaltungsleiter aktive, freigeschaltete Mitglieder nach Name oder
E-Mail suchen und je Treffer ausschließlich ID und vollständigen Namen
liefern, höchstens 20 Treffer. Suchbegriffe unter zwei Zeichen MUST mit
400 abgelehnt werden; Mitgliedern ohne diese Rolle MUST das System mit
403 antworten.

#### Scenario: Veranstaltungsleiter sucht einen Helfer
- **WHEN** ein Veranstaltungsleiter `GET /api/v1/members/lookup?q=must` aufruft und „Max Mustermann“ aktiv ist
- **THEN** enthält die Antwort einen Eintrag mit ID und Name „Max Mustermann“, aber keine E-Mail, Rolle oder Stundenziel

#### Scenario: Suche durch einfaches Mitglied
- **WHEN** ein Mitglied ohne Veranstaltungsleiterrolle `GET /api/v1/members/lookup?q=must` aufruft
- **THEN** lehnt das System die Anfrage mit 403 ab
