/* Member (Mitglied) screens */

// gather the current user's shift signups across live events
function collectMy(events, uid) {
  const today = '2026-05-30';
  const up = [], past = [];
  events.forEach(ev => ev.days.forEach(day => day.shifts.forEach(sh => {
    const su = sh.signups.find(s => s.memberId === uid);
    if (!su) return;
    const rec = { ev, day, sh, su, dur: window.SM.durH(sh.start, sh.end) };
    if (day.date >= today && su.status !== 'bestätigt' && su.status !== 'nichterschienen') up.push(rec);
    else past.push(rec);
  })));
  up.sort((a, b) => (a.day.date + a.sh.start).localeCompare(b.day.date + b.sh.start));
  past.sort((a, b) => (b.day.date).localeCompare(a.day.date));
  return { up, past };
}

function MemberDashboard() {
  const { state, nav, settings, role } = useApp();
  const SM = window.SM, uid = SM.currentUserId;
  const me = SM.M[uid];
  const { up } = collectMy(state.events, uid);

  let confirmed = 0;
  state.events.forEach(ev => ev.days.forEach(d => d.shifts.forEach(sh =>
    sh.signups.forEach(s => { if (s.memberId === uid && s.status === 'bestätigt') confirmed += s.hours || 0; }))));
  state.manualBookings.forEach(b => { if (b.memberId === uid) confirmed += b.hours; });
  const reservedExtra = up.reduce((a, r) => a + r.dur, 0);
  const incl = confirmed + reservedExtra;
  const goal = me.goal || settings.yearGoal;
  const remaining = Math.max(goal - incl, 0);

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">{settings.clubName}</div>
          <div className="sm-title">Hallo, {me.first}</div>
        </div>
        <div style={{ display: 'flex', gap: 9 }}>
          <IconCircle name="bell" onClick={() => nav.go('profil')} badge />
          <Avatar id={uid} size={42} />
        </div>
      </div>

      <div className="sm-pad">
        {/* hour account */}
        <div className="sm-card pad" style={{ padding: 18 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 14 }}>
            <span style={{ fontWeight: 700, fontSize: 14.5, color: 'var(--ink-2)' }}>Stundenkonto</span>
            <Badge kind="primary" dot={false}>Vereinsjahr {settings.clubYear}</Badge>
          </div>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
            <span style={{ fontFamily: 'Bricolage Grotesque', fontSize: 46, fontWeight: 800, lineHeight: 1, letterSpacing: '-0.03em' }}>
              {SM.hrs(confirmed).replace(' h', '')}
            </span>
            <span style={{ color: 'var(--muted)', fontWeight: 700, fontSize: 18 }}>/ {goal} h</span>
          </div>
          <div style={{ color: 'var(--muted)', fontSize: 13, fontWeight: 600, marginTop: 2, marginBottom: 14 }}>bestätigte Stunden</div>
          <HourBar confirmed={confirmed} reserved={incl} goal={goal} />
          <div style={{ display: 'flex', gap: 18, marginTop: 14 }}>
            <Legend swatch="seg-c" label="Bestätigt" val={SM.hrs(confirmed)} />
            <Legend swatch="seg-r" label="Inkl. Reservierungen" val={SM.hrs(incl)} />
          </div>
          <div style={{ marginTop: 14, padding: '11px 13px', background: 'var(--surface-2)', borderRadius: 13, fontSize: 13.5, fontWeight: 600, color: 'var(--ink-2)', display: 'flex', alignItems: 'center', gap: 8 }}>
            <Icon name={remaining > 0 ? 'spark' : 'check'} size={17} color="var(--warn)" />
            {remaining > 0
              ? <span>Noch <b style={{ color: 'var(--ink)' }}>{SM.hrs(remaining)}</b> bis zum Jahresziel – inkl. deiner Reservierungen.</span>
              : <span>Stark! Du hast dein Jahresziel bereits erreicht.</span>}
          </div>
        </div>

        <div style={{ marginTop: 14 }}>
          <Button icon="search" onClick={() => nav.go('entdecken')}>Freie Schichten finden</Button>
        </div>

        <Section title="Deine nächsten Schichten" action={up.length > 0 && <TextLink onClick={() => nav.go('schichten')}>Alle</TextLink>} />
        {up.length === 0
          ? <EmptyState icon="calendar" title="Noch nichts geplant" text="Melde dich für eine Schicht an – sie erscheint dann hier." />
          : up.slice(0, 3).map(r => <MyShiftCard key={r.sh.id} r={r} onClick={() => nav.push('event', { id: r.ev.id })} />)}
      </div>
    </div>
  );
}

