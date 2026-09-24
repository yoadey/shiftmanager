import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import type { OIDCProvider } from '@/api/auth';

const useOIDCProviders = vi.fn();

vi.mock('@/api/auth', async () => {
  const actual = await vi.importActual<typeof import('@/api/auth')>('@/api/auth');
  return {
    ...actual,
    useOIDCProviders: () => useOIDCProviders(),
  };
});

import { LoginPage } from './LoginPage';

function queryResult(data: OIDCProvider[] | undefined, opts: { isLoading?: boolean } = {}) {
  return { data, isLoading: opts.isLoading ?? false };
}

describe('LoginPage', () => {
  const originalLocation = window.location;

  beforeEach(() => {
    useOIDCProviders.mockReset();
    // window.location.href is normally not writable in jsdom; replace the
    // whole object so the click handler's assignment can be observed.
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { ...originalLocation, href: '' },
    });
  });

  afterEach(() => {
    Object.defineProperty(window, 'location', { configurable: true, value: originalLocation });
  });

  it('single configured provider → one button, login URL names it explicitly', () => {
    useOIDCProviders.mockReturnValue(queryResult([{ name: 'default', label: 'Vereinskonto' }]));
    render(<LoginPage />);

    const button = screen.getByRole('button', { name: /Anmelden mit Vereinskonto/ });
    button.click();

    expect(window.location.href).toBe('/api/v1/auth/login?provider=default');
  });

  it('multiple configured providers → one button per provider, each with its own ?provider=', () => {
    useOIDCProviders.mockReturnValue(queryResult([
      { name: 'verein', label: 'Vereins-SSO' },
      { name: 'google', label: 'Google' },
    ]));
    render(<LoginPage />);

    const vereinBtn = screen.getByRole('button', { name: /Anmelden mit Vereins-SSO/ });
    const googleBtn = screen.getByRole('button', { name: /Anmelden mit Google/ });

    googleBtn.click();
    expect(window.location.href).toBe('/api/v1/auth/login?provider=google');

    vereinBtn.click();
    expect(window.location.href).toBe('/api/v1/auth/login?provider=verein');
  });

  it('providers still loading → button disabled, no premature navigation', () => {
    useOIDCProviders.mockReturnValue(queryResult(undefined, { isLoading: true }));
    render(<LoginPage />);

    const button = screen.getByRole('button', { name: /Anmelden mit Vereinskonto/ });
    expect(button).toBeDisabled();
  });

  it('providers fetch failed → falls back to a single generic button with no ?provider=', () => {
    useOIDCProviders.mockReturnValue(queryResult(undefined, { isLoading: false }));
    render(<LoginPage />);

    const button = screen.getByRole('button', { name: /Anmelden mit Vereinskonto/ });
    expect(button).not.toBeDisabled();
    button.click();

    expect(window.location.href).toBe('/api/v1/auth/login');
  });
});
