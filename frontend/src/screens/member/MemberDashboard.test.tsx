import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { useAuthStore } from '@/store/auth.store';
import type { MemberHourAccount } from '@/api/hours';

let account: MemberHourAccount;

vi.mock('@/api/events', () => ({ useEvents: () => ({ data: [], isLoading: false, isError: false }) }));
vi.mock('@/api/hours', () => ({ useMyHours: () => ({ data: account, isLoading: false, isError: false }) }));
vi.mock('@/api/settings', () => ({
  useSettings: () => ({ data: { yearGoal: 20 } }),
  useBranding: () => ({ data: { clubName: 'TSC' } }),
}));

import { MemberDashboard } from './MemberDashboard';

function renderDashboard() {
  return render(
    <MemoryRouter>
      <MemberDashboard />
    </MemoryRouter>,
  );
}

describe('MemberDashboard', () => {
  beforeEach(() => {
    useAuthStore.setState({ user: { id: 'me', name: 'Max Muster', email: 'me@x.de', role: 'mitglied' } });
    account = { confirmed: 17.5, reserved: 0, goal: 20, entries: [] };
  });

  it('shows what the member would owe without further hours', () => {
    account.openFeeCents = 3250;
    renderDashboard();
    expect(screen.getByTestId('open-fee')).toHaveTextContent(/Ohne weitere Stunden wären 32,50\s€ Abgeltung fällig\./);
  });

  it('shows no fee once nothing is owed or no fee tiers exist', () => {
    account.openFeeCents = 0;
    const { unmount } = renderDashboard();
    expect(screen.queryByTestId('open-fee')).not.toBeInTheDocument();
    unmount();

    delete account.openFeeCents;
    renderDashboard();
    expect(screen.queryByTestId('open-fee')).not.toBeInTheDocument();
  });
});
