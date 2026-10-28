import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import MarkdownEditor from './MarkdownEditor';

describe('MarkdownEditor', () => {
  it('renders the initial Markdown content in the WYSIWYG view without loss (V-009)', async () => {
    render(<MarkdownEditor markdown={'# Willkommen\n\nEin **wichtiger** Hinweis für Helfer.'} onChange={vi.fn()} />);
    await waitFor(() => expect(screen.getByRole('heading', { name: 'Willkommen' })).toBeInTheDocument());
    expect(screen.getByText('wichtiger')).toBeInTheDocument();
  });

  it('renders a placeholder when the initial content is empty', async () => {
    render(<MarkdownEditor markdown="" onChange={vi.fn()} placeholder="Worum geht es?" />);
    await waitFor(() => expect(screen.getByText('Worum geht es?')).toBeInTheDocument());
  });

  it('offers a toggle to the raw Markdown source view (diffSourcePlugin)', async () => {
    render(<MarkdownEditor markdown={'Hallo Welt'} onChange={vi.fn()} />);
    await waitFor(() => expect(screen.getByText('Hallo Welt')).toBeInTheDocument());
    // The rich-text/source toggle is one of the toolbar's radio groups.
    expect(screen.getAllByRole('radiogroup').length).toBeGreaterThan(0);
    expect(screen.getByRole('radio', { name: /rich text/i })).toBeInTheDocument();
  });
});
