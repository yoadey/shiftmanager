import { Icon } from '@/components/ui/Icon';
import { Button } from '@/components/ui/Button';
import { oidcLoginUrl } from '@/api/auth';

export function LoginPage() {
  return (
    <div
      style={{
        minHeight: '100vh',
        background: 'var(--bg)',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
      }}
    >
      <div className="sm-card pad fade-in" style={{ width: '100%', maxWidth: 380, padding: 28, textAlign: 'center' }}>
        <div className="rb-logo" style={{ width: 56, height: 56, fontSize: 21, margin: '0 auto 18px' }}>SG</div>
        <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 24 }}>ShiftManager</div>
        <div style={{ color: 'var(--muted)', fontSize: 13.5, fontWeight: 600, marginTop: 4 }}>TSC Schwarz-Gelb Aachen</div>

        <p style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, lineHeight: 1.5, margin: '22px 0 20px' }}>
          Melde dich mit deinem Vereinskonto an, um deine Schichten und dein Stundenkonto zu verwalten.
        </p>

        <Button icon="lock" onClick={() => { window.location.href = oidcLoginUrl(); }}>
          Login via OIDC
        </Button>

        <div style={{ display: 'flex', alignItems: 'center', gap: 8, marginTop: 18, color: 'var(--muted)', fontSize: 12, fontWeight: 600, justifyContent: 'center' }}>
          <Icon name="shield" size={14} color="var(--muted)" />
          Sichere Anmeldung über deinen Vereins-Identity-Provider
        </div>
      </div>
    </div>
  );
}
