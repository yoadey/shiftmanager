import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { Button } from '@/components/ui/Button';
import { useAppStore } from '@/store/app.store';
import { useEmailLog, useResendEmail } from '@/api/settings';
import { LoadingState, MessageState, ErrorState } from '@/components/ui/States';

function fmtTs(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return new Intl.DateTimeFormat('de-DE', {
    day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit',
  }).format(d);
}

export function EmailLog() {
  const { back, showToast } = useAppStore();
  const { data, isLoading, isError } = useEmailLog();
  const resend = useResendEmail();
  const entries = data ?? [];

  const onResend = (id: string) => {
    resend.mutate(id, {
      onSuccess: () => showToast('E-Mail erneut gesendet.'),
      onError: () => showToast('Erneutes Senden fehlgeschlagen.', 'crit'),
    });
  };

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
        <div className="sm-title" style={{ fontSize: 22 }}>E-Mail-Protokoll</div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 4 }}>
        {isLoading && <LoadingState />}
        {isError && <ErrorState />}
        {!isLoading && !isError && entries.length === 0 && (
          <MessageState icon="mail" title="Keine E-Mails" text="Es wurden noch keine E-Mails versendet." />
        )}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 10 }}>
          {entries.map((e) => {
            const failed = e.status === 'failed';
            return (
              <div key={e.id} className="sm-card" style={{ padding: 13, borderLeft: `3px solid ${failed ? 'var(--crit)' : 'var(--ok)'}` }}>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                  <div style={{ flex: 1, minWidth: 0 }}>
                    <div style={{ fontWeight: 700, fontSize: 14, whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>{e.subject || e.template}</div>
                    <div style={{ color: 'var(--muted)', fontSize: 12.5, fontWeight: 600 }}>{e.to}</div>
                  </div>
                  <Badge kind={failed ? 'crit' : 'ok'}>{failed ? 'Fehler' : 'Gesendet'}</Badge>
                </div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 7, flexWrap: 'wrap' }}>
                  <span style={{ fontSize: 11.5, color: 'var(--muted)', fontWeight: 600 }}>{e.template}</span>
                  <span style={{ fontSize: 11.5, color: 'var(--muted)', fontWeight: 600 }}>· {fmtTs(e.createdAt)}</span>
                </div>
                {failed && e.error && (
                  <div style={{ marginTop: 8, fontSize: 12.5, fontWeight: 600, color: 'var(--crit)', background: 'var(--surface-2)', borderRadius: 8, padding: '8px 10px' }}>{e.error}</div>
                )}
                {failed && (
                  <div style={{ marginTop: 10 }}>
                    <Button variant="soft" size="sm" icon="mail" loading={resend.isPending} onClick={() => onResend(e.id)}>Erneut senden</Button>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
