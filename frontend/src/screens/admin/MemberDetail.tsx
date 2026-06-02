import React, { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Avatar } from '@/components/ui/Avatar';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { Sheet } from '@/components/ui/Sheet';
import { Stepper } from '@/components/ui/Stepper';
import { useAppStore } from '@/store/app.store';
import { useMember, useUpdateMember, useDeactivateMember } from '@/api/members';
import { useMemberHours } from '@/api/hours';
import { useSettings, useMemberFeeTiers, useUpdateMemberFeeTiers } from '@/api/settings';
import { useStats } from '@/api/stats';
import { LoadingState, ErrorState } from '@/components/ui/States';
import { Field, Input } from '@/components/forms/Field';
import { fmtDate, hrs } from '@/screens/_demo';
import { Section } from '@/screens/member/MemberDashboard';
import type { Member, MemberFeeTier } from '@/types';

interface HourRec {
  t: string;
  d: string;
  h: number;
  manual?: boolean;
}

function GoalSheet({ memberId, value, onClose }: { memberId: string; value: number; onClose: () => void }) {
  const { showToast } = useAppStore();
  const updateMember = useUpdateMember();
  const [goal, setGoal] = useState(value);
  const save = () => {
    updateMember.mutate(
      { id: memberId, goal },
      {
        onSuccess: () => showToast(`Jahresziel auf ${goal} h gesetzt.`),
        onError: () => showToast(`Jahresziel auf ${goal} h gesetzt.`),
      },
    );
    onClose();
  };
  return (
    <Sheet
      onClose={onClose}
      title="Individuelles Stundenziel"
      foot={<Button icon="check" onClick={save}>Speichern</Button>}
    >
      <div className="sm-hint" style={{ marginBottom: 14 }}>Überschreibt das globale Jahresziel für dieses Mitglied.</div>
      <div style={{ display: 'flex', justifyContent: 'center', padding: '8px 0 4px' }}>
        <Stepper value={goal} set={setGoal} min={0} max={80} label="h / Jahr" />
      </div>
    </Sheet>
  );
}

/** Draft tier row used in the editor (cents shown to the user as EUR). */
interface TierDraft {
  amountEur: string;
}

function FeeTierSheet({ memberId, clubYearId, onClose }: { memberId: string; clubYearId: string; onClose: () => void }) {
  const { showToast } = useAppStore();
  const tiersQ = useMemberFeeTiers(memberId, clubYearId);
  const update = useUpdateMemberFeeTiers();
  const [rows, setRows] = useState<TierDraft[] | null>(null);

  // Hydrate the editable rows from the loaded overrides once.
  const draft: TierDraft[] = rows ?? (tiersQ.data ?? [])
    .slice()
    .sort((a, b) => a.position - b.position)
    .map((t) => ({ amountEur: (t.amountCents / 100).toString().replace('.', ',') }));

  const setRow = (i: number, v: string) => {
    const next = draft.map((r, j) => (j === i ? { amountEur: v } : r));
    setRows(next);
  };
  const addRow = () => setRows([...draft, { amountEur: '' }]);
  const removeRow = (i: number) => setRows(draft.filter((_, j) => j !== i));

  const save = () => {
    // Empty list = inherit the club-year fee schedule.
    const tiers: MemberFeeTier[] = draft
      .map((r) => parseFloat(r.amountEur.replace(',', '.')))
      .filter((eur) => !Number.isNaN(eur))
      .map((eur, i) => ({
        id: '00000000-0000-0000-0000-000000000000',
        clubYearId,
        position: i + 1,
        amountCents: Math.round(eur * 100),
      }));
    update.mutate(
      { memberId, clubYearId, tiers },
      {
        onSuccess: () => { showToast('Abgeltungsbeträge gespeichert.'); onClose(); },
        onError: () => showToast('Beträge konnten nicht gespeichert werden.', 'crit'),
      },
    );
  };

  return (
    <Sheet
      onClose={onClose}
      title="Individuelle Abgeltung"
      foot={<Button icon="check" loading={update.isPending} onClick={save}>Speichern</Button>}
    >
      <div className="sm-hint" style={{ marginBottom: 14 }}>
        Überschreibt die Vereins-Abgeltungsliste für dieses Mitglied. Leer lassen, um die Standardliste zu übernehmen.
      </div>
      {tiersQ.isLoading ? (
        <LoadingState />
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
          {draft.length === 0 && (
            <div style={{ color: 'var(--muted)', fontWeight: 600, fontSize: 13.5 }}>Keine eigenen Beträge — Vereinsliste gilt.</div>
          )}
          {draft.map((r, i) => (
            <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <span style={{ width: 78, fontSize: 13, fontWeight: 700, color: 'var(--ink-2)' }}>{i + 1}. Stunde</span>
              <div style={{ flex: 1, position: 'relative' }}>
                <Input
                  type="number"
                  inputMode="decimal"
                  value={r.amountEur}
                  onChange={(e) => setRow(i, e.target.value)}
                  style={{ paddingRight: 28 }}
                />
                <span style={{ position: 'absolute', right: 12, top: 13, color: 'var(--muted)', fontWeight: 700 }}>€</span>
              </div>
              <button
                onClick={() => removeRow(i)}
                className="pressable"
                aria-label="Eintrag entfernen"
                style={{ width: 34, height: 34, borderRadius: '50%', border: 'none', background: 'var(--surface-2)', cursor: 'pointer', display: 'flex', alignItems: 'center', justifyContent: 'center' }}
              >
                <Icon name="x" size={16} stroke={2.4} color="var(--ink-2)" />
              </button>
            </div>
          ))}
          <Button variant="soft" size="sm" icon="plus" onClick={addRow} style={{ marginTop: 4 }}>Stufe hinzufügen</Button>
        </div>
      )}
    </Sheet>
  );
}

