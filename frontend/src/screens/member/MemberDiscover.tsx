import { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { useAppStore } from '@/store/app.store';
import { useEvents } from '@/api/events';
import { calcOccupancy } from '@/hooks/useOccupancy';
import { LoadingState, ErrorState } from '@/components/ui/States';
import { fmtDate, catGradient } from '@/screens/_demo';
import { EmptyState } from '@/screens/member/MemberDashboard';
import type { Event } from '@/types';

function EventCard({ ev, onClick }: { ev: Event; onClick: () => void }) {
  const days = ev.days ?? [];
  const allShifts = days.flatMap((d) => d.shifts);
  const free = allShifts.reduce((a, s) => a + calcOccupancy(s).free, 0);
  const needs = allShifts.some((s) => calcOccupancy(s).needsMore);
  const d0 = days[0]?.date;
  const d1 = days[days.length - 1]?.date;
  const dateLabel = !d0
    ? 'Termin offen'
    : days.length > 1
      ? `${fmtDate(d0, 'daymon')} – ${fmtDate(d1, 'daymon')}`
      : fmtDate(d0, 'weekday');

  return (
    <div className="sm-card pressable" style={{ marginBottom: 12, overflow: 'hidden' }} onClick={onClick}>
      <div style={{ height: 74, background: catGradient(ev.category), position: 'relative', display: 'flex', alignItems: 'flex-end', padding: 12 }}>
        <span className="sm-badge" style={{ background: 'rgba(255,255,255,0.92)', color: 'var(--ink)', backdropFilter: 'blur(4px)' }}>
          <Icon name="tag" size={12} stroke={2.2} />{ev.category}
        </span>
        {days.length > 1 && (
          <span style={{ position: 'absolute', top: 12, right: 12 }} className="sm-badge b-neutral">{days.length} Tage</span>
        )}
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
            {needs && <Badge kind="crit" dot={false}>Helfer gesucht</Badge>}
            <Icon name="arrowR" size={17} stroke={2.2} />
          </span>
        </div>
      </div>
    </div>
  );
}

interface FilterDef {
  id: string;
  label: string;
}

export function MemberDiscover() {
  const { push } = useAppStore();
  const [filter, setFilter] = useState('alle');
  const [dateFrom, setDateFrom] = useState('');
  const [dateTo, setDateTo] = useState('');
  const eventsQ = useEvents();
  const pub = (eventsQ.data ?? []).filter((e) => e.status === 'veröffentlicht' || e.status === 'published');

  const filters: FilterDef[] = [
    { id: 'alle', label: 'Alle' },
    { id: 'frei', label: 'Freie Plätze' },
    { id: 'Turnier', label: 'Turnier' },
    { id: 'Vereinsleben', label: 'Vereinsleben' },
    { id: 'Mitgliederwerbung', label: 'Werbung' },
  ];

  const shown = pub.filter((ev) => {
    if (filter !== 'alle') {
      if (filter === 'frei' && !(ev.days ?? []).some((d) => d.shifts.some((s) => calcOccupancy(s).free > 0))) return false;
      if (filter !== 'frei' && ev.category !== filter) return false;
    }
    if (dateFrom || dateTo) {
      const days = ev.days ?? [];
      const inRange = days.some((d) => {
        if (dateFrom && d.date < dateFrom) return false;
        if (dateTo && d.date > dateTo) return false;
        return true;
      });
      if (!inRange) return false;
    }
    return true;
  });

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">Veranstaltungen</div>
          <div className="sm-title">Entdecken</div>
        </div>
        <button
          className="pressable"
          style={{ position: 'relative', width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }}
        >
          <Icon name="calendar" size={20} color="var(--ink)" />
        </button>
      </div>
      <div style={{ padding: '0 18px 4px' }}>
        <div className="sm-chiprow">
          {filters.map((f) => (
            <button
              key={f.id}
              className={'sm-chip' + (filter === f.id ? ' active' : '')}
              onClick={() => setFilter(f.id)}
            >
              {f.id === 'frei' && <Icon name="filter" size={14} stroke={2.2} />}{f.label}
            </button>
          ))}
        </div>
        <div style={{ display: 'flex', gap: 8, marginTop: 10 }}>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--muted)', marginBottom: 4 }}>Von</div>
            <input
              type="date"
              value={dateFrom}
              onChange={(e) => setDateFrom(e.target.value)}
              style={{ width: '100%', padding: '8px 10px', border: '1px solid var(--line)', borderRadius: 10, fontFamily: 'inherit', fontSize: 13.5, fontWeight: 600, background: 'var(--surface)', color: 'var(--ink)', boxSizing: 'border-box' }}
            />
          </div>
          <div style={{ flex: 1 }}>
            <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--muted)', marginBottom: 4 }}>Bis</div>
            <input
              type="date"
              value={dateTo}
              onChange={(e) => setDateTo(e.target.value)}
              style={{ width: '100%', padding: '8px 10px', border: '1px solid var(--line)', borderRadius: 10, fontFamily: 'inherit', fontSize: 13.5, fontWeight: 600, background: 'var(--surface)', color: 'var(--ink)', boxSizing: 'border-box' }}
            />
          </div>
        </div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 14 }}>
        {eventsQ.isLoading ? (
          <LoadingState />
        ) : eventsQ.isError ? (
          <ErrorState />
        ) : shown.length === 0 ? (
          <EmptyState icon="compass" title="Keine Treffer" text="Für diesen Filter gibt es gerade keine Veranstaltungen." />
        ) : (
          shown.map((ev) => (
            <EventCard key={ev.id} ev={ev} onClick={() => push('event', { id: ev.id })} />
          ))
        )}
      </div>
    </div>
  );
}

export { EventCard };
