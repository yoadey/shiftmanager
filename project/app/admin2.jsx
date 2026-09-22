/* Create-event flow, settings, manual booking, audit */

function datesBetween(start, end) {
  const out = []; let d = new Date(start + 'T12:00:00'); const e = new Date(end + 'T12:00:00');
  while (d <= e) { out.push(d.toISOString().slice(0, 10)); d.setDate(d.getDate() + 1); }
  return out.length ? out : [start];
}

function CreateEventFlow({ onClose }) {
  const { createEvent } = useApp();
  const SM = window.SM;
  const [step, setStep] = useState(0);
  const [f, setF] = useState({ name: '', description: '', location: '', category: 'Turnier', status: 'veröffentlicht', multiDay: false, startDate: '2026-08-22', endDate: '2026-08-23' });
  const [shifts, setShifts] = useState([]);
  const set = (k, v) => setF(p => ({ ...p, [k]: v }));
  const dates = f.multiDay ? datesBetween(f.startDate, f.endDate) : [f.startDate];

  const step0valid = f.name.trim() && f.location.trim() && f.startDate && (!f.multiDay || f.endDate >= f.startDate);
  const labels = ['Eckdaten', 'Schichten', 'Prüfen'];

  const finish = () => {
    const byDate = {};
    dates.forEach(d => { byDate[d] = []; });
    shifts.forEach(s => { (byDate[s.date] || (byDate[s.date] = [])).push({ id: 'ns-' + Math.random().toString(36).slice(2, 7), name: s.name, start: s.start, end: s.end, min: s.min, max: s.max, desc: s.desc, signups: [] }); });
    const days = Object.keys(byDate).sort().filter(d => byDate[d].length).map(d => ({ date: d, shifts: byDate[d].sort((a, b) => a.start.localeCompare(b.start)) }));
    createEvent({ id: 'ev-' + Date.now(), name: f.name.trim(), description: f.description.trim() || 'Keine Beschreibung.', location: f.location.trim(), category: f.category, status: f.status, days });
    onClose();
  };

  return (
    <Sheet variant="full" onClose={onClose}>
      {/* header */}
      <div className="cef-head" style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '54px 18px 14px', borderBottom: '1px solid var(--line)' }}>
        <button onClick={step === 0 ? onClose : () => setStep(step - 1)} className="pressable" style={{ width: 38, height: 38, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}>
          <Icon name={step === 0 ? 'x' : 'chevL'} size={19} stroke={2.4} />
        </button>
        <div style={{ flex: 1 }}>
          <div className="sm-eyebrow">Schritt {step + 1} von 3</div>
          <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 20 }}>{labels[step]}</div>
        </div>
      </div>
      {/* progress */}
      <div style={{ display: 'flex', gap: 6, padding: '12px 18px 0' }}>
        {labels.map((l, i) => <div key={l} style={{ flex: 1, height: 4, borderRadius: 4, background: i <= step ? 'var(--primary)' : 'var(--line)', transition: 'background .2s' }} />)}
      </div>

      <div className="sheet-body" style={{ flex: 1 }}>
        {step === 0 && (
          <div className="fade-in">
            <Field label="Name der Veranstaltung"><Input placeholder="z. B. Sommerturnier 2026" value={f.name} onChange={e => set('name', e.target.value)} /></Field>
            <Field label="Beschreibung (optional)"><Textarea placeholder="Worum geht es? Was sollen Helfer wissen?" value={f.description} onChange={e => set('description', e.target.value)} /></Field>
            <Field label="Ort"><Input placeholder="z. B. Sporthalle Aachen-Brand" value={f.location} onChange={e => set('location', e.target.value)} /></Field>
            <Field label="Kategorie">
              <Select value={f.category} onChange={e => set('category', e.target.value)}>
                {['Turnier', 'Vereinsleben', 'Mitgliederwerbung', 'Training', 'Sonstiges'].map(c => <option key={c}>{c}</option>)}
              </Select>
            </Field>
            <div className="sm-card pad" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 14 }}>
              <div><div style={{ fontWeight: 700, fontSize: 14.5 }}>Mehrtägige Veranstaltung</div><div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>Schichten je Tag planen</div></div>
              <Toggle on={f.multiDay} onClick={() => set('multiDay', !f.multiDay)} />
            </div>
            <div style={{ display: 'flex', gap: 10 }}>
              <Field label={f.multiDay ? 'Von' : 'Datum'}><Input type="date" value={f.startDate} onChange={e => set('startDate', e.target.value)} /></Field>
              {f.multiDay && <Field label="Bis"><Input type="date" value={f.endDate} min={f.startDate} onChange={e => set('endDate', e.target.value)} /></Field>}
            </div>
            <Field label="Sichtbarkeit">
              <div style={{ display: 'flex', gap: 9 }}>
                <ChoiceCard active={f.status === 'entwurf'} onClick={() => set('status', 'entwurf')} icon="edit" title="Entwurf" sub="Nur intern" />
                <ChoiceCard active={f.status === 'veröffentlicht'} onClick={() => set('status', 'veröffentlicht')} icon="users" title="Veröffentlicht" sub="Für Mitglieder" />
              </div>
            </Field>
          </div>
        )}

        {step === 1 && <ShiftBuilder dates={dates} multiDay={f.multiDay} shifts={shifts} setShifts={setShifts} />}

        {step === 2 && (
          <div className="fade-in">
            <div className="sm-card pad">
              <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 19 }}>{f.name || 'Unbenannt'}</div>
              <div style={{ color: 'var(--ink-2)', fontSize: 13, fontWeight: 600, marginTop: 8, display: 'flex', flexDirection: 'column', gap: 6 }}>
                <span style={{ display: 'flex', gap: 8, alignItems: 'center' }}><Icon name="tag" size={15} color="var(--muted)" />{f.category}</span>
                <span style={{ display: 'flex', gap: 8, alignItems: 'center' }}><Icon name="calendar" size={15} color="var(--muted)" />{dates.length > 1 ? `${SM.fmtDate(dates[0], 'daymon')} – ${SM.fmtDate(dates[dates.length - 1], 'daymon')} (${dates.length} Tage)` : SM.fmtDate(dates[0], 'weekday')}</span>
                <span style={{ display: 'flex', gap: 8, alignItems: 'center' }}><Icon name="pin" size={15} color="var(--muted)" />{f.location}</span>
              </div>
              <div style={{ marginTop: 10 }}><Badge kind={f.status === 'veröffentlicht' ? 'ok' : 'neutral'}>{f.status === 'veröffentlicht' ? 'Wird veröffentlicht' : 'Als Entwurf gespeichert'}</Badge></div>
            </div>
            <Section title={`${shifts.length} Schicht${shifts.length === 1 ? '' : 'en'}`} />
            {dates.map((d, di) => {
              const ds = shifts.filter(s => s.date === d);
              if (!ds.length) return null;
              return (
                <div key={d} style={{ marginBottom: 12 }}>
                  {f.multiDay && <div style={{ fontWeight: 700, fontSize: 12.5, color: 'var(--muted)', marginBottom: 7 }}>Tag {di + 1} · {SM.fmtDate(d, 'weekday')}</div>}
                  {ds.map(s => (
                    <div key={s.id} className="sm-card" style={{ padding: 12, marginBottom: 8, display: 'flex', gap: 11, alignItems: 'center' }}>
                      <div style={{ width: 4, alignSelf: 'stretch', borderRadius: 4, background: 'var(--primary)' }} />
                      <div style={{ flex: 1 }}>
                        <div style={{ fontWeight: 700, fontSize: 14.5 }}>{s.name}</div>
                        <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{s.start}–{s.end} · {s.min}–{s.max} Helfer</div>
                      </div>
                    </div>
                  ))}
                </div>
              );
            })}
          </div>
        )}
      </div>

      <div className="sheet-foot">
        {step < 2
          ? <Button disabled={step === 0 ? !step0valid : shifts.length === 0} icon="arrowR" onClick={() => setStep(step + 1)}>{step === 1 && shifts.length === 0 ? 'Mind. eine Schicht' : 'Weiter'}</Button>
          : <Button icon="check" onClick={finish}>Veranstaltung erstellen</Button>}
      </div>
    </Sheet>
  );
}

