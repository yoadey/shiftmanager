## MODIFIED Requirements

### Requirement: E-Mail-Fluss für bekannte Adressen (K-006)
Gehört die eingegebene E-Mail einem aktiven Mitglied, MUST das System
ihm eine Bestätigungsaufforderung mit Link senden und die Eintragung
für es vormerken. Bestätigt das Mitglied per Link oder ein
Veranstaltungsleiter, Vorstand oder Admin, MUST das System die
Eintragung dem Mitglied zuordnen, ohne die Anwesenheit zu bestätigen;
der Link MUST eine Reservierung verbindlich machen. Bis dahin MUST die
Schicht den eingegebenen Anzeigenamen gemäß Namensmodus zeigen.

#### Scenario: Bekannte E-Mail
- **WHEN** am Kiosk die E-Mail-Adresse eines aktiven Mitglieds mit dem Anzeigenamen „Maxi“ eingegeben wird
- **THEN** sendet das System eine Bestätigungsaufforderung an diese Adresse, und die Schicht zeigt bis zur Bestätigung „Maxi“

#### Scenario: Bestätigung über den Link
- **WHEN** das Mitglied den Bestätigungslink öffnet
- **THEN** ist die Eintragung dem Mitglied zugeordnet, die Schicht zeigt seinen Namen, und der Anmeldestatus ist „angemeldet“

#### Scenario: Reservierung wird durch den Link verbindlich
- **WHEN** die Schicht voll war, die vorgemerkte Eintragung deshalb reserviert ist und das Mitglied den Bestätigungslink öffnet
- **THEN** ist die Eintragung dem Mitglied zugeordnet, hat den Status „angemeldet“ und verfällt nicht mehr

#### Scenario: Bestätigung durch Veranstaltungsleiter
- **WHEN** ein Veranstaltungsleiter die vorgemerkte Zuordnung in der Helferliste bestätigt
- **THEN** ist die Eintragung dem Mitglied zugeordnet

### Requirement: E-Mail-Fluss für unbekannte Adressen (K-007)
Ist die eingegebene E-Mail keinem aktiven Mitglied zugeordnet, MUST das
System die Eintragung als Gast mit dem eingegebenen Anzeigenamen
speichern, einen Bestätigungslink an die Adresse senden und die
Eintragung in der Helferliste als nicht zugeordnet kennzeichnen, damit
Veranstaltungsleiter sie zuordnen können (K-015). Öffnet der Gast den
Link, MUST das System die Eintragung verbindlich machen (Status
„angemeldet“, kein Verfall nach K-009), MUST NOT aber die Anwesenheit
bestätigen.

#### Scenario: Unbekannte E-Mail
- **WHEN** eine eingegebene E-Mail keinem Mitglied entspricht
- **THEN** speichert das System eine Gast-Eintragung mit dem Anzeigenamen, sendet den Bestätigungslink an die Adresse und kennzeichnet die Eintragung in der Helferliste als „Zuordnung offen“

#### Scenario: Gast bestätigt, danach Zuordnung
- **WHEN** der Gast den Bestätigungslink öffnet und ein Veranstaltungsleiter die Eintragung danach einem Mitglied zuordnet
- **THEN** hat die Eintragung den Status „angemeldet“, und dem Mitglied werden keine Stunden gutgeschrieben, bevor die Anwesenheit bestätigt ist

### Requirement: Zeitlich begrenzte, einmalige Bestätigungslinks (K-011)
Bestätigungslinks MUST zeitlich begrenzt (entsprechend K-010) und nur
einmal verwendbar sein. Ein Bestätigungslink MUST NOT den Status
„bestätigt“ (Anwesenheit) setzen; diesen setzt ausschließlich ein
Veranstaltungsleiter, Vorstand oder Admin.

#### Scenario: Bereits verwendeter Link
- **WHEN** ein Bestätigungslink ein zweites Mal aufgerufen wird
- **THEN** lehnt das System die erneute Bestätigung ab

#### Scenario: Link bestätigt keine Anwesenheit
- **WHEN** ein Bestätigungslink geöffnet wird
- **THEN** hat die Eintragung danach den Status „angemeldet“, nicht „bestätigt“