function EditSheet({ member, onClose }: { member: Member; onClose: () => void }) {
  const { showToast } = useAppStore();
  const update = useUpdateMember();
  const deactivate = useDeactivateMember();
  const [first, setFirst] = useState(member.first);
  const [last, setLast] = useState(member.last);
  const [email, setEmail] = useState(member.email);
  const [since, setSince] = useState(member.since?.slice(0, 10) ?? '');
  const [confirmDeactivate, setConfirmDeactivate] = useState(false);

  const save = () => {
    if (!first.trim() || !last.trim() || !email.trim()) return;
    update.mutate(
      { id: member.id, first: first.trim(), last: last.trim(), email: email.trim(), since },
      {
        onSuccess: () => { showToast('Mitglied gespeichert.'); onClose(); },
        onError: () => showToast('Speichern fehlgeschlagen.', 'crit'),
      },
    );
  };

  const onDeactivate = () => {
    deactivate.mutate(member.id, {
      onSuccess: () => { showToast('Mitglied deaktiviert.'); onClose(); },
      onError: () => showToast('Deaktivieren fehlgeschlagen.', 'crit'),
    });
  };

  return (
    <Sheet
      onClose={onClose}
      title="Mitglied bearbeiten"
      foot={<Button icon="check" loading={update.isPending} onClick={save} disabled={!first.trim() || !last.trim() || !email.trim()}>Speichern</Button>}
    >
      <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
          <Field label="Vorname">
            <Input value={first} onChange={(e) => setFirst(e.target.value)} autoFocus />
          </Field>
          <Field label="Nachname">
            <Input value={last} onChange={(e) => setLast(e.target.value)} />
          </Field>
        </div>
        <Field label="E-Mail-Adresse">
          <Input type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
        </Field>
        <Field label="Eintrittsdatum">
          <Input type="date" value={since} onChange={(e) => setSince(e.target.value)} />
        </Field>
      </div>
      {member.active !== false && (
        <div style={{ marginTop: 20, paddingTop: 16, borderTop: '1px solid var(--line)' }}>
          {!confirmDeactivate ? (
            <button
              onClick={() => setConfirmDeactivate(true)}
              style={{ background: 'none', border: 'none', color: 'var(--crit)', fontWeight: 700, fontSize: 13.5, cursor: 'pointer', padding: 0 }}
            >
              Mitglied deaktivieren…
            </button>
          ) : (
            <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
              <span style={{ fontSize: 13.5, fontWeight: 600, color: 'var(--ink-2)' }}>Wirklich deaktivieren?</span>
              <Button size="sm" variant="ghost" onClick={() => setConfirmDeactivate(false)}>Abbrechen</Button>
              <Button size="sm" loading={deactivate.isPending} onClick={onDeactivate} style={{ background: 'var(--crit)', color: '#fff' }}>Ja, deaktivieren</Button>
            </div>
          )}
        </div>
      )}
    </Sheet>
  );
}

