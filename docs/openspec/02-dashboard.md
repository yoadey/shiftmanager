# OpenSpec: Seite 02 – Dashboard

**Version:** 1.0  
**Stand:** 2026-06-04  
**Status:** Draft

---

## 1. Zweck

Die Dashboard-Seite ist die Startseite nach dem Login. Sie gibt einen schnellen Überblick über relevante Informationen abhängig von der Rolle des Nutzers.

---

## 2. Route

| Route | Beschreibung |
|---|---|
| `/dashboard` | Dashboard (rollenabhängige Ansicht) |

---

## 3. User Stories

| ID | Als... | möchte ich... | damit... |
|---|---|---|---|
| DASH-01 | Helfer | meine nächsten Schichten sehen | ich weiß, wann ich eingesetzt bin |
| DASH-02 | Helfer | offene Schichten mit freien Plätzen sehen | ich mich schnell anmelden kann |
| DASH-03 | Koordinator | sehen wie viele Schichten in meinem Bereich besetzt sind | ich den Status im Blick habe |
| DASH-04 | Admin | eine Gesamtübersicht aller Veranstaltungen sehen | ich den Überblick über alle Aktivitäten habe |
| DASH-05 | Admin | Schnellzugriff auf häufig genutzte Verwaltungsfunktionen haben | ich effizient arbeiten kann |

---

## 4. UI-Anforderungen

### 4.1 Helfer-Dashboard

**Bereich: Meine nächsten Schichten**
- Liste der nächsten 3–5 angemeldeten Schichten mit:
  - Datum, Uhrzeit (Start–Ende)
  - Bereich/Schichtname
  - Status (Angemeldet / Bestätigt)
- Button "Alle Schichten anzeigen" → `/meine-schichten`

**Bereich: Offene Schichten**
- Liste der nächsten 3–5 Schichten mit freien Plätzen
- Anzeige: Name, Bereich, Datum/Zeit, freie Plätze
- Button "Jetzt anmelden" je Schicht (öffnet Bestätigungsdialog)
- Button "Alle offenen Schichten" → `/schichten`

**Bereich: Statistik-Kacheln (klein)**
- Anzahl meiner angemeldeten Schichten
- Anzahl bereits absolvierter Schichten
- Gesamtstunden eingeplant

### 4.2 Koordinator-Dashboard

Alles wie Helfer-Dashboard, zusätzlich:

**Bereich: Mein Bereich – Status**
- Name des zugewiesenen Bereichs
- Belegungs-Fortschrittsbalken: X von Y Schichtplätzen besetzt
- Liste der Schichten mit Ampelstatus:
  - Grün: ≥ min_helfer besetzt
  - Gelb: 1 bis (min_helfer - 1) besetzt
  - Rot: 0 Helfer
- Direktlink zu Schicht-Verwaltung des Bereichs

### 4.3 Admin-Dashboard

**Bereich: Veranstaltungs-Übersicht**
- Kacheln für alle aktiven Veranstaltungen mit:
  - Name, Datum, Gesamtbelegungsquote (%)
  - Quick-Links: "Schichten", "Helfer", "Bereiche"

**Bereich: System-Statistiken**
- Gesamtzahl registrierter Helfer
- Schichten gesamt / besetzt / offen / überfüllt
- Ausstehende Registrierungen (falls manual-approve aktiv)

**Bereich: Schnellaktionen**
- "Neue Schicht anlegen"
- "Helfer einladen"
- "Veranstaltung anlegen"

---

## 5. API-Endpunkte

| Methode | Pfad | Beschreibung |
|---|---|---|
| `GET` | `/api/v1/dashboard` | Rollenabhängige Dashboard-Daten |
| `GET` | `/api/v1/dashboard/next-shifts` | Nächste Schichten des eingeloggten Users |
| `GET` | `/api/v1/dashboard/open-shifts` | Offene Schichten mit freien Plätzen |
| `GET` | `/api/v1/dashboard/stats` | Statistiken (Admin/Koordinator) |

---

## 6. Business Rules

- Das Dashboard zeigt immer die Daten der **aktiven** Veranstaltung. Falls mehrere Veranstaltungen aktiv sind, wird die nächste (nach Startdatum) bevorzugt.
- Vergangene Schichten (Ende > jetzt) werden nicht als "offen" angezeigt.
- Helfer sehen nur Schichten in aktiven Veranstaltungen.

---

## 7. Akzeptanzkriterien

- [ ] Nach Login erscheint das Dashboard als erste Seite
- [ ] Helfer sieht nur seine Schichten und offene Schichten
- [ ] Koordinator sieht zusätzlich Bereichs-Status
- [ ] Admin sieht Gesamtüberblick aller Veranstaltungen
- [ ] "Jetzt anmelden"-Button öffnet Bestätigungsdialog und meldet Helfer an
- [ ] Alle Daten laden in < 1s (gecacht)
- [ ] Dashboard ist auf Mobile gut nutzbar (Kacheln untereinander)
