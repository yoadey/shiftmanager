import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { useAppStore } from '@/store/app.store';
import { routes } from '@/routes';
import { useSmartBack } from '@/hooks/useSmartBack';
import {
  useClubYears,
  useCreateClubYear,
  useUpdateClubYear,
  useDeleteClubYear,
  useBilling,
  useExportBillingCSV,
  useExportBillingPDF,
  useClubYearFeeTiers,
  useUpdateClubYearFeeTiers,
} from '@/api/billing';
import { useSettings } from '@/api/settings';
import { LoadingState, ErrorState } from '@/components/ui/States';
import type { BillingResult } from '@/api/billing';
import type { ClubYear, FeeTier } from '@/types';

// ── Formatters ──────────────────────────────────────────────────────────────────

function euro(cents: number): string {
  return (cents / 100).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' });
}
function hrsLabel(h: number): string {
  return h % 1 === 0 ? `${h} Std.` : `${h.toFixed(1)} Std.`;
}
function toDateInputValue(iso: string): string {
  return iso ? iso.slice(0, 10) : '';
}
function fmtDate(iso: string): string {
  return new Date(iso).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' });
}

// ── Member row ─────────────────────────────────────────────────────────────────

function MemberRow({ r }: { r: BillingResult }) {
  const missing = r.missingHours > 0;
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px' }}>
      <div style={{ width: 38, height: 38, borderRadius: '50%', flexShrink: 0, background: missing ? 'var(--crit-bg, #fff1f0)' : 'var(--ok-bg, #f0faf4)', display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
        <Icon name={missing ? 'alert' : 'check'} size={18} color={missing ? 'var(--crit)' : 'var(--ok)'} />
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontWeight: 700, fontSize: 14.5 }}>{r.member.firstName} {r.member.lastName}</div>
        <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>
          {hrsLabel(r.confirmedHours)} / {hrsLabel(r.targetHours)} Ziel
          {missing && <span style={{ color: 'var(--crit)' }}> · {hrsLabel(r.missingHours)} fehlend</span>}
        </div>
      </div>
      {missing
        ? r.totalCents > 0
          ? <Badge kind="crit">{euro(r.totalCents)}</Badge>
          : <Badge kind="warn">kein Betrag</Badge>
        : <Badge kind="ok">Erfüllt</Badge>}
    </div>
  );
}

// ── Bottom-sheet modal wrapper ─────────────────────────────────────────────────

function BottomSheet({ onClose, children }: { onClose: () => void; children: React.ReactNode }) {
  return (
    <div
      style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.45)', zIndex: 200, display: 'flex', alignItems: 'flex-end', justifyContent: 'center' }}
      onClick={onClose}
    >
      <div
        className="fade-in"
        style={{ background: 'var(--surface)', borderRadius: '20px 20px 0 0', width: '100%', maxWidth: 520, padding: 24, paddingBottom: 32, maxHeight: '90vh', overflowY: 'auto' }}
        onClick={(e) => e.stopPropagation()}
      >
        {children}
      </div>
    </div>
  );
}

// ── Reusable fee tier rows (used inside create modal) ──────────────────────────

interface TierRow { amountEuros: string }

