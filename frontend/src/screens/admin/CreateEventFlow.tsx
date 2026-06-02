import React, { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Sheet } from '@/components/ui/Sheet';
import { Stepper } from '@/components/ui/Stepper';
import { Toggle } from '@/components/ui/Toggle';
import { Field, Input, Textarea, Select } from '@/components/forms/Field';
import { useAppStore } from '@/store/app.store';
import { useCreateEvent } from '@/api/events';
import { useCreateShift } from '@/api/shifts';
import { fmtDate } from '@/screens/_demo';
import { Section, EmptyState } from '@/screens/member/MemberDashboard';
import type { EventStatus } from '@/types';

function datesBetween(start: string, end: string): string[] {
  const out: string[] = [];
  const d = new Date(start + 'T12:00:00');
  const e = new Date(end + 'T12:00:00');
  while (d <= e) {
    out.push(d.toISOString().slice(0, 10));
    d.setDate(d.getDate() + 1);
  }
  return out.length ? out : [start];
}

interface DraftShift {
  id: string;
  name: string;
  date: string;
  start: string;
  end: string;
  min: number;
  max: number;
  desc: string;
}

interface EventForm {
  name: string;
  description: string;
  location: string;
  category: string;
  status: EventStatus;
  multiDay: boolean;
  startDate: string;
  endDate: string;
}

function ChoiceCard({ active, onClick, icon, title, sub }: { active: boolean; onClick: () => void; icon: string; title: string; sub: string }) {
  return (
    <button
      onClick={onClick}
      className="pressable"
      style={{ flex: 1, textAlign: 'left', cursor: 'pointer', padding: 13, borderRadius: 14, border: active ? '1.5px solid var(--primary)' : '1.5px solid var(--line-2)', background: active ? 'color-mix(in srgb, var(--primary) 12%, #fff)' : 'var(--surface)', fontFamily: 'inherit' }}
    >
      <Icon name={icon} size={19} color={active ? 'var(--on-primary)' : 'var(--muted)'} />
      <div style={{ fontWeight: 700, fontSize: 14, marginTop: 7 }}>{title}</div>
      <div style={{ color: 'var(--muted)', fontSize: 11.5, fontWeight: 600 }}>{sub}</div>
    </button>
  );
}

