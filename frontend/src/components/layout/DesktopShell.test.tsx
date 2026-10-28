import { describe, it, expect, beforeEach } from 'vitest';
import { render, fireEvent, screen } from '@testing-library/react';
import { MemoryRouter, Routes, Route, useLocation } from 'react-router-dom';
import { useAuthStore } from '@/store/auth.store';
import { DesktopShell } from './DesktopShell';
import type { NavTab } from '@/utils/navTabs';

vi.mock('@/api/settings', () => ({
  useBranding: () => ({ data: undefined }),
}));

const TABS: NavTab[] = [
  { key: 'start', label: 'Start', icon: 'home' },
  { key: 'profil', label: 'Profil', icon: 'user' },
  { key: 'uebersicht', label: 'Übersicht', icon: 'chart', section: 'Verwaltung' },
];

function LocationProbe() {
  const location = useLocation();
  return <span data-testid="path">{location.pathname}</span>;
}

function renderAt(path: string) {
  return render(
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route
          path="*"
          element={
            <DesktopShell tabs={TABS}>
              <LocationProbe />
            </DesktopShell>
          }
        />
      </Routes>
    </MemoryRouter>,
  );
}

describe('DesktopShell', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: { id: 'm1', name: 'Jonas Berger', email: 'jonas@test.local', role: 'mitglied', first: 'Jonas', last: 'Berger' },
    });
  });

  it('has no separate profil tab in the sidebar nav (the profile card replaces it, PR-005)', () => {
    renderAt('/start');
    const nav = document.querySelector('.dt-nav');
    expect(nav?.textContent).not.toContain('Profil');
  });

  it('still renders other tabs (start, übersicht) in the sidebar nav', () => {
    renderAt('/start');
    const nav = document.querySelector('.dt-nav');
    expect(nav?.textContent).toContain('Start');
    expect(nav?.textContent).toContain('Übersicht');
  });

  it('navigates to /profil when the sidebar profile card is clicked', () => {
    renderAt('/start');
    fireEvent.click(screen.getByText('Jonas Berger'));
    expect(screen.getByTestId('path').textContent).toBe('/profil');
  });

  it('marks the profile card active on /profil', () => {
    renderAt('/profil');
    const card = document.querySelector('.dt-user');
    expect(card?.className).toContain('active');
  });
});
