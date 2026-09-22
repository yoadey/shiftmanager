import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';

const apiPost = vi.fn().mockResolvedValue({ message: 'ok' });

vi.mock('./client', () => ({
  apiPost: (...args: unknown[]) => apiPost(...args),
  apiGet: vi.fn(),
}));

import { useKioskRegister } from './kiosk';

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return React.createElement(QueryClientProvider, { client: qc }, children);
}

describe('useKioskRegister', () => {
  beforeEach(() => apiPost.mockClear());

  it('posts to the shift-scoped register route with the body (without shiftId)', async () => {
    const { result } = renderHook(() => useKioskRegister(), { wrapper });
    act(() => {
      result.current.mutate({ shiftId: 's-42', email: 'x@example.de' });
    });
    await waitFor(() => expect(apiPost).toHaveBeenCalled());
    expect(apiPost).toHaveBeenCalledWith('/kiosk/shifts/s-42/register', { email: 'x@example.de' });
  });

  it('passes memberId in the body when given', async () => {
    const { result } = renderHook(() => useKioskRegister(), { wrapper });
    act(() => {
      result.current.mutate({ shiftId: 's-7', memberId: 'm-1' });
    });
    await waitFor(() => expect(apiPost).toHaveBeenCalled());
    expect(apiPost).toHaveBeenCalledWith('/kiosk/shifts/s-7/register', { memberId: 'm-1' });
  });
});
