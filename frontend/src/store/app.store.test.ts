import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { useAppStore } from './app.store';

function reset() {
  useAppStore.setState({
    role: 'mitglied',
    tab: 'start',
    navStack: [],
    toast: null,
    _toastTimer: null,
    tweaks: { primaryColor: '#F4B63F', radius: 'standard', warmth: 'warm' },
  });
}

describe('app.store', () => {
  beforeEach(reset);

  describe('navigation stack', () => {
    it('push appends frames with params', () => {
      const { push } = useAppStore.getState();
      push('event', { id: 'e1' });
      push('shift', { id: 's2' });
      expect(useAppStore.getState().navStack).toEqual([
        { name: 'event', params: { id: 'e1' } },
        { name: 'shift', params: { id: 's2' } },
      ]);
    });

    it('push defaults params to empty object', () => {
      useAppStore.getState().push('audit');
      expect(useAppStore.getState().navStack).toEqual([{ name: 'audit', params: {} }]);
    });

    it('back pops the last frame', () => {
      const { push, back } = useAppStore.getState();
      push('a');
      push('b');
      back();
      expect(useAppStore.getState().navStack).toEqual([{ name: 'a', params: {} }]);
    });

    it('go switches tab and clears the nav stack', () => {
      const { push, go } = useAppStore.getState();
      push('detail', { id: 'x' });
      go('profil');
      expect(useAppStore.getState().tab).toBe('profil');
      expect(useAppStore.getState().navStack).toEqual([]);
    });
  });

  it('setRole switches role and resets tab + nav stack', () => {
    const { push, setRole } = useAppStore.getState();
    push('detail');
    setRole('vorstand');
    const s = useAppStore.getState();
    expect(s.role).toBe('vorstand');
    expect(s.tab).toBe('start');
    expect(s.navStack).toEqual([]);
    expect(localStorage.getItem('sm_role')).toBe('vorstand');
  });

  describe('showToast', () => {
    beforeEach(() => vi.useFakeTimers());
    afterEach(() => vi.useRealTimers());

    it('sets the toast then clears it after the timeout', () => {
      useAppStore.getState().showToast('Gespeichert', 'ok');
      expect(useAppStore.getState().toast).toEqual({ msg: 'Gespeichert', kind: 'ok' });
      vi.advanceTimersByTime(2600);
      expect(useAppStore.getState().toast).toBeNull();
    });

    it('defaults the kind to "ok"', () => {
      useAppStore.getState().showToast('Hallo');
      expect(useAppStore.getState().toast).toEqual({ msg: 'Hallo', kind: 'ok' });
    });

    it('a second toast resets the previous timer (only one clear fires)', () => {
      const store = useAppStore.getState();
      store.showToast('first', 'warn');
      vi.advanceTimersByTime(2000);
      store.showToast('second', 'crit');
      // 1000ms after the second toast — the first timer must not have cleared it
      vi.advanceTimersByTime(1000);
      expect(useAppStore.getState().toast).toEqual({ msg: 'second', kind: 'crit' });
      vi.advanceTimersByTime(2600);
      expect(useAppStore.getState().toast).toBeNull();
    });
  });

  it('setTweak updates a single tweak key', () => {
    useAppStore.getState().setTweak('primaryColor', '#000000');
    expect(useAppStore.getState().tweaks.primaryColor).toBe('#000000');
    // other keys untouched
    expect(useAppStore.getState().tweaks.radius).toBe('standard');
    useAppStore.getState().setTweak('radius', 'weich');
    expect(useAppStore.getState().tweaks.radius).toBe('weich');
  });
});