function Legend({ swatch, label, val }) {
  return (
    <div style={{ flex: 1 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
        <span className={'sm-track ' + ''} style={{ width: 18, height: 10, border: 'none', display: 'inline-block' }}>
          <span className={swatch} style={{ display: 'block', width: '100%', height: '100%', borderRadius: 3 }} />
        </span>
        <span style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)' }}>{label}</span>
      </div>
      <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 18, marginTop: 3 }}>{val}</div>
    </div>
  );
}

function MyShiftCard({ r, onClick }) {
  const SM = window.SM;
  const o = SM.occ(r.sh);
  const statusBadge = r.su.status === 'reserviert'
    ? <Badge kind="warn">Reserviert</Badge>
    : <Badge kind="info">Angemeldet</Badge>;
  return (
    <div className="sm-card pressable" style={{ padding: 14, marginBottom: 10, display: 'flex', gap: 13, alignItems: 'center' }} onClick={onClick}>
      <DateChip date={r.day.date} />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontWeight: 700, fontSize: 15.5, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{r.sh.name}</div>
        <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{r.ev.name}</div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 7 }}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 12.5, fontWeight: 700, color: 'var(--ink-2)' }}>
            <Icon name="clock" size={14} stroke={2.2} />{r.sh.start}–{r.sh.end}
          </span>
          {statusBadge}
        </div>
      </div>
    </div>
  );
}

function DateChip({ date }) {
  const d = new Date(date + 'T12:00:00');
  const wd = new Intl.DateTimeFormat('de-DE', { weekday: 'short' }).format(d).replace('.', '');
  const day = d.getDate();
  const mon = new Intl.DateTimeFormat('de-DE', { month: 'short' }).format(d).replace('.', '');
  return (
    <div style={{ width: 52, height: 56, borderRadius: 14, background: 'var(--surface-2)', border: '1px solid var(--line)', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
      <span style={{ fontSize: 10.5, fontWeight: 800, color: 'var(--muted)', textTransform: 'uppercase' }}>{wd}</span>
      <span style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 21, lineHeight: 1 }}>{day}</span>
      <span style={{ fontSize: 10, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase' }}>{mon}</span>
    </div>
  );
}

// ── Discover (Entdecken) ──
function MemberDiscover() {
  const { state, nav } = useApp();
  const SM = window.SM;
  const [filter, setFilter] = useState('alle');
  const pub = state.events.filter(e => e.status === 'veröffentlicht');

  const filters = [
    { id: 'alle', label: 'Alle' },
    { id: 'frei', label: 'Freie Plätze' },
    { id: 'Turnier', label: 'Turnier' },
    { id: 'Vereinsleben', label: 'Vereinsleben' },
    { id: 'Mitgliederwerbung', label: 'Werbung' },
  ];
  const shown = pub.filter(ev => {
    if (filter === 'alle') return true;
    if (filter === 'frei') return ev.days.some(d => d.shifts.some(s => SM.occ(s).free > 0));
    return ev.category === filter;
  });

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">Veranstaltungen</div>
          <div className="sm-title">Entdecken</div>
        </div>
        <IconCircle name="calendar" onClick={() => {}} />
      </div>
      <div style={{ padding: '0 18px 4px' }}>
        <div className="sm-chiprow">
          {filters.map(f => (
            <button key={f.id} className={'sm-chip' + (filter === f.id ? ' active' : '')} onClick={() => setFilter(f.id)}>
              {f.id === 'frei' && <Icon name="filter" size={14} stroke={2.2} />}{f.label}
            </button>
          ))}
        </div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 14 }}>
        {shown.map(ev => <EventCard key={ev.id} ev={ev} onClick={() => nav.push('event', { id: ev.id })} />)}
        {shown.length === 0 && <EmptyState icon="compass" title="Keine Treffer" text="Für diesen Filter gibt es gerade keine Veranstaltungen." />}
      </div>
    </div>
  );
}