function ChoiceCard({ active, onClick, icon, title, sub }) {
  return (
    <button onClick={onClick} className="pressable" style={{ flex: 1, textAlign: 'left', cursor: 'pointer', padding: 13, borderRadius: 14, border: active ? '1.5px solid var(--primary)' : '1.5px solid var(--line-2)', background: active ? 'color-mix(in srgb, var(--primary) 12%, #fff)' : 'var(--surface)', fontFamily: 'inherit' }}>
      <Icon name={icon} size={19} color={active ? 'var(--on-primary)' : 'var(--muted)'} />
      <div style={{ fontWeight: 700, fontSize: 14, marginTop: 7 }}>{title}</div>
      <div style={{ color: 'var(--muted)', fontSize: 11.5, fontWeight: 600 }}>{sub}</div>
    </button>
  );
}

function ShiftBuilder({ dates, multiDay, shifts, setShifts }) {
  const SM = window.SM;
  const [form, setForm] = useState(null);   // null = closed
  const blank = () => ({ name: '', date: dates[0], start: '10:00', end: '14:00', min: 2, max: 4, desc: '' });
  const fset = (k, v) => setForm(p => ({ ...p, [k]: v }));
  const add = () => {
    if (!form.name.trim() || form.end <= form.start || form.max < form.min) return;
    setShifts(s => [...s, { ...form, id: 'tmp-' + Math.random().toString(36).slice(2, 7), name: form.name.trim() }]);
    setForm(null);
  };

  return (
    <div className="fade-in">
      {shifts.length === 0 && !form && <EmptyState icon="layers" title="Noch keine Schichten" text="Lege fest, welche Helfer-Schichten es bei dieser Veranstaltung gibt." />}

      {/* grouped list */}
      {dates.map((d, di) => {
        const ds = shifts.filter(s => s.date === d);
        if (!ds.length) return null;
        return (
          <div key={d} style={{ marginBottom: 14 }}>
            {multiDay && <div className="tl-dayhead"><span className="tl-daybadge">Tag {di + 1}</span><span style={{ fontWeight: 700, fontSize: 13, color: 'var(--ink-2)' }}>{SM.fmtDate(d, 'weekday')}</span></div>}
            {ds.map(s => (
              <div key={s.id} className="sm-card" style={{ padding: 12, marginBottom: 8, display: 'flex', gap: 11, alignItems: 'center' }}>
                <div style={{ width: 4, alignSelf: 'stretch', borderRadius: 4, background: 'var(--primary)' }} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontWeight: 700, fontSize: 14.5 }}>{s.name}</div>
                  <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{s.start}–{s.end} · {s.min}–{s.max} Helfer</div>
                </div>
                <button className="pressable" onClick={() => setShifts(x => x.filter(i => i.id !== s.id))} style={{ width: 34, height: 34, borderRadius: 10, border: 'none', background: 'var(--surface-2)', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}><Icon name="trash" size={16} color="var(--crit)" /></button>
              </div>
            ))}
          </div>
        );
      })}

      {form ? (
        <div className="sm-card pad fade-in" style={{ border: '1.5px solid var(--primary)' }}>
          <div style={{ fontWeight: 800, fontFamily: 'Bricolage Grotesque', fontSize: 16, marginBottom: 12 }}>Neue Schicht</div>
          <Field label="Bezeichnung"><Input placeholder="z. B. Einlass & Kasse" value={form.name} onChange={e => fset('name', e.target.value)} /></Field>
          {multiDay && <Field label="Tag"><Select value={form.date} onChange={e => fset('date', e.target.value)}>{dates.map((d, i) => <option key={d} value={d}>Tag {i + 1} · {SM.fmtDate(d, 'weekday')}</option>)}</Select></Field>}
          <div style={{ display: 'flex', gap: 10 }}>
            <Field label="Von"><Input type="time" value={form.start} onChange={e => fset('start', e.target.value)} /></Field>
            <Field label="Bis"><Input type="time" value={form.end} onChange={e => fset('end', e.target.value)} /></Field>
          </div>
          <div style={{ display: 'flex', gap: 22, margin: '4px 0 14px' }}>
            <div><div className="sm-label">Min. Helfer</div><Stepper value={form.min} set={v => fset('min', Math.min(v, form.max))} min={1} max={20} /></div>
            <div><div className="sm-label">Max. Helfer</div><Stepper value={form.max} set={v => fset('max', Math.max(v, form.min))} min={1} max={20} /></div>
          </div>
          <Field label="Beschreibung (optional)"><Textarea placeholder="Aufgaben dieser Schicht…" value={form.desc} onChange={e => fset('desc', e.target.value)} /></Field>
          <div style={{ display: 'flex', gap: 10 }}>
            <Button variant="ghost" onClick={() => setForm(null)}>Abbrechen</Button>
            <Button icon="check" disabled={!form.name.trim() || form.end <= form.start} onClick={add}>Hinzufügen</Button>
          </div>
        </div>
      ) : (
        <Button variant="soft" icon="plus" onClick={() => setForm(blank())}>Schicht hinzufügen</Button>
      )}
    </div>
  );
}

