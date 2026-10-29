import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import type { KioskLink } from '@/api/settings';

let links: KioskLink[] = [];
const createMutate = vi.fn();
const revokeMutate = vi.fn();

vi.mock('@/api/settings', async (orig) => ({
  ...(await orig<typeof import('@/api/settings')>()),
  useKioskLinks: () => ({ data: links }),
  useCreateKioskLink: () => ({ mutate: createMutate, isPending: false }),
  useRevokeKioskLink: () => ({ mutate: revokeMutate }),
}));

import { KioskLinks } from './KioskLinks';

describe('KioskLinks', () => {
  beforeEach(() => {
    createMutate.mockReset();
    revokeMutate.mockReset();
    links = [
      { id: 'l1', label: 'Tablet Vereinsheim', createdAt: '2026-10-01T10:00:00Z', expiresAt: '2099-01-01T00:00:00Z' },
      { id: 'l2', label: 'Alt', createdAt: '2025-01-01T10:00:00Z', expiresAt: '2025-02-01T00:00:00Z' },
    ];
  });

  it('creates a link with the chosen validity and shows it once', () => {
    createMutate.mockImplementation((_v, o) => o.onSuccess({ id: 'l3', label: 'Halle', createdAt: '', expiresAt: '2027-01-27T00:00:00Z', token: 'secret-token' }));
    render(<KioskLinks />);
    fireEvent.change(screen.getByLabelText('Bezeichnung'), { target: { value: ' Halle ' } });
    fireEvent.click(screen.getByRole('button', { name: '90 Tage' }));
    fireEvent.click(screen.getByRole('button', { name: /Kiosk-Link erstellen/ }));
    expect(createMutate).toHaveBeenCalledWith({ label: 'Halle', validDays: 90 }, expect.anything());
    expect(screen.getByTestId('kiosk-link-url')).toHaveTextContent('/kiosk?token=secret-token');
  });

  it('lists active links and revokes after confirmation', () => {
    render(<KioskLinks />);
    expect(screen.getByText('Tablet Vereinsheim')).toBeInTheDocument();
    expect(screen.queryByText('Alt')).not.toBeInTheDocument();
    expect(screen.getByText('1 abgelaufene oder widerrufene Link')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Widerrufen' }));
    expect(revokeMutate).not.toHaveBeenCalled();
    fireEvent.click(screen.getAllByRole('button', { name: 'Widerrufen' }).at(-1)!);
    expect(revokeMutate).toHaveBeenCalledWith('l1', expect.anything());
  });

  it('rejects a validity outside 1 to 3650 days', () => {
    render(<KioskLinks />);
    fireEvent.change(screen.getByLabelText('Gültig (Tage)'), { target: { value: '0' } });
    expect(screen.getByRole('button', { name: /Kiosk-Link erstellen/ })).toBeDisabled();
  });
});
