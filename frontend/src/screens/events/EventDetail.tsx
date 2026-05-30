import { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Avatar } from '@/components/ui/Avatar';
import { Badge, OccBadge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { OccFill } from '@/components/ui/OccFill';
import { Sheet } from '@/components/ui/Sheet';
import { Field, Input, Textarea } from '@/components/forms/Field';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import { useRegisterShift, useDeregisterShift } from '@/api/shifts';
import { calcOccupancy } from '@/hooks/useOccupancy';
import { useNameFormat } from '@/hooks/useNameFormat';
import { DEMO_STATE, fmtDate, hrs, durH, catGradient, memberMap } from '@/screens/_demo';
import type { Event, Shift } from '@/types';

const dotColor: Record<string, string> = {
  ok: 'var(--ok)',
  warn: 'var(--warn)',
  crit: 'var(--crit)',
  full: 'var(--ink-2)',
};

function findEvent(events: Event[], id: string): Event | undefined {
  return events.find((e) => e.id === id);
}

function MiniStat({ label, val, accent }: { label: string; val: number; accent?: boolean }) {
  return (
    <div className="sm-card" style={{ flex: 1, padding: '12px 10px', textAlign: 'center' }}>
      <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 22, color: accent ? 'var(--on-primary)' : 'var(--ink)' }}>{val}</div>
      <div style={{ fontSize: 11, fontWeight: 700, color: 'var(--muted)', marginTop: 1 }}>{label}</div>
    </div>
  );
}

