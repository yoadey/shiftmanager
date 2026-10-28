import { useQuery, useMutation, useQueryClient, keepPreviousData } from '@tanstack/react-query';
import { apiGet, apiPost, apiPut, apiDelete, apiClient } from './client';
import type { Member } from '@/types';

// ── Backend ↔ UI shape mapping ───────────────────────────────────────────────
// The backend serialises domain.Member with camelCase field names
// (firstName/lastName/joinedAt/individualGoalHours/isActive); the UI model uses
// shorter names (first/last/since/goal/active). These mappers bridge both so
// names/dates render and updates hit the right fields.

interface RawMember {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
  joinedAt: string;
  leftAt?: string | null;
  isActive: boolean;
  individualGoalHours?: number | null;
  role?: string;
  reminderOptOut?: boolean;
  notifyNewEvents?: boolean;
}

function toMember(r: RawMember): Member {
  return {
    id: r.id,
    first: r.firstName ?? '',
    last: r.lastName ?? '',
    email: r.email ?? '',
    since: r.joinedAt ?? '',
    goal: r.individualGoalHours ?? null,
    active: r.isActive,
    left: r.leftAt ?? null,
    reminderOptOut: r.reminderOptOut,
    notifyNewEvents: r.notifyNewEvents,
  };
}

/** Fields accepted by the create/update mutations, in UI naming. */
export type MemberWrite = Partial<{
  first: string;
  last: string;
  email: string;
  goal: number | null;
  active: boolean;
  role: string;
  since: string;
}>;

function toRawWrite(data: MemberWrite): Record<string, unknown> {
  const body: Record<string, unknown> = {};
  if (data.first !== undefined) body.firstName = data.first;
  if (data.last !== undefined) body.lastName = data.last;
  if (data.email !== undefined) body.email = data.email;
  if (data.goal !== undefined) body.individualGoalHours = data.goal;
  if (data.active !== undefined) body.isActive = data.active;
  if (data.role !== undefined) body.role = data.role;
  if (data.since !== undefined) body.joinedAt = data.since;
  return body;
}

// ── Queries ────────────────────────────────────────────────────────────────

/** True for an account that registered itself via OIDC and still waits for
 * activation — inactive, but (unlike a deactivated member) without a leave date. */
export function isPendingApproval(m: Pick<Member, 'active' | 'left'>): boolean {
  return m.active === false && !m.left;
}

/** All members (unpaginated) — for name lookups and pickers. Pass
 * enabled=false where the list is not needed, sparing the full download. */
export function useMembers(enabled = true) {
  return useQuery({
    queryKey: ['members'],
    // limit=0 lifts the backend's default page size of 50.
    queryFn: () => apiGet<RawMember[]>('/members', { params: { limit: 0 } }).then((rows) => rows.map(toMember)),
    enabled,
  });
}

export type MemberStatus = 'active' | 'pending' | 'left' | 'approved';

export interface MemberPageParams {
  page: number;      // 0-based
  pageSize: number;
  search?: string;
  status?: MemberStatus;
}

export interface MemberPage {
  items: Member[];
  total: number;
}

/** One page of members, sorted by last/first name; total comes from X-Total-Count. */
export function useMemberPage({ page, pageSize, search = '', status }: MemberPageParams) {
  return useQuery({
    queryKey: ['members', 'page', { page, pageSize, search, status }],
    queryFn: async (): Promise<MemberPage> => {
      const res = await apiClient.get<RawMember[]>('/members', {
        params: { limit: pageSize, offset: page * pageSize, search: search || undefined, status },
      });
      const items = res.data.map(toMember);
      const total = Number(res.headers['x-total-count']);
      return { items, total: Number.isFinite(total) ? total : items.length };
    },
    // Keep the current page on screen while the next one loads.
    placeholderData: keepPreviousData,
  });
}

/** Self-registered accounts waiting for activation (usually only a handful). */
export function usePendingMembers(enabled = true) {
  return useQuery({
    queryKey: ['members', 'pending'],
    queryFn: () =>
      apiGet<RawMember[]>('/members', { params: { status: 'pending', limit: 0 } }).then((rows) => rows.map(toMember)),
    enabled,
    // New registrations arrive without any action in this tab; poll gently.
    refetchInterval: 5 * 60 * 1000,
  });
}