function FeeTierRows({ rows, onChange }: { rows: TierRow[]; onChange: (rows: TierRow[]) => void }) {
  const set = (i: number, v: string) => onChange(rows.map((r, j) => j === i ? { amountEuros: v } : r));
  const add = () => {
    const last = rows.length > 0 ? rows[rows.length - 1].amountEuros : '0.00';
    onChange([...rows, { amountEuros: last }]);
  };
  const remove = (i: number) => onChange(rows.filter((_, j) => j !== i));

  return (
    <>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 8, marginBottom: 10 }}>
        {rows.map((row, i) => (
          <div key={i} style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
            <span style={{ width: 80, fontSize: 13, fontWeight: 700, color: 'var(--ink-2)', flexShrink: 0 }}>
              {i + 1}. Stunde
            </span>
            <div style={{ flex: 1, position: 'relative' }}>
              <input
                type="number" min={0} step={0.01} value={row.amountEuros}
                onChange={(e) => set(i, e.target.value)}
                style={{ width: '100%', padding: '8px 28px 8px 10px', borderRadius: 8, border: '1.5px solid var(--line)', background: 'var(--surface)', fontSize: 14, fontWeight: 600, color: 'var(--ink)', boxSizing: 'border-box' as const }}
              />
              <span style={{ position: 'absolute', right: 10, top: '50%', transform: 'translateY(-50%)', color: 'var(--muted)', fontWeight: 700, fontSize: 13 }}>€</span>
            </div>
            {i === rows.length - 1 && rows.length > 1 && <Badge kind="neutral" dot={false}>ab hier</Badge>}
            {rows.length > 1 && (
              <button className="pressable" onClick={() => remove(i)}
                style={{ flexShrink: 0, width: 28, height: 28, borderRadius: 8, border: '1.5px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', color: 'var(--crit)' }}>
                <Icon name="x" size={13} color="currentColor" />
              </button>
            )}
          </div>
        ))}
      </div>
      <button className="sm-btn soft" style={{ width: '100%' }} onClick={add}>
        <Icon name="plus" size={16} stroke={2.2} /> Stufe hinzufügen
      </button>
    </>
  );
}

function toTierRows(schedule: number[]): TierRow[] {
  return schedule.map((v) => ({ amountEuros: v.toFixed(2) }));
}
function tierRowsToCents(rows: TierRow[]): { position: number; amountCents: number }[] {
  return rows.map((r, i) => ({
    position: i + 1,
    amountCents: Math.round(parseFloat(r.amountEuros.replace(',', '.') || '0') * 100),
  }));
}

// ── Create year modal ──────────────────────────────────────────────────────────

interface CreateYearModalProps {
  defaultTargetHours: number;
  globalFeeSchedule: number[];
  onClose: () => void;
  onSave: (label: string, startDate: string, endDate: string, targetHours: number, tiers: { position: number; amountCents: number }[]) => void;
  isSaving: boolean;
}

function CreateYearModal({ defaultTargetHours, globalFeeSchedule, onClose, onSave, isSaving }: CreateYearModalProps) {
  const curYear = new Date().getFullYear().toString();
  const [label, setLabel] = useState(curYear);
  const [startDate, setStartDate] = useState(`${curYear}-01-01`);
  const [endDate, setEndDate] = useState(`${curYear}-12-31`);
  const [targetHours, setTargetHours] = useState(defaultTargetHours);
  const [tierRows, setTierRows] = useState<TierRow[]>(
    globalFeeSchedule.length > 0 ? toTierRows(globalFeeSchedule) : [{ amountEuros: '0.00' }],
  );

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!label.trim() || !startDate || !endDate) return;
    onSave(label.trim(), startDate, endDate, targetHours, tierRowsToCents(tierRows));
  };

  return (
    <BottomSheet onClose={onClose}>
      <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 19, marginBottom: 16 }}>
        Abrechnungsjahr anlegen
      </div>
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div>
          <label className="sm-field-label">Bezeichnung</label>
          <input type="text" value={label} onChange={(e) => setLabel(e.target.value)} required autoFocus className="sm-input" style={{ width: '100%', boxSizing: 'border-box' as const }} />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
          <div>
            <label className="sm-field-label">Beginn</label>
            <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} required className="sm-input" style={{ width: '100%', boxSizing: 'border-box' as const }} />
          </div>
          <div>
            <label className="sm-field-label">Ende</label>
            <input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} required className="sm-input" style={{ width: '100%', boxSizing: 'border-box' as const }} />
          </div>
        </div>
        <div>
          <label className="sm-field-label">Standard-Stundenziel</label>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <input type="number" min={1} max={100} value={targetHours} onChange={(e) => setTargetHours(+e.target.value)} className="sm-input" style={{ flex: 1 }} />
            <span style={{ fontWeight: 700, color: 'var(--muted)' }}>h</span>
          </div>
        </div>

        <hr className="sm-divider" style={{ margin: '4px 0' }} />

        <div>
          <div style={{ fontWeight: 700, fontSize: 14, marginBottom: 4 }}>Abgeltungsbeträge je Fehlstunde</div>
          <div className="sm-hint" style={{ marginBottom: 10 }}>Letzter Wert gilt für alle weiteren Fehlstunden.</div>
          <FeeTierRows rows={tierRows} onChange={setTierRows} />
        </div>

        <div className="sm-hint" style={{ marginTop: 4 }}>
          Das neue Jahr wird sofort als aktives Abrechnungsjahr gesetzt.
        </div>

        <button type="submit" className="sm-btn" disabled={isSaving || !label.trim() || !startDate || !endDate} style={{ width: '100%', marginTop: 4 }}>
          {isSaving ? 'Wird angelegt…' : 'Anlegen'}
        </button>
        <button type="button" className="sm-btn ghost" onClick={onClose} style={{ width: '100%' }}>
          Abbrechen
        </button>
      </form>
    </BottomSheet>
  );
}

