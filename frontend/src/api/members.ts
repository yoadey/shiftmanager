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
