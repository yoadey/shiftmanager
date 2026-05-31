import { client } from './generated/client.gen';

/**
 * Configure the generated API client once at app startup.
 * - base URL from env (falls back to /api/v1 for same-origin serving)
 * - JWT injected from localStorage via the auth callback
 * - 401 responses clear the token and redirect to login
 */
export function setupApiClient(): void {
  client.setConfig({
    baseUrl: (import.meta.env.VITE_API_URL as string | undefined) ?? '/api/v1',
    auth: () => localStorage.getItem('sm_token') ?? '',
  });
}
