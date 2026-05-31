import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiGet, apiPost, apiPut, apiPatch, apiClient } from './client';
import type { Member } from '@/types';

// ── Queries ────────────────────────────────────────────────────────────────

export function useMembers() {
  return useQuery({
    queryKey: ['members'],
    queryFn: () => apiGet<Member[]>('/members'),
  });
}

export function useMember(id: string) {
  return useQuery({
    queryKey: ['members', id],
    queryFn: () => apiGet<Member>(`/members/${id}`),
    enabled: !!id,
  });
}

// ── Mutations ──────────────────────────────────────────────────────────────

export function useCreateMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Omit<Member, 'id'>) => apiPost<Member>('/members', data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['members'] }),
  });
}

export function useUpdateMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, ...data }: Partial<Member> & { id: string }) =>
      apiPut<Member>(`/members/${id}`, data),
    onSuccess: (_d, vars) => {
      qc.invalidateQueries({ queryKey: ['members'] });
      qc.invalidateQueries({ queryKey: ['members', vars.id] });
    },
  });
}

export function useDeactivateMember() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiPatch<Member>(`/members/${id}/deactivate`, {}),
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
