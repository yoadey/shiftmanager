# Spec Delta

## Purpose

Pflegt die Mitgliederliste unabhängig vom OIDC-Provider, sodass Stunden
auch für Mitglieder ohne aktives OIDC-Konto erfasst werden können.

## ADDED Requirements

### Requirement: Mitglieder anlegen, bearbeiten, deaktivieren (ML-001)
Vorstand und Admin MUST Mitglieder anlegen, bearbeiten und deaktivieren
können.

#### Scenario: Mitglied anlegen
- **WHEN** ein Vorstandsmitglied `POST /api/v1/members` mit gültigen Pflichtfeldern aufruft
- **THEN** legt das System das Mitglied an

#### Scenario: Fremdrolle ohne Zugriff
- **WHEN** ein Mitglied ohne Vorstandsrolle versucht, ein anderes Mitglied zu bearbeiten
- **THEN** lehnt das System die Anfrage ab

### Requirement: Pflichtfelder eines Mitglieds (ML-002)
Ein Mitglied MUST mindestens Vorname, Nachname, E-Mail und Eintrittsdatum
enthalten; ein Austrittsdatum ist optional.

#### Scenario: Anlage ohne Pflichtfeld
- **WHEN** beim Anlegen eines Mitglieds ein Pflichtfeld fehlt
- **THEN** lehnt das System die Anfrage ab

### Requirement: Eintrittsdatum-Default (ML-003)
Das Eintrittsdatum MUST beim Anlegen auf den Anlagezeitpunkt vorbelegt
sein und SHOULD manuell überschreibbar sein.

#### Scenario: Manuelles Eintrittsdatum
- **WHEN** ein Vorstandsmitglied beim Anlegen ein abweichendes Eintrittsdatum angibt
- **THEN** übernimmt das System den angegebenen Wert statt des Anlagezeitpunkts

### Requirement: Verknüpfung mit OIDC-Benutzer (ML-004)
Das System MUST ein Mitglied automatisch per E-Mail-Abgleich beim ersten
Login mit einem OIDC-Benutzer verknüpfen und SHOULD eine manuelle
Verknüpfung durch den Admin erlauben.

#### Scenario: Automatische Verknüpfung beim ersten Login
- **WHEN** sich ein OIDC-Nutzer erstmals anmeldet und dessen E-Mail einem bestehenden Mitglied ohne OIDC-Verknüpfung entspricht
- **THEN** verknüpft das System das Mitglied automatisch mit diesem OIDC-Konto

### Requirement: Mitglieder ohne OIDC-Verknüpfung (ML-005)
Mitglieder ohne OIDC-Verknüpfung MUST im Kiosk-Modus und in der manuellen
Stundenbuchung nutzbar sein.

#### Scenario: Manuelle Stundenbuchung ohne OIDC-Konto
- **WHEN** ein Vorstandsmitglied Stunden manuell für ein Mitglied ohne OIDC-Verknüpfung bucht
- **THEN** akzeptiert das System die Buchung

### Requirement: CSV-Import (ML-006)
Das System MUST einen CSV-Import der Mitgliederliste unterstützen, der
bestehende Einträge per E-Mail abgleicht (Update) und neue Einträge
anlegt.

#### Scenario: Import mit bekannter E-Mail
- **WHEN** eine importierte Zeile eine E-Mail enthält, die einem bestehenden Mitglied entspricht
- **THEN** aktualisiert das System das bestehende Mitglied statt ein Duplikat anzulegen

### Requirement: Import-Vorschau (ML-007)
Das System MUST vor der endgültigen Übernahme eine Vorschau der
Änderungen anzeigen.

#### Scenario: Vorschau vor Bestätigung
- **WHEN** eine CSV-Datei hochgeladen wird
- **THEN** zeigt das System an, welche Mitglieder neu angelegt und welche aktualisiert würden, bevor der Import bestätigt wird

### Requirement: CSV-Export (ML-008)
Das System MUST einen CSV-Export der Mitgliederliste bereitstellen.

#### Scenario: Export anfordern
- **WHEN** ein Vorstandsmitglied `GET /api/v1/members/export` aufruft
- **THEN** liefert das System eine CSV-Datei mit der aktuellen Mitgliederliste

### Requirement: Änderungen im Audit-Log (ML-009)
Änderungen an Mitgliedern MUST im Audit-Log protokolliert werden.

#### Scenario: Protokollierung einer Deaktivierung
- **WHEN** ein Mitglied deaktiviert wird
- **THEN** legt das System einen Audit-Eintrag mit Aktion, Nutzer und Zeitpunkt an

### Requirement: Leseberechtigung der Mitgliederliste (ML-010)
Die Mitgliederliste MUST für jeden authentifizierten Nutzer lesbar sein
(z. B. für die Namensanzeige in Schichtlisten); Schreiboperationen sind
Vorstand und Admin vorbehalten.

#### Scenario: Lesezugriff durch einfaches Mitglied
- **WHEN** ein authentifiziertes Mitglied ohne Vorstandsrolle `GET /api/v1/members` aufruft
- **THEN** liefert das System die Liste
