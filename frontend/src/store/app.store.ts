import { create } from 'zustand';
import type { Tweaks } from '@/types';

// Veranstaltungsleiter and above (event managers, see useIsEventManager)
// land on the admin overview, plain members on the member start screen.
// Mirrors the same localStorage key auth.store reads, so a hard page reload
// while already logged in picks the right home screen without waiting for a
// login effect.
function initialTab(): string {
  try {
    const raw = localStorage.getItem('sm_user');
    if (raw) {
      const user = JSON.parse(raw) as { role?: string };
      if (user.role && user.role !== 'mitglied') return 'uebersicht';
    }
  } catch {
    // ignore malformed/inaccessible storage — fall back to the member start screen
  }
  return 'start';
}

interface NavFrame {
  name: string;
  params: Record<string, string>;
}

interface Toast {
  msg: string;
  kind: 'ok' | 'warn' | 'crit';
}

interface AppState {
  // Navigation
  tab: string;
  navStack: NavFrame[];

  // Toast
  toast: Toast | null;
  _toastTimer: ReturnType<typeof setTimeout> | null;

  // Tweaks / branding
  tweaks: Tweaks;

  // Global name-display mode (NM-005) — hydrated from the settings API at app root.
  nameMode: 'abbrev' | 'full';

  // Actions
  go: (tab: string) => void;
  push: (name: string, params?: Record<string, string>) => void;
  back: () => void;
  showToast: (msg: string, kind?: 'ok' | 'warn' | 'crit') => void;
  setTweak: <K extends keyof Tweaks>(key: K, value: Tweaks[K]) => void;
  setNameMode: (mode: 'abbrev' | 'full') => void;
}

export const useAppStore = create<AppState>((set, get) => ({
  tab: initialTab(),
  navStack: [],

  toast: null,
  _toastTimer: null,

  tweaks: {
    primaryColor: '#F4B63F',
    radius: 'standard',
    warmth: 'warm',
  },

  nameMode: 'abbrev',

  go: (tab) => set({ tab, navStack: [] }),

  push: (name, params = {}) =>
    set((s) => ({ navStack: [...s.navStack, { name, params }] })),

  back: () => set((s) => ({ navStack: s.navStack.slice(0, -1) })),

  showToast: (msg, kind = 'ok') => {
    const prev = get()._toastTimer;
    if (prev) clearTimeout(prev);
    const timer = setTimeout(() => set({ toast: null, _toastTimer: null }), 2600);
    set({ toast: { msg, kind }, _toastTimer: timer });
  },

  setTweak: (key, value) =>
    set((s) => ({ tweaks: { ...s.tweaks, [key]: value } })),

  setNameMode: (mode) => set({ nameMode: mode }),
}));
