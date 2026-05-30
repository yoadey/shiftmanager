/* Admin / Vorstand screens */

function AdminDashboard() {
  const { state, nav, setRole, openCreate } = useApp();
  const SM = window.SM;
  let totalConfirmed = 0, openSlots = 0;
  const understaffed = [];
  state.events.forEach(ev => {
    if (ev.status === 'abgesagt' || ev.status === 'entwurf') return;
    ev.days.forEach(d => d.shifts.forEach(sh => {
      sh.signups.forEach(s => { if (s.status === 'bestätigt') totalConfirmed += s.hours || 0; });
      const o = SM.occ(sh);
      openSlots += o.free;
      if (o.needsMore && ev.status === 'veröffentlicht' && d.date >= '2026-05-30') understaffed.push({ ev, d, sh, o });
    }));
  });
  const activeEvents = state.events.filter(e => e.status === 'veröffentlicht').length;

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div><div className="sm-eyebrow">{settingsName(SM)}</div><div className="sm-title">Übersicht</div></div>
        <IconCircle name="user" onClick={() => setRole('mitglied')} />
      </div>
      <div className="sm-pad">
        <div className="sm-stat-grid">
          <StatCard icon="hours" label="Geleistete Stunden" val={SM.hrs(totalConfirmed)} sub={`Vereinsjahr ${state.settings.clubYear}`} />
          <StatCard icon="users" label="Mitglieder" val={SM.members.length} sub="aktiv" />
          <StatCard icon="calendar" label="Veröffentlicht" val={activeEvents} sub="Veranstaltungen" />
          <StatCard icon="layers" label="Offene Plätze" val={openSlots} sub="über alle Schichten" accent />
        </div>

        <Button icon="plus" onClick={openCreate} style={{ marginTop: 14 }}>Neue Veranstaltung</Button>
        <div style={{ height: 14 }} />
        <Button variant="soft" icon="plus" onClick={() => nav.push('manual')}>Stunden manuell buchen</Button>

        <Section title="Unterbesetzte Schichten" />
        {understaffed.length === 0
          ? <EmptyState icon="check" title="Alles im grünen Bereich" text="Keine bevorstehende Schicht unter dem Mindesthelfer-Wert." />
          : understaffed.map(({ ev, d, sh, o }) => (
            <div key={sh.id} className="sm-card pressable" style={{ padding: 13, marginBottom: 10, display: 'flex', gap: 12, alignItems: 'center', borderLeft: '3px solid var(--crit)' }} onClick={() => nav.push('event', { id: ev.id })}>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontWeight: 700, fontSize: 15 }}>{sh.name}</div>
                <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{ev.name} · {SM.fmtDate(d.date, 'daymon')}</div>
              </div>
              <Badge kind={o.key === 'crit' ? 'crit' : 'warn'}>{o.count}/{sh.min} min.</Badge>
            </div>
          ))}
      </div>
    </div>
  );
}
function settingsName(SM) { return SM.settings.clubName; }

function StatCard({ icon, label, val, sub, accent }) {
  return (
    <div className="sm-card" style={{ padding: 14, background: accent ? 'var(--ink)' : 'var(--surface)' }}>
      <Icon name={icon} size={20} color={accent ? 'var(--primary)' : 'var(--muted)'} />
      <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 27, marginTop: 8, color: accent ? '#fff' : 'var(--ink)', lineHeight: 1 }}>{val}</div>
      <div style={{ fontWeight: 700, fontSize: 12.5, marginTop: 4, color: accent ? 'rgba(255,255,255,0.92)' : 'var(--ink-2)' }}>{label}</div>
      <div style={{ fontSize: 11, fontWeight: 600, color: accent ? 'rgba(255,255,255,0.55)' : 'var(--muted)' }}>{sub}</div>
    </div>
  );
}

