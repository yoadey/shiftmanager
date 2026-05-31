import { apiGet } from './client';
import type { UserRole } from '@/types';

const BASE_URL = (import.meta.env.VITE_API_URL as string | undefined) || '/api/v1';

/** Full URL the login page redirects to in order to start the OIDC flow. */
export function oidcLoginUrl(): string {
  return `${BASE_URL}/auth/login`;
}

/** Identity of the currently authenticated user, derived from the session JWT. */
export interface MeResponse {
  id: string;
  role: UserRole;
  email: string;
}

/**
 * Fetches the authenticated user's identity. Called by the OIDC callback page
 * after the backend redirects with a freshly issued token in the URL fragment,
 * to hydrate the auth store.
 */
export async function getMe(): Promise<MeResponse> {
  return apiGet<MeResponse>('/auth/me');
}
