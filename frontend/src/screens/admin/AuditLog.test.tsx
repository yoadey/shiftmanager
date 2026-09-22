import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { AuditEntry } from '@/types';

const fixture: AuditEntry[] = [
  { ts: '30.05.2026 14:02', who: 'Anna Admin', what: 'Jahresziel auf 40 h geändert', cat: 'Einstellungen' },
  { ts: '29.05.2026 09:15', who: 'Bernd Board', what: 'Veranstaltung „Sommerfest“ veröffentlicht', cat: 'Veranstaltung' },
];

// Mock the API module so the real export returns our fixture (read api/settings.ts).
vi.mock('@/api/settings', () => ({
  useAuditLog: () => ({ data: fixture, isLoading: false, isError: false }),
}));

import { AuditLog } from './AuditLog';

function renderWithProviders(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

describe('AuditLog screen', () => {
  it('renders the heading and audit entries from the mocked hook', () => {
    renderWithProviders(<AuditLog />);

    // Key German UI text from the header
    expect(screen.getByText('Audit-Log')).toBeInTheDocument();

    // Entries from the fixture
    expect(screen.getByText('Jahresziel auf 40 h geändert')).toBeInTheDocument();
    expect(screen.getByText('Veranstaltung „Sommerfest“ veröffentlicht')).toBeInTheDocument();
    expect(screen.getByText('Anna Admin')).toBeInTheDocument();
    expect(screen.getByText('Einstellungen')).toBeInTheDocument();
  });
});
