# Change: A-005 — mehrere gleichzeitige OIDC-Provider

**Status:** offen · **Anforderung:** `A-005` (KANN) · **Quelle:**
`project/requirements_extracted.txt` 3.1.1

## Problem

> Es müssen mehrere OIDC-Provider gleichzeitig unterstützt werden können
> (z. B. Vereins-SSO + Google).

Aktuell ist das System fest auf **einen** OIDC-Provider verdrahtet:

- `internal/config/config.go`: genau ein Satz `OIDC_ISSUER` /
  `OIDC_CLIENT_ID` / `OIDC_CLIENT_SECRET` / `OIDC_REDIRECT_URL`.
- `cmd/server/main.go`: erzeugt eine einzelne `oidc.Service`-Instanz
  (`internal/adapter/oidc/service.go`) und injiziert sie als
  `port.OIDCService` in den `AuthHandler`.
- `GET /api/v1/auth/login` kennt keinen Provider-Parameter — es gibt genau
  eine Login-URL.

## Vorgeschlagene Lösung

1. **Konfiguration:** `OIDC_PROVIDERS`-Liste (JSON oder
   `NAME1:ISSUER1:CLIENT_ID1:...,NAME2:...`) statt einzelner
   `OIDC_*`-Variablen; Rückwärtskompatibilität zu den bisherigen
   `OIDC_*`-Variablen als impliziter Single-Provider mit Namen `default`.
2. **Backend:** `port.OIDCService` wird pro Provider instanziiert und in
   einer `map[string]port.OIDCService` gehalten; `AuthHandler` bekommt diese
   Map statt einer einzelnen Instanz.
3. **Routing:** `GET /api/v1/auth/login?provider=<name>` wählt den
   Provider; der State-Parameter (aktuell nur CSRF-Schutz) trägt den
   Provider-Namen verschlüsselt/signiert mit, damit
   `/api/v1/auth/callback` weiß, welchen Provider er zur Token-Verifikation
   verwenden muss (der OIDC-Callback selbst enthält keinen Provider-Namen).
4. **Mitgliederzuordnung (ML-004):** Der E-Mail-Abgleich beim ersten Login
   bleibt unverändert — provider-übergreifend über die E-Mail-Adresse. Bei
   Konflikt (gleiche E-Mail bei zwei Providern) wird die bestehende
   Zuordnung nicht automatisch überschrieben; das erfordert eine explizite
   Fehlermeldung/Admin-Eingriff (Detail noch offen).
5. **Frontend:** `LoginPage.tsx` zeigt bei konfigurierten Mehrfach-Providern
   mehrere Buttons ("Mit Vereins-SSO anmelden", "Mit Google anmelden") statt
   des aktuellen Einzel-Buttons.

## Betroffene Bereiche

- `internal/config`, `internal/adapter/oidc`, `internal/adapter/http/handler/auth_handler.go`
- `frontend/src/screens/auth/LoginPage.tsx`
- `docs/openspec/01-auth-mitglieder.md` (nach Umsetzung aktualisieren)

## Migration

Kein Datenbank-Migrationsscript nötig — `Member.OIDCSubject` bleibt ein
einzelnes Feld; bei Bedarf könnte optional der Provider-Name mitgespeichert
werden, ist aber nicht zwingend, solange E-Mail-Abgleich provider-übergreifend
eindeutig funktioniert.
