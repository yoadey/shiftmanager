# datenschutz-sicherheit Specification

## Purpose
Stellt den DSGVO-konformen Betrieb sicher: Transportsicherheit,
Auskunfts- und Löschrecht, Passwortlosigkeit, Token-Lebensdauer und
Schutz vor gängigen Web-Angriffen.

## Requirements

### Requirement: DSGVO-konformer Betrieb (DS-001)
Das System MUST insgesamt DSGVO-konform betrieben werden können — siehe
die folgenden Einzelanforderungen dieser Capability.

#### Scenario: Gesamtkonformität
- **WHEN** alle Einzelanforderungen dieser Capability (DS-002 bis DS-008) erfüllt sind
- **THEN** gilt der Betrieb als DSGVO-konform im Rahmen dieser Spezifikation

### Requirement: Verschlüsselte Übertragung (DS-002)
Die Kommunikation mit dem System MUST ausschließlich über HTTPS/TLS
erfolgen (Deployment-Anforderung, Reverse Proxy/TLS-Terminierung).

#### Scenario: Unverschlüsselter Zugriffsversuch
- **WHEN** ein Client versucht, das Deployment über unverschlüsseltes HTTP zu erreichen
- **THEN** MUST der Betrieb dies über die Infrastruktur (Reverse Proxy) verhindern oder auf HTTPS umleiten

### Requirement: Auskunftsrecht (DS-003)
Ein Mitglied MUST seine gespeicherten Daten abrufen können; fremde Daten
MUST nur Vorstand und Admin abrufen können.

#### Scenario: Eigene Daten exportieren
- **WHEN** ein Mitglied `GET /api/v1/members/{id}/export-data` für die eigene ID aufruft
- **THEN** liefert das System die gespeicherten Daten dieses Mitglieds

#### Scenario: Fremde Daten ohne Vorstandsrolle
- **WHEN** ein Mitglied ohne Vorstandsrolle die Daten eines anderen Mitglieds exportieren will
- **THEN** lehnt das System die Anfrage ab

### Requirement: Recht auf Löschung (DS-004)
Ein Mitglied MUST eine Löschung seiner Daten beantragen können (Recht auf
Vergessen, mit Ausnahmen für Aufbewahrungspflichten); die Durchführung
MUST Vorstand/Admin vorbehalten sein.

#### Scenario: Löschung durch Vorstand
- **WHEN** ein Vorstandsmitglied `POST /api/v1/members/{id}/gdpr-delete` aufruft
- **THEN** anonymisiert das System die personenbezogenen Daten dieses Mitglieds, soweit keine Aufbewahrungspflicht entgegensteht

### Requirement: Keine Passwortspeicherung (DS-005)
Das System MUST NOT Passwörter selbst speichern; die Authentifizierung
MUST vollständig an den OIDC-Provider ausgelagert sein (siehe `auth`).

#### Scenario: Kein Passwortfeld
- **WHEN** ein Mitglied angelegt wird
- **THEN** verlangt das System kein Passwort und speichert keines

### Requirement: Kurze JWT-Lebensdauer (DS-006)
JWT-Tokens MUST eine konfigurierbare, kurze Lebensdauer haben.

#### Scenario: Ablauf eines Tokens
- **WHEN** die konfigurierte JWT-Lebensdauer (`JWT_EXPIRATION`) überschritten wird
- **THEN** lehnt das System Anfragen mit diesem Token ab und der Client muss sich über den Refresh-Flow (A-004) oder erneut anmelden

### Requirement: Schutz vor gängigen Web-Angriffen (DS-007)
Das System MUST vor SQL-Injection, XSS und CSRF über Framework-Features
geschützt sein.

#### Scenario: Parametrisierte Datenbankzugriffe
- **WHEN** Nutzereingaben in eine Datenbankabfrage einfließen
- **THEN** verwendet das System ausschließlich parametrisierte Queries (ORM), keine String-Konkatenation

### Requirement: Kryptografisch zufällige, einmalige Bestätigungslinks (DS-008)
Kiosk-Bestätigungslinks MUST kryptografisch zufällig und nur einmal
verwendbar sein (siehe `kiosk` K-011).

#### Scenario: Vorhersagbarkeit ausgeschlossen
- **WHEN** ein Bestätigungslink erzeugt wird
- **THEN** verwendet das System einen kryptografisch sicheren Zufallswert als Token
