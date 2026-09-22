import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Button } from './Button';

describe('Button', () => {
  it('renders label and applies the variant class', () => {
    render(<Button variant="danger">Löschen</Button>);
    const btn = screen.getByRole('button', { name: 'Löschen' });
    expect(btn).toHaveClass('sm-btn', 'danger');
  });

  it('fires onClick when clicked', async () => {
    const onClick = vi.fn();
    render(<Button onClick={onClick}>Speichern</Button>);
    await userEvent.click(screen.getByRole('button', { name: 'Speichern' }));
    expect(onClick).toHaveBeenCalledTimes(1);
  });

  it('disabled blocks onClick and sets the disabled attribute', async () => {
    const onClick = vi.fn();
    render(
      <Button onClick={onClick} disabled>
        Gesperrt
      </Button>,
    );
    const btn = screen.getByRole('button', { name: 'Gesperrt' });
    expect(btn).toBeDisabled();
    await userEvent.click(btn);
    expect(onClick).not.toHaveBeenCalled();
  });

  it('loading also disables the button', () => {
    render(<Button loading>Lädt</Button>);
    expect(screen.getByRole('button')).toBeDisabled();
  });

  it('renders an icon svg when icon prop is given', () => {
    const { container } = render(<Button icon="check">Bestätigen</Button>);
    expect(container.querySelector('svg')).toBeInTheDocument();
  });
});
