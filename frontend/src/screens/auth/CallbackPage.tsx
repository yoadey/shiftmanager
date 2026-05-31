import { useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Icon } from '@/components/ui/Icon';
import { Button } from '@/components/ui/Button';
import { useAuthStore } from '@/store/auth.store';
import { useAppStore } from '@/store/app.store';
import { getMe } from '@/api/auth';
import type { AuthUser, UserRole } from '@/types';

// Friendly German messages for the error codes the backend passes in the
// callback URL fragment (#error=...).
const ERROR_MESSAGES: Record<string, string> = {
  invalid_state: 'Die Anmeldesitzung ist abgelaufen oder ungültig. Bitte versuche es erneut.',
  missing_code: 'Vom Identity-Provider kam kein Autorisierungscode zurück.',
  exchange_failed: 'Der Token-Austausch mit dem Identity-Provider ist fehlgeschlagen.',
  token_invalid: 'Das ID-Token konnte nicht verifiziert werden.',
  inactive: 'Dein Konto ist (noch) nicht freigeschaltet. Bitte wende dich an den Vorstand.',
  server_error: 'Bei der Anmeldung ist ein interner Fehler aufgetreten. Bitte versuche es erneut.',
};

// "registered" is not an error: the account was just auto-created and is
// waiting for an administrator to activate it.
const PENDING_CODES = new Set(['registered', 'inactive']);

function roleView(role: UserRole): 'mitglied' | 'vorstand' {
  return role === 'mitglied' ? 'mitglied' : 'vorstand';
}

type Status =
  | { kind: 'loading' }
  | { kind: 'error'; message: string }
  | { kind: 'pending'; message: string };

export function CallbackPage() {
  const navigate = useNavigate();
  const login = useAuthStore((s) => s.login);
  const setRole = useAppStore((s) => s.setRole);
  const [status, setStatus] = useState<Status>({ kind: 'loading' });
  const ran = useRef(false);

  useEffect(() => {
    if (ran.current) return;
    ran.current = true;

    const frag = new URLSearchParams(window.location.hash.replace(/^#/, ''));
    const err = frag.get('error');
    const token = frag.get('token');

    // Strip the sensitive fragment from the URL/history immediately.
    window.history.replaceState(null, '', window.location.pathname + window.location.search);

    if (err) {
      if (err === 'registered') {
        setStatus({
          kind: 'pending',
          message:
            'Dein Zugang wurde erstellt und wartet auf Freischaltung durch den Vorstand. Du erhältst Zugriff, sobald dein Konto aktiviert wurde.',
        });
      } else if (PENDING_CODES.has(err)) {
        setStatus({ kind: 'pending', message: ERROR_MESSAGES[err] });
      } else {
        setStatus({ kind: 'error', message: ERROR_MESSAGES[err] ?? 'Anmeldung fehlgeschlagen. Bitte erneut versuchen.' });
      }
      return;
    }

    if (!token) {
      setStatus({ kind: 'error', message: 'Es wurde kein Anmelde-Token empfangen. Bitte erneut versuchen.' });
      return;
    }

    // Persist the token first so the API client attaches it to /auth/me.
    localStorage.setItem('sm_token', token);
    getMe()
      .then((me) => {
        const user: AuthUser = { id: me.id, name: me.email, email: me.email, role: me.role };
        login(token, user);
        setRole(roleView(me.role));
        navigate('/', { replace: true });
      })
      .catch(() => {
        localStorage.removeItem('sm_token');
        setStatus({ kind: 'error', message: 'Dein Profil konnte nicht geladen werden. Bitte erneut anmelden.' });
      });
  }, [login, setRole, navigate]);

  return (
    <div style={{ minHeight: '100vh', background: 'var(--bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', padding: 24 }}>
      <div className="sm-card pad fade-in" style={{ width: '100%', maxWidth: 360, padding: 28, textAlign: 'center' }}>
        {status.kind === 'error' && (
          <>
            <div style={{ width: 54, height: 54, borderRadius: '50%', background: 'var(--crit-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
              <Icon name="x" size={26} stroke={2.4} color="var(--crit)" />
            </div>
            <div style={{ fontWeight: 800, fontSize: 18 }}>Anmeldung fehlgeschlagen</div>
            <p style={{ color: 'var(--ink-2)', fontSize: 13.5, fontWeight: 600, margin: '8px 0 18px' }}>{status.message}</p>
            <Button icon="arrowR" onClick={() => navigate('/auth/login', { replace: true })}>Zurück zur Anmeldung</Button>
          </>
        )}
        {status.kind === 'pending' && (
          <>
            <div style={{ width: 54, height: 54, borderRadius: '50%', background: 'var(--warn-bg)', display: 'flex', alignItems: 'center', justifyContent: 'center', margin: '0 auto 14px' }}>
              <Icon name="clock" size={26} stroke={2.4} color="var(--warn)" />
            </div>
            <div style={{ fontWeight: 800, fontSize: 18 }}>Warten auf Freischaltung</div>
            <p style={{ color: 'var(--ink-2)', fontSize: 13.5, fontWeight: 600, margin: '8px 0 18px' }}>{status.message}</p>
            <Button icon="arrowR" onClick={() => navigate('/auth/login', { replace: true })}>Zurück zur Anmeldung</Button>
          </>
        )}
        {status.kind === 'loading' && (
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