function EventCard({ ev, onClick }) {
  const SM = window.SM;
  const allShifts = ev.days.flatMap(d => d.shifts);
  const free = allShifts.reduce((a, s) => a + SM.occ(s).free, 0);
  const needs = allShifts.some(s => SM.occ(s).needsMore);
  const d0 = ev.days[0].date, d1 = ev.days[ev.days.length - 1].date;
  const dateLabel = ev.days.length > 1 ? `${SM.fmtDate(d0, 'daymon')} – ${SM.fmtDate(d1, 'daymon')}` : SM.fmtDate(d0, 'weekday');
  return (
    <div className="sm-card pressable" style={{ marginBottom: 12, overflow: 'hidden' }} onClick={onClick}>
      <div style={{ height: 74, background: catGradient(ev.category), position: 'relative', display: 'flex', alignItems: 'flex-end', padding: 12 }}>
        <span className="sm-badge" style={{ background: 'rgba(255,255,255,0.92)', color: 'var(--ink)', backdropFilter: 'blur(4px)' }}>
          <Icon name="tag" size={12} stroke={2.2} />{ev.category}
        </span>
        {ev.days.length > 1 && <span style={{ position: 'absolute', top: 12, right: 12 }} className="sm-badge b-neutral">{ev.days.length} Tage</span>}
      </div>
      <div style={{ padding: 15 }}>
        <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 18, lineHeight: 1.15, marginBottom: 7 }}>{ev.name}</div>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 5, color: 'var(--ink-2)', fontSize: 13, fontWeight: 600 }}>
          <span style={{ display: 'flex', alignItems: 'center', gap: 7 }}><Icon name="calendar" size={15} stroke={2} color="var(--muted)" />{dateLabel}</span>
          <span style={{ display: 'flex', alignItems: 'center', gap: 7 }}><Icon name="pin" size={15} stroke={2} color="var(--muted)" />{ev.location}</span>
        </div>
        <hr className="sm-divider" style={{ margin: '13px 0' }} />
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          {free > 0
            ? <span style={{ fontSize: 13.5, fontWeight: 700 }}>{free} freie {free === 1 ? 'Platz' : 'Plätze'}</span>
            : <span style={{ fontSize: 13.5, fontWeight: 700, color: 'var(--muted)' }}>Alle Schichten besetzt</span>}
          <span style={{ display: 'flex', alignItems: 'center', gap: 5, fontWeight: 700, fontSize: 13.5, color: needs ? 'var(--crit)' : 'var(--ink)' }}>
            {needs && <span className="sm-badge b-crit" style={{ padding: '3px 7px' }}>Helfer gesucht</span>}
            <Icon name="arrowR" size={17} stroke={2.2} />
          </span>
        </div>
      </div>
    </div>
  );
}

function catGradient(cat) {
  const map = {
    'Turnier': 'linear-gradient(120deg, #2A2722 0%, #4a4336 100%)',
    'Vereinsleben': 'linear-gradient(120deg, #6b5a1f 0%, #b89436 100%)',
    'Mitgliederwerbung': 'linear-gradient(120deg, #3a4a3f 0%, #5f7a64 100%)',
  };
  return map[cat] || 'linear-gradient(120deg,#444,#666)';
}

