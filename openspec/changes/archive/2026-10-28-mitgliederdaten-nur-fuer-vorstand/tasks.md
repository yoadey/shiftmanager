# Tasks

## 1. Backend: Zugriffsrechte

- [x] 1.1 `router.go`: `GET /members`, `/members/export` und `/hours/summary` nur ab Vorstand; neue Route `GET /members/lookup` ab Veranstaltungsleiter
- [x] 1.2 `MemberHandler.Get`: fremde Mitglieder nur für Vorstand (403); `MemberHandler.Lookup`
- [x] 1.3 `HourHandler.GetMemberAccountByID`/`GetMemberAccount`: eigenes Konto oder Vorstand (403); Entfernen von `openFeeCents` entfällt
- [x] 1.4 `KioskHandler.SearchMembers`: nur `{id, name}` mit abgekürztem Nachnamen; 403 bei deaktivierter Suche

## 2. Backend: Zeitleiste

- [x] 2.1 `domain.Registration.ConfirmationToken` nicht serialisieren
- [x] 2.2 Zeitleiste pro Betrachter aufbereiten: `displayName` je Anmeldung; unterhalb Veranstaltungsleiter fremde `guestEmail`/`comment`/`bookedHours`/`reservedUntil` leeren (Event-, Timeline- und Kiosk-Endpunkte)
- [x] 2.3 Tests (Handler/Integration: 403-Fälle, Lookup, Zeitleiste, Kiosk-Suche)

## 3. API

- [x] 3.1 `api/openapi.yaml` aktualisieren und `make generate` ausführen

## 4. Frontend

- [x] 4.1 `mappers.ts`: `displayName` übernehmen; `EventDetail.tsx` nutzt Namen aus der Zeitleiste statt `useMembers`
- [x] 4.2 Mitglieder sehen die Namen der eingetragenen Helfer pro Schicht
- [x] 4.3 „Helfer hinzufügen“ sucht über `/members/lookup`
- [x] 4.4 `api/kiosk.ts` an die neue Suchantwort anpassen
- [x] 4.5 Route `stunden-buchen` nur für Vorstand
- [x] 4.6 Tests (vitest), `tsc --noEmit`
