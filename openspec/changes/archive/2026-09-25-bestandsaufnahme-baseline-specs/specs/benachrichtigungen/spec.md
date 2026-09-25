# Spec Delta

## Purpose

Regelt die automatisierten E-Mail-Benachrichtigungen des Systems, ihre
Auslöser, Anpassbarkeit und Protokollierung.

## ADDED Requirements

### Requirement: Opt-out für Erinnerungsmails (N-001)
Ein Mitglied MUST Erinnerungsmails abbestellen können; Pflicht-Mails
(z. B. Schicht-Bestätigung) MUST NOT abbestellbar sein.

#### Scenario: Erinnerungen abbestellen
- **WHEN** ein Mitglied `PUT /api/v1/members/me/preferences` mit deaktivierten Erinnerungen aufruft
- **THEN** versendet das System künftig keine Erinnerungsmails mehr an dieses Mitglied, wohl aber weiterhin Pflicht-Mails

### Requirement: Konfigurierbarer Versandzeitpunkt (N-002)
Der Versandzeitpunkt für Erinnerungen MUST global konfigurierbar sein.

#### Scenario: Versandzeit ändern
- **WHEN** der Admin Uhrzeit oder Vorlaufwochen für Erinnerungen ändert
- **THEN** verwendet das System die neuen Werte für künftige Erinnerungen

### Requirement: Versand über SMTP mit Transportverschlüsselung (N-003)
Der E-Mail-Versand MUST über SMTP mit TLS/STARTTLS erfolgen.

#### Scenario: Versand über verschlüsselte Verbindung
- **WHEN** das System eine E-Mail versendet
- **THEN** baut es die SMTP-Verbindung mit TLS/STARTTLS auf

### Requirement: Protokollierung fehlgeschlagener E-Mails (N-004)
Fehlgeschlagene E-Mails MUST protokolliert und manuell erneut versendbar
sein.

#### Scenario: Fehlgeschlagener Versand erneut auslösen
- **WHEN** ein Vorstandsmitglied im E-Mail-Protokoll einen fehlgeschlagenen Versand über `POST /api/v1/settings/email-log/{id}/resend` erneut auslöst
- **THEN** versucht das System den Versand erneut

### Requirement: Anpassbare E-Mail-Vorlagen (N-005)
E-Mail-Vorlagen MUST über die Administrationsoberfläche anpassbar sein.

#### Scenario: Vorlage bearbeiten
- **WHEN** ein Vorstandsmitglied eine E-Mail-Vorlage über `PUT /api/v1/settings/email-templates/{name}` bearbeitet
- **THEN** verwendet das System die angepasste Vorlage für künftige Versände dieses Typs

### Requirement: Definierte E-Mail-Typen und Auslöser (N-006)
Das System MUST mindestens folgende E-Mail-Typen mit ihren Auslösern
unterstützen: Kiosk-Bestätigung (bei Kiosk-Eintragung), Schicht-Bestätigung
(bei Anmeldung), Schicht-Abmeldung (bei Abmeldung), Erinnerung eine Woche
und einen Tag vor Schichtbeginn, Schicht abgesagt (bei Organisator-Absage),
Jahresabrechnung (Ende Vereinsjahr), Warnung Fehlstunden (X Wochen vor
Jahresende), Neue Veranstaltung (bei Veröffentlichung, Opt-in),
Unterschreitung Minimum (an den Veranstaltungsleiter) und Stunden
bestätigt (an das Mitglied).

#### Scenario: Schicht-Bestätigung nach Anmeldung
- **WHEN** ein Mitglied sich erfolgreich für eine Schicht anmeldet
- **THEN** sendet das System eine Schicht-Bestätigungs-Mail an das Mitglied

#### Scenario: Erinnerung eine Woche vorher
- **WHEN** eine Schicht in sieben Tagen beginnt
- **THEN** sendet das System allen angemeldeten Helfern eine Erinnerung
