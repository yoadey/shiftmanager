import { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { useAppStore } from '@/store/app.store';
import { useEvents } from '@/api/events';
import { calcOccupancy } from '@/hooks/useOccupancy';
import { LoadingState, ErrorState } from '@/components/ui/States';
import { EmptyState } from '@/screens/member/MemberDashboard';
import { fmtDate } from '@/screens/_demo';
import { CreateEventFlow } from '@/screens/admin/CreateEventFlow';
import type { EventStatus } from '@/types';

const order: Record<string, number> = {
  veröffentlicht: 0,
  entwurf: 1,
  abgeschlossen: 2,
  abgesagt: 3,
};

const statusKind: Record<string, 'ok' | 'neutral' | 'full' | 'crit'> = {
  veröffentlicht: 'ok',
  entwurf: 'neutral',
  abgeschlossen: 'full',
  abgesagt: 'crit',
};

function statusLabel(s: EventStatus): string {
  return s.charAt(0).toUpperCase() + s.slice(1);
}

export function AdminEvents() {
  const { push } = useAppStore();
  const [createOpen, setCreateOpen] = useState(false);
  const eventsQ = useEvents();

  const evs = [...(eventsQ.data ?? [])].sort(
    (a, b) => (order[a.status] - order[b.status]) || (a.days?.[0]?.date ?? '').localeCompare(b.days?.[0]?.date ?? ''),
  );

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">Verwaltung</div>
          <div className="sm-title">Veranstaltungen</div>
        </div>
        <button
          className="pressable"
          onClick={() => setCreateOpen(true)}
          style={{ width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }}
        >
          <Icon name="plus" size={20} color="var(--ink)" />
        </button>
      </div>
      <div className="sm-pad" style={{ paddingTop: 8 }}>
        {eventsQ.isLoading && <LoadingState />}
        {eventsQ.isError && <ErrorState />}
        {!eventsQ.isLoading && !eventsQ.isError && evs.length === 0 && (
          <EmptyState icon="calendar" title="Keine Veranstaltungen" text="Lege über das Plus-Symbol die erste Veranstaltung an." />
        )}
        {!eventsQ.isLoading && !eventsQ.isError && evs.map((ev) => {
          const days = ev.days ?? [];
          const shifts = days.flatMap((d) => d.shifts);
          const filled = shifts.reduce((a, s) => a + calcOccupancy(s).count, 0);
          const cap = shifts.reduce((a, s) => a + s.max, 0);
          return (
            <div key={ev.id} className="sm-card pressable" style={{ padding: 15, marginBottom: 11 }} onClick={() => push('event', { id: ev.id })}>
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: 10, alignItems: 'flex-start' }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 16.5, lineHeight: 1.15 }}>{ev.name}</div>
                  <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600, marginTop: 3 }}>
                    {days[0] ? fmtDate(days[0].date, 'daymon') : 'Termin offen'}{days.length > 1 ? ` – ${fmtDate(days[days.length - 1].date, 'daymon')}` : ''} · {ev.location}
                  </div>
                </div>
                <Badge kind={statusKind[ev.status] ?? 'neutral'} dot={ev.status === 'veröffentlicht'}>{statusLabel(ev.status)}</Badge>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginTop: 12 }}>
                <div style={{ flex: 1 }}>
                  <div className="occ-track">
                    <div className="occ-fill" style={{ width: (cap ? filled / cap * 100 : 0) + '%', background: 'var(--primary)' }} />
                  </div>
                </div>
                <span style={{ fontSize: 12, fontWeight: 800, color: 'var(--ink-2)' }}>{filled}/{cap} besetzt · {shifts.length} Schichten</span>
              </div>
            </div>
          );
        })}
      </div>

      {createOpen && <CreateEventFlow onClose={() => setCreateOpen(false)} />}
    </div>
  );
}
