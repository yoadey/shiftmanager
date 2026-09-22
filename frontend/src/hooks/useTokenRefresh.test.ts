import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, waitFor } from '@testing-library/react';
import { useAuthStore } from '@/store/auth.store';
import type { AuthUser } from '@/types';

const refreshTokenMock = vi.fn();

vi.mock('@/api/auth', () => ({
  refreshToken: () => refreshTokenMock(),
}));

import { useTokenRefresh } from './useTokenRefresh';

const user: AuthUser = { id: 'm1', name: 'Max', email: 'max@example.com', role: 'mitglied' };

describe('useTokenRefresh', () => {
  beforeEach(() => {
    refreshTokenMock.mockReset();
    useAuthStore.setState({ token: 'old-token', user, expiresAt: null });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('does nothing when there is no session', async () => {
    useAuthStore.setState({ token: null, user: null, expiresAt: null });
    renderHook(() => useTokenRefresh());
    await Promise.resolve();
    expect(refreshTokenMock).not.toHaveBeenCalled();
  });

  it('refreshes immediately when the expiry is unknown', async () => {
    refreshTokenMock.mockResolvedValue({ token: 'new-token', expiresIn: 3600 });
    renderHook(() => useTokenRefresh());

    await waitFor(() => expect(refreshTokenMock).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(useAuthStore.getState().token).toBe('new-token'));
    expect(useAuthStore.getState().expiresAt).not.toBeNull();
  });

  it('does not refresh while comfortably within the token lifetime', async () => {
    useAuthStore.setState({ token: 'old-token', user, expiresAt: Date.now() + 60 * 60_000 });
    renderHook(() => useTokenRefresh());

    await Promise.resolve();
    expect(refreshTokenMock).not.toHaveBeenCalled();
  });

  it('refreshes once the token is close to expiry', async () => {
    refreshTokenMock.mockResolvedValue({ token: 'new-token', expiresIn: 3600 });
    useAuthStore.setState({ token: 'old-token', user, expiresAt: Date.now() + 60_000 });
    renderHook(() => useTokenRefresh());

    await waitFor(() => expect(refreshTokenMock).toHaveBeenCalledTimes(1));
    await waitFor(() => expect(useAuthStore.getState().token).toBe('new-token'));
  });

  it('leaves the session untouched when the refresh call fails', async () => {
    refreshTokenMock.mockRejectedValue(new Error('401'));
    useAuthStore.setState({ token: 'old-token', user, expiresAt: Date.now() - 1 });
    renderHook(() => useTokenRefresh());

    await waitFor(() => expect(refreshTokenMock).toHaveBeenCalledTimes(1));
    expect(useAuthStore.getState().token).toBe('old-token');
  });
});
