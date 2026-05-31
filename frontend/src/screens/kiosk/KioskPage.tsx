import { useMemo, useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/forms/Field';
import { useKioskEvents, useKioskMemberSearch, useKioskRegister } from '@/api/kiosk';
import { useSettings, useBranding } from '@/api/settings';
import { calcOccupancy } from '@/hooks/useOccupancy';
import { LoadingState } from '@/components/ui/States';
import { DEMO_STATE, fmtDate, hrs, durH, catGradient } from '@/screens/_demo';
import type { Event, Shift, Member } from '@/types';

type Step = 'event' | 'shift' | 'identify' | 'done';

function clubInitials(name: string): string {
  const words = name.split(/\s+/).filter(Boolean);
  if (words.length === 0) return 'SG';
  if (words.length === 1) return words[0].slice(0, 2).toUpperCase();
  return (words[0][0] + words[words.length - 1][0]).toUpperCase();
}

function KioskHeader({ onBack, title, sub }: { onBack?: () => void; title: string; sub: string }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 16, marginBottom: 28 }}>
      {onBack && (
        <button
          onClick={onBack}
          className="pressable"
          style={{ width: 56, height: 56, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', flexShrink: 0 }}
        >
          <Icon name="chevL" size={26} stroke={2.4} />
        </button>
      )}
      <div>
        <div className="sm-eyebrow" style={{ fontSize: 13 }}>{sub}</div>
        <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 32, lineHeight: 1.1 }}>{title}</div>
      </div>
    </div>
  );
}