// ── My shifts ──
function MyShifts() {
  const { state, nav } = useApp();
  const SM = window.SM, uid = SM.currentUserId;
  const [tab, setTab] = useState('up');
  const { up, past } = collectMy(state.events, uid);
  const list = tab === 'up' ? up : past;
  return (
    <div className="fade-in">
      <div className="sm-header"><div><div className="sm-eyebrow">Mein Engagement</div><div className="sm-title">Meine Schichten</div></div></div>
      <div style={{ padding: '0 18px' }}>
        <div style={{ display: 'flex', gap: 4, background: 'var(--surface-2)', borderRadius: 13, padding: 4, border: '1px solid var(--line)' }}>
          {[['up', 'Bevorstehend', up.length], ['past', 'Vergangen', past.length]].map(([k, l, n]) => (
            <button key={k} onClick={() => setTab(k)} style={{ flex: 1, border: 'none', cursor: 'pointer', padding: '9px 0', borderRadius: 10, fontWeight: 700, fontSize: 14, fontFamily: 'inherit', background: tab === k ? 'var(--surface)' : 'transparent', color: tab === k ? 'var(--ink)' : 'var(--muted)', boxShadow: tab === k ? 'var(--shadow)' : 'none' }}>{l} · {n}</button>
          ))}
        </div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 14 }}>
        {list.length === 0 && <EmptyState icon="calendar" title={tab === 'up' ? 'Keine kommenden Schichten' : 'Noch keine Historie'} text={tab === 'up' ? 'Finde freie Schichten im Tab „Entdecken".' : 'Abgeschlossene Schichten erscheinen hier.'} />}
        {list.map(r => tab === 'up'
          ? <MyShiftCard key={r.sh.id} r={r} onClick={() => nav.push('event', { id: r.ev.id })} />
          : <PastShiftCard key={r.sh.id} r={r} />)}
      </div>
    </div>
  );
}

function PastShiftCard({ r }) {
  const SM = window.SM;
  const no = r.su.status === 'nichterschienen';
  return (
    <div className="sm-card" style={{ padding: 14, marginBottom: 10, display: 'flex', gap: 13, alignItems: 'center', opacity: no ? 0.85 : 1 }}>
      <DateChip date={r.day.date} />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontWeight: 700, fontSize: 15.5 }}>{r.sh.name}</div>
        <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{r.ev.name}</div>
        <div style={{ marginTop: 7 }}>
          {no ? <Badge kind="crit">Nicht erschienen · 0 h</Badge> : <Badge kind="ok">Bestätigt · {SM.hrs(r.su.hours || 0)}</Badge>}
        </div>
      </div>
    </div>
  );
}

