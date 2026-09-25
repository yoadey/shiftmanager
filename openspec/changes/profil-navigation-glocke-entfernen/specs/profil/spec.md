# Spec Delta

## MODIFIED Requirements

### Requirement: Erreichbarkeit des Profils (PR-005)
Das Profil MUST auf dem Desktop über eine klickbare Profilkarte unten
links in der Sidebar-Navigation erreichbar sein. Der `profil`-Tab MUST
NOT zusätzlich in der Desktop-Sidebar-Navigationsliste erscheinen (die
Profilkarte ersetzt ihn dort). In der mobilen Ansicht, die keine
Sidebar-Profilkarte hat, MUST das Profil weiterhin über den `profil`-Tab
in der unteren Navigationsleiste erreichbar sein. Ein Glocken-Icon oder
sonstiger Benachrichtigungs-Shortcut zum Profil MUST NOT verwendet
werden.

#### Scenario: Über die Desktop-Profilkarte navigieren
- **WHEN** ein Mitglied auf dem Desktop die Profilkarte unten links in der Sidebar anklickt
- **THEN** öffnet das System die eigene Profilseite

#### Scenario: Kein redundanter Profil-Tab auf Desktop
- **WHEN** die Desktop-Sidebar-Navigation gerendert wird
- **THEN** enthält sie keinen separaten `profil`-Tab-Eintrag

#### Scenario: Über den Profil-Tab navigieren
- **WHEN** ein Mitglied in der mobilen Ansicht den `profil`-Tab in der unteren Navigationsleiste anklickt (die einzige Ansicht, in der dieser Tab noch erscheint)
- **THEN** öffnet das System die eigene Profilseite

#### Scenario: Über den Dashboard-Shortcut navigieren
- **WHEN** ein Mitglied sein Dashboard öffnet
- **THEN** zeigt das System dort keinen Glocken-Shortcut (oder sonstigen Benachrichtigungs-Shortcut) mehr an, der zum Profil führt — dieser Zugang wurde ersatzlos entfernt