// ── Veranstaltungen ──
function AdminEvents() {
  const { state, nav, openCreate } = useApp();
  const SM = window.SM;
  const order = { 'veröffentlicht': 0, 'entwurf': 1, 'abgeschlossen': 2, 'abgesagt': 3 };
  const evs = [...state.events].sort((a, b) => (order[a.status] - order[b.status]) || a.days[0].date.localeCompare(b.days[0].date));
  const statusKind = { 'veröffentlicht': 'ok', 'entwurf': 'neutral', 'abgeschlossen': 'full', 'abgesagt': 'crit' };
  return (
    <div className="fade-in">
      <div className="sm-header">
        <div><div className="sm-eyebrow">Verwaltung</div><div className="sm-title">Veranstaltungen</div></div>
        <IconCircle name="plus" onClick={openCreate} />
      </div>
      <div className="sm-pad" style={{ paddingTop: 8 }}>
        {evs.map(ev => {
          const shifts = ev.days.flatMap(d => d.shifts);
          const filled = shifts.reduce((a, s) => a + SM.occ(s).count, 0);
          const cap = shifts.reduce((a, s) => a + s.max, 0);
          return (
            <div key={ev.id} className="sm-card pressable" style={{ padding: 15, marginBottom: 11 }} onClick={() => nav.push('event', { id: ev.id })}>
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: 10, alignItems: 'flex-start' }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 16.5, lineHeight: 1.15 }}>{ev.name}</div>
                  <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600, marginTop: 3 }}>{SM.fmtDate(ev.days[0].date, 'daymon')}{ev.days.length > 1 ? ` – ${SM.fmtDate(ev.days[ev.days.length - 1].date, 'daymon')}` : ''} · {ev.location}</div>
                </div>
                <Badge kind={statusKind[ev.status]} dot={ev.status === 'veröffentlicht'}>{ev.status[0].toUpperCase() + ev.status.slice(1)}</Badge>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 12 }}>
                <div style={{ flex: 1 }}><div className="occ-track"><div className="occ-fill" style={{ width: (cap ? filled / cap * 100 : 0) + '%', background: 'var(--primary)' }} /></div></div>
                <span style={{ fontSize: 12, fontWeight: 800, color: 'var(--ink-2)' }}>{filled}/{cap} besetzt · {shifts.length} Schichten</span>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

// ── Mitglieder ──
function AdminMembers() {
  const { state, nav } = useApp();
  const SM = window.SM;
  const [q, setQ] = useState('');
  const hours = {};
  SM.members.forEach(m => { hours[m.id] = 0; });
  state.events.forEach(ev => ev.days.forEach(d => d.shifts.forEach(sh => sh.signups.forEach(s => { if (s.status === 'bestätigt') hours[s.memberId] = (hours[s.memberId] || 0) + (s.hours || 0); }))));
  state.manualBookings.forEach(b => { hours[b.memberId] = (hours[b.memberId] || 0) + b.hours; });
  const list = SM.members.filter(m => (m.first + ' ' + m.last).toLowerCase().includes(q.toLowerCase()));

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div><div className="sm-eyebrow">{SM.members.length} Einträge</div><div className="sm-title">Mitglieder</div></div>
        <IconCircle name="download" onClick={() => {}} />
      </div>
      <div style={{ padding: '0 18px 4px' }}>
        <div style={{ position: 'relative' }}>
          <Icon name="search" size={18} color="var(--muted)" style={{ position: 'absolute', left: 13, top: 13 }} />
          <Input placeholder="Mitglied suchen…" value={q} onChange={e => setQ(e.target.value)} style={{ paddingLeft: 40 }} />
        </div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 12 }}>
        <div className="sm-card" style={{ overflow: 'hidden' }}>
          {list.map((m, i) => {
            const goal = m.goal || state.settings.yearGoal;
            const h = hours[m.id] || 0;
            return (
              <React.Fragment key={m.id}>
                {i > 0 && <hr className="sm-divider" style={{ marginLeft: 64 }} />}
                <div className="pressable" onClick={() => nav.push('member', { id: m.id })} style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px', cursor: 'pointer' }}>
                  <Avatar id={m.id} size={40} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: 700, fontSize: 14.5 }}>{m.first} {m.last}{m.id === SM.currentUserId && <span style={{ color: 'var(--muted)', fontWeight: 600 }}> · du</span>}</div>
                    <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{SM.hrs(h)} / {goal} h{m.goal && ' · ind. Ziel'}</div>
                  </div>
                  <div style={{ width: 46 }}><div className="occ-track"><div className="occ-fill" style={{ width: Math.min(h / goal * 100, 100) + '%', background: h >= goal ? 'var(--ok)' : 'var(--primary)' }} /></div></div>
                  <Icon name="chevR" size={16} color="var(--line-2)" stroke={2.4} />
                </div>
              </React.Fragment>
            );
          })}
        </div>
      </div>
    </div>
  );
}

