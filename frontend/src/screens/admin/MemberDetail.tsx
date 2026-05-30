import React, { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Avatar } from '@/components/ui/Avatar';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Sheet } from '@/components/ui/Sheet';
import { Stepper } from '@/components/ui/Stepper';
import { useAppStore } from '@/store/app.store';
import { DEMO_STATE, fmtDate, hrs, memberMap } from '@/screens/_demo';
import { Section } from '@/screens/member/MemberDashboard';

interface HourRec {
  t: string;
  d: string;
  h: number;
  manual?: boolean;
}

function GoalSheet({ value, onClose }: { value: number; onClose: () => void }) {
  const { showToast } = useAppStore();
  const [goal, setGoal] = useState(value);
  return (
    <Sheet
      onClose={onClose}
      title="Individuelles Stundenziel"
      foot={<Button icon="check" onClick={() => { showToast(`Jahresziel auf ${goal} h gesetzt.`); onClose(); }}>Speichern</Button>}
    >
      <div className="sm-hint" style={{ marginBottom: 14 }}>Überschreibt das globale Jahresziel für dieses Mitglied.</div>
      <div style={{ display: 'flex', justifyContent: 'center', padding: '8px 0 4px' }}>
        <Stepper value={goal} set={setGoal} min={0} max={80} label="h / Jahr" />
      </div>
    </Sheet>
  );
}

export function MemberDetail({ id }: { id: string }) {
  const { back, push } = useAppStore();
  const [goalOpen, setGoalOpen] = useState(false);
  const m = memberMap[id] ?? DEMO_STATE.members[0];

  let h = 0;
  const recs: HourRec[] = [];
  DEMO_STATE.events.forEach((ev) =>
    ev.days.forEach((d) =>
      d.shifts.forEach((sh) =>
        sh.signups.forEach((s) => {
          if (s.memberId === id && s.status === 'bestätigt') {
            h += s.hours ?? 0;
            recs.push({ t: `${ev.name} · ${sh.name}`, d: d.date, h: s.hours ?? 0 });
          }
        }),
      ),
    ),
  );
  DEMO_STATE.manualBookings.forEach((b) => {
    if (b.memberId === id) {
      h += b.hours;
      recs.push({ t: b.desc, d: b.date, h: b.hours, manual: true });
    }
  });
  recs.sort((a, b) => b.d.localeCompare(a.d));
  const goal = m.goal ?? DEMO_STATE.settings.yearGoal;

  return (
    <div className="fade-in">
      <div className="sm-header detail" style={{ alignItems: 'center', gap: 12 }}>
        <button
          onClick={back}
          className="pressable"
          style={{ width: 40, height: 40, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
        >
          <Icon name="chevL" size={20} stroke={2.4} />
        </button>
        <div className="sm-title" style={{ fontSize: 22 }}>Mitglied</div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 4 }}>
        <div className="sm-card pad" style={{ display: 'flex', gap: 14, alignItems: 'center' }}>
          <Avatar memberId={id} members={memberMap} size={56} />
          <div style={{ flex: 1 }}>
            <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 19 }}>{m.first} {m.last}</div>
            <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{m.email}</div>
            <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>seit {fmtDate(m.since, 'short')}</div>
          </div>
        </div>
        <div className="sm-card pad" style={{ marginTop: 12 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
            <span style={{ fontWeight: 700, color: 'var(--ink-2)', fontSize: 14 }}>Stundenkonto {DEMO_STATE.settings.clubYear}</span>
            <span><b style={{ fontFamily: 'Bricolage Grotesque', fontSize: 19 }}>{hrs(h)}</b> <span style={{ color: 'var(--muted)', fontWeight: 700 }}>/ {goal} h</span></span>
          </div>
          <div style={{ marginTop: 10 }}>
            <div className="occ-track" style={{ height: 9 }}>
              <div className="occ-fill" style={{ width: Math.min(h / goal * 100, 100) + '%', background: h >= goal ? 'var(--ok)' : 'var(--primary)' }} />
            </div>
          </div>
          <div style={{ display: 'flex', gap: 9, marginTop: 13 }}>
            <Button variant="soft" size="sm" icon="plus" onClick={() => push('manual', { id })}>Stunden buchen</Button>
            <Button variant="soft" size="sm" icon="edit" onClick={() => setGoalOpen(true)}>Ziel anpassen</Button>
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
                  <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{fmtDate(r.d, 'short')}{r.manual && ' · manuell'}</div>
                </div>
                {r.manual && <Badge kind="primary" dot={false}>manuell</Badge>}
                <span style={{ fontWeight: 800, fontFamily: 'Bricolage Grotesque' }}>{hrs(r.h)}</span>
              </div>
            </React.Fragment>
          ))}
        </div>
      </div>

      {goalOpen && <GoalSheet value={goal} onClose={() => setGoalOpen(false)} />}
    </div>
  );
}
