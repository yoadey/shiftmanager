import { useEffect, useRef } from 'react';
import { useAuthStore } from '@/store/auth.store';
import { refreshToken } from '@/api/auth';

const CHECK_INTERVAL_MS = 60_000;
/** Refresh once less than this much of the session's lifetime remains. */
const REFRESH_MARGIN_MS = 10 * 60_000;

/**
 * Keeps the session JWT alive while the app stays open, without requiring a
 * full OIDC round-trip (A-004: "Token-Refresh muss automatisch erfolgen,
 * solange die Session aktiv ist."). Polls the token's remaining lifetime and
 * silently re-issues it via POST /auth/refresh shortly before it expires.
 * A failed refresh (session actually gone) is left to the API client's 401
 * interceptor, which redirects to login.
 */
export function useTokenRefresh(): void {
  const token = useAuthStore((s) => s.token);
  const expiresAt = useAuthStore((s) => s.expiresAt);
  const setToken = useAuthStore((s) => s.setToken);
  const refreshing = useRef(false);

  useEffect(() => {
    if (!token) return;

    const maybeRefresh = async () => {
      if (refreshing.current) return;
      const due = expiresAt === null || expiresAt - Date.now() < REFRESH_MARGIN_MS;
      if (!due) return;

      refreshing.current = true;
      try {
        const res = await refreshToken();
        setToken(res.token, Date.now() + res.expiresIn * 1000);
      } catch {
        // Session no longer valid; the API client's 401 handling takes over.
      } finally {
        refreshing.current = false;
      }
    };

    void maybeRefresh();
    const id = window.setInterval(() => void maybeRefresh(), CHECK_INTERVAL_MS);
    return () => window.clearInterval(id);
  }, [token, expiresAt, setToken]);
}
