# Tasks

## 1. Desktop-Profilkarte klickbar machen

- [x] 1.1 `DesktopShell.tsx`: `dt-user`/`dt-side-foot`-Block zu einem Button/Link mit `onClick={() => navigate(routes.profil)}` und aktivem Zustand machen; manuell verifizieren, dass ein Klick zur Profilseite navigiert
- [x] 1.2 `DesktopShell.tsx`: `profil`-Eintrag aus der gerenderten Sidebar-Nav-Liste filtern, ohne ihn aus der gemeinsamen `tabs`-Liste in `App.tsx` zu entfernen; verifizieren, dass `MobileShell.tsx` den Tab weiterhin erhält

## 2. Glocke entfernen

- [x] 2.1 `MemberDashboard.tsx`: Glocken-Button (Icon, Klick-Handler, Benachrichtigungs-Punkt) entfernen
- [x] 2.2 Prüfen, ob der Icon-Eintrag `bell` in `Icon.tsx` noch anderswo verwendet wird (`grep -rn "name=\"bell\""`); falls nicht, den Eintrag entfernen — keine weiteren Verwendungen gefunden, Eintrag entfernt

## 3. Tests

- [x] 3.1 Frontend-Unit-/Snapshot-Tests anpassen, die auf das Glocken-Icon oder die alte Profil-Navigation referenzieren (`npm test -- --run` grün) — neuer `DesktopShell.test.tsx` (4 Tests: kein Profil-Tab im Sidebar-Nav, andere Tabs bleiben, Klick auf Profilkarte navigiert zu `/profil`, aktiver Zustand auf `/profil`); volle Suite 21 Dateien/86 Tests grün, `npx tsc --noEmit` sauber
- [x] 3.2 Playwright-E2E-Test ergänzen/anpassen: Desktop-Profilkarte anklicken öffnet `/profil`; mobiler `profil`-Tab weiterhin funktionsfähig — `e2e/helpers/app.ts` um `openProfileCard()` ergänzt, `member-dashboard.member.spec.ts` angepasst (kein `Profil`-Nav-Tab-Assert mehr, neuer Test für die Profilkarte, `Member — Profil tab`-Block nutzt jetzt `openProfileCard`); einzeln/isoliert grün — ein Voll-Lauf der ganzen Datei kollabiert nach ~3-10 sequenziellen Browser-Kontexten in dieser Sandbox, reproduzierbar identisch auf unveränderten, unabhängigen Spec-Dateien (z. B. `admin-members.spec.ts`), also eine Umgebungs-Ressourcengrenze dieser Sandbox, keine Regression dieser Change
