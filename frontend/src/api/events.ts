import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiGet, apiPost, apiPut, apiDelete } from './client';
import type { Event, EventTimeline } from '@/types';
import { flatToEvent, timelineToEvent, type RawEvent, type RawTimeline } from './mappers';

// ── Queries ────────────────────────────────────────────────────────────────

/**
 * Lists all events, enriched with their per-event timeline (days/shifts/
 * signups) so dashboards and discovery screens can render shift-level data.
 * Each timeline is fetched independently and failures degrade gracefully to a
 * flat event with an empty timeline rather than failing the whole list.
 */
export function useEvents() {
  return useQuery({
    queryKey: ['events'],
    queryFn: async (): Promise<Event[]> => {
      const list = await apiGet<RawEvent[]>('/events');
      const rows = Array.isArray(list) ? list : [];
      return Promise.all(
        rows.map(async (e) => {
          try {
            const tl = await apiGet<RawTimeline>(`/events/${e.id}/timeline`);
            return timelineToEvent(tl);
          } catch {
            return flatToEvent(e);
          }
        }),
      );
    },
  });
}

export function useEvent(id: string) {
  return useQuery({
    queryKey: ['events', id],
    queryFn: () => apiGet<RawTimeline>(`/events/${id}`).then(timelineToEvent),
    enabled: !!id,
  });
}

export function useEventTimeline(id: string) {
  return useQuery({
    queryKey: ['events', id, 'timeline'],
    queryFn: async (): Promise<EventTimeline> => {
      const tl = await apiGet<RawTimeline>(`/events/${id}/timeline`);
      const event = timelineToEvent(tl);
      return { event, days: event.days };
    },
    enabled: !!id,
  });
}

// ── Mutations ──────────────────────────────────────────────────────────────

export function useCreateEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Omit<Event, 'id'>) => apiPost<RawEvent>('/events', data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events'] }),
  });
}

export function useUpdateEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: Partial<Event> & { id: string }) =>
      apiPut<RawEvent>(`/events/${id}`, data),
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
    mutationFn: (id: string) => apiPost<RawEvent>(`/events/${id}/publish`),
    onSuccess: (_d, id) => {
      qc.invalidateQueries({ queryKey: ['events'] });
      qc.invalidateQueries({ queryKey: ['events', id] });
    },
  });
}
