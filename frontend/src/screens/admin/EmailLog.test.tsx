import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { EmailLogEntry } from '@/types';

const resendMutate = vi.fn();

const fixture: EmailLogEntry[] = [
  { id: 'e1', to: 'a@example.de', template: 'reminder-1d', subject: 'Erinnerung', body: '', status: 'sent', error: '', createdAt: '2026-05-30T08:00:00Z' },
  { id: 'e2', to: 'b@example.de', template: 'shift-confirmation', subject: 'Bestätigung', body: '', status: 'failed', error: 'SMTP timeout', createdAt: '2026-05-30T09:00:00Z' },
];

vi.mock('@/api/settings', () => ({
  useEmailLog: () => ({ data: fixture, isLoading: false, isError: false }),
  useResendEmail: () => ({ mutate: resendMutate, isPending: false }),
}));

import { EmailLog } from './EmailLog';

function renderWithProviders(ui: React.ReactElement) {
  const qc = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(<QueryClientProvider client={qc}>{ui}</QueryClientProvider>);
}

describe('EmailLog screen', () => {
  beforeEach(() => resendMutate.mockClear());

  it('renders entries with status and error text', () => {
    renderWithProviders(<EmailLog />);
    expect(screen.getByText('E-Mail-Protokoll')).toBeInTheDocument();
    expect(screen.getByText('a@example.de')).toBeInTheDocument();
    expect(screen.getByText('SMTP timeout')).toBeInTheDocument();
    expect(screen.getByText('Fehler')).toBeInTheDocument();
    expect(screen.getByText('Gesendet')).toBeInTheDocument();
  });

  it('shows the resend button only on failed rows and resends by id', () => {
    renderWithProviders(<EmailLog />);
    const buttons = screen.getAllByText('Erneut senden');
    // Only the failed entry (e2) gets a resend button.
    expect(buttons).toHaveLength(1);
    fireEvent.click(buttons[0]);
    expect(resendMutate).toHaveBeenCalledTimes(1);
    expect(resendMutate.mock.calls[0][0]).toBe('e2');
  });
});
