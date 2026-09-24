import { create } from 'zustand';
import type { Tweaks } from '@/types';

interface Toast {
  msg: string;
  kind: 'ok' | 'warn' | 'crit';
}

interface AppState {
  // Toast
  toast: Toast | null;
  _toastTimer: ReturnType<typeof setTimeout> | null;

  // Tweaks / branding
  tweaks: Tweaks;

  // Global name-display mode (NM-005) — hydrated from the settings API at app root.
  nameMode: 'abbrev' | 'full';

  // Actions
  showToast: (msg: string, kind?: 'ok' | 'warn' | 'crit') => void;
  setTweak: <K extends keyof Tweaks>(key: K, value: Tweaks[K]) => void;
  setNameMode: (mode: 'abbrev' | 'full') => void;
}

export const useAppStore = create<AppState>((set, get) => ({
  toast: null,
  _toastTimer: null,

  tweaks: {
    primaryColor: '#F4B63F',
    radius: 'standard',
    warmth: 'warm',
  },

  nameMode: 'abbrev',

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