// ── Manual booking (SA-005) ──
function ManualBooking({ id }) {
  const { state, nav, manualBook } = useApp();
  const SM = window.SM;
  const [mid, setMid] = useState(id || SM.members[0].id);
  const [date, setDate] = useState('2026-05-30');
  const [hours, setHours] = useState('2');
  const [desc, setDesc] = useState('');
  const valid = mid && date && parseFloat(hours) > 0 && desc.trim();
  return (
    <div className="fade-in">
      <div className="sm-header detail" style={{ alignItems: 'center', gap: 12 }}>
        <button onClick={nav.back} className="pressable" style={{ width: 40, height: 40, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}><Icon name="chevL" size={20} stroke={2.4} /></button>
        <div className="sm-title" style={{ fontSize: 22 }}>Stunden buchen</div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 4 }}>
        <div style={{ display: 'flex', gap: 9, background: 'var(--info)', borderRadius: 13, padding: 12, fontSize: 12.5, color: '#fff', fontWeight: 600, lineHeight: 1.4, marginBottom: 16, background: '#E8EFF8' }}>
          <Icon name="info" size={17} color="var(--info)" style={{ flexShrink: 0 }} />
          <span style={{ color: '#2F5C97' }}>Manuelle Buchungen sind keiner Veranstaltung zugeordnet und werden im Audit-Log protokolliert.</span>
        </div>
        <Field label="Mitglied"><Select value={mid} onChange={e => setMid(e.target.value)}>{SM.members.map(m => <option key={m.id} value={m.id}>{m.first} {m.last}</option>)}</Select></Field>
        <div style={{ display: 'flex', gap: 10 }}>
          <Field label="Datum"><Input type="date" value={date} onChange={e => setDate(e.target.value)} /></Field>
          <Field label="Stunden"><Input type="number" step="0.5" min="0" value={hours} onChange={e => setHours(e.target.value)} /></Field>
        </div>
        <Field label="Beschreibung" hint="Pflichtfeld – wofür wurden die Stunden geleistet?"><Textarea placeholder="z. B. Pflege der Vereinshomepage" value={desc} onChange={e => setDesc(e.target.value)} /></Field>
        <Button icon="check" disabled={!valid} onClick={() => { manualBook(mid, date, parseFloat(hours), desc.trim()); nav.back(); }}>Buchung speichern</Button>
      </div>
    </div>
  );
}

