import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiGet, apiPut } from './client';
import type { AppSettings, BrandingConfig, FeeTier, AuditEntry } from '@/types';

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

export function useUpdateBranding() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<BrandingConfig>) => apiPut<BrandingConfig>('/settings/branding', data),
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
