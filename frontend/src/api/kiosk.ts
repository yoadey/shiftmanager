import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiGet, apiPost } from './client';
import type { Event, Member } from '@/types';
import { flatToEvent, timelineToEvent, type RawEvent, type RawTimeline } from './mappers';

/** The kiosk search returns only id and first name plus last initial (K-003). */
interface RawKioskMember {
  id: string;
  name?: string;
}

// ── Queries ────────────────────────────────────────────────────────────────

export interface KioskConfig {
  kioskSearch: boolean;
  clubName?: string;
  logoUrl?: string;
  primaryColor?: string;
}

/** Kiosk settings and branding; fails with 401 without a valid kiosk link
 * and 403 while the kiosk is locked (K-016, K-012). */
export function useKioskConfig() {
  return useQuery({
    queryKey: ['kiosk', 'config'],
    queryFn: () => apiGet<KioskConfig>('/kiosk/config'),
    retry: false,
  });
}

/**
 * Published events for the standalone kiosk (no auth), enriched with their
 * public timeline (days/shifts) via GET /kiosk/events/{id}. Per-event failures
 * degrade to a flat event rather than failing the whole list.
 */
export function useKioskEvents(enabled = true) {
  return useQuery({
    queryKey: ['kiosk', 'events'],
    enabled,
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
      // The name already comes abbreviated ("Max M."); carry it as the first
      // name so `${first} ${last}` renders it unchanged.
      return (Array.isArray(rows) ? rows : []).map((r) => ({
        id: r.id,
        first: r.name ?? '',
        last: '',
        email: '',
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
  /** Full name (K-014), or display name together with an e-mail (K-006/K-007). */
  name?: string;
  email?: string;
  /** Member picked via the kiosk search (only when it is enabled). */
  memberId?: string;
}

/**
 * Registers a single person via the kiosk. One call per person —
 * POST /kiosk/shifts/{id}/register (K-008).
 */
export function useKioskRegister() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ shiftId, ...body }: KioskRegisterPayload) =>
      apiPost<{ message: string }>(`/kiosk/shifts/${shiftId}/register`, body),
    // Refresh the free places before the next person picks a shift (K-008).
    onSuccess: () => qc.invalidateQueries({ queryKey: ['kiosk', 'events'] }),
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
