import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { ErrorBoundary } from './ErrorBoundary';

function Boom(): JSX.Element {
  throw new Error('kaboom');
}

describe('ErrorBoundary', () => {
  beforeEach(() => vi.spyOn(console, 'error').mockImplementation(() => {}));
  afterEach(() => vi.restoreAllMocks());

  it('renders children when there is no error', () => {
    render(
      <ErrorBoundary>
        <span>alles gut</span>
      </ErrorBoundary>,
    );
    expect(screen.getByText('alles gut')).toBeInTheDocument();
  });

  it('shows a contained error panel (with label) when a child throws', () => {
    render(
      <ErrorBoundary label="der Übersicht">
        <Boom />
      </ErrorBoundary>,
    );
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText(/Fehler ist aufgetreten in der Übersicht/)).toBeInTheDocument();
    expect(screen.getByText(/kaboom/)).toBeInTheDocument();
  });

  it('isolates the failure so sibling content keeps rendering', () => {
    render(
      <div>
        <ErrorBoundary>
          <Boom />
        </ErrorBoundary>
        <span>noch da</span>
      </div>,
    );
    expect(screen.getByRole('alert')).toBeInTheDocument();
    expect(screen.getByText('noch da')).toBeInTheDocument();
  });

  it('recovers via "Erneut versuchen" once the child stops throwing', () => {
    let healthy = false;
    function Flaky(): JSX.Element {
      if (!healthy) throw new Error('temporär');
      return <span>wieder da</span>;
    }

    render(
      <ErrorBoundary>
        <Flaky />
      </ErrorBoundary>,
    );
    expect(screen.getByRole('alert')).toBeInTheDocument();

    // Underlying cause resolved → reset re-renders the healthy child.
    healthy = true;
    fireEvent.click(screen.getByText('Erneut versuchen'));
    expect(screen.getByText('wieder da')).toBeInTheDocument();
  });
});
