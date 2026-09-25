# profil Specification

## Purpose
Stellt jedem Mitglied eine eigene Profilseite mit
Namensanzeige-Kontext, Erinnerungspräferenzen und Logout bereit.

## Requirements

### Requirement: Eigene Profilseite (PR-001)
Jedes Mitglied MUST eine eigene Profilseite mit seinen Stammdaten und
seinem Namensanzeige-Kontext einsehen können.

#### Scenario: Profil öffnen
- **WHEN** ein Mitglied seine Profilseite öffnet
- **THEN** zeigt das System seine eigenen Stammdaten und den aktuellen Namensanzeige-Kontext an

### Requirement: Erinnerungspräferenzen im Profil (PR-002)
Das Profil MUST die Verwaltung der eigenen Erinnerungspräferenzen (siehe
`benachrichtigungen` N-001) ermöglichen.

#### Scenario: Präferenzen im Profil ändern
- **WHEN** ein Mitglied auf der Profilseite Erinnerungen deaktiviert
- **THEN** übernimmt das System diese Präferenz

### Requirement: Logout im Profil (PR-003)
Das Profil MUST eine Möglichkeit bieten, die aktuelle Session zu beenden
(Logout).

#### Scenario: Logout auslösen
- **WHEN** ein Mitglied auf der Profilseite "Sitzung beenden" auswählt
- **THEN** beendet das System die aktuelle Session und leitet zum Login weiter

### Requirement: DSGVO-Löschantrag aus dem Profil (PR-004)
Das Profil MUST einen Weg bieten, eine Löschung der eigenen Daten beim
Vorstand zu beantragen (siehe `datenschutz-sicherheit` DS-004), da die
Selbstlöschung aus Aufbewahrungspflichten heraus nicht möglich ist.

#### Scenario: Löschantrag stellen
- **WHEN** ein Mitglied auf der Profilseite eine Löschung beantragt
- **THEN** informiert das System, dass die Anonymisierung durch den Vorstand bestätigt werden muss, und leitet den Antrag entsprechend weiter

### Requirement: Erreichbarkeit des Profils (PR-005)
Das Profil MUST über den `profil`-Tab in der regulären Navigation
(alle Rollen) sowie über einen Profil-Shortcut im Verwaltungs-Dashboard
erreichbar sein.

#### Scenario: Über den Profil-Tab navigieren
- **WHEN** ein Mitglied den `profil`-Tab in der Navigation anklickt
- **THEN** öffnet das System die eigene Profilseite

#### Scenario: Über den Dashboard-Shortcut navigieren
- **WHEN** ein Mitglied im Mitglieder-Dashboard den Profil-Shortcut (Glocken-Symbol oben rechts) anklickt
- **THEN** öffnet das System die eigene Profilseite
