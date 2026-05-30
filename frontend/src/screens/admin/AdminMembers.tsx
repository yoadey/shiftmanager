import React, { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Avatar } from '@/components/ui/Avatar';
import { Input } from '@/components/forms/Field';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import { DEMO_STATE, hrs, memberMap } from '@/screens/_demo';

export function AdminMembers() {
  const { push, showToast } = useAppStore();
  const { user } = useAuthStore();
  const uid = user?.id ?? DEMO_STATE.currentUserId;
  const [q, setQ] = useState('');

  const hours: Record<string, number> = {};
  DEMO_STATE.members.forEach((m) => { hours[m.id] = 0; });
  DEMO_STATE.events.forEach((ev) =>
    ev.days.forEach((d) =>
      d.shifts.forEach((sh) =>
        sh.signups.forEach((s) => {
          if (s.status === 'bestätigt') hours[s.memberId] = (hours[s.memberId] ?? 0) + (s.hours ?? 0);
        }),
      ),
    ),
  );
  DEMO_STATE.manualBookings.forEach((b) => { hours[b.memberId] = (hours[b.memberId] ?? 0) + b.hours; });

  const list = DEMO_STATE.members.filter((m) =>
    (m.first + ' ' + m.last).toLowerCase().includes(q.toLowerCase()),
  );

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">{DEMO_STATE.members.length} Einträge</div>
          <div className="sm-title">Mitglieder</div>
        </div>
        <button
          className="pressable"
          onClick={() => showToast('Mitglieder-Export wird erstellt …')}
          style={{ width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }}
        >
          <Icon name="download" size={20} color="var(--ink)" />
        </button>
      </div>
      <div style={{ padding: '0 18px 4px' }}>
        <div style={{ position: 'relative' }}>
          <Icon name="search" size={18} color="var(--muted)" style={{ position: 'absolute', left: 13, top: 13 }} />
          <Input placeholder="Mitglied suchen…" value={q} onChange={(e) => setQ(e.target.value)} style={{ paddingLeft: 40 }} />
        </div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 12 }}>
        <div className="sm-card" style={{ overflow: 'hidden' }}>
          {list.map((m, i) => {
            const goal = m.goal ?? DEMO_STATE.settings.yearGoal;
            const h = hours[m.id] ?? 0;
            return (
              <React.Fragment key={m.id}>
                {i > 0 && <hr className="sm-divider" style={{ marginLeft: 64 }} />}
                <div
                  className="pressable"
                  onClick={() => push('member', { id: m.id })}
                  style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px', cursor: 'pointer' }}
                >
                  <Avatar memberId={m.id} members={memberMap} size={40} />
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: 700, fontSize: 14.5 }}>
                      {m.first} {m.last}{m.id === uid && <span style={{ color: 'var(--muted)', fontWeight: 600 }}> · du</span>}
                    </div>
                    <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{hrs(h)} / {goal} h{m.goal && ' · ind. Ziel'}</div>
                  </div>
                  <div style={{ width: 46 }}>
                    <div className="occ-track">
                      <div className="occ-fill" style={{ width: Math.min(h / goal * 100, 100) + '%', background: h >= goal ? 'var(--ok)' : 'var(--primary)' }} />
                    </div>
                  </div>
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
