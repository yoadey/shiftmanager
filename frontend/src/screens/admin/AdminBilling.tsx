import { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { useAppStore } from '@/store/app.store';
import { useClubYears, useBilling, useExportBillingCSV, useExportBillingPDF } from '@/api/billing';
import { LoadingState, ErrorState } from '@/components/ui/States';
import type { BillingResult } from '@/api/billing';
import type { ClubYear } from '@/types';

function euro(cents: number): string {
  return (cents / 100).toLocaleString('de-DE', { style: 'currency', currency: 'EUR' });
}

function hrsLabel(h: number): string {
  return h % 1 === 0 ? `${h} Std.` : `${h.toFixed(1)} Std.`;
}

function MemberRow({ r }: { r: BillingResult }) {
  const missing = r.missingHours > 0;
  return (
    <div style={{ display: 'flex', alignItems: 'center', gap: 12, padding: '12px 14px' }}>
      <div
        style={{
          width: 38, height: 38, borderRadius: '50%', flexShrink: 0,
          background: missing ? 'var(--crit-bg, #fff1f0)' : 'var(--ok-bg, #f0faf4)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
        }}
      >
        <Icon name={missing ? 'alert' : 'check'} size={18} color={missing ? 'var(--crit)' : 'var(--ok)'} />
      </div>
      <div style={{ flex: 1, minWidth: 0 }}>
        <div style={{ fontWeight: 700, fontSize: 14.5 }}>
          {r.member.firstName} {r.member.lastName}
        </div>
        <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>
          {hrsLabel(r.confirmedHours)} / {hrsLabel(r.targetHours)} Ziel
          {missing && (
            <span style={{ color: 'var(--crit)' }}> · {hrsLabel(r.missingHours)} fehlend</span>
          )}
        </div>
      </div>
      {r.totalCents > 0
        ? <Badge kind="crit">{euro(r.totalCents)}</Badge>
        : missing
          ? <Badge kind="warn">{hrsLabel(r.missingHours)} fehlend</Badge>
          : <Badge kind="ok">Erfüllt</Badge>
      }
    </div>
  );
}

function YearPicker({ years, selected, onChange }: { years: ClubYear[]; selected: string; onChange: (id: string) => void }) {
  return (
    <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', padding: '0 0 14px' }}>
      {years.map((y) => (
        <button
          key={y.id}
          onClick={() => onChange(y.id)}
          style={{
            padding: '5px 14px', borderRadius: 20, border: '1.5px solid',
            borderColor: selected === y.id ? 'var(--primary)' : 'var(--line)',
            background: selected === y.id ? 'var(--primary)' : 'var(--surface)',
            color: selected === y.id ? '#fff' : 'var(--ink)',
            fontWeight: 700, fontSize: 13.5, cursor: 'pointer',
          }}
        >
          {y.label}{y.isActive && ' ·'}
        </button>
      ))}
    </div>
  );
}

export function AdminBilling() {
  const { back, showToast } = useAppStore();

  const yearsQ = useClubYears();
  const years = yearsQ.data ?? [];

  // Default to active year
  const activeYear = years.find((y) => y.isActive) ?? years[0];
  const [selectedYearId, setSelectedYearId] = useState<string>('');
  const effectiveYearId = selectedYearId || activeYear?.id || '';

  const billingQ = useBilling(effectiveYearId || undefined);
  const exportCSV = useExportBillingCSV();
  const exportPDF = useExportBillingPDF();

  const report = billingQ.data;
  const selectedYear = years.find((y) => y.id === effectiveYearId);

  const onExportCSV = () => {
    if (!effectiveYearId) return;
    exportCSV.mutate(effectiveYearId, {
      onError: () => showToast('CSV-Export fehlgeschlagen.', 'crit'),
    });
  };

  const onExportPDF = () => {
    if (!effectiveYearId) return;
    exportPDF.mutate(effectiveYearId, {
      onError: () => showToast('PDF-Export fehlgeschlagen.', 'crit'),
    });
  };

  const withFee = report?.results.filter((r) => r.missingHours > 0) ?? [];
  const noFee = report?.results.filter((r) => r.missingHours === 0) ?? [];

  const noFeeTiers = !billingQ.isLoading && !billingQ.data && billingQ.isError;

  return (
    <div className="fade-in">
      <div className="sm-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: 10 }}>
          <button
            className="pressable"
            onClick={back}
            style={{ width: 36, height: 36, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer' }}
          >
            <Icon name="chevL" size={18} color="var(--ink)" stroke={2.4} />
          </button>
          <div>
            <div className="sm-eyebrow">{selectedYear?.label ?? ''}</div>
            <div className="sm-title">Jahresabrechnung</div>
          </div>
        </div>
        <div style={{ display: 'flex', gap: 8 }}>
          <button
            className="pressable"
            onClick={onExportPDF}
            disabled={exportPDF.isPending || !effectiveYearId || !report}
            style={{ width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }}
            title="PDF exportieren"
          >
            <Icon name="list" size={19} color="var(--ink)" />
          </button>
          <button
            className="pressable"
            onClick={onExportCSV}
            disabled={exportCSV.isPending || !effectiveYearId || !report}
            style={{ width: 42, height: 42, borderRadius: '50%', border: '1px solid var(--line)', background: 'var(--surface)', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', boxShadow: 'var(--shadow)' }}
            title="CSV exportieren"
          >
            <Icon name="download" size={20} color="var(--ink)" />
          </button>
        </div>
      </div>

      <div className="sm-pad">
        {yearsQ.isLoading && <LoadingState />}
        {yearsQ.isError && <ErrorState />}

        {!yearsQ.isLoading && years.length > 1 && (
          <YearPicker years={years} selected={effectiveYearId} onChange={setSelectedYearId} />
        )}

        {years.length === 0 && !yearsQ.isLoading && (
          <div className="sm-card" style={{ padding: 28, textAlign: 'center', color: 'var(--muted)', fontWeight: 600 }}>
            Kein Vereinsjahr konfiguriert.
          </div>
        )}

        {billingQ.isLoading && <LoadingState />}

        {noFeeTiers && (
          <div className="sm-card" style={{ padding: 20, display: 'flex', gap: 14, alignItems: 'flex-start', background: 'var(--warn-bg, #fffbe6)', borderLeft: '3px solid var(--warn)' }}>
            <Icon name="info" size={20} color="var(--warn)" style={{ flexShrink: 0, marginTop: 2 }} />
            <div>
              <div style={{ fontWeight: 800, fontSize: 14.5 }}>Keine Abgeltungsbeträge konfiguriert</div>
              <div style={{ color: 'var(--ink-2)', fontSize: 13, fontWeight: 600, marginTop: 4 }}>
                Bitte hinterlege zuerst eine Abgeltungsliste unter Einstellungen → Abgeltungsbeträge.
              </div>
            </div>
          </div>
        )}

        {report && (
          <>
            {/* Summary */}
            <div className="sm-card" style={{ padding: '14px 16px', marginBottom: 18, display: 'flex', gap: 20 }}>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.5 }}>Gesamt fällig</div>
                <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 26, color: report.totalCents > 0 ? 'var(--crit)' : withFee.length > 0 ? 'var(--warn)' : 'var(--ok)', marginTop: 2 }}>
                  {report.totalCents > 0 ? euro(report.totalCents) : withFee.length > 0 ? '—' : euro(0)}
                </div>
              </div>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.5 }}>Mit Fehlstunden</div>
                <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 26, color: 'var(--ink)', marginTop: 2 }}>
                  {withFee.length}
                </div>
              </div>
              <div style={{ flex: 1 }}>
                <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.5 }}>Ziel erfüllt</div>
                <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 26, color: 'var(--ok)', marginTop: 2 }}>
                  {noFee.length}
                </div>
              </div>
            </div>

            {withFee.length > 0 && (
              <>
                <div style={{ fontSize: 11.5, fontWeight: 700, color: 'var(--muted)', textTransform: 'uppercase', letterSpacing: 0.5, padding: '0 4px', marginBottom: 8 }}>
                  Fehlstunden ({withFee.length})
                </div>
                <div className="sm-card" style={{ overflow: 'hidden', marginBottom: 18 }}>
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
                Keine aktiven Mitglieder für dieses Vereinsjahr.
              </div>
            )}

            <div style={{ fontSize: 11.5, color: 'var(--muted)', fontWeight: 600, textAlign: 'center', marginTop: 18 }}>
              Berechnet am {new Date(report.computedAt).toLocaleString('de-DE', { dateStyle: 'medium', timeStyle: 'short' })}
            </div>
          </>
        )}
      </div>
    </div>
  );
}