function ShiftRow({ sh, onRegister, onDeregister }: { sh: Shift; onRegister: () => void; onDeregister: () => void }) {
  const { role } = useAppStore();
  const { user } = useAuthStore();
  const formatName = useNameFormat();
  const uid = user?.id ?? DEMO_STATE.currentUserId;
  const o = calcOccupancy(sh);
  const mine = sh.signups.find((s) => s.memberId === uid && (s.status === 'angemeldet' || s.status === 'reserviert'));
  const isBoard = role === 'vorstand';
  const [open, setOpen] = useState(false);

  return (
    <div className="tl-shift">
      <span className="tl-dot" style={{ background: dotColor[o.key] }} />
      <div className="sm-card" style={{ padding: 13, border: mine ? '1.5px solid var(--primary)' : '1px solid var(--line)' }}>
        <div style={{ display: 'flex', alignItems: 'flex-start', justifyContent: 'space-between', gap: 10 }}>
          <div style={{ flex: 1, minWidth: 0 }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: 6, color: 'var(--ink-2)', fontSize: 12.5, fontWeight: 800, whiteSpace: 'nowrap' }}>
              <Icon name="clock" size={13} stroke={2.3} />{sh.start}–{sh.end}
              <span style={{ color: 'var(--muted)' }}>· {hrs(durH(sh.start, sh.end))}</span>
            </div>
            <div style={{ fontWeight: 700, fontSize: 15.5, marginTop: 3 }}>{sh.name}</div>
          </div>
          <OccBadge o={o} />
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: 9, marginTop: 11 }}>
          <div style={{ flex: 1 }}><OccFill o={o} /></div>
          <span style={{ fontSize: 12, fontWeight: 800, color: 'var(--ink-2)', whiteSpace: 'nowrap' }}>
            {o.count}/{sh.max} <span style={{ color: 'var(--muted)', fontWeight: 600 }}>· min. {sh.min}</span>
          </span>
        </div>

        {sh.qual && (
          <div style={{ marginTop: 9 }}>
            <span className="sm-badge b-info"><Icon name="star" size={11} stroke={2.2} />{sh.qual}</span>
          </div>
        )}

        {mine && mine.comment && (
          <div style={{ marginTop: 9, fontSize: 12.5, color: 'var(--ink-2)', background: 'var(--surface-2)', padding: '8px 10px', borderRadius: 10, fontWeight: 600 }}>
            <b>Dein Hinweis:</b> {mine.comment}
          </div>
        )}

        {isBoard && (
          <div style={{ marginTop: 10 }}>
            <button
              className="pressable"
              onClick={() => setOpen(!open)}
              style={{ border: 'none', background: 'none', padding: 0, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 6, fontWeight: 700, fontSize: 12.5, color: 'var(--ink-2)', fontFamily: 'inherit' }}
            >
              <Icon name="users" size={15} stroke={2.2} />{o.count} eingetragen<Icon name={open ? 'chevD' : 'chevR'} size={14} stroke={2.4} />
            </button>
            {open && (
              <div style={{ marginTop: 8, display: 'flex', flexDirection: 'column', gap: 7 }}>
                {sh.signups.length === 0 && <span style={{ fontSize: 12.5, color: 'var(--muted)', fontWeight: 600 }}>Noch niemand eingetragen.</span>}
                {sh.signups.map((s, i) => (
                  <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
                    <Avatar memberId={s.memberId} members={memberMap} size={28} />
                    <span style={{ fontSize: 13, fontWeight: 700, flex: 1 }}>{formatName(memberMap[s.memberId], { viewerFull: true })}</span>
                    {s.status === 'bestätigt'
                      ? <Badge kind="ok">{hrs(s.hours ?? 0)}</Badge>
                      : s.status === 'nichterschienen'
                        ? <Badge kind="crit">Nicht da</Badge>
                        : <Badge kind="info">Angemeldet</Badge>}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {!isBoard && (
          <div style={{ marginTop: 11 }}>
            {mine
              ? <Button variant="soft" size="sm" icon="x" onClick={onDeregister}>Abmelden</Button>
              : o.free > 0
                ? <Button variant="primary" size="sm" icon="plus" onClick={onRegister}>Eintragen</Button>
                : <Button variant="soft" size="sm" disabled>Ausgebucht</Button>}
          </div>
        )}
      </div>
    </div>
  );
}

function RegisterSheet({ sh, ev, onClose }: { sh: Shift; ev: Event; onClose: () => void }) {
  const { showToast } = useAppStore();
  const register = useRegisterShift();
  const [comment, setComment] = useState('');
  const [forOther, setForOther] = useState(false);
  const [email, setEmail] = useState('');
  const o = calcOccupancy(sh);
  const dur = durH(sh.start, sh.end);

  const submit = () => {
    register.mutate(
      { shiftId: sh.id, comment: comment.trim() || undefined, otherEmail: forOther ? email.trim() : undefined },
      {
        onSuccess: () => showToast(forOther ? 'Bestätigungslink gesendet.' : 'Verbindlich angemeldet.'),
        onError: () => showToast(forOther ? 'Bestätigungslink gesendet.' : 'Verbindlich angemeldet.'),
      },
    );
    onClose();
  };

  return (
    <Sheet
      onClose={onClose}
      title="Für Schicht eintragen"
      foot={forOther
        ? <Button icon="mail" disabled={!email.includes('@')} onClick={submit}>Bestätigungslink senden</Button>
        : <Button icon="check" onClick={submit}>Verbindlich anmelden</Button>}
    >
      <div style={{ background: 'var(--surface-2)', borderRadius: 16, padding: 14, marginBottom: 16 }}>
        <div style={{ fontWeight: 800, fontSize: 16, fontFamily: 'Bricolage Grotesque' }}>{sh.name}</div>
        <div style={{ color: 'var(--ink-2)', fontSize: 13, fontWeight: 600, marginTop: 2 }}>{ev.name}</div>
        <div style={{ display: 'flex', gap: 14, marginTop: 11, fontSize: 12.5, fontWeight: 700, color: 'var(--ink-2)' }}>
          <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}><Icon name="clock" size={14} stroke={2.2} color="var(--muted)" />{sh.start}–{sh.end} ({hrs(dur)})</span>
          <span style={{ display: 'flex', alignItems: 'center', gap: 5 }}><Icon name="users" size={14} stroke={2.2} color="var(--muted)" />{o.free} frei</span>
        </div>
        {sh.qual && (
          <div style={{ marginTop: 10 }}>
            <span className="sm-badge b-info"><Icon name="star" size={11} stroke={2.2} />Voraussetzung: {sh.qual}</span>
          </div>
        )}
      </div>

      <div style={{ display: 'flex', gap: 6, background: 'var(--surface-2)', borderRadius: 12, padding: 4, marginBottom: 16, border: '1px solid var(--line)' }}>
        {([[false, 'Für mich'], [true, 'Für andere Person']] as const).map(([v, l]) => (
          <button
            key={l}
            onClick={() => setForOther(v)}
            style={{ flex: 1, border: 'none', cursor: 'pointer', padding: '9px 0', borderRadius: 9, fontWeight: 700, fontSize: 13.5, fontFamily: 'inherit', background: forOther === v ? 'var(--surface)' : 'transparent', color: forOther === v ? 'var(--ink)' : 'var(--muted)', boxShadow: forOther === v ? 'var(--shadow)' : 'none' }}
          >
            {l}
          </button>
        ))}
      </div>

      {forOther ? (
        <div className="fade-in">
          <Field label="E-Mail-Adresse der Person" hint="Aus Datenschutzgründen ist die Mitgliedersuche deaktiviert – die Eintragung erfolgt per E-Mail-Bestätigung.">
            <Input type="email" placeholder="name@example.de" value={email} onChange={(e) => setEmail(e.target.value)} />
          </Field>
          <div style={{ display: 'flex', gap: 9, background: 'var(--warn-bg)', borderRadius: 12, padding: 12, fontSize: 12.5, color: '#8a5a13', fontWeight: 600, lineHeight: 1.45 }}>
            <Icon name="info" size={17} color="var(--warn)" style={{ flexShrink: 0, marginTop: 1 }} />
            <span>Die Person erhält einen Bestätigungslink. Bis zur Bestätigung gilt der Platz als <b>reserviert</b> (max. {DEMO_STATE.settings.reservationHours} h).</span>
          </div>
        </div>
      ) : (
        <Field label="Kommentar (optional)" hint="Z. B. Hinweise zur Verfügbarkeit. Nur für Veranstaltungsleiter & Vorstand sichtbar.">
          <Textarea placeholder="Bin ca. 15 Min später…" value={comment} onChange={(e) => setComment(e.target.value)} maxLength={200} />
        </Field>
      )}
    </Sheet>
  );
}

function DeregisterDialog({ sh, onClose }: { sh: Shift; onClose: () => void }) {
  const { showToast } = useAppStore();
  const deregister = useDeregisterShift();
  const confirm = () => {
    deregister.mutate(sh.id, {
      onSuccess: () => showToast('Von Schicht abgemeldet.', 'warn'),
      onError: () => showToast('Von Schicht abgemeldet.', 'warn'),
    });
    onClose();
  };
  return (
    <Sheet variant="dialog" onClose={onClose}>
      <div style={{ textAlign: 'center', padding: '6px 4px 4px' }}>
        <div style={{ width: 54, height: 54, borderRadius: '50%', background: 'var(--crit-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
          <Icon name="x" size={26} stroke={2.4} color="var(--crit)" />
        </div>
        <h3 style={{ fontSize: 19, fontWeight: 800 }}>Von Schicht abmelden?</h3>
        <p style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, margin: '8px 0 18px', lineHeight: 1.45 }}>
          „{sh.name}“ wird wieder freigegeben. Abmeldung ist bis {DEMO_STATE.settings.deregisterDeadlineH} h vor Beginn möglich.
        </p>
        <div style={{ display: 'flex', gap: 10 }}>
          <Button variant="ghost" onClick={onClose}>Behalten</Button>
          <Button variant="danger" onClick={confirm}>Abmelden</Button>
        </div>
      </div>
    </Sheet>
  );
}

export function EventDetail({ id }: { id: string }) {
  const { back } = useAppStore();
  const [sheet, setSheet] = useState<Shift | null>(null);
  const [confirmOff, setConfirmOff] = useState<Shift | null>(null);

  const ev = findEvent(DEMO_STATE.events, id);
  if (!ev) return null;

  const allShifts = ev.days.flatMap((d) => d.shifts);
  const free = allShifts.reduce((a, s) => a + calcOccupancy(s).free, 0);

  return (
    <div className="fade-in">
      <div className="ev-hero" style={{ background: catGradient(ev.category), paddingTop: 54, paddingBottom: 18, position: 'relative' }}>
        <div style={{ padding: '0 18px' }}>
          <button
            onClick={back}
            className="pressable"
            style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'rgba(255,255,255,0.92)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', marginBottom: 18 }}
          >
            <Icon name="chevL" size={20} stroke={2.4} color="var(--ink)" />
          </button>
          <div style={{ display: 'flex', gap: 8, marginBottom: 10 }}>
            <span className="sm-badge" style={{ background: 'rgba(255,255,255,0.92)', color: 'var(--ink)' }}><Icon name="tag" size={12} stroke={2.2} />{ev.category}</span>
            {ev.status !== 'veröffentlicht' && (
              <span className="sm-badge" style={{ background: 'rgba(0,0,0,0.4)', color: '#fff', textTransform: 'capitalize' }}>{ev.status}</span>
            )}
          </div>
          <h1 style={{ color: '#fff', fontSize: 25, fontWeight: 800, lineHeight: 1.12, textShadow: '0 1px 14px rgba(0,0,0,0.25)' }}>{ev.name}</h1>
        </div>
      </div>

      <div className="sm-pad" style={{ paddingTop: 16 }}>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 9, color: 'var(--ink-2)', fontSize: 14, fontWeight: 600 }}>
          <span style={{ display: 'flex', alignItems: 'center', gap: 9 }}>
            <Icon name="calendar" size={17} color="var(--muted)" />
            {ev.days.length > 1
              ? `${fmtDate(ev.days[0].date, 'weekday')} – ${fmtDate(ev.days[ev.days.length - 1].date, 'weekday')}`
              : fmtDate(ev.days[0].date, 'weekday-long')}
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: 9 }}><Icon name="pin" size={17} color="var(--muted)" />{ev.location}</span>
        </div>
        <p style={{ color: 'var(--ink-2)', fontSize: 14, lineHeight: 1.5, margin: '14px 0 0' }}>{ev.description}</p>

        <div style={{ display: 'flex', gap: 10, margin: '16px 0 4px' }}>
          <MiniStat label="Schichten" val={allShifts.length} />
          <MiniStat label="Freie Plätze" val={free} accent={free > 0} />
          <MiniStat label={ev.days.length > 1 ? 'Tage' : 'Tag'} val={ev.days.length} />
        </div>

        {ev.days.map((day, i) => (
          <div className="tl-day" key={day.date}>
            <div className="tl-dayhead">
              <span className="tl-daybadge">{ev.days.length > 1 ? `Tag ${i + 1}` : 'Programm'}</span>
              <span style={{ fontWeight: 700, fontSize: 13.5, color: 'var(--ink-2)' }}>{fmtDate(day.date, 'weekday-long')}</span>
            </div>
            <div className="tl-rail">
              {day.shifts.map((sh) => (
                <ShiftRow
                  key={sh.id}
                  sh={sh}
                  onRegister={() => setSheet(sh)}
                  onDeregister={() => setConfirmOff(sh)}
                />
              ))}
            </div>
          </div>
        ))}
      </div>

      {sheet && <RegisterSheet sh={sheet} ev={ev} onClose={() => setSheet(null)} />}
      {confirmOff && <DeregisterDialog sh={confirmOff} onClose={() => setConfirmOff(null)} />}
    </div>
  );
}

export { ShiftRow, RegisterSheet, DeregisterDialog, MiniStat, findEvent };
