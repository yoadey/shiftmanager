# Tasks

## 1. Domain & Persistenz

- [x] 1.1 `domain.NameSimilarity` mit Tests
- [x] 1.2 `Registration.PendingMemberID`, `SelfRegistered`, `PendingMemberName`; `domain.KioskLink`
- [x] 1.3 Modelle/Repos: `registrations.pending_member_id`, `self_registered`; `kiosk_links` (`KioskLinkRepository`); Migration `011_kiosk_links_assignment` inkl. Vorlage `kiosk-assigned`

## 2. Usecases

- [x] 2.1 `KioskUsecase`: Link erzeugen (Gültigkeit in Tagen), auflisten, widerrufen, Token prüfen; Audit
- [x] 2.2 `RegistrationUsecase.KioskRegister` (Name / E-Mail+Name / Mitglied), Bestätigungs-URL aus `BASE_URL`
- [x] 2.3 `ConfirmByLink` ordnet bei vorgemerktem Mitglied zu; `AssignRegistration` mit E-Mail; `SuggestMembers`
- [x] 2.4 Zeitleiste: `pendingMemberName` für Organisatoren, Felder für andere entfernen

## 3. API

- [x] 3.1 Kiosk-Token-Middleware (Header `X-Kiosk-Token`) für alle Kiosk-Routen außer Bestätigung; `GET /kiosk/config`
- [x] 3.2 `POST /kiosk/shifts/{id}/register` mit `name`; Validierung
- [x] 3.3 `GET/POST /settings/kiosk-links`, `DELETE /settings/kiosk-links/{id}` (Vorstand+)
- [x] 3.4 `GET /registrations/{id}/suggestions`, `POST /registrations/{id}/assign` (Veranstaltungsleiter+)
- [x] 3.5 `api/openapi.yaml`, `make generate`; Integrationstests

## 4. Frontend

- [x] 4.1 Kiosk: Token aus Link übernehmen, Header, Hinweis ohne gültigen Token, Konfiguration über `/kiosk/config`
- [x] 4.2 Kiosk: Eintragung per Name, per E-Mail+Anzeigename, optional Mitgliedersuche; Abschlusstexte ohne Rückschluss auf bekannte Adressen
- [x] 4.3 Helferliste: Kennzeichnung „Zuordnung offen“ / „Bestätigung ausstehend“, Zuordnen-Dialog mit Vorschlägen und Suche
- [x] 4.4 Einstellungen: Kiosk-Links erzeugen (Bezeichnung, Gültigkeit), Link einmalig anzeigen/kopieren, Liste, Widerrufen
- [x] 4.5 Tests (vitest), `tsc --noEmit`
