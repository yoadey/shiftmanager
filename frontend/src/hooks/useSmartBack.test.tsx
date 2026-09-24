import { describe, it, expect } from 'vitest';
import { render, fireEvent } from '@testing-library/react';
import { MemoryRouter, Routes, Route, useLocation } from 'react-router-dom';
import { useSmartBack } from './useSmartBack';

function Probe({ fallback }: { fallback: string }) {
  const back = useSmartBack(fallback);
  const location = useLocation();
  return (
    <div>
      <span data-testid="path">{location.pathname}</span>
      <button onClick={back}>close</button>
    </div>
  );
}

// MemoryRouter gives its initial entry the key "default" — exactly the case
// of a direct link or a page reload landing straight on a modal route, where
// there is nothing in-app to go back to.
describe('useSmartBack', () => {
  it('replaces to the fallback when there is no in-app history (direct/deep link)', () => {
    const { getByTestId, getByText } = render(
      <MemoryRouter initialEntries={['/events/e1/bearbeiten']}>
        <Probe fallback="/events/e1" />
      </MemoryRouter>,
    );

    expect(getByTestId('path').textContent).toBe('/events/e1/bearbeiten');
    fireEvent.click(getByText('close'));
    expect(getByTestId('path').textContent).toBe('/events/e1');
  });

  it('goes back in history when the modal was opened in-app', () => {
    const { getByTestId, getByText } = render(
      <MemoryRouter initialEntries={['/events/e1', '/events/e1/bearbeiten']} initialIndex={1}>
        <Routes>
          <Route path="/events/:id" element={<Probe fallback="/events/e1" />} />
          <Route path="/events/:id/bearbeiten" element={<Probe fallback="/events/e1" />} />
        </Routes>
      </MemoryRouter>,
    );

    expect(getByTestId('path').textContent).toBe('/events/e1/bearbeiten');
    fireEvent.click(getByText('close'));
    expect(getByTestId('path').textContent).toBe('/events/e1');
  });
});
