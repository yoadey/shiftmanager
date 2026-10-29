# Spec Delta

## ADDED Requirements

### Requirement: Sichtbarkeit der Stundenkonten (SA-011)
Jedes Mitglied MUST sein eigenes Stundenkonto einsehen können. Das
Stundenkonto eines anderen Mitglieds (`GET /api/v1/hours/{memberId}`,
`GET /api/v1/hours/account`) und die Jahresübersicht aller Konten
(`GET /api/v1/hours/summary`) MUST ausschließlich Vorstand und Admin
lesen dürfen; anderen Nutzern, auch Veranstaltungsleitern, MUST das
System mit 403 antworten.

#### Scenario: Fremdes Stundenkonto durch Mitglied
- **WHEN** ein Mitglied `GET /api/v1/hours/{memberId}` für ein anderes Mitglied aufruft
- **THEN** lehnt das System die Anfrage mit 403 ab

#### Scenario: Eigenes Stundenkonto über die ID
- **WHEN** ein Mitglied `GET /api/v1/hours/{memberId}` mit seiner eigenen ID aufruft
- **THEN** liefert das System sein Stundenkonto

#### Scenario: Jahresübersicht durch Veranstaltungsleiter
- **WHEN** ein Veranstaltungsleiter ohne Vorstandsrolle `GET /api/v1/hours/summary` aufruft
- **THEN** lehnt das System die Anfrage mit 403 ab

#### Scenario: Vorstand sieht fremdes Konto
- **WHEN** ein Vorstandsmitglied `GET /api/v1/hours/{memberId}` aufruft
- **THEN** liefert das System das Stundenkonto des Mitglieds
