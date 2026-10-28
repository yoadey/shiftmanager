import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { useAuthStore } from '@/store/auth.store';
import type { AuthUser, Member } from '@/types';

let pending: Member[] = [];
const usePendingMembers = vi.fn((_enabled: boolean) => ({ data: pending }));
vi.mock('@/api/members', () => ({
  usePendingMembers: (enabled: boolean) => usePendingMembers(enabled),
}));

import { PendingMembersBanner } from './PendingMembersBanner';

const member = (id: string): Member => ({ id, first: 'Neu', last: id, email: `${id}@x.de`, since: '2026-10-01', goal: null, active: false, left: null });
const asRole = (role: AuthUser['role']) =>
  useAuthStore.setState({ user: { id: 'u', name: 'U', email: 'u@x.de', role } });

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <PendingMembersBanner />
    </MemoryRouter>,
  );
}

describe('PendingMembersBanner', () => {
  beforeEach(() => {
    pending = [member('a'), member('b')];
    usePendingMembers.mockClear();
  });

  it('tells board members how many registrations wait for activation', () => {
    asRole('admin');
    renderAt('/uebersicht');
    expect(screen.getByText(/2 neue Mitglieder warten auf Freischaltung/)).toBeInTheDocument();
    expect(usePendingMembers).toHaveBeenCalledWith(true);
  });

  it('uses the singular for one registration', () => {
    asRole('vorstand');
    pending = [member('a')];
    renderAt('/start');
    expect(screen.getByText(/1 neues Mitglied wartet auf Freischaltung/)).toBeInTheDocument();
  });

  it('is hidden on the member list, which shows the registrations itself', () => {
    asRole('admin');
    const { container } = renderAt('/mitglieder');
    expect(container).toBeEmptyDOMElement();
  });

  it('is hidden on the create/import dialogs, which render over the member list', () => {
    asRole('admin');
    for (const path of ['/mitglieder/neu', '/mitglieder/import']) {
      const { container, unmount } = renderAt(path);
      expect(container).toBeEmptyDOMElement();
      unmount();
    }
  });

  it('is hidden when nobody is waiting', () => {
    asRole('admin');
    pending = [];
    const { container } = renderAt('/uebersicht');
    expect(container).toBeEmptyDOMElement();
  });

  it('is never shown to (nor fetched for) roles that cannot activate members', () => {
    asRole('veranstaltungsleiter');
    const { container } = renderAt('/uebersicht');
    expect(container).toBeEmptyDOMElement();
    expect(usePendingMembers).toHaveBeenCalledWith(false);
  });
});
