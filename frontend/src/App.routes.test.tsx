import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import { Suspense } from 'react';
import { useAuthStore } from '@/store/auth.store';
import type { AuthUser } from '@/types';

// Every screen App.tsx lazy-loads is mocked out to a one-line marker so this
// test exercises only the route table and RequireRole guards in ScreenRoutes
// — not the real screens (which need their own API/query mocking, already
// covered by their own test files where they exist).
vi.mock('@/screens/member/MemberDashboard', () => ({ MemberDashboard: () => <div>screen:start</div> }));
vi.mock('@/screens/member/MemberDiscover', () => ({ MemberDiscover: () => <div>screen:entdecken</div> }));
vi.mock('@/screens/member/MyShifts', () => ({ MyShifts: () => <div>screen:schichten</div> }));
vi.mock('@/screens/member/MemberProfile', () => ({ MemberProfile: () => <div>screen:profil</div> }));
vi.mock('@/screens/events/EventDetail', () => ({
  EventDetail: () => <div>screen:event-detail</div>,
}));
vi.mock('@/screens/admin/AdminDashboard', () => ({ AdminDashboard: () => <div>screen:uebersicht</div> }));
vi.mock('@/screens/admin/AdminEvents', () => ({ AdminEvents: () => <div>screen:events</div> }));
vi.mock('@/screens/admin/AdminMembers', () => ({ AdminMembers: () => <div>screen:mitglieder</div> }));
vi.mock('@/screens/admin/MemberDetail', () => ({ MemberDetail: () => <div>screen:member-detail</div> }));
vi.mock('@/screens/admin/ManualBooking', () => ({ ManualBooking: () => <div>screen:manual-booking</div> }));
vi.mock('@/screens/admin/AdminSettings', () => ({ AdminSettings: () => <div>screen:settings</div> }));
vi.mock('@/screens/admin/AuditLog', () => ({ AuditLog: () => <div>screen:audit</div> }));
vi.mock('@/screens/admin/EmailTemplates', () => ({ EmailTemplates: () => <div>screen:email-templates</div> }));
vi.mock('@/screens/admin/EmailLog', () => ({ EmailLog: () => <div>screen:email-log</div> }));
vi.mock('@/screens/admin/CreateEventFlow', () => ({ CreateEventFlow: () => <div>screen:create-event</div> }));
vi.mock('@/screens/admin/AdminBilling', () => ({ AdminBilling: () => <div>screen:abrechnungen</div> }));

import { ScreenRoutes } from './App';

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Suspense fallback={<div>loading</div>}>
        <ScreenRoutes />
      </Suspense>
    </MemoryRouter>,
  );
}

function setRole(role: AuthUser['role']) {
  useAuthStore.setState({
    user: { id: 'u1', name: 'Test User', email: 't@test.local', role, first: 'Test', last: 'User' },
  });
}

describe('ScreenRoutes route/role matrix', () => {
  beforeEach(() => setRole('mitglied'));

  it('lets a plain member open an event detail page (P1 regression: must not be event-manager-gated)', async () => {
    setRole('mitglied');
    renderAt('/events/e1');
    expect(await screen.findByText('screen:event-detail')).toBeInTheDocument();
  });

  it('lets a plain member open a sub-route of an event (register/helper/edit sheets)', async () => {
    setRole('mitglied');
    renderAt('/events/e1/anmelden/s1');
    expect(await screen.findByText('screen:event-detail')).toBeInTheDocument();
  });

  it('lets an event manager (veranstaltungsleiter) reach the management routes', async () => {
    setRole('veranstaltungsleiter');
    renderAt('/uebersicht');
    expect(await screen.findByText('screen:uebersicht')).toBeInTheDocument();
  });

  it('lets vorstand reach vorstand-only routes', async () => {
    setRole('vorstand');
    renderAt('/mitglieder');
    expect(await screen.findByText('screen:mitglieder')).toBeInTheDocument();
  });

  it('redirects a plain member away from a vorstand-only route to their home screen', async () => {
    setRole('mitglied');
    renderAt('/mitglieder');
    expect(await screen.findByText('screen:start')).toBeInTheDocument();
  });

  it('redirects a plain member away from an event-manager-only route to their home screen', async () => {
    setRole('mitglied');
    renderAt('/uebersicht');
    expect(await screen.findByText('screen:start')).toBeInTheDocument();
  });

  it('redirects an event manager (not vorstand) away from a vorstand-only route to the admin overview', async () => {
    setRole('veranstaltungsleiter');
    renderAt('/abrechnungen');
    expect(await screen.findByText('screen:uebersicht')).toBeInTheDocument();
  });

  it('sends the root path home based on role', async () => {
    setRole('vorstand');
    renderAt('/');
    expect(await screen.findByText('screen:uebersicht')).toBeInTheDocument();
  });
});
