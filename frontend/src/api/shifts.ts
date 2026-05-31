import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiPost, apiPut, apiDelete } from './client';
import type { Shift } from '@/types';

interface CreateShiftPayload extends Omit<Shift, 'id' | 'signups'> {
  eventId: string;
  date: string;
}

interface UpdateShiftPayload extends Partial<Omit<Shift, 'signups'>> {
  id: string;
  eventId: string;
}

// ── Mutations ──────────────────────────────────────────────────────────────

export function useCreateShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ eventId, date, ...data }: CreateShiftPayload) =>
      apiPost<Shift>(`/events/${eventId}/shifts`, { ...data, date }),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['events', vars.eventId] });
      qc.invalidateQueries({ queryKey: ['events'] });
    },
  });
}

export function useUpdateShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, eventId: _eventId, ...data }: UpdateShiftPayload) =>
      apiPut<Shift>(`/shifts/${id}`, data),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['events', vars.eventId] });
    },
  });
}

export function useDeleteShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, eventId: _eventId }: { id: string; eventId: string }) =>
      apiDelete<void>(`/shifts/${id}`),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['events', vars.eventId] });
    },
  });
}

export function useRegisterShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      shiftId,
      comment,
      otherEmail,
    }: {
      shiftId: string;
      comment?: string;
      otherEmail?: string;
    }) => apiPost<{ message: string }>(`/shifts/${shiftId}/register`, { comment, otherEmail }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['events'] });
      qc.invalidateQueries({ queryKey: ['my-hours'] });
    },
  });
}

export function useDeregisterShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (shiftId: string) =>
      apiDelete<{ message: string }>(`/shifts/${shiftId}/register`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['events'] });
      qc.invalidateQueries({ queryKey: ['my-hours'] });
    },
  });
}

export function useConfirmShiftHours() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      shiftId,
      memberId,
      hours,
      description,
    }: {
      shiftId: string;
      memberId: string;
      hours: number;
      description?: string;
    }) => apiPost<{ message: string }>('/hours/confirm', { shiftId, memberId, hours, description }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['events'] });
      qc.invalidateQueries({ queryKey: ['members'] });
    },
  });
}
