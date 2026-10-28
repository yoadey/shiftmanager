import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';

vi.mock('@/api/events', () => ({
  useCreateEvent: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

vi.mock('@/api/shifts', () => ({
  useCreateShift: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

import { CreateEventFlow } from './CreateEventFlow';

describe('CreateEventFlow', () => {
  it('renders as a real page, not inside a Sheet overlay (UX-001, bedienkonzept)', () => {
    render(<CreateEventFlow onClose={vi.fn()} />);
    expect(screen.getByText('Eckdaten')).toBeInTheDocument();
    expect(document.querySelector('.sm-overlay')).not.toBeInTheDocument();
    expect(document.querySelector('.sm-sheet')).not.toBeInTheDocument();
  });

  it('calls onClose when the header back/close button is pressed on the first step', () => {
    const onClose = vi.fn();
    render(<CreateEventFlow onClose={onClose} />);
    fireEvent.click(screen.getByText('Eckdaten').closest('.sm-header')!.querySelector('button')!);
    expect(onClose).toHaveBeenCalled();
  });
});
