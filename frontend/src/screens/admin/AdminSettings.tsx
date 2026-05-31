import React, { useState, useEffect } from 'react';
import { Badge } from '@/components/ui/Badge';
import { Input } from '@/components/forms/Field';
import { Stepper } from '@/components/ui/Stepper';
import { Toggle } from '@/components/ui/Toggle';
import { useAppStore } from '@/store/app.store';
import { useSettings, useUpdateSettings, useBranding } from '@/api/settings';
import { LoadingState, ErrorState } from '@/components/ui/States';
import { DEMO_STATE } from '@/screens/_demo';
import { Section } from '@/screens/member/MemberDashboard';
import { LinkRow } from '@/screens/member/MemberProfile';
import type { AppSettings } from '@/types';

function RowBetween({ label, sub, children }: { label: string; sub?: string; children: React.ReactNode }) {
  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', gap: 12 }}>
      <div style={{ flex: 1 }}>
        <div style={{ fontWeight: 700, fontSize: 14.5 }}>{label}</div>
        {sub && <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>{sub}</div>}
      </div>
      {children}
    </div>
  );
}

function SegRadio<T extends string>({ value, onChange, options }: { value: T; onChange: (v: T) => void; options: [T, string][] }) {
  return (
    <div style={{ display: 'flex', gap: 6, background: 'var(--surface-2)', borderRadius: 12, padding: 4, border: '1px solid var(--line)' }}>
      {options.map(([v, l]) => (
        <button
          key={v}
          onClick={() => onChange(v)}
          style={{ flex: 1, border: 'none', cursor: 'pointer', padding: '9px 0', borderRadius: 9, fontWeight: 700, fontSize: 13.5, fontFamily: 'inherit', background: value === v ? 'var(--surface)' : 'transparent', color: value === v ? 'var(--ink)' : 'var(--muted)', boxShadow: value === v ? 'var(--shadow)' : 'none' }}
        >
          {l}
        </button>
      ))}
    </div>
  );
}

