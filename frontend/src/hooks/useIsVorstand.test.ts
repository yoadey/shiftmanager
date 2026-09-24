import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useIsVorstand } from './useIsVorstand';
import { useIsEventManager } from './useIsEventManager';
import { useAuthStore } from '@/store/auth.store';
import type { AuthUser } from '@/types';

function userWithRole(role: AuthUser['role']): AuthUser {
  return { id: 'm1', name: 'Test User', email: 'test@example.com', role };
}

describe('useIsVorstand / useIsEventManager', () => {
  beforeEach(() => {
    useAuthStore.setState({ token: null, user: null, expiresAt: null });
  });

  it('no user logged in → both false', () => {
    expect(renderHook(() => useIsVorstand()).result.current).toBe(false);
    expect(renderHook(() => useIsEventManager()).result.current).toBe(false);
  });

  it('mitglied → neither vorstand nor event manager', () => {
    useAuthStore.setState({ user: userWithRole('mitglied') });
    expect(renderHook(() => useIsVorstand()).result.current).toBe(false);
    expect(renderHook(() => useIsEventManager()).result.current).toBe(false);
  });

  it('veranstaltungsleiter → event manager but not vorstand (matches RoleVeranstaltungsleiter backend gate)', () => {
    useAuthStore.setState({ user: userWithRole('veranstaltungsleiter') });
    expect(renderHook(() => useIsVorstand()).result.current).toBe(false);
    expect(renderHook(() => useIsEventManager()).result.current).toBe(true);
  });

  it('vorstand → both true', () => {
    useAuthStore.setState({ user: userWithRole('vorstand') });
    expect(renderHook(() => useIsVorstand()).result.current).toBe(true);
    expect(renderHook(() => useIsEventManager()).result.current).toBe(true);
  });

  it('admin → both true', () => {
    useAuthStore.setState({ user: userWithRole('admin') });
    expect(renderHook(() => useIsVorstand()).result.current).toBe(true);
    expect(renderHook(() => useIsEventManager()).result.current).toBe(true);
  });
});
