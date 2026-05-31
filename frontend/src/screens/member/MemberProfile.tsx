import { useEffect, useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Avatar } from '@/components/ui/Avatar';
import { Button } from '@/components/ui/Button';
import { Toggle } from '@/components/ui/Toggle';
import { Sheet } from '@/components/ui/Sheet';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import { useMember, useUpdatePreferences, useExportMemberData } from '@/api/members';
import { fmtDate } from '@/screens/_demo';
import { Section } from '@/screens/member/MemberDashboard';
import type { Member } from '@/types';

function PrefRow({ label, sub, on, set }: { label: string; sub: string; on: boolean; set: () => void }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '13px 15px' }}>
      <div style={{ flex: 1 }}>
        <div style={{ fontWeight: 700, fontSize: 14.5 }}>{label}</div>
        <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{sub}</div>
      </div>
      <Toggle on={on} onClick={set} />
    </div>
  );
}

function LinkRow({ icon, label, sub, onClick }: { icon: string; label: string; sub?: string; onClick?: () => void }) {
  return (
    <div
      className="pressable"
      onClick={onClick}
      style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '13px 15px', cursor: 'pointer' }}
    >
      <div style={{ width: 36, height: 36, borderRadius: 10, background: 'var(--surface-2)', display: 'flex', alignItems: 'center', justifyContent: 'center', flexShrink: 0 }}>
        <Icon name={icon} size={18} color="var(--ink-2)" />
      </div>
      <div style={{ flex: 1 }}>
        <div style={{ fontWeight: 700, fontSize: 14.5 }}>{label}</div>
        {sub && <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{sub}</div>}
      </div>
      <Icon name="chevR" size={18} color="var(--line-2)" stroke={2.4} />
    </div>
  );
}

interface Prefs {
  week: boolean;
  day: boolean;
  news: boolean;
}

export function MemberProfile() {
  const { setRole, showToast } = useAppStore();
  const { user, logout } = useAuthStore();
  const uid = user?.id ?? '';
  const { data: profile } = useMember(uid);
  const updatePrefs = useUpdatePreferences();
  const exportData = useExportMemberData();
  const [deleteOpen, setDeleteOpen] = useState(false);

  // Prefer the live profile; fall back to the auth user while it loads.
  const me: Member = profile ?? {
    id: uid,
    first: user?.first ?? user?.name?.split(' ')[0] ?? '',
    last: user?.last ?? user?.name?.split(' ')[1] ?? '',
    email: user?.email ?? '',
    since: '',
    goal: null,
  };
  const memberMap = { [me.id]: { first: me.first, last: me.last } };
  const [prefs, setPrefs] = useState<Prefs>({ week: true, day: true, news: false });

  // N-001: reminder opt-out is persisted; the local toggle reflects the inverse
  // ("Erinnerungen aktiv" = !reminderOptOut). Hydrate from the live profile.
  const [remindersOn, setRemindersOn] = useState(true);
  useEffect(() => {
    if (profile) setRemindersOn(!profile.reminderOptOut);
  }, [profile]);

  const toggleReminders = () => {
    const next = !remindersOn;
    setRemindersOn(next);
    // optOut is the inverse of "reminders on".
    updatePrefs.mutate(!next, {
      onSuccess: () => showToast(next ? 'Erinnerungen aktiviert.' : 'Erinnerungen deaktiviert.'),
      onError: () => { setRemindersOn(!next); showToast('Einstellung konnte nicht gespeichert werden.', 'crit'); },
    });
  };

  const onExport = () => {
    exportData.mutate(uid, {
      onSuccess: () => showToast('Datenexport heruntergeladen.'),
      onError: () => showToast('Export fehlgeschlagen.', 'crit'),
    });
  };

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">Konto</div>
          <div className="sm-title">Profil</div>
        </div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 8 }}>
        <div className="sm-card pad" style={{ display: 'flex', alignItems: 'center', gap: 14 }}>
          <Avatar memberId={me.id} members={memberMap} size={56} />
          <div>
            <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 19 }}>{me.first} {me.last}</div>
            <div style={{ color: 'var(--muted)', fontSize: 13, fontWeight: 600 }}>{me.email}</div>
            {me.since && <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600, marginTop: 2 }}>Mitglied seit {fmtDate(me.since, 'short')}</div>}
          </div>
        </div>

        <Section title="Erinnerungen" />
        <div className="sm-card">
          <PrefRow label="Schicht-Erinnerungen" sub="Erinnerungs-Mails vor Schichtbeginn" on={remindersOn} set={toggleReminders} />
          <hr className="sm-divider" />
          <PrefRow label="Erinnerung 1 Tag vorher" sub="24 h vor Schichtbeginn" on={prefs.day} set={() => setPrefs((p) => ({ ...p, day: !p.day }))} />
          <hr className="sm-divider" />
          <PrefRow label="Neue Veranstaltungen" sub="Bei Veröffentlichung benachrichtigen" on={prefs.news} set={() => setPrefs((p) => ({ ...p, news: !p.news }))} />
        </div>
        <div className="sm-hint" style={{ padding: '0 4px' }}>Pflicht-Mails (z. B. Absage einer Schicht) können nicht deaktiviert werden.</div>

        <Section title="Datenschutz" />
        <div className="sm-card">
          <LinkRow icon="download" label="Meine Daten exportieren" sub="Auskunftsrecht (DSGVO)" onClick={onExport} />
          <hr className="sm-divider" />
          <LinkRow icon="shield" label="Löschung beantragen" sub="Recht auf Vergessen" onClick={() => setDeleteOpen(true)} />
        </div>

        <Section title="Demo" />
        <div className="sm-card pad" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div>
            <div style={{ fontWeight: 700, fontSize: 14.5 }}>Vorstands-Ansicht testen</div>
            <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>Wechselt in die Admin-Oberfläche</div>
          </div>
          <Button variant="dark" size="sm" icon="arrowR" onClick={() => setRole('vorstand')}>Wechseln</Button>
        </div>

        <Section title="Konto" />
        <div className="sm-card">
          <LinkRow
            icon="lock"
            label="Abmelden"
            sub="Sitzung beenden"
            onClick={() => {
              logout();
              setRole('mitglied');
            }}
          />
        </div>
      </div>

      {deleteOpen && (
        <Sheet
          onClose={() => setDeleteOpen(false)}
          variant="dialog"
          title="Löschung beantragen"
          foot={
            <Button
              icon="mail"
              onClick={() => {
                // Self-service deletion is not permitted: the gdpr-delete route is
                // Vorstand+ only. Members request anonymization from the board.
                showToast('Antrag an den Vorstand gesendet.', 'warn');
                setDeleteOpen(false);
              }}
            >
              Antrag senden
            </Button>
          }
        >
          <div style={{ fontSize: 14, fontWeight: 600, color: 'var(--ink-2)', lineHeight: 1.5 }}>
            Aus rechtlichen Gründen (Aufbewahrungspflicht für Abrechnungs- und Audit-Daten)
            kann dein Konto nur durch den Vorstand anonymisiert werden. Wir leiten deinen
            Antrag an den Vorstand weiter; dieser bestätigt die Anonymisierung deiner
            persönlichen Daten.
          </div>
        </Sheet>
      )}
    </div>
  );
}

export { PrefRow, LinkRow };
