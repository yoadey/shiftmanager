import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { useEventPage } from '@/api/events';
import { calcOccupancy } from '@/hooks/useOccupancy';
import { LoadingState, ErrorState } from '@/components/ui/States';
import { fmtDate, catGradient } from '@/screens/_demo';
import { EmptyState } from '@/screens/member/MemberDashboard';
import { routes } from '@/routes';
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

export const DISCOVER_PAGE_SIZE = 10;

/** Today as YYYY-MM-DD in the viewer's local time zone. */
function todayLocal(): string {
  const d = new Date();
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

export function MemberDiscover() {
  const navigate = useNavigate();
  const [filter, setFilter] = useState('alle');
  // Every visit starts from today, so past events stay hidden unless the
  // member clears "Von" (D-005).
  const [dateFrom, setDateFrom] = useState(todayLocal);
  const [dateTo, setDateTo] = useState('');
  const [page, setPage] = useState(0);
  const listTop = useRef<HTMLDivElement>(null);

  const filters: FilterDef[] = [
    { id: 'alle', label: 'Alle' },
    { id: 'frei', label: 'Freie Plätze' },
    { id: 'Turnier', label: 'Turnier' },
    { id: 'Vereinsleben', label: 'Vereinsleben' },
    { id: 'Mitgliederwerbung', label: 'Werbung' },
  ];

  const eventsQ = useEventPage({
    page,
    pageSize: DISCOVER_PAGE_SIZE,
    from: dateFrom,
    to: dateTo,
    category: filter !== 'alle' && filter !== 'frei' ? filter : undefined,
    freeSlots: filter === 'frei',
  });
  const shown = eventsQ.data?.items ?? [];
  const total = eventsQ.data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / DISCOVER_PAGE_SIZE));

  // Any filter change starts over at the first page.
  const changeFilter = (set: (v: string) => void) => (v: string) => {
    set(v);
    setPage(0);
  };
  const goToPage = (p: number) => {
    setPage(p);
    listTop.current?.scrollIntoView?.({ block: 'start' });
  };

  // A page beyond the end (events were unpublished meanwhile) snaps back to
  // the last existing page once the real page has loaded.
  useEffect(() => {
    if (eventsQ.data && !eventsQ.isPlaceholderData && page > 0 && page >= pageCount) setPage(pageCount - 1);
  }, [eventsQ.data, eventsQ.isPlaceholderData, page, pageCount]);

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
              onClick={() => changeFilter(setFilter)(f.id)}
            >
              {f.id === 'frei' && <Icon name="filter" size={14} stroke={2.2} />}{f.label}
            </button>
          ))}
        </div>
        <div style={{ display: 'flex', gap: 8, marginTop: 10 }}>
          <label style={{ flex: 1 }}>
            <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--muted)', marginBottom: 4 }}>Von</div>
            <input
              type="date"
              value={dateFrom}
              onChange={(e) => changeFilter(setDateFrom)(e.target.value)}
              style={{ width: '100%', padding: '8px 10px', border: '1px solid var(--line)', borderRadius: 10, fontFamily: 'inherit', fontSize: 13.5, fontWeight: 600, background: 'var(--surface)', color: 'var(--ink)', boxSizing: 'border-box' }}
            />
          </label>
          <label style={{ flex: 1 }}>
            <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--muted)', marginBottom: 4 }}>Bis</div>
            <input
              type="date"
              value={dateTo}
              onChange={(e) => changeFilter(setDateTo)(e.target.value)}
              style={{ width: '100%', padding: '8px 10px', border: '1px solid var(--line)', borderRadius: 10, fontFamily: 'inherit', fontSize: 13.5, fontWeight: 600, background: 'var(--surface)', color: 'var(--ink)', boxSizing: 'border-box' }}
            />
          </label>
        </div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 14 }} ref={listTop}>
        {eventsQ.isLoading ? (
          <LoadingState />
        ) : eventsQ.isError ? (
          <ErrorState />
        ) : shown.length === 0 ? (
          <EmptyState icon="compass" title="Keine Treffer" text="Für diesen Filter gibt es gerade keine Veranstaltungen." />
        ) : (
          shown.map((ev) => (
            <EventCard key={ev.id} ev={ev} onClick={() => navigate(routes.event(ev.id))} />
          ))
        )}
        {!eventsQ.isError && pageCount > 1 && (
          <nav aria-label="Seiten" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 10, marginTop: 4 }}>
            <Button size="sm" variant="ghost" disabled={page === 0} onClick={() => goToPage(page - 1)}>
              <Icon name="chevL" size={14} stroke={2.4} /> Zurück
            </Button>
            <span style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 700, textAlign: 'center' }}>
              <span style={{ display: 'block' }}>Seite {page + 1} von {pageCount}</span>
              <span style={{ display: 'block', fontWeight: 600, fontSize: 11.5 }}>
                {page * DISCOVER_PAGE_SIZE + 1}–{Math.min((page + 1) * DISCOVER_PAGE_SIZE, total)} von {total}
              </span>
            </span>
            <Button size="sm" variant="ghost" disabled={page >= pageCount - 1} onClick={() => goToPage(page + 1)}>
              Weiter <Icon name="chevR" size={14} stroke={2.4} />
            </Button>
          </nav>
        )}
      </div>
    </div>
  );
}

export { EventCard };