export function KioskPage() {
  const { data: settings } = useSettings();
  const { data: branding } = useBranding();
  const { data: remoteEvents, isLoading: eventsLoading } = useKioskEvents();
  const register = useKioskRegister();

  const kioskSearch = settings?.kioskSearch ?? DEMO_STATE.settings.kioskSearch;
  const clubName = branding?.clubName ?? settings?.clubName ?? DEMO_STATE.settings.clubName;
  const events = useMemo<Event[]>(
    () => (remoteEvents ?? DEMO_STATE.events).filter((e) => e.status === 'veröffentlicht'),
    [remoteEvents],
  );

  const [step, setStep] = useState<Step>('event');
  const [event, setEvent] = useState<Event | null>(null);
  const [shift, setShift] = useState<Shift | null>(null);
  const [email, setEmail] = useState('');
  const [query, setQuery] = useState('');

  const { data: searchResults } = useKioskMemberSearch(query, kioskSearch);

  const reset = () => {
    setStep('event');
    setEvent(null);
    setShift(null);
    setEmail('');
    setQuery('');
  };

  const submit = (member?: Member) => {
    if (!shift) return;
    register.mutate(
      member ? { shiftId: shift.id, memberId: member.id } : { shiftId: shift.id, email: email.trim() },
      { onSuccess: () => setStep('done'), onError: () => setStep('done') },
    );
  };

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)', display: 'flex', flexDirection: 'column' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '20px 32px', borderBottom: '1px solid var(--line)', background: 'var(--surface)' }}>
        {branding?.logoUrl ? (
          <img src={branding.logoUrl} alt={clubName} style={{ width: 40, height: 40, borderRadius: '50%', objectFit: 'cover' }} />
        ) : (
          <span className="rb-logo" style={{ width: 40, height: 40, fontSize: 15 }}>{clubInitials(clubName)}</span>
        )}
        <div>
          <div style={{ fontWeight: 800, fontSize: 16, fontFamily: 'Bricolage Grotesque' }}>ShiftManager Kiosk</div>
          <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{clubName}</div>
        </div>
      </div>

      <div style={{ flex: 1, width: '100%', maxWidth: 720, margin: '0 auto', padding: '32px 32px 48px' }}>
        {step === 'event' && (
          <div className="fade-in">
            <KioskHeader title="Veranstaltung wählen" sub="Schritt 1 von 3" />
            {eventsLoading && !remoteEvents && <LoadingState label="Veranstaltungen werden geladen…" />}
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              {events.map((ev) => {
                const free = ev.days.flatMap((d) => d.shifts).reduce((a, s) => a + calcOccupancy(s).free, 0);
                return (
                  <button
                    key={ev.id}
                    className="sm-card pressable"
                    onClick={() => { setEvent(ev); setStep('shift'); }}
                    style={{ textAlign: 'left', cursor: 'pointer', border: '1px solid var(--line)', padding: 0, overflow: 'hidden', fontFamily: 'inherit' }}
                  >
                    <div style={{ height: 10, background: catGradient(ev.category) }} />
                    <div style={{ padding: 20, display: 'flex', alignItems: 'center', gap: 14 }}>
                      <div style={{ flex: 1 }}>
                        <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 20 }}>{ev.name}</div>
                        <div style={{ color: 'var(--muted)', fontSize: 14, fontWeight: 600, marginTop: 4 }}>{fmtDate(ev.days[0].date, 'weekday')} · {ev.location}</div>
                      </div>
                      <Badge kind={free > 0 ? 'ok' : 'full'}>{free} frei</Badge>
                      <Icon name="chevR" size={24} color="var(--line-2)" stroke={2.4} />
                    </div>
                  </button>
                );
              })}
              {events.length === 0 && (
                <div style={{ textAlign: 'center', color: 'var(--muted)', fontWeight: 600, padding: 40 }}>Keine Veranstaltungen verfügbar.</div>
              )}
            </div>
          </div>
        )}

        {step === 'shift' && event && (
          <div className="fade-in">
            <KioskHeader onBack={() => setStep('event')} title={event.name} sub="Schritt 2 von 3 · Schicht wählen" />
            <div style={{ display: 'flex', flexDirection: 'column', gap: 14 }}>
              {event.days.flatMap((d) =>
                d.shifts.map((sh) => {
                  const o = calcOccupancy(sh);
                  const full = o.free <= 0;
                  return (
                    <button
                      key={sh.id}
                      disabled={full}
                      className={full ? '' : 'pressable'}
                      onClick={() => { if (!full) { setShift(sh); setStep('identify'); } }}
                      style={{ textAlign: 'left', cursor: full ? 'not-allowed' : 'pointer', opacity: full ? 0.55 : 1, border: '1px solid var(--line)', borderRadius: 'var(--radius)', background: 'var(--surface)', padding: 20, display: 'flex', alignItems: 'center', gap: 14, fontFamily: 'inherit' }}
                    >
                      <div style={{ flex: 1 }}>
                        <div style={{ fontWeight: 700, fontSize: 19 }}>{sh.name}</div>
                        <div style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, marginTop: 4, display: 'flex', alignItems: 'center', gap: 6 }}>
                          <Icon name="clock" size={15} stroke={2.2} color="var(--muted)" />{sh.start}–{sh.end} · {hrs(durH(sh.start, sh.end))} · {fmtDate(d.date, 'daymon')}
                        </div>
                      </div>
                      <Badge kind={full ? 'full' : o.key === 'crit' ? 'crit' : 'ok'}>{full ? 'Ausgebucht' : `${o.free} frei`}</Badge>
                      {!full && <Icon name="chevR" size={24} color="var(--line-2)" stroke={2.4} />}
                    </button>
                  );
                }),
              )}
            </div>
          </div>
        )}

        {step === 'identify' && shift && (
          <div className="fade-in">
            <KioskHeader onBack={() => setStep('shift')} title="Anmeldung" sub={`Schritt 3 von 3 · ${shift.name}`} />
            {kioskSearch ? (
              <div>
                <div style={{ fontWeight: 700, fontSize: 16, marginBottom: 10 }}>Mitglied suchen</div>
                <Input placeholder="Name eingeben…" value={query} onChange={(e) => setQuery(e.target.value)} style={{ fontSize: 18, padding: '16px 16px' }} />
                <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginTop: 16 }}>
                  {(searchResults ?? DEMO_STATE.members.filter((m) => query.trim().length >= 2 && (m.first + ' ' + m.last).toLowerCase().includes(query.toLowerCase()))).map((m) => (
                    <button
                      key={m.id}
                      className="pressable"
                      onClick={() => submit(m)}
                      style={{ textAlign: 'left', cursor: 'pointer', border: '1px solid var(--line)', borderRadius: 'var(--radius-sm)', background: 'var(--surface)', padding: 16, display: 'flex', alignItems: 'center', gap: 12, fontFamily: 'inherit' }}
                    >
                      <div style={{ flex: 1, fontWeight: 700, fontSize: 16 }}>{m.first} {m.last}</div>
                      <Icon name="arrowR" size={22} stroke={2.2} />
                    </button>
                  ))}
                </div>
              </div>
            ) : (
              <div>
                <div style={{ fontWeight: 700, fontSize: 16, marginBottom: 10 }}>E-Mail-Adresse</div>
                <Input type="email" placeholder="name@example.de" value={email} onChange={(e) => setEmail(e.target.value)} style={{ fontSize: 18, padding: '16px 16px' }} />
                <div style={{ display: 'flex', gap: 10, background: 'var(--surface-2)', borderRadius: 12, padding: 14, fontSize: 13.5, color: 'var(--ink-2)', fontWeight: 600, lineHeight: 1.5, margin: '16px 0' }}>
                  <Icon name="info" size={18} color="var(--info)" style={{ flexShrink: 0, marginTop: 1 }} />
                  <span>Du erhältst einen Bestätigungslink per E-Mail. Erst nach Bestätigung ist dein Platz verbindlich.</span>
                </div>
                <Button icon="mail" disabled={!email.includes('@')} loading={register.isPending} onClick={() => submit()}>Bestätigungslink senden</Button>
              </div>
            )}
          </div>
        )}

        {step === 'done' && (
          <div className="fade-in" style={{ textAlign: 'center', padding: '60px 20px' }}>
            <div style={{ width: 88, height: 88, borderRadius: '50%', background: 'var(--ok-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 22px' }}>
              <Icon name="check" size={44} stroke={2.4} color="var(--ok)" />
            </div>
            <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 28 }}>Bestätigung gesendet</div>
            <p style={{ color: 'var(--ink-2)', fontSize: 15.5, fontWeight: 600, lineHeight: 1.5, margin: '12px auto 28px', maxWidth: 420 }}>
              {kioskSearch
                ? 'Die Anmeldung wurde gespeichert. Vielen Dank für deine Unterstützung!'
                : 'Bitte prüfe dein Postfach und bestätige die Anmeldung über den Link in der E-Mail.'}
            </p>
            <div style={{ maxWidth: 320, margin: '0 auto' }}>
              <Button icon="plus" onClick={reset}>Weitere Anmeldung</Button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