export function MemberDetail({ id }: { id: string }) {
  const { back, push, showToast } = useAppStore();
  const [goalOpen, setGoalOpen] = useState(false);
  const [feeOpen, setFeeOpen] = useState(false);
  const [editOpen, setEditOpen] = useState(false);
  const { data: stats } = useStats();
  const clubYearId = stats?.clubYearId ?? '';
  const activate = useUpdateMember();
  const memberQ = useMember(id);
  const hoursQ = useMemberHours(id);
  const { data: settings } = useSettings();

  const memberMap = memberQ.data ? { [id]: { first: memberQ.data.first, last: memberQ.data.last } } : {};

  if (memberQ.isLoading || hoursQ.isLoading) return <LoadingState />;
  if (memberQ.isError || !memberQ.data) return <ErrorState text="Dieses Mitglied wurde nicht gefunden." />;

  const m = memberQ.data;

  // Hour account + booking history come from the hours API (includes manual bookings).
  const h = hoursQ.data?.confirmed ?? 0;
  const recs: HourRec[] = (hoursQ.data?.entries ?? [])
    .map((e) => ({
      t: e.manual
        ? (e.desc || 'Manuelle Buchung')
        : [e.eventName, e.shiftName].filter(Boolean).join(' · ') || e.desc || 'Schicht',
      d: e.date ?? '',
      h: e.hours,
      manual: e.manual,
    }))
    .sort((a, b) => b.d.localeCompare(a.d));
  const goal = hoursQ.data?.goal ?? m.goal ?? settings?.yearGoal ?? 20;

  return (
    <div className="fade-in">
      <div className="sm-header detail" style={{ alignItems: 'center', gap: 12 }}>
        <button
          onClick={back}
          className="pressable"
          style={{ width: 40, height: 40, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
        >
          <Icon name="chevL" size={20} stroke={2.4} />
        </button>
        <div className="sm-title" style={{ fontSize: 22 }}>Mitglied</div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 4 }}>
        <div className="sm-card pad" style={{ display: 'flex', gap: 14, alignItems: 'center' }}>
          <Avatar memberId={id} members={memberMap} size={56} />
          <div style={{ flex: 1 }}>
            <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 19 }}>{m.first} {m.last}</div>
            <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{m.email}</div>
            <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>seit {fmtDate(m.since, 'short')}</div>
          </div>
          {m.active === false && <Badge kind="warn">Nicht freigeschaltet</Badge>}
          <button
            className="pressable"
            onClick={() => setEditOpen(true)}
            title="Mitglied bearbeiten"
            style={{ width: 34, height: 34, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface-2)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
          >
            <Icon name="edit" size={16} color="var(--ink-2)" />
          </button>
        </div>
        {m.active === false && (
          <div className="sm-card pad" style={{ marginTop: 12, display: 'flex', alignItems: 'center', gap: 12, background: 'var(--warn-bg)' }}>
            <Icon name="clock" size={20} color="var(--warn)" />
            <div style={{ flex: 1 }}>
              <div style={{ fontWeight: 800, fontSize: 14.5 }}>Konto wartet auf Freischaltung</div>
              <div style={{ color: 'var(--ink-2)', fontSize: 12.5, fontWeight: 600 }}>Per OIDC registriert, aber noch nicht aktiviert.</div>
            </div>
            <Button
              size="sm"
              icon="check"
              loading={activate.isPending}
              onClick={() =>
                activate.mutate(
                  { id, active: true },
                  {
                    onSuccess: () => showToast('Mitglied freigeschaltet.'),
                    onError: () => showToast('Freischaltung fehlgeschlagen.', 'crit'),
                  },
                )
              }
            >
              Freischalten
            </Button>
          </div>
        )}
        <div className="sm-card pad" style={{ marginTop: 12 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'baseline' }}>
            <span style={{ fontWeight: 700, color: 'var(--ink-2)', fontSize: 14 }}>Stundenkonto {settings?.clubYear ?? new Date().getFullYear()}</span>
            <span><b style={{ fontFamily: 'Bricolage Grotesque', fontSize: 19 }}>{hrs(h)}</b> <span style={{ color: 'var(--muted)', fontWeight: 700 }}>/ {goal} h</span></span>
          </div>
          <div style={{ marginTop: 10 }}>
            <div className="occ-track" style={{ height: 9 }}>
              <div className="occ-fill" style={{ width: Math.min(h / goal * 100, 100) + '%', background: h >= goal ? 'var(--ok)' : 'var(--primary)' }} />
            </div>
          </div>
          <div style={{ display: 'flex', gap: 9, marginTop: 13, flexWrap: 'wrap' }}>
            <Button variant="soft" size="sm" icon="plus" onClick={() => push('manual', { id })}>Stunden buchen</Button>
            <Button variant="soft" size="sm" icon="edit" onClick={() => setGoalOpen(true)}>Ziel anpassen</Button>
            <Button variant="soft" size="sm" icon="edit" disabled={!clubYearId} onClick={() => setFeeOpen(true)}>Abgeltung</Button>
          </div>
        </div>
        <Section title="Gebuchte Stunden" />
        <div className="sm-card" style={{ overflow: 'hidden' }}>
          {recs.length === 0 && <div style={{ padding: 16, color: 'var(--muted)', fontWeight: 600, fontSize: 13.5 }}>Noch keine bestätigten Stunden.</div>}
          {recs.map((r, i) => (
            <React.Fragment key={i}>
              {i > 0 && <hr className="sm-divider" />}
              <div style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '12px 14px' }}>
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: 700, fontSize: 14 }}>{r.t}</div>
                  <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{fmtDate(r.d, 'short')}{r.manual && ' · manuell'}</div>
                </div>
                {r.manual && <Badge kind="primary" dot={false}>manuell</Badge>}
                <span style={{ fontWeight: 800, fontFamily: 'Bricolage Grotesque' }}>{hrs(r.h)}</span>
              </div>
            </React.Fragment>
          ))}
        </div>
      </div>

      {goalOpen && <GoalSheet memberId={id} value={goal} onClose={() => setGoalOpen(false)} />}
      {feeOpen && clubYearId && <FeeTierSheet memberId={id} clubYearId={clubYearId} onClose={() => setFeeOpen(false)} />}
      {editOpen && <EditSheet member={m} onClose={() => setEditOpen(false)} />}
    </div>
  );
}
