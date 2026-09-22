import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiGet, apiPost, apiPut } from './client';
import type { HourEntry } from '@/types';

// ── Queries ────────────────────────────────────────────────────────────────

export function useMyHours() {
  return useQuery({
    queryKey: ['my-hours'],
    queryFn: () => apiGet<{ entries: HourEntry[]; confirmed: number; reserved: number; goal: number }>('/hours/me'),
  });
}

export function useMemberHours(memberId: string) {
  return useQuery({
    queryKey: ['hours', memberId],
    queryFn: () => apiGet<{ entries: HourEntry[]; confirmed: number; reserved: number; goal: number }>(`/hours/${memberId}`),
    enabled: !!memberId,
  });
}

// ── Mutations ──────────────────────────────────────────────────────────────

export function useManualBooking() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { memberId: string; date?: string; hours: number; desc: string }) =>
      apiPost<HourEntry>('/hours/manual', { memberId: data.memberId, hours: data.hours, desc: data.desc }),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['hours', vars.memberId] });
      qc.invalidateQueries({ queryKey: ['my-hours'] });
      qc.invalidateQueries({ queryKey: ['members'] });
    },
  });
}

export function useUpdateHourEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: Partial<HourEntry> & { id: string }) =>
      apiPut<HourEntry>(`/hours/${id}`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['hours'] });
      qc.invalidateQueries({ queryKey: ['my-hours'] });
    },
  });
}
