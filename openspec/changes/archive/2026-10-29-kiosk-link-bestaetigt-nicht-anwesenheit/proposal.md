# Proposal

## Why

Der Bestätigungslink aus der Kiosk-E-Mail setzt eine Gast-Eintragung
heute auf „bestätigt“. „Bestätigt“ ist aber auch der Status, mit dem ein
Veranstaltungsleiter die Anwesenheit bestätigt, und nur solche
Eintragungen zählen als geleistete Stunden. Wird eine per Link bestätigte
Gast-Eintragung später einem Mitglied zugeordnet (K-015), bekommt das
Mitglied die Stunden gutgeschrieben, bevor die Schicht stattgefunden hat.

Umgekehrt ordnet der Link eine vorgemerkte Mitglieds-Eintragung (K-006)
nur zu und lässt eine Reservierung (volle Schicht) bestehen. Sie
verfällt dann nach K-009, obwohl das Mitglied bestätigt hat und die
Seite „verbindlich“ meldet.

## What Changes

- Der Bestätigungslink setzt nie mehr den Status „bestätigt“. Er
  bestätigt nur die E-Mail-Adresse bzw. die Zuordnung (K-006).
- Eine Reservierung wird durch den Link verbindlich: Status
  „angemeldet“, kein Verfallsdatum mehr. Das gilt für Gäste und für
  vorgemerkte Mitglieder gleichermaßen.
- „Bestätigt“ (Anwesenheit, zählt für Stunden) setzt ausschließlich ein
  Veranstaltungsleiter, Vorstand oder Admin.

## Capabilities

### Modified Capabilities

- `kiosk`: K-006, K-007, K-011

## Impact

- `RegistrationUsecase.ConfirmByLink`
- Bestätigungsseite: Text „verbindlich angemeldet“ statt „Schicht
  bestätigt“
- Keine Schema- oder API-Änderung
