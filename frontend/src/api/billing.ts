import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiGet, apiPost, apiClient } from './client';
import type { ClubYear } from '@/types';

// ── Club Years ─────────────────────────────────────────────────────────────────

export function useClubYears() {
  return useQuery({
    queryKey: ['club-years'],
    queryFn: () => apiGet<ClubYear[]>('/hours/club-years'),
  });
}

export interface CreateClubYearInput {
  label: string;
  startDate: string;   // ISO datetime string
  endDate: string;
  defaultTargetHours: number;
  setActive: boolean;
}

export function useCreateClubYear() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateClubYearInput) => apiPost<ClubYear>('/hours/club-years', data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['club-years'] });
      qc.invalidateQueries({ queryKey: ['stats'] });
      qc.invalidateQueries({ queryKey: ['billing'] });
    },
  });
}

// ── Types ──────────────────────────────────────────────────────────────────────

export interface TierLine {
  tierPosition: number;
  amountCents: number;
  hours: number;
  subtotalCents: number;
}

export interface BillingMember {
  id: string;
  firstName: string;
  lastName: string;
  email: string;
}

export interface BillingResult {
  memberId: string;
  member: BillingMember;
  yearId: string;
  targetHours: number;
  confirmedHours: number;
  missingHours: number;
  totalCents: number;
  tierBreakdown: TierLine[];
}

export interface YearBillingReport {
  clubYear: ClubYear;
  results: BillingResult[];
  totalCents: number;
  computedAt: string;
}

// ── Queries ────────────────────────────────────────────────────────────────────

export function useBilling(clubYearId: string | undefined) {
  return useQuery({
    queryKey: ['billing', clubYearId],
    queryFn: () => apiGet<YearBillingReport>(`/billing/${clubYearId}`),
    enabled: !!clubYearId,
  });
}

// ── Mutations / Downloads ──────────────────────────────────────────────────────

export function useExportBillingCSV() {
  return useMutation({
    mutationFn: async (clubYearId: string) => {
      const res = await apiClient.get(`/billing/${clubYearId}/export.csv`, {
        responseType: 'blob',
      });
      const url = window.URL.createObjectURL(new Blob([res.data as BlobPart]));
      const a = document.createElement('a');
      a.href = url;
      a.download = 'jahresabrechnung.csv';
      a.click();
      window.URL.revokeObjectURL(url);
    },
  });
}

export function useExportBillingPDF() {
  return useMutation({
    mutationFn: async (clubYearId: string) => {
      const res = await apiClient.get(`/billing/${clubYearId}/export.pdf`, {
        responseType: 'blob',
      });
      const url = window.URL.createObjectURL(new Blob([res.data as BlobPart]));
      const a = document.createElement('a');
      a.href = url;
      a.download = 'jahresabrechnung.pdf';
      a.click();
      window.URL.revokeObjectURL(url);
    },
  });
}