// ── Settings ──
function AdminSettings() {
  const { state, setSetting, nav, t } = useApp();
  const SM = window.SM, s = state.settings;
  return (
    <div className="fade-in">
      <div className="sm-header"><div><div className="sm-eyebrow">{s.clubName}</div><div className="sm-title">Einstellungen</div></div></div>
      <div className="sm-pad" style={{ paddingTop: 6 }}>
        <Section title="Stundenziel" />
        <div className="sm-card pad">
          <RowBetween label="Globales Jahresziel" sub="Gilt, sofern kein individuelles Ziel gesetzt ist">
            <Stepper value={s.yearGoal} set={v => setSetting('yearGoal', v)} min={1} max={80} label="h" />
          </RowBetween>
        </div>

        <Section title="Datenschutz" />
        <div className="sm-card pad">
          <div style={{ fontWeight: 700, fontSize: 14.5, marginBottom: 9 }}>Namensanzeige</div>
          <SegRadio value={s.nameMode} onChange={v => setSetting('nameMode', v)} options={[['abbrev', 'Abgekürzt'], ['full', 'Vollständig']]} />
          <div className="sm-hint">{s.nameMode === 'abbrev' ? '„Maximilian M." – Nachname gekürzt. Vorstand sieht immer den vollen Namen.' : '„Maximilian Müller" – voller Name für alle Rollen.'}</div>
          <hr className="sm-divider" style={{ margin: '14px 0' }} />
          <RowBetween label="Mitgliedersuche im Kiosk" sub="Sonst nur Eintragung per E-Mail">
            <Toggle on={s.kioskSearch} onClick={() => setSetting('kioskSearch', !s.kioskSearch)} />
          </RowBetween>
        </div>

        <Section title="Reservierungen" />
        <div className="sm-card pad">
          <RowBetween label="Reservierung gültig für" sub="Frist für E-Mail-Bestätigung">
            <Stepper value={s.reservationHours} set={v => setSetting('reservationHours', v)} min={1} max={168} label="h" />
          </RowBetween>
          <hr className="sm-divider" style={{ margin: '14px 0' }} />
          <RowBetween label="Abmeldefrist vor Beginn" sub="">
            <Stepper value={s.deregisterDeadlineH} set={v => setSetting('deregisterDeadlineH', v)} min={0} max={72} label="h" />
          </RowBetween>
        </div>

        <Section title="Abrechnung" />
        <div className="sm-card pad">
          <div style={{ fontWeight: 700, fontSize: 14.5, marginBottom: 9 }}>Abrechnungsmodus</div>
          <SegRadio value={s.billingMode} onChange={v => setSetting('billingMode', v)} options={[['manuell', 'Manuell'], ['auto', 'Automatisch']]} />
          <hr className="sm-divider" style={{ margin: '14px 0' }} />
          <div style={{ fontWeight: 700, fontSize: 14.5, marginBottom: 4 }}>Abgeltungsbeträge je Fehlstunde</div>
          <div className="sm-hint" style={{ marginBottom: 10 }}>Letzter Wert gilt für alle weiteren Fehlstunden.</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {s.feeSchedule.map((v, i) => (
              <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <span style={{ width: 78, fontSize: 13, fontWeight: 700, color: 'var(--ink-2)' }}>{i + 1}. Stunde</span>
                <div style={{ flex: 1, position: 'relative' }}>
                  <Input type="number" value={v} onChange={e => { const f = [...s.feeSchedule]; f[i] = +e.target.value; setSetting('feeSchedule', f); }} style={{ paddingRight: 28 }} />
                  <span style={{ position: 'absolute', right: 12, top: 13, color: 'var(--muted)', fontWeight: 700 }}>€</span>
                </div>
                {i === s.feeSchedule.length - 1 && <span className="sm-badge b-neutral" dot={false}>ab hier</span>}
              </div>
            ))}
          </div>
        </div>

        <Section title="Branding" />
        <div className="sm-card pad">
          <RowBetween label="Primärfarbe" sub="Über das Tweaks-Panel anpassbar">
            <div style={{ width: 30, height: 30, borderRadius: 9, background: 'var(--primary)', border: '1px solid var(--line-2)' }} />
          </RowBetween>
          <hr className="sm-divider" style={{ margin: '14px 0' }} />
          <RowBetween label="Vereinslogo" sub="PNG/SVG, min. 200×200 px">
            <div style={{ width: 30, height: 30, borderRadius: '50%', background: 'var(--ink)', color: 'var(--primary)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 800, fontSize: 12, fontFamily: 'Bricolage Grotesque' }}>SG</div>
          </RowBetween>
        </div>

        <Section title="System" />
        <div className="sm-card">
          <LinkRow icon="shield" label="Audit-Log" sub="Protokoll aller Änderungen" onClick={() => nav.push('audit')} />
          <hr className="sm-divider" />
          <LinkRow icon="download" label="Datenbank-Export" sub="CSV / Backup" />
        </div>
      </div>
    </div>
  );
}

function RowBetween({ label, sub, children }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}>
      <div style={{ flex: 1 }}><div style={{ fontWeight: 700, fontSize: 14.5 }}>{label}</div>{sub && <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{sub}</div>}</div>
      {children}
    </div>
  );
}
function SegRadio({ value, onChange, options }) {
  return (
    <div style={{ display: 'flex', gap: 6, background: 'var(--surface-2)', borderRadius: 12, padding: 4, border: '1px solid var(--line)' }}>
      {options.map(([v, l]) => (
        <button key={v} onClick={() => onChange(v)} style={{ flex: 1, border: 'none', cursor: 'pointer', padding: '9px 0', borderRadius: 9, fontWeight: 700, fontSize: 13.5, fontFamily: 'inherit', background: value === v ? 'var(--surface)' : 'transparent', color: value === v ? 'var(--ink)' : 'var(--muted)', boxShadow: value === v ? 'var(--shadow)' : 'none' }}>{l}</button>
      ))}
    </div>
  );
}

