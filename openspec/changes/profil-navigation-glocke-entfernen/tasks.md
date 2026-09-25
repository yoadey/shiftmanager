# Tasks

## 1. Desktop-Profilkarte klickbar machen

- [ ] 1.1 `DesktopShell.tsx`: `dt-user`/`dt-side-foot`-Block zu einem Button/Link mit `onClick={() => navigate(routes.profil)}` und aktivem Zustand machen; manuell verifizieren, dass ein Klick zur Profilseite navigiert
- [ ] 1.2 `DesktopShell.tsx`: `profil`-Eintrag aus der gerenderten Sidebar-Nav-Liste filtern, ohne ihn aus der gemeinsamen `tabs`-Liste in `App.tsx` zu entfernen; verifizieren, dass `MobileShell.tsx` den Tab weiterhin erhält

## 2. Glocke entfernen

- [ ] 2.1 `MemberDashboard.tsx`: Glocken-Button (Icon, Klick-Handler, Benachrichtigungs-Punkt) entfernen
- [ ] 2.2 Prüfen, ob der Icon-Eintrag `bell` in `Icon.tsx` noch anderswo verwendet wird (`grep -rn "name=\"bell\""`); falls nicht, den Eintrag entfernen

## 3. Tests

- [ ] 3.1 Frontend-Unit-/Snapshot-Tests anpassen, die auf das Glocken-Icon oder die alte Profil-Navigation referenzieren (`npm test -- --run` grün)
- [ ] 3.2 Playwright-E2E-Test ergänzen/anpassen: Desktop-Profilkarte anklicken öffnet `/profil`; mobiler `profil`-Tab weiterhin funktionsfähig
