# mitgliederverwaltung Specification

## Purpose
Pflegt die Mitgliederliste unabhängig vom OIDC-Provider, sodass Stunden
auch für Mitglieder ohne aktives OIDC-Konto erfasst werden können.

## Requirements

### Requirement: Mitglieder anlegen, bearbeiten, deaktivieren (ML-001)
Vorstand und Admin MUST Mitglieder anlegen, bearbeiten und deaktivieren
können. Eine Deaktivierung MUST das Austrittsdatum setzen (sofern noch
keines gesetzt ist), damit ein ausgetretenes Mitglied von einer
Registrierung, die auf Freischaltung wartet (ML-011), unterscheidbar
bleibt. Ein ausgetretenes Mitglied MUST durch Vorstand oder Admin
reaktiviert werden können; die Reaktivierung entfernt das
Austrittsdatum.

#### Scenario: Mitglied anlegen
- **WHEN** ein Vorstandsmitglied `POST /api/v1/members` mit gültigen Pflichtfeldern aufruft
- **THEN** legt das System das Mitglied an

#### Scenario: Fremdrolle ohne Zugriff
- **WHEN** ein Mitglied ohne Vorstandsrolle versucht, ein anderes Mitglied zu bearbeiten
- **THEN** lehnt das System die Anfrage ab

#### Scenario: Deaktivierung über die Bearbeitung setzt Austrittsdatum
- **WHEN** ein Vorstandsmitglied ein aktives Mitglied per `PUT /api/v1/members/{id}` mit `isActive=false` deaktiviert
- **THEN** setzt das System das Austrittsdatum auf den aktuellen Zeitpunkt und das Mitglied erscheint nicht als wartende Registrierung

#### Scenario: Ausgetretenes Mitglied reaktivieren
- **WHEN** ein Vorstandsmitglied in der Detailansicht eines ausgetretenen Mitglieds „Reaktivieren“ wählt
- **THEN** ist das Mitglied wieder aktiv und hat kein Austrittsdatum mehr

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
Vorstand und Admin vorbehalten. Registrierungen, die auf Freischaltung
warten (ML-011), MUST NOT Nutzern ohne Vorstandsrolle geliefert werden.

#### Scenario: Lesezugriff durch einfaches Mitglied
- **WHEN** ein authentifiziertes Mitglied ohne Vorstandsrolle `GET /api/v1/members` aufruft
- **THEN** liefert das System die Liste ohne wartende Registrierungen

#### Scenario: Wartende Registrierungen ohne Vorstandsrolle
- **WHEN** ein authentifiziertes Mitglied ohne Vorstandsrolle `GET /api/v1/members?status=pending` aufruft
- **THEN** lehnt das System die Anfrage mit 403 ab

### Requirement: Selbstregistrierung mit Freischaltung (ML-011)
Meldet sich ein OIDC-Nutzer an, der keinem Mitglied zugeordnet werden
kann (weder über die OIDC-Verknüpfung noch per E-Mail, ML-004), MUST das
System ihn als inaktives Mitglied mit Rolle `mitglied`, Name und E-Mail
aus den OIDC-Claims und ohne Austrittsdatum anlegen — es sei denn, er ist
Bootstrap-Admin (A-005) oder erhält über die Rollenzuordnung (A-006) eine
Rolle. Ein solches Konto gilt als „wartet auf Freischaltung“ (inaktiv,
ohne Austrittsdatum). Der Nutzer MUST NOT eingeloggt werden, solange das
Konto wartet, und MUST einen Hinweis erhalten, dass sein Zugang auf
Freischaltung durch den Vorstand wartet. Vorstand und Admin MUST eine
wartende Registrierung freischalten (aktivieren) oder ablehnen können;
eine Ablehnung deaktiviert das Konto mit Austrittsdatum, sodass es nicht
mehr als wartend erscheint.

#### Scenario: Unbekannter Nutzer meldet sich an
- **WHEN** sich ein OIDC-Nutzer anmeldet, dessen Subject und E-Mail keinem Mitglied zugeordnet sind, und keine Rolle über A-006 erhält
- **THEN** legt das System ein inaktives Mitglied ohne Austrittsdatum an, stellt kein JWT aus und zeigt den Hinweis „wartet auf Freischaltung“

#### Scenario: Registrierung freischalten
- **WHEN** ein Vorstandsmitglied bei einer wartenden Registrierung „Freischalten“ wählt
- **THEN** ist das Mitglied aktiv und kann sich beim nächsten Login anmelden

#### Scenario: Registrierung ablehnen
- **WHEN** ein Vorstandsmitglied bei einer wartenden Registrierung „Ablehnen“ wählt und die Rückfrage bestätigt
- **THEN** deaktiviert das System das Konto mit Austrittsdatum, es erscheint nicht mehr als wartend und ein erneuter Login wird als „inaktiv“ abgewiesen

