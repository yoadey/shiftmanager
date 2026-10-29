# Proposal

## Why

Am Kiosk im Vereinsheim sollen sich Helfer schnell eintragen können, ohne
Konto und ohne dass der Kiosk Mitgliederdaten preisgibt. Heute geht das
nur per Mitgliedersuche (gibt anonym Namen preis) oder per E-Mail mit
Bestätigungslink. Wer seine E-Mail nicht eingeben will oder kann, kommt
nicht weiter; Eintragungen mit unbekannter Adresse landen ohne
Zuordnung. Außerdem ist der Kiosk für jeden erreichbar, der die URL
kennt, und die Kiosk-Seite kann ohne Anmeldung weder Einstellungen noch
Vereinsdaten laden (sie fällt auf Demo-Werte zurück). Der
Bestätigungslink in der E-Mail zeigt fest auf `localhost:8080`, und eine
per Link bestätigte Kiosk-Anmeldung gilt als „anwesend bestätigt“, was
Stunden gutschreiben würde, bevor die Schicht stattgefunden hat.

## What Changes

- **Freitext-Name:** Am Kiosk kann man sich mit seinem vollständigen
  Namen eintragen. Die Eintragung ist zunächst keinem Mitglied
  zugeordnet.
- **E-Mail mit Anzeigename:** Alternativ gibt man E-Mail und Anzeigenamen
  ein. Gehört die Adresse einem Mitglied, erhält es eine E-Mail mit
  Bestätigungsaufforderung; bestätigt es oder bestätigt ein
  Veranstaltungsleiter/Vorstand/Admin, wird die Eintragung dem Mitglied
  zugeordnet. Bei unbekannter Adresse bleibt es eine Gast-Eintragung mit
  Bestätigungslink wie bisher. Der Kiosk verrät nicht, ob die Adresse
  bekannt ist.
- **Anzeige bis zur Zuordnung:** Bis dahin zeigt die Schicht den am Kiosk
  eingegebenen Namen, gemäß Namensmodus voll oder abgekürzt.
- **Zuordnung durch Veranstaltungsleiter/Vorstand/Admin:** Die
  Helferliste kennzeichnet offene Kiosk-Eintragungen. Beim Zuordnen schlägt
  das System Mitglieder nach Namensähnlichkeit vor, auch ohne exakten
  Treffer (z. B. „Max Mustermann“ ↔ „Maximilian Mustermann“, „Müller“ ↔
  „Mueller“). Das zugeordnete Mitglied wird per E-Mail informiert.
- **Kiosk-Link mit Token:** Vorstand/Admin erzeugen in den Einstellungen
  Kiosk-Links mit einstellbarer Gültigkeit (in Tagen), sehen aktive Links
  und können sie widerrufen. Der Kiosk ist nur noch mit gültigem Link
  nutzbar; der Bestätigungslink aus der E-Mail bleibt ohne Token
  erreichbar.
- **Kiosk-Konfiguration ohne Anmeldung:** `GET /api/v1/kiosk/config`
  liefert dem Kiosk Suche-an/aus sowie Vereinsname und Logo.
- **Korrekturen:** Bestätigungslink zeigt auf `BASE_URL` und die
  Bestätigungsseite der App; die Bestätigung einer Mitglieds-Zuordnung
  ändert den Anmeldestatus nicht mehr auf „bestätigt“.

## Capabilities

### New Capabilities

_Keine._

### Modified Capabilities

- `kiosk`: K-001, K-004, K-006, K-007 geändert; K-014 (Freitext-Name),
  K-015 (Zuordnung mit Vorschlägen), K-016 (Kiosk-Link mit Token) neu.

## Impact

- Datenbank: `registrations.pending_member_id`,
  `registrations.self_registered`, neue Tabelle `kiosk_links`
  (Migration `011_kiosk_links_assignment`, GORM-AutoMigrate), neue
  E-Mail-Vorlage `kiosk-assigned`.
- Backend: Kiosk-Token-Middleware, `KioskUsecase` (Links),
  `RegistrationUsecase` (Kiosk-Anmeldung, Bestätigung, Zuordnung,
  Vorschläge), Namensähnlichkeit in `domain`, Zeitleiste
  (`pendingMemberName`), neue Routen.
- Frontend: Kiosk-Seite (Token aus Link, drei Eintragungswege,
  Konfiguration), Helferliste (Kennzeichnung, Zuordnen-Dialog),
  Einstellungen (Kiosk-Links).
- Betrieb: Bestehende Kiosk-Geräte brauchen einmalig einen Kiosk-Link.
