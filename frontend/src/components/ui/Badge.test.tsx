import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { Badge, OccBadge } from './Badge';
import type { ShiftOccupancy } from '@/types';

describe('Badge', () => {
  it('renders children text', () => {
    render(<Badge>Hallo</Badge>);
    expect(screen.getByText('Hallo')).toBeInTheDocument();
  });

  it('applies the kind class (default neutral)', () => {
    const { rerender } = render(<Badge>Standard</Badge>);
    expect(screen.getByText('Standard')).toHaveClass('sm-badge', 'b-neutral');

    rerender(<Badge kind="crit">Offen</Badge>);
    expect(screen.getByText('Offen')).toHaveClass('b-crit');
  });

  it('renders a dot by default and omits it when dot=false', () => {
    const { container, rerender } = render(<Badge>Mit Punkt</Badge>);
    expect(container.querySelector('.dot')).toBeInTheDocument();

    rerender(<Badge dot={false}>Ohne Punkt</Badge>);
    expect(container.querySelector('.dot')).not.toBeInTheDocument();
  });
});

describe('OccBadge', () => {
  it('maps occupancy key to badge kind and shows the label', () => {
    const o: ShiftOccupancy = {
      count: 1,
      min: 2,
      max: 4,
      key: 'warn',
      label: 'Teilweise',
      free: 3,
      needsMore: true,
    };
    render(<OccBadge o={o} />);
    const el = screen.getByText('Teilweise');
    expect(el).toHaveClass('b-warn');
  });
});
