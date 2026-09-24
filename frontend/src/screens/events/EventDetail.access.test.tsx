import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route } from 'react-router-dom';
import { useAuthStore } from '@/store/auth.store';
import type { Event, AuthUser } from '@/types';

// Regression coverage for a URL-only access-control bug: EditEventSheet and
// AddMemberSheet were derived purely from the route's splat segment, with no
// isBoard check, so a plain member navigating straight to
// /events/:id/bearbeiten or /events/:id/helfer/:shiftId got the board-only
// sheet even though the buttons that normally open them are isBoard-gated.

const fixtureEvent: Event = {
  id: 'e1',
  name: 'Sommerfest',
  category: 'Vereinsleben',
  location: 'Vereinsheim',
  status: 'veröffentlicht',
  description: 'Das jährliche Sommerfest',
  days: [
    {
      date: '2026-07-04',
      shifts: [
        { id: 's1', name: 'Aufbau', start: '14:00', end: '18:00', min: 2, max: 5, signups: [] },
      ],
    },
  ],
};

vi.mock('@/api/events', () => ({
  useEvent: () => ({ data: fixtureEvent, isLoading: false, isError: false }),
  useEventTimeline: () => ({ data: { event: fixtureEvent, days: fixtureEvent.days }, isLoading: false, isError: false }),
  useUpdateEvent: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
  useDeleteEvent: () => ({ mutate: vi.fn(), isPending: false }),
  useCompleteEvent: () => ({ mutate: vi.fn(), isPending: false }),
  useEventAttachments: () => ({ data: [] }),
  useUploadEventAttachment: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteEventAttachment: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('@/api/shifts', () => ({
  useRegisterShift: () => ({ mutate: vi.fn() }),
  useDeregisterShift: () => ({ mutate: vi.fn() }),
  useCreateShift: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useUpdateShift: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteShift: () => ({ mutateAsync: vi.fn(), isPending: false }),
  usePatchRegistration: () => ({ mutate: vi.fn() }),
  useForceDeleteRegistration: () => ({ mutate: vi.fn() }),
  useAddMemberToShift: () => ({ mutate: vi.fn() }),
  useAddGuestToShift: () => ({ mutate: vi.fn() }),
}));

vi.mock('@/api/members', () => ({
  useMembers: () => ({ data: [] }),
}));

vi.mock('@/api/settings', () => ({
  useSettings: () => ({ data: { reservationHours: 48, deregisterDeadlineH: 24 } }),
}));

import { EventDetail } from './EventDetail';

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/events/:id/*" element={<EventDetail />} />
      </Routes>
    </MemoryRouter>,
  );
}

function setRole(role: AuthUser['role']) {
  useAuthStore.setState({
    user: { id: 'u1', name: 'Test User', email: 't@test.local', role, first: 'Test', last: 'User' },
  });
}

describe('EventDetail board-only sheets via direct URL', () => {
  it('does not open the edit sheet for a plain member deep-linking to /bearbeiten', async () => {
    setRole('mitglied');
    renderAt('/events/e1/bearbeiten');
    await screen.findByText('Sommerfest');
    expect(screen.queryByText('Veranstaltung bearbeiten')).not.toBeInTheDocument();
  });

  it('opens the edit sheet for an event manager at the same URL', async () => {
    setRole('veranstaltungsleiter');
    renderAt('/events/e1/bearbeiten');
    expect(await screen.findByText('Veranstaltung bearbeiten')).toBeInTheDocument();
  });

  it('does not open the add-helper sheet for a plain member deep-linking to /helfer/:shiftId', async () => {
    setRole('mitglied');
    renderAt('/events/e1/helfer/s1');
    await screen.findByText('Sommerfest');
    expect(screen.queryByText('Helfer hinzufügen')).not.toBeInTheDocument();
  });

  it('opens the add-helper sheet for an event manager at the same URL', async () => {
    setRole('veranstaltungsleiter');
    renderAt('/events/e1/helfer/s1');
    expect(await screen.findByText('Helfer hinzufügen')).toBeInTheDocument();
  });
});