// ── Profile ──
function MemberProfile() {
  const { settings, setSetting, role, setRole } = useApp();
  const SM = window.SM, me = SM.M[SM.currentUserId];
  const [prefs, setPrefs] = useState({ week: true, day: true, news: false });
  return (
    <div className="fade-in">
      <div className="sm-header"><div><div className="sm-eyebrow">Konto</div><div className="sm-title">Profil</div></div></div>
      <div className="sm-pad" style={{ paddingTop: 8 }}>
        <div className="sm-card pad" style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <Avatar id={me.id} size={56} />
          <div>
            <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 19 }}>{me.first} {me.last}</div>
            <div style={{ color: 'var(--muted)', fontSize: 13, fontWeight: 600 }}>{me.email}</div>
            <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600, marginTop: 2 }}>Mitglied seit {SM.fmtDate(me.since, 'short')}</div>
          </div>
        </div>

        <Section title="Erinnerungen" />
        <div className="sm-card">
          <PrefRow label="Erinnerung 1 Woche vorher" sub="7 Tage vor Schichtbeginn" on={prefs.week} set={() => setPrefs(p => ({ ...p, week: !p.week }))} />
          <hr className="sm-divider" />
          <PrefRow label="Erinnerung 1 Tag vorher" sub="24 h vor Schichtbeginn" on={prefs.day} set={() => setPrefs(p => ({ ...p, day: !p.day }))} />
          <hr className="sm-divider" />
          <PrefRow label="Neue Veranstaltungen" sub="Bei Veröffentlichung benachrichtigen" on={prefs.news} set={() => setPrefs(p => ({ ...p, news: !p.news }))} />
        </div>
        <div className="sm-hint" style={{ padding: '0 4px' }}>Pflicht-Mails (z. B. Absage einer Schicht) können nicht deaktiviert werden.</div>

        <Section title="Datenschutz" />
        <div className="sm-card">
          <LinkRow icon="download" label="Meine Daten exportieren" sub="Auskunftsrecht (DSGVO)" />
          <hr className="sm-divider" />
          <LinkRow icon="shield" label="Löschung beantragen" sub="Recht auf Vergessen" />
        </div>

        <Section title="Demo" />
        <div className="sm-card pad" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <div style={{ fontWeight: 700, fontSize: 14.5 }}>Vorstands-Ansicht testen</div>
            <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>Wechselt in die Admin-Oberfläche</div>
          </div>
          <Button variant="dark" size="sm" icon="arrowR" onClick={() => setRole('vorstand')}>Wechseln</Button>
        </div>
      </div>
    </div>
  );
}
function PrefRow({ label, sub, on, set }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '13px 15px' }}>
      <div style={{ flex: 1 }}>
        <div style={{ fontWeight: 700, fontSize: 14.5 }}>{label}</div>
        <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{sub}</div>
      </div>
      <Toggle on={on} onClick={set} />
    </div>
  );
}
function LinkRow({ icon, label, sub, onClick }) {
  return (
    <div className="pressable" onClick={onClick} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '13px 15px', cursor: 'pointer' }}>
      <div style={{ width: 36, height: 36, borderRadius: 10, background: 'var(--surface-2)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}><Icon name={icon} size={18} color="var(--ink-2)" /></div>
      <div style={{ flex: 1 }}>
        <div style={{ fontWeight: 700, fontSize: 14.5 }}>{label}</div>
        {sub && <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{sub}</div>}
      </div>
      <Icon name="chevR" size={18} color="var(--line-2)" stroke={2.4} />
    </div>
  );
}

// shared small bits
function IconCircle({ name, onClick, badge }) {
  return (
    <button onClick={onClick} className="pressable" style={{ position: 'relative', width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }}>
      <Icon name={name} size={20} color="var(--ink)" />
      {badge && <span style={{ position: 'absolute', top: 9, right: 10, width: 8, height: 8, borderRadius: '50%', background: 'var(--crit)', border: '1.5px solid var(--surface)' }} />}
    </button>
  );
}
function TextLink({ onClick, children }) { return <span className="pressable" onClick={onClick} style={{ fontWeight: 700, fontSize: 13.5, color: 'var(--ink-2)', cursor: 'pointer', display: 'inline-flex', alignItems: 'center', gap: 3 }}>{children}<Icon name="chevR" size={14} stroke={2.4} /></span>; }
function EmptyState({ icon, title, text }) {
  return (
    <div style={{ textAlign: 'center', padding: '34px 20px' }}>
      <div style={{ width: 58, height: 58, borderRadius: '50%', background: 'var(--surface-2)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}><Icon name={icon} size={26} color="var(--muted)" /></div>
      <div style={{ fontWeight: 700, fontSize: 16 }}>{title}</div>
      <div style={{ color: 'var(--muted)', fontSize: 13.5, fontWeight: 600, marginTop: 4, maxWidth: 250, marginInline: 'auto' }}>{text}</div>
    </div>
  );
}

Object.assign(window, {
  MemberDashboard, MemberDiscover, MyShifts, MemberProfile,
  EventCard, MyShiftCard, DateChip, IconCircle, TextLink, EmptyState, collectMy, catGradient, Legend,
});
