import { useState } from 'react';
import { Icon } from '@/components/ui/Icon';
import { Button } from '@/components/ui/Button';
import { Sheet } from '@/components/ui/Sheet';
import { Field, Input } from '@/components/forms/Field';
import { useAppStore } from '@/store/app.store';
import { useKioskLinks, useCreateKioskLink, useRevokeKioskLink, kioskLinkUrl, type KioskLink } from '@/api/settings';
import { fmtDate } from '@/screens/_demo';

const VALIDITY_PRESETS = [30, 90, 365];

function linkState(l: KioskLink, now = Date.now()): 'aktiv' | 'abgelaufen' | 'widerrufen' {
  if (l.revokedAt) return 'widerrufen';
  return new Date(l.expiresAt).getTime() > now ? 'aktiv' : 'abgelaufen';
}

/**
 * Kiosk links (K-016): a link with a long-lived token unlocks the kiosk on a
 * device. The token is shown once, right after creating the link.
 */
export function KioskLinks() {
  const { showToast } = useAppStore();
  const linksQ = useKioskLinks();
  const create = useCreateKioskLink();
  const revoke = useRevokeKioskLink();
  const [label, setLabel] = useState('');
  const [days, setDays] = useState(365);
  const [created, setCreated] = useState<KioskLink | null>(null);
  const [confirmRevoke, setConfirmRevoke] = useState<KioskLink | null>(null);

  const validDays = Number.isInteger(days) && days >= 1 && days <= 3650;
  const onCreate = () => {
    create.mutate({ label: label.trim(), validDays: days }, {
      onSuccess: (link) => { setCreated(link); setLabel(''); },
      onError: () => showToast('Kiosk-Link konnte nicht erstellt werden.', 'crit'),
    });
  };
  const copy = async (url: string) => {
    try {
      await navigator.clipboard.writeText(url);
      showToast('Link kopiert.');
    } catch {
      showToast('Kopieren nicht möglich – bitte den Link markieren und kopieren.', 'crit');
    }
  };

  const links = linksQ.data ?? [];
  const active = links.filter((l) => linkState(l) === 'aktiv');
  const inactive = links.filter((l) => linkState(l) !== 'aktiv');

  return (
    <div className="sm-card pad" data-testid="kiosk-links">
      <div style={{ fontWeight: 700, fontSize: 14.5 }}>Kiosk-Links</div>
      <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600, marginBottom: 12 }}>
        Der Kiosk lässt sich nur über einen solchen Link öffnen. Den Link einmal auf dem Kiosk-Gerät aufrufen.
      </div>

      <div style={{ display: 'flex', gap: 10, flexWrap: 'wrap', alignItems: 'flex-end' }}>
        <div style={{ flex: '2 1 200px' }}>
          <Field label="Bezeichnung" htmlFor="kiosk-link-label">
            <Input id="kiosk-link-label" placeholder="z. B. Tablet Vereinsheim" value={label} onChange={(e) => setLabel(e.target.value)} />
          </Field>
        </div>
        <div style={{ flex: '1 1 140px' }}>
          <Field label="Gültig (Tage)" htmlFor="kiosk-link-days">
            <Input id="kiosk-link-days" type="number" min={1} max={3650} value={String(days)} onChange={(e) => setDays(Number(e.target.value))} />
          </Field>
        </div>
      </div>
      <div style={{ display: 'flex', gap: 6, flexWrap: 'wrap', margin: '-4px 0 12px' }}>
        {VALIDITY_PRESETS.map((d) => (
          <button key={d} className={'sm-chip' + (days === d ? ' active' : '')} onClick={() => setDays(d)}>
            {d === 365 ? '1 Jahr' : `${d} Tage`}
          </button>
        ))}
      </div>
      <Button icon="plus" size="sm" disabled={!validDays} loading={create.isPending} onClick={onCreate}>Kiosk-Link erstellen</Button>

      {active.length > 0 && (
        <div style={{ marginTop: 16, display: 'flex', flexDirection: 'column', gap: 8 }}>
          {active.map((l) => (
            <div key={l.id} style={{ display: 'flex', alignItems: 'center', gap: 10, padding: '10px 12px', borderRadius: 12, border: '1px solid var(--line)', background: 'var(--surface-2)' }}>
              <Icon name="lock" size={16} color="var(--muted)" />
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontWeight: 700, fontSize: 13.5 }}>{l.label || 'Kiosk-Link'}</div>
                <div style={{ color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>gültig bis {fmtDate(l.expiresAt.slice(0, 10))}</div>
              </div>
              <Button size="sm" variant="ghost" onClick={() => setConfirmRevoke(l)}>Widerrufen</Button>
            </div>
          ))}
        </div>
      )}
      {inactive.length > 0 && (
        <div style={{ marginTop: 10, color: 'var(--muted)', fontSize: 12, fontWeight: 600 }}>
          {inactive.length} abgelaufene oder widerrufene {inactive.length === 1 ? 'Link' : 'Links'}
        </div>
      )}

      {created?.token && (
        <Sheet variant="dialog" onClose={() => setCreated(null)}>
          <div style={{ padding: '4px 2px' }}>
            <h3 style={{ fontSize: 19, fontWeight: 800 }}>Kiosk-Link erstellt</h3>
            <p style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, margin: '8px 0 12px', lineHeight: 1.45 }}>
              Dieser Link wird nur jetzt angezeigt. Rufe ihn auf dem Kiosk-Gerät auf oder kopiere ihn dorthin. Gültig bis {fmtDate(created.expiresAt.slice(0, 10))}.
            </p>
            <div data-testid="kiosk-link-url" style={{ padding: '10px 12px', borderRadius: 10, background: 'var(--surface-2)', border: '1px solid var(--line)', fontFamily: 'ui-monospace, monospace', fontSize: 12.5, wordBreak: 'break-all', userSelect: 'all' }}>
              {kioskLinkUrl(created.token)}
            </div>
            <div style={{ display: 'flex', gap: 10, marginTop: 14 }}>
              <Button variant="ghost" onClick={() => setCreated(null)}>Schließen</Button>
              <Button icon="paperclip" onClick={() => copy(kioskLinkUrl(created.token ?? ''))}>Link kopieren</Button>
            </div>
          </div>
        </Sheet>
      )}

      {confirmRevoke && (
        <Sheet variant="dialog" onClose={() => setConfirmRevoke(null)}>
          <div style={{ textAlign: 'center', padding: '6px 4px 4px' }}>
            <h3 style={{ fontSize: 19, fontWeight: 800 }}>Kiosk-Link widerrufen?</h3>
            <p style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, margin: '8px 0 18px', lineHeight: 1.45 }}>
              Geräte mit „{confirmRevoke.label || 'Kiosk-Link'}“ können den Kiosk danach nicht mehr öffnen.
            </p>
            <div style={{ display: 'flex', gap: 10 }}>
              <Button variant="ghost" onClick={() => setConfirmRevoke(null)}>Abbrechen</Button>
              <Button
                variant="danger"
                onClick={() => {
                  revoke.mutate(confirmRevoke.id, { onError: () => showToast('Widerrufen fehlgeschlagen.', 'crit') });
                  setConfirmRevoke(null);
                }}
              >
                Widerrufen
              </Button>
            </div>
          </div>
        </Sheet>
      )}
    </div>
  );
}
