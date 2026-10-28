import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { useAuthStore } from '@/store/auth.store';
import type { Member } from '@/types';
import type { MemberPageParams } from '@/api/members';

const m = (id: string, extra: Partial<Member> = {}): Member => ({
  id, first: 'Vor' + id, last: 'Nach' + id, email: `${id}@x.de`, since: '2026-10-01', goal: null, active: true, left: null, ...extra,
});

let pageItems: Member[] = [];
let pageTotal = 0;
let pending: Member[] = [];
let pendingError = false;
const pageCalls: MemberPageParams[] = [];
const approve = vi.fn();
const reject = vi.fn();
const idle = () => ({ mutate: vi.fn(), isPending: false });

vi.mock('@/api/members', () => ({
  useMemberPage: (p: MemberPageParams) => {
    pageCalls.push(p);
    return { data: { items: pageItems, total: pageTotal }, isLoading: false, isError: false };
  },
  usePendingMembers: () => ({ data: pendingError ? undefined : pending, isError: pendingError }),
  useUpdateMember: () => ({ mutate: approve, isPending: false, variables: undefined }),
  useDeactivateMember: () => ({ mutate: reject, isPending: false }),
  useCreateMember: () => idle(),
  useExportMembersCSV: () => idle(),
  useImportMembersCSV: () => idle(),
  useImportMembersPreview: () => idle(),
}));
vi.mock('@/api/events', () => ({ useEvents: () => ({ data: [] }) }));
vi.mock('@/api/settings', () => ({ useSettings: () => ({ data: { yearGoal: 20 } }) }));

import { AdminMembers, MEMBERS_PAGE_SIZE } from './AdminMembers';

function renderAt(path = '/mitglieder') {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <AdminMembers />
    </MemoryRouter>,
  );
}

describe('AdminMembers', () => {
  beforeEach(() => {
    useAuthStore.setState({ user: { id: 'me', name: 'Me', email: 'me@x.de', role: 'admin' } });
    pageItems = [m('1'), m('2', { active: false, left: '2026-01-01' })];
    pageTotal = 2;
    pending = [];
    pendingError = false;
    pageCalls.length = 0;
    approve.mockClear();
    reject.mockClear();
  });

  it('lists pending registrations on top and activates them', () => {
    pending = [m('p', { first: 'Paula', last: 'Neu', active: false })];
    renderAt();

    expect(screen.getByText('Warten auf Freischaltung')).toBeInTheDocument();
    expect(screen.getByText('Paula Neu')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /Freischalten/ }));
    expect(approve).toHaveBeenCalledWith({ id: 'p', active: true }, expect.anything());
  });

  it('asks for confirmation before rejecting a registration', () => {
    pending = [m('p', { first: 'Paula', last: 'Neu', active: false })];
    renderAt();

    fireEvent.click(screen.getByRole('button', { name: 'Ablehnen' }));
    expect(reject).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole('button', { name: /Ja, ablehnen/ }));
    expect(reject).toHaveBeenCalledWith('p', expect.anything());
  });

  it('says so when pending registrations could not be loaded', () => {
    pendingError = true;
    renderAt();
    expect(screen.getByRole('alert')).toHaveTextContent('Wartende Registrierungen konnten nicht geladen werden.');
  });

  it('requests approved members only and marks deactivated ones', () => {
    renderAt();
    expect(pageCalls.at(-1)).toMatchObject({ page: 0, pageSize: MEMBERS_PAGE_SIZE, status: 'approved' });
    expect(screen.queryByText('Warten auf Freischaltung')).not.toBeInTheDocument();
    expect(screen.getByText('Ausgetreten')).toBeInTheDocument();
  });

  it('pages through the list', () => {
    pageTotal = MEMBERS_PAGE_SIZE * 2 + 5;
    renderAt();

    expect(screen.getByText(/Seite 1 von 3/)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Zurück/ })).toBeDisabled();
    fireEvent.click(screen.getByRole('button', { name: /Weiter/ }));
    expect(pageCalls.at(-1)).toMatchObject({ page: 1 });
    expect(screen.getByText(/Seite 2 von 3/)).toBeInTheDocument();
  });

  it('restores the page from the URL', () => {
    pageTotal = MEMBERS_PAGE_SIZE * 2 + 5;
    renderAt('/mitglieder?seite=3');
    expect(pageCalls.at(-1)).toMatchObject({ page: 2 });
    expect(screen.getByRole('button', { name: /Weiter/ })).toBeDisabled();
  });
});
