import { create } from 'zustand';
import type { AuthUser } from '@/types';
import { useAppStore } from '@/store/app.store';

interface AuthState {
  token: string | null;
  /** Epoch milliseconds at which `token` expires, or null if unknown. */
  expiresAt: number | null;
  user: AuthUser | null;
  login: (token: string, user: AuthUser, expiresAt: number | null) => void;
  /** Swaps in a freshly refreshed token without touching the user profile (A-004). */
  setToken: (token: string, expiresAt: number | null) => void;
  logout: () => void;
}

function readExpiresAt(): number | null {
  const raw = localStorage.getItem('sm_expires_at');
  const n = raw ? Number(raw) : NaN;
  return Number.isFinite(n) ? n : null;
}

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('sm_token'),
  expiresAt: readExpiresAt(),
  user: (() => {
    try {
      const raw = localStorage.getItem('sm_user');
      return raw ? (JSON.parse(raw) as AuthUser) : null;
    } catch {
      return null;
    }
  })(),

  login: (token, user, expiresAt) => {
    localStorage.setItem('sm_token', token);
    localStorage.setItem('sm_user', JSON.stringify(user));
    if (expiresAt !== null) {
      localStorage.setItem('sm_expires_at', String(expiresAt));
    } else {
      localStorage.removeItem('sm_expires_at');
    }
    set({ token, user, expiresAt });
  },

  setToken: (token, expiresAt) => {
    localStorage.setItem('sm_token', token);
    if (expiresAt !== null) {
      localStorage.setItem('sm_expires_at', String(expiresAt));
    } else {
      localStorage.removeItem('sm_expires_at');
    }
    set({ token, expiresAt });
  },

  logout: () => {
    localStorage.removeItem('sm_token');
    localStorage.removeItem('sm_user');
    localStorage.removeItem('sm_expires_at');
    set({ token: null, user: null, expiresAt: null });
    // `useAppStore` is a module-level singleton that outlives this logout (no
    // page reload happens), so a stale tab/navStack from the previous user
    // must be cleared here or the next login on the same tab can land on a
    // screen the new user's role no longer has access to.
    useAppStore.getState().go('start');
  },
}));