// ── Edit year modal ────────────────────────────────────────────────────────────

interface EditYearModalProps {
  year: ClubYear;
  onClose: () => void;
  onSave: (label: string, startDate: string, endDate: string, targetHours: number, isActive: boolean) => void;
  isSaving: boolean;
}

function EditYearModal({ year, onClose, onSave, isSaving }: EditYearModalProps) {
  const [label, setLabel] = useState(year.label);
  const [startDate, setStartDate] = useState(toDateInputValue(year.startDate));
  const [endDate, setEndDate] = useState(toDateInputValue(year.endDate));
  const [targetHours, setTargetHours] = useState(year.defaultTargetHours);
  const [isActive, setIsActive] = useState(year.isActive);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!label.trim() || !startDate || !endDate) return;
    onSave(label.trim(), startDate, endDate, targetHours, isActive);
  };

  return (
    <BottomSheet onClose={onClose}>
      <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 19, marginBottom: 16 }}>
        Abrechnungsjahr bearbeiten
      </div>
      <form onSubmit={handleSubmit} style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        <div>
          <label className="sm-field-label">Bezeichnung</label>
          <input type="text" value={label} onChange={(e) => setLabel(e.target.value)} required autoFocus className="sm-input" style={{ width: '100%', boxSizing: 'border-box' as const }} />
        </div>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 10 }}>
          <div>
            <label className="sm-field-label">Beginn</label>
            <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} required className="sm-input" style={{ width: '100%', boxSizing: 'border-box' as const }} />
          </div>
          <div>
            <label className="sm-field-label">Ende</label>
            <input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} required className="sm-input" style={{ width: '100%', boxSizing: 'border-box' as const }} />
          </div>
        </div>
        <div>
          <label className="sm-field-label">Standard-Stundenziel</label>
          <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
            <input type="number" min={1} max={100} value={targetHours} onChange={(e) => setTargetHours(+e.target.value)} className="sm-input" style={{ flex: 1 }} />
            <span style={{ fontWeight: 700, color: 'var(--muted)' }}>h</span>
          </div>
        </div>
        <div className="sm-hint">Ändert nur den Standardwert für neue Mitglieder — individuelle Ziele bleiben unberührt.</div>
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '10px 14px', borderRadius: 12, border: '1px solid var(--line)', background: 'var(--surface)' }}>
          <div>
            <div style={{ fontWeight: 700, fontSize: 14 }}>Aktives Jahr</div>
            <div style={{ fontSize: 12, color: 'var(--muted)', fontWeight: 600 }}>Wird für Stundenbuchungen und Abrechnung verwendet</div>
          </div>
          <button
            type="button"
            onClick={() => setIsActive((v) => !v)}
            style={{ width: 44, height: 26, borderRadius: 13, border: 'none', cursor: 'pointer', background: isActive ? 'var(--primary)' : 'var(--line)', transition: 'background .2s', position: 'relative', flexShrink: 0 }}
          >
            <span style={{ position: 'absolute', top: 3, left: isActive ? 21 : 3, width: 20, height: 20, borderRadius: '50%', background: '#fff', transition: 'left .2s', boxShadow: '0 1px 3px rgba(0,0,0,.2)' }} />
          </button>
        </div>
        <button type="submit" className="sm-btn" disabled={isSaving || !label.trim() || !startDate || !endDate} style={{ width: '100%', marginTop: 4 }}>
          {isSaving ? 'Speichern…' : 'Speichern'}
        </button>
        <button type="button" className="sm-btn ghost" onClick={onClose} style={{ width: '100%' }}>
          Abbrechen
        </button>
      </form>
    </BottomSheet>
  );
}

