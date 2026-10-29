# Tasks

## 1. Backend: Veranstaltungsliste

- [x] 1.1 `port.EventFilter` um `Category`, `FreeSlots`, `Ascending` erweitern; `EventRepository.CountFiltered`; gemeinsamer `applyEventFilter` in `event_repo.go` (inkl. `EXISTS`-Filter für freie Plätze, Sortierung mit `id` als Tiebreaker)
- [x] 1.2 `EventUsecase.ListEvents` liefert Seite + Gesamtzahl; Kiosk-Handler angepasst
- [x] 1.3 `EventHandler.List`: Parameter `category`, `freeSlots`, `order`; `to` inklusive ganzer Tag; Header `X-Total-Count`; 400 bei negativem `limit`/`offset` oder ungültigem `order`
- [x] 1.4 Tests (Repo, Integration)

## 2. Backend: Abgeltungsprognose

- [x] 2.1 Auflösung der Gebührenliste aus `BillingUsecase` in gemeinsame Funktion extrahieren
- [x] 2.2 `HourUsecase` erhält `SettingsRepository`; `GetMemberAccountFull` nutzt `resolveTargetHours` und liefert `openFeeCents`
- [x] 2.3 `HourHandler.GetMemberAccountByID` entfernt `openFeeCents` für fremde Konten ohne Vorstandsrolle
- [x] 2.4 Tests (Usecase, Integration)

## 3. API

- [x] 3.1 `api/openapi.yaml` aktualisieren und `make generate` ausführen

## 4. Frontend

- [x] 4.1 `api/events.ts`: `useEventPage` (liest `X-Total-Count`, lädt Zeitleisten nur für die Seite)
- [x] 4.2 `MemberDiscover.tsx`: „Von“ = heute bei jedem Öffnen, serverseitige Filter, Paginierung (10 pro Seite)
- [x] 4.3 `api/hours.ts` + `MemberDashboard.tsx`: Abgeltungsprognose im Stundenkonto
- [x] 4.4 `MemberDetail.tsx`: Abgeltungsprognose für den Vorstand
- [x] 4.5 Tests (vitest), `tsc --noEmit`
