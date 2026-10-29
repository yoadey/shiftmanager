# Design

## Context

`MemberDiscover.tsx` ruft `useEvents()` auf: `GET /api/v1/events` ohne
Parameter (Backend-Default `limit=50`, sortiert `start_date DESC`), danach
pro Veranstaltung `GET /events/{id}/timeline`. Kategorie-, „Freie
Plätze“- und Datumsfilter laufen clientseitig. `EventRepo.List` kann
bereits nach `status`, `from`, `to`, `limit`, `offset` filtern.

Die Abgeltung wird nur in `BillingUsecase.ComputeYearBilling` berechnet
(Vorstand-only, G-010). `HourUsecase.GetMemberAccountFull` liefert das
Stundenkonto für `/hours/me` und `/hours/{memberId}` – letzteres ist für
jeden angemeldeten Nutzer lesbar.

## Goals / Non-Goals

**Goals:**
- Entdecken zeigt standardmäßig nur aktuelle und künftige
  Veranstaltungen, vollständig paginiert.
- Mitglied sieht seine voraussichtliche Abgeltung, berechnet exakt wie in
  der Jahresabrechnung.

**Non-Goals:**
- Kein Speichern der Entdecken-Filter (weder URL noch Local Storage); der
  Standard „ab heute“ gilt bei jedem Öffnen.
- Keine Prognose inklusive Reservierungen/Anmeldungen: gefragt ist der
  Betrag, wenn keine weiteren Stunden geleistet werden.
- Keine Änderung an den Billing-Endpunkten oder ihrer Berechtigung.

## Decisions

- **Serverseitige Paginierung und Filter.** Clientseitig über alle
  Veranstaltungen zu paginieren hieße weiter alle Zeitleisten zu laden.
  Deshalb filtert das Backend Status (`published`), Datum, Kategorie und
  freie Plätze; die Zeitleisten werden nur für die aktuelle Seite geladen.
  Gesamtzahl wie bei der Mitgliederliste im Header `X-Total-Count`, der
  Body bleibt ein Array (kompatibel zu bestehenden Clients).
- **„Freie Plätze“ in SQL.** Eine Schicht hat einen freien Platz, wenn sie
  unbegrenzt ist (`max_helpers` 0) oder `max_helpers` größer ist als die
  Zahl ihrer Anmeldungen im Zustand `registered`, `reserved` oder
  `confirmed` – dieselbe Zählung wie die angezeigte Belegung
  (`computeOccupancy`, Frontend `calcOccupancy`), damit der Filter keine
  Veranstaltung liefert, deren Karte „besetzt“ zeigt. Umgesetzt als
  `EXISTS`-Unterabfrage, damit `LIMIT/OFFSET` und `COUNT` stimmen.
- **Sortierung.** Neuer Parameter `order=asc|desc`; Default bleibt `desc`,
  damit die Verwaltungsansichten unverändert bleiben. `id` als zweiter
  Sortierschlüssel macht die Seiten stabil.
- **`to` inklusive.** `to=YYYY-MM-DD` filtert auf `start_date < to + 1 Tag`.
- **Seitengröße 10**, Seite nur im Komponentenzustand; jede Filteränderung
  springt auf Seite 1. Die vorherige Seite bleibt beim Blättern sichtbar
  (`keepPreviousData`).
- **Prognose im Stundenkonto statt eigener Billing-Route.** `HourUsecase`
  erhält das `SettingsRepository` und berechnet `openFeeCents` mit
  `domain.ComputeBilling` über die fehlenden bestätigten Stunden (Ziel −
  bestätigt). Die Auflösung der Gebührenliste (Liste des Vereinsjahres,
  ersatzweise `feeSchedule`-Einstellung, individuelle Liste des Mitglieds
  hat Vorrang) wird aus `BillingUsecase` in eine gemeinsame Funktion
  extrahiert. Ist keine Gebührenliste konfiguriert, fehlt das Feld.
- **Sichtbarkeit.** `/hours/me` liefert `openFeeCents` immer;
  `/hours/{memberId}` nur, wenn der Aufrufer das Mitglied selbst ist oder
  mindestens Vorstand. Sonst würde jedes Mitglied die Beträge anderer
  sehen können.
- **Ziel-Auflösung angleichen.** `GetMemberAccountFull` verwendet
  `resolveTargetHours` (individuelles Ziel am Mitglied, dann Ziel pro
  Vereinsjahr) wie die Abrechnung.

## Risks / Trade-offs

- Die `EXISTS`-Unterabfrage läuft pro Veranstaltung; bei Vereinsgröße
  (einige hundert Veranstaltungen) unkritisch, Indizes auf
  `shifts.event_id` und `registrations.shift_id` existieren.
- Filter werden beim Zurückkehren aus einer Veranstaltung zurückgesetzt –
  gewollt („bei jedem Laden ab heute“), kostet aber die gewählte Seite.
