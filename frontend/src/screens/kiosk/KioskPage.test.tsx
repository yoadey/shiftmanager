import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { AxiosError, AxiosHeaders } from 'axios';
import type { Event } from '@/types';
import type { KioskRegisterPayload } from '@/api/kiosk';

const axiosError = (status: number) =>
  new AxiosError('x', String(status), undefined, undefined, { status, statusText: '', data: {}, headers: {}, config: { headers: new AxiosHeaders() } });

let configState: { data?: object; error?: unknown } = {};
const registerCalls: KioskRegisterPayload[] = [];
let registerError: unknown = null;

const event: Event = {
  id: 'e1', name: 'Sommerfest', category: 'Vereinsleben', location: 'Vereinsheim', status: 'published', description: '',
  days: [{ date: '2026-11-12', shifts: [{ id: 's1', name: 'Theke', start: '10:00', end: '14:00', min: 2, max: 4, signups: [] }] }],
};

vi.mock('@/api/kiosk', () => ({
  useKioskConfig: () => ({ data: configState.data, error: configState.error }),
  useKioskEvents: () => ({ data: configState.data ? [event] : undefined, isLoading: false }),
  useKioskMemberSearch: () => ({ data: [{ id: 'm1', first: 'Max M.', last: '', email: '', since: '', goal: null }] }),
  useKioskRegister: () => ({
    isPending: false,
    mutate: (p: KioskRegisterPayload, o: { onSuccess: () => void; onError: (e: unknown) => void }) => {
      registerCalls.push(p);
      if (registerError) o.onError(registerError);
      else o.onSuccess();
    },
  }),
}));

import { KioskPage } from './KioskPage';
import { getKioskToken } from '@/api/client';

function toIdentify() {
  fireEvent.click(screen.getByText('Sommerfest'));
  fireEvent.click(screen.getByText('Theke'));
}

describe('KioskPage', () => {
  beforeEach(() => {
    localStorage.clear();
    localStorage.setItem('sm_kiosk_token', 'tok');
    configState = { data: { kioskSearch: false, clubName: 'TSC' } };
    registerCalls.length = 0;
    registerError = null;
    window.history.replaceState(null, '', '/kiosk');
  });

  it('keeps the token from the kiosk link and drops it from the address bar', () => {
    localStorage.clear();
    window.history.replaceState(null, '', '/kiosk?token=abc123');
    render(<KioskPage />);
    expect(getKioskToken()).toBe('abc123');
    expect(window.location.search).toBe('');
  });

  it('asks for a kiosk link without a valid token', () => {
    localStorage.clear();
    configState = { error: axiosError(401) };
    render(<KioskPage />);
    expect(screen.getByText('Kiosk nicht freigeschaltet')).toBeInTheDocument();
    expect(screen.queryByText('Sommerfest')).not.toBeInTheDocument();
  });

  it('shows the lock notice while the kiosk is locked', () => {
    configState = { error: axiosError(403) };
    render(<KioskPage />);
    expect(screen.getByText('Kiosk gesperrt')).toBeInTheDocument();
  });

  it('registers by full name (K-014)', () => {
    render(<KioskPage />);
    toIdentify();
    expect(screen.queryByRole('tab', { name: /Mitglied suchen/ })).not.toBeInTheDocument();
    const btn = screen.getByRole('button', { name: /Eintragen/ });
    expect(btn).toBeDisabled();
    fireEvent.change(screen.getByLabelText('Vor- und Nachname'), { target: { value: '  Maximilian   Mustermann ' } });
    fireEvent.click(btn);
    expect(registerCalls).toEqual([{ shiftId: 's1', name: 'Maximilian Mustermann' }]);
    expect(screen.getByText('Eintragung gespeichert')).toBeInTheDocument();
    expect(screen.getByText(/ordnet die Eintragung deinem Mitgliedskonto zu/)).toBeInTheDocument();
  });

  it('registers with e-mail and display name, without revealing whether the address is known (K-005)', () => {
    render(<KioskPage />);
    toIdentify();
    fireEvent.click(screen.getByRole('tab', { name: /Mit E-Mail/ }));
    fireEvent.change(screen.getByLabelText('Anzeigename'), { target: { value: 'Maxi' } });
    fireEvent.change(screen.getByLabelText('E-Mail-Adresse'), { target: { value: 'max@example.org' } });
    fireEvent.click(screen.getByRole('button', { name: /Eintragen und Link senden/ }));
    expect(registerCalls).toEqual([{ shiftId: 's1', name: 'Maxi', email: 'max@example.org' }]);
    expect(screen.getByText(/Link, den wir an die angegebene Adresse geschickt haben/)).toBeInTheDocument();
  });

  it('offers the member search only when enabled', () => {
    configState = { data: { kioskSearch: true } };
    render(<KioskPage />);
    toIdentify();
    fireEvent.click(screen.getByRole('tab', { name: /Mitglied suchen/ }));
    fireEvent.click(screen.getByText('Max M.'));
    expect(registerCalls).toEqual([{ shiftId: 's1', memberId: 'm1' }]);
  });

  it('says so when the registration fails instead of claiming success', () => {
    registerError = axiosError(409);
    render(<KioskPage />);
    toIdentify();
    fireEvent.change(screen.getByLabelText('Vor- und Nachname'), { target: { value: 'Max Mustermann' } });
    fireEvent.click(screen.getByRole('button', { name: /Eintragen/ }));
    expect(screen.getByRole('alert')).toHaveTextContent('bereits eingetragen');
    expect(screen.queryByText('Eintragung gespeichert')).not.toBeInTheDocument();
  });
});
