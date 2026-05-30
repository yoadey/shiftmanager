import { apiPost } from './client';
import type { AuthUser } from '@/types';

const BASE_URL = (import.meta.env.VITE_API_URL as string | undefined) || '/api/v1';

export interface AuthCallbackResponse {
  token: string;
  user: AuthUser;
}

/** Full URL the login page redirects to in order to start the OIDC flow. */
export function oidcLoginUrl(): string {
  return `${BASE_URL}/auth/login`;
}

/** Exchanges the OIDC authorization code for a JWT + user. */
export async function exchangeCode(code: string): Promise<AuthCallbackResponse> {
  return apiPost<AuthCallbackResponse>('/auth/callback', { code });
}
