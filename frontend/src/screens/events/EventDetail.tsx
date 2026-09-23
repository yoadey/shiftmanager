import { useRef, useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Avatar } from '@/components/ui/Avatar';
import { OccBadge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { OccFill } from '@/components/ui/OccFill';
import { Sheet } from '@/components/ui/Sheet';
import { Field, Input, Textarea, Select } from '@/components/forms/Field';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import {
  useRegisterShift,
  useDeregisterShift,
  useCreateShift,
  useUpdateShift,
  useDeleteShift,
  usePatchRegistration,
  useForceDeleteRegistration,
  useAddMemberToShift,
  useAddGuestToShift,
} from '@/api/shifts';
import {
  useEvent,
  useEventTimeline,
  useUpdateEvent,
  useDeleteEvent,
  useCompleteEvent,
  useEventAttachments,
  useUploadEventAttachment,
  useDeleteEventAttachment,
} from '@/api/events';
import { useMembers } from '@/api/members';
import { useSettings } from '@/api/settings';
import { calcOccupancy } from '@/hooks/useOccupancy';
import { useNameFormat } from '@/hooks/useNameFormat';
import { LoadingState, ErrorState } from '@/components/ui/States';
import { fmtDate, hrs, durH, catGradient } from '@/screens/_demo';
import type { Event, Shift, Signup, Member, EventAttachment } from '@/types';

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

const MINUTES = [0, 5, 10, 15, 20, 25, 30, 35, 40, 45, 50, 55];

function decimalToHhMm(decimal: number): [number, number] {
  const h = Math.floor(decimal);
  const m = Math.round(((decimal % 1) * 60) / 5) * 5;
  return [h, Math.min(m, 55)];
}

function TimePickerSheet({ hh, mm, dur, onSave, onClose }: {
  hh: number; mm: number; dur: number;
  onSave: (h: number, m: number) => void;
  onClose: () => void;
}) {
  const [h, setH] = useState(hh);
  const [m, setM] = useState(mm);
  return (
    <Sheet variant="dialog" onClose={onClose}>
      <div style={{ textAlign: 'center', padding: '4px 0 4px' }}>
        <div style={{ fontSize: 38, fontWeight: 800, fontFamily: 'Bricolage Grotesque', margin: '4px 0 20px', letterSpacing: '-1px' }}>
          {h}<span style={{ fontSize: 22, color: 'var(--muted)', fontWeight: 700 }}>&thinsp;h&thinsp;</span>
          {String(m).padStart(2, '0')}<span style={{ fontSize: 22, color: 'var(--muted)', fontWeight: 700 }}>&thinsp;min</span>
        </div>

        <div style={{ marginBottom: 20 }}>
          <div style={{ fontSize: 11, fontWeight: 800, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '.06em', marginBottom: 10 }}>Stunden</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 18, justifyContent: 'center' }}>
            <button onClick={() => setH(Math.max(0, h - 1))} className="pressable"
              style={{ width: 52, height: 52, borderRadius: 14, border: '1.5px solid var(--line)', background: 'var(--surface)', cursor: 'pointer', fontSize: 24, fontWeight: 700, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>−</button>
            <span style={{ fontWeight: 800, fontSize: 30, minWidth: 48, textAlign: 'center', fontFamily: 'Bricolage Grotesque' }}>{h}</span>
            <button onClick={() => setH(Math.min(23, h + 1))} className="pressable"
              style={{ width: 52, height: 52, borderRadius: 14, border: '1.5px solid var(--line)', background: 'var(--surface)', cursor: 'pointer', fontSize: 24, fontWeight: 700, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>+</button>
          </div>
        </div>

        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 11, fontWeight: 800, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '.06em', marginBottom: 10 }}>Minuten</div>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: 7 }}>
            {MINUTES.map((min) => (
              <button key={min} onClick={() => setM(min)} className="pressable"
                style={{ padding: '11px 0', borderRadius: 11, border: m === min ? '2px solid var(--primary)' : '1.5px solid var(--line)', background: m === min ? 'color-mix(in srgb, var(--primary) 12%, var(--surface))' : 'var(--surface)', cursor: 'pointer', fontWeight: 700, fontSize: 14, fontFamily: 'inherit', color: m === min ? 'var(--primary)' : 'var(--ink)' }}>
                :{String(min).padStart(2, '0')}
              </button>
            ))}
          </div>
        </div>

        <div style={{ fontSize: 12, color: 'var(--muted)', fontWeight: 600, marginBottom: 18 }}>
          Schichtdauer: {hrs(dur)}
        </div>

        <div style={{ display: 'flex', gap: 10 }}>
          <Button variant="ghost" onClick={onClose}>Abbrechen</Button>
          <Button icon="check" onClick={() => onSave(h, m)}>Übernehmen</Button>
        </div>
      </div>
    </Sheet>
  );
}

