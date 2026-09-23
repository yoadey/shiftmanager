import { describe, it, expect, vi } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import React from 'react';
import { apiClient } from './client';
import { useUploadEventAttachment } from './events';

function wrapper({ children }: { children: React.ReactNode }) {
  const qc = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return React.createElement(QueryClientProvider, { client: qc }, children);
}

describe('useUploadEventAttachment', () => {
  it('overrides the default JSON Content-Type so axios sends real multipart, not a JSON-stringified FormData', async () => {
    const postSpy = vi.spyOn(apiClient, 'post').mockResolvedValue({ data: { id: 'a1' } });
    const file = new File(['hello'], 'flyer.png', { type: 'image/png' });

    const { result } = renderHook(() => useUploadEventAttachment('ev-1'), { wrapper });
    act(() => {
      result.current.mutate(file);
    });
    await waitFor(() => expect(postSpy).toHaveBeenCalled());

    const [url, body, config] = postSpy.mock.calls[0];
    expect(url).toBe('/events/ev-1/attachments');
    expect(body).toBeInstanceOf(FormData);
    // apiClient has an instance-level default `Content-Type: application/json`
    // header. Without overriding it to null per-request, axios treats that
    // default as authoritative and JSON.stringify()s the FormData instead of
    // sending it as multipart — the file's bytes never reach the server.
    // This assertion is the regression guard for that bug.
    expect(config).toEqual({ headers: { 'Content-Type': null } });

    postSpy.mockRestore();
  });
});
