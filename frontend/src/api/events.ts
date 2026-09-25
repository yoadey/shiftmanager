import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiGet, apiPost, apiPut, apiDelete, apiClient } from './client';
import type { Event, EventTimeline, EventAttachment, RecurrenceFrequency } from '@/types';
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

export interface CreateEventPayload {
  name: string;
  description: string;
  location: string;
  category: string;
  status: string;
  startDate: string;
  endDate: string;
  visibility?: string;
}

export function useCreateEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateEventPayload) => apiPost<RawEvent>('/events', data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events'] }),
  });
}

type UpdateEventPayload = {
  id: string;
  name?: string;
  description?: string;
  location?: string;
  category?: string;
  startDate?: string;
  endDate?: string;
  visibility?: string;
  status?: string;
};

export function useUpdateEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: UpdateEventPayload) =>
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

export function useCopyEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiClient.post(`/events/${id}/copy`).then((r) => r.data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events'] }),
  });
}

export function useCompleteEvent() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiPost<RawEvent>(`/events/${id}/complete`),
    onSuccess: (_d, id) => {
      qc.invalidateQueries({ queryKey: ['events'] });
      qc.invalidateQueries({ queryKey: ['events', id] });
    },
  });
}

// ── Recurrence (V-007) ────────────────────────────────────────────────────────

export interface GenerateRecurrencePayload {
  frequency: RecurrenceFrequency;
  until: string; // ISO date-time, inclusive
}

/**
 * Turns an event into a recurring series: the source event is stamped with
 * the recurrence config and follow-up occurrences (full copies including
 * shifts) are created weekly/monthly up to `until`.
 */
export function useGenerateEventRecurrence(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: GenerateRecurrencePayload) =>
      apiPost<RawEvent[]>(`/events/${eventId}/recurrence`, data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['events'] });
      qc.invalidateQueries({ queryKey: ['events', eventId] });
    },
  });
}

// ── Attachments (V-008) ──────────────────────────────────────────────────────

export function useEventAttachments(eventId: string) {
  return useQuery({
    queryKey: ['events', eventId, 'attachments'],
    queryFn: () => apiGet<EventAttachment[]>(`/events/${eventId}/attachments`),
    enabled: !!eventId,
  });
}

/**
 * Uploads an image or document for an event via multipart FormData.
 * apiClient sets a default `Content-Type: application/json` header; without
 * overriding it to null per-request, axios treats that default as
 * authoritative and JSON-serializes the FormData instead of sending it as
 * multipart (the file's bytes never reach the server), mirroring useUploadLogo.
 */
export function useUploadEventAttachment(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => {
      const form = new FormData();
      form.append('file', file);
      return apiClient
        .post<EventAttachment>(`/events/${eventId}/attachments`, form, { headers: { 'Content-Type': null } })
        .then((r) => r.data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events', eventId, 'attachments'] }),
  });
}

export function useDeleteEventAttachment(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (attachmentId: string) => apiDelete<void>(`/events/${eventId}/attachments/${attachmentId}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['events', eventId, 'attachments'] }),
  });
}

// ── Header image (V-010) ──────────────────────────────────────────────────────
// A single image shown at the top of the event detail page, kept separate
// from the Attachments list above (V-008) — its own field on the event, not
// an attachment entry.

/** Uploads (or replaces) an event's header image via multipart FormData. */
export function useUploadEventHeaderImage(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => {
      const form = new FormData();
      form.append('file', file);
      return apiClient
        .post<RawEvent>(`/events/${eventId}/header-image`, form, { headers: { 'Content-Type': null } })
        .then((r) => r.data);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['events', eventId] });
      qc.invalidateQueries({ queryKey: ['events'] });
    },
  });
}

export function useDeleteEventHeaderImage(eventId: string) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => apiDelete<void>(`/events/${eventId}/header-image`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['events', eventId] });
      qc.invalidateQueries({ queryKey: ['events'] });
    },
  });
}
