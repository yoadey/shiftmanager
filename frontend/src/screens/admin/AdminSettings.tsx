import React, { useState, useEffect, useRef } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { Field, Input } from '@/components/forms/Field';
import { Stepper } from '@/components/ui/Stepper';
import { Toggle } from '@/components/ui/Toggle';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import { useSettings, useUpdateSettings, useBranding, useUploadLogo, useUpdateBranding, useBrandingHistory, useRollbackBranding } from '@/api/settings';
import { useClubYears, useCreateClubYear } from '@/api/billing';
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
  const { push, setNameMode, showToast, setRole } = useAppStore();
  const { user, logout } = useAuthStore();
  const isAdmin = user?.role === 'admin' || user?.role === 'vorstand';
  const settingsQ = useSettings();
  const { data: branding } = useBranding();
  const clubYearsQ = useClubYears();
  const createClubYear = useCreateClubYear();

  // Club year creation dialog
  const [yearDialogOpen, setYearDialogOpen] = useState(false);
  const currentYear = new Date().getFullYear().toString();
  const [newYearLabel, setNewYearLabel] = useState(currentYear);
  const [newYearStart, setNewYearStart] = useState(`${currentYear}-01-01`);
  const [newYearEnd, setNewYearEnd] = useState(`${currentYear}-12-31`);
  const [newYearGoal, setNewYearGoal] = useState(20);

  const openYearDialog = () => {
    const y = new Date().getFullYear().toString();
    setNewYearLabel(y);
    setNewYearStart(`${y}-01-01`);
    setNewYearEnd(`${y}-12-31`);
    setNewYearGoal(s.yearGoal ?? 20);
    setYearDialogOpen(true);
  };

  const onCreateYear = () => {
    createClubYear.mutate(
      {
        label: newYearLabel,
        startDate: new Date(newYearStart + 'T00:00:00Z').toISOString(),
        endDate: new Date(newYearEnd + 'T23:59:59Z').toISOString(),
        defaultTargetHours: newYearGoal,
        setActive: true,
      },
      {
        onSuccess: () => { showToast(`Vereinsjahr ${newYearLabel} angelegt.`); setYearDialogOpen(false); },
        onError: () => showToast('Vereinsjahr konnte nicht angelegt werden.', 'crit'),
      },
    );
  };
  const updateSettings = useUpdateSettings();
  const uploadLogo = useUploadLogo();
  const updateBranding = useUpdateBranding();
  const { data: brandingHistory } = useBrandingHistory();
  const rollbackBranding = useRollbackBranding();
  const fileRef = useRef<HTMLInputElement>(null);
  const remote = settingsQ.data;

  // Branding history modal
  const [historyOpen, setHistoryOpen] = useState(false);

  // Branding contrast warnings (B-003) surfaced after the last colour save.
  const [warnings, setWarnings] = useState<string[]>([]);
  const [primaryColor, setPrimaryColor] = useState('');
  useEffect(() => {
    if (branding?.primaryColor) setPrimaryColor(branding.primaryColor);
  }, [branding?.primaryColor]);

  const saveColor = (hex: string) => {
    setPrimaryColor(hex);
    updateBranding.mutate(
      { primaryColor: hex },
      {
        onSuccess: (res) => {
          setWarnings(res.warnings ?? []);
          showToast('Farbe gespeichert.');
        },
        onError: () => showToast('Farbe konnte nicht gespeichert werden.', 'crit'),
      },
    );
  };

  const [s, setS] = useState<AppSettings>(remote ?? DEMO_STATE.settings);

  // Hydrate the editable form once the real settings arrive.
  useEffect(() => {
    if (remote) setS(remote);
  }, [remote]);

  const setSetting = <K extends keyof AppSettings>(key: K, value: AppSettings[K]) => {
    const updated = { ...s, [key]: value };
    setS(updated);
    updateSettings.mutate(updated);
    // NM-005: keep the global name-display mode in sync so every screen updates immediately.
    if (key === 'nameMode') setNameMode(value as 'abbrev' | 'full');
  };

  const onPickLogo = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    uploadLogo.mutate(file, {
      onSuccess: () => showToast('Logo aktualisiert.'),
      onError: () => showToast('Logo konnte nicht hochgeladen werden.', 'crit'),
    });
    e.target.value = '';
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

        <Section title="Vereinsjahr" />
        <div className="sm-card pad">
          {clubYearsQ.isLoading && <div style={{ color: 'var(--muted)', fontWeight: 600, fontSize: 13.5 }}>Lade…</div>}
          {!clubYearsQ.isLoading && (clubYearsQ.data ?? []).length === 0 && (
            <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
              <Icon name="info" size={18} color="var(--warn)" />
              <div style={{ flex: 1, color: 'var(--ink-2)', fontWeight: 600, fontSize: 13.5 }}>Kein Vereinsjahr konfiguriert — Jahresabrechnung und Stundenziele sind erst nach dem Anlegen verfügbar.</div>
            </div>
          )}
          {(clubYearsQ.data ?? []).map((y) => (
            <div key={y.id} style={{ display: 'flex', alignItems: 'center', gap: 10, marginBottom: 4 }}>
              <span style={{ fontWeight: 700, fontSize: 14.5 }}>{y.label}</span>
              {y.isActive && <Badge kind="ok">Aktiv</Badge>}
              <span style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>
                {new Date(y.startDate).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' })} – {new Date(y.endDate).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' })}
              </span>
              <span style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>· Ziel: {y.defaultTargetHours} h</span>
            </div>
          ))}
          <button
            className="sm-btn soft"
            style={{ width: '100%', marginTop: (clubYearsQ.data ?? []).length > 0 ? 12 : 10 }}
            onClick={openYearDialog}
          >
            <Icon name="plus" size={17} stroke={2.2} />
            Neues Vereinsjahr anlegen
          </button>
        </div>

        {yearDialogOpen && (
          <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.45)', zIndex: 200, display: 'flex', alignItems: 'flex-end', justifyContent: 'center' }} onClick={() => setYearDialogOpen(false)}>
            <div style={{ background: 'var(--surface)', borderRadius: '20px 20px 0 0', width: '100%', maxWidth: 520, padding: 24, paddingBottom: 32 }} onClick={(e) => e.stopPropagation()}>
              <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 19, marginBottom: 16 }}>Vereinsjahr anlegen</div>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 10, marginBottom: 18 }}>
                <Field label="Bezeichnung (z. B. 2026)">
                  <Input value={newYearLabel} onChange={(e) => setNewYearLabel(e.target.value)} autoFocus />
                </Field>
                <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
                  <Field label="Beginn">
                    <Input type="date" value={newYearStart} onChange={(e) => setNewYearStart(e.target.value)} />
                  </Field>
                  <Field label="Ende">
                    <Input type="date" value={newYearEnd} onChange={(e) => setNewYearEnd(e.target.value)} />
                  </Field>
                </div>
                <Field label="Standard-Stundenziel">
                  <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                    <Input
                      type="number"
                      min={1}
                      max={100}
                      value={newYearGoal}
                      onChange={(e) => setNewYearGoal(+e.target.value)}
                      style={{ flex: 1 }}
                    />
                    <span style={{ fontWeight: 700, color: 'var(--muted)' }}>h</span>
                  </div>
                </Field>
              </div>
              <div className="sm-hint" style={{ marginBottom: 14 }}>Das neue Jahr wird sofort als aktives Vereinsjahr gesetzt.</div>
              <button
                className="sm-btn"
                onClick={onCreateYear}
                disabled={createClubYear.isPending || !newYearLabel.trim()}
                style={{ width: '100%', marginBottom: 10 }}
              >
                {createClubYear.isPending ? 'Wird angelegt…' : 'Anlegen'}
              </button>
              <button className="sm-btn ghost" onClick={() => setYearDialogOpen(false)} style={{ width: '100%' }}>
                Abbrechen
              </button>
            </div>
          </div>
        )}

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
          <hr className="sm-divider" style={{ margin: '14px 0' }} />
          <RowBetween label="Erinnerung Vorlaufzeit" sub="Erinnerungsmail wird N Wochen vor Beginn versendet">
            <Stepper value={s.reminderLeadWeeks ?? 2} set={(v) => setSetting('reminderLeadWeeks', v)} min={1} max={4} label="Wo" />
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
            {(s.feeSchedule ?? []).map((v, i) => (
              <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
                <span style={{ width: 78, fontSize: 13, fontWeight: 700, color: 'var(--ink-2)' }}>{i + 1}. Stunde</span>
                <div style={{ flex: 1, position: 'relative' }}>
                  <Input
                    type="number"
                    value={v}
                    onChange={(e) => {
                      const fee = [...(s.feeSchedule ?? [])];
                      fee[i] = +e.target.value;
                      setSetting('feeSchedule', fee);
                    }}
                    style={{ paddingRight: 28 }}
                  />
                  <span style={{ position: 'absolute', right: 12, top: 13, color: 'var(--muted)', fontWeight: 700 }}>€</span>
                </div>
                {i === (s.feeSchedule ?? []).length - 1 && <Badge kind="neutral" dot={false}>ab hier</Badge>}
              </div>
            ))}
          </div>
        </div>

        <Section title="Branding" />
        <div className="sm-card pad">
          <RowBetween label="Primärfarbe" sub="WCAG-Kontrast wird beim Speichern geprüft">
            <input
              type="color"
              value={primaryColor || '#000000'}
              onChange={(e) => saveColor(e.target.value)}
              aria-label="Primärfarbe"
              style={{ width: 36, height: 30, border: '1px solid var(--line-2)', borderRadius: 9, background: 'transparent', cursor: 'pointer', padding: 0 }}
            />
          </RowBetween>
          {warnings.length > 0 && (
            <div style={{ display: 'flex', gap: 8, marginTop: 12, background: 'var(--warn-bg, #FFF6E5)', borderRadius: 10, padding: '10px 12px', fontSize: 12.5, color: 'var(--ink-2)', fontWeight: 600, lineHeight: 1.4 }}>
              <Icon name="info" size={16} color="var(--warn, #C58A00)" style={{ flexShrink: 0, marginTop: 1 }} />
              <div>
                {warnings.map((w, i) => <div key={i}>Kontrast unter 4,5:1 — {w}</div>)}
              </div>
            </div>
          )}
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
          <input
            ref={fileRef}
            type="file"
            accept="image/png,image/svg+xml,.png,.svg"
            onChange={onPickLogo}
            style={{ display: 'none' }}
          />
          <div style={{ display: 'flex', gap: 8, marginTop: 12 }}>
            <button
              className="sm-btn soft sm"
              onClick={() => fileRef.current?.click()}
              disabled={uploadLogo.isPending}
              style={{ flex: 1 }}
            >
              <Icon name="download" size={17} stroke={2.2} />
              {uploadLogo.isPending ? 'Wird hochgeladen…' : 'Logo hochladen (PNG/SVG)'}
            </button>
            <button
              className="sm-btn soft sm"
              onClick={() => setHistoryOpen(true)}
              style={{ flexShrink: 0 }}
            >
              Verlauf
            </button>
          </div>
        </div>

        {/* Branding History Modal */}
        {historyOpen && (
          <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.45)', zIndex: 200, display: 'flex', alignItems: 'flex-end', justifyContent: 'center' }} onClick={() => setHistoryOpen(false)}>
            <div style={{ background: 'var(--surface)', borderRadius: '20px 20px 0 0', width: '100%', maxWidth: 520, padding: 24, paddingBottom: 32, maxHeight: '70vh', overflowY: 'auto' }} onClick={(e) => e.stopPropagation()}>
              <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 19, marginBottom: 16 }}>Branding-Verlauf</div>
              {(!brandingHistory || brandingHistory.length === 0) ? (
                <div style={{ color: 'var(--muted)', fontSize: 13.5, fontWeight: 600 }}>Noch keine gespeicherten Snapshots.</div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
                  {brandingHistory.map((snap) => (
                    <div key={snap.id} style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '12px 14px', background: 'var(--surface-2)', borderRadius: 12 }}>
                      <div>
                        <div style={{ fontWeight: 700, fontSize: 13.5 }}>{new Date(snap.createdAt).toLocaleString('de-DE', { dateStyle: 'medium', timeStyle: 'short' })}</div>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginTop: 4 }}>
                          <div style={{ width: 16, height: 16, borderRadius: '50%', background: snap.branding.primaryColor, border: '1px solid var(--line)' }} />
                          <span style={{ fontSize: 12, fontWeight: 600, color: 'var(--muted)' }}>{snap.branding.primaryColor}</span>
                        </div>
                      </div>
                      <button
                        className="sm-btn soft sm"
                        disabled={rollbackBranding.isPending}
                        onClick={() => rollbackBranding.mutate(snap.id, {
                          onSuccess: () => {
                            showToast('Branding wiederhergestellt.');
                            setHistoryOpen(false);
                          },
                          onError: () => showToast('Wiederherstellung fehlgeschlagen.', 'crit'),
                        })}
                      >
                        Wiederherstellen
                      </button>
                    </div>
                  ))}
                </div>
              )}
              <button className="sm-btn ghost" onClick={() => setHistoryOpen(false)} style={{ width: '100%', marginTop: 16 }}>
                Schließen
              </button>
            </div>
          </div>
        )}

        <Section title="Kiosk" />
        <div className="sm-card pad">
          <RowBetween label="Kiosk sperren" sub="Deaktiviert die öffentliche Kiosk-Eintragung (K-012)">
            <Toggle on={!!s.kioskLocked} onClick={() => setSetting('kioskLocked', !s.kioskLocked)} />
          </RowBetween>
        </div>

        {isAdmin && (
          <>
            <Section title="E-Mail" />
            <div className="sm-card">
              <LinkRow icon="mail" label="E-Mail-Vorlagen" sub="Betreff & Text bearbeiten" onClick={() => push('email-templates')} />
              <hr className="sm-divider" />
              <LinkRow icon="mail" label="E-Mail-Protokoll" sub="Versand prüfen & erneut senden" onClick={() => push('email-log')} />
            </div>
          </>
        )}

        <Section title="System" />
        <div className="sm-card">
          <LinkRow icon="shield" label="Audit-Log" sub="Protokoll aller Änderungen" onClick={() => push('audit')} />
          <hr className="sm-divider" />
          <LinkRow icon="download" label="Datenbank-Export" sub="CSV / Backup" />
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
    </div>
  );
}

export { RowBetween, SegRadio };
