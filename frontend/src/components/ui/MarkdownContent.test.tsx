import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/react';
import { MarkdownContent } from './MarkdownContent';

describe('MarkdownContent', () => {
  it('renders basic Markdown formatting', () => {
    const { container } = render(<MarkdownContent text={'**Fett** und eine Liste:\n\n- Eins\n- Zwei'} />);
    expect(container.querySelector('strong')?.textContent).toBe('Fett');
    expect(container.querySelectorAll('li')).toHaveLength(2);
  });

  it('strips an embedded <script> tag instead of rendering/executing it (V-009 XSS guard)', () => {
    const { container } = render(
      <MarkdownContent text={'Hallo <script>window.__pwned = true;</script> Welt'} />,
    );
    expect(container.querySelector('script')).toBeNull();
    expect((window as unknown as { __pwned?: boolean }).__pwned).toBeUndefined();
  });

  it('strips a javascript: link href instead of leaving it clickable', () => {
    const { container } = render(<MarkdownContent text={'[klick mich](javascript:alert(1))'} />);
    const link = container.querySelector('a');
    const href = link?.getAttribute('href') ?? null;
    expect(href === null || !/^javascript:/i.test(href)).toBe(true);
  });

  it('renders nothing for an empty description', () => {
    const { container } = render(<MarkdownContent text="" />);
    expect(container.firstChild).toBeNull();
  });
});
