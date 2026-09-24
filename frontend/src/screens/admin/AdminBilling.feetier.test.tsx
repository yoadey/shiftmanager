import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import type { ClubYear, FeeTier } from '@/types';

// Regression coverage: a direct link or reload on
// /abrechnungen/:yearId/staffeln must load (and, on save, overwrite) the
// tiers of the URL's year — not whichever year happens to be selected in
// the list (which resets to the active year on a fresh load, since
// `selectedYearId` is component-local state).

const activeYear: ClubYear = {
  id: 'active-year', label: '2026 (aktiv)', startDate: '2026-01-01T00:00:00Z',
  endDate: '2026-12-31T00:00:00Z', defaultTargetHours: 20, isActive: true,
};
const otherYear: ClubYear = {
  id: 'other-year', label: '2025', startDate: '2025-01-01T00:00:00Z',
  endDate: '2025-12-31T00:00:00Z', defaultTargetHours: 20, isActive: false,
};

const activeTiers: FeeTier[] = [{ id: 't1', clubYearId: activeYear.id, position: 1, amountCents: 999900 }];
const otherTiers: FeeTier[] = [{ id: 't2', clubYearId: otherYear.id, position: 1, amountCents: 12300 }];

vi.mock('@/api/billing', () => ({
  useClubYears: () => ({ data: [activeYear, otherYear], isLoading: false, isError: false }),
  useCreateClubYear: () => ({ mutate: vi.fn(), isPending: false }),
  useUpdateClubYear: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteClubYear: () => ({ mutate: vi.fn(), isPending: false }),
  useBilling: () => ({ data: undefined, isLoading: false }),
  useExportBillingCSV: () => ({ mutate: vi.fn(), isPending: false }),
  useExportBillingPDF: () => ({ mutate: vi.fn(), isPending: false }),
  // Mirrors the real per-year query: the returned tiers depend on which
  // club year id the caller asks for.
  useClubYearFeeTiers: (clubYearId?: string) => ({
    data: clubYearId === otherYear.id ? otherTiers : activeTiers,
    isLoading: false,
  }),
  useUpdateClubYearFeeTiers: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('@/api/settings', () => ({
  useSettings: () => ({ data: { feeSchedule: [], yearGoal: 20 } }),
}));

import { AdminBilling } from './AdminBilling';

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/abrechnungen/*" element={<AdminBilling />} />
      </Routes>
    </MemoryRouter>,
  );
}

describe('AdminBilling fee-tier modal deep link', () => {
  it('pre-fills the URL year\'s tiers on a fresh load, not the active year\'s', async () => {
    renderAt(`/abrechnungen/${otherYear.id}/staffeln`);

    // The editable amount is the URL year's tier (123.00), not the active
    // year's (9999.00) — this is the value a no-op save would persist.
    expect(await screen.findByDisplayValue('123.00')).toBeInTheDocument();
    expect(screen.queryByDisplayValue('9999.00')).not.toBeInTheDocument();

    // Modal header names the URL year (appears once for the year-list row,
    // once for the modal subtitle).
    expect(screen.getAllByText(otherYear.label).length).toBeGreaterThanOrEqual(2);
  });
});
