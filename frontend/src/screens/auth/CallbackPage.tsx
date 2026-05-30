import { useEffect, useRef, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { Icon } from '@/components/ui/Icon';
import { Button } from '@/components/ui/Button';
import { useAuthStore } from '@/store/auth.store';
import { exchangeCode } from '@/api/auth';

export function CallbackPage() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);
  const [error, setError] = useState<string | null>(null);
  const ran = useRef(false);

  useEffect(() => {
    if (ran.current) return;
    ran.current = true;

    const code = params.get('code');
    if (!code) {
      setError('Kein Autorisierungscode gefunden.');
      return;
    }

    exchangeCode(code)
      .then((res) => {
        login(res.token, res.user);
        navigate('/', { replace: true });
      })
      .catch(() => setError('Anmeldung fehlgeschlagen. Bitte erneut versuchen.'));
  }, [params, login, navigate]);

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 24 }}>
      <div className="sm-card pad fade-in" style={{ width: '100%', maxWidth: 360, padding: 28, textAlign: 'center' }}>
        {error ? (
          <>
            <div style={{ width: 54, height: 54, borderRadius: '50%', background: 'var(--crit-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
              <Icon name="x" size={26} stroke={2.4} color="var(--crit)" />
            </div>
            <div style={{ fontWeight: 800, fontSize: 18 }}>Anmeldung fehlgeschlagen</div>
            <p style={{ color: 'var(--ink-2)', fontSize: 13.5, fontWeight: 600, margin: '8px 0 18px' }}>{error}</p>
            <Button icon="arrowR" onClick={() => navigate('/auth/login', { replace: true })}>Zurück zur Anmeldung</Button>
          </>
        ) : (
          <>
            <div className="rb-logo" style={{ width: 50, height: 50, fontSize: 19, margin: '0 auto 16px' }}>SG</div>
            <div style={{ fontWeight: 700, fontSize: 16 }}>Anmeldung wird abgeschlossen …</div>
            <div style={{ color: 'var(--muted)', fontSize: 13, fontWeight: 600, marginTop: 4 }}>Einen Moment bitte.</div>
          </>
        )}
      </div>
    </div>
  );
}
