import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter, Routes, Route, useLocation } from 'react-router-dom';
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
        {
          id: 's1', name: 'Aufbau', start: '14:00', end: '18:00', min: 2, max: 5,
          signups: [{ id: 'reg1', memberId: 'm1', status: 'angemeldet' }],
        },
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
  useUploadEventHeaderImage: () => ({ mutate: vi.fn(), isPending: false }),
  useDeleteEventHeaderImage: () => ({ mutate: vi.fn(), isPending: false }),
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

  it('renders event editing as a real page, not inside a Sheet overlay (UX-001, bedienkonzept)', async () => {
    setRole('veranstaltungsleiter');
    renderAt('/events/e1/bearbeiten');
    await screen.findByText('Veranstaltung bearbeiten');
    expect(document.querySelector('.sm-overlay')).not.toBeInTheDocument();
    expect(document.querySelector('.sm-sheet')).not.toBeInTheDocument();
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
    // Two matches: the registrant list's own trigger button (now expanded,
    // since this deep link needs it open — see ShiftRow's `open` state) and
    // the sheet's title.
    expect((await screen.findAllByText('Helfer hinzufügen')).length).toBe(2);
  });
});

// Regression coverage for a second bug: App.tsx keys the screen's
// ErrorBoundary on the URL so it resets when navigating to a genuinely
// different screen, but it used to key on the *full* pathname — including a
// modal's own sub-route. That forced a full remount of EventDetail on every
// click that opens a route-backed sheet, wiping local UI state the sheet
// depends on. "Zeit bearbeiten" was the visible casualty: it lives inside
// the registrant list's expand toggle (local state, not URL-derived), so
// clicking it navigated to /events/:id/zeit/:signupId, which remounted the
// list collapsed again — and the popup that navigation was supposed to open
// never appeared. Reproduced here by wrapping EventDetail the same way
// App.tsx does: a key derived from the first two path segments only.
function AppShellLike() {
  const location = useLocation();
  const navKey = location.pathname.split('/').filter(Boolean).slice(0, 2).join('/') || 'home';
  return (
    <div key={navKey}>
      <EventDetail />
    </div>
  );
}

describe('EventDetail click-through under the app-level remount key', () => {
  it('opens the time-edit sheet after expanding the list and clicking "Zeit bearbeiten"', async () => {
    setRole('veranstaltungsleiter');
    render(
      <MemoryRouter initialEntries={['/events/e1']}>
        <Routes>
          <Route path="/events/:id/*" element={<AppShellLike />} />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(await screen.findByText(/eingetragen/));
    fireEvent.click(await screen.findByTitle('Zeit bearbeiten'));

    // The time picker's "Übernehmen" save button only exists once it's open.
    expect(await screen.findByText('Übernehmen')).toBeInTheDocument();
  });

  it('keeps the registrant list expanded after closing the time-edit sheet', async () => {
    setRole('veranstaltungsleiter');
    render(
      <MemoryRouter initialEntries={['/events/e1']}>
        <Routes>
          <Route path="/events/:id/*" element={<AppShellLike />} />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(await screen.findByText(/eingetragen/));
    fireEvent.click(await screen.findByTitle('Zeit bearbeiten'));
    fireEvent.click(await screen.findByText('Abbrechen'));

    // Closing the sheet navigates back to /events/e1. If that remounts the
    // shift row (the old, full-pathname remount key), the registrant list's
    // local expand state is lost and collapses — even though the user never
    // asked for that.
    expect(screen.queryByText('Übernehmen')).not.toBeInTheDocument();
    expect(await screen.findByTitle('Zeit bearbeiten')).toBeInTheDocument();
  });
});

describe('EventDetail description and header image (V-009, V-010)', () => {
  it('renders the Markdown description formatted, not as raw text', async () => {
    setRole('mitglied');
    fixtureEvent.description = 'Bitte **pünktlich** da sein.';
    fixtureEvent.headerImageUrl = undefined;
    renderAt('/events/e1');
    await screen.findByText('Sommerfest');
    expect(screen.queryByText('Bitte **pünktlich** da sein.')).not.toBeInTheDocument();
    expect(screen.getByText('pünktlich').tagName).toBe('STRONG');
  });

  it('shows the header image in the hero when set', async () => {
    setRole('mitglied');
    fixtureEvent.description = 'Das jährliche Sommerfest';
    fixtureEvent.headerImageUrl = '/uploads/event-header-abc.png';
    const { container } = renderAt('/events/e1');
    await screen.findByText('Sommerfest');
    expect(container.querySelector('.ev-hero img')).toHaveAttribute('src', '/uploads/event-header-abc.png');
    fixtureEvent.headerImageUrl = undefined;
  });

  it('shows no hero image when none is set', async () => {
    setRole('mitglied');
    fixtureEvent.headerImageUrl = undefined;
    const { container } = renderAt('/events/e1');
    await screen.findByText('Sommerfest');
    expect(container.querySelector('.ev-hero img')).toBeNull();
  });
});