// ── Delete year confirmation ────────────────────────────────────────────────────

function DeleteYearModal({ year, onClose, onConfirm, isDeleting }: { year: ClubYear; onClose: () => void; onConfirm: () => void; isDeleting: boolean }) {
  return (
    <BottomSheet onClose={onClose}>
      <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 19, marginBottom: 10 }}>
        Abrechnungsjahr löschen?
      </div>
      {year.isActive && (
        <div style={{ background: 'var(--crit-bg, #fff1f0)', borderLeft: '3px solid var(--crit)', borderRadius: 8, padding: '9px 12px', marginBottom: 12, fontSize: 13, fontWeight: 700, color: 'var(--crit)' }}>
          Das ist das aktive Abrechnungsjahr. Nach dem Löschen gibt es kein aktives Jahr mehr.
        </div>
      )}
      <div style={{ fontSize: 14, color: 'var(--ink-2)', fontWeight: 600, marginBottom: 20 }}>
        <span style={{ color: 'var(--ink)' }}>{year.label}</span> wird unwiderruflich gelöscht. Alle zugehörigen Stundenbuchungen und Abgeltungsbeträge gehen verloren.
      </div>
      <button onClick={onConfirm} disabled={isDeleting} className="sm-btn" style={{ width: '100%', marginBottom: 10, background: 'var(--crit)', borderColor: 'var(--crit)' }}>
        {isDeleting ? 'Löschen…' : 'Endgültig löschen'}
      </button>
      <button onClick={onClose} className="sm-btn ghost" style={{ width: '100%' }}>
        Abbrechen
      </button>
    </BottomSheet>
  );
}

// ── Fee tier editor modal ──────────────────────────────────────────────────────

function FeeTierModal({ yearLabel, initialTiers, isGlobalFallback, onClose, onSave, isSaving }: {
  yearLabel: string;
  initialTiers: FeeTier[];
  isGlobalFallback: boolean;
  onClose: () => void;
  onSave: (tiers: { position: number; amountCents: number }[]) => void;
  isSaving: boolean;
}) {
  const [rows, setRows] = useState<TierRow[]>(
    initialTiers.length > 0
      ? initialTiers.map((t) => ({ amountEuros: (t.amountCents / 100).toFixed(2) }))
      : [{ amountEuros: '0.00' }],
  );

  return (
    <BottomSheet onClose={onClose}>
      <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 700, fontSize: 19, marginBottom: 4 }}>
        Abgeltungsbeträge
      </div>
      <div style={{ fontSize: 12.5, color: 'var(--muted)', fontWeight: 600, marginBottom: 14 }}>{yearLabel}</div>

      {isGlobalFallback && (
        <div style={{ background: 'var(--warn-bg, #fffbe6)', borderLeft: '3px solid var(--warn)', borderRadius: 8, padding: '9px 12px', marginBottom: 14 }}>
          <div className="sm-hint" style={{ margin: 0 }}>Aktuell globale Einstellung als Fallback. Hier gespeicherte Beträge gelten nur für dieses Jahr.</div>
        </div>
      )}

      <div style={{ fontWeight: 700, fontSize: 14, marginBottom: 4 }}>Betrag je Fehlstunde</div>
      <div className="sm-hint" style={{ marginBottom: 12 }}>Letzter Wert gilt für alle weiteren Fehlstunden.</div>

      <FeeTierRows rows={rows} onChange={setRows} />

      <button onClick={() => onSave(tierRowsToCents(rows))} disabled={isSaving || rows.length === 0}
        className="sm-btn" style={{ width: '100%', marginTop: 16, marginBottom: 10 }}>
        {isSaving ? 'Speichern…' : 'Speichern'}
      </button>
      <button onClick={onClose} className="sm-btn ghost" style={{ width: '100%' }}>Abbrechen</button>
    </BottomSheet>
  );
}

