import { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { useAppStore } from '@/store/app.store';
import { useEvents } from '@/api/events';
import { useMembers } from '@/api/members';
import { useSettings, useBranding } from '@/api/settings';
import { calcOccupancy } from '@/hooks/useOccupancy';
import { LoadingState, ErrorState } from '@/components/ui/States';
import { DEMO_STATE, fmtDate, hrs } from '@/screens/_demo';
import { Section, EmptyState } from '@/screens/member/MemberDashboard';
import { CreateEventFlow } from '@/screens/admin/CreateEventFlow';
import type { Event, Shift, ShiftDay, ShiftOccupancy } from '@/types';

function StatCard({ icon, label, val, sub, accent }: { icon: string; label: string; val: string | number; sub: string; accent?: boolean }) {
  return (
    <div className="sm-card" style={{ padding: 14, background: accent ? 'var(--ink)' : 'var(--surface)' }}>
      <Icon name={icon} size={20} color={accent ? 'var(--primary)' : 'var(--muted)'} />
      <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 27, marginTop: 8, color: accent ? '#fff' : 'var(--ink)', lineHeight: 1 }}>{val}</div>
      <div style={{ fontWeight: 700, fontSize: 12.5, marginTop: 4, color: accent ? 'rgba(255,255,255,0.92)' : 'var(--ink-2)' }}>{label}</div>
      <div style={{ fontSize: 11, fontWeight: 600, color: accent ? 'rgba(255,255,255,0.55)' : 'var(--muted)' }}>{sub}</div>
    </div>
  );
}

interface Understaffed {
  ev: Event;
  d: ShiftDay;
  sh: Shift;
  o: ShiftOccupancy;
}

export function AdminDashboard() {
  const { push, setRole } = useAppStore();
  const [createOpen, setCreateOpen] = useState(false);
  const today = new Date().toISOString().slice(0, 10);

  const eventsQ = useEvents();
  const membersQ = useMembers();
  const { data: settings } = useSettings();
  const { data: branding } = useBranding();

  const events = eventsQ.data ?? [];
  const clubName = branding?.clubName ?? settings?.clubName ?? '';
  const clubYear = settings?.clubYear ?? new Date().getFullYear().toString();

  let totalConfirmed = 0;
  let openSlots = 0;
  const understaffed: Understaffed[] = [];

  events.forEach((ev) => {
    const published = ev.status === 'veröffentlicht' || ev.status === 'published';
    const draftOrCancelled = ev.status === 'abgesagt' || ev.status === 'entwurf'
      || ev.status === 'cancelled' || ev.status === 'draft';
    if (draftOrCancelled) return;
    (ev.days ?? []).forEach((d) =>
      d.shifts.forEach((sh) => {
        sh.signups.forEach((s) => { if (s.status === 'bestätigt') totalConfirmed += s.hours ?? 0; });
        const o = calcOccupancy(sh);
        openSlots += o.free;
        if (o.needsMore && published && d.date >= today) understaffed.push({ ev, d, sh, o });
      }),
    );
  });

  const activeEvents = events.filter((e) => e.status === 'veröffentlicht' || e.status === 'published').length;
  const memberCount = membersQ.data?.length ?? 0;

  // TODO(stats): no system-stats endpoint yet — `totalConfirmed` is summed from the events
  // payload (only counts hours visible on shifts, not manual bookings). Replace with a
  // dedicated /stats hook once the backend exposes it. Demo number kept as a last-resort fallback.
  const totalConfirmedDisplay = totalConfirmed || DEMO_STATE.events.reduce(
    (sum, ev) => sum + ev.days.flatMap((d) => d.shifts).flatMap((s) => s.signups)
      .filter((s) => s.status === 'bestätigt').reduce((a, s) => a + (s.hours ?? 0), 0),
    0,
  );

  if (eventsQ.isLoading || membersQ.isLoading) return <LoadingState />;
  if (eventsQ.isError || membersQ.isError) return <ErrorState />;

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">{clubName}</div>
          <div className="sm-title">Übersicht</div>
        </div>
        <button
          className="pressable"
          onClick={() => setRole('mitglied')}
          style={{ width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }}
        >
          <Icon name="user" size={20} color="var(--ink)" />
        </button>
      </div>
      <div className="sm-pad">
        <div className="sm-stat-grid">
          <StatCard icon="hours" label="Geleistete Stunden" val={hrs(totalConfirmedDisplay)} sub={`Vereinsjahr ${clubYear}`} />
          <StatCard icon="users" label="Mitglieder" val={memberCount} sub="aktiv" />
          <StatCard icon="calendar" label="Veröffentlicht" val={activeEvents} sub="Veranstaltungen" />
          <StatCard icon="layers" label="Offene Plätze" val={openSlots} sub="über alle Schichten" accent />
        </div>

        <Button icon="plus" onClick={() => setCreateOpen(true)} style={{ marginTop: 14 }}>Neue Veranstaltung</Button>
        <div style={{ height: 14 }} />
        <Button variant="soft" icon="plus" onClick={() => push('manual')}>Stunden manuell buchen</Button>

        <Section title="Unterbesetzte Schichten" />
        {understaffed.length === 0
          ? <EmptyState icon="check" title="Alles im grünen Bereich" text="Keine bevorstehende Schicht unter dem Mindesthelfer-Wert." />
          : understaffed.map(({ ev, d, sh, o }) => (
            <div
              key={sh.id}
              className="sm-card pressable"
              style={{ padding: 13, marginBottom: 10, display: 'flex', gap: 12, alignItems: 'center', borderLeft: '3px solid var(--crit)' }}
              onClick={() => push('event', { id: ev.id })}
            >
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontWeight: 700, fontSize: 15 }}>{sh.name}</div>
                <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{ev.name} · {fmtDate(d.date, 'daymon')}</div>
              </div>
              <Badge kind={o.key === 'crit' ? 'crit' : 'warn'}>{o.count}/{sh.min} min.</Badge>
            </div>
          ))}
      </div>

      {createOpen && <CreateEventFlow onClose={() => setCreateOpen(false)} />}
    </div>
  );
}

export { StatCard };
