import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiGet, apiPost, apiPut, apiDelete } from './client';
import type { Event, EventTimeline } from '@/types';

// ── Queries ────────────────────────────────────────────────────────────────

export function useEvents() {
  return useQuery({
    queryKey: ['events'],
    queryFn: () => apiGet<Event[]>('/events'),
  });
}

export function useEvent(id: string) {
  return useQuery({
    queryKey: ['events', id],
    queryFn: () => apiGet<Event>(`/events/${id}`),
    enabled: !!id,
  });
}

export function useEventTimeline(id: string) {
  return useQuery({
    queryKey: ['events', id, 'timeline'],
    queryFn: () => apiGet<EventTimeline>(`/events/${id}/timeline`),
    enabled: !!id,
  });
}

// ── Mutations ──────────────────────────────────────────────────────────────

export function useCreateEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Omit<Event, 'id'>) => apiPost<Event>('/events', data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events'] }),
  });
}

export function useUpdateEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: Partial<Event> & { id: string }) =>
      apiPut<Event>(`/events/${id}`, data),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['events'] });
      qc.invalidateQueries({ queryKey: ['events', vars.id] });
    },
  });
}

export function useDeleteEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiDelete<void>(`/events/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events'] }),
  });
}

export function usePublishEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiPost<Event>(`/events/${id}/publish`),
    onSuccess: (_d, id) => {
      qc.invalidateQueries({ queryKey: ['events'] });
      qc.invalidateQueries({ queryKey: ['events', id] });
    },
  });
}
