import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiPost, apiPut, apiPatch, apiDelete } from './client';

export interface ShiftPayload {
  eventId: string;
  date: string;     // YYYY-MM-DD
  name: string;
  start: string;    // HH:MM (local time)
  end: string;      // HH:MM (local time)
  min: number;
  max: number;
  qual?: string;
}

function toISO(date: string, time: string): string {
  return new Date(`${date}T${time}:00`).toISOString();
}

// ── Mutations ──────────────────────────────────────────────────────────────

export function useCreateShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ eventId, date, name, start, end, min, max, qual }: ShiftPayload) =>
      apiPost(`/events/${eventId}/shifts`, {
        name,
        startAt: toISO(date, start),
        endAt: toISO(date, end),
        minHelpers: min,
        maxHelpers: max,
        requiredQualification: qual ?? '',
      }),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['events', vars.eventId] });
      qc.invalidateQueries({ queryKey: ['events'] });
    },
  });
}

export function useUpdateShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, eventId: _eventId, date, name, start, end, min, max, qual }: ShiftPayload & { id: string }) =>
      apiPut(`/shifts/${id}`, {
        name,
        startAt: toISO(date, start),
        endAt: toISO(date, end),
        minHelpers: min,
        maxHelpers: max,
        requiredQualification: qual ?? '',
      }),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['events', vars.eventId] });
      qc.invalidateQueries({ queryKey: ['events'] });
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
      qc.invalidateQueries({ queryKey: ['events'] });
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

export function usePatchRegistration() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ registrationId, state, bookedHours }: {
      registrationId: string;
      eventId: string;
      state?: string;
      bookedHours?: number;
    }) => apiPatch(`/registrations/${registrationId}`, { state, bookedHours }),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['events'] });
      qc.invalidateQueries({ queryKey: ['events', vars.eventId] });
    },
  });
}

export function useForceDeleteRegistration() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ registrationId }: { registrationId: string; eventId: string }) =>
      apiDelete<void>(`/registrations/${registrationId}`),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['events'] });
      qc.invalidateQueries({ queryKey: ['events', vars.eventId] });
    },
  });
}

export function useAddMemberToShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ shiftId, memberId }: { shiftId: string; memberId: string; eventId: string }) =>
      apiPost(`/shifts/${shiftId}/add-member`, { memberId }),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['events'] });
      qc.invalidateQueries({ queryKey: ['events', vars.eventId] });
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