// ── Main screen ────────────────────────────────────────────────────────────────

export function AdminBilling() {
  const { showToast } = useAppStore();
  const { data: settings } = useSettings();
  const navigate = useNavigate();
  const params = useParams();
  const rest = params['*'] ?? '';
  const closeModal = useSmartBack(routes.abrechnungen);

  const yearsQ = useClubYears();
  const years = yearsQ.data ?? [];

  const activeYear = years.find((y) => y.isActive) ?? years[0];
  const [selectedYearId, setSelectedYearId] = useState<string>('');
  const effectiveYearId = selectedYearId || activeYear?.id || '';
  const selectedYear = years.find((y) => y.id === effectiveYearId);

  const [deleteTarget, setDeleteTarget] = useState<ClubYear | null>(null);
  const createOpen = rest === 'neu';
  const editMatch = /^([^/]+)\/bearbeiten$/.exec(rest);
  const feeMatch = /^([^/]+)\/staffeln$/.exec(rest);
  const editYear = editMatch ? years.find((y) => y.id === editMatch[1]) : undefined;
  const feeYear = feeMatch ? years.find((y) => y.id === feeMatch[1]) : undefined;
  // The fee-tier modal's own year (from the URL) takes precedence over
  // whichever year happens to be selected in the list — otherwise a direct
  // link or a reload on /abrechnungen/:yearId/staffeln would fetch (and, on
  // save, overwrite) the *selected* year's tiers instead of the URL year's.
  const feeTierYearId = feeYear?.id ?? effectiveYearId;

  const billingQ = useBilling(effectiveYearId || undefined);
  const feeTiersQ = useClubYearFeeTiers(feeTierYearId || undefined);
  const exportCSV = useExportBillingCSV();
  const exportPDF = useExportBillingPDF();
  const createYear = useCreateClubYear();
  const updateYear = useUpdateClubYear();
  const deleteYear = useDeleteClubYear();
  const updateFeeTiers = useUpdateClubYearFeeTiers();

  const report = billingQ.data;
  const feeTiers = feeTiersQ.data ?? [];
  const isGlobalFallback = !feeTiersQ.isLoading && feeTiers.length === 0;
  const globalFeeSchedule: number[] = (settings?.feeSchedule ?? []).map(Number);

  const onSaveCreate = (label: string, startDate: string, endDate: string, targetHours: number, tiers: { position: number; amountCents: number }[]) => {
    createYear.mutate(
      {
        label,
        startDate: new Date(startDate + 'T00:00:00Z').toISOString(),
        endDate: new Date(endDate + 'T23:59:59Z').toISOString(),
        defaultTargetHours: targetHours,
        setActive: true,
      },
      {
        onSuccess: (y) => {
          if (tiers.length > 0) {
            updateFeeTiers.mutate({ clubYearId: y.id, tiers }, {
              onError: () => showToast('Abgeltungsbeträge konnten nicht gespeichert werden.', 'crit'),
            });
          }
          showToast(`Abrechnungsjahr ${label} angelegt.`, 'ok');
          setSelectedYearId(y.id);
          closeModal();
        },
        onError: (e: unknown) => showToast(e instanceof Error ? e.message : 'Fehler beim Anlegen.', 'crit'),
      },
    );
  };

  const onSaveEdit = (label: string, startDate: string, endDate: string, targetHours: number, isActive: boolean) => {
    if (!editYear) return;
    updateYear.mutate(
      {
        id: editYear.id,
        data: {
          label,
          startDate: new Date(startDate + 'T00:00:00Z').toISOString(),
          endDate: new Date(endDate + 'T23:59:59Z').toISOString(),
          defaultTargetHours: targetHours,
          setActive: isActive,
        },
      },
      {
        onSuccess: () => { showToast('Abrechnungsjahr aktualisiert.', 'ok'); closeModal(); },
        onError: (e: unknown) => showToast(e instanceof Error ? e.message : 'Fehler beim Speichern.', 'crit'),
      },
    );
  };

  const onConfirmDelete = () => {
    if (!deleteTarget) return;
    const yearId = deleteTarget.id;
    deleteYear.mutate(yearId, {
      onSuccess: () => {
        showToast('Abrechnungsjahr gelöscht.', 'ok');
        setDeleteTarget(null);
        if (selectedYearId === yearId) setSelectedYearId('');
      },
      onError: (e: unknown) => showToast(e instanceof Error ? e.message : 'Fehler beim Löschen.', 'crit'),
    });
  };

  const onSaveFeeTiers = (tiers: { position: number; amountCents: number }[]) => {
    if (!feeTierYearId) return;
    updateFeeTiers.mutate(
      { clubYearId: feeTierYearId, tiers },
      {
        onSuccess: () => { showToast('Abgeltungsbeträge gespeichert.', 'ok'); closeModal(); },
        onError: (e: unknown) => showToast(e instanceof Error ? e.message : 'Fehler beim Speichern.', 'crit'),
      },
    );
  };

  const withFee = report?.results.filter((r) => r.missingHours > 0) ?? [];
  const noFee = report?.results.filter((r) => r.missingHours === 0) ?? [];
  const hasAmounts = (report?.totalCents ?? 0) > 0;

  return (
    <div className="fade-in">
      {/* ── Header ── */}
      <div className="sm-header">
        <div>
          <div className="sm-title">Abrechnungen</div>
        </div>
        {report && (
          <div style={{ display: 'flex', gap: 8 }}>
            <button className="pressable" onClick={() => exportPDF.mutate(effectiveYearId, { onError: () => showToast('PDF-Export fehlgeschlagen.', 'crit') })}
              disabled={exportPDF.isPending}
              style={{ width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }} title="PDF exportieren">
              <Icon name="list" size={19} color="var(--ink)" />
            </button>
            <button className="pressable" onClick={() => exportCSV.mutate(effectiveYearId, { onError: () => showToast('CSV-Export fehlgeschlagen.', 'crit') })}
              disabled={exportCSV.isPending}
              style={{ width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }} title="CSV exportieren">
              <Icon name="download" size={20} color="var(--ink)" />
            </button>
          </div>
        )}
      </div>

      <div className="sm-pad">
        {yearsQ.isLoading && <LoadingState />}
        {yearsQ.isError && <ErrorState />}

        {/* ── Year list ── */}
        {!yearsQ.isLoading && (
          <>
            <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.5, padding: '0 4px', marginBottom: 8 }}>
              Abrechnungsjahre
            </div>

            {years.length === 0 ? (
              <div className="sm-card" style={{ padding: 28, textAlign: 'center', color: 'var(--muted)', fontWeight: 600, marginBottom: 12 }}>
                Noch kein Abrechnungsjahr angelegt.
              </div>
            ) : (
              <div className="sm-card" style={{ overflow: 'hidden', marginBottom: 12 }}>
                {years.map((y, i) => (
                  <div key={y.id}>
                    {i > 0 && <hr className="sm-divider" />}
                    <div
                      className="pressable"
                      onClick={() => setSelectedYearId(y.id)}
                      style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px', background: y.id === effectiveYearId ? 'var(--primary-tint, rgba(0,0,0,0.04))' : 'transparent', cursor: 'pointer' }}
                    >
                      <div style={{ width: 4, alignSelf: 'stretch', borderRadius: 2, background: y.id === effectiveYearId ? 'var(--primary)' : 'transparent', flexShrink: 0 }} />
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{ display: 'flex', alignItems: 'center', gap: 7 }}>
                          <span style={{ fontWeight: 700, fontSize: 15 }}>{y.label}</span>
                          {y.isActive && <Badge kind="ok">Aktiv</Badge>}
                        </div>
                        <div style={{ fontSize: 12, color: 'var(--muted)', fontWeight: 600, marginTop: 2 }}>
                          {fmtDate(y.startDate)} – {fmtDate(y.endDate)} · Ziel: {y.defaultTargetHours} h
                        </div>
                      </div>
                      <button
                        className="pressable"
                        onClick={(e) => { e.stopPropagation(); navigate(routes.abrechnungBearbeiten(y.id)); }}
                        title="Bearbeiten"
                        style={{ width: 32, height: 32, borderRadius: 8, border: '1.5px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', flexShrink: 0 }}
                      >
                        <Icon name="edit" size={14} color="var(--muted)" />
                      </button>
                      <button
                        className="pressable"
                        onClick={(e) => { e.stopPropagation(); setDeleteTarget(y); }}
                        title="Löschen"
                        style={{ width: 32, height: 32, borderRadius: 8, border: '1.5px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', flexShrink: 0 }}
                      >
                        <Icon name="trash" size={14} color="var(--crit)" />
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}

            <button className="sm-btn soft" style={{ width: '100%', marginBottom: 22 }} onClick={() => navigate(routes.abrechnungenNeu)}>
              <Icon name="plus" size={17} stroke={2.2} /> Neues Abrechnungsjahr anlegen
            </button>
          </>
        )}

        {/* ── Billing for selected year ── */}
        {effectiveYearId && selectedYear && (
          <>
            <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.5, padding: '0 4px', marginBottom: 8 }}>
              Abrechnung — {selectedYear.label}
            </div>

            {!feeTiersQ.isLoading && (
              <div
                className="sm-card pressable"
                style={{ padding: '11px 14px', marginBottom: 14, display: 'flex', alignItems: 'center', gap: 12, cursor: 'pointer' }}
                onClick={() => navigate(routes.abrechnungStaffeln(selectedYear.id))}
              >
                <div style={{ flex: 1 }}>
                  <div style={{ fontWeight: 700, fontSize: 13.5 }}>Abgeltungsbeträge</div>
                  <div style={{ fontSize: 12, color: 'var(--muted)', fontWeight: 600, marginTop: 1 }}>
                    {feeTiers.length > 0
                      ? `${feeTiers.length} Stufe${feeTiers.length !== 1 ? 'n' : ''} · ${feeTiers.map((t) => euro(t.amountCents)).join(', ')}`
                      : isGlobalFallback
                        ? `Globale Einstellung: ${globalFeeSchedule.length > 0 ? globalFeeSchedule.map((v) => euro(v * 100)).join(', ') : '–'}`
                        : 'Laden…'}
                  </div>
                </div>
                <Icon name="edit" size={16} color="var(--muted)" />
              </div>
            )}

            {billingQ.isLoading && <LoadingState />}

            {report && (
              <>
                <div className="sm-card" style={{ padding: '14px 16px', marginBottom: 14, display: 'flex', gap: 20 }}>
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.5 }}>Vereinseinnahmen</div>
                    <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 26, color: hasAmounts ? 'var(--crit)' : 'var(--ink)', marginTop: 2 }}>{euro(report.totalCents)}</div>
                  </div>
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.5 }}>Mit Fehlstunden</div>
                    <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 26, color: 'var(--ink)', marginTop: 2 }}>{withFee.length}</div>
                  </div>
                  <div style={{ flex: 1 }}>
                    <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.5 }}>Ziel erfüllt</div>
                    <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 26, color: 'var(--ok)', marginTop: 2 }}>{noFee.length}</div>
                  </div>
                </div>

                {!hasAmounts && withFee.length > 0 && (
                  <div className="sm-card" style={{ padding: 14, display: 'flex', gap: 12, alignItems: 'flex-start', background: 'var(--warn-bg, #fffbe6)', borderLeft: '3px solid var(--warn)', marginBottom: 14 }}>
                    <Icon name="info" size={19} color="var(--warn)" style={{ flexShrink: 0, marginTop: 1 }} />
                    <div>
                      <div style={{ fontWeight: 800, fontSize: 14 }}>Keine Abgeltungsbeträge konfiguriert</div>
                      <div style={{ color: 'var(--ink-2)', fontSize: 13, fontWeight: 600, marginTop: 3 }}>Konfiguriere oben die Abgeltungsbeträge für dieses Jahr.</div>
                    </div>
                  </div>
                )}

                {withFee.length > 0 && (
                  <>
                    <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.5, padding: '0 4px', marginBottom: 8 }}>
                      Fehlstunden ({withFee.length})
                    </div>
                    <div className="sm-card" style={{ overflow: 'hidden', marginBottom: 14 }}>
                      {withFee.map((r, i) => (
                        <div key={r.memberId}>
                          {i > 0 && <hr className="sm-divider" style={{ marginLeft: 62 }} />}
                          <MemberRow r={r} />
                        </div>
                      ))}
                    </div>
                  </>
                )}

                {noFee.length > 0 && (
                  <>
                    <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.5, padding: '0 4px', marginBottom: 8 }}>
                      Ziel erfüllt ({noFee.length})
                    </div>
                    <div className="sm-card" style={{ overflow: 'hidden' }}>
                      {noFee.map((r, i) => (
                        <div key={r.memberId}>
                          {i > 0 && <hr className="sm-divider" style={{ marginLeft: 62 }} />}
                          <MemberRow r={r} />
                        </div>
                      ))}
                    </div>
                  </>
                )}

                {report.results.length === 0 && (
                  <div className="sm-card" style={{ padding: 28, textAlign: 'center', color: 'var(--muted)', fontWeight: 600 }}>
                    Keine aktiven Mitglieder für dieses Abrechnungsjahr.
                  </div>
                )}

                <div style={{ fontSize: 11.5, color: 'var(--muted)', fontWeight: 600, textAlign: 'center', marginTop: 16 }}>
                  Berechnet am {new Date(report.computedAt).toLocaleString('de-DE', { dateStyle: 'medium', timeStyle: 'short' })}
                </div>
              </>
            )}
          </>
        )}
      </div>

      {/* ── Modals ── */}
      {createOpen && (
        <CreateYearModal
          defaultTargetHours={settings?.yearGoal ?? 20}
          globalFeeSchedule={globalFeeSchedule}
          onClose={closeModal}
          onSave={onSaveCreate}
          isSaving={createYear.isPending}
        />
      )}
      {editYear && (
        <EditYearModal year={editYear} onClose={closeModal} onSave={(l, s, e, h, a) => onSaveEdit(l, s, e, h, a)} isSaving={updateYear.isPending} />
      )}
      {deleteTarget && (
        <DeleteYearModal year={deleteTarget} onClose={() => setDeleteTarget(null)} onConfirm={onConfirmDelete} isDeleting={deleteYear.isPending} />
      )}
      {feeYear && (
        <FeeTierModal
          yearLabel={feeYear.label}
          initialTiers={feeTiers}
          isGlobalFallback={isGlobalFallback}
          onClose={closeModal}
          onSave={onSaveFeeTiers}
          isSaving={updateFeeTiers.isPending}
        />
      )}
    </div>
  );
}