function MemberDetail({ id }) {
  const { state, nav } = useApp();
  const SM = window.SM, m = SM.M[id];
  let h = 0; const recs = [];
  state.events.forEach(ev => ev.days.forEach(d => d.shifts.forEach(sh => sh.signups.forEach(s => { if (s.memberId === id && s.status === 'bestätigt') { h += s.hours || 0; recs.push({ t: `${ev.name} · ${sh.name}`, d: d.date, h: s.hours }); } }))));
  state.manualBookings.forEach(b => { if (b.memberId === id) { h += b.hours; recs.push({ t: b.desc, d: b.date, h: b.hours, manual: true }); } });
  recs.sort((a, b) => b.d.localeCompare(a.d));
  const goal = m.goal || state.settings.yearGoal;
  return (
    <div className="fade-in">
      <div className="sm-header detail" style={{ alignItems: 'center', gap: 12 }}>
        <button onClick={nav.back} className="pressable" style={{ width: 40, height: 40, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}><Icon name="chevL" size={20} stroke={2.4} /></button>
        <div className="sm-title" style={{ fontSize: 22 }}>Mitglied</div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 4 }}>
        <div className="sm-card pad" style={{ display: 'flex', gap: 14, alignItems: 'center' }}>
          <Avatar id={id} size={56} />
          <div style={{ flex: 1 }}>
            <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 19 }}>{m.first} {m.last}</div>
            <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{m.email}</div>
            <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>seit {SM.fmtDate(m.since, 'short')}</div>
          </div>
        </div>
        <div className="sm-card pad" style={{ marginTop: 12 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
            <span style={{ fontWeight: 700, color: 'var(--ink-2)', fontSize: 14 }}>Stundenkonto {state.settings.clubYear}</span>
            <span><b style={{ fontFamily: 'Bricolage Grotesque', fontSize: 19 }}>{SM.hrs(h)}</b> <span style={{ color: 'var(--muted)', fontWeight: 700 }}>/ {goal} h</span></span>
          </div>
          <div style={{ marginTop: 10 }}><div className="occ-track" style={{ height: 9 }}><div className="occ-fill" style={{ width: Math.min(h / goal * 100, 100) + '%', background: h >= goal ? 'var(--ok)' : 'var(--primary)' }} /></div></div>
          <div style={{ display: 'flex', gap: 9, marginTop: 13 }}>
            <Button variant="soft" size="sm" icon="plus" onClick={() => nav.push('manual', { id })}>Stunden buchen</Button>
            <Button variant="soft" size="sm" icon="edit">Ziel anpassen</Button>
          </div>
        </div>
        <Section title="Gebuchte Stunden" />
        <div className="sm-card" style={{ overflow: 'hidden' }}>
          {recs.length === 0 && <div style={{ padding: 16, color: 'var(--muted)', fontWeight: 600, fontSize: 13.5 }}>Noch keine bestätigten Stunden.</div>}
          {recs.map((r, i) => (
            <React.Fragment key={i}>
              {i > 0 && <hr className="sm-divider" />}
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '12px 14px' }}>
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: 700, fontSize: 14 }}>{r.t}</div>
                  <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{SM.fmtDate(r.d, 'short')}{r.manual && ' · manuell'}</div>
                </div>
                {r.manual && <Badge kind="primary" dot={false}>manuell</Badge>}
                <span style={{ fontWeight: 800, fontFamily: 'Bricolage Grotesque' }}>{SM.hrs(r.h)}</span>
              </div>
            </React.Fragment>
          ))}
        </div>
      </div>
    </div>
  );
}

Object.assign(window, { AdminDashboard, AdminEvents, AdminMembers, MemberDetail, StatCard });
