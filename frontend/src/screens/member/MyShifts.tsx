import { useState } from 'react';
import { Badge } from '@/components/ui/Badge';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import { useEvents } from '@/api/events';
import { LoadingState, ErrorState } from '@/components/ui/States';
import { hrs } from '@/screens/_demo';
import {
  collectMy,
  MyShiftCard,
  DateChip,
  EmptyState,
  type ShiftRec,
} from '@/screens/member/MemberDashboard';

function PastShiftCard({ r }: { r: ShiftRec }) {
  const no = r.su.status === 'nichterschienen';
  return (
    <div
      className="sm-card"
      style={{ padding: 14, marginBottom: 10, display: 'flex', gap: 13, alignItems: 'center', opacity: no ? 0.85 : 1 }}
    >
      <DateChip date={r.day.date} />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontWeight: 700, fontSize: 15.5 }}>{r.sh.name}</div>
        <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{r.ev.name}</div>
        <div style={{ marginTop: 7 }}>
          {no
            ? <Badge kind="crit">Nicht erschienen · 0 h</Badge>
            : <Badge kind="ok">Bestätigt · {hrs(r.su.hours ?? 0)}</Badge>}
        </div>
      </div>
    </div>
  );
}

export function MyShifts() {
  const { push } = useAppStore();
  const { user } = useAuthStore();
  const uid = user?.id ?? '';
  const [tab, setTab] = useState<'up' | 'past'>('up');
  const eventsQ = useEvents();
  const { up, past } = collectMy(eventsQ.data ?? [], uid);
  const list = tab === 'up' ? up : past;

  const tabs: [('up' | 'past'), string, number][] = [
    ['up', 'Bevorstehend', up.length],
    ['past', 'Vergangen', past.length],
  ];

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">Mein Engagement</div>
          <div className="sm-title">Meine Schichten</div>
        </div>
      </div>
      <div style={{ padding: '0 18px' }}>
        <div style={{ display: 'flex', gap: 4, background: 'var(--surface-2)', borderRadius: 13, padding: 4, border: '1px solid var(--line)' }}>
          {tabs.map(([k, l, n]) => (
            <button
              key={k}
              onClick={() => setTab(k)}
              style={{ flex: 1, border: 'none', cursor: 'pointer', padding: '9px 0', borderRadius: 10, fontWeight: 700, fontSize: 14, fontFamily: 'inherit', background: tab === k ? 'var(--surface)' : 'transparent', color: tab === k ? 'var(--ink)' : 'var(--muted)', boxShadow: tab === k ? 'var(--shadow)' : 'none' }}
            >
              {l} · {n}
            </button>
          ))}
        </div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 14 }}>
        {eventsQ.isLoading && <LoadingState />}
        {eventsQ.isError && <ErrorState />}
        {!eventsQ.isLoading && !eventsQ.isError && list.length === 0 && (
          <EmptyState
            icon="calendar"
            title={tab === 'up' ? 'Keine kommenden Schichten' : 'Noch keine Historie'}
            text={tab === 'up' ? 'Finde freie Schichten im Tab „Entdecken“.' : 'Abgeschlossene Schichten erscheinen hier.'}
          />
        )}
        {!eventsQ.isLoading && !eventsQ.isError && list.map((r) =>
          tab === 'up'
            ? <MyShiftCard key={r.sh.id} r={r} onClick={() => push('event', { id: r.ev.id })} />
            : <PastShiftCard key={r.sh.id} r={r} />,
        )}
      </div>
    </div>
  );
}

export { PastShiftCard };
