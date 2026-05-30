import React from 'react';
import { Icon } from '@/components/ui/Icon';
import { Avatar } from '@/components/ui/Avatar';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { HourBar } from '@/components/ui/HourBar';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import { calcOccupancy } from '@/hooks/useOccupancy';
import type { Event, Shift, ShiftDay, Signup } from '@/types';

// ── Demo data (until API is connected) ───────────────────────────────────────
import { DEMO_STATE } from '@/screens/_demo';

// ── Date helpers ──────────────────────────────────────────────────────────────
function fmtDate(iso: string, style?: string): string {
  const d = new Date(iso + 'T12:00:00');
  if (style === 'weekday')
    return new Intl.DateTimeFormat('de-DE', { weekday: 'short', day: 'numeric', month: 'long' }).format(d);
  if (style === 'daymon')
    return new Intl.DateTimeFormat('de-DE', { day: 'numeric', month: 'short' }).format(d);
  return new Intl.DateTimeFormat('de-DE', { day: 'numeric', month: 'long', year: 'numeric' }).format(d);
}
function durH(start: string, end: string): number {
  const [sh, sm] = start.split(':').map(Number);
  const [eh, em] = end.split(':').map(Number);
  return ((eh * 60 + em) - (sh * 60 + sm)) / 60;
}
function hrs(n: number): string {
  return (Number.isInteger(n) ? n : n.toLocaleString('de-DE', { minimumFractionDigits: 1 })) + ' h';
}

// ── Collect upcoming/past shifts for current user ─────────────────────────────
interface ShiftRec {
  ev: Event;
  day: ShiftDay;
  sh: Shift;
  su: Signup;
  dur: number;
}
function collectMy(events: Event[], uid: string): { up: ShiftRec[]; past: ShiftRec[] } {
  const today = new Date().toISOString().slice(0, 10);
  const up: ShiftRec[] = [], past: ShiftRec[] = [];
  events.forEach((ev) =>
    ev.days.forEach((day) =>
      day.shifts.forEach((sh) => {
        const su = sh.signups.find(
          (s) => s.memberId === uid && (s.status === 'angemeldet' || s.status === 'reserviert'),
        );
        if (!su) return;
        const rec: ShiftRec = { ev, day, sh, su, dur: durH(sh.start, sh.end) };
        if (day.date >= today && su.status !== 'bestätigt' && su.status !== 'nichterschienen') {
          up.push(rec);
        } else {
          past.push(rec);
        }
      }),
    ),
  );
  up.sort((a, b) => (a.day.date + a.sh.start).localeCompare(b.day.date + b.sh.start));
  past.sort((a, b) => b.day.date.localeCompare(a.day.date));
  return { up, past };
}

// ── Sub-components ────────────────────────────────────────────────────────────

function DateChip({ date }: { date: string }) {
  const d = new Date(date + 'T12:00:00');
  const wd = new Intl.DateTimeFormat('de-DE', { weekday: 'short' }).format(d).replace('.', '');
  const day = d.getDate();
  const mon = new Intl.DateTimeFormat('de-DE', { month: 'short' }).format(d).replace('.', '');
  return (
    <div
      style={{
        width: 52, height: 56, borderRadius: 14,
        background: 'var(--surface-2)', border: '1px solid var(--line)',
        display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center',
        flexShrink: 0,
      }}
    >
      <span style={{ fontSize: 10.5, fontWeight: 800, color: 'var(--muted)', textTransform: 'uppercase' }}>{wd}</span>
      <span style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 21, lineHeight: 1 }}>{day}</span>
      <span style={{ fontSize: 10, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase' }}>{mon}</span>
    </div>
  );
}

function MyShiftCard({ r, onClick }: { r: ShiftRec; onClick: () => void }) {
  const statusBadge =
    r.su.status === 'reserviert'
      ? <Badge kind="warn">Reserviert</Badge>
      : <Badge kind="info">Angemeldet</Badge>;
  return (
    <div
      className="sm-card pressable"
      style={{ padding: 14, marginBottom: 10, display: 'flex', gap: 13, alignItems: 'center' }}
      onClick={onClick}
    >
      <DateChip date={r.day.date} />
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontWeight: 700, fontSize: 15.5, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{r.sh.name}</div>
        <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{r.ev.name}</div>
        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 7 }}>
          <span style={{ display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 12.5, fontWeight: 700, color: 'var(--ink-2)' }}>
            <Icon name="clock" size={14} stroke={2.2} />{r.sh.start}–{r.sh.end}
          </span>
          {statusBadge}
        </div>
      </div>
    </div>
  );
}

function Legend({ swatch, label, val }: { swatch: string; label: string; val: string }) {
  return (
    <div style={{ flex: 1 }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 6 }}>
        <span className="sm-track" style={{ width: 18, height: 10, border: 'none', display: 'inline-block' }}>
          <span className={swatch} style={{ display: 'block', width: '100%', height: '100%', borderRadius: 3 }} />
        </span>
        <span style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)' }}>{label}</span>
      </div>
      <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 18, marginTop: 3 }}>{val}</div>
    </div>
  );
}

