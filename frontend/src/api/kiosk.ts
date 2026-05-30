import { useQuery, useMutation } from '@tanstack/react-query';
import { apiGet, apiPost } from './client';
import type { Event, Member } from '@/types';

// ── Queries ────────────────────────────────────────────────────────────────

/** Published events for the standalone kiosk (no auth). */
export function useKioskEvents() {
  return useQuery({
    queryKey: ['kiosk', 'events'],
    queryFn: () => apiGet<Event[]>('/kiosk/events'),
  });
}

/** Member search for the kiosk — only used when settings.kioskSearch is enabled. */
export function useKioskMemberSearch(query: string, enabled: boolean) {
  return useQuery({
    queryKey: ['kiosk', 'members', query],
    queryFn: () => apiGet<Member[]>(`/kiosk/members?q=${encodeURIComponent(query)}`),
    enabled: enabled && query.trim().length >= 2,
  });
}

// ── Mutations ──────────────────────────────────────────────────────────────

export interface KioskRegisterPayload {
  shiftId: string;
  email?: string;
  memberId?: string;
}

/** Registers via the kiosk; the backend sends a confirmation e-mail. */
export function useKioskRegister() {
  return useMutation({
    mutationFn: (data: KioskRegisterPayload) =>
      apiPost<{ message: string }>('/kiosk/register', data),
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
    queryFn: () => apiPost<ConfirmResult>(`/shifts/confirm/${token}`),
    enabled: !!token,
    retry: false,
  });
}
