import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiGet, apiPut, apiPost, apiClient } from './client';
import type {
  AppSettings,
  BrandingConfig,
  BrandingUpdateResult,
  FeeTier,
  AuditEntry,
  EmailTemplate,
  EmailLogEntry,
  MemberFeeTier,
} from '@/types';

// ── Queries ────────────────────────────────────────────────────────────────

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: () => apiGet<AppSettings>('/settings'),
  });
}

export function useBranding() {
  return useQuery({
    queryKey: ['branding'],
    queryFn: () => apiGet<BrandingConfig>('/settings/branding'),
  });
}

export function useFeeTiers() {
  return useQuery({
    queryKey: ['fee-tiers'],
    queryFn: () => apiGet<FeeTier[]>('/settings/fee-tiers'),
  });
}

export function useAuditLog() {
  return useQuery({
    queryKey: ['audit-log'],
    queryFn: () => apiGet<AuditEntry[]>('/audit'),
  });
}

// ── Mutations ──────────────────────────────────────────────────────────────

export function useUpdateSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<AppSettings>) => apiPut<AppSettings>('/settings', data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  });
}

/**
 * Persists branding. The backend returns `{ branding, warnings }`; the warnings
 * array surfaces WCAG contrast issues (B-003) without rejecting the update.
 */
export function useUpdateBranding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<BrandingConfig>) =>
      apiPut<BrandingUpdateResult>('/settings/branding', data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['branding'] }),
  });
}

export function useUpdateFeeTiers() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: FeeTier[]) => apiPut<FeeTier[]>('/settings/fee-tiers', data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['fee-tiers'] }),
  });
}

// ── Logo upload (B-004) ─────────────────────────────────────────────────────

/**
 * Uploads a PNG/SVG club logo via multipart FormData. We deliberately do not set
 * a JSON content-type — axios infers the multipart boundary from the FormData.
 */
export function useUploadLogo() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => {
      const form = new FormData();
      form.append('file', file);
      return apiClient
        .post<BrandingConfig>('/settings/logo', form)
        .then((r) => r.data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['branding'] }),
  });
}

// ── Email templates (Section 4) ─────────────────────────────────────────────

export function useEmailTemplates() {
  return useQuery({
    queryKey: ['email-templates'],
    queryFn: () => apiGet<EmailTemplate[]>('/settings/email-templates'),
  });
}

export function useUpdateEmailTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ name, subject, body }: { name: string; subject: string; body: string }) =>
      apiPut<EmailTemplate>(`/settings/email-templates/${encodeURIComponent(name)}`, { subject, body }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['email-templates'] }),
  });
}

// ── Email log + resend (N-004) ──────────────────────────────────────────────

export function useEmailLog() {
  return useQuery({
    queryKey: ['email-log'],
    queryFn: () => apiGet<EmailLogEntry[]>('/settings/email-log'),
  });
}

export function useResendEmail() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => apiPost<{ status: string }>(`/settings/email-log/${id}/resend`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['email-log'] }),
  });
}

// ── Per-member fee-tier overrides (G-004) ───────────────────────────────────

export function useMemberFeeTiers(memberId: string, clubYearId: string) {
  return useQuery({
    queryKey: ['member-fee-tiers', memberId, clubYearId],
    queryFn: () =>
      apiGet<MemberFeeTier[]>(
        `/settings/members/${memberId}/fee-tiers?clubYearId=${encodeURIComponent(clubYearId)}`,
      ),
    enabled: !!memberId && !!clubYearId,
  });
}

export function useUpdateMemberFeeTiers() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      memberId,
      clubYearId,
      tiers,
    }: {
      memberId: string;
      clubYearId: string;
      tiers: MemberFeeTier[];
    }) =>
      apiPut<MemberFeeTier[]>(
        `/settings/members/${memberId}/fee-tiers?clubYearId=${encodeURIComponent(clubYearId)}`,
        { tiers },
      ),
    onSuccess: (_d, vars) =>
      qc.invalidateQueries({ queryKey: ['member-fee-tiers', vars.memberId, vars.clubYearId] }),
  });
}
