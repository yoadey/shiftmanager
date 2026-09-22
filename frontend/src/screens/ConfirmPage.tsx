import { useParams, useNavigate } from 'react-router-dom';
import { Icon } from '@/components/ui/Icon';
import { Button } from '@/components/ui/Button';
import { useConfirmRegistration } from '@/api/kiosk';

export function ConfirmPage() {
  const { token = '' } = useParams<{ token: string }>();
  const navigate = useNavigate();
  const { data, isPending, isError } = useConfirmRegistration(token);

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 24 }}>
      <div className="sm-card pad fade-in" style={{ width: '100%', maxWidth: 380, padding: 28, textAlign: 'center' }}>
        {isPending ? (
          <>
            <div className="rb-logo" style={{ width: 50, height: 50, fontSize: 19, margin: '0 auto 16px' }}>SG</div>
            <div style={{ fontWeight: 700, fontSize: 16 }}>Anmeldung wird bestätigt …</div>
            <div style={{ color: 'var(--muted)', fontSize: 13, fontWeight: 600, marginTop: 4 }}>Einen Moment bitte.</div>
          </>
        ) : isError ? (
          <>
            <div style={{ width: 54, height: 54, borderRadius: '50%', background: 'var(--crit-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
              <Icon name="x" size={26} stroke={2.4} color="var(--crit)" />
            </div>
            <div style={{ fontWeight: 800, fontSize: 19 }}>Link ungültig</div>
            <p style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, margin: '8px 0 18px', lineHeight: 1.45 }}>
              Dieser Bestätigungslink ist abgelaufen oder wurde bereits verwendet.
            </p>
            <Button icon="arrowR" onClick={() => navigate('/')}>Zur App</Button>
          </>
        ) : (
          <>
            <div style={{ width: 64, height: 64, borderRadius: '50%', background: 'var(--ok-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 16px' }}>
              <Icon name="check" size={32} stroke={2.4} color="var(--ok)" />
            </div>
            <div style={{ fontFamily: 'Bricolage Grotesque', fontWeight: 800, fontSize: 22 }}>Schicht bestätigt</div>
            <p style={{ color: 'var(--ink-2)', fontSize: 14, fontWeight: 600, margin: '8px 0 18px', lineHeight: 1.45 }}>
              {data?.shiftName
                ? <>Deine Anmeldung für <b style={{ color: 'var(--ink)' }}>{data.shiftName}</b>{data.eventName ? <> ({data.eventName})</> : null} ist jetzt verbindlich.</>
                : 'Deine Anmeldung ist jetzt verbindlich. Vielen Dank für deine Unterstützung!'}
            </p>
            <Button icon="arrowR" onClick={() => navigate('/')}>Zur App</Button>
          </>
        )}
      </div>
    </div>
  );
}
