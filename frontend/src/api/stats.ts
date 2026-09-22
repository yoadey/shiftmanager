import { useQuery } from '@tanstack/react-query';
import { apiGet } from './client';
import type { SystemStats } from '@/types';

// ── Queries ────────────────────────────────────────────────────────────────

/** System-wide admin statistics (D-004). Veranstaltungsleiter+ only. */
export function useStats() {
  return useQuery({
    queryKey: ['stats'],
    queryFn: () => apiGet<SystemStats>('/stats'),
  });
}
