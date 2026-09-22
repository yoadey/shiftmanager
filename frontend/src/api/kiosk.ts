import { useQuery, useMutation } from '@tanstack/react-query';
import { apiGet, apiPost } from './client';
import type { Event, Member } from '@/types';
import { flatToEvent, timelineToEvent, type RawEvent, type RawTimeline } from './mappers';

interface RawKioskMember {
  id: string;
  firstName?: string;
  lastName?: string;
  email?: string;
}

// ── Queries ────────────────────────────────────────────────────────────────

/**
 * Published events for the standalone kiosk (no auth), enriched with their
 * public timeline (days/shifts) via GET /kiosk/events/{id}. Per-event failures
 * degrade to a flat event rather than failing the whole list.
 */
export function useKioskEvents() {
  return useQuery({
    queryKey: ['kiosk', 'events'],
    queryFn: async (): Promise<Event[]> => {
      const list = await apiGet<RawEvent[]>('/kiosk/events');
      const rows = Array.isArray(list) ? list : [];
      return Promise.all(
        rows.map(async (e) => {
          try {
            const tl = await apiGet<RawTimeline>(`/kiosk/events/${e.id}`);
            return timelineToEvent(tl);
          } catch {
            return flatToEvent(e);
          }
        }),
      );
    },
  });
}

/** Member search for the kiosk — only used when settings.kioskSearch is enabled. */
export function useKioskMemberSearch(query: string, enabled: boolean) {
  return useQuery({
    queryKey: ['kiosk', 'members', query],
    queryFn: async (): Promise<Member[]> => {
      const rows = await apiGet<RawKioskMember[]>(`/kiosk/members?q=${encodeURIComponent(query.trim())}`);
      return (Array.isArray(rows) ? rows : []).map((r) => ({
        id: r.id,
        first: r.firstName ?? '',
        last: r.lastName ?? '',
        email: r.email ?? '',
        since: '',
        goal: null,
      }));
    },
    enabled: enabled && query.trim().length >= 2,
  });
}

// ── Mutations ──────────────────────────────────────────────────────────────

export interface KioskRegisterPayload {
  shiftId: string;
  email?: string;
  memberId?: string;
}

/**
 * Registers a single person via the kiosk; the backend sends a confirmation
 * e-mail. One call per person — POST /kiosk/shifts/{id}/register (K-008).
 */
export function useKioskRegister() {
  return useMutation({
    mutationFn: ({ shiftId, ...body }: KioskRegisterPayload) =>
      apiPost<{ message: string }>(`/kiosk/shifts/${shiftId}/register`, body),
  });
}

export interface ConfirmResult {
  message: string;
  eventName?: string;
  shiftName?: string;
}

/** Confirms a reserved registration from the e-mail link token. */
export function useConfirmRegistration(token: string) {
  return useQuery({
    queryKey: ['shift-confirm', token],
    queryFn: () => apiGet<ConfirmResult>(`/kiosk/confirm/${encodeURIComponent(token)}`),
    enabled: !!token,
    retry: false,
  });
}
