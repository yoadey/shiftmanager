import { useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Sheet } from '@/components/ui/Sheet';
import { Field, Input, Select } from '@/components/forms/Field';
import { useAppStore } from '@/store/app.store';
import { useEvents, useCopyEvent, useGenerateEventRecurrence } from '@/api/events';
import { calcOccupancy } from '@/hooks/useOccupancy';
import { LoadingState, ErrorState } from '@/components/ui/States';
import { EmptyState } from '@/screens/member/MemberDashboard';
import { fmtDate } from '@/screens/_demo';
import { routes } from '@/routes';
import { useSmartBack } from '@/hooks/useSmartBack';
import type { Event, EventStatus } from '@/types';

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

// Turns an event into a recurring series (V-007): weekly/monthly copies
// (including shifts) up to a chosen end date.
function RecurrenceDialog({ ev, onClose }: { ev: Event; onClose: () => void }) {
  const { showToast } = useAppStore();
  const [frequency, setFrequency] = useState<'weekly' | 'monthly'>('weekly');
  const [until, setUntil] = useState('');
  const generate = useGenerateEventRecurrence(ev.id);
  const startDate = ev.days[0]?.date;

  const submit = () => {
    if (!until) return;
    generate.mutate(
      { frequency, until: `${until}T23:59:59Z` },
      {
        onSuccess: (created) => {
          showToast(`${Math.max(created.length - 1, 0)} weitere Termine angelegt.`);
          onClose();
        },
        onError: () => showToast('Serientermin konnte nicht angelegt werden.', 'crit'),
      },
    );
  };

  return (
    <Sheet variant="dialog" onClose={onClose}>
      <div style={{ textAlign: 'center', padding: '6px 4px 4px' }}>
        <div style={{ width: 54, height: 54, borderRadius: '50%', background: 'var(--surface-2, var(--line))', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
          <Icon name="repeat" size={26} stroke={2.4} color="var(--primary)" />
        </div>
        <h3 style={{ fontSize: 19, fontWeight: 800 }}>Als Serie anlegen</h3>
        <p style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, margin: '8px 0 18px', lineHeight: 1.45 }}>
          Erzeugt Kopien von „{ev.name}" inklusive aller Schichten, wöchentlich oder monatlich, bis zum gewählten Datum.
        </p>
        <div style={{ textAlign: 'left', display: 'flex', flexDirection: 'column', gap: 10, marginBottom: 18 }}>
          <Field label="Wiederholung">
            <Select value={frequency} onChange={(e) => setFrequency(e.target.value as 'weekly' | 'monthly')}>
              <option value="weekly">Wöchentlich</option>
              <option value="monthly">Monatlich</option>
            </Select>
          </Field>
          <Field label="Bis einschließlich">
            <Input type="date" value={until} min={startDate} onChange={(e) => setUntil(e.target.value)} />
          </Field>
        </div>
        <div style={{ display: 'flex', gap: 10 }}>
          <Button variant="ghost" onClick={onClose}>Abbrechen</Button>
          <Button onClick={submit} disabled={!until || generate.isPending}>Serie anlegen</Button>
        </div>
      </div>
    </Sheet>
  );
}

export function AdminEvents() {
  const { showToast } = useAppStore();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const serieId = searchParams.get('serie');
  const eventsQ = useEvents();
  const copyEvent = useCopyEvent();
  const recurrenceFor = serieId ? (eventsQ.data ?? []).find((e) => e.id === serieId) ?? null : null;
  const closeRecurrence = useSmartBack(routes.events);

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
          onClick={() => navigate(routes.eventNeu)}
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
            <div key={ev.id} className="sm-card pressable" style={{ padding: 15, marginBottom: 11 }} onClick={() => navigate(routes.event(ev.id))}>
              <div style={{ display: 'flex', justifyContent: 'space-between', gap: 10, alignItems: 'flex-start' }}>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 16.5, lineHeight: 1.15 }}>{ev.name}</div>
                  <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600, marginTop: 3 }}>
                    {days[0] ? fmtDate(days[0].date, 'daymon') : 'Termin offen'}{days.length > 1 ? ` – ${fmtDate(days[days.length - 1].date, 'daymon')}` : ''} · {ev.location}
                  </div>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  {!ev.recurrenceFrequency && (
                    <button
                      className="pressable"
                      onClick={(e) => {
                        e.stopPropagation();
                        navigate(routes.eventSerie(ev.id));
                      }}
                      style={{ width: 32, height: 32, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
                      title="Als Serie anlegen"
                    >
                      <Icon name="repeat" size={15} color="var(--ink)" />
                    </button>
                  )}
                  <button
                    className="pressable"
                    onClick={(e) => {
                      e.stopPropagation();
                      copyEvent.mutate(ev.id, {
                        onSuccess: () => showToast('Veranstaltung wurde kopiert'),
                        onError: () => showToast('Kopieren fehlgeschlagen', 'crit'),
                      });
                    }}
                    disabled={copyEvent.isPending}
                    style={{ width: 32, height: 32, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
                    title="Veranstaltung kopieren"
                  >
                    <Icon name="layers" size={15} color="var(--ink)" />
                  </button>
                  <Badge kind={statusKind[ev.status] ?? 'neutral'} dot={ev.status === 'veröffentlicht'}>{statusLabel(ev.status)}</Badge>
                </div>
              </div>
              {ev.recurrenceFrequency && (
                <div style={{ display: 'flex', alignItems: 'center', gap: 5, marginTop: 8, color: 'var(--muted)', fontSize: 11.5, fontWeight: 700 }}>
                  <Icon name="repeat" size={13} color="var(--muted)" />
                  {ev.recurrenceFrequency === 'weekly' ? 'Wöchentliche Serie' : 'Monatliche Serie'}
                </div>
              )}
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

      {recurrenceFor && <RecurrenceDialog ev={recurrenceFor} onClose={closeRecurrence} />}
    </div>
  );
}