function ShiftBuilder({ dates, multiDay, shifts, setShifts }: {
  dates: string[];
  multiDay: boolean;
  shifts: DraftShift[];
  setShifts: React.Dispatch<React.SetStateAction<DraftShift[]>>;
}) {
  const [form, setForm] = useState<DraftShift | null>(null);
  const blank = (): DraftShift => ({ id: '', name: '', date: dates[0], start: '10:00', end: '14:00', min: 2, max: 4, desc: '' });
  const fset = <K extends keyof DraftShift>(k: K, v: DraftShift[K]) =>
    setForm((p) => (p ? { ...p, [k]: v } : p));
  const add = () => {
    if (!form || !form.name.trim() || form.end <= form.start || form.max < form.min) return;
    setShifts((s) => [...s, { ...form, id: 'tmp-' + Math.random().toString(36).slice(2, 7), name: form.name.trim() }]);
    setForm(null);
  };

  return (
    <div className="fade-in">
      {shifts.length === 0 && !form && (
        <EmptyState icon="layers" title="Noch keine Schichten" text="Lege fest, welche Helfer-Schichten es bei dieser Veranstaltung gibt." />
      )}

      {dates.map((d, di) => {
        const ds = shifts.filter((s) => s.date === d);
        if (!ds.length) return null;
        return (
          <div key={d} style={{ marginBottom: 14 }}>
            {multiDay && (
              <div className="tl-dayhead">
                <span className="tl-daybadge">Tag {di + 1}</span>
                <span style={{ fontWeight: 700, fontSize: 13, color: 'var(--ink-2)' }}>{fmtDate(d, 'weekday')}</span>
              </div>
            )}
            {ds.map((s) => (
              <div key={s.id} className="sm-card" style={{ padding: 12, marginBottom: 8, display: 'flex', gap: 11, alignItems: 'center' }}>
                <div style={{ width: 4, alignSelf: 'stretch', borderRadius: 4, background: 'var(--primary)' }} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontWeight: 700, fontSize: 14.5 }}>{s.name}</div>
                  <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{s.start}–{s.end} · {s.min}–{s.max} Helfer</div>
                </div>
                <button
                  className="pressable"
                  onClick={() => setShifts((x) => x.filter((i) => i.id !== s.id))}
                  style={{ width: 34, height: 34, borderRadius: 10, border: 'none', background: 'var(--surface-2)', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
                >
                  <Icon name="trash" size={16} color="var(--crit)" />
                </button>
              </div>
            ))}
          </div>
        );
      })}

      {form ? (
        <div className="sm-card pad fade-in" style={{ border: '1.5px solid var(--primary)' }}>
          <div style={{ fontWeight: 800, fontFamily: 'Bricolage Grotesque', fontSize: 16, marginBottom: 12 }}>Neue Schicht</div>
          <Field label="Bezeichnung"><Input placeholder="z. B. Einlass & Kasse" value={form.name} onChange={(e) => fset('name', e.target.value)} /></Field>
          {multiDay && (
            <Field label="Tag">
              <Select value={form.date} onChange={(e) => fset('date', e.target.value)}>
                {dates.map((d, i) => <option key={d} value={d}>Tag {i + 1} · {fmtDate(d, 'weekday')}</option>)}
              </Select>
            </Field>
          )}
          <div style={{ display: 'flex', gap: 10 }}>
            <Field label="Von"><Input type="time" value={form.start} onChange={(e) => fset('start', e.target.value)} /></Field>
            <Field label="Bis"><Input type="time" value={form.end} onChange={(e) => fset('end', e.target.value)} /></Field>
          </div>
          <div style={{ display: 'flex', gap: 22, margin: '4px 0 14px' }}>
            <div><div className="sm-label">Min. Helfer</div><Stepper value={form.min} set={(v) => fset('min', Math.min(v, form.max))} min={1} max={20} /></div>
            <div><div className="sm-label">Max. Helfer</div><Stepper value={form.max} set={(v) => fset('max', Math.max(v, form.min))} min={1} max={20} /></div>
          </div>
          <Field label="Beschreibung (optional)"><Textarea placeholder="Aufgaben dieser Schicht…" value={form.desc} onChange={(e) => fset('desc', e.target.value)} /></Field>
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

export function CreateEventFlow({ onClose }: { onClose: () => void }) {
  const { showToast } = useAppStore();
  const createEvent = useCreateEvent();
  const createShift = useCreateShift();
  const [step, setStep] = useState(0);
  const [f, setF] = useState<EventForm>({
    name: '', description: '', location: '', category: 'Turnier',
    status: 'published', multiDay: false, startDate: '', endDate: '',
  });
  const [shifts, setShifts] = useState<DraftShift[]>([]);
  const set = <K extends keyof EventForm>(k: K, v: EventForm[K]) => setF((p) => ({ ...p, [k]: v }));
  const dates = f.multiDay ? datesBetween(f.startDate, f.endDate) : [f.startDate];

  const step0valid = !!(f.name.trim() && f.location.trim() && f.startDate && (!f.multiDay || f.endDate >= f.startDate));
  const labels = ['Eckdaten', 'Schichten', 'Prüfen'];

  const finish = async () => {
    try {
      const endDate = f.multiDay ? f.endDate : f.startDate;
      const newEvent = await createEvent.mutateAsync({
        name: f.name.trim(),
        description: f.description.trim() || 'Keine Beschreibung.',
        location: f.location.trim(),
        category: f.category,
        status: f.status,
        startDate: `${f.startDate}T12:00:00Z`,
        endDate: `${endDate}T12:00:00Z`,
      });
      for (const s of shifts) {
        await createShift.mutateAsync({
          eventId: newEvent.id,
          date: s.date,
          name: s.name,
          start: s.start,
          end: s.end,
          min: s.min,
          max: s.max,
          qual: s.desc || '',
        });
      }
      showToast('Veranstaltung erstellt.');
      onClose();
    } catch {
      showToast('Erstellen fehlgeschlagen.', 'crit');
    }
  };

  return (
    <Sheet variant="full" onClose={onClose}>
      <div className="cef-head" style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '54px 18px 14px', borderBottom: '1px solid var(--line)' }}>
        <button
          onClick={step === 0 ? onClose : () => setStep(step - 1)}
          className="pressable"
          style={{ width: 38, height: 38, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
        >
          <Icon name={step === 0 ? 'x' : 'chevL'} size={19} stroke={2.4} />
        </button>
        <div style={{ flex: 1 }}>
          <div className="sm-eyebrow">Schritt {step + 1} von 3</div>
          <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 20 }}>{labels[step]}</div>
        </div>
      </div>
      <div style={{ display: 'flex', gap: 6, padding: '12px 18px 0' }}>
        {labels.map((l, i) => (
          <div key={l} style={{ flex: 1, height: 4, borderRadius: 4, background: i <= step ? 'var(--primary)' : 'var(--line)', transition: 'background .2s' }} />
        ))}
      </div>

      <div className="sheet-body" style={{ flex: 1 }}>
        {step === 0 && (
          <div className="fade-in">
            <Field label="Name der Veranstaltung"><Input placeholder="z. B. Sommerturnier 2026" value={f.name} onChange={(e) => set('name', e.target.value)} /></Field>
            <Field label="Beschreibung (optional)"><Textarea placeholder="Worum geht es? Was sollen Helfer wissen?" value={f.description} onChange={(e) => set('description', e.target.value)} /></Field>
            <Field label="Ort"><Input placeholder="z. B. Sporthalle Aachen-Brand" value={f.location} onChange={(e) => set('location', e.target.value)} /></Field>
            <Field label="Kategorie">
              <Select value={f.category} onChange={(e) => set('category', e.target.value)}>
                {['Turnier', 'Vereinsleben', 'Mitgliederwerbung', 'Training', 'Sonstiges'].map((c) => <option key={c}>{c}</option>)}
              </Select>
            </Field>
            <div className="sm-card pad" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 14 }}>
              <div>
                <div style={{ fontWeight: 700, fontSize: 14.5 }}>Mehrtägige Veranstaltung</div>
                <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>Schichten je Tag planen</div>
              </div>
              <Toggle on={f.multiDay} onClick={() => set('multiDay', !f.multiDay)} />
            </div>
            <div style={{ display: 'flex', gap: 10 }}>
              <Field label={f.multiDay ? 'Von' : 'Datum'}><Input type="date" value={f.startDate} onChange={(e) => set('startDate', e.target.value)} /></Field>
              {f.multiDay && <Field label="Bis"><Input type="date" value={f.endDate} min={f.startDate} onChange={(e) => set('endDate', e.target.value)} /></Field>}
            </div>
            <Field label="Sichtbarkeit">
              <div style={{ display: 'flex', gap: 9 }}>
                <ChoiceCard active={f.status === 'draft'} onClick={() => set('status', 'draft')} icon="edit" title="Entwurf" sub="Nur intern" />
                <ChoiceCard active={f.status === 'published'} onClick={() => set('status', 'published')} icon="users" title="Veröffentlicht" sub="Für Mitglieder" />
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
                <span style={{ display: 'flex', gap: 8, alignItems: 'center' }}>
                  <Icon name="calendar" size={15} color="var(--muted)" />
                  {dates.length > 1 ? `${fmtDate(dates[0], 'daymon')} – ${fmtDate(dates[dates.length - 1], 'daymon')} (${dates.length} Tage)` : fmtDate(dates[0], 'weekday')}
                </span>
                <span style={{ display: 'flex', gap: 8, alignItems: 'center' }}><Icon name="pin" size={15} color="var(--muted)" />{f.location}</span>
              </div>
              <div style={{ marginTop: 10 }}>
                <Badge kind={f.status === 'published' ? 'ok' : 'neutral'}>{f.status === 'published' ? 'Wird veröffentlicht' : 'Als Entwurf gespeichert'}</Badge>
              </div>
            </div>
            <Section title={`${shifts.length} Schicht${shifts.length === 1 ? '' : 'en'}`} />
            {dates.map((d, di) => {
              const ds = shifts.filter((s) => s.date === d);
              if (!ds.length) return null;
              return (
                <div key={d} style={{ marginBottom: 12 }}>
                  {f.multiDay && <div style={{ fontWeight: 700, fontSize: 12.5, color: 'var(--muted)', marginBottom: 7 }}>Tag {di + 1} · {fmtDate(d, 'weekday')}</div>}
                  {ds.map((s) => (
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
          : <Button icon="check" disabled={createEvent.isPending || createShift.isPending} onClick={finish}>Veranstaltung erstellen</Button>}
      </div>
    </Sheet>
  );
}

export { ShiftBuilder, ChoiceCard, datesBetween };
