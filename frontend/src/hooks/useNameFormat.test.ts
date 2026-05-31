import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import { useNameFormat } from './useNameFormat';
import { useAppStore } from '@/store/app.store';
import { useAuthStore } from '@/store/auth.store';
import type { Member, AuthUser } from '@/types';

const max: Member = {
  id: 'm1',
  first: 'Maximilian',
  last: 'Müller',
  email: 'max@example.com',
  since: '2020-01-01',
  goal: null,
};

const self: AuthUser = {
  id: 'me',
  name: 'Self Tester',
  email: 'me@example.com',
  role: 'mitglied',
};

const selfMember: Member = {
  id: 'me',
  first: 'Self',
  last: 'Tester',
  email: 'me@example.com',
  since: '2020-01-01',
  goal: null,
};

describe('useNameFormat', () => {
  beforeEach(() => {
    // Default: a member viewer, logged in as `self`.
    useAuthStore.setState({ token: 't', user: self });
    useAppStore.setState({ role: 'mitglied' });
  });

  it('abbrev mode → "Maximilian M."', () => {
    const { result } = renderHook(() => useNameFormat());
    expect(result.current(max)).toBe('Maximilian M.');
  });

  it('explicit full mode → full name', () => {
    const { result } = renderHook(() => useNameFormat());
    expect(result.current(max, { mode: 'full' })).toBe('Maximilian Müller');
  });

  it('own name always shown in full', () => {
    const { result } = renderHook(() => useNameFormat());
    expect(result.current(selfMember)).toBe('Self Tester');
  });

  it('board (vorstand) viewer sees full name', () => {
    useAppStore.setState({ role: 'vorstand' });
    const { result } = renderHook(() => useNameFormat());
    expect(result.current(max)).toBe('Maximilian Müller');
  });

  it('viewerFull option forces full name', () => {
    const { result } = renderHook(() => useNameFormat());
    expect(result.current(max, { viewerFull: true })).toBe('Maximilian Müller');
  });

  it('returns "Unbekannt" for missing member', () => {
    const { result } = renderHook(() => useNameFormat());
    expect(result.current(null)).toBe('Unbekannt');
    expect(result.current(undefined)).toBe('Unbekannt');
  });
});
