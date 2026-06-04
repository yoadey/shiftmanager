# Analyse & Verbesserungsvorschläge

**Stand:** 2026-06-04

---

## Bewertung der Spezifikation

Die 10 definierten Seiten decken den Kern-Workflow eines Schichtmanagers vollständig ab. Im Folgenden eine kritische Analyse mit konkreten Empfehlungen.

---

## Was gut ist

- **Rollenkonzept** (Helfer / Koordinator / Admin) ist sauber und deckt reale Veranstaltungsorganisation gut ab.
- **Kern-Entitäten** (Veranstaltung → Bereich → Schicht → Anmeldung) sind logisch hierarchisch.
- **DSGVO** ist von Anfang an berücksichtigt (Datenlöschung, Anonymisierung).
- **Anwesenheitsliste als PDF** ist ein konkreter, praxisnaher Feature-Wunsch der bei solchen Tools oft fehlt.

---

## Verbesserungsvorschläge

### 1. KRITISCH: Navigation / Shell fehlt in der Spezifikation

**Problem:** Es gibt keine Spec für die globale Navigation (Sidebar/Navbar), Breadcrumbs, Mobile-Navigation (Hamburger-Menü).

**Empfehlung:** Eine eigene Mini-Spec für das Layout-Shell erstellen:
- Welche Navpunkte sind für welche Rolle sichtbar?
- Wo erscheint der Nutzer-Avatar / Logout-Button?
- Mobile: Bottom-Tab-Navigation vs. Hamburger?

---

### 2. WICHTIG: Einladungs-Flow für Helfer ist unklar

**Problem:** Es gibt zwei Wege zum System:
- Selbstregistrierung (`/register`) – konfigurierbar deaktivierbar
- Admin-Einladung (`/admin/helfer/einladen`)

Aber: Was passiert wenn ein Helfer über einen Einladungslink registriert wird? Gibt es eine eigene `/invite/:token` Seite? Diese Seite fehlt in der Spec.

**Empfehlung:** Route `/invite/:token` in Auth-Spec ergänzen:
- Token validieren → Formular "Konto erstellen" (Passwort setzen)
- Vorname/Nachname vorausgefüllt aus Einladungsdaten

---

### 3. WICHTIG: Konflikt zwischen Seite 03 und Seite 05 bei Abmeldefrist

**Problem:** Die Abmeldefrist wird auf Seite 03 (Schichtenplan) und Seite 04 (Meine Schichten) erwähnt, ist aber in Seite 08 (Veranstaltungen) als konfigurierbarer Wert definiert. Seite 05 (Schicht-Verwaltung) erwähnt die Frist nicht.

**Empfehlung:** In Seite 05 explizit aufnehmen, dass Koordinatoren/Admins auch nach Ablauf der Frist Anmeldungen bearbeiten können (Override). Das ist ein wichtiger Business-Rule-Unterschied.

---

### 4. WICHTIG: Schicht-Status-Logik ist inkonsistent

**Problem:** Schichten haben Status `open | full | closed | cancelled`. Aber `full` ist eigentlich kein explizit gesetzter Status – er ergibt sich aus `registrations.count >= max_helfer`. Wenn der Status explizit gesetzt werden muss, entsteht ein Synchronisierungsproblem.

**Empfehlung:** Status-Logik klarstellen:
- `draft` – Entwurf, nicht sichtbar für Helfer
- `open` – Anmeldung möglich
- `closed` – manuell geschlossen (keine Anmeldung, aber nicht abgesagt)
- `cancelled` – abgesagt
- `full` als **berechnetes Attribut** im API-Response, nicht als DB-Status

---

### 5. DESIGN: Separate Admin-Routen vs. eingebettete Verwaltung

**Problem:** Admin-Seiten liegen unter `/admin/*`, normale Helfer-Seiten direkt unter `/`. Das ist sauber, aber es gibt eine Chance für eine bessere UX:

**Alternative Überlegung:** Ein einziges Interface mit rollenabhängig sichtbaren Bereichen (z. B. der Schichtenplan zeigt für Admins inline Edit-Icons). Das reduziert die Gesamtkomplexität erheblich.

**Empfehlung:** Klare Entscheidung treffen und in der Spec festhalten:
- Option A: Klare Trennung `/admin/*` (einfacher zu sichern, klare Rollen-URL)
- Option B: Rollenbasierte UI im gleichen Interface (bessere UX, komplexere Implementierung)

