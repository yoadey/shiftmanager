# Spec Delta

## ADDED Requirements

### Requirement: Helferliste direkt in der Schicht (SC-012)
In der Veranstaltungsansicht MUST Veranstaltungsleitern in jeder Schicht
die eingetragenen Helfer ohne zusätzliches Aufklappen angezeigt werden,
zusammen mit einer direkt sichtbaren Aktion „Helfer hinzufügen“, die
hervorgehoben ist, solange die Schicht freie Plätze hat. Eine eigene
Angabe der Anzahl Eingetragener neben dem Belegungsbalken MUST entfallen,
da der Balken sie bereits zeigt.

#### Scenario: Helfer sofort sichtbar
- **WHEN** ein Veranstaltungsleiter eine Veranstaltung öffnet, in deren Schicht zwei Helfer eingetragen sind
- **THEN** sieht er beide Helfer mit ihren Aktionen (Zeit, Anwesenheit, Entfernen) direkt in der Schicht, ohne vorher etwas aufzuklappen

#### Scenario: Helfer hinzufügen
- **WHEN** ein Veranstaltungsleiter eine Schicht mit freien Plätzen ansieht
- **THEN** ist „Helfer hinzufügen“ direkt in der Schicht als hervorgehobene Schaltfläche sichtbar und öffnet die Helfersuche