function AuditLog() {
  const { state, nav } = useApp();
  return (
    <div className="fade-in">
      <div className="sm-header detail" style={{ alignItems: 'center', gap: 12 }}>
        <button onClick={nav.back} className="pressable" style={{ width: 40, height: 40, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}><Icon name="chevL" size={20} stroke={2.4} /></button>
        <div className="sm-title" style={{ fontSize: 22 }}>Audit-Log</div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 4 }}>
        <div className="tl-rail">
          {state.audit.map((a, i) => (
            <div key={i} className="tl-shift">
              <span className="tl-dot" style={{ background: 'var(--ink-2)', top: 14 }} />
              <div className="sm-card" style={{ padding: 13 }}>
                <div style={{ fontWeight: 700, fontSize: 14, lineHeight: 1.35 }}>{a.what}</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 7, flexWrap: 'wrap' }}>
                  <Badge kind="neutral" dot={false}>{a.cat}</Badge>
                  <span style={{ fontSize: 12, color: 'var(--muted)', fontWeight: 600 }}>{a.who}</span>
                  <span style={{ fontSize: 12, color: 'var(--muted)', fontWeight: 600 }}>· {a.ts}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

Object.assign(window, { CreateEventFlow, ShiftBuilder, ChoiceCard, ManualBooking, AdminSettings, AuditLog, RowBetween, SegRadio, datesBetween });
