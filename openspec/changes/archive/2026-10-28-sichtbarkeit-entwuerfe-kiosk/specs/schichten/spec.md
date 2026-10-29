# Spec Delta

## MODIFIED Requirements

### Requirement: Helfernamen pro Schicht (SC-011)
In einer Veranstaltung MUST jedes Mitglied die Namen der in der
jeweiligen Schicht eingetragenen Helfer sehen können, nach Namensmodus
(NM-001 bis NM-006). Weitere Daten fremder Anmeldungen (Mitglieds-ID,
voller Gastname, E-Mail eines Gastes, Kommentar, gebuchte Stunden,
Reservierungsfrist) MUST das System Nutzern unterhalb der Rolle
Veranstaltungsleiter und im Kiosk NOT ausliefern; die eigene Anmeldung
bleibt vollständig sichtbar.

#### Scenario: Mitglied sieht Helfernamen
- **WHEN** ein Mitglied eine Veranstaltung öffnet, in deren Schicht „Max Mustermann“ eingetragen ist, und der Namensmodus „Abgekürzter Name“ aktiv ist
- **THEN** zeigt die Schicht den Helfer als „Max M.“

#### Scenario: Keine fremden Kommentare und E-Mails
- **WHEN** ein Mitglied die Zeitleiste einer Veranstaltung abruft, in der ein anderes Mitglied mit Kommentar und ein Gast mit E-Mail-Adresse eingetragen sind
- **THEN** enthält die Antwort weder den fremden Kommentar noch die E-Mail des Gastes

#### Scenario: Keine fremden Mitglieds-IDs
- **WHEN** der Kiosk oder ein Mitglied die Zeitleiste einer Veranstaltung abruft, in der ein anderes Mitglied eingetragen ist
- **THEN** enthält dessen Anmeldung keine Mitglieds-ID

#### Scenario: Veranstaltungsleiter sieht die vollständige Helferliste
- **WHEN** ein Veranstaltungsleiter die Zeitleiste derselben Veranstaltung abruft
- **THEN** enthält die Antwort volle Namen, Kommentare und die E-Mail des Gastes
