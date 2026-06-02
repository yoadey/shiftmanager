import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
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
    reminderOptOut: r.reminderOptOut,
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

export function useMembers() {
  return useQuery({
    queryKey: ['members'],
    queryFn: () => apiGet<RawMember[]>('/members').then((rows) => rows.map(toMember)),
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

/** Updates the authenticated member's own reminder opt-out (N-001). */
export function useUpdatePreferences() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (reminderOptOut: boolean) =>
      apiPut<{ reminderOptOut: boolean }>('/members/me/preferences', { reminderOptOut }),
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
