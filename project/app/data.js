/* ShiftManager demo data + helpers — TSC Schwarz-Gelb Aachen */
(function () {
  // ── members ──
  const members = [
    { id: 'm-jonas', first: 'Jonas', last: 'Berger', email: 'jonas.berger@example.de', since: '2021-09-01', goal: null },
    { id: 'm-maxi', first: 'Maximilian', last: 'Müller', email: 'm.mueller@example.de', since: '2019-03-12', goal: null },
    { id: 'm-lena', first: 'Lena', last: 'Brandt', email: 'lena.brandt@example.de', since: '2022-01-20', goal: 12 },
    { id: 'm-sophie', first: 'Sophie', last: 'Klein', email: 's.klein@example.de', since: '2020-11-04', goal: null },
    { id: 'm-tobi', first: 'Tobias', last: 'Hofmann', email: 'tobias.h@example.de', since: '2023-06-15', goal: null },
    { id: 'm-aylin', first: 'Aylin', last: 'Demir', email: 'aylin.demir@example.de', since: '2018-08-30', goal: null },
    { id: 'm-mark', first: 'Mark', last: 'Petersen', email: 'mark.p@example.de', since: '2024-02-01', goal: null },
    { id: 'm-carla', first: 'Carla', last: 'Vogt', email: 'carla.vogt@example.de', since: '2021-04-18', goal: 25 },
    { id: 'm-finn', first: 'Finn', last: 'Schäfer', email: 'finn.s@example.de', since: '2023-10-09', goal: null },
    { id: 'm-nina', first: 'Nina', last: 'Roth', email: 'nina.roth@example.de', since: '2017-05-22', goal: null },
  ];

  const currentUserId = 'm-jonas';

  // ── events with day-grouped shifts ──
  // signup status: 'angemeldet' (reservation, counts to incl.), 'bestätigt' (confirmed hours), 'nichterschienen'
  const events = [
    {
      id: 'ev-turnier',
      name: 'Frühjahrsturnier Standard & Latein',
      category: 'Turnier',
      location: 'Sporthalle Aachen-Brand',
      status: 'veröffentlicht',
      description: 'Zweitägiges Ranglistenturnier mit Paaren aus ganz NRW. Helfer für Auf-/Abbau, Einlass, Catering und Garderobe gesucht.',
      days: [
        { date: '2026-06-13', shifts: [
          { id: 's1', name: 'Aufbau Parkett & Technik', start: '08:00', end: '10:00', min: 2, max: 4, qual: 'Tanzfläche-Aufbau', desc: 'Schwingboden verlegen, Technik & Banden stellen.', signups: [{ memberId: 'm-maxi', status: 'angemeldet' }] },
          { id: 's2', name: 'Einlass & Kasse', start: '10:00', end: '14:00', min: 2, max: 3, desc: 'Tickets, Tageskasse, Startnummern ausgeben.', signups: [{ memberId: 'm-jonas', status: 'angemeldet', comment: 'Bin etwas später, gegen 10:15 da.' }, { memberId: 'm-lena', status: 'angemeldet' }] },
          { id: 's3', name: 'Cateringtheke Mittag', start: '11:00', end: '15:00', min: 3, max: 5, desc: 'Getränke- und Kuchenverkauf während der Vorrunden.', signups: [{ memberId: 'm-sophie', status: 'angemeldet' }] },
          { id: 's4', name: 'Garderobe', start: '12:00', end: '18:00', min: 1, max: 2, desc: 'Garderobenannahme für Gäste.', signups: [] },
        ]},
        { date: '2026-06-14', shifts: [
          { id: 's5', name: 'Cateringtheke Finaltag', start: '11:00', end: '16:00', min: 3, max: 5, desc: 'Verkauf während der Endrunden.', signups: [{ memberId: 'm-aylin', status: 'angemeldet' }, { memberId: 'm-carla', status: 'angemeldet' }] },
          { id: 's6', name: 'Siegerehrung & Abbau', start: '16:00', end: '20:00', min: 4, max: 6, qual: 'Tanzfläche-Aufbau', desc: 'Pokale stellen, Saal & Parkett abbauen.', signups: [{ memberId: 'm-maxi', status: 'angemeldet' }] },
        ]},
      ],
    },
    {
      id: 'ev-cafe',
      name: 'Tanzcafé für Senioren',
      category: 'Vereinsleben',
      location: 'Vereinsheim, Saal 1',
      status: 'veröffentlicht',
      description: 'Gemütlicher Tanznachmittag mit Kaffee und Kuchen für unsere Seniorengruppe.',
      days: [
        { date: '2026-06-28', shifts: [
          { id: 's7', name: 'Empfang & Kasse', start: '14:00', end: '15:00', min: 1, max: 2, desc: 'Gäste begrüßen, Eintritt einsammeln.', signups: [{ memberId: 'm-nina', status: 'angemeldet' }, { memberId: 'm-tobi', status: 'angemeldet' }] },
          { id: 's8', name: 'Kuchentheke', start: '14:00', end: '17:00', min: 2, max: 3, desc: 'Kuchen schneiden und ausgeben, Kaffee kochen.', signups: [{ memberId: 'm-jonas', status: 'angemeldet' }] },
          { id: 's9', name: 'Aufräumen & Spülen', start: '17:00', end: '18:30', min: 2, max: 4, desc: 'Saal aufräumen, Geschirr spülen.', signups: [] },
        ]},
      ],
    },
    {
      id: 'ev-tdot',
      name: 'Tag der offenen Tür',
      category: 'Mitgliederwerbung',
      location: 'Vereinsheim & Außengelände',
      status: 'veröffentlicht',
      description: 'Schnupperkurse, Showtanz und Infostände – wir öffnen unsere Türen für alle Tanzbegeisterten.',
      days: [
        { date: '2026-07-11', shifts: [
          { id: 's10', name: 'Infostand Mitgliedschaft', start: '11:00', end: '16:00', min: 2, max: 3, desc: 'Interessierte beraten, Flyer verteilen.', signups: [] },
          { id: 's11', name: 'Schnupperkurs-Betreuung', start: '13:00', end: '15:00', min: 2, max: 4, desc: 'Trainer beim Schnupperkurs unterstützen.', signups: [{ memberId: 'm-finn', status: 'angemeldet' }] },
          { id: 's12', name: 'Getränke & Grill', start: '12:00', end: '17:00', min: 2, max: 3, desc: 'Außenbewirtung am Grillstand.', signups: [] },
        ]},
      ],
    },
    {
      id: 'ev-mv',
      name: 'Jahres-Mitgliederversammlung',
      category: 'Vereinsleben',
      location: 'Vereinsheim, Saal 1',
      status: 'entwurf',
      description: 'Ordentliche Mitgliederversammlung mit Vorstandswahl.',
      days: [
        { date: '2026-07-18', shifts: [
          { id: 's13', name: 'Bestuhlung & Technik', start: '17:00', end: '19:00', min: 2, max: 3, desc: 'Stuhlreihen stellen, Mikro & Beamer.', signups: [] },
          { id: 's14', name: 'Getränkeausgabe', start: '18:30', end: '22:00', min: 1, max: 2, desc: '', signups: [] },
        ]},
      ],
    },
    {
      id: 'ev-neujahr',
      name: 'Neujahrsball 2026',
      category: 'Turnier',
      location: 'Altes Kurhaus Aachen',
      status: 'abgeschlossen',
      description: 'Festlicher Gesellschaftsball zum Jahresauftakt.',
      days: [
        { date: '2026-01-11', shifts: [
          { id: 's15', name: 'Garderobe', start: '18:00', end: '23:00', min: 2, max: 3, desc: '', signups: [{ memberId: 'm-jonas', status: 'bestätigt', hours: 5 }, { memberId: 'm-sophie', status: 'bestätigt', hours: 5 }] },
        ]},
      ],
    },
    {
      id: 'ev-putz',
      name: 'Vereinsheim-Frühjahrsputz',
      category: 'Vereinsleben',
      location: 'Vereinsheim',
      status: 'abgeschlossen',
      description: 'Großreinemachen vor Saisonstart.',
      days: [
        { date: '2026-03-21', shifts: [
          { id: 's16', name: 'Saal & Küche', start: '09:00', end: '13:00', min: 3, max: 6, desc: '', signups: [{ memberId: 'm-jonas', status: 'bestätigt', hours: 4 }, { memberId: 'm-lena', status: 'bestätigt', hours: 4 }, { memberId: 'm-mark', status: 'nichterschienen', hours: 0 }] },
        ]},
      ],
    },
  ];

  // manual hour bookings (SA-005) not tied to an event
  const manualBookings = [
    { id: 'mb1', memberId: 'm-jonas', date: '2026-02-09', hours: 2.5, desc: 'Protokoll Vorstandssitzung', by: 'Vorstand' },
  ];

  // settings (configurable in admin)
  const settings = {
    clubName: 'TSC Schwarz-Gelb Aachen',
    yearGoal: 20,
    clubYear: '2026',
    nameMode: 'abbrev',          // 'abbrev' | 'full'   (NM-001/004)
    reservationHours: 48,         // K-010
    billingMode: 'manuell',       // 'auto' | 'manuell' (G-009)
    kioskSearch: false,           // K-002
    feeSchedule: [5, 7, 10, 15],  // G-001 per missing hour
    deregisterDeadlineH: 24,      // SC-006
  };

  // audit log sample
  const audit = [
    { ts: '2026-05-28 14:02', who: 'Sophie K. (Vorstand)', what: 'Stundenziel global auf 20 h geändert', cat: 'Stunden' },
    { ts: '2026-05-26 09:41', who: 'Sophie K. (Vorstand)', what: 'Manuelle Buchung +2,5 h für Jonas Berger', cat: 'Stunden' },
    { ts: '2026-05-20 18:15', who: 'Mark P. (Admin)', what: 'Namensmodus auf „Abgekürzt" gesetzt', cat: 'Datenschutz' },
    { ts: '2026-05-12 11:30', who: 'Sophie K. (Vorstand)', what: 'Abgeltungsliste geändert (3. Fehlstunde 8→10 €)', cat: 'Abrechnung' },
  ];

  // ── helpers ──
  const M = {};
  members.forEach(m => { M[m.id] = m; });

  function fmtName(id, opts) {
    const m = typeof id === 'string' ? M[id] : id;
    if (!m) return 'Unbekannt';
    opts = opts || {};
    const full = `${m.first} ${m.last}`;
    if (m.id === currentUserId) return full;            // NM-006 own name
    if (opts.viewerFull) return full;                    // NM-003 board+
    if (opts.mode === 'full') return full;
    return `${m.first} ${m.last.charAt(0)}.`;            // NM-002 abbreviated
  }

  function durH(start, end) {
    const [sh, sm] = start.split(':').map(Number);
    const [eh, em] = end.split(':').map(Number);
    return ((eh * 60 + em) - (sh * 60 + sm)) / 60;
  }

  function occ(shift) {
    const count = shift.signups.filter(s => s.status === 'angemeldet' || s.status === 'bestätigt' || s.status === 'reserviert').length;
    let key, label;
    if (count >= shift.max) { key = 'full'; label = 'Ausgebucht'; }
    else if (count >= shift.min) { key = 'ok'; label = 'Besetzt'; }
    else if (count > 0) { key = 'warn'; label = 'Teilweise'; }
    else { key = 'crit'; label = 'Offen'; }
    return { count, min: shift.min, max: shift.max, key, label, free: shift.max - count, needsMore: count < shift.min };
  }

  const fmtDate = (iso, style) => {
    const d = new Date(iso + 'T12:00:00');
    if (style === 'weekday-long') return new Intl.DateTimeFormat('de-DE', { weekday: 'long', day: 'numeric', month: 'long' }).format(d);
    if (style === 'weekday') return new Intl.DateTimeFormat('de-DE', { weekday: 'short', day: 'numeric', month: 'long' }).format(d);
    if (style === 'short') return new Intl.DateTimeFormat('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' }).format(d);
    if (style === 'daymon') return new Intl.DateTimeFormat('de-DE', { day: 'numeric', month: 'short' }).format(d);
    return new Intl.DateTimeFormat('de-DE', { day: 'numeric', month: 'long', year: 'numeric' }).format(d);
  };

  const eur = (n) => n.toLocaleString('de-DE', { style: 'currency', currency: 'EUR' });
  const hrs = (n) => (Number.isInteger(n) ? n : n.toLocaleString('de-DE', { minimumFractionDigits: 1 })) + ' h';

  window.SM = {
    members, currentUserId, events, manualBookings, settings, audit,
    M, fmtName, durH, occ, fmtDate, eur, hrs,
  };
})();
