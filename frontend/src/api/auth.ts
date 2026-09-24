import { useQuery } from '@tanstack/react-query';
import { apiGet, apiPost } from './client';
import type { UserRole } from '@/types';

const BASE_URL = (import.meta.env.VITE_API_URL as string | undefined) || '/api/v1';

/**
 * Full URL the login page redirects to in order to start the OIDC flow.
 * `provider` (from useOIDCProviders/GET /auth/providers) selects which
 * configured provider to use (A-005); omit it when only one is configured.
 */
export function oidcLoginUrl(provider?: string): string {
  const url = `${BASE_URL}/auth/login`;
  return provider ? `${url}?provider=${encodeURIComponent(provider)}` : url;
}

/** A configured OIDC provider, as listed by GET /auth/providers (A-005). */
export interface OIDCProvider {
  name: string;
  label: string;
}

/**
 * Lists the configured OIDC providers, so the login page can render one
 * button per provider. Called before the user is authenticated, so this
 * hits the public (no-JWT) /auth/providers endpoint.
 */
export function useOIDCProviders() {
  return useQuery({
    queryKey: ['oidc-providers'],
    queryFn: () => apiGet<OIDCProvider[]>('/auth/providers'),
    staleTime: Infinity,
  });
}

/** Identity of the currently authenticated user, derived from the session JWT. */
export interface MeResponse {
  id: string;
  role: UserRole;
  email: string;
  firstName?: string;
  lastName?: string;
}

/**
 * Fetches the authenticated user's identity. Called by the OIDC callback page
 * after the backend redirects with a freshly issued token in the URL fragment,
 * to hydrate the auth store.
 */
export async function getMe(): Promise<MeResponse> {
  return apiGet<MeResponse>('/auth/me');
}

/** Freshly issued token returned by the session-refresh endpoint (A-004). */
export interface TokenResponse {
  token: string;
  expiresIn: number;
}

/**
 * Re-issues a JWT with a fresh expiry for the current session, without a
 * full OIDC round-trip. Requires the current token to still be valid.
 */
export async function refreshToken(): Promise<TokenResponse> {
  return apiPost<TokenResponse>('/auth/refresh');
}