function EmptyState({ icon, title, text }: { icon: string; title: string; text: string }) {
  return (
    <div style={{ textAlign: 'center', padding: '34px 20px' }}>
      <div style={{ width: 58, height: 58, borderRadius: '50%', background: 'var(--surface-2)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
        <Icon name={icon} size={26} color="var(--muted)" />
      </div>
      <div style={{ fontWeight: 700, fontSize: 16 }}>{title}</div>
      <div style={{ color: 'var(--muted)', fontSize: 13.5, fontWeight: 600, marginTop: 4, maxWidth: 250, marginInline: 'auto' }}>{text}</div>
    </div>
  );
}

function TextLink({ onClick, children }: { onClick: () => void; children: React.ReactNode }) {
  return (
    <span className="pressable" onClick={onClick} style={{ fontWeight: 700, fontSize: 13.5, color: 'var(--ink-2)', cursor: 'pointer', display: 'inline-flex', alignItems: 'center', gap: 3 }}>
      {children}<Icon name="chevR" size={14} stroke={2.4} />
    </span>
  );
}

function Section({ title, action }: { title: string; action?: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', margin: '22px 0 11px' }}>
      <h3 style={{ fontSize: 17.5, fontWeight: 700 }}>{title}</h3>
      {action}
    </div>
  );
}

// ── Main component ─────────────────────────────────────────────────────────────

export function MemberDashboard() {
  const { go, push } = useAppStore();
  const { user } = useAuthStore();

  const { events, manualBookings, settings } = DEMO_STATE;
  const uid = user?.id ?? 'm-jonas';
  const me = DEMO_STATE.members.find((m) => m.id === uid) ?? DEMO_STATE.members[0];
  const memberMap = Object.fromEntries(DEMO_STATE.members.map((m) => [m.id, m]));

  const { up } = collectMy(events, uid);

  let confirmed = 0;
  events.forEach((ev) =>
    ev.days.forEach((d) =>
      d.shifts.forEach((sh) =>
        sh.signups.forEach((s) => {
          if (s.memberId === uid && s.status === 'bestätigt') confirmed += s.hours ?? 0;
        }),
      ),
    ),
  );
  manualBookings.forEach((b) => { if (b.memberId === uid) confirmed += b.hours; });

  const reservedExtra = up.reduce((a, r) => a + r.dur, 0);
  const incl = confirmed + reservedExtra;
  const goal = me.goal ?? settings.yearGoal;
  const remaining = Math.max(goal - incl, 0);

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">{settings.clubName}</div>
          <div className="sm-title">Hallo, {me.first}</div>
        </div>
        <div style={{ display: 'flex', gap: 9 }}>
          <button
            onClick={() => go('profil')}
            className="pressable"
            style={{ position: 'relative', width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }}
          >
            <Icon name="bell" size={20} color="var(--ink)" />
            <span style={{ position: 'absolute', top: 9, right: 10, width: 8, height: 8, borderRadius: '50%', background: 'var(--crit)', border: '1.5px solid var(--surface)' }} />
          </button>
          <Avatar memberId={uid} members={memberMap} size={42} />
        </div>
      </div>

      <div className="sm-pad">
        {/* Hour account card */}
        <div className="sm-card pad" style={{ padding: 18 }}>
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 14 }}>
            <span style={{ fontWeight: 700, fontSize: 14.5, color: 'var(--ink-2)' }}>Stundenkonto</span>
            <Badge kind="primary" dot={false}>Vereinsjahr {settings.clubYear}</Badge>
          </div>
          <div style={{ display: 'flex', alignItems: 'baseline', gap: 8 }}>
            <span style={{ fontFamily: 'Bricolage Grotesque', fontSize: 46, fontWeight: 800, lineHeight: 1, letterSpacing: '-0.03em' }}>
              {hrs(confirmed).replace(' h', '')}
            </span>
            <span style={{ color: 'var(--muted)', fontWeight: 700, fontSize: 18 }}>/ {goal} h</span>
          </div>
          <div style={{ color: 'var(--muted)', fontSize: 13, fontWeight: 600, marginTop: 2, marginBottom: 14 }}>bestätigte Stunden</div>
          <HourBar confirmed={confirmed} reserved={incl} goal={goal} />
          <div style={{ display: 'flex', gap: 18, marginTop: 14 }}>
            <Legend swatch="seg-c" label="Bestätigt" val={hrs(confirmed)} />
            <Legend swatch="seg-r" label="Inkl. Reservierungen" val={hrs(incl)} />
          </div>
          <div style={{ marginTop: 14, padding: '11px 13px', background: 'var(--surface-2)', borderRadius: 13, fontSize: 13.5, fontWeight: 600, color: 'var(--ink-2)', display: 'flex', alignItems: 'center', gap: 8 }}>
            <Icon name={remaining > 0 ? 'spark' : 'check'} size={17} color="var(--warn)" />
            {remaining > 0
              ? <span>Noch <b style={{ color: 'var(--ink)' }}>{hrs(remaining)}</b> bis zum Jahresziel – inkl. deiner Reservierungen.</span>
              : <span>Stark! Du hast dein Jahresziel bereits erreicht.</span>}
          </div>
        </div>

        <div style={{ marginTop: 14 }}>
          <Button icon="search" onClick={() => go('entdecken')}>Freie Schichten finden</Button>
        </div>

        <Section
          title="Deine nächsten Schichten"
          action={up.length > 0 ? <TextLink onClick={() => go('schichten')}>Alle</TextLink> : undefined}
        />
        {up.length === 0
          ? <EmptyState icon="calendar" title="Noch nichts geplant" text="Melde dich für eine Schicht an – sie erscheint dann hier." />
          : up.slice(0, 3).map((r) => (
            <MyShiftCard key={r.sh.id} r={r} onClick={() => push('event', { id: r.ev.id })} />
          ))}
      </div>
    </div>
  );
}

export { collectMy, DateChip, MyShiftCard, EmptyState, TextLink, Section, fmtDate, durH, hrs };
