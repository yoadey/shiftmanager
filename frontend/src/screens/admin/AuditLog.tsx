import { Icon } from '@/components/ui/Icon';
import { Badge } from '@/components/ui/Badge';
import { useAppStore } from '@/store/app.store';
import { useAuditLog } from '@/api/settings';
import { DEMO_STATE } from '@/screens/_demo';

export function AuditLog() {
  const { back } = useAppStore();
  const { data } = useAuditLog();
  const entries = data ?? DEMO_STATE.audit;

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
        <div className="sm-title" style={{ fontSize: 22 }}>Audit-Log</div>
      </div>
      <div className="sm-pad" style={{ paddingTop: 4 }}>
        <div className="tl-rail">
          {entries.map((a, i) => (
            <div key={i} className="tl-shift">
              <span className="tl-dot" style={{ background: 'var(--ink-2)', top: 14 }} />
              <div className="sm-card" style={{ padding: 13 }}>
                <div style={{ fontWeight: 700, fontSize: 14, lineHeight: 1.35 }}>{a.what}</div>
                <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 7, flexWrap: 'wrap' }}>
                  <Badge kind="neutral" dot={false}>{a.cat}</Badge>
                  <span style={{ fontSize: 12, color: 'var(--muted)', fontWeight: 600 }}>{a.who}</span>
                  <span style={{ fontSize: 12, color: 'var(--muted)', fontWeight: 600 }}>· {a.ts}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