#### Scenario: Ablehnen nur mit Bestätigung
- **WHEN** ein Vorstandsmitglied „Ablehnen“ wählt, die Rückfrage aber verneint
- **THEN** bleibt die Registrierung unverändert wartend

### Requirement: Paginierte Mitgliederliste (ML-012)
`GET /api/v1/members` MUST die Mitglieder nach Nachname und Vorname
sortiert seitenweise liefern (`limit`, Standard 50; `offset`) und die
Gesamtzahl der Treffer unabhängig von `limit`/`offset` im Header
`X-Total-Count` angeben; `limit=0` MUST alle Treffer liefern. Die Suche
(`search`) MUST serverseitig über Vorname, Nachname, vollständigen Namen
und E-Mail erfolgen (Groß-/Kleinschreibung egal). Der Parameter `status`
MUST nach Kontostatus filtern: `active` (aktiv), `pending` (wartet auf
Freischaltung), `left` (ausgetreten) und `approved` (alle außer wartenden
Registrierungen); ein unbekannter Wert MUST mit 400 abgelehnt werden.
Die Mitgliederübersicht MUST die Liste seitenweise anzeigen, den Wechsel
zur vorherigen und nächsten Seite ermöglichen, die aktuelle Seite und
Suche in der URL halten (Zurück-Navigation landet auf derselben Seite)
und ausgetretene Mitglieder als „Ausgetreten“ kennzeichnen.

#### Scenario: Seite abrufen
- **WHEN** `GET /api/v1/members?limit=25&offset=25` bei 60 Mitgliedern aufgerufen wird
- **THEN** liefert das System die Mitglieder 26 bis 50 in Namensreihenfolge und `X-Total-Count: 60`

#### Scenario: Alle Mitglieder für Namensauflösung
- **WHEN** `GET /api/v1/members?limit=0` aufgerufen wird
- **THEN** liefert das System alle passenden Mitglieder ohne Seitenbegrenzung

#### Scenario: Suche nach vollständigem Namen
- **WHEN** `GET /api/v1/members?search=max muster` aufgerufen wird und ein Mitglied „Max Mustermann“ existiert
- **THEN** ist dieses Mitglied im Ergebnis enthalten, auch wenn es nicht auf der ersten Seite der ungefilterten Liste stünde

#### Scenario: Wartende Registrierungen filtern
- **WHEN** `GET /api/v1/members?status=pending` aufgerufen wird
- **THEN** enthält das Ergebnis genau die inaktiven Mitglieder ohne Austrittsdatum, aber keine ausgetretenen

#### Scenario: Unbekannter Status
- **WHEN** `GET /api/v1/members?status=foo` aufgerufen wird
- **THEN** antwortet das System mit 400

#### Scenario: Seitenwechsel in der Übersicht
- **WHEN** ein Vorstandsmitglied in der Mitgliederübersicht bei mehr Mitgliedern als einer Seite „Weiter“ wählt
- **THEN** zeigt die Übersicht die nächste Seite, die URL enthält `?seite=2`, und „Zurück“ im Browser nach dem Öffnen eines Mitglieds kehrt zu dieser Seite zurück

### Requirement: Hinweis auf wartende Registrierungen (ML-013)
Solange Registrierungen auf Freischaltung warten (ML-011), MUST das
System sie Vorstand und Admin anzeigen: in der Mitgliederübersicht in
einem eigenen Abschnitt oberhalb der Liste mit „Freischalten“ und
„Ablehnen“, und auf allen anderen Seiten der App als Hinweis oben mit
der Anzahl wartender Registrierungen und einem Link zur
Mitgliederübersicht. Nutzern ohne Vorstandsrolle MUST NOT dieser Hinweis
angezeigt werden.

#### Scenario: Hinweis für den Vorstand
- **WHEN** zwei Registrierungen warten und ein Vorstandsmitglied eine beliebige Seite außer der Mitgliederübersicht öffnet
- **THEN** zeigt die App oben den Hinweis „2 neue Mitglieder warten auf Freischaltung“, der zur Mitgliederübersicht führt

#### Scenario: Abschnitt in der Mitgliederübersicht
- **WHEN** ein Vorstandsmitglied die Mitgliederübersicht öffnet, während Registrierungen warten
- **THEN** stehen diese unabhängig von Seite und Suche in einem Abschnitt „Warten auf Freischaltung“ oberhalb der paginierten Liste und nicht zusätzlich in der Liste

#### Scenario: Kein Hinweis ohne Vorstandsrolle
- **WHEN** ein Veranstaltungsleiter oder Mitglied die App nutzt, während Registrierungen warten
- **THEN** zeigt die App keinen Hinweis auf wartende Registrierungen

#### Scenario: Kein Hinweis ohne wartende Registrierungen
- **WHEN** keine Registrierung wartet
- **THEN** zeigt die App weder Hinweis noch Abschnitt an
