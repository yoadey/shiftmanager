import { create } from 'zustand';
import type { Tweaks } from '@/types';

type RoleView = 'mitglied' | 'vorstand';

interface NavFrame {
  name: string;
  params: Record<string, string>;
}

interface Toast {
  msg: string;
  kind: 'ok' | 'warn' | 'crit';
}

interface AppState {
  // Role / view
  role: RoleView;
  tab: string;
  navStack: NavFrame[];

  // Toast
  toast: Toast | null;
  _toastTimer: ReturnType<typeof setTimeout> | null;

  // Tweaks / branding
  tweaks: Tweaks;

  // Actions
  setRole: (role: RoleView) => void;
  go: (tab: string) => void;
  push: (name: string, params?: Record<string, string>) => void;
  back: () => void;
  showToast: (msg: string, kind?: 'ok' | 'warn' | 'crit') => void;
  setTweak: <K extends keyof Tweaks>(key: K, value: Tweaks[K]) => void;
}

export const useAppStore = create<AppState>((set, get) => ({
  role: (localStorage.getItem('sm_role') as RoleView) || 'mitglied',
  tab: 'start',
  navStack: [],

  toast: null,
  _toastTimer: null,

  tweaks: {
    primaryColor: '#F4B63F',
    radius: 'standard',
    warmth: 'warm',
  },

  setRole: (role) => {
    localStorage.setItem('sm_role', role);
    set({ role, tab: 'start', navStack: [] });
  },

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
}));
