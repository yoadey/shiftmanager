import { describe, it, expect, vi, afterEach } from 'vitest';
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
          signups: [{ id: 'reg1', memberId: 'm1', status: 'angemeldet', name: 'Max M.' }],
        },
      ],
    },
  ],
};

let attachmentsData: { id: string; eventId: string; fileName: string; url: string; contentType: string; sizeBytes: number; uploadedAt: string }[] = [];

vi.mock('@/api/events', () => ({
  useEvent: () => ({ data: fixtureEvent, isLoading: false, isError: false }),
  useEventTimeline: () => ({ data: { event: fixtureEvent, days: fixtureEvent.days }, isLoading: false, isError: false }),
  useUpdateEvent: () => ({ mutate: vi.fn(), mutateAsync: vi.fn(), isPending: false }),
  useDeleteEvent: () => ({ mutate: vi.fn(), isPending: false }),
  useCompleteEvent: () => ({ mutate: vi.fn(), isPending: false }),
  useEventAttachments: () => ({ data: attachmentsData }),
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

const lookupCalls: { q: string; exclude: string[] }[] = [];
vi.mock('@/api/members', () => ({
  // Like the server: helpers in `exclude` never come back.
  useMemberLookup: (q: string, exclude: string[] = []) => {
    lookupCalls.push({ q, exclude });
    const all = [{ id: 'm1', name: 'Max Mustermann' }, { id: 'm2', name: 'Maria Musterfrau' }];
    return { data: q.length >= 2 ? all.filter((m) => !exclude.includes(m.id)) : undefined, isFetching: false };
  },
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

// Members see only the helpers' names per shift (SC-011); organisers search
// helpers by name instead of loading the member list (ML-014).
describe('EventDetail helper names', () => {
  it('shows a plain member the names of the helpers in each shift', async () => {
    setRole('mitglied');
    renderAt('/events/e1');
    expect(await screen.findByTestId('helper-names')).toHaveTextContent('Max M.');
  });

  it('shows the organiser the helpers and "Helfer hinzufügen" right away (SC-012)', async () => {
    setRole('veranstaltungsleiter');
    renderAt('/events/e1');
    expect(await screen.findByTestId('registrations')).toHaveTextContent('Max M.');
    expect(screen.queryByTestId('helper-names')).not.toBeInTheDocument();
    expect(screen.queryByText(/eingetragen/)).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Helfer hinzufügen/ })).toHaveClass('primary');
  });

  it('searches helpers by name and hides those already in the shift', async () => {
    setRole('veranstaltungsleiter');
    lookupCalls.length = 0;
    renderAt('/events/e1/helfer/s1');
    expect(await screen.findByText('Mindestens zwei Buchstaben eingeben.')).toBeInTheDocument();
    fireEvent.change(screen.getByPlaceholderText('Name oder E-Mail…'), { target: { value: 'mus' } });
    expect(await screen.findByText('Maria Musterfrau')).toBeInTheDocument();
    expect(screen.queryByText('Max Mustermann')).not.toBeInTheDocument();
    expect(lookupCalls).toContainEqual({ q: 'mus', exclude: ['m1'] });
  });
});

// Compact event view (V-008, V-011).
describe('EventDetail compact view', () => {
  afterEach(() => {
    attachmentsData = [];
    vi.restoreAllMocks();
  });

  it('shows no summary tiles and no empty attachments section', async () => {
    setRole('veranstaltungsleiter');
    renderAt('/events/e1');
    await screen.findByText('Sommerfest');
    expect(screen.queryByText('Freie Plätze')).not.toBeInTheDocument();
    expect(screen.queryByText('Tag')).not.toBeInTheDocument();
    expect(screen.queryByTestId('attachments')).not.toBeInTheDocument();
    // The single-day badge above the shifts is labelled "Schichten" (not "Programm").
    expect(screen.getByText('Schichten')).toHaveClass('tl-daybadge');
  });

  it('lists existing attachments read-only in the view, editable on the edit page', async () => {
    attachmentsData = [{ id: 'a1', eventId: 'e1', fileName: 'plan.pdf', url: '/uploads/plan.pdf', contentType: 'application/pdf', sizeBytes: 1, uploadedAt: '' }];
    setRole('veranstaltungsleiter');
    const { unmount } = renderAt('/events/e1');
    expect(await screen.findByTestId('attachments')).toHaveTextContent('plan.pdf');
    expect(screen.queryByTitle('Anhang entfernen')).not.toBeInTheDocument();
    expect(screen.queryByText('Bild oder Datei anhängen')).not.toBeInTheDocument();
    unmount();

    renderAt('/events/e1/bearbeiten');
    expect(await screen.findByText('Bild oder Datei anhängen')).toBeInTheDocument();
    expect(screen.getByTitle('Anhang entfernen')).toBeInTheDocument();
  });

  it('collapses a long description to its first lines until expanded', async () => {
    vi.spyOn(HTMLElement.prototype, 'scrollHeight', 'get').mockReturnValue(400);
    setRole('mitglied');
    renderAt('/events/e1');
    const body = await screen.findByTestId('event-description');
    expect(screen.getByRole('region', { name: 'Beschreibung' })).toBeInTheDocument();
    expect(body.style.maxHeight).not.toBe('');

    fireEvent.click(screen.getByRole('button', { name: /Mehr anzeigen/ }));
    expect(body.style.maxHeight).toBe('');
    fireEvent.click(screen.getByRole('button', { name: /Weniger anzeigen/ }));
    expect(body.style.maxHeight).not.toBe('');
  });

  it('shows a short description in full without a toggle', async () => {
    setRole('mitglied');
    renderAt('/events/e1');
    const body = await screen.findByTestId('event-description');
    expect(body.style.maxHeight).toBe('');
    expect(screen.queryByRole('button', { name: /Mehr anzeigen/ })).not.toBeInTheDocument();
  });
});

// Regression coverage for a second bug: App.tsx keys the screen's
// ErrorBoundary on the URL so it resets when navigating to a genuinely
// different screen, but it used to key on the *full* pathname — including a
// modal's own sub-route. That forced a full remount of EventDetail on every
// click that opens a route-backed sheet, wiping local UI state the sheet
// depends on. "Zeit bearbeiten" was the visible casualty: it then lived
// inside the registrant list's expand toggle (local state, not URL-derived),
// so clicking it navigated to /events/:id/zeit/:signupId, which remounted
// the list collapsed again — and the popup never appeared. The list is now
// always visible, but the tests still guard against the remount. Reproduced
// by wrapping EventDetail the same way App.tsx does: a key derived from the
// first two path segments only.
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

    fireEvent.click(await screen.findByTitle('Zeit bearbeiten'));

    // The time picker's "Übernehmen" save button only exists once it's open.
    expect(await screen.findByText('Übernehmen')).toBeInTheDocument();
  });

  it('keeps the registrant list visible after closing the time-edit sheet', async () => {
    setRole('veranstaltungsleiter');
    render(
      <MemoryRouter initialEntries={['/events/e1']}>
        <Routes>
          <Route path="/events/:id/*" element={<AppShellLike />} />
        </Routes>
      </MemoryRouter>,
    );

    fireEvent.click(await screen.findByTitle('Zeit bearbeiten'));
    fireEvent.click(await screen.findByText('Abbrechen'));

    // Closing the sheet navigates back to /events/e1; the registrant list
    // must still be there.
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