function RegistrationRow({
  s, shift, eventId, memberMap,
}: {
  s: Signup; shift: Shift; eventId: string; memberMap: Record<string, Member>;
}) {
  const formatName = useNameFormat();
  const { showToast } = useAppStore();
  const patch = usePatchRegistration();
  const forceDelete = useForceDeleteRegistration();

  const dur = durH(shift.start, shift.end);
  const [[hh, mm], setHhMm] = useState<[number, number]>(() =>
    decimalToHhMm(s.hours != null ? s.hours : dur)
  );
  const [timeOpen, setTimeOpen] = useState(false);
  const [deleteConfirm, setDeleteConfirm] = useState(false);

  const isConfirmed = s.status === 'bestätigt';
  const isNoShow = s.status === 'nichterschienen';

  const doSaveHours = (h: number, m: number) => {
    patch.mutate({ registrationId: s.id ?? '', eventId, bookedHours: h + m / 60 },
      { onError: () => showToast('Fehler.', 'crit') });
  };

  const onTimeSave = (h: number, m: number) => {
    setHhMm([h, m]);
    doSaveHours(h, m);
    setTimeOpen(false);
  };

  const onConfirm = () => {
    const bookedH = hh + mm / 60;
    patch.mutate({
      registrationId: s.id ?? '', eventId,
      state: isConfirmed ? 'registered' : 'confirmed',
      bookedHours: !isConfirmed && bookedH > 0 ? bookedH : undefined,
    }, { onError: () => showToast('Fehler.', 'crit') });
  };

  const onNoShow = () => {
    patch.mutate({
      registrationId: s.id ?? '', eventId,
      state: isNoShow ? 'registered' : 'no_show',
    }, { onError: () => showToast('Fehler.', 'crit') });
  };

  const onRemove = () => {
    forceDelete.mutate({ registrationId: s.id ?? '', eventId },
      { onError: () => showToast('Entfernen fehlgeschlagen.', 'crit') });
  };

  const member = memberMap[s.memberId];
  const displayName = s.guest ?? (member ? formatName(member, { viewerFull: true }) : '—');
  const bg = isNoShow
    ? 'color-mix(in srgb, var(--crit) 8%, var(--surface))'
    : isConfirmed
      ? 'color-mix(in srgb, var(--ok) 8%, var(--surface))'
      : 'var(--surface-2)';

  return (
    <>
      <div style={{ display: 'flex', alignItems: 'center', gap: 5, padding: '6px 8px', borderRadius: 10, border: '1px solid var(--line)', background: bg }}>
        <Avatar memberId={s.memberId} members={memberMap} size={24} />
        <span style={{ fontSize: 12.5, fontWeight: 700, flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {displayName}
        </span>
        {!isNoShow && (
          <button onClick={() => setTimeOpen(true)} title="Zeit bearbeiten" className="pressable"
            style={{ padding: '3px 9px', borderRadius: 8, border: '1px solid var(--line)', background: 'var(--surface)', fontSize: 12.5, fontWeight: 800, fontFamily: 'inherit', cursor: 'pointer', flexShrink: 0, whiteSpace: 'nowrap' }}>
            {hh}:{String(mm).padStart(2, '0')}
          </button>
        )}
        <button onClick={onConfirm} title={isConfirmed ? 'Bestätigung aufheben' : 'Anwesend bestätigen'} className="pressable"
          style={{ width: 28, height: 28, borderRadius: 8, border: isConfirmed ? 'none' : '1px solid var(--ok)', background: isConfirmed ? 'var(--ok)' : 'transparent', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
          <Icon name="check" size={13} stroke={2.4} color={isConfirmed ? '#fff' : 'var(--ok)'} />
        </button>
        <button onClick={onNoShow} title={isNoShow ? 'Status zurücksetzen' : 'Nicht erschienen markieren'} className="pressable"
          style={{ width: 28, height: 28, borderRadius: 8, border: isNoShow ? 'none' : '1px solid var(--crit)', background: isNoShow ? 'var(--crit)' : 'transparent', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
          <Icon name="x" size={13} stroke={2.4} color={isNoShow ? '#fff' : 'var(--crit)'} />
        </button>
        <button onClick={() => setDeleteConfirm(true)} title="Entfernen" className="pressable"
          style={{ width: 28, height: 28, borderRadius: 8, border: '1px solid var(--line)', background: 'transparent', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
          <Icon name="trash" size={13} stroke={2} color="var(--muted)" />
        </button>
      </div>
      {timeOpen && (
        <TimePickerSheet hh={hh} mm={mm} dur={dur} onSave={onTimeSave} onClose={() => setTimeOpen(false)} />
      )}
      {deleteConfirm && (
        <Sheet variant="dialog" onClose={() => setDeleteConfirm(false)}>
          <div style={{ textAlign: 'center', padding: '6px 4px 4px' }}>
            <div style={{ width: 54, height: 54, borderRadius: '50%', background: 'var(--crit-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
              <Icon name="trash" size={26} stroke={2.4} color="var(--crit)" />
            </div>
            <h3 style={{ fontSize: 19, fontWeight: 800 }}>Person entfernen?</h3>
            <p style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, margin: '8px 0 18px', lineHeight: 1.45 }}>
              {displayName} wird aus der Schicht entfernt.
            </p>
            <div style={{ display: 'flex', gap: 10 }}>
              <Button variant="ghost" onClick={() => setDeleteConfirm(false)}>Abbrechen</Button>
              <Button variant="danger" onClick={() => { setDeleteConfirm(false); onRemove(); }}>Entfernen</Button>
            </div>
          </div>
        </Sheet>
      )}
    </>
  );
}

function AddMemberSheet({
  shift, eventId, members, onClose,
}: {
  shift: Shift; eventId: string; members: Member[]; onClose: () => void;
}) {
  const { showToast } = useAppStore();
  const formatName = useNameFormat();
  const addMember = useAddMemberToShift();
  const addGuest = useAddGuestToShift();
  const [mode, setMode] = useState<'member' | 'guest'>('member');
  const [search, setSearch] = useState('');
  const [guestName, setGuestName] = useState('');
  const [guestEmail, setGuestEmail] = useState('');

  const existingIds = new Set(shift.signups.map((s) => s.memberId));
  const filtered = members.filter((m) => {
    if (existingIds.has(m.id)) return false;
    if (m.active === false) return false;
    const q = search.toLowerCase();
    if (!q) return true;
    return `${m.first} ${m.last}`.toLowerCase().includes(q) || m.email.toLowerCase().includes(q);
  });

  const add = (memberId: string) => {
    addMember.mutate({ shiftId: shift.id, memberId, eventId }, {
      onSuccess: () => showToast('Helfer hinzugefügt.'),
      onError: () => showToast('Hinzufügen fehlgeschlagen.', 'crit'),
    });
    onClose();
  };

  const addGuestHelper = () => {
    const name = guestName.trim();
    if (!name) return;
    addGuest.mutate({ shiftId: shift.id, name, email: guestEmail.trim() || undefined, eventId }, {
      onSuccess: () => showToast('Helfer hinzugefügt.'),
      onError: () => showToast('Hinzufügen fehlgeschlagen.', 'crit'),
    });
    onClose();
  };

  const segStyle = (active: boolean): React.CSSProperties => ({
    flex: 1, border: 'none', cursor: 'pointer', padding: '9px 0', borderRadius: 9, fontWeight: 700, fontSize: 13.5, fontFamily: 'inherit',
    background: active ? 'var(--surface)' : 'transparent', color: active ? 'var(--ink)' : 'var(--muted)', boxShadow: active ? 'var(--shadow)' : 'none',
  });

  return (
    <Sheet onClose={onClose} title="Helfer hinzufügen">
      <div style={{ display: 'flex', gap: 6, marginBottom: 14, background: 'var(--surface-2)', borderRadius: 12, padding: 4, border: '1px solid var(--line)' }}>
        <button className="pressable" onClick={() => setMode('member')} style={segStyle(mode === 'member')}>Mitglied</button>
        <button className="pressable" onClick={() => setMode('guest')} style={segStyle(mode === 'guest')}>Gast</button>
      </div>

      {mode === 'member' ? (
        <>
          <Field label="Suche">
            <Input placeholder="Name oder E-Mail…" value={search} onChange={(e) => setSearch(e.target.value)} autoFocus />
          </Field>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6, marginTop: 4 }}>
            {filtered.length === 0 && (
              <span style={{ fontSize: 13, color: 'var(--muted)', fontWeight: 600, padding: '10px 0' }}>
                {search ? 'Keine Mitglieder gefunden.' : 'Alle Mitglieder bereits eingetragen.'}
              </span>
            )}
            {filtered.slice(0, 50).map((m) => (
              <button key={m.id} onClick={() => add(m.id)} className="pressable"
                style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '10px 12px', borderRadius: 12, border: '1px solid var(--line)', background: 'var(--surface)', cursor: 'pointer', fontFamily: 'inherit', textAlign: 'left', width: '100%' }}
              >
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontWeight: 700, fontSize: 14 }}>{formatName(m, { viewerFull: true })}</div>
                  <div style={{ fontSize: 12, color: 'var(--muted)', fontWeight: 600 }}>{m.email}</div>
                </div>
                <Icon name="plus" size={16} stroke={2.4} color="var(--primary)" />
              </button>
            ))}
          </div>
        </>
      ) : (
        <>
          <Field label="Name">
            <Input placeholder="Vor- und Nachname" value={guestName} onChange={(e) => setGuestName(e.target.value)} autoFocus />
          </Field>
          <Field label="E-Mail (optional)">
            <Input type="email" placeholder="name@beispiel.de" value={guestEmail} onChange={(e) => setGuestEmail(e.target.value)} />
          </Field>
          <Button
            icon="plus"
            onClick={addGuestHelper}
            disabled={!guestName.trim() || addGuest.isPending}
            style={{ width: '100%', marginTop: 8 }}
          >
            Hinzufügen
          </Button>
        </>
      )}
    </Sheet>
  );
}

function ShiftRow({ sh, eventId, onRegister, onDeregister, memberMap }: { sh: Shift; eventId: string; onRegister: () => void; onDeregister: () => void; memberMap: Record<string, Member> }) {
  const { role } = useAppStore();
  const { user } = useAuthStore();
  const uid = user?.id ?? '';
  const o = calcOccupancy(sh);
  const mine = sh.signups.find((s) => s.memberId === uid && (s.status === 'angemeldet' || s.status === 'reserviert'));
  const isBoard = role === 'vorstand';
  const [open, setOpen] = useState(false);
  const [addingMember, setAddingMember] = useState(false);

  return (
    <>
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
              <div style={{ marginTop: 8, display: 'flex', flexDirection: 'column', gap: 6 }}>
                {sh.signups.length === 0 && <span style={{ fontSize: 12.5, color: 'var(--muted)', fontWeight: 600 }}>Noch niemand eingetragen.</span>}
                {sh.signups.map((s) => (
                  <RegistrationRow
                    key={s.id ?? s.memberId}
                    s={s}
                    shift={sh}
                    eventId={eventId}
                    memberMap={memberMap}
                  />
                ))}
                <Button variant="soft" size="sm" icon="plus" onClick={() => setAddingMember(true)}>Helfer hinzufügen</Button>
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

    {addingMember && (
      <AddMemberSheet
        shift={sh}
        eventId={eventId}
        members={Object.values(memberMap)}
        onClose={() => setAddingMember(false)}
      />
    )}
    </>
  );
}

function RegisterSheet({ sh, ev, onClose, reservationHours }: { sh: Shift; ev: Event; onClose: () => void; reservationHours: number }) {
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
            <span>Die Person erhält einen Bestätigungslink. Bis zur Bestätigung gilt der Platz als <b>reserviert</b> (max. {reservationHours} h).</span>
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

function DeregisterDialog({ sh, onClose, deregisterDeadlineH }: { sh: Shift; onClose: () => void; deregisterDeadlineH: number }) {
  const { showToast } = useAppStore();
  const deregister = useDeregisterShift();
  const confirm = () => {
    deregister.mutate(sh.id, {
      onSuccess: () => showToast('Von Schicht abgemeldet.', 'warn'),
      onError: () => showToast('Abmelden fehlgeschlagen.', 'crit'),
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
          „{sh.name}“ wird wieder freigegeben. Abmeldung ist bis {deregisterDeadlineH} h vor Beginn möglich.
        </p>
        <div style={{ display: 'flex', gap: 10 }}>
          <Button variant="ghost" onClick={onClose}>Behalten</Button>
          <Button variant="danger" onClick={confirm}>Abmelden</Button>
        </div>
      </div>
    </Sheet>
  );
}

type ShiftDraft = {
  key: string;
  id: string;
  name: string;
  date: string;
  start: string;
  end: string;
  min: number;
  max: number;
  qual: string;
  deleted: boolean;
};

const CATEGORIES = ['Turnier', 'Vereinsleben', 'Mitgliederwerbung', 'Training', 'Sonstiges'];
const STATUS_OPTIONS: { value: string; label: string; sub: string }[] = [
  { value: 'draft',     label: 'Entwurf',       sub: 'Nur intern sichtbar' },
  { value: 'published', label: 'Veröffentlicht', sub: 'Für Mitglieder sichtbar' },
  { value: 'cancelled', label: 'Abgesagt',       sub: 'Veranstaltung abgesagt' },
];

function ShiftForm({
  initial, eventStartDate, onSave, onCancel,
}: {
  initial: Omit<ShiftDraft, 'key' | 'id' | 'deleted'>;
  eventStartDate: string;
  onSave: (v: Omit<ShiftDraft, 'key' | 'id' | 'deleted'>) => void;
  onCancel: () => void;
}) {
  const [name, setName] = useState(initial.name);
  const [date, setDate] = useState(initial.date || eventStartDate);
  const [start, setStart] = useState(initial.start || '10:00');
  const [end, setEnd] = useState(initial.end || '14:00');
  const [min, setMin] = useState(initial.min || 2);
  const [max, setMax] = useState(initial.max || 4);
  const [qual, setQual] = useState(initial.qual || '');
  const valid = !!(name.trim() && date && end > start);
  return (
    <div className="sm-card" style={{ padding: 14, border: '1.5px solid var(--primary)', marginBottom: 10 }}>
      <Field label="Bezeichnung">
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="z. B. Einlass & Kasse" />
      </Field>
      <Field label="Datum">
        <Input type="date" value={date} onChange={(e) => setDate(e.target.value)} />
      </Field>
      <div style={{ display: 'flex', gap: 10 }}>
        <Field label="Von"><Input type="time" value={start} onChange={(e) => setStart(e.target.value)} /></Field>
        <Field label="Bis"><Input type="time" value={end} onChange={(e) => setEnd(e.target.value)} /></Field>
      </div>
      <div style={{ display: 'flex', gap: 22, margin: '4px 0 14px' }}>
        <div>
          <div className="sm-label">Min. Helfer</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 6 }}>
            <button className="pressable" onClick={() => setMin(Math.max(1, min - 1))} style={{ width: 30, height: 30, borderRadius: 8, border: '1px solid var(--line)', background: 'var(--surface)', cursor: 'pointer', fontWeight: 800, fontSize: 16 }}>−</button>
            <span style={{ fontWeight: 800, minWidth: 20, textAlign: 'center' }}>{min}</span>
            <button className="pressable" onClick={() => setMin(Math.min(min + 1, max))} style={{ width: 30, height: 30, borderRadius: 8, border: '1px solid var(--line)', background: 'var(--surface)', cursor: 'pointer', fontWeight: 800, fontSize: 16 }}>+</button>
          </div>
        </div>
        <div>
          <div className="sm-label">Max. Helfer</div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 6 }}>
            <button className="pressable" onClick={() => setMax(Math.max(max - 1, min))} style={{ width: 30, height: 30, borderRadius: 8, border: '1px solid var(--line)', background: 'var(--surface)', cursor: 'pointer', fontWeight: 800, fontSize: 16 }}>−</button>
            <span style={{ fontWeight: 800, minWidth: 20, textAlign: 'center' }}>{max}</span>
            <button className="pressable" onClick={() => setMax(max + 1)} style={{ width: 30, height: 30, borderRadius: 8, border: '1px solid var(--line)', background: 'var(--surface)', cursor: 'pointer', fontWeight: 800, fontSize: 16 }}>+</button>
          </div>
        </div>
      </div>
      <Field label="Qualifikation (optional)">
        <Input value={qual} onChange={(e) => setQual(e.target.value)} placeholder="z. B. Ersthelfer" />
      </Field>
      <div style={{ display: 'flex', gap: 10 }}>
        <Button variant="ghost" onClick={onCancel}>Abbrechen</Button>
        <Button icon="check" disabled={!valid} onClick={() => onSave({ name: name.trim(), date, start, end, min, max, qual })}>Übernehmen</Button>
      </div>
    </div>
  );
}

function EditEventSheet({ ev, onClose }: { ev: Event; onClose: () => void }) {
  const { showToast } = useAppStore();
  const updateEvent = useUpdateEvent();
  const createShift = useCreateShift();
  const updateShift = useUpdateShift();
  const deleteShift = useDeleteShift();

  const startDay = ev.days[0]?.date ?? '';
  const endDay = ev.days[ev.days.length - 1]?.date ?? startDay;

  const [name, setName] = useState(ev.name);
  const [description, setDescription] = useState(ev.description);
  const [location, setLocation] = useState(ev.location);
  const [category, setCategory] = useState(ev.category);
  const [startDate, setStartDate] = useState(startDay);
  const [endDate, setEndDate] = useState(endDay);
  const [status, setStatus] = useState(ev.status as string);

  const initShifts: ShiftDraft[] = ev.days.flatMap((d) =>
    d.shifts.map((s) => ({
      key: s.id,
      id: s.id,
      name: s.name,
      date: d.date,
      start: s.start,
      end: s.end,
      min: s.min,
      max: s.max,
      qual: s.qual ?? '',
      deleted: false,
    })),
  );
  const [shifts, setShifts] = useState<ShiftDraft[]>(initShifts);
  const [editingKey, setEditingKey] = useState<string | null>(null);
  const [addingNew, setAddingNew] = useState(false);

  const isPending = updateEvent.isPending || createShift.isPending || updateShift.isPending || deleteShift.isPending;

  const submit = async () => {
    try {
      const sd = startDate || startDay;
      const ed = endDate || endDay || sd;
      await updateEvent.mutateAsync({
        id: ev.id,
        name: name.trim(),
        description: description.trim(),
        location: location.trim(),
        category,
        status,
        startDate: sd ? `${sd}T12:00:00Z` : undefined,
        endDate: ed ? `${ed}T12:00:00Z` : undefined,
      });
      for (const s of shifts) {
        if (s.deleted && s.id) {
          await deleteShift.mutateAsync({ id: s.id, eventId: ev.id });
        } else if (!s.deleted && s.id) {
          await updateShift.mutateAsync({ id: s.id, eventId: ev.id, date: s.date, name: s.name, start: s.start, end: s.end, min: s.min, max: s.max, qual: s.qual });
        } else if (!s.deleted && !s.id) {
          await createShift.mutateAsync({ eventId: ev.id, date: s.date, name: s.name, start: s.start, end: s.end, min: s.min, max: s.max, qual: s.qual });
        }
      }
      showToast('Veranstaltung gespeichert.');
      onClose();
    } catch {
      showToast('Speichern fehlgeschlagen.', 'crit');
    }
  };

  const visibleShifts = shifts.filter((s) => !s.deleted);

  return (
    <Sheet variant="full" onClose={onClose} title="Veranstaltung bearbeiten" foot={<Button icon="check" onClick={submit} disabled={!name.trim() || isPending}>Speichern</Button>}>
      {/* Eckdaten */}
      <div style={{ fontWeight: 800, fontSize: 13, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '.06em', marginBottom: 10 }}>Eckdaten</div>
      <Field label="Name">
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="z. B. Sommerturnier 2026" />
      </Field>
      <Field label="Beschreibung">
        <Textarea value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Beschreibung für Helfer…" />
      </Field>
      <Field label="Ort">
        <Input value={location} onChange={(e) => setLocation(e.target.value)} placeholder="z. B. Sporthalle Aachen-Brand" />
      </Field>
      <Field label="Kategorie">
        <Select value={category} onChange={(e) => setCategory(e.target.value)}>
          {CATEGORIES.map((c) => <option key={c}>{c}</option>)}
        </Select>
      </Field>
      <div style={{ display: 'flex', gap: 10 }}>
        <Field label="Von">
          <Input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} />
        </Field>
        <Field label="Bis">
          <Input type="date" value={endDate} min={startDate} onChange={(e) => setEndDate(e.target.value)} />
        </Field>
      </div>
      <Field label="Status">
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {STATUS_OPTIONS.map((opt) => (
            <button
              key={opt.value}
              onClick={() => setStatus(opt.value)}
              className="pressable"
              style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '11px 14px', borderRadius: 12, border: status === opt.value ? '1.5px solid var(--primary)' : '1.5px solid var(--line-2)', background: status === opt.value ? 'color-mix(in srgb, var(--primary) 10%, var(--surface))' : 'var(--surface)', cursor: 'pointer', fontFamily: 'inherit', textAlign: 'left' }}
            >
              <div style={{ width: 10, height: 10, borderRadius: '50%', background: status === opt.value ? 'var(--primary)' : 'var(--line)', flexShrink: 0 }} />
              <div>
                <div style={{ fontWeight: 700, fontSize: 14 }}>{opt.label}</div>
                <div style={{ fontSize: 12, color: 'var(--muted)', fontWeight: 600 }}>{opt.sub}</div>
              </div>
            </button>
          ))}
        </div>
      </Field>

      {/* Schichten */}
      <div style={{ fontWeight: 800, fontSize: 13, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: '.06em', margin: '20px 0 10px' }}>
        Schichten ({visibleShifts.length})
      </div>

      {visibleShifts.length === 0 && !addingNew && (
        <div style={{ color: 'var(--muted)', fontSize: 13.5, fontWeight: 600, padding: '12px 0 4px' }}>Noch keine Schichten.</div>
      )}

      {shifts.filter((s) => !s.deleted).map((s) =>
        editingKey === s.key ? (
          <ShiftForm
            key={s.key}
            initial={s}
            eventStartDate={startDate || startDay}
            onSave={(v) => {
              setShifts((prev) => prev.map((x) => x.key === s.key ? { ...x, ...v } : x));
              setEditingKey(null);
            }}
            onCancel={() => setEditingKey(null)}
          />
        ) : (
          <div key={s.key} className="sm-card" style={{ padding: 12, marginBottom: 8, display: 'flex', gap: 11, alignItems: 'center' }}>
            <div style={{ width: 4, alignSelf: 'stretch', borderRadius: 4, background: 'var(--primary)' }} />
            <div style={{ flex: 1, minWidth: 0 }}>
              <div style={{ fontWeight: 700, fontSize: 14.5 }}>{s.name}</div>
              <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>
                {fmtDate(s.date, 'daymon')} · {s.start}–{s.end} · {s.min}–{s.max} Helfer
              </div>
            </div>
            <button className="pressable" onClick={() => setEditingKey(s.key)} style={{ width: 32, height: 32, borderRadius: 9, border: '1px solid var(--line)', background: 'var(--surface)', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <Icon name="edit" size={15} color="var(--ink-2)" />
            </button>
            <button className="pressable" onClick={() => setShifts((prev) => prev.map((x) => x.key === s.key ? { ...x, deleted: true } : x))} style={{ width: 32, height: 32, borderRadius: 9, border: '1px solid var(--line)', background: 'var(--surface)', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <Icon name="trash" size={15} color="var(--crit)" />
            </button>
          </div>
        ),
      )}

      {addingNew ? (
        <ShiftForm
          initial={{ name: '', date: startDate || startDay, start: '10:00', end: '14:00', min: 2, max: 4, qual: '' }}
          eventStartDate={startDate || startDay}
          onSave={(v) => {
            setShifts((prev) => [...prev, { key: 'new-' + Date.now(), id: '', ...v, deleted: false }]);
            setAddingNew(false);
          }}
          onCancel={() => setAddingNew(false)}
        />
      ) : (
        <Button variant="soft" icon="plus" onClick={() => { setEditingKey(null); setAddingNew(true); }}>Schicht hinzufügen</Button>
      )}
    </Sheet>
  );
}

function CompleteEventDialog({ ev, onClose }: { ev: Event; onClose: () => void }) {
  const { showToast } = useAppStore();
  const completeEvent = useCompleteEvent();
  const confirm = () => {
    completeEvent.mutate(ev.id, {
      onSuccess: () => showToast('Veranstaltung abgeschlossen. Alle angemeldeten Schichten wurden bestätigt.', 'ok'),
      onError: () => showToast('Abschließen fehlgeschlagen.', 'crit'),
    });
    onClose();
  };
  return (
    <Sheet variant="dialog" onClose={onClose}>
      <div style={{ textAlign: 'center', padding: '6px 4px 4px' }}>
        <div style={{ width: 54, height: 54, borderRadius: '50%', background: 'color-mix(in srgb, var(--ok) 15%, var(--surface))', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
          <Icon name="check" size={26} stroke={2.4} color="var(--ok)" />
        </div>
        <h3 style={{ fontSize: 19, fontWeight: 800 }}>Veranstaltung abschließen?</h3>
        <p style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, margin: '8px 0 18px', lineHeight: 1.45 }}>
          „{ev.name}" wird als <b>abgeschlossen</b> markiert. Alle angemeldeten Helferschichten werden automatisch bestätigt.
        </p>
        <div style={{ display: 'flex', gap: 10 }}>
          <Button variant="ghost" onClick={onClose}>Abbrechen</Button>
          <Button variant="primary" icon="check" onClick={confirm}>Abschließen</Button>
        </div>
      </div>
    </Sheet>
  );
}

function DeleteEventDialog({ ev, onClose, onDeleted }: { ev: Event; onClose: () => void; onDeleted: () => void }) {
  const { showToast } = useAppStore();
  const deleteEvent = useDeleteEvent();
  const confirm = () => {
    deleteEvent.mutate(ev.id, {
      onSuccess: () => { showToast('Veranstaltung gelöscht.', 'warn'); onDeleted(); },
      onError: () => showToast('Löschen fehlgeschlagen.', 'crit'),
    });
    onClose();
  };
  return (
    <Sheet variant="dialog" onClose={onClose}>
      <div style={{ textAlign: 'center', padding: '6px 4px 4px' }}>
        <div style={{ width: 54, height: 54, borderRadius: '50%', background: 'var(--crit-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
          <Icon name="trash" size={26} stroke={2.4} color="var(--crit)" />
        </div>
        <h3 style={{ fontSize: 19, fontWeight: 800 }}>Veranstaltung löschen?</h3>
        <p style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, margin: '8px 0 18px', lineHeight: 1.45 }}>
          „{ev.name}" und alle Schichten werden unwiderruflich gelöscht.
        </p>
        <div style={{ display: 'flex', gap: 10 }}>
          <Button variant="ghost" onClick={onClose}>Abbrechen</Button>
          <Button variant="danger" onClick={confirm}>Löschen</Button>
        </div>
      </div>
    </Sheet>
  );
}

// SVG is intentionally excluded: it can embed <script>, and unlike the
// Vorstand-only logo upload this is reachable by any Veranstaltungsleiter.
const ALLOWED_ATTACHMENT_TYPES = 'image/png,image/jpeg,image/gif,image/webp,application/pdf';

// Confirms removing a single attachment before it happens — a stray tap on
// the small delete button (20x20, easy to misclick, especially on touch)
// would otherwise delete it with no undo, unlike full event deletion which
// already goes through DeleteEventDialog.
function DeleteAttachmentDialog({ attachment, onClose, onConfirm }: { attachment: EventAttachment; onClose: () => void; onConfirm: () => void }) {
  return (
    <Sheet variant="dialog" onClose={onClose}>
      <div style={{ textAlign: 'center', padding: '6px 4px 4px' }}>
        <div style={{ width: 54, height: 54, borderRadius: '50%', background: 'var(--crit-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
          <Icon name="trash" size={26} stroke={2.4} color="var(--crit)" />
        </div>
        <h3 style={{ fontSize: 19, fontWeight: 800 }}>Anhang entfernen?</h3>
        <p style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, margin: '8px 0 18px', lineHeight: 1.45 }}>
          „{attachment.fileName}" wird unwiderruflich entfernt.
        </p>
        <div style={{ display: 'flex', gap: 10 }}>
          <Button variant="ghost" onClick={onClose}>Abbrechen</Button>
          <Button variant="danger" onClick={onConfirm}>Entfernen</Button>
        </div>
      </div>
    </Sheet>
  );
}

// Images and other attachments for an event (V-008). Board members can add
// and remove files; everyone else just sees what's there (and the section is
// hidden entirely for them when it's empty, to keep the detail page tidy).
function AttachmentsSection({ eventId, isBoard }: { eventId: string; isBoard: boolean }) {
  const { showToast } = useAppStore();
  const { data: attachments } = useEventAttachments(eventId);
  const upload = useUploadEventAttachment(eventId);
  const remove = useDeleteEventAttachment(eventId);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [confirmDelete, setConfirmDelete] = useState<EventAttachment | null>(null);
  // Tracked separately from remove.isPending, which is shared across the
  // whole mutation and would otherwise disable every attachment's delete
  // button while any single one is in flight. A Set (not a single id) so
  // confirming a second delete while a first is still in flight doesn't
  // re-enable the first attachment's button before its request settles.
  const [deletingIds, setDeletingIds] = useState<ReadonlySet<string>>(new Set());

  const onFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) return;
    upload.mutate(file, {
      onError: () => showToast('Datei konnte nicht hochgeladen werden.', 'crit'),
    });
  };

  const confirmRemove = () => {
    if (!confirmDelete) return;
    const id = confirmDelete.id;
    setDeletingIds((current) => new Set(current).add(id));
    remove.mutate(id, {
      onError: () => showToast('Anhang konnte nicht entfernt werden.', 'crit'),
      onSettled: () =>
        setDeletingIds((current) => {
          const next = new Set(current);
          next.delete(id);
          return next;
        }),
    });
    setConfirmDelete(null);
  };

  if (!isBoard && (!attachments || attachments.length === 0)) return null;

  return (
    <div style={{ margin: '18px 0 4px' }}>
      <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: 9 }}>
        <span style={{ fontWeight: 700, fontSize: 12.5, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.3 }}>Anhänge</span>
        {isBoard && (
          <>
            <button
              onClick={() => fileInputRef.current?.click()}
              disabled={upload.isPending}
              className="pressable"
              title="Bild oder Datei hinzufügen"
              style={{ width: 26, height: 26, borderRadius: '50%', border: '1px solid var(--line-2)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
            >
              <Icon name="plus" size={13} stroke={2.6} color="var(--ink-2)" />
            </button>
            <input
              ref={fileInputRef}
              type="file"
              accept={ALLOWED_ATTACHMENT_TYPES}
              style={{ display: 'none' }}
              onChange={onFileChange}
            />
          </>
        )}
      </div>
      {attachments && attachments.length > 0 && (
        <div style={{ display: 'flex', gap: 8, overflowX: 'auto', paddingBottom: 2 }}>
          {attachments.map((a) => (
            <div key={a.id} className="sm-card" style={{ position: 'relative', flexShrink: 0, width: 84, padding: 0, overflow: 'hidden' }}>
              <a href={a.url} target="_blank" rel="noreferrer" style={{ display: 'block', color: 'inherit', textDecoration: 'none' }}>
                {a.contentType.startsWith('image/') ? (
                  <img src={a.url} alt={a.fileName} style={{ width: '100%', height: 64, objectFit: 'cover', display: 'block' }} />
                ) : (
                  <div style={{ width: '100%', height: 64, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'var(--surface-2)' }}>
                    <Icon name="paperclip" size={20} color="var(--muted)" />
                  </div>
                )}
                <div style={{ fontSize: 10, fontWeight: 600, color: 'var(--muted)', padding: '4px 6px', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                  {a.fileName}
                </div>
              </a>
              {isBoard && (
                <button
                  onClick={() => setConfirmDelete(a)}
                  disabled={deletingIds.has(a.id)}
                  className="pressable"
                  title="Anhang entfernen"
                  style={{ position: 'absolute', top: 4, right: 4, width: 20, height: 20, borderRadius: '50%', border: 'none', background: 'rgba(0,0,0,0.55)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
                >
                  <Icon name="x" size={11} stroke={3} color="#fff" />
                </button>
              )}
            </div>
          ))}
        </div>
      )}
      {confirmDelete && (
        <DeleteAttachmentDialog attachment={confirmDelete} onClose={() => setConfirmDelete(null)} onConfirm={confirmRemove} />
      )}
    </div>
  );
}

export function EventDetail({ id }: { id: string }) {
  const { back } = useAppStore();
  const { role } = useAppStore();
  const isBoard = role === 'vorstand';
  const isOrganizer = isBoard;
  const [sheet, setSheet] = useState<Shift | null>(null);
  const [confirmOff, setConfirmOff] = useState<Shift | null>(null);
  const [editOpen, setEditOpen] = useState(false);
  const [deleteOpen, setDeleteOpen] = useState(false);
  const [completeOpen, setCompleteOpen] = useState(false);

  const eventQ = useEvent(id);
  const timelineQ = useEventTimeline(id);
  const { data: settings } = useSettings();
  // Member names are only needed in the board view (otherwise spare the request).
  const { data: members } = useMembers();
  const memberMap: Record<string, Member> = (members ?? []).reduce(
    (acc, m) => { acc[m.id] = m; return acc; },
    {} as Record<string, Member>,
  );

  const reservationHours = settings?.reservationHours ?? 48;
  const deregisterDeadlineH = settings?.deregisterDeadlineH ?? 24;

  if (eventQ.isLoading || timelineQ.isLoading) return <LoadingState />;
  if (eventQ.isError) return <ErrorState />;

  const ev = eventQ.data;
  if (!ev) return <ErrorState text="Diese Veranstaltung wurde nicht gefunden." />;

  // Prefer the dedicated timeline (multi-day grouped, occupancy ready); fall back to the event days.
  const days = timelineQ.data?.days ?? ev.days ?? [];
  const allShifts = days.flatMap((d) => d.shifts);
  const free = allShifts.reduce((a, s) => a + calcOccupancy(s).free, 0);

  return (
    <div className="fade-in">
      <div className="ev-hero" style={{ background: catGradient(ev.category), paddingTop: 54, paddingBottom: 18, position: 'relative' }}>
        <div style={{ padding: '0 18px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 18 }}>
            <button
              onClick={back}
              className="pressable"
              style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'rgba(255,255,255,0.92)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
            >
              <Icon name="chevL" size={20} stroke={2.4} color="var(--ink)" />
            </button>
            {isOrganizer && (
              <div style={{ display: 'flex', gap: 8, marginLeft: 'auto' }}>
                {isOrganizer && (ev.status === 'published' || ev.status === 'veröffentlicht') && (
                  <button
                    onClick={() => setCompleteOpen(true)}
                    className="pressable"
                    title="Veranstaltung abschließen"
                    style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'rgba(255,255,255,0.92)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
                  >
                    <Icon name="check" size={18} stroke={2.4} color="var(--ok)" />
                  </button>
                )}
                {isBoard && (
                  <>
                    <button
                      onClick={() => setEditOpen(true)}
                      className="pressable"
                      title="Veranstaltung bearbeiten"
                      style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'rgba(255,255,255,0.92)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
                    >
                      <Icon name="edit" size={18} stroke={2.2} color="var(--ink)" />
                    </button>
                    <button
                      onClick={() => setDeleteOpen(true)}
                      className="pressable"
                      title="Veranstaltung löschen"
                      style={{ width: 40, height: 40, borderRadius: '50%', border: 'none', background: 'rgba(255,255,255,0.92)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
                    >
                      <Icon name="trash" size={18} stroke={2.2} color="var(--crit)" />
                    </button>
                  </>
                )}
              </div>
            )}
          </div>
          <div style={{ display: 'flex', gap: 8, marginBottom: 10 }}>
            <span className="sm-badge" style={{ background: 'rgba(255,255,255,0.92)', color: 'var(--ink)' }}><Icon name="tag" size={12} stroke={2.2} />{ev.category}</span>
            {ev.status !== 'veröffentlicht' && ev.status !== 'published' && (
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
            {days.length === 0
              ? 'Termin offen'
              : days.length > 1
                ? `${fmtDate(days[0].date, 'weekday')} – ${fmtDate(days[days.length - 1].date, 'weekday')}`
                : fmtDate(days[0].date, 'weekday-long')}
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: 9 }}><Icon name="pin" size={17} color="var(--muted)" />{ev.location}</span>
        </div>
        <p style={{ color: 'var(--ink-2)', fontSize: 14, lineHeight: 1.5, margin: '14px 0 0' }}>{ev.description}</p>

        <AttachmentsSection eventId={id} isBoard={isBoard} />

        <div style={{ display: 'flex', gap: 10, margin: '16px 0 4px' }}>
          <MiniStat label="Schichten" val={allShifts.length} />
          <MiniStat label="Freie Plätze" val={free} accent={free > 0} />
          <MiniStat label={days.length > 1 ? 'Tage' : 'Tag'} val={days.length} />
        </div>

        {days.map((day, i) => (
          <div className="tl-day" key={day.date}>
            <div className="tl-dayhead">
              <span className="tl-daybadge">{days.length > 1 ? `Tag ${i + 1}` : 'Programm'}</span>
              <span style={{ fontWeight: 700, fontSize: 13.5, color: 'var(--ink-2)' }}>{fmtDate(day.date, 'weekday-long')}</span>
            </div>
            <div className="tl-rail">
              {day.shifts.map((sh) => (
                <ShiftRow
                  key={sh.id}
                  sh={sh}
                  eventId={id}
                  memberMap={memberMap}
                  onRegister={() => setSheet(sh)}
                  onDeregister={() => setConfirmOff(sh)}
                />
              ))}
            </div>
          </div>
        ))}
      </div>

      {sheet && <RegisterSheet sh={sheet} ev={ev} onClose={() => setSheet(null)} reservationHours={reservationHours} />}
      {confirmOff && <DeregisterDialog sh={confirmOff} onClose={() => setConfirmOff(null)} deregisterDeadlineH={deregisterDeadlineH} />}
      {editOpen && <EditEventSheet ev={ev} onClose={() => setEditOpen(false)} />}
      {deleteOpen && <DeleteEventDialog ev={ev} onClose={() => setDeleteOpen(false)} onDeleted={back} />}
      {completeOpen && <CompleteEventDialog ev={ev} onClose={() => setCompleteOpen(false)} />}
    </div>
  );
}

export { ShiftRow, RegisterSheet, DeregisterDialog, MiniStat, findEvent };