Die aktuelle Spec impliziert Option A – das ist gut und sollte explizit dokumentiert werden.

---

### 6. DESIGN: Fehlende Benachrichtigungs-Seite (Inbox)

**Problem:** Email-Benachrichtigungen sind konfigurierbar, aber es gibt keine In-App-Notification-Inbox. Das bedeutet: Wenn ein Helfer keine Emails liest, verpasst er wichtige Infos (z. B. Schicht-Absagen).

**Empfehlung (Phase 2):** Eine `/benachrichtigungen`-Seite ergänzen:
- Chronologische Liste aller System-Events für den User
- Gelesen/Ungelesen-Status
- Glocken-Icon in der Navigation mit Unread-Count-Badge

Dies kann als Phase-2-Feature markiert werden, aber das Datenmodell sollte es von Anfang an unterstützen.

---

### 7. DESIGN: Schicht-Vorlagen (Templates)

**Problem:** Das Klonen einer Veranstaltung (Seite 08) kopiert die Schicht-Struktur. Aber für wiederkehrende Events (z. B. jährliches Festival) wäre ein explizites Template-Konzept mächtiger.

**Empfehlung (nice-to-have):**
- Templates für Bereiche + Schichtstrukturen (ohne Datum)
- Beim Erstellen einer neuen Veranstaltung: "Template auswählen" → Daten werden reingezogen, Datum muss gesetzt werden
- Einfachere Alternative: Das bestehende Clone-Feature ist ausreichend für MVP

---

### 8. TECHNIK: Echtzeit-Updates nicht spezifiziert

**Problem:** Wenn Admin die Schichtbelegung ändert, sieht ein Helfer auf der Schichtenplan-Seite die Änderung erst nach Page-Reload.

**Empfehlung:**
- **MVP:** Polling alle 60 Sekunden für Schicht-Verfügbarkeit (einfach)
- **Phase 2:** Server-Sent Events (SSE) oder WebSocket für Echtzeit-Belegungsanzeige (skalierbar)
- In der Spec sollte explizit stehen, ob Echtzeit erwartet wird oder nicht

---

### 9. TECHNIK: API-Versioning und Breaking Changes

**Problem:** Die API liegt unter `/api/v1/`, aber es gibt keine Spec dafür, wie mit Breaking Changes umgegangen wird.

**Empfehlung:** Kurze API-Changelog-Regel in der General-Spec:
- Minor Changes (neue Felder) → kein Version-Bump
- Breaking Changes → `/api/v2/` mit Deprecation-Frist

---

### 10. FEHLEND: Onboarding-Flow / erste Schritte

**Problem:** Was passiert wenn ein frisch registrierter Admin sich einloggt und noch keine Veranstaltung existiert? Das Dashboard ist dann leer und verwirrend.

**Empfehlung:** Empty States klar definieren:
- Dashboard leer → "Starte mit dem Anlegen deiner ersten Veranstaltung" mit direktem CTA-Button
- Schichtenplan ohne Schichten → "Es wurden noch keine Schichten angelegt"
- Diese Empty States sollten in den jeweiligen Seiten-Specs ergänzt werden

---

## Priorisierungs-Empfehlung (MVP vs. Phase 2)

### MVP (v1.0 – das Minimum für eine erste Veranstaltung)

| Seite | Priorität |
|---|---|
| 01 – Login/Registrierung | MUSS |
| 02 – Dashboard (vereinfacht) | MUSS |
| 03 – Schichtenplan | MUSS |
| 04 – Meine Schichten (ohne iCal/PDF) | MUSS |
| 05 – Schicht-Verwaltung | MUSS |
| 06 – Helfer-Verwaltung (ohne Massen-Einladung) | MUSS |
| 07 – Bereiche-Verwaltung | MUSS |
| 08 – Veranstaltungs-Verwaltung | MUSS |
| 09 – Berichte (nur CSV-Export) | SOLL |
| 10 – Profil (ohne Session-Verwaltung) | SOLL |

### Phase 2 (nach erster produktiver Nutzung)

- In-App Benachrichtigungs-Inbox
- iCal-Export
- Anwesenheitsliste als PDF
- Echtzeit-Updates (SSE/WebSocket)
- Template-System für Veranstaltungen
- Schicht-Bestätigungs-Flow durch Koordinator
- Statistik-Diagramme in Berichten
- 2FA für Admin-Accounts