export function AdminSettings() {
  const { push, setNameMode } = useAppStore();
  const settingsQ = useSettings();
  const { data: branding } = useBranding();
  const updateSettings = useUpdateSettings();
  const remote = settingsQ.data;

  const [s, setS] = useState<AppSettings>(remote ?? DEMO_STATE.settings);

  // Hydrate the editable form once the real settings arrive.
  useEffect(() => {
    if (remote) setS(remote);
  }, [remote]);

  const setSetting = <K extends keyof AppSettings>(key: K, value: AppSettings[K]) => {
    setS((prev) => ({ ...prev, [key]: value }));
    updateSettings.mutate({ [key]: value } as Partial<AppSettings>);
    // NM-005: keep the global name-display mode in sync so every screen updates immediately.
    if (key === 'nameMode') setNameMode(value as 'abbrev' | 'full');
  };

  if (settingsQ.isLoading) return <LoadingState />;
  if (settingsQ.isError && !remote) return <ErrorState />;

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div>
          <div className="sm-eyebrow">{branding?.clubName ?? s.clubName}</div>
          <div className="sm-title">Einstellungen</div>
        </div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 6 }}>
        <Section title="Stundenziel" />
        <div className="sm-card pad">
          <RowBetween label="Globales Jahresziel" sub="Gilt, sofern kein individuelles Ziel gesetzt ist">
            <Stepper value={s.yearGoal} set={(v) => setSetting('yearGoal', v)} min={1} max={80} label="h" />
          </RowBetween>
        </div>

        <Section title="Datenschutz" />
        <div className="sm-card pad">
          <div style={{ fontWeight: 700, fontSize: 14.5, marginBottom: 9 }}>Namensanzeige</div>
          <SegRadio
            value={s.nameMode}
            onChange={(v) => setSetting('nameMode', v)}
            options={[['abbrev', 'Abgekürzt'], ['full', 'Vollständig']]}
          />
          <div className="sm-hint">{s.nameMode === 'abbrev' ? '„Maximilian M.“ – Nachname gekürzt. Vorstand sieht immer den vollen Namen.' : '„Maximilian Müller“ – voller Name für alle Rollen.'}</div>
          <hr className="sm-divider" style={{ margin: '14px 0' }} />
          <RowBetween label="Mitgliedersuche im Kiosk" sub="Sonst nur Eintragung per E-Mail">
            <Toggle on={s.kioskSearch} onClick={() => setSetting('kioskSearch', !s.kioskSearch)} />
          </RowBetween>
        </div>

        <Section title="Reservierungen" />
        <div className="sm-card pad">
          <RowBetween label="Reservierung gültig für" sub="Frist für E-Mail-Bestätigung">
            <Stepper value={s.reservationHours} set={(v) => setSetting('reservationHours', v)} min={1} max={168} label="h" />
          </RowBetween>
          <hr className="sm-divider" style={{ margin: '14px 0' }} />
          <RowBetween label="Abmeldefrist vor Beginn">
            <Stepper value={s.deregisterDeadlineH} set={(v) => setSetting('deregisterDeadlineH', v)} min={0} max={72} label="h" />
          </RowBetween>
        </div>

        <Section title="Abrechnung" />
        <div className="sm-card pad">
          <div style={{ fontWeight: 700, fontSize: 14.5, marginBottom: 9 }}>Abrechnungsmodus</div>
          <SegRadio
            value={s.billingMode}
            onChange={(v) => setSetting('billingMode', v)}
            options={[['manuell', 'Manuell'], ['auto', 'Automatisch']]}
          />
          <hr className="sm-divider" style={{ margin: '14px 0' }} />
          <div style={{ fontWeight: 700, fontSize: 14.5, marginBottom: 4 }}>Abgeltungsbeträge je Fehlstunde</div>
          <div className="sm-hint" style={{ marginBottom: 10 }}>Letzter Wert gilt für alle weiteren Fehlstunden.</div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {s.feeSchedule.map((v, i) => (
              <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <span style={{ width: 78, fontSize: 13, fontWeight: 700, color: 'var(--ink-2)' }}>{i + 1}. Stunde</span>
                <div style={{ flex: 1, position: 'relative' }}>
                  <Input
                    type="number"
                    value={v}
                    onChange={(e) => {
                      const fee = [...s.feeSchedule];
                      fee[i] = +e.target.value;
                      setSetting('feeSchedule', fee);
                    }}
                    style={{ paddingRight: 28 }}
                  />
                  <span style={{ position: 'absolute', right: 12, top: 13, color: 'var(--muted)', fontWeight: 700 }}>€</span>
                </div>
                {i === s.feeSchedule.length - 1 && <Badge kind="neutral" dot={false}>ab hier</Badge>}
              </div>
            ))}
          </div>
        </div>

        <Section title="Branding" />
        <div className="sm-card pad">
          <RowBetween label="Primärfarbe" sub="Über das Tweaks-Panel anpassbar">
            <div style={{ width: 30, height: 30, borderRadius: 9, background: 'var(--primary)', border: '1px solid var(--line-2)' }} />
          </RowBetween>
          <hr className="sm-divider" style={{ margin: '14px 0' }} />
          <RowBetween label="Vereinslogo" sub="PNG/SVG, min. 200×200 px">
            {branding?.logoUrl ? (
              <img src={branding.logoUrl} alt="Logo" style={{ width: 30, height: 30, borderRadius: '50%', objectFit: 'cover' }} />
            ) : (
              <div style={{ width: 30, height: 30, borderRadius: '50%', background: 'var(--ink)', color: 'var(--primary)', display: 'flex', alignItems: 'center', justifyContent: 'center', fontWeight: 800, fontSize: 12, fontFamily: 'Bricolage Grotesque' }}>
                {(branding?.clubName ?? s.clubName).split(/\s+/).filter(Boolean).map((w) => w[0]).slice(0, 2).join('').toUpperCase() || 'SG'}
              </div>
            )}
          </RowBetween>
        </div>

        <Section title="System" />
        <div className="sm-card">
          <LinkRow icon="shield" label="Audit-Log" sub="Protokoll aller Änderungen" onClick={() => push('audit')} />
          <hr className="sm-divider" />
          <LinkRow icon="download" label="Datenbank-Export" sub="CSV / Backup" />
        </div>
      </div>
    </div>
  );
}

export { RowBetween, SegRadio };
