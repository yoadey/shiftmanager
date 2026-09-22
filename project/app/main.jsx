/* App root — state, navigation, role switch, tweaks */
const { useState: uS, useEffect: uE, useRef: uR, useMemo } = React;

const TWEAK_DEFAULTS = /*EDITMODE-BEGIN*/{
  "primaryColor": "#F4B63F",
  "radius": "standard",
  "warmth": "warm"
}/*EDITMODE-END*/;

function pickOn(hex) {
  const h = hex.replace('#', '');
  const r = parseInt(h.slice(0, 2), 16), g = parseInt(h.slice(2, 4), 16), b = parseInt(h.slice(4, 6), 16);
  const lum = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
  return lum > 0.6 ? '#1A1813' : '#ffffff';
}

const clone = (o) => JSON.parse(JSON.stringify(o));

function App() {
  const [t, setTweak] = useTweaks(TWEAK_DEFAULTS);
  const [role, setRoleRaw] = uS(() => localStorage.getItem('sm_role') || 'mitglied');
  const [tab, setTab] = uS(role === 'mitglied' ? 'start' : 'start');
  const [stack, setStack] = uS([]);
  const [createOpen, setCreateOpen] = uS(false);
  const [toast, setToast] = uS(null);
  const toastTimer = uR();

  const [state, setState] = uS(() => ({
    events: clone(window.SM.events),
    manualBookings: clone(window.SM.manualBookings),
    settings: clone(window.SM.settings),
    audit: clone(window.SM.audit),
  }));

  // apply branding tweaks
  uE(() => {
    const r = document.documentElement;
    r.style.setProperty('--primary', t.primaryColor);
    r.style.setProperty('--on-primary', pickOn(t.primaryColor));
    const radii = { klein: ['12px', '9px'], standard: ['20px', '13px'], weich: ['28px', '18px'] };
    const [rad, radSm] = radii[t.radius] || radii.standard;
    r.style.setProperty('--radius', rad); r.style.setProperty('--radius-sm', radSm);
    if (t.warmth === 'neutral') {
      r.style.setProperty('--bg', '#F5F5F3'); r.style.setProperty('--surface-2', '#F3F3F1');
      r.style.setProperty('--line', '#EAEAE7'); r.style.setProperty('--line-2', '#DEDEDA');
    } else {
      r.style.setProperty('--bg', '#FBF7EF'); r.style.setProperty('--surface-2', '#FAF6EC');
      r.style.setProperty('--line', '#ECE6D8'); r.style.setProperty('--line-2', '#E2DBCB');
    }
  }, [t.primaryColor, t.radius, t.warmth]);

  uE(() => { localStorage.setItem('sm_role', role); }, [role]);

  const showToast = (msg, kind = 'ok') => {
    setToast({ msg, kind });
    clearTimeout(toastTimer.current);
    toastTimer.current = setTimeout(() => setToast(null), 2600);
  };

  const nav = {
    tab, stack,
    go: (tb) => { setStack([]); setTab(tb); },
    push: (name, params) => setStack(s => [...s, { name, params: params || {} }]),
    back: () => setStack(s => s.slice(0, -1)),
  };

  const setRole = (rr) => { setStack([]); setTab('start'); setRoleRaw(rr); };

  // ── actions ──
  const register = (shiftId, comment, otherEmail) => {
    setState(prev => {
      const ns = clone(prev);
      ns.events.forEach(ev => ev.days.forEach(d => d.shifts.forEach(sh => {
        if (sh.id === shiftId) {
          if (otherEmail) sh.signups.push({ memberId: '_guest_' + Date.now(), guest: otherEmail, status: 'reserviert' });
          else sh.signups.push({ memberId: window.SM.currentUserId, status: 'angemeldet', comment: comment || undefined });
        }
      })));
      return ns;
    });
    if (otherEmail) showToast('Bestätigungslink an ' + otherEmail + ' gesendet', 'warn');
    else showToast('Angemeldet – zählt als Reservierung', 'ok');
  };

  const deregister = (shiftId) => {
    setState(prev => {
      const ns = clone(prev);
      ns.events.forEach(ev => ev.days.forEach(d => d.shifts.forEach(sh => {
        if (sh.id === shiftId) sh.signups = sh.signups.filter(s => s.memberId !== window.SM.currentUserId);
      })));
      return ns;
    });
    showToast('Von Schicht abgemeldet', 'crit');
  };

  const createEvent = (ev) => {
    setState(prev => { const ns = clone(prev); ns.events.unshift(ev); return ns; });
    setTab('events'); setStack([]);
    showToast(ev.status === 'veröffentlicht' ? 'Veranstaltung veröffentlicht' : 'Als Entwurf gespeichert', 'ok');
  };

  const setSetting = (k, v) => setState(prev => { const ns = clone(prev); ns.settings[k] = v; return ns; });

  const manualBook = (mid, date, hours, desc) => {
    setState(prev => {
      const ns = clone(prev);
      ns.manualBookings.push({ id: 'mb-' + Date.now(), memberId: mid, date, hours, desc, by: 'Vorstand' });
      ns.audit.unshift({ ts: '2026-05-30 ' + new Date().toTimeString().slice(0, 5), who: 'Du (Vorstand)', what: `Manuelle Buchung +${window.SM.hrs(hours)} für ${window.SM.M[mid].first} ${window.SM.M[mid].last}`, cat: 'Stunden' });
      return ns;
    });
    showToast(window.SM.hrs(hours) + ' gebucht', 'ok');
  };

  const ctx = { state, nav, role, setRole, showToast, settings: state.settings, setSetting, register, deregister, createEvent, manualBook, openCreate: () => setCreateOpen(true), t };

  // ── tabs ──
  const memberTabs = [['start', 'Start', 'home'], ['entdecken', 'Entdecken', 'compass'], ['schichten', 'Schichten', 'calendar'], ['profil', 'Profil', 'user']];
  const adminTabs = [['start', 'Übersicht', 'chart'], ['events', 'Termine', 'calendar'], ['mitglieder', 'Mitglieder', 'users'], ['settings', 'Einstellungen', 'settings']];
  const tabs = role === 'mitglied' ? memberTabs : adminTabs;

  function renderScreen() {
    if (stack.length) {
      const top = stack[stack.length - 1];
      if (top.name === 'event') return <EventDetail id={top.params.id} />;
      if (top.name === 'member') return <MemberDetail id={top.params.id} />;
      if (top.name === 'manual') return <ManualBooking id={top.params.id} />;
      if (top.name === 'audit') return <AuditLog />;
    }
    if (role === 'mitglied') {
      if (tab === 'start') return <MemberDashboard />;
      if (tab === 'entdecken') return <MemberDiscover />;
      if (tab === 'schichten') return <MyShifts />;
      if (tab === 'profil') return <MemberProfile />;
    } else {
      if (tab === 'start') return <AdminDashboard />;
      if (tab === 'events') return <AdminEvents />;
      if (tab === 'mitglieder') return <AdminMembers />;
      if (tab === 'settings') return <AdminSettings />;
    }
    return null;
  }

  // ── scaling / responsive ──
  const [scale, setScale] = uS(1);
  const [isDesktop, setIsDesktop] = uS(() => window.innerWidth >= 900);
  uE(() => {
    const fit = () => {
      setIsDesktop(window.innerWidth >= 900);
      const pad = window.innerWidth < 520 ? 0 : 24;
      const s = Math.min(1, (window.innerHeight - 86 - pad) / 874, (window.innerWidth - pad) / 402);
      setScale(Math.max(0.4, s));
    };
    fit(); window.addEventListener('resize', fit); return () => window.removeEventListener('resize', fit);
  }, []);

  const toastEl = toast && (
    <div className="sm-toast-wrap">
      <div className="sm-toast"><span className="tdot" style={{ background: toast.kind === 'crit' ? 'var(--crit)' : toast.kind === 'warn' ? 'var(--warn)' : 'var(--ok)' }} />{toast.msg}</div>
    </div>
  );

  const tweaks = (
    <TweaksPanel>
      <TweakSection label="Branding" />
      <TweakColor label="Primärfarbe" value={t.primaryColor}
        options={['#F4B63F', '#6C4AD9', '#2A6FDB', '#1F8A5B', '#C0392B']}
        onChange={v => setTweak('primaryColor', v)} />
      <TweakRadio label="Hintergrund" value={t.warmth} options={['warm', 'neutral']} onChange={v => setTweak('warmth', v)} />
      <TweakSection label="Form" />
      <TweakRadio label="Ecken" value={t.radius} options={['klein', 'standard', 'weich']} onChange={v => setTweak('radius', v)} />
    </TweaksPanel>
  );

  // ════════ DESKTOP SHELL ════════
  if (isDesktop) {
    const me = window.SM.M[window.SM.currentUserId];
    return (
      <AppCtx.Provider value={ctx}>
        <div className="stage desktop">
          <aside className="dt-sidebar">
            <div className="dt-brand">
              <span className="rb-logo" style={{ width: 38, height: 38, fontSize: 15 }}>SG</span>
              <div>
                <div className="dt-brand-name">ShiftManager</div>
                <div className="dt-brand-sub">{state.settings.clubName}</div>
              </div>
            </div>
            <nav className="dt-nav">
              {tabs.map(([k, l, ic]) => (
                <button key={k} className={'dt-navitem' + (tab === k && !stack.length ? ' active' : '')} onClick={() => nav.go(k)}>
                  <Icon name={ic} size={20} stroke={2} />{l}
                </button>
              ))}
            </nav>
            <div className="dt-side-foot">
              <div className="dt-role-label">Ansicht wechseln</div>
              <div className="dt-roleseg">
                {[['mitglied', 'Mitglied'], ['vorstand', 'Vorstand']].map(([v, l]) => (
                  <button key={v} className={role === v ? 'active' : ''} onClick={() => setRole(v)}>{l}</button>
                ))}
              </div>
              <div className="dt-user">
                <Avatar id={me.id} size={36} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div className="dt-user-name">{me.first} {me.last}</div>
                  <div className="dt-user-role">{role === 'mitglied' ? 'Mitglied' : 'Vorstand'}</div>
                </div>
              </div>
            </div>
          </aside>
          <main className="dt-content">
            <div className="sm-app desktop">
              <div className="sm-main" key={role + tab + stack.length}>{renderScreen()}</div>
              {toastEl}
              {createOpen && <CreateEventFlow onClose={() => setCreateOpen(false)} />}
            </div>
          </main>
          {tweaks}
        </div>
      </AppCtx.Provider>
    );
  }

  // ════════ MOBILE SHELL ════════
  return (
    <AppCtx.Provider value={ctx}>
      <div className="stage">
        {/* role switch chrome */}
        <div className="rolebar">
          <div className="rb-brand"><span className="rb-logo">SG</span><span>ShiftManager</span></div>
          <div className="rb-seg">
            {[['mitglied', 'Mitglied'], ['vorstand', 'Vorstand']].map(([v, l]) => (
              <button key={v} className={'rb-btn' + (role === v ? ' active' : '')} onClick={() => setRole(v)}>{l}</button>
            ))}
          </div>
        </div>

        <div className="scaler" style={{ transform: `scale(${scale})` }}>
          <IOSDevice>
            <div className="sm-app">
              <div className="sm-main" key={role + tab + stack.length}>{renderScreen()}</div>
              {stack.length === 0 && (
                <nav className="sm-nav">
                  {tabs.map(([k, l, ic]) => (
                    <button key={k} className={'sm-navitem' + (tab === k ? ' active' : '')} onClick={() => nav.go(k)}>
                      <span className="ni-ico"><Icon name={ic} size={21} stroke={tab === k ? 2.4 : 2} /></span>{l}
                    </button>
                  ))}
                </nav>
              )}

              {toastEl}

              {createOpen && <CreateEventFlow onClose={() => setCreateOpen(false)} />}
            </div>
          </IOSDevice>
        </div>

        {tweaks}
      </div>
    </AppCtx.Provider>
  );
}

ReactDOM.createRoot(document.getElementById('root')).render(<App />);