export interface MemberName {
  id: string;
  name: string;
}

/** Name-only member search for assigning helpers (Veranstaltungsleiter+,
 * ML-014); runs once the query has at least two characters. `exclude` (the
 * helpers already on the shift) is dropped server-side before the result
 * limit, so they don't crowd out other matches. */
export function useMemberLookup(query: string, exclude: string[] = []) {
  const q = query.trim();
  return useQuery({
    queryKey: ['members', 'lookup', q, exclude],
    queryFn: () =>
      apiGet<MemberName[]>('/members/lookup', {
        params: { q, exclude },
        // Repeated keys (exclude=a&exclude=b), as the API expects.
        paramsSerializer: { indexes: null },
      }),
    enabled: q.length >= 2,
    placeholderData: keepPreviousData,
  });
}

export function useMember(id: string) {
  return useQuery({
    queryKey: ['members', id],
    queryFn: () => apiGet<RawMember>(`/members/${id}`).then(toMember),
    enabled: !!id,
  });
}

// ── Mutations ──────────────────────────────────────────────────────────────

export function useCreateMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: MemberWrite) => apiPost<RawMember>('/members', toRawWrite(data)).then(toMember),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['members'] }),
  });
}

export function useUpdateMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: MemberWrite & { id: string }) =>
      apiPut<RawMember>(`/members/${id}`, toRawWrite(data)).then(toMember),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['members'] });
      qc.invalidateQueries({ queryKey: ['members', vars.id] });
    },
  });
}

export function useDeactivateMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiDelete<void>(`/members/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['members'] }),
  });
}

export function useImportMembersCSV() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => {
      const form = new FormData();
      form.append('file', file);
      return apiClient.post<{ imported: number }>('/members/import', form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      }).then((r) => r.data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['members'] }),
  });
}

export interface ImportPreviewResult {
  toCreate: Array<{ email: string; firstName: string; lastName: string }>;
  toUpdate: Array<{ email: string; firstName: string; lastName: string; changes: string[] }>;
  unchanged: number;
}

export function useImportMembersPreview() {
  return useMutation({
    mutationFn: async (file: File): Promise<ImportPreviewResult> => {
      const form = new FormData();
      form.append('file', file);
      const res = await apiClient.post<ImportPreviewResult>('/members/import?preview=true', form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
      return res.data;
    },
  });
}

export function useExportMembersCSV() {
  return useMutation({
    mutationFn: async () => {
      const res = await apiClient.get('/members/export', { responseType: 'blob' });
      const url = window.URL.createObjectURL(new Blob([res.data as BlobPart]));
      const a = document.createElement('a');
      a.href = url;
      a.download = 'mitglieder.csv';
      a.click();
      window.URL.revokeObjectURL(url);
    },
  });
}

// ── Privacy / GDPR (DS-003, DS-004, N-001) ──────────────────────────────────

/**
 * Updates the authenticated member's own preferences: reminder opt-out
 * (N-001) and the "new event published" opt-in. The backend expects both
 * fields on every write, so callers must pass the full desired state (not
 * just the one being changed) to avoid clobbering the other.
 */
export function useUpdatePreferences() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (prefs: { reminderOptOut: boolean; notifyNewEvents: boolean }) =>
      apiPut<{ reminderOptOut: boolean; notifyNewEvents: boolean }>('/members/me/preferences', prefs),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['members'] }),
  });
}

/**
 * Triggers a browser download of a member's GDPR data export as JSON (DS-003).
 * A member may export their own data; Vorstand+ may export any member's.
 */
export function useExportMemberData() {
  return useMutation({
    mutationFn: async (id: string) => {
      const data = await apiGet<unknown>(`/members/${id}/export-data`);
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `mitglied-${id}-export.json`;
      a.click();
      window.URL.revokeObjectURL(url);
    },
  });
}

/** Anonymizes a member (DS-004, Vorstand+ only). */
export function useGdprDelete() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiPost<{ status: string }>(`/members/${id}/gdpr-delete`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['members'] }),
  });
}
