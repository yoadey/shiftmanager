import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';
import type { Event } from '@/types';
import type { EventPageParams } from '@/api/events';

const ev = (id: string): Event => ({
  id, name: 'Event ' + id, category: 'Turnier', location: 'Halle', status: 'published', description: '', days: [],
});

let pageItems: Event[] = [];
let pageTotal = 0;
const pageCalls: EventPageParams[] = [];

vi.mock('@/api/events', () => ({
  useEventPage: (p: EventPageParams) => {
    pageCalls.push(p);
    return { data: { items: pageItems, total: pageTotal }, isLoading: false, isError: false, isPlaceholderData: false };
  },
}));

import { MemberDiscover, DISCOVER_PAGE_SIZE } from './MemberDiscover';

const lastCall = () => pageCalls[pageCalls.length - 1];

function renderDiscover() {
  return render(
    <MemoryRouter>
      <MemberDiscover />
    </MemoryRouter>,
  );
}

describe('MemberDiscover', () => {
  beforeEach(() => {
    vi.useFakeTimers({ toFake: ['Date'] });
    vi.setSystemTime(new Date(2026, 9, 28, 9, 30));
    pageItems = [ev('1'), ev('2')];
    pageTotal = 2;
    pageCalls.length = 0;
  });
  afterEach(() => vi.useRealTimers());

  it('starts every visit filtered from today without an end date', () => {
    renderDiscover();

    expect(screen.getByLabelText('Von')).toHaveValue('2026-10-28');
    expect(screen.getByLabelText('Bis')).toHaveValue('');
    expect(lastCall()).toMatchObject({ page: 0, pageSize: DISCOVER_PAGE_SIZE, from: '2026-10-28', to: '', freeSlots: false });
    expect(screen.getByText('Event 1')).toBeInTheDocument();
  });

  it('shows past events once "Von" is cleared', () => {
    renderDiscover();

    fireEvent.change(screen.getByLabelText('Von'), { target: { value: '' } });
    expect(lastCall()).toMatchObject({ from: '' });
  });

  it('pages through the results and restarts at page 1 on a filter change', () => {
    pageTotal = 25;
    renderDiscover();

    expect(screen.getByText('Seite 1 von 3')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Zurück/ })).toBeDisabled();
    fireEvent.click(screen.getByRole('button', { name: /Weiter/ }));
    expect(lastCall()).toMatchObject({ page: 1 });
    expect(screen.getByText('Seite 2 von 3')).toBeInTheDocument();
    expect(screen.getByText('11–20 von 25')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Turnier' }));
    expect(lastCall()).toMatchObject({ page: 0, category: 'Turnier', freeSlots: false });

    fireEvent.click(screen.getByRole('button', { name: /Freie Plätze/ }));
    expect(lastCall()).toMatchObject({ page: 0, category: undefined, freeSlots: true });
  });

  it('hides the page controls when everything fits on one page', () => {
    renderDiscover();
    expect(screen.queryByRole('navigation', { name: 'Seiten' })).not.toBeInTheDocument();
  });
});
